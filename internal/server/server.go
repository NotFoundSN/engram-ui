// Package server wires the chi router and handlers for engram-ui.
package server

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Gentleman-Programming/engram-ui/internal/client"
	"github.com/Gentleman-Programming/engram-ui/internal/views"
)

type Server struct {
	client *client.Client
	router *chi.Mux
}

func New(c *client.Client) *Server {
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

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	stats, err := s.client.Stats()
	if err != nil {
		renderError(w, r, "could not fetch stats", err)
		return
	}

	recent, err := s.client.RecentObservations(client.RecentOptions{Limit: 20})
	if err != nil {
		renderError(w, r, "could not fetch recent observations", err)
		return
	}

	_ = views.Home(stats, recent).Render(r.Context(), w)
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

	_ = views.ObservationDetail(obs).Render(r.Context(), w)
}

func renderError(w http.ResponseWriter, r *http.Request, msg string, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	_ = views.ErrorPage(msg, err.Error()).Render(r.Context(), w)
}
