// engram-ui is a web viewer for engram.
//
// On startup it checks whether engram's HTTP API is reachable; if not, it
// spawns "engram serve" as a child process (unless --no-spawn is set).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/Gentleman-Programming/engram-ui/internal/client"
	"github.com/Gentleman-Programming/engram-ui/internal/server"
)

func main() {
	engramAddr := flag.String("engram", "http://localhost:7437", "engram REST API base URL")
	listenAddr := flag.String("listen", ":7438", "address engram-ui listens on")
	noSpawn := flag.Bool("no-spawn", false, "fail instead of auto-spawning 'engram serve' when unreachable")
	flag.Parse()

	c := client.New(*engramAddr)

	var spawned *exec.Cmd
	if err := waitForEngram(c, 1*time.Second); err != nil {
		if *noSpawn {
			log.Fatalf("engram unreachable at %s and --no-spawn set: %v", *engramAddr, err)
		}
		log.Printf("engram unreachable, spawning 'engram serve'...")
		cmd, err := spawnEngram()
		if err != nil {
			log.Fatalf("failed to spawn engram serve: %v", err)
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
