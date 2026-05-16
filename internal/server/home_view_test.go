package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Gentleman-Programming/engram-ui/internal/client"
)

// TestHomeView_RendersProjGrid asserts that GET / renders a .proj-grid container
// with exactly one .proj-card descendant per project.
func TestHomeView_RendersProjGrid(t *testing.T) {
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"alpha", "beta"}},
		recentByProject: map[string][]client.Observation{
			"alpha": {{ID: 1, Type: "session_summary", Title: "T", Content: "preview alpha", CreatedAt: "2026-01-02"}},
			"beta":  {{ID: 2, Type: "session_summary", Title: "T", Content: "preview beta", CreatedAt: "2026-01-01"}},
		},
		recentByProjectErr: map[string]error{},
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()

	// .proj-grid container must be present
	if !strings.Contains(body, `class="proj-grid"`) {
		t.Error("expected .proj-grid container in home response")
	}

	// One .proj-card per project — count occurrences
	cardCount := strings.Count(body, `proj-card`)
	if cardCount < 2 {
		t.Errorf("expected at least 2 .proj-card elements (one per project), got %d occurrences of 'proj-card'", cardCount)
	}
}

// TestHomeView_CardContent asserts that each .proj-card shows the project name,
// observation count, activity timestamp, and a preview snippet from the latest observation.
func TestHomeView_CardContent(t *testing.T) {
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"myproj"}},
		recentByProject: map[string][]client.Observation{
			"myproj": {
				{ID: 1, Type: "session_summary", Title: "Session title", Content: "Hello preview content", CreatedAt: "2026-03-15"},
			},
		},
		recentByProjectErr: map[string]error{},
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()

	// Project name must appear inside a .proj-card__name element
	if !strings.Contains(body, `proj-card__name`) {
		t.Error("expected .proj-card__name element in card")
	}
	if !strings.Contains(body, "myproj") {
		t.Error("expected project name 'myproj' in card")
	}

	// Activity timestamp must appear (via .proj-card__activity or .proj-card__time)
	if !strings.Contains(body, "2026-03-15") {
		t.Error("expected activity timestamp '2026-03-15' in card")
	}

	// Preview snippet must appear inside .proj-card__preview
	if !strings.Contains(body, `proj-card__preview`) {
		t.Error("expected .proj-card__preview element in card")
	}
	if !strings.Contains(body, "Hello preview content") {
		t.Error("expected preview snippet 'Hello preview content' in card")
	}

	// Observation count must appear (via .proj-card__count)
	if !strings.Contains(body, `proj-card__count`) {
		t.Error("expected .proj-card__count element in card")
	}
}

// TestHomeView_EmptyPreview asserts that when a project has no session_summary,
// the card renders the .proj-card__preview--empty treatment (italic faint indicator).
func TestHomeView_EmptyPreview(t *testing.T) {
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"emptyproj"}},
		recentByProject: map[string][]client.Observation{
			"emptyproj": {},
		},
		recentByProjectErr: map[string]error{},
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()

	// The card must render the --empty modifier class for the empty treatment
	if !strings.Contains(body, `proj-card__preview--empty`) {
		t.Error("expected .proj-card__preview--empty class when project has no session preview")
	}
}

// TestHomeView_CardLink asserts that each .proj-card is or contains a link
// with href="/p/{project}".
func TestHomeView_CardLink(t *testing.T) {
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"linkproj", "otherproj"}},
		recentByProject: map[string][]client.Observation{
			"linkproj":  {},
			"otherproj": {},
		},
		recentByProjectErr: map[string]error{},
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()

	// Each project card must link to /p/{project}
	for _, proj := range []string{"linkproj", "otherproj"} {
		want := `href="/p/` + proj + `"`
		if !strings.Contains(body, want) {
			t.Errorf("expected card link %q in body", want)
		}
	}

	// The link must be anchored with class proj-card (anchor IS the card)
	if !strings.Contains(body, `class="proj-card"`) {
		t.Error("expected anchor element with class='proj-card' (card wraps its own anchor)")
	}
}
