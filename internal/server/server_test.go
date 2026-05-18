package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/NotFoundSN/engram-ui/internal/client"
)

// stubEngramClient implements engramClient for testing.
type stubEngramClient struct {
	statsOut  *client.Stats
	statsErr  error
	recentOut []client.Observation
	recentErr error
	// recentByProject allows different results per project name
	recentByProject    map[string][]client.Observation
	recentByProjectErr map[string]error
	searchOut          []client.SearchResult
	searchErr          error
	searchCalled       bool
	recentCalled       bool
	obsOut             *client.Observation
	obsErr             error
}

func (s *stubEngramClient) Stats() (*client.Stats, error) {
	return s.statsOut, s.statsErr
}

func (s *stubEngramClient) RecentObservations(opts client.RecentOptions) ([]client.Observation, error) {
	s.recentCalled = true
	if s.recentByProject != nil {
		if err, ok := s.recentByProjectErr[opts.Project]; ok {
			return nil, err
		}
		if obs, ok := s.recentByProject[opts.Project]; ok {
			return obs, nil
		}
		return nil, nil
	}
	return s.recentOut, s.recentErr
}

func (s *stubEngramClient) Search(q string, opts client.SearchOptions) ([]client.SearchResult, error) {
	s.searchCalled = true
	return s.searchOut, s.searchErr
}

func (s *stubEngramClient) Observation(id int64) (*client.Observation, error) {
	return s.obsOut, s.obsErr
}

func (s *stubEngramClient) Health() error { return nil }

// --- handleHome tests ---

func TestHandleHome_ProjectGrid(t *testing.T) {
	// Three projects: alpha has session_summary, beta and gamma do not.
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"alpha", "beta", "gamma"}},
		recentByProject: map[string][]client.Observation{
			"alpha": {{ID: 1, Type: "session_summary", Title: "Session for alpha", Content: "Alpha session content", CreatedAt: "2026-01-03"}},
			"beta":  {},
			"gamma": {},
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

	// Each project must have a card
	for _, proj := range []string{"alpha", "beta", "gamma"} {
		if !strings.Contains(body, proj) {
			t.Errorf("expected project %q in response body", proj)
		}
	}

	// alpha has session content; beta and gamma have placeholder
	if !strings.Contains(body, "Alpha session content") {
		t.Error("expected alpha session preview in body")
	}
	if strings.Count(body, "No session yet") < 2 {
		t.Error("expected 'No session yet' placeholder for beta and gamma")
	}
}

func TestHandleHome_EmptyState(t *testing.T) {
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{}},
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "No projects yet") {
		t.Error("expected empty-state message in body")
	}
}

func TestHandleHome_PreviewTruncated(t *testing.T) {
	// Build a content string > 140 runes
	longContent := strings.Repeat("a", 200)
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"alpha"}},
		recentByProject: map[string][]client.Observation{
			"alpha": {{ID: 1, Type: "session_summary", Title: "T", Content: longContent, CreatedAt: "2026-01-01"}},
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

	// Find the preview in the body; it should not contain more than 140 'a' chars in sequence
	idx := strings.Index(body, strings.Repeat("a", 140))
	if idx == -1 {
		t.Error("expected exactly 140 'a' chars in body (truncated preview)")
	}
	if strings.Contains(body, strings.Repeat("a", 141)) {
		t.Error("preview should not contain 141+ consecutive 'a' chars (should be truncated)")
	}
}

func TestHandleHome_PerProjectFetchError_GracefulDegradation(t *testing.T) {
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"good", "bad"}},
		recentByProject: map[string][]client.Observation{
			"good": {{ID: 1, Type: "session_summary", Title: "Good session", Content: "preview", CreatedAt: "2026-01-01"}},
		},
		recentByProjectErr: map[string]error{
			"bad": errors.New("connection refused"),
		},
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	// Page must NOT return an error — graceful degradation
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 even with per-project error, got %d", rr.Code)
	}
	body := rr.Body.String()

	// Both project names must appear
	if !strings.Contains(body, "good") {
		t.Error("expected 'good' project in body")
	}
	if !strings.Contains(body, "bad") {
		t.Error("expected 'bad' project in body even when fetch fails")
	}

	// bad project should show placeholder
	if !strings.Contains(body, "No session yet") {
		t.Error("expected 'No session yet' placeholder for bad project")
	}
}

// --- handleProject tests (no q) ---

