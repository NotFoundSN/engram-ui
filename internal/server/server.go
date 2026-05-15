// Package server wires the chi router and handlers for engram-ui.
package server

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/sync/errgroup"

	"github.com/Gentleman-Programming/engram-ui/internal/client"
	"github.com/Gentleman-Programming/engram-ui/internal/engramconv"
	"github.com/Gentleman-Programming/engram-ui/internal/render"
	"github.com/Gentleman-Programming/engram-ui/internal/views"
)

// engramClient is the subset of client.Client that handlers need.
// Kept in the server package so tests can stub it without touching
// the public client API surface.
type engramClient interface {
	Stats() (*client.Stats, error)
	RecentObservations(client.RecentOptions) ([]client.Observation, error)
	Search(string, client.SearchOptions) ([]client.SearchResult, error)
	Observation(int64) (*client.Observation, error)
	Health() error
}

// compile-time assertion: *client.Client satisfies engramClient.
var _ engramClient = (*client.Client)(nil)

type Server struct {
	client engramClient
	router *chi.Mux
}

func New(c *client.Client) *Server {
	return newWithClient(c)
}

func newWithClient(c engramClient) *Server {
	s := &Server{client: c, router: chi.NewRouter()}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler { return s.router }

func (s *Server) routes() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Logger)

	s.router.Get("/", s.handleHome)
	s.router.Get("/p/{project}", s.handleProject)
	s.router.Get("/observations/{id}", s.handleObservation)
	s.router.Get("/healthz", s.handleHealthz)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if err := s.client.Health(); err != nil {
		http.Error(w, "engram unreachable: "+err.Error(), http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

const (
	homeSessionFetchLimit = 10
	homePreviewMaxRunes   = 140
	homeFanoutLimit       = 8
)

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	stats, err := s.client.Stats()
	if err != nil {
		renderError(w, r, "could not fetch stats", err)
		return
	}

	projects := stats.Projects
	cards := make([]views.ProjectCard, len(projects))
	for i, p := range projects {
		cards[i] = views.ProjectCard{Name: p}
	}

	var g errgroup.Group
	g.SetLimit(homeFanoutLimit)

	for i, p := range projects {
		i, p := i, p // capture loop vars
		g.Go(func() error {
			obs, fetchErr := s.client.RecentObservations(client.RecentOptions{
				Project: p,
				Limit:   homeSessionFetchLimit,
			})
			if fetchErr != nil {
				// Graceful degradation: keep HasSession=false for this card.
				return nil
			}
			// Client-side filter for session_summary (engram ?type= not supported).
			for _, o := range obs {
				if o.Type == "session_summary" {
					cards[i].HasSession = true
					cards[i].SessionPreview = render.Truncate(o.Content, homePreviewMaxRunes)
					break
				}
			}
			return nil
		})
	}
	// errgroup.Wait never returns non-nil here (errors are absorbed above).
	_ = g.Wait()

	_ = views.Home(views.HomeData{Cards: cards}).Render(r.Context(), w)
}

func (s *Server) handleProject(w http.ResponseWriter, r *http.Request) {
	project := chi.URLParam(r, "project")
	q := r.URL.Query().Get("q")
	activeType := r.URL.Query().Get("type")
	sortParam := r.URL.Query().Get("sort")
	_, sortExplicit := r.URL.Query()["sort"] // track explicit presence before normalizing

	// Validate sort; any unrecognized value defaults to date_desc.
	// sortExplicit stays true even on a bad value (user explicitly set something).
	if sortParam != "date_asc" && sortParam != "date_desc" {
		sortParam = "date_desc"
	}

	var observations []client.Observation
	isSearch := q != ""

	if isSearch {
		results, err := s.client.Search(q, client.SearchOptions{
			Project: project,
			Limit:   100,
		})
		if err != nil {
			renderError(w, r, "could not search observations", err)
			return
		}
		observations = make([]client.Observation, len(results))
		for i, r := range results {
			observations[i] = r.Observation
		}
	} else {
		recent, err := s.client.RecentObservations(client.RecentOptions{
			Project: project,
			Limit:   100,
		})
		if err != nil {
			renderError(w, r, "could not fetch observations", err)
			return
		}
		observations = recent
		// Apply type filter client-side.
		if activeType != "" {
			observations = filterByType(observations, activeType)
		}
		// Sort client-side (engram returns newest-first by default; handle date_asc).
		observations = sortedObservations(observations, sortParam)
	}

	presentTypes := distinctTypes(observations)
	typeOptions := unionTypes(engramconv.CanonicalTypes, presentTypes, activeType)
	sourceURL := buildSourceURL(project, activeType, q, sortParam, sortExplicit)

	data := views.ProjectData{
		Project:        project,
		Observations:   observations,
		AvailableTypes: typeOptions,
		ActiveType:     activeType,
		Query:          q,
		Sort:           sortParam,
		SortExplicit:   sortExplicit,
		IsSearch:       isSearch,
		SourceURL:      sourceURL,
	}
	_ = views.Project(data).Render(r.Context(), w)
}

func (s *Server) handleObservation(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid observation id", http.StatusBadRequest)
		return
	}

	obs, err := s.client.Observation(id)
	if err != nil {
		renderError(w, r, "could not fetch observation", err)
		return
	}

	renderMD := r.URL.Query().Get("raw") == ""
	back := validateFrom(r.URL.Query().Get("from"))

	_ = views.ObservationDetail(obs, renderMD, back).Render(r.Context(), w)
}

func renderError(w http.ResponseWriter, r *http.Request, msg string, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	_ = views.ErrorPage(msg, err.Error()).Render(r.Context(), w)
}

func filterByType(obs []client.Observation, typ string) []client.Observation {
	if typ == "" {
		return obs
	}
	out := make([]client.Observation, 0, len(obs))
	for _, o := range obs {
		if o.Type == typ {
			out = append(out, o)
		}
	}
	return out
}

func distinctTypes(obs []client.Observation) []string {
	seen := make(map[string]struct{}, len(obs))
	for _, o := range obs {
		if o.Type == "" {
			continue
		}
		seen[o.Type] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for t := range seen {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

func containsString(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

// unionTypes returns an alphabetically sorted, deduplicated union of canonical
// types, project-present types, and the active phantom type (if any).
// The active type is added to the union even when it appears in neither
// canonical nor present slices (FR-4.6).
func unionTypes(canonical, present []string, active string) []string {
	seen := make(map[string]struct{}, len(canonical)+len(present)+1)
	for _, t := range canonical {
		seen[t] = struct{}{}
	}
	for _, t := range present {
		if t == "" {
			continue
		}
		seen[t] = struct{}{}
	}
	if active != "" {
		seen[active] = struct{}{} // phantom inclusion
	}
	out := make([]string, 0, len(seen))
	for t := range seen {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

func sortedObservations(obs []client.Observation, direction string) []client.Observation {
	out := make([]client.Observation, len(obs))
	copy(out, obs)
	if direction == "date_asc" {
		sort.Slice(out, func(i, j int) bool {
			return out[i].CreatedAt < out[j].CreatedAt
		})
	} else {
		// default: date_desc
		sort.Slice(out, func(i, j int) bool {
			return out[i].CreatedAt > out[j].CreatedAt
		})
	}
	return out
}
