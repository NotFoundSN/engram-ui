package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NotFoundSN/engram-ui/internal/client"
)

// projectViewStub returns a server with project "alpha" having two observations
// with distinct types and a multi-segment topic_key.
func projectViewStub() *Server {
	tk1 := "architecture/auth-model/design"
	tk2 := "bugfix/login"
	return newWithClient(&stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"alpha"}},
		recentOut: []client.Observation{
			{ID: 1, Type: "architecture", Title: "Auth Model Design", Content: "c1", CreatedAt: "2026-01-02", TopicKey: &tk1},
			{ID: 2, Type: "bugfix", Title: "Login Fix", Content: "c2", CreatedAt: "2026-01-01", TopicKey: &tk2},
		},
	})
}

// --- proj-head ---

func TestProjectView_ProjHead(t *testing.T) {
	s := projectViewStub()
	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()

	if !strings.Contains(body, `class="proj-head"`) {
		t.Fatal(`expected element with class="proj-head" in body`)
	}
	if !strings.Contains(body, "alpha") {
		t.Error("expected project name 'alpha' inside .proj-head")
	}
	// Count "2" must appear in the head along with the "observations" label.
	if !strings.Contains(body, ">2<") {
		t.Error("expected observation count '2' inside .proj-head")
	}
	if !strings.Contains(body, "observations") {
		t.Error(`expected the literal "observations" label in .proj-head`)
	}
}

// --- filter bar ---

func TestProjectView_FilterBarStructure(t *testing.T) {
	s := projectViewStub()
	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	if !strings.Contains(body, `class="filter-bar"`) {
		t.Fatal(`expected element with class="filter-bar" in body`)
	}

	// Search input for q parameter.
	if !strings.Contains(body, `type="search"`) || !strings.Contains(body, `name="q"`) {
		t.Error("expected a search input for q in .filter-bar")
	}

	// SVG icon decorating the search input.
	if !strings.Contains(body, `filter-bar__search-icon`) {
		t.Error(`expected .filter-bar__search-icon (SVG) inside the search form`)
	}

	// Type select replaces the old chip row.
	if !strings.Contains(body, `name="type"`) {
		t.Error(`expected <select name="type"> filter in .filter-bar`)
	}

	// Sort select preserved (default IsSearch=false on this stub).
	if !strings.Contains(body, `name="sort"`) {
		t.Error(`expected <select name="sort"> in .filter-bar`)
	}

	// Both selects must be wrapped in .filter-bar__selects.
	if !strings.Contains(body, `class="filter-bar__selects"`) {
		t.Error(`expected .filter-bar__selects container around type/sort selects`)
	}
}

func TestProjectView_TypeSelectAllOptionDefault(t *testing.T) {
	s := projectViewStub()
	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	// When no ?type= is set, the "All types" option must be the selected one.
	// templ renders `selected` as the bare attribute when truthy.
	if !strings.Contains(body, `<option value="" selected`) {
		t.Error(`expected default "All types" option to be selected (value="" selected)`)
	}
	if !strings.Contains(body, "All types") {
		t.Error(`expected the "All types" option label`)
	}
}

func TestProjectView_TypeSelectActiveSelected(t *testing.T) {
	s := newWithClient(&stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"alpha"}},
		recentOut: []client.Observation{
			{ID: 1, Type: "decision", Title: "D1", Content: "c", CreatedAt: "2026-01-01"},
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/p/alpha?type=decision", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	// The decision option must be marked selected; default empty option must not.
	if !strings.Contains(body, `<option value="decision" selected`) {
		t.Error(`expected the "decision" option to be selected when ?type=decision`)
	}
}

func TestProjectView_TypeSelectPreservesParams(t *testing.T) {
	s := newWithClient(&stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"alpha"}},
		recentOut: []client.Observation{
			{ID: 1, Type: "decision", Title: "D1", Content: "c", CreatedAt: "2026-01-01"},
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/p/alpha?q=foo&sort=date_asc&topic_key_prefix=sdd/auth/", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	// Each filter form must carry the other params as hidden inputs so a
	// select change keeps the current state.
	if !strings.Contains(body, `<input type="hidden" name="q" value="foo"`) {
		t.Error(`expected q preserved as a hidden input in filter forms`)
	}
	if !strings.Contains(body, `<input type="hidden" name="sort" value="date_asc"`) {
		t.Error(`expected sort preserved as a hidden input in filter forms`)
	}
	if !strings.Contains(body, `<input type="hidden" name="topic_key_prefix" value="sdd/auth/"`) {
		t.Error(`expected topic_key_prefix preserved as a hidden input`)
	}
}

// --- prefix chip ---