func TestHandleProject_DefaultRender(t *testing.T) {
	obs := []client.Observation{
		{ID: 10, Type: "decision", Title: "First", Content: "c1", CreatedAt: "2026-01-05"},
		{ID: 11, Type: "bugfix", Title: "Second", Content: "c2", CreatedAt: "2026-01-04"},
		{ID: 12, Type: "decision", Title: "Third", Content: "c3", CreatedAt: "2026-01-03"},
		{ID: 13, Type: "bugfix", Title: "Fourth", Content: "c4", CreatedAt: "2026-01-02"},
		{ID: 14, Type: "plan", Title: "Fifth", Content: "c5", CreatedAt: "2026-01-01"},
	}
	stub := &stubEngramClient{
		statsOut:  &client.Stats{Projects: []string{"alpha"}},
		recentOut: obs,
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()

	// All 5 observations must appear
	for _, title := range []string{"First", "Second", "Third", "Fourth", "Fifth"} {
		if !strings.Contains(body, title) {
			t.Errorf("expected observation %q in body", title)
		}
	}

	// Most recent (2026-01-05) must appear before oldest (2026-01-01) — date_desc default
	posFirst := strings.Index(body, "First")
	posFifth := strings.Index(body, "Fifth")
	if posFirst == -1 || posFifth == -1 {
		t.Fatal("could not find First or Fifth in body")
	}
	if posFirst > posFifth {
		t.Error("most recent observation (First, 2026-01-05) should appear before oldest (Fifth, 2026-01-01)")
	}
}

func TestHandleProject_TypeFilter(t *testing.T) {
	obs := []client.Observation{
		{ID: 1, Type: "decision", Title: "D1", Content: "c", CreatedAt: "2026-01-03"},
		{ID: 2, Type: "bugfix", Title: "B1", Content: "c", CreatedAt: "2026-01-02"},
		{ID: 3, Type: "decision", Title: "D2", Content: "c", CreatedAt: "2026-01-01"},
	}
	stub := &stubEngramClient{
		statsOut:  &client.Stats{Projects: []string{"alpha"}},
		recentOut: obs,
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/p/alpha?type=decision", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()

	if !strings.Contains(body, "D1") {
		t.Error("expected D1 in filtered body")
	}
	if !strings.Contains(body, "D2") {
		t.Error("expected D2 in filtered body")
	}
	if strings.Contains(body, "B1") {
		t.Error("B1 (bugfix) should be filtered out")
	}
}

func TestHandleProject_SortAscending(t *testing.T) {
	obs := []client.Observation{
		{ID: 1, Type: "decision", Title: "Newer", Content: "c", CreatedAt: "2026-01-02"},
		{ID: 2, Type: "decision", Title: "Older", Content: "c", CreatedAt: "2026-01-01"},
	}
	stub := &stubEngramClient{
		statsOut:  &client.Stats{Projects: []string{"alpha"}},
		recentOut: obs,
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/p/alpha?sort=date_asc", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()

	posOlder := strings.Index(body, "Older")
	posNewer := strings.Index(body, "Newer")
	if posOlder == -1 || posNewer == -1 {
		t.Fatal("could not find Older or Newer in body")
	}
	if posOlder > posNewer {
		t.Error("date_asc: oldest (Older, 2026-01-01) should appear before newest (Newer, 2026-01-02)")
	}
}

func TestHandleProject_EmptyProject(t *testing.T) {
	stub := &stubEngramClient{
		statsOut:  &client.Stats{Projects: []string{}},
		recentOut: []client.Observation{},
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/p/does-not-exist", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for unknown project, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "No observations to show") {
		t.Error("expected empty-state message for project with zero observations")
	}
}

func TestHandleProject_ObservationDeepLink(t *testing.T) {
	// v2: observation rows now carry ?from= with the project source URL.
	// GET /p/alpha (no params) → from=%2Fp%2Falpha
	obs := []client.Observation{
		{ID: 42, Type: "decision", Title: "The Answer", Content: "c", CreatedAt: "2026-01-01"},
	}
	stub := &stubEngramClient{
		statsOut:  &client.Stats{Projects: []string{"alpha"}},
		recentOut: obs,
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	// v2: rows carry ?from= with the encoded project URL
	want := `href="/observations/42?from=%2Fp%2Falpha"`
	if !strings.Contains(body, want) {
		t.Errorf("expected %q in body", want)
	}
}

func TestHandleProject_SortGarbage(t *testing.T) {
	obs := []client.Observation{
		{ID: 1, Type: "decision", Title: "Only", Content: "c", CreatedAt: "2026-01-01"},
	}
	stub := &stubEngramClient{
		statsOut:  &client.Stats{Projects: []string{"alpha"}},
		recentOut: obs,
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/p/alpha?sort=garbage", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	// Must not 500 — unknown sort value defaults to date_desc
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for unknown sort value, got %d", rr.Code)
	}
}

// --- handleProject tests (with q) ---

func TestHandleProject_TopicKeyPrefixFilter(t *testing.T) {
	// ?topic_key_prefix=sdd/auth/ should narrow the list to observations
	// whose topic_key starts with that prefix. Observations without a
	// matching prefix (or no topic_key at all) are filtered out.
	k1, k2, k3 := "sdd/auth/spec", "sdd/auth/design", "decision/cookie-vs-localstorage"
	obs := []client.Observation{
		{ID: 1, Type: "spec", Title: "Auth Spec", Content: "c", CreatedAt: "2026-01-01", TopicKey: &k1},
		{ID: 2, Type: "design", Title: "Auth Design", Content: "c", CreatedAt: "2026-01-02", TopicKey: &k2},
		{ID: 3, Type: "decision", Title: "Other Decision", Content: "c", CreatedAt: "2026-01-03", TopicKey: &k3},
		{ID: 4, Type: "discovery", Title: "No Key Obs", Content: "c", CreatedAt: "2026-01-04"},
	}
	stub := &stubEngramClient{
		statsOut:  &client.Stats{Projects: []string{"alpha"}},
		recentOut: obs,
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/p/alpha?topic_key_prefix=sdd/auth/", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()

	if !strings.Contains(body, "Auth Spec") {
		t.Error("expected 'Auth Spec' in filtered body (matches prefix)")
	}
	if !strings.Contains(body, "Auth Design") {
		t.Error("expected 'Auth Design' in filtered body (matches prefix)")
	}
	if strings.Contains(body, "Other Decision") {
		t.Error("'Other Decision' must be filtered out (different prefix)")
	}
	if strings.Contains(body, "No Key Obs") {
		t.Error("observation without topic_key must be filtered out when prefix is active")
	}
}

func TestHandleProject_TopicKeyPrefixPreservedInForms(t *testing.T) {
	// When ?topic_key_prefix= is active, every filter form on the page must
	// carry it as a hidden input so submitting one form does not drop the
	// active prefix filter. There are 3 forms: search, type, and sort.
	k := "sdd/auth/spec"
	obs := []client.Observation{
		{ID: 1, Type: "spec", Title: "Auth Spec", Content: "c", CreatedAt: "2026-01-01", TopicKey: &k},
	}
	stub := &stubEngramClient{
		statsOut:  &client.Stats{Projects: []string{"alpha"}},
		recentOut: obs,
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/p/alpha?topic_key_prefix=sdd/auth/", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()

	// Hidden inputs must appear in all 3 forms (search, type, sort).
	needle := `type="hidden" name="topic_key_prefix" value="sdd/auth/"`
	count := strings.Count(body, needle)
	if count < 3 {
		t.Errorf("expected hidden topic_key_prefix input in 3 forms (search, type, sort); found %d occurrences of %q", count, needle)
	}
}

func TestHandleProject_ActivePrefixChipRendered(t *testing.T) {
	// When ?topic_key_prefix= is active, the page shows a visible chip
	// containing the prefix and a clear link that drops it.
	k := "sdd/auth/spec"
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"alpha"}},
		recentOut: []client.Observation{
			{ID: 10, Type: "spec", Title: "Auth Spec", Content: "c", CreatedAt: "2026-01-01", TopicKey: &k},
		},
	}
	s := newWithClient(stub)
	req := httptest.NewRequest(http.MethodGet, "/p/alpha?topic_key_prefix=sdd/auth/", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	// The chip element must contain the prefix value visible to the user.
	if !strings.Contains(body, `id="prefix-chip"`) {
		t.Error(`expected id="prefix-chip" in body when ?topic_key_prefix= is active`)
	}
	if !strings.Contains(body, "sdd/auth/") {
		t.Error("expected prefix value 'sdd/auth/' visible in chip body")
	}

	// Clear link must point to /p/alpha (no prefix). Use chip id as anchor.
	chipIdx := strings.Index(body, `id="prefix-chip"`)
	if chipIdx == -1 {
		t.Fatal("could not locate chip in body")
	}
	// Find next 200 chars after chip to look for clear link
	tail := body[chipIdx:]
	if !strings.Contains(tail[:min(len(tail), 500)], `href="/p/alpha"`) {
		t.Error("expected clear link href=/p/alpha inside or near the chip")
	}
}

func TestHandleProject_NoChipWhenPrefixInactive(t *testing.T) {
	// Without ?topic_key_prefix=, no chip should render.
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"alpha"}},
		recentOut: []client.Observation{
			{ID: 1, Type: "decision", Title: "Some Dec", Content: "c", CreatedAt: "2026-01-01"},
		},
	}
	s := newWithClient(stub)
	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()
	if strings.Contains(body, `id="prefix-chip"`) {
		t.Error(`prefix-chip must NOT render when topic_key_prefix is empty`)
	}
}

func TestHandleProject_RowsHaveFromWithTopicKeyPrefix(t *testing.T) {
	// Observation row hrefs must embed ?from= URL that includes the active
	// topic_key_prefix, so back-link from detail preserves the filter view.
	k := "sdd/auth/spec"
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"alpha"}},
		recentOut: []client.Observation{
			{ID: 10, Type: "spec", Title: "Auth Spec", Content: "c", CreatedAt: "2026-01-01", TopicKey: &k},
		},
	}
	s := newWithClient(stub)
	req := httptest.NewRequest(http.MethodGet, "/p/alpha?topic_key_prefix=sdd/auth/", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()
	// /p/alpha?topic_key_prefix=sdd/auth/ → encoded = %2Fp%2Falpha%3Ftopic_key_prefix%3Dsdd%252Fauth%252F
	// Detail link should contain ?from= with the encoded source URL.
	want := "topic_key_prefix"
	if !strings.Contains(body, want) {
		t.Errorf("expected %q in ?from= attribute of row hrefs", want)
	}
}

