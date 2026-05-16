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
