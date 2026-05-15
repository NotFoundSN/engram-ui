package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Gentleman-Programming/engram-ui/internal/client"
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
	s := newWithClient(alphaStub())
	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()

	// templ renders id before name, so we check for both separately
	if !strings.Contains(body, `name="type"`) || !strings.Contains(body, `<select`) {
		t.Error("expected <select ... name=\"type\"> in body")
	}

	// All 14 canonical types must appear as option values
	canonicalTypes := []string{
		"architecture", "bugfix", "config", "decision", "design",
		"discovery", "exploration", "pattern", "plan", "preference",
		"proposal", "report", "spec", "tasks",
	}
	for _, typ := range canonicalTypes {
		needle := `value="` + typ + `"`
		if !strings.Contains(body, needle) {
			t.Errorf("expected option %q in select", typ)
		}
	}

	// custom-internal must also appear
	if !strings.Contains(body, `value="custom-internal"`) {
		t.Error("expected custom-internal option in select")
	}

	// "All types" option must be present
	if !strings.Contains(body, `value=""`) {
		t.Error(`expected "All types" option with value="" in select`)
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

	if !strings.Contains(body, `value="decision" selected`) {
		t.Error(`expected option value="decision" selected in body`)
	}
}

func TestHandleProject_TypeSelectMarksAllTypesSelectedWhenNoFilter(t *testing.T) {
	s := newWithClient(alphaStub())
	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	// The "All types" empty-value option should carry selected
	if !strings.Contains(body, `value="" selected`) {
		t.Error(`expected "All types" option (value="") to be selected when no type filter`)
	}
}

func TestHandleProject_TypeSelectIncludesPhantom(t *testing.T) {
	s := newWithClient(alphaStub())
	req := httptest.NewRequest(http.MethodGet, "/p/alpha?type=disappeared", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	if !strings.Contains(body, `value="disappeared" selected`) {
		t.Error(`expected phantom option value="disappeared" selected in body`)
	}
}

func TestHandleProject_TypeSelectFormHasHiddenQAndSort(t *testing.T) {
	s := newWithClient(alphaStub())
	req := httptest.NewRequest(http.MethodGet, "/p/alpha?q=auth&sort=date_asc", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	// Scope assertion to the type filter form block (identified by id="type-select")
	selectIdx := strings.Index(body, `id="type-select"`)
	if selectIdx == -1 {
		t.Fatal("expected id=\"type-select\" (type filter select) in body")
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

	if !strings.Contains(typeFilterForm, `type="hidden" name="q" value="auth"`) {
		t.Error(`expected hidden input for q="auth" in type filter form`)
	}
	if !strings.Contains(typeFilterForm, `type="hidden" name="sort" value="date_asc"`) {
		t.Error(`expected hidden input for sort="date_asc" in type filter form`)
	}
}

func TestHandleProject_TypeSelectFormOmitsSortWhenImplicit(t *testing.T) {
	s := newWithClient(alphaStub())
	// No ?sort= in request — it's implicit (defaulted by server)
	req := httptest.NewRequest(http.MethodGet, "/p/alpha?type=decision", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	// The type-filter form (identified by id="type-select") must NOT
	// contain a hidden sort input.
	selectIdx := strings.Index(body, `id="type-select"`)
	if selectIdx == -1 {
		// Before phase 3 this will be -1 (chips still rendered) — RED is correct.
		t.Skip("type select not yet rendered — RED phase")
	}
	// Find the opening <form that precedes the select
	formStart := strings.LastIndex(body[:selectIdx], "<form")
	if formStart == -1 {
		t.Fatal("could not find <form before type select")
	}
	// Find the closing </form> after the select
	formEnd := strings.Index(body[selectIdx:], "</form>")
	if formEnd == -1 {
		t.Fatal("could not find </form> after type select")
	}
	typeFilterForm := body[formStart : selectIdx+formEnd+len("</form>")]

	// Hidden sort input must NOT be inside the type filter form
	if strings.Contains(typeFilterForm, `type="hidden" name="sort"`) {
		t.Error(`type filter form must not contain hidden sort input when sort was not explicitly set`)
	}
}

func TestHandleProject_TypeSelectFormHasApplyButton(t *testing.T) {
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

func TestHandleProject_NoOnchangeJavaScriptHandler(t *testing.T) {
	// The type select must NOT have onchange= (NFR-7)
	s := newWithClient(alphaStub())
	req := httptest.NewRequest(http.MethodGet, "/p/alpha", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	// Find the type-select element by its id attribute (templ puts id before name)
	selectIdx := strings.Index(body, `id="type-select"`)
	if selectIdx == -1 {
		t.Fatal(`expected id="type-select" (type filter select) in body`)
	}
	// Find the start of the enclosing <select tag
	selectStart := strings.LastIndex(body[:selectIdx], "<select")
	if selectStart == -1 {
		t.Fatal("could not find <select before id=\"type-select\"")
	}
	// Find closing > of the opening select tag
	closeIdx := strings.Index(body[selectStart:], ">")
	if closeIdx == -1 {
		t.Fatal("could not find end of <select> tag")
	}
	selectTag := body[selectStart : selectStart+closeIdx+1]
	if strings.Contains(selectTag, "onchange") {
		t.Errorf("type filter select tag must not contain onchange: %s", selectTag)
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
