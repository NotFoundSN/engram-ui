// SCN: observation view — .detail-grid + .meta-sidebar + .markdown
// AUX: makeObsStub (server_test.go), newWithClient (server_test.go)
package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Gentleman-Programming/engram-ui/internal/client"
)

// --- Task 6.1 ---

// TestObservationView_DetailGridPresent asserts the observation detail page
// contains a .detail-grid element.
func TestObservationView_DetailGridPresent(t *testing.T) {
	s := newWithClient(makeObsStub(42))
	req := httptest.NewRequest(http.MethodGet, "/observations/42", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `class="detail-grid"`) {
		t.Error(`expected element with class="detail-grid" in observation detail body`)
	}
}

// TestObservationView_ObsDetailWrapper asserts the observation detail page
// wraps content in an .obs-detail element containing the .detail-grid.
func TestObservationView_ObsDetailWrapper(t *testing.T) {
	s := newWithClient(makeObsStub(42))
	req := httptest.NewRequest(http.MethodGet, "/observations/42", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	// .obs-detail must be present.
	if !strings.Contains(body, `class="obs-detail"`) {
		t.Error(`expected element with class="obs-detail" in observation detail body`)
	}
	// .obs-detail must appear before .detail-grid (wrapper → inner).
	obsDetailIdx := strings.Index(body, `class="obs-detail"`)
	detailGridIdx := strings.Index(body, `class="detail-grid"`)
	if obsDetailIdx == -1 || detailGridIdx == -1 {
		t.Fatal("obs-detail or detail-grid missing")
	}
	if obsDetailIdx > detailGridIdx {
		t.Error(".obs-detail must appear before (wrap) .detail-grid in the HTML")
	}
}

// TestObservationView_MetadataSidebar asserts all 9 required metadata fields
// appear in the .meta-sidebar section.
func TestObservationView_MetadataSidebar(t *testing.T) {
	proj := "alpha"
	topicKey := "sdd/auth/spec"
	lastSeen := "2026-01-20"
	stub := &stubEngramClient{
		obsOut: &client.Observation{
			ID:             55,
			Title:          "Meta Test Obs",
			Content:        "body content",
			Type:           "decision",
			Scope:          "project",
			CreatedAt:      "2026-01-15",
			Project:        &proj,
			TopicKey:       &topicKey,
			SessionID:      "my-session-123",
			RevisionCount:  3,
			DuplicateCount: 2,
			LastSeenAt:     &lastSeen,
		},
	}
	s := newWithClient(stub)
	req := httptest.NewRequest(http.MethodGet, "/observations/55", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	// Locate the meta-sidebar.
	sidebarIdx := strings.Index(body, `class="meta-sidebar"`)
	if sidebarIdx == -1 {
		t.Fatal(`expected element with class="meta-sidebar" in observation detail body`)
	}
	// Extract from sidebar to end so we scope subsequent checks.
	sidebarRegion := body[sidebarIdx:]

	// All 9 fields must be present in the sidebar region.
	wantLabels := []string{
		"project", "scope", "type", "topic_key",
		"session", "created_at", "revisions", "duplicates", "last_seen_at",
	}
	for _, label := range wantLabels {
		// Check case-insensitive presence of the label text.
		if !strings.Contains(strings.ToLower(sidebarRegion), label) {
			t.Errorf("expected metadata label %q in .meta-sidebar region", label)
		}
	}
}

// TestObservationView_BackLinkFromParam asserts the back-link href equals
// the decoded ?from= value when the from param is valid.
//
// Encoded from:  /p/alpha?topic_key_prefix=sdd%2Fauth%2F
// URL-encoded:   %2Fp%2Falpha%3Ftopic_key_prefix%3Dsdd%252Fauth%252F
//
// After net/http decodes the query string once:
//   decoded `from` value = /p/alpha?topic_key_prefix=sdd%2Fauth%2F
// validateFrom must accept this (starts with /, no .., length OK).
// The back-link href should equal that decoded value verbatim.
func TestObservationView_BackLinkFromParam(t *testing.T) {
	s := newWithClient(makeObsStub(42))
	// Build request with ?from= carrying a once-encoded path.
	// The `from` value (pre-encoding for inclusion in the URL) is:
	//   /p/alpha?topic_key_prefix=sdd%2Fauth%2F
	// Encoded for the outer query string:
	//   %2Fp%2Falpha%3Ftopic_key_prefix%3Dsdd%252Fauth%252F
	req := httptest.NewRequest(http.MethodGet,
		"/observations/42?from=%2Fp%2Falpha%3Ftopic_key_prefix%3Dsdd%252Fauth%252F",
		nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()
	// The href must equal the once-decoded value.
	want := `href="/p/alpha?topic_key_prefix=sdd%2Fauth%2F"`
	if !strings.Contains(body, want) {
		t.Errorf("expected back-link %q in body, got:\n%s", want, body)
	}
}

// --- Task 6.2 ---

// TestObservationView_MarkdownContainer asserts that when ?raw= is NOT set,
// the markdown output is wrapped in a .markdown container.
func TestObservationView_MarkdownContainer(t *testing.T) {
	stub := &stubEngramClient{
		obsOut: &client.Observation{
			ID:        99,
			Title:     "Markdown Test",
			Content:   "## Hello\n\nSome **bold** text.",
			Type:      "decision",
			Scope:     "project",
			CreatedAt: "2026-01-01",
		},
	}
	s := newWithClient(stub)
	req := httptest.NewRequest(http.MethodGet, "/observations/99", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()
	// The rendered markdown must be inside a .markdown wrapper.
	if !strings.Contains(body, `class="markdown"`) {
		t.Error(`expected element with class="markdown" wrapping rendered markdown output`)
	}
}

// TestObservationView_DetailHeadInlineMeta asserts the detail page renders a
// .detail-head__row with the type badge + #id + formatted date + time-ago.
// Optional "rev N" appears only when RevisionCount > 0.
func TestObservationView_DetailHeadInlineMeta(t *testing.T) {
	stub := &stubEngramClient{
		obsOut: &client.Observation{
			ID:            77,
			Title:         "Head Meta Test",
			Content:       "body",
			Type:          "decision",
			Scope:         "project",
			CreatedAt:     "2026-03-15",
			RevisionCount: 2,
		},
	}
	s := newWithClient(stub)
	req := httptest.NewRequest(http.MethodGet, "/observations/77", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	if !strings.Contains(body, `class="detail-head__row"`) {
		t.Error(`expected .detail-head__row container in detail head`)
	}
	// #id rendered in the inline row.
	if !strings.Contains(body, `class="detail-head__id mono"`) || !strings.Contains(body, "#77") {
		t.Error("expected #77 in .detail-head__id")
	}
	// Formatted date present (FormatDate("2026-03-15") -> "Mar 15, 2026").
	if !strings.Contains(body, "Mar 15, 2026") {
		t.Error(`expected "Mar 15, 2026" in detail head row`)
	}
	// rev N appears because RevisionCount=2.
	if !strings.Contains(body, "rev 2") {
		t.Error(`expected "rev 2" indicator in detail head when RevisionCount > 0`)
	}
}

// TestObservationView_DetailHeadOmitsRevWhenZero asserts the "rev N" segment
// is dropped when RevisionCount == 0.
func TestObservationView_DetailHeadOmitsRevWhenZero(t *testing.T) {
	stub := &stubEngramClient{
		obsOut: &client.Observation{
			ID:            78,
			Title:         "No Revisions",
			Content:       "body",
			Type:          "discovery",
			Scope:         "project",
			CreatedAt:     "2026-03-15",
			RevisionCount: 0,
		},
	}
	s := newWithClient(stub)
	req := httptest.NewRequest(http.MethodGet, "/observations/78", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	if strings.Contains(body, "rev 0") || strings.Contains(body, ">rev ") {
		t.Error(`"rev N" segment must NOT render when RevisionCount == 0`)
	}
}

// TestObservationView_ViewToggleHasBothTabs asserts the view toggle renders
// both Rendered and Raw tabs at all times. The current state is marked
// aria-current="page" and the other state is a link to switch.
func TestObservationView_ViewToggleHasBothTabs(t *testing.T) {
	stub := &stubEngramClient{
		obsOut: &client.Observation{
			ID:        12,
			Title:     "Toggle Test",
			Content:   "## Hello",
			Type:      "decision",
			Scope:     "project",
			CreatedAt: "2026-01-01",
		},
	}
	s := newWithClient(stub)

	t.Run("rendered mode", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/observations/12", nil)
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		body := rr.Body.String()

		if !strings.Contains(body, "Rendered") {
			t.Error(`expected "Rendered" tab label`)
		}
		if !strings.Contains(body, "Raw") {
			t.Error(`expected "Raw" tab label`)
		}
		if !strings.Contains(body, `aria-current="page"`) {
			t.Error("expected aria-current=\"page\" on the active tab")
		}
		// Raw must be the navigable one (anchor with href to ?raw=1).
		if !strings.Contains(body, `href="/observations/12?raw=1"`) {
			t.Error(`expected Raw tab to link to ?raw=1`)
		}
	})

	t.Run("raw mode", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/observations/12?raw=1", nil)
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		body := rr.Body.String()

		// In raw mode, the Rendered tab becomes the navigable anchor.
		if !strings.Contains(body, `href="/observations/12"`) {
			t.Error(`expected Rendered tab to link to /observations/12 (no raw param)`)
		}
	})
}

// TestObservationView_SidebarShowsIDAndToolName asserts the new sidebar rows
// (id always, tool_name when present) appear in the meta-card region.
func TestObservationView_SidebarShowsIDAndToolName(t *testing.T) {
	tool := "claude-code"
	stub := &stubEngramClient{
		obsOut: &client.Observation{
			ID:        88,
			Title:     "Sidebar Extras",
			Content:   "body",
			Type:      "decision",
			Scope:     "project",
			CreatedAt: "2026-01-01",
			ToolName:  &tool,
		},
	}
	s := newWithClient(stub)
	req := httptest.NewRequest(http.MethodGet, "/observations/88", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()
	sidebarIdx := strings.Index(body, `class="meta-sidebar"`)
	if sidebarIdx == -1 {
		t.Fatal(`expected .meta-sidebar in body`)
	}
	sidebar := strings.ToLower(body[sidebarIdx:])

	// id row with #88 inside the sidebar
	if !strings.Contains(sidebar, "#88") {
		t.Error("expected #88 in sidebar id row")
	}
	// tool_name label + value
	if !strings.Contains(sidebar, "tool_name") {
		t.Error(`expected "tool_name" label in sidebar`)
	}
	if !strings.Contains(sidebar, "claude-code") {
		t.Error("expected tool_name value 'claude-code' in sidebar")
	}
}

// TestObservationView_SidebarHidesToolNameWhenAbsent asserts the tool_name
// row is omitted entirely when the observation has no ToolName.
func TestObservationView_SidebarHidesToolNameWhenAbsent(t *testing.T) {
	stub := &stubEngramClient{
		obsOut: &client.Observation{
			ID:        89,
			Title:     "No Tool",
			Content:   "body",
			Type:      "decision",
			Scope:     "project",
			CreatedAt: "2026-01-01",
		},
	}
	s := newWithClient(stub)
	req := httptest.NewRequest(http.MethodGet, "/observations/89", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := strings.ToLower(rr.Body.String())
	if strings.Contains(body, "tool_name") {
		t.Error("tool_name row must NOT render when ToolName is nil/empty")
	}
}

// TestObservationView_RawBypass asserts that when ?raw=1 is set, the body is
// rendered as raw text without a .markdown wrapper.
func TestObservationView_RawBypass(t *testing.T) {
	stub := &stubEngramClient{
		obsOut: &client.Observation{
			ID:        99,
			Title:     "Raw Test",
			Content:   "## Hello\n\nSome **bold** text.",
			Type:      "decision",
			Scope:     "project",
			CreatedAt: "2026-01-01",
		},
	}
	s := newWithClient(stub)
	req := httptest.NewRequest(http.MethodGet, "/observations/99?raw=1", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()
	// No .markdown wrapper must be present in raw mode.
	if strings.Contains(body, `class="markdown"`) {
		t.Error(`class="markdown" must NOT be present when ?raw=1 is set`)
	}
	// Raw content must still be present (unrendered).
	if !strings.Contains(body, "## Hello") {
		t.Error("expected raw markdown source in body when ?raw=1 is set")
	}
}
