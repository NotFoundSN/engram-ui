// Package server wires the chi router and handlers for engram-ui.
package server

import (
	"mime"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/sync/errgroup"

	"github.com/Gentleman-Programming/engram-ui/internal/client"
	"github.com/Gentleman-Programming/engram-ui/internal/engramconv"
	"github.com/Gentleman-Programming/engram-ui/internal/render"
	"github.com/Gentleman-Programming/engram-ui/internal/static"
	"github.com/Gentleman-Programming/engram-ui/internal/views"
)

func init() {
	// Windows does not always register font/woff2 in the system MIME table;
	// register it explicitly so http.FileServer returns the correct Content-Type.
	_ = mime.AddExtensionType(".woff2", "font/woff2")
}

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

	// Static asset subrouter — registered first so chi's trie never confuses it
	// with parameterized routes. Long-cache headers applied at the subrouter scope.
	s.router.Route("/static", func(r chi.Router) {
		r.Use(longCacheMiddleware)
		fs := http.FileServer(http.FS(static.FS()))
		r.Handle("/*", http.StripPrefix("/static", fs))
	})

	s.router.Get("/", s.handleHome)
	s.router.Get("/p/{project}", s.handleProject)
	s.router.Get("/observations/{id}", s.handleObservation)
	// Short alias for agent emission — web templates keep using /observations/{id}.
	s.router.Get("/m/{id}", s.handleObservation)
	s.router.Get("/healthz", s.handleHealthz)
}

// longCacheMiddleware sets an immutable long-cache header on all responses it
// wraps. Mounted only on the /static/* subrouter so no other handler is affected.
func longCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
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
					cards[i].LastActivity = o.CreatedAt
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
	topicKeyPrefix := r.URL.Query().Get("topic_key_prefix")
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

	// Apply topic_key_prefix filter (works for both search and recent paths).
	if topicKeyPrefix != "" {
		observations = filterByTopicKeyPrefix(observations, topicKeyPrefix)
	}

	presentTypes := distinctTypes(observations)
	typeOptions := unionTypes(engramconv.CanonicalTypes, presentTypes, activeType)
	sourceURL := buildSourceURL(project, activeType, q, sortParam, sortExplicit, topicKeyPrefix)

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
		TopicKeyPrefix: topicKeyPrefix,
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

	siblings, prefix, hasMore := s.computeSiblings(obs)

	_ = views.ObservationDetail(obs, renderMD, back, siblings, prefix, hasMore).Render(r.Context(), w)
}

const siblingsCap = 20

// computeSiblings applies the X+ rule: returns siblings only when the
// observation's topic_key has ≥2 `/` (3+ segments). Returns the sibling slice
// sorted by created_at asc, the prefix used for the lookup, and a hasMore
// flag if more siblings exist beyond the cap.
func (s *Server) computeSiblings(obs *client.Observation) ([]client.Observation, string, bool) {
	if obs == nil || obs.TopicKey == nil {
		return nil, "", false
	}
	topicKey := *obs.TopicKey
	if strings.Count(topicKey, "/") < 2 {
		return nil, "", false
	}
	// strip-last-segment prefix: everything up to and including the final `/`.
	lastSlash := strings.LastIndex(topicKey, "/")
	prefix := topicKey[:lastSlash+1]

	project := ""
	if obs.Project != nil {
		project = *obs.Project
	}
	recent, err := s.client.RecentObservations(client.RecentOptions{
		Project: project,
		Limit:   200,
	})
	if err != nil {
		return nil, prefix, false
	}
	// Filter by prefix.
	matched := make([]client.Observation, 0, len(recent))
	for _, o := range recent {
		if o.TopicKey == nil {
			continue
		}
		if strings.HasPrefix(*o.TopicKey, prefix) {
			matched = append(matched, o)
		}
	}
	// Sort by created_at asc.
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt < matched[j].CreatedAt
	})
	// Cap at siblingsCap.
	hasMore := len(matched) > siblingsCap
	if hasMore {
		matched = matched[:siblingsCap]
	}
	return matched, prefix, hasMore
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

func filterByTopicKeyPrefix(obs []client.Observation, prefix string) []client.Observation {
	if prefix == "" {
		return obs
	}
	out := make([]client.Observation, 0, len(obs))
	for _, o := range obs {
		if o.TopicKey == nil {
			continue
		}
		if strings.HasPrefix(*o.TopicKey, prefix) {
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