func TestHandleProject_SearchPath(t *testing.T) {
	searchResults := []client.SearchResult{
		{Observation: client.Observation{ID: 5, Type: "decision", Title: "Auth flow", Content: "c", CreatedAt: "2026-01-01"}, Rank: 1.0},
		{Observation: client.Observation{ID: 6, Type: "bugfix", Title: "Auth fix", Content: "c", CreatedAt: "2026-01-02"}, Rank: 0.5},
	}
	stub := &stubEngramClient{
		statsOut:  &client.Stats{Projects: []string{"alpha"}},
		searchOut: searchResults,
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/p/alpha?q=auth", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()

	// Search must be called; RecentObservations must NOT be called
	if !stub.searchCalled {
		t.Error("expected Search() to be called when ?q= is set")
	}
	if stub.recentCalled {
		t.Error("expected RecentObservations() NOT to be called when ?q= is set")
	}

	// Search results must appear in returned order (rank order)
	if !strings.Contains(body, "Auth flow") {
		t.Error("expected 'Auth flow' in search results")
	}
	if !strings.Contains(body, "Auth fix") {
		t.Error("expected 'Auth fix' in search results")
	}
	posFlow := strings.Index(body, "Auth flow")
	posFix := strings.Index(body, "Auth fix")
	if posFlow > posFix {
		t.Error("search results should appear in rank order (Auth flow before Auth fix)")
	}
}

func TestHandleProject_SearchSortIgnored(t *testing.T) {
	searchResults := []client.SearchResult{
		{Observation: client.Observation{ID: 7, Type: "decision", Title: "Result A", Content: "c", CreatedAt: "2026-01-01"}, Rank: 1.0},
		{Observation: client.Observation{ID: 8, Type: "bugfix", Title: "Result B", Content: "c", CreatedAt: "2026-01-02"}, Rank: 0.5},
	}
	stub := &stubEngramClient{
		statsOut:  &client.Stats{Projects: []string{"alpha"}},
		searchOut: searchResults,
	}
	s := newWithClient(stub)

	// sort=date_asc is set, but q is also set — sort should be ignored
	req := httptest.NewRequest(http.MethodGet, "/p/alpha?q=auth&sort=date_asc", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()

	// Results in search rank order (A before B)
	posA := strings.Index(body, "Result A")
	posB := strings.Index(body, "Result B")
	if posA == -1 || posB == -1 {
		t.Fatal("could not find Result A or Result B in body")
	}
	if posA > posB {
		t.Error("search results should be in rank order (Result A before Result B), sort param ignored")
	}
}

// --- unionTypes tests ---

func TestUnionTypes(t *testing.T) {
	canonical14 := []string{
		"architecture", "bugfix", "config", "decision", "design",
		"discovery", "exploration", "pattern", "plan", "preference",
		"proposal", "report", "spec", "tasks",
	}

	cases := []struct {
		name      string
		canonical []string
		present   []string
		active    string
		wantLen   int
		wantItems []string
	}{
		{
			name:      "canonical only, no present, no active",
			canonical: canonical14,
			present:   nil,
			active:    "",
			wantLen:   14,
			wantItems: []string{"architecture", "tasks"}, // first and last alphabetically
		},
		{
			name:      "canonical plus extra present type",
			canonical: canonical14,
			present:   []string{"custom-internal"},
			active:    "",
			wantLen:   15,
			wantItems: []string{"custom-internal", "decision"},
		},
		{
			name:      "canonical plus active phantom",
			canonical: canonical14,
			present:   nil,
			active:    "phantom-type",
			wantLen:   15,
			wantItems: []string{"phantom-type", "architecture"},
		},
		{
			name:      "duplicate present already in canonical",
			canonical: canonical14,
			present:   []string{"decision"},
			active:    "",
			wantLen:   14,
			wantItems: []string{"decision"},
		},
		{
			name:      "empty present slice",
			canonical: canonical14,
			present:   []string{},
			active:    "",
			wantLen:   14,
			wantItems: []string{"bugfix"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := unionTypes(tc.canonical, tc.present, tc.active)
			if len(got) != tc.wantLen {
				t.Errorf("len(unionTypes(...)) = %d; want %d; got %v", len(got), tc.wantLen, got)
			}
			for _, item := range tc.wantItems {
				found := false
				for _, g := range got {
					if g == item {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected %q in result %v", item, got)
				}
			}
			// Verify sorted
			for i := 1; i < len(got); i++ {
				if got[i] < got[i-1] {
					t.Errorf("result not sorted: got[%d]=%q < got[%d]=%q", i, got[i], i-1, got[i-1])
				}
			}
		})
	}
}

// --- handleProject v2 tests (FR-4, FR-6, Scenarios 1-6) ---

// Shared setup: project "alpha" with observations of types decision and custom-internal.
func alphaStub() *stubEngramClient {
	return &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"alpha"}},
		recentOut: []client.Observation{
			{ID: 10, Type: "decision", Title: "Dec1", Content: "c", CreatedAt: "2026-01-03"},
			{ID: 11, Type: "custom-internal", Title: "Cust1", Content: "c", CreatedAt: "2026-01-02"},
		},
	}
}

