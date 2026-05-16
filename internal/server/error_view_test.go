package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestErrorView_RendersErrorPageClass — error response body contains an
// element with class="error-page" and shows both title + body message.
func TestErrorView_RendersErrorPageClass(t *testing.T) {
	stub := &stubEngramClient{
		statsErr: errors.New("engram down"),
	}
	s := newWithClient(stub)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 on stats error, got %d", rr.Code)
	}
	body := rr.Body.String()

	// Must contain error-page wrapper
	if !strings.Contains(body, `class="error-page"`) {
		t.Error(`expected element with class="error-page" in error response`)
	}

	// Must contain error-page__title element
	if !strings.Contains(body, `class="error-page__title"`) {
		t.Error(`expected element with class="error-page__title" in error response`)
	}

	// Must contain error-page__body element
	if !strings.Contains(body, `class="error-page__body"`) {
		t.Error(`expected element with class="error-page__body" in error response`)
	}

	// Must display the error title (message passed to renderError)
	if !strings.Contains(body, "could not fetch stats") {
		t.Error(`expected error title "could not fetch stats" in response body`)
	}

	// Must display the error detail (err.Error())
	if !strings.Contains(body, "engram down") {
		t.Error(`expected error body "engram down" in response body`)
	}
}
