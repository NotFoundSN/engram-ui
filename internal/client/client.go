// Package client is an HTTP client for engram's REST API.
//
// It targets engram's public contract (the routes registered in
// internal/server/server.go) so that engram-ui stays decoupled from
// engram's internal schema. If engram changes storage internals,
// this client keeps working as long as the REST shape is stable.
package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	base string
	http *http.Client
}

func New(base string) *Client {
	return &Client{
		base: base,
		http: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) BaseURL() string { return c.base }

func (c *Client) Health() error {
	resp, err := c.http.Get(c.base + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("engram /health returned status %d", resp.StatusCode)
	}
	return nil
}

// Stats mirrors engram's store.Stats JSON shape.
type Stats struct {
	TotalSessions     int      `json:"total_sessions"`
	TotalObservations int      `json:"total_observations"`
	TotalPrompts      int      `json:"total_prompts"`
	Projects          []string `json:"projects"`
}

func (c *Client) Stats() (*Stats, error) {
	var out Stats
	if err := c.getJSON("/stats", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Observation mirrors engram's store.Observation JSON shape.
type Observation struct {
	ID             int64   `json:"id"`
	SyncID         string  `json:"sync_id"`
	SessionID      string  `json:"session_id"`
	Type           string  `json:"type"`
	Title          string  `json:"title"`
	Content        string  `json:"content"`
	ToolName       *string `json:"tool_name,omitempty"`
	Project        *string `json:"project,omitempty"`
	Scope          string  `json:"scope"`
	TopicKey       *string `json:"topic_key,omitempty"`
	RevisionCount  int     `json:"revision_count"`
	DuplicateCount int     `json:"duplicate_count"`
	LastSeenAt     *string `json:"last_seen_at,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
	DeletedAt      *string `json:"deleted_at,omitempty"`
}

// SearchResult is an observation plus a relevance rank.
type SearchResult struct {
	Observation
	Rank float64 `json:"rank"`
}

type SearchOptions struct {
	Type    string
	Project string
	Scope   string
	Limit   int
}

func (c *Client) Search(query string, opts SearchOptions) ([]SearchResult, error) {
	q := url.Values{}
	q.Set("q", query)
	if opts.Type != "" {
		q.Set("type", opts.Type)
	}
	if opts.Project != "" {
		q.Set("project", opts.Project)
	}
	if opts.Scope != "" {
		q.Set("scope", opts.Scope)
	}
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	}

	var out []SearchResult
	if err := c.getJSON("/search", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) Observation(id int64) (*Observation, error) {
	var out Observation
	path := fmt.Sprintf("/observations/%d", id)
	if err := c.getJSON(path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type RecentOptions struct {
	Project string
	Scope   string
	Limit   int
}

func (c *Client) RecentObservations(opts RecentOptions) ([]Observation, error) {
	q := url.Values{}
	if opts.Project != "" {
		q.Set("project", opts.Project)
	}
	if opts.Scope != "" {
		q.Set("scope", opts.Scope)
	}
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	}

	var out []Observation
	if err := c.getJSON("/observations/recent", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) getJSON(path string, query url.Values, out any) error {
	u := c.base + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	resp, err := c.http.Get(u)
	if err != nil {
		return fmt.Errorf("GET %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: status %d", path, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("GET %s: decode: %w", path, err)
	}
	return nil
}