func TestHandleProject_TypeSelectIncludesAllCanonical(t *testing.T) {
	// Type filter is a <select> with one option per canonical type plus any
	// project-present types and a default "All types" option.
	s := newWithClient(alphaStub())
	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()

	if !strings.Contains(body, `<select id="type" name="type"`) {
		t.Error(`expected <select id="type" name="type"> in body`)
	}

	canonicalTypes := []string{
		"architecture", "bugfix", "config", "decision", "design",
		"discovery", "exploration", "pattern", "plan", "preference",
		"proposal", "report", "spec", "tasks",
	}
	for _, typ := range canonicalTypes {
		if !strings.Contains(body, `<option value="`+typ+`"`) {
			t.Errorf("expected <option value=%q> in the type select", typ)
		}
	}

	if !strings.Contains(body, `<option value="custom-internal"`) {
		t.Error(`expected <option value="custom-internal"> in the type select`)
	}

	if !strings.Contains(body, `<option value="" selected`) {
		t.Error(`expected default "All types" option (value="") to be selected`)
	}
}

func TestHandleProject_TypeSelectMarksActiveSelected(t *testing.T) {
	s := newWithClient(alphaStub())
	req := httptest.NewRequest(http.MethodGet, "/p/alpha?type=decision", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()

	if !strings.Contains(body, `<option value="decision" selected`) {
		t.Error(`expected <option value="decision" selected> when ?type=decision`)
	}
}

func TestHandleProject_TypeSelectMarksAllTypesSelectedWhenNoFilter(t *testing.T) {
	s := newWithClient(alphaStub())
	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	if !strings.Contains(body, `<option value="" selected`) {
		t.Error(`expected default "All types" option (value="") to be selected when no ?type= filter is set`)
	}
}

func TestHandleProject_TypeSelectIncludesPhantom(t *testing.T) {
	// Phantom type — not present in the project's observations — still
	// appears as a selected option when ?type= names it.
	s := newWithClient(alphaStub())
	req := httptest.NewRequest(http.MethodGet, "/p/alpha?type=disappeared", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	if !strings.Contains(body, `<option value="disappeared" selected`) {
		t.Error(`expected phantom <option value="disappeared" selected> in type select`)
	}
}

func TestHandleProject_TypeSelectFormHasHiddenQAndSort(t *testing.T) {
	// q and sort survive a type-select submission via hidden inputs in the
	// type form (one per other active filter).
	s := newWithClient(alphaStub())
	req := httptest.NewRequest(http.MethodGet, "/p/alpha?q=auth&sort=date_asc", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	if !strings.Contains(body, `<select id="type"`) {
		t.Fatal(`expected <select id="type"> in body`)
	}
	if !strings.Contains(body, `<input type="hidden" name="q" value="auth"`) {
		t.Error(`expected hidden q=auth input in the type filter form`)
	}
	if !strings.Contains(body, `<input type="hidden" name="sort" value="date_asc"`) {
		t.Error(`expected hidden sort=date_asc input in the type filter form`)
	}
}

func TestHandleProject_TypeSelectFormOmitsSortWhenImplicit(t *testing.T) {
	// When sort is not explicitly set (defaults to date_desc), the type
	// filter form must NOT carry a hidden sort input — submitting it would
	// otherwise pin the default sort into the URL unnecessarily.
	s := newWithClient(alphaStub())
	req := httptest.NewRequest(http.MethodGet, "/p/alpha?type=decision", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	selectIdx := strings.Index(body, `<select id="type"`)
	if selectIdx == -1 {
		t.Fatal(`expected <select id="type"> in body`)
	}
	formStart := strings.LastIndex(body[:selectIdx], "<form")
	if formStart == -1 {
		t.Fatal("could not find <form before type select")
	}
	formEnd := strings.Index(body[selectIdx:], "</form>")
	if formEnd == -1 {
		t.Fatal("could not find </form> after type select")
	}
	typeFilterForm := body[formStart : selectIdx+formEnd+len("</form>")]

	if strings.Contains(typeFilterForm, `type="hidden" name="sort"`) {
		t.Error(`type filter form must not contain hidden sort input when sort was not explicitly set`)
	}
}

func TestHandleProject_TypeSelectFormHasApplyButton(t *testing.T) {
	// Every filter form keeps a submit button (visually hidden via sr-only)
	// so JS-disabled clients can still submit the form by pressing Enter.
	s := newWithClient(alphaStub())
	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	if !strings.Contains(body, `type="submit"`) || !strings.Contains(body, "Apply") {
		t.Error(`expected <button type="submit">Apply</button> in body`)
	}
}

func TestHandleProject_RowsHaveFromWithFilters(t *testing.T) {
	// Scenario 5: GET /p/alpha?type=decision&sort=date_asc
	// Observation id=10 row href must be /observations/10?from=%2Fp%2Falpha%3Ftype%3Ddecision%26sort%3Ddate_asc
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"alpha"}},
		recentOut: []client.Observation{
			{ID: 10, Type: "decision", Title: "Dec1", Content: "c", CreatedAt: "2026-01-03"},
		},
	}
	s := newWithClient(stub)
	req := httptest.NewRequest(http.MethodGet, "/p/alpha?type=decision&sort=date_asc", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	want := `href="/observations/10?from=%2Fp%2Falpha%3Ftype%3Ddecision%26sort%3Ddate_asc"`
	if !strings.Contains(body, want) {
		t.Errorf("expected %q in body, got:\n%s", want, body)
	}
}

func TestHandleProject_RowsHaveFromWithNoFilters(t *testing.T) {
	// Scenario 6: GET /p/alpha (no params)
	// row href must be /observations/42?from=%2Fp%2Falpha
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"alpha"}},
		recentOut: []client.Observation{
			{ID: 42, Type: "decision", Title: "The Answer", Content: "c", CreatedAt: "2026-01-01"},
		},
	}
	s := newWithClient(stub)
	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	want := `href="/observations/42?from=%2Fp%2Falpha"`
	if !strings.Contains(body, want) {
		t.Errorf("expected %q in body, got:\n%s", want, body)
	}
}