func TestProjectView_PrefixChipUsesDesignClass(t *testing.T) {
	k := "sdd/auth/spec"
	s := newWithClient(&stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"alpha"}},
		recentOut: []client.Observation{
			{ID: 1, Type: "spec", Title: "Auth Spec", Content: "c", CreatedAt: "2026-01-01", TopicKey: &k},
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/p/alpha?topic_key_prefix=sdd/auth/", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	if !strings.Contains(body, `id="prefix-chip"`) {
		t.Error(`expected id="prefix-chip" when topic_key_prefix is active`)
	}
	if !strings.Contains(body, `class="prefix-chip"`) {
		t.Error(`expected class="prefix-chip"`)
	}
}

// --- obs list / row ---

func TestProjectView_ObsRowsRendered(t *testing.T) {
	s := projectViewStub()
	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	if !strings.Contains(body, `class="obs-list"`) {
		t.Error(`expected element with class="obs-list" in body`)
	}

	// Two observations → two anchors with class="obs-row".
	if got := strings.Count(body, `class="obs-row"`); got < 2 {
		t.Errorf("expected at least 2 .obs-row anchors, got %d", got)
	}

	// Both observation titles must appear.
	if !strings.Contains(body, "Auth Model Design") {
		t.Error("expected observation title 'Auth Model Design' in body")
	}
	if !strings.Contains(body, "Login Fix") {
		t.Error("expected observation title 'Login Fix' in body")
	}
}

func TestProjectView_ObsRowRichColumns(t *testing.T) {
	s := projectViewStub()
	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	// Index column carries zero-padded numbers (01, 02, …).
	if !strings.Contains(body, `class="obs-row__idx mono"`) {
		t.Error(`expected .obs-row__idx column`)
	}
	if !strings.Contains(body, ">01<") {
		t.Error(`expected the "01" index for the first row`)
	}
	if !strings.Contains(body, ">02<") {
		t.Error(`expected the "02" index for the second row`)
	}

	// Type dot + type label.
	if !strings.Contains(body, `class="obs-row__dot"`) {
		t.Error(`expected .obs-row__dot column`)
	}
	if !strings.Contains(body, `class="type-dot"`) {
		t.Error(`expected nested .type-dot element`)
	}
	if !strings.Contains(body, `class="obs-row__type mono"`) {
		t.Error(`expected .obs-row__type column`)
	}

	// ID column shows #<id>.
	if !strings.Contains(body, `class="obs-row__id mono"`) {
		t.Error(`expected .obs-row__id column`)
	}
	if !strings.Contains(body, "#1") || !strings.Contains(body, "#2") {
		t.Error(`expected "#1" and "#2" id labels`)
	}

	// Time column carries the absolute timestamp in the title attribute via
	// render.FormatDateTime — "2026-01-02" parses to "Jan 2, 2026 00:00 UTC".
	if !strings.Contains(body, `class="obs-row__time mono"`) {
		t.Error(`expected .obs-row__time column`)
	}
	if !strings.Contains(body, "Jan 2, 2026") {
		t.Error(`expected formatted timestamp "Jan 2, 2026" in row tooltip`)
	}
}

func TestProjectView_ObsRowDataTypeAttribute(t *testing.T) {
	s := projectViewStub()
	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	// Each row's anchor must carry data-type so CSS resolves the right hue.
	if !strings.Contains(body, `data-type="architecture"`) {
		t.Error(`expected data-type="architecture" on a row anchor`)
	}
	if !strings.Contains(body, `data-type="bugfix"`) {
		t.Error(`expected data-type="bugfix" on a row anchor`)
	}
}

func TestProjectView_TopicPillSegmented(t *testing.T) {
	tk := "architecture/auth-model/design"
	s := newWithClient(&stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"alpha"}},
		recentOut: []client.Observation{
			{ID: 1, Type: "architecture", Title: "Auth Design", Content: "c", CreatedAt: "2026-01-01", TopicKey: &tk},
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	if !strings.Contains(body, `class="topic-pill topic-pill--xs"`) {
		t.Fatal(`expected .topic-pill.topic-pill--xs (compact variant for dense rows)`)
	}

	if got := strings.Count(body, `class="topic-pill__seg"`); got < 3 {
		t.Errorf("expected 3 .topic-pill__seg elements for the 3-segment key, got %d", got)
	}
	for _, seg := range []string{"architecture", "auth-model", "design"} {
		if !strings.Contains(body, seg) {
			t.Errorf("expected segment %q in topic pill", seg)
		}
	}
}

func TestProjectView_TopicPillAbsentWhenNoTopicKey(t *testing.T) {
	s := newWithClient(&stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"alpha"}},
		recentOut: []client.Observation{
			{ID: 1, Type: "decision", Title: "No Key", Content: "c", CreatedAt: "2026-01-01"},
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	if !strings.Contains(body, `class="obs-list"`) {
		t.Error(`expected obs-list to render even without a topic_key`)
	}
	if strings.Contains(body, `class="topic-pill`) {
		t.Error("topic-pill must not render when observation has no topic_key")
	}
}
