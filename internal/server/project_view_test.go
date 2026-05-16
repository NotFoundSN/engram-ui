package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Gentleman-Programming/engram-ui/internal/client"
)

// --- Phase 5 project view tests ---

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

// --- Task 5.1: .proj-head + .filter-bar skeleton ---

func TestProjectView_ProjHead(t *testing.T) {
	s := projectViewStub()
	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()

	// .proj-head must contain project name and observation count
	projHeadIdx := strings.Index(body, `class="proj-head"`)
	if projHeadIdx == -1 {
		t.Fatal(`expected element with class="proj-head" in body`)
	}

	// Find the closing tag of proj-head to scope the check
	tail := body[projHeadIdx:]
	// Look for closing div within next 600 chars (generous)
	projHeadBlock := tail[:min(len(tail), 600)]

	if !strings.Contains(projHeadBlock, "alpha") {
		t.Error("expected project name 'alpha' inside .proj-head")
	}
	// Observation count (2 observations)
	if !strings.Contains(projHeadBlock, "2") {
		t.Error("expected observation count '2' inside .proj-head")
	}
}

func TestProjectView_FilterBarStructure(t *testing.T) {
	s := projectViewStub()
	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	// .filter-bar must be present
	filterBarIdx := strings.Index(body, `class="filter-bar"`)
	if filterBarIdx == -1 {
		t.Fatal(`expected element with class="filter-bar" in body`)
	}

	// Must contain a text input for search
	if !strings.Contains(body, `type="search"`) && !strings.Contains(body, `name="q"`) {
		t.Error("expected a text/search input for search in .filter-bar")
	}

	// Must contain a sort selector (select element with name="sort")
	if !strings.Contains(body, `name="sort"`) {
		t.Error("expected sort selector (name=\"sort\") in .filter-bar")
	}

	// Must contain .filter-chip-row
	if !strings.Contains(body, `class="filter-chip-row"`) {
		t.Error(`expected element with class="filter-chip-row" in .filter-bar`)
	}
}

func TestProjectView_NoSelectTypeElement(t *testing.T) {
	s := projectViewStub()
	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	// The body must NOT contain <select name="type">
	// templ renders attributes in order: id before name, so check both patterns
	if strings.Contains(body, `name="type"`) && strings.Contains(body, `<select`) {
		// Find if the select with name type is present — more precise check
		selectIdx := strings.Index(body, `<select`)
		for selectIdx != -1 {
			selectEnd := strings.Index(body[selectIdx:], ">")
			if selectEnd != -1 {
				selectTag := body[selectIdx : selectIdx+selectEnd+1]
				if strings.Contains(selectTag, `name="type"`) {
					t.Error("body must NOT contain <select name=\"type\"> — chips should replace it")
					break
				}
			}
			next := strings.Index(body[selectIdx+1:], `<select`)
			if next == -1 {
				break
			}
			selectIdx = selectIdx + 1 + next
		}
	}
}

// --- Task 5.2: .filter-chip-row replaces <select> ---

func TestProjectView_ChipsReplaceSelect(t *testing.T) {
	s := projectViewStub()
	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	// .filter-chip-row must be present
	chipRowIdx := strings.Index(body, `class="filter-chip-row"`)
	if chipRowIdx == -1 {
		t.Fatal(`expected element with class="filter-chip-row" in body`)
	}

	// Must contain anchor chips (class="chip")
	if !strings.Contains(body, `class="chip`) {
		t.Error("expected anchor chips with class=\"chip\" in .filter-chip-row")
	}

	// Must contain an "all" reset chip
	if !strings.Contains(body, "all") {
		t.Error("expected an 'all' reset chip in .filter-chip-row")
	}

	// No <select name="type"> should be in the body
	if strings.Contains(body, `<select`) {
		// Only the sort select is allowed
		selectCount := strings.Count(body, "<select")
		// There should be exactly 1 select (the sort select), not one for type
		if strings.Count(body, `name="type"`) > 0 && strings.Contains(body, `<select`) {
			// Verify no select has name="type"
			idx := strings.Index(body, "<select")
			for idx != -1 {
				end := strings.Index(body[idx:], ">")
				if end != -1 {
					tag := body[idx : idx+end+1]
					if strings.Contains(tag, `name="type"`) {
						t.Errorf("found <select name=\"type\"> but it should be replaced with chips (found %d selects total)", selectCount)
					}
				}
				next := strings.Index(body[idx+1:], "<select")
				if next == -1 {
					break
				}
				idx = idx + 1 + next
			}
		}
	}
}

func TestProjectView_ActiveChipMarked(t *testing.T) {
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

	// The decision chip must have is-active class
	if !strings.Contains(body, `is-active`) {
		t.Error("expected is-active class on the active chip")
	}

	// The decision chip must have aria-current="page"
	if !strings.Contains(body, `aria-current="page"`) {
		t.Error("expected aria-current=\"page\" on the active chip")
	}

	// Find the chip for "decision" specifically
	// Look for a chip link that contains "decision" and has is-active
	decisionChipFound := false
	searchIn := body
	for {
		chipIdx := strings.Index(searchIn, `class="chip`)
		if chipIdx == -1 {
			break
		}
		// Find the end of this anchor tag (up to </a>)
		aEnd := strings.Index(searchIn[chipIdx:], "</a>")
		if aEnd == -1 {
			break
		}
		chipBlock := searchIn[chipIdx : chipIdx+aEnd+4]
		if strings.Contains(chipBlock, "decision") && strings.Contains(chipBlock, "is-active") {
			decisionChipFound = true
			break
		}
		searchIn = searchIn[chipIdx+1:]
	}
	if !decisionChipFound {
		t.Error("expected chip for 'decision' type with is-active class")
	}
}