func TestHandleProject_RowsHaveFromWithExplicitDefaultSort(t *testing.T) {
	// GET /p/alpha?sort=date_desc — explicit, so from= must include sort
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"alpha"}},
		recentOut: []client.Observation{
			{ID: 7, Type: "decision", Title: "D", Content: "c", CreatedAt: "2026-01-01"},
		},
	}
	s := newWithClient(stub)
	req := httptest.NewRequest(http.MethodGet, "/p/alpha?sort=date_desc", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	want := `href="/observations/7?from=%2Fp%2Falpha%3Fsort%3Ddate_desc"`
	if !strings.Contains(body, want) {
		t.Errorf("expected %q in body, got:\n%s", want, body)
	}
}


// --- handleObservation v2 tests (FR-5, Scenarios 7-11) ---

func makeObsStub(id int64) *stubEngramClient {
	topicKey := "test/topic"
	return &stubEngramClient{
		obsOut: &client.Observation{
			ID:        id,
			Title:     "Test Obs",
			Content:   "content",
			Type:      "decision",
			Scope:     "project",
			CreatedAt: "2026-01-01",
			TopicKey:  &topicKey,
		},
	}
}

func TestHandleObservation_RendersOK(t *testing.T) {
	s := newWithClient(makeObsStub(42))
	req := httptest.NewRequest(http.MethodGet, "/observations/42", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Test Obs") {
		t.Error("expected observation title in body")
	}
}

// siblingsStub builds an engramClient stub where Observation(id) returns the
// matching one in the observations slice (by ID), and RecentObservations
// returns all of them. Useful for testing the siblings fetch path.
type siblingsStub struct {
	stubEngramClient
	observations []client.Observation
}

