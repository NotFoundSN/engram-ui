package cli

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
	"time"

	eclient "github.com/Gentleman-Programming/engram-ui/internal/client"
)

func TestIsAlreadyRunning(t *testing.T) {
	cases := []struct {
		name    string
		handler http.HandlerFunc
		want    bool
	}{
		{
			name: "200 with ok body returns true",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
			},
			want: true,
		},
		{
			name: "200 with wrong body returns false",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("not-engram"))
			},
			want: false,
		},
		{
			name: "500 returns false",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			want: false,
		},
		{
			name: "200 ok with trailing whitespace returns true",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok\n"))
			},
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(tc.handler)
			defer srv.Close()
			// Use the test server's address (host:port form, without http://).
			// IsAlreadyRunning accepts a listenAddr like ":7438" but we pass the full URL base here.
			got := IsAlreadyRunning(srv.URL)
			if got != tc.want {
				t.Errorf("IsAlreadyRunning(%q) = %v, want %v", srv.URL, got, tc.want)
			}
		})
	}
}

func TestCmdServe_AlreadyRunning(t *testing.T) {
	// Spin up a stub server that mimics engram-ui's /healthz response.
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}
	}))
	defer stub.Close()

	// Inject: override the alreadyRunningCheck so cmdServe uses the stub URL.
	origCheck := alreadyRunningCheck
	defer func() { alreadyRunningCheck = origCheck }()
	alreadyRunningCheck = func(addr string) bool {
		return IsAlreadyRunning(stub.URL)
	}

	// cmdServe should detect already-running and return 0 without binding.
	code := cmdServe([]string{})
	if code != 0 {
		t.Errorf("cmdServe (already running): exit code = %d, want 0", code)
	}
}

func TestIsAlreadyRunning_NoServer(t *testing.T) {
	// Port that is almost certainly not listening.
	got := IsAlreadyRunning("http://localhost:19999")
	if got {
		t.Error("IsAlreadyRunning: expected false when no server is listening")
	}
}

func TestCmdServe_FlagParseError(t *testing.T) {
	// A completely invalid flag should return exit 2.
	var buf bytes.Buffer
	origStderr := stderr
	stderr = &buf
	defer func() { stderr = origStderr }()

	code := cmdServe([]string{"--invalid-flag-that-does-not-exist=xyz"})
	if code != 2 {
		t.Errorf("cmdServe with invalid flag = %d, want 2", code)
	}
}

func TestWaitForEngram_Healthy(t *testing.T) {
	// Stub HTTP server returning health OK.
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer stub.Close()

	c := newEngram(stub.URL)
	err := waitForEngram(c, 2*time.Second)
	if err != nil {
		t.Errorf("waitForEngram: expected nil, got %v", err)
	}
}

func TestWaitForEngram_Timeout(t *testing.T) {
	// No server on this port — should timeout.
	c := newEngram("http://localhost:29887")
	err := waitForEngram(c, 300*time.Millisecond)
	if err == nil {
		t.Error("waitForEngram: expected error on unreachable server, got nil")
	}
}

func TestNormalizeListenAddr(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{":7438", "http://localhost:7438"},
		{":9000", "http://localhost:9000"},
		{"http://localhost:7438", "http://localhost:7438"},
		{"http://127.0.0.1:7438", "http://127.0.0.1:7438"},
		{"https://example.com:443", "https://example.com:443"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := normalizeListenAddr(tc.input)
			if got != tc.want {
				t.Errorf("normalizeListenAddr(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// newEngram creates a real *client.Client pointing at the given base URL.
func newEngram(baseURL string) *eclient.Client {
	return eclient.New(baseURL)
}

func TestStopSpawned_Nil(t *testing.T) {
	// stopSpawned(nil) must not panic.
	stopSpawned(nil)
}

func TestStopSpawned_NilProcess(t *testing.T) {
	// stopSpawned with cmd but nil Process must not panic.
	cmd := &exec.Cmd{}
	stopSpawned(cmd)
}
