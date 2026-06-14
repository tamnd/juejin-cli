// Package juejin is the library behind the juejin command: the HTTP client,
// request shaping, and the typed data models for Juejin (掘金, juejin.cn).
//
// The client posts queries to the public Juejin recommendation API at
// https://api.juejin.cn. No authentication is required. It sets a real
// User-Agent, paces requests, and retries transient 429/5xx errors with
// exponential backoff.
package juejin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// DefaultUserAgent identifies the client to Juejin.
const DefaultUserAgent = "juejin/dev (+https://github.com/tamnd/juejin-cli)"

// Config holds constructor parameters.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Retries   int
	Timeout   time.Duration
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://api.juejin.cn",
		UserAgent: DefaultUserAgent,
		Rate:      500 * time.Millisecond,
		Retries:   3,
		Timeout:   30 * time.Second,
	}
}

// Client talks to the Juejin recommendation API.
type Client struct {
	cfg        Config
	httpClient *http.Client
	mu         sync.Mutex
	last       time.Time
}

// NewClient returns a Client with the given config.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}
}

// Feed fetches articles from the Juejin recommendation API.
//
// sortType controls ordering: 3=newest, 200=hot.
// categoryID filters by category; pass "" to fetch the global all-feed.
// cursor is the pagination cursor, use "0" for the first page.
// limit is the number of articles to return.
func (c *Client) Feed(ctx context.Context, sortType int, categoryID string, cursor string, limit int) ([]Article, error) {
	var url string
	var bodyMap map[string]any

	if categoryID == "" {
		url = c.cfg.BaseURL + "/recommend_api/v1/article/recommend_all_feed"
		bodyMap = map[string]any{
			"id_type":   2,
			"sort_type": sortType,
			"cursor":    cursor,
			"limit":     limit,
		}
	} else {
		url = c.cfg.BaseURL + "/recommend_api/v1/article/recommend_cate_feed"
		bodyMap = map[string]any{
			"id_type":     2,
			"sort_type":   sortType,
			"cursor":      cursor,
			"limit":       limit,
			"category_id": categoryID,
		}
	}

	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, err
	}

	raw, err := c.post(ctx, url, bodyBytes)
	if err != nil {
		return nil, err
	}

	var resp wireResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("decode feed: %w", err)
	}
	if resp.ErrNo != 0 {
		return nil, fmt.Errorf("juejin api error %d: %s", resp.ErrNo, resp.ErrMsg)
	}

	out := make([]Article, 0, len(resp.Data))
	for i, item := range resp.Data {
		out = append(out, wireToArticle(item, i+1))
	}
	return out, nil
}

func (c *Client) post(ctx context.Context, url string, body []byte) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		b, retry, err := c.do(ctx, url, body)
		if err == nil {
			return b, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("post %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, url string, body []byte) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}