func (s *siblingsStub) Observation(id int64) (*client.Observation, error) {
	for i := range s.observations {
		if s.observations[i].ID == id {
			return &s.observations[i], nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (s *siblingsStub) RecentObservations(opts client.RecentOptions) ([]client.Observation, error) {
	return s.observations, nil
}

func TestHandleObservation_SiblingsRenderedWhenTopicKeyHasTwoSlashes(t *testing.T) {
	// Current obs is sdd/auth/spec — has 2 slashes → prefix = sdd/auth/
	// Three sibling-rule candidates: spec, design, tasks → all should appear.
	// Plus one unrelated obs (decision/X) → must NOT appear.
	proj := "alpha"
	k1, k2, k3, k4 := "sdd/auth/spec", "sdd/auth/design", "sdd/auth/tasks", "decision/x"
	obs := []client.Observation{
		{ID: 10, Type: "spec", Title: "Auth Spec", Content: "c", CreatedAt: "2026-01-01", TopicKey: &k1, Project: &proj},
		{ID: 11, Type: "design", Title: "Auth Design", Content: "c", CreatedAt: "2026-01-02", TopicKey: &k2, Project: &proj},
		{ID: 12, Type: "tasks", Title: "Auth Tasks", Content: "c", CreatedAt: "2026-01-03", TopicKey: &k3, Project: &proj},
		{ID: 99, Type: "decision", Title: "Other Decision", Content: "c", CreatedAt: "2026-01-04", TopicKey: &k4, Project: &proj},
	}
	stub := &siblingsStub{observations: obs}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/observations/10", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()

	// Siblings section must be present.
	if !strings.Contains(body, `id="siblings-section"`) {
		t.Error(`expected id="siblings-section" in body when topic_key has ≥2 slashes`)
	}
	// All three sibling titles must appear.
	for _, want := range []string{"Auth Spec", "Auth Design", "Auth Tasks"} {
		if !strings.Contains(body, want) {
			t.Errorf("expected sibling title %q in body", want)
		}
	}
	// Unrelated obs must NOT appear in the siblings section.
	// (We assert it's not in the body at all — it isn't the current obs either.)
	if strings.Contains(body, "Other Decision") {
		t.Error("'Other Decision' (different prefix) must NOT appear in siblings")
	}
}

func TestHandleObservation_NoSiblingsSectionWhenTopicKeyHasOneSlash(t *testing.T) {
	// Current obs is decision/cookies (1 slash) → no siblings section.
	proj := "alpha"
	k := "decision/cookies"
	obs := []client.Observation{
		{ID: 5, Type: "decision", Title: "Cookies Decision", Content: "c", CreatedAt: "2026-01-01", TopicKey: &k, Project: &proj},
	}
	stub := &siblingsStub{observations: obs}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/observations/5", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()
	if strings.Contains(body, `id="siblings-section"`) {
		t.Error("siblings section must NOT render when topic_key has <2 slashes")
	}
}

func TestHandleObservation_NoSiblingsSectionWhenTopicKeyNil(t *testing.T) {
	// Obs with no topic_key → no siblings section.
	proj := "alpha"
	obs := []client.Observation{
		{ID: 7, Type: "discovery", Title: "Nameless", Content: "c", CreatedAt: "2026-01-01", Project: &proj},
	}
	stub := &siblingsStub{observations: obs}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/observations/7", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()
	if strings.Contains(body, `id="siblings-section"`) {
		t.Error("siblings section must NOT render when topic_key is nil")
	}
}

func TestHandleObservation_SiblingsSortedByCreatedAtAsc(t *testing.T) {
	// proposal (date_1) → spec (date_2) → design (date_3) in chronological order.
	// Even when fetched in random order, the rendered list must be asc.
	proj := "alpha"
	k1, k2, k3 := "sdd/auth/proposal", "sdd/auth/spec", "sdd/auth/design"
	obs := []client.Observation{
		// Inserted out of order intentionally.
		{ID: 12, Type: "design", Title: "Auth Design", Content: "c", CreatedAt: "2026-01-03", TopicKey: &k3, Project: &proj},
		{ID: 10, Type: "proposal", Title: "Auth Proposal", Content: "c", CreatedAt: "2026-01-01", TopicKey: &k1, Project: &proj},
		{ID: 11, Type: "spec", Title: "Auth Spec", Content: "c", CreatedAt: "2026-01-02", TopicKey: &k2, Project: &proj},
	}
	stub := &siblingsStub{observations: obs}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/observations/11", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()
	// Scope the order check to the siblings section <ul> (the non-current
	// siblings list). The current obs is pinned to the top via <summary>
	// so its position in the section is not part of the asc check.
	sectionStart := strings.Index(body, `id="siblings-section"`)
	if sectionStart == -1 {
		t.Fatal(`expected id="siblings-section" in body`)
	}
	sectionEnd := strings.Index(body[sectionStart:], "</section>")
	if sectionEnd == -1 {
		t.Fatal("could not find </section> closing tag")
	}
	siblingsBlock := body[sectionStart : sectionStart+sectionEnd]

	// The non-current siblings inside the <ul> must be sorted asc:
	// proposal (2026-01-01) before design (2026-01-03).
	pProposal := strings.Index(siblingsBlock, "Auth Proposal")
	pDesign := strings.Index(siblingsBlock, "Auth Design")
	if pProposal == -1 || pDesign == -1 {
		t.Fatalf("missing non-current siblings within section: proposal=%d, design=%d", pProposal, pDesign)
	}
	if !(pProposal < pDesign) {
		t.Errorf("non-current siblings not sorted asc by created_at: proposal@%d, design@%d", pProposal, pDesign)
	}
}

func TestHandleObservation_CurrentObsMarkedInSiblings(t *testing.T) {
	// The current obs appears in its own siblings list, marked with a badge.
	proj := "alpha"
	k1, k2 := "sdd/auth/spec", "sdd/auth/design"
	obs := []client.Observation{
		{ID: 10, Type: "spec", Title: "Auth Spec", Content: "c", CreatedAt: "2026-01-01", TopicKey: &k1, Project: &proj},
		{ID: 11, Type: "design", Title: "Auth Design", Content: "c", CreatedAt: "2026-01-02", TopicKey: &k2, Project: &proj},
	}
	stub := &siblingsStub{observations: obs}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/observations/10", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()
	// The current obs must have a "current" badge.
	if !strings.Contains(body, "current") {
		t.Error(`expected "current" badge near current obs in siblings list`)
	}
}

func TestHandleObservation_SiblingsCollapsedByDefault(t *testing.T) {
	// Siblings section must be wrapped in a <details> element WITHOUT the
	// `open` attribute, so the list is collapsed by default. Only the
	// current row (in <summary>) is visible until the user clicks.
	proj := "alpha"
	k1, k2 := "sdd/auth/spec", "sdd/auth/design"
	obs := []client.Observation{
		{ID: 10, Type: "spec", Title: "Auth Spec", Content: "c", CreatedAt: "2026-01-01", TopicKey: &k1, Project: &proj},
		{ID: 11, Type: "design", Title: "Auth Design", Content: "c", CreatedAt: "2026-01-02", TopicKey: &k2, Project: &proj},
	}
	stub := &siblingsStub{observations: obs}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/observations/10", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	// Scope to the siblings section.
	sectionStart := strings.Index(body, `id="siblings-section"`)
	if sectionStart == -1 {
		t.Fatal(`expected id="siblings-section" in body`)
	}
	sectionEnd := strings.Index(body[sectionStart:], "</section>")
	siblingsBlock := body[sectionStart : sectionStart+sectionEnd]

	// Must contain <details> and <summary> elements.
	if !strings.Contains(siblingsBlock, "<details") {
		t.Error("expected <details> element inside siblings section")
	}
	if !strings.Contains(siblingsBlock, "<summary") {
		t.Error("expected <summary> element inside siblings section")
	}
	// The <details> tag must NOT carry the `open` attribute (collapsed by default).
	// Match any of: `<details>`, `<details class=`, etc., but reject `<details open`.
	if strings.Contains(siblingsBlock, "<details open") {
		t.Error("siblings <details> must NOT be open by default (collapsed expected)")
	}
}

func TestHandleObservation_SiblingsSectionRendersWhenOnlyCurrent(t *testing.T) {
	// Even when the current obs is the only thing under the prefix, we still
	// render the sib-card — pinned with just the current row — so the page
	// always has the topic-prefix context block. The expand toggle is omitted
	// because there's nothing else to expand.
	proj := "alpha"
	k := "sdd/lonely/spec"
	obs := []client.Observation{
		{ID: 20, Type: "spec", Title: "Lonely Spec", Content: "c", CreatedAt: "2026-01-01", TopicKey: &k, Project: &proj},
	}
	stub := &siblingsStub{observations: obs}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/observations/20", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	if !strings.Contains(body, `id="siblings-section"`) {
		t.Error(`expected id="siblings-section" to render even when only the current obs matches the prefix`)
	}
	if !strings.Contains(body, "Lonely Spec") {
		t.Error(`expected current obs title in the pinned sib-card`)
	}
	// "show N related" toggle MUST NOT render when there are zero non-current siblings.
	if strings.Contains(body, "show 0 related") {
		t.Error(`"show 0 related" toggle must NOT render when current is the only sibling`)
	}
}

func TestHandleObservation_TopicKeyMetaRowClickable(t *testing.T) {
	// In the detail view, topic_key with ≥1 slash must render as a clickable
	// link to the project view filtered by strip-last-segment prefix.
	proj := "alpha"
	k := "sdd/auth/spec"
	stub := &siblingsStub{
		observations: []client.Observation{
			{ID: 10, Type: "spec", Title: "Auth Spec", Content: "c", CreatedAt: "2026-01-01", TopicKey: &k, Project: &proj},
		},
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/observations/10", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()
	want := `href="/p/alpha?topic_key_prefix=sdd%2Fauth%2F"`
	if !strings.Contains(body, want) {
		t.Errorf("expected topic_key meta-row link %q in body", want)
	}
}

func TestHandleObservation_TopicKeyMetaRowNotClickableWhenNoSlash(t *testing.T) {
	// Single-segment topic_key (no slash) → no strip-last possible → plain text.
	proj := "alpha"
	k := "scratchpad"
	stub := &siblingsStub{
		observations: []client.Observation{
			{ID: 11, Type: "discovery", Title: "Scratch", Content: "c", CreatedAt: "2026-01-01", TopicKey: &k, Project: &proj},
		},
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/observations/11", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()
	if strings.Contains(body, `?topic_key_prefix=`) {
		t.Error("topic_key with no slash must NOT render as a prefix link")
	}
}

func TestHandleObservation_ShortAliasRoute(t *testing.T) {
	// /m/{id} is an additive alias for /observations/{id} (same handler).
	// Web links keep using the long form; the short form is for agent emission.
	s := newWithClient(makeObsStub(42))
	req := httptest.NewRequest(http.MethodGet, "/m/42", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 from /m/42 alias, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Test Obs") {
		t.Error("expected observation title in body via /m/ alias")
	}
}

func TestHandleObservation_BackPlumbedWhenFromValid(t *testing.T) {
	// ?from=%2Fp%2Falpha%3Ftype%3Ddecision → decoded: /p/alpha?type=decision
	s := newWithClient(makeObsStub(42))
	req := httptest.NewRequest(http.MethodGet, "/observations/42?from=%2Fp%2Falpha%3Ftype%3Ddecision", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, `href="/p/alpha?type=decision"`) {
		t.Errorf("expected back link href=/p/alpha?type=decision in body, got:\n%s", body)
	}
}

func TestHandleObservation_BackFallsBackWhenFromMissing(t *testing.T) {
	s := newWithClient(makeObsStub(42))
	req := httptest.NewRequest(http.MethodGet, "/observations/42", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, `href="/"`) {
		t.Errorf("expected back link href=/ in body, got:\n%s", body)
	}
}

func TestHandleObservation_BackFallsBackWhenFromScheme(t *testing.T) {
	// ?from=https%3A%2F%2Fevil.example.com → decoded: https://evil.example.com
	s := newWithClient(makeObsStub(42))
	req := httptest.NewRequest(http.MethodGet, "/observations/42?from=https%3A%2F%2Fevil.example.com", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, `href="/"`) {
		t.Errorf("expected fallback href=/ for scheme URL, got:\n%s", body)
	}
}

func TestHandleObservation_BackFallsBackWhenFromProtocolRelative(t *testing.T) {
	// ?from=%2F%2Fevil.example.com → decoded: //evil.example.com
	s := newWithClient(makeObsStub(42))
	req := httptest.NewRequest(http.MethodGet, "/observations/42?from=%2F%2Fevil.example.com", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, `href="/"`) {
		t.Errorf("expected fallback href=/ for protocol-relative URL, got:\n%s", body)
	}
}

func TestHandleObservation_BackFallsBackWhenFromTraversal(t *testing.T) {
	// ?from=%2F..%2Fetc%2Fpasswd → decoded: /../etc/passwd
	s := newWithClient(makeObsStub(42))
	req := httptest.NewRequest(http.MethodGet, "/observations/42?from=%2F..%2Fetc%2Fpasswd", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, `href="/"`) {
		t.Errorf("expected fallback href=/ for traversal URL, got:\n%s", body)
	}
}

func TestHandleObservation_BackFallsBackWhenFromOverLength(t *testing.T) {
	// Build a 2049-rune ?from= value starting with /
	longVal := "/" + strings.Repeat("a", 2048)
	s := newWithClient(makeObsStub(42))
	req := httptest.NewRequest(http.MethodGet, "/observations/42?from="+longVal, nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, `href="/"`) {
		t.Errorf("expected fallback href=/ for over-length from, got:\n%s", body)
	}
}

// TestHandleHome_ProjectCardsHaveNoFrom asserts home project cards do NOT inject ?from=
// into their hrefs (FR-6.2, Scenario 12).
func TestHandleHome_ProjectCardsHaveNoFrom(t *testing.T) {
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"alpha", "beta"}},
		recentByProject: map[string][]client.Observation{
			"alpha": {},
			"beta":  {},
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

	// Each project card must have a clean href with no ?from=
	for _, proj := range []string{"alpha", "beta"} {
		cleanHref := fmt.Sprintf(`href="/p/%s"`, proj)
		if !strings.Contains(body, cleanHref) {
			t.Errorf("expected %q in body (no ?from= on home card)", cleanHref)
		}
	}
	// No ?from= should appear anywhere in the home page links
	if strings.Contains(body, "?from=") {
		t.Error("home page must not inject ?from= into any project card links")
	}
}

// TestHandleHome_PreviewRuneSafe tests that multi-byte runes are not split
func TestHandleHome_PreviewRuneSafe(t *testing.T) {
	// Build content: 50 emoji (each 4 bytes) = 50 runes, then long ASCII tail
	content := strings.Repeat("🎉", 50) + strings.Repeat("x", 200)
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"proj"}},
		recentByProject: map[string][]client.Observation{
			"proj": {{ID: 1, Type: "session_summary", Title: "T", Content: content, CreatedAt: "2026-01-01"}},
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

	// The body must be valid UTF-8 (no broken multi-byte sequences)
	if !utf8.ValidString(body) {
		t.Error("response body is not valid UTF-8")
	}
}

// --- static asset handler tests (Task 1.3) ---

// TestStaticHandler_CSS verifies that GET /static/app.css returns 200,
// Content-Type text/css, and Cache-Control: public, max-age=31536000, immutable.
func TestStaticHandler_CSS(t *testing.T) {
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{}},
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/static/app.css", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for /static/app.css, got %d (body: %s)", rr.Code, rr.Body.String())
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/css") {
		t.Errorf("expected Content-Type text/css, got %q", ct)
	}
	cc := rr.Header().Get("Cache-Control")
	if cc != "public, max-age=31536000, immutable" {
		t.Errorf("expected Cache-Control 'public, max-age=31536000, immutable', got %q", cc)
	}
}