func TestProjectView_AllChipActiveWhenNoType(t *testing.T) {
	s := projectViewStub()
	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	// Find the "all" chip and verify it has is-active
	allChipFound := false
	searchIn := body
	for {
		chipIdx := strings.Index(searchIn, `class="chip`)
		if chipIdx == -1 {
			break
		}
		aEnd := strings.Index(searchIn[chipIdx:], "</a>")
		if aEnd == -1 {
			break
		}
		chipBlock := searchIn[chipIdx : chipIdx+aEnd+4]
		if strings.Contains(chipBlock, "all") && strings.Contains(chipBlock, "is-active") {
			allChipFound = true
			break
		}
		searchIn = searchIn[chipIdx+1:]
	}
	if !allChipFound {
		t.Error("expected 'all' chip to have is-active class when no ?type= is set")
	}
}

func TestProjectView_ChipHrefsPreserveParams(t *testing.T) {
	s := newWithClient(&stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"alpha"}},
		recentOut: []client.Observation{
			{ID: 1, Type: "decision", Title: "D1", Content: "c", CreatedAt: "2026-01-01"},
		},
	})
	// Request with q, sort, and topic_key_prefix active
	req := httptest.NewRequest(http.MethodGet, "/p/alpha?q=foo&sort=date_asc&topic_key_prefix=sdd/auth/", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	// Each chip's href must include q=foo, sort=date_asc, topic_key_prefix
	// The projectFilterHref helper encodes these — check that at least one chip href
	// contains the preserved params
	if !strings.Contains(body, "q=foo") {
		t.Error("expected q=foo preserved in chip hrefs")
	}
	if !strings.Contains(body, "sort=date_asc") {
		t.Error("expected sort=date_asc preserved in chip hrefs")
	}
	if !strings.Contains(body, "topic_key_prefix") {
		t.Error("expected topic_key_prefix preserved in chip hrefs")
	}
}

// --- Task 5.3: .prefix-chip class (existing activePrefixChip, just verify classes) ---
// The existing test TestHandleProject_ActivePrefixChipRenders already covers id="prefix-chip".
// We verify the new .prefix-chip class is used in addition.

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

	// Must still have id="prefix-chip" (existing test contract)
	if !strings.Contains(body, `id="prefix-chip"`) {
		t.Error(`expected id="prefix-chip" in body when topic_key_prefix is active`)
	}

	// Must use .prefix-chip class
	if !strings.Contains(body, `class="prefix-chip"`) {
		t.Error(`expected class="prefix-chip" in the prefix chip element`)
	}
}

// --- Task 5.4: .obs-list + .obs-row + segmented .topic-pill ---

func TestProjectView_ObsRowsRendered(t *testing.T) {
	s := projectViewStub()
	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	// .obs-list must be present
	if !strings.Contains(body, `class="obs-list"`) {
		t.Error(`expected element with class="obs-list" in body`)
	}

	// Each observation must be in a .obs-row element
	obsRowCount := strings.Count(body, `class="obs-row"`)
	if obsRowCount < 2 {
		t.Errorf("expected at least 2 .obs-row elements, got %d", obsRowCount)
	}

	// Both observation titles must appear
	if !strings.Contains(body, "Auth Model Design") {
		t.Error("expected observation title 'Auth Model Design' in body")
	}
	if !strings.Contains(body, "Login Fix") {
		t.Error("expected observation title 'Login Fix' in body")
	}
}

func TestProjectView_TopicPillSegmented(t *testing.T) {
	// Observation with topic_key = "architecture/auth-model/design" (3 segments)
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

	// .topic-pill must be present
	if !strings.Contains(body, `class="topic-pill"`) {
		t.Fatal(`expected element with class="topic-pill" in body`)
	}

	// Segments must be present as .topic-pill__seg spans
	segCount := strings.Count(body, `class="topic-pill__seg"`)
	if segCount < 3 {
		t.Errorf("expected at least 3 .topic-pill__seg elements for 'architecture/auth-model/design', got %d", segCount)
	}

	// Each segment text must appear
	if !strings.Contains(body, "architecture") {
		t.Error("expected segment 'architecture' in topic pill")
	}
	if !strings.Contains(body, "auth-model") {
		t.Error("expected segment 'auth-model' in topic pill")
	}
	if !strings.Contains(body, "design") {
		t.Error("expected segment 'design' in topic pill")
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

	// .obs-list and .obs-row must still render
	if !strings.Contains(body, `class="obs-list"`) {
		t.Error(`expected obs-list even when observation has no topic_key`)
	}
	// topic-pill should not render for this obs (no topic_key)
	if strings.Contains(body, `class="topic-pill"`) {
		t.Error("topic-pill must not render when observation has no topic_key")
	}
}
