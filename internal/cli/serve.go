package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Gentleman-Programming/engram-ui/internal/client"
	"github.com/Gentleman-Programming/engram-ui/internal/server"
)

// alreadyRunningCheck is a function variable for test injection.
var alreadyRunningCheck = IsAlreadyRunning

// cmdServe parses the serve flags and runs the engram-ui web server.
// It mirrors the v2 main.go logic exactly, preserving backward compatibility.
func cmdServe(args []string) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(stderr)

	engramAddr := fs.String("engram", "http://localhost:7437", "engram REST API base URL")
	listenAddr := fs.String("listen", ":7438", "address engram-ui listens on")
	noSpawn := fs.Bool("no-spawn", false, "fail instead of auto-spawning 'engram serve' when unreachable")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	// Already-running guard (FR-5).
	if alreadyRunningCheck(*listenAddr) {
		log.Printf("engram-ui already running on %s, exiting cleanly", *listenAddr)
		return 0
	}

	c := client.New(*engramAddr)

	var spawned *exec.Cmd
	if err := waitForEngram(c, 1*time.Second); err != nil {
		if *noSpawn {
			log.Fatalf("engram unreachable at %s and --no-spawn set: %v", *engramAddr, err)
		}
		log.Printf("engram unreachable, spawning 'engram serve'...")
		cmd, spawnErr := spawnEngram()
		if spawnErr != nil {
			log.Fatalf("failed to spawn engram serve: %v", spawnErr)
		}
		spawned = cmd
		if err := waitForEngram(c, 10*time.Second); err != nil {
			stopSpawned(spawned)
			log.Fatalf("engram still unreachable after spawn: %v", err)
		}
		log.Printf("engram serve up (pid=%d)", spawned.Process.Pid)
	} else {
		log.Printf("engram already reachable at %s", *engramAddr)
	}

	srv := server.New(c)
	httpSrv := &http.Server{
		Addr:              *listenAddr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("engram-ui listening on %s", *listenAddr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)

	if spawned != nil {
		stopSpawned(spawned)
	}

	return 0
}

// IsAlreadyRunning checks whether an engram-ui instance is already listening
// on the given address. It accepts either a full URL (http://host:port) or a
// bare listen address like ":7438" (which is normalized to http://localhost:port).
// Returns true only when the probe returns 200 OK with body "ok".
func IsAlreadyRunning(listenAddr string) bool {
	baseURL := normalizeListenAddr(listenAddr)

	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get(baseURL + "/healthz")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 32))
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(body)) == "ok"
}

// normalizeListenAddr converts bare ":port" to "http://localhost:port".
// Full URLs (starting with "http") are returned unchanged.
func normalizeListenAddr(addr string) string {
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return addr
	}
	// Strip leading colon if present (e.g. ":7438" -> "7438").
	port := strings.TrimPrefix(addr, ":")
	return "http://localhost:" + port
}

func waitForEngram(c *client.Client, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		if err := c.Health(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(200 * time.Millisecond)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("timeout waiting for engram")
	}
	return lastErr
}

func spawnEngram() (*exec.Cmd, error) {
	cmd := exec.Command("engram", "serve")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func stopSpawned(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	log.Printf("stopping engram serve (pid=%d)", cmd.Process.Pid)
	_ = cmd.Process.Kill()
	_, _ = cmd.Process.Wait()
}