// TestStaticHandler_Font verifies that GET /static/fonts/GeistMono-Regular.woff2
// returns 200, Content-Type font/woff2, and the long-cache header.
func TestStaticHandler_Font(t *testing.T) {
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{}},
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/static/fonts/GeistMono-Regular.woff2", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for /static/fonts/GeistMono-Regular.woff2, got %d", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if ct != "font/woff2" {
		t.Errorf("expected Content-Type font/woff2, got %q", ct)
	}
	cc := rr.Header().Get("Cache-Control")
	if cc != "public, max-age=31536000, immutable" {
		t.Errorf("expected Cache-Control 'public, max-age=31536000, immutable', got %q", cc)
	}
}

// --- layout shell tests (Task 3.1) ---

// SCN: TestLayout_LinksStylesheet — every rendered page MUST contain
// <link rel="stylesheet" href="/static/app.css">.
func TestLayout_LinksStylesheet(t *testing.T) {
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{}},
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `rel="stylesheet"`) || !strings.Contains(body, `href="/static/app.css"`) {
		t.Error(`expected <link rel="stylesheet" href="/static/app.css"> in response body`)
	}
}

// SCN: TestLayout_NoTailwindCDN — no rendered page MUST contain the Tailwind CDN script.
func TestLayout_NoTailwindCDN(t *testing.T) {
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{}},
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if strings.Contains(rr.Body.String(), "cdn.tailwindcss.com") {
		t.Error("response body must NOT contain cdn.tailwindcss.com")
	}
}

// SCN: TestLayout_HeaderChrome — header MUST contain brand chrome (dot +
// name + read-only sub) + <nav> with primary actions.
func TestLayout_HeaderChrome(t *testing.T) {
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{}},
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `class="brand__dot"`) {
		t.Error(`expected element with class="brand__dot" in response body`)
	}
	if !strings.Contains(body, `class="brand__name"`) {
		t.Error(`expected element with class="brand__name" in response body`)
	}
	if !strings.Contains(body, `class="brand__sub"`) {
		t.Error(`expected "read-only" sub chip (class="brand__sub") in response body`)
	}
	if !strings.Contains(body, "engram-ui") {
		t.Error("expected text 'engram-ui' inside header")
	}
	if !strings.Contains(body, "<nav") {
		t.Error("expected <nav> element in response body")
	}
}

// SCN: TestLayout_FontPreloads — at least 4 font preload <link> hints MUST be present.
func TestLayout_FontPreloads(t *testing.T) {
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{}},
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	// Count occurrences of preload font link hints
	preloadCount := strings.Count(body, `rel="preload"`)
	if preloadCount < 4 {
		t.Errorf("expected at least 4 font preload <link> hints, found %d", preloadCount)
	}
	if !strings.Contains(body, `as="font"`) {
		t.Error(`expected as="font" attribute on preload links`)
	}
	if !strings.Contains(body, `type="font/woff2"`) {
		t.Error(`expected type="font/woff2" attribute on preload links`)
	}
}

// SCN: TestLayout_FooterEndpoint — footer MUST display "localhost:7437".
func TestLayout_FooterEndpoint(t *testing.T) {
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{}},
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "localhost:7437") {
		t.Error("expected 'localhost:7437' in footer region")
	}
}

// AUX: TestStaticHandler_NoShadowing verifies that existing routes still return correct
// responses after /static/* is mounted.
func TestStaticHandler_NoShadowing(t *testing.T) {
	stub := &stubEngramClient{
		statsOut: &client.Stats{Projects: []string{"alpha"}},
		recentByProject: map[string][]client.Observation{
			"alpha": {{ID: 1, Type: "decision", Title: "Some obs", Content: "c", CreatedAt: "2026-01-01"}},
		},
		recentByProjectErr: map[string]error{},
	}
	s := newWithClient(stub)

	// GET /healthz must return 200 (not shadowed by /static/*)
	t.Run("healthz", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200 from /healthz, got %d", rr.Code)
		}
		body := rr.Body.String()
		if strings.HasPrefix(body, "font/woff2") || rr.Code == http.StatusNotFound {
			t.Error("/healthz must not return a static file response")
		}
	})

	// GET / must return 200
	t.Run("home", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200 from /, got %d", rr.Code)
		}
	})

	// GET /p/alpha must return 200 (not 404)
	t.Run("project", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		if rr.Code == http.StatusNotFound {
			t.Errorf("expected non-404 from /p/alpha, got %d", rr.Code)
		}
	})
}

// SCN: TestToggleTheme_FlipsCookie — POST /toggle-theme with no cookie sets
// theme=light; with theme=light set, it flips back to dark.
func TestToggleTheme_FlipsCookie(t *testing.T) {
	stub := &stubEngramClient{statsOut: &client.Stats{Projects: []string{}}}
	s := newWithClient(stub)

	cases := []struct {
		name    string
		cookie  string // value of "theme" cookie on the incoming request; "" = no cookie
		wantNew string // expected value of the cookie set in the response
	}{
		{"no cookie defaults to light", "", "light"},
		{"dark flips to light", "dark", "light"},
		{"light flips to dark", "light", "dark"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/toggle-theme", nil)
			if tc.cookie != "" {
				req.AddCookie(&http.Cookie{Name: "theme", Value: tc.cookie})
			}
			rr := httptest.NewRecorder()
			s.Handler().ServeHTTP(rr, req)

			if rr.Code != http.StatusSeeOther {
				t.Fatalf("expected 303, got %d", rr.Code)
			}
			var got *http.Cookie
			for _, c := range rr.Result().Cookies() {
				if c.Name == "theme" {
					got = c
					break
				}
			}
			if got == nil {
				t.Fatal("expected response to set the theme cookie")
			}
			if got.Value != tc.wantNew {
				t.Errorf("expected theme=%q, got theme=%q", tc.wantNew, got.Value)
			}
			if !got.HttpOnly {
				t.Error("theme cookie must be HttpOnly")
			}
			if got.SameSite != http.SameSiteLaxMode {
				t.Errorf("theme cookie must be SameSite=Lax, got %v", got.SameSite)
			}
			if got.Path != "/" {
				t.Errorf("theme cookie must have Path=/, got %q", got.Path)
			}
		})
	}
}

// SCN: TestToggleTheme_SafeReferer — same-host Referer is preserved on redirect;
// cross-origin or empty Referer falls back to "/".
func TestToggleTheme_SafeReferer(t *testing.T) {
	stub := &stubEngramClient{statsOut: &client.Stats{Projects: []string{}}}
	s := newWithClient(stub)

	cases := []struct {
		name     string
		referer  string
		wantLoc  string
		hostHdr  string // Host of the simulated request — needed to compare against Referer's Host
	}{
		{"no referer falls back to /", "", "/", "example.com"},
		{"same-host path preserved", "http://example.com/p/alpha?q=foo", "/p/alpha?q=foo", "example.com"},
		{"cross-origin referer rejected", "http://evil.example/steal", "/", "example.com"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/toggle-theme", nil)
			req.Host = tc.hostHdr
			if tc.referer != "" {
				req.Header.Set("Referer", tc.referer)
			}
			rr := httptest.NewRecorder()
			s.Handler().ServeHTTP(rr, req)

			if rr.Code != http.StatusSeeOther {
				t.Fatalf("expected 303, got %d", rr.Code)
			}
			if loc := rr.Header().Get("Location"); loc != tc.wantLoc {
				t.Errorf("expected Location=%q, got %q", tc.wantLoc, loc)
			}
		})
	}
}

// SCN: TestLayout_ThemeAttributeAndToggle — every rendered page MUST carry the
// data-theme attribute on <html> and expose a POST form to /toggle-theme.
func TestLayout_ThemeAttributeAndToggle(t *testing.T) {
	stub := &stubEngramClient{statsOut: &client.Stats{Projects: []string{}}}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `data-theme="dark"`) {
		t.Error(`expected <html data-theme="dark"> when no theme cookie is present`)
	}
	if !strings.Contains(body, `action="/toggle-theme"`) {
		t.Error(`expected a form with action="/toggle-theme" in the header`)
	}
	if !strings.Contains(body, `class="theme-toggle"`) {
		t.Error(`expected a button with class="theme-toggle"`)
	}
}
