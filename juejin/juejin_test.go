package juejin_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tamnd/juejin-cli/juejin"
)

const mockFeedResponse = `{
  "err_no": 0,
  "err_msg": "success",
  "data": [
    {
      "item_type": 2,
      "item_info": {
        "article_id": "7123456789012345678",
        "article_info": {
          "title": "Go 语言并发编程实践",
          "brief_content": "本文介绍 Go 语言并发编程的最佳实践",
          "view_count": 5000,
          "digg_count": 200,
          "comment_count": 42,
          "collect_count": 88
        },
        "author_user_info": {
          "user_name": "gopher42",
          "user_id": "uid001"
        },
        "tags": [
          {"tag_name": "Go"},
          {"tag_name": "并发"},
          {"tag_name": "后端"}
        ],
        "category": {
          "category_name": "后端"
        }
      }
    },
    {
      "item_type": 2,
      "item_info": {
        "article_id": "7999999999999999999",
        "article_info": {
          "title": "React 18 新特性解析",
          "brief_content": "深入解析 React 18 的并发特性",
          "view_count": 12000,
          "digg_count": 500,
          "comment_count": 120,
          "collect_count": 300
        },
        "author_user_info": {
          "user_name": "frontend_dev",
          "user_id": "uid002"
        },
        "tags": [
          {"tag_name": "React"},
          {"tag_name": "前端"},
          {"tag_name": "JavaScript"},
          {"tag_name": "extra-tag"}
        ],
        "category": {
          "category_name": "前端"
        }
      }
    }
  ]
}`

const mockErrorResponse = `{"err_no":403,"err_msg":"permission denied","data":[]}`

func newTestClient(ts *httptest.Server) *juejin.Client {
	cfg := juejin.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return juejin.NewClient(cfg)
}

func TestFeedAllSendsUserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write([]byte(mockFeedResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Feed(context.Background(), 200, "", "0", 20)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFeedAllUsesAllFeedEndpoint(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(mockFeedResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Feed(context.Background(), 200, "", "0", 20)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/recommend_api/v1/article/recommend_all_feed" {
		t.Errorf("path = %q, want /recommend_api/v1/article/recommend_all_feed", gotPath)
	}
}

func TestFeedCategoryUsesCateFeedEndpoint(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		_, _ = w.Write([]byte(mockFeedResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Feed(context.Background(), 3, "6809637769959178254", "0", 10)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/recommend_api/v1/article/recommend_cate_feed" {
		t.Errorf("path = %q, want /recommend_api/v1/article/recommend_cate_feed", gotPath)
	}
	if gotBody["category_id"] != "6809637769959178254" {
		t.Errorf("category_id = %v, want 6809637769959178254", gotBody["category_id"])
	}
}

func TestFeedParsesArticles(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockFeedResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	articles, err := c.Feed(context.Background(), 200, "", "0", 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 2 {
		t.Fatalf("got %d articles, want 2", len(articles))
	}

	a := articles[0]
	if a.Rank != 1 {
		t.Errorf("rank = %d, want 1", a.Rank)
	}
	if a.Title != "Go 语言并发编程实践" {
		t.Errorf("title = %q", a.Title)
	}
	if a.Author != "gopher42" {
		t.Errorf("author = %q, want gopher42", a.Author)
	}
	if a.Category != "后端" {
		t.Errorf("category = %q, want 后端", a.Category)
	}
	if a.Views != 5000 {
		t.Errorf("views = %d, want 5000", a.Views)
	}
	if a.Diggs != 200 {
		t.Errorf("diggs = %d, want 200", a.Diggs)
	}
	if a.Comments != 42 {
		t.Errorf("comments = %d, want 42", a.Comments)
	}
	if a.Tags != "Go, 并发, 后端" {
		t.Errorf("tags = %q, want 3 tags", a.Tags)
	}
	if a.URL != "https://juejin.cn/post/7123456789012345678" {
		t.Errorf("url = %q", a.URL)
	}

	// Second article: tags capped at 3 even when 4 present.
	a2 := articles[1]
	if a2.Rank != 2 {
		t.Errorf("rank = %d, want 2", a2.Rank)
	}
	if strings.Contains(a2.Tags, "extra-tag") {
		t.Errorf("tags should be capped at 3, got %q", a2.Tags)
	}
}

func TestFeedSendsContentTypeJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", r.Header.Get("Content-Type"))
		}
		_, _ = w.Write([]byte(mockFeedResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Feed(context.Background(), 200, "", "0", 5)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFeedRetriesOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(mockFeedResponse))
	}))
	defer srv.Close()

	cfg := juejin.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := juejin.NewClient(cfg)

	_, err := c.Feed(context.Background(), 200, "", "0", 5)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
}

func TestFeedAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockErrorResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Feed(context.Background(), 200, "", "0", 5)
	if err == nil {
		t.Fatal("expected error from api err_no, got nil")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("error = %v, want 403", err)
	}
}

// --- Search tests ---

const mockSearchResponse = `{
  "err_no": 0,
  "err_msg": "success",
  "data": {
    "count": 2,
    "has_more": false,
    "cursor": "2",
    "items": [
      {
        "result_model": {
          "article_id": "7111111111111111111",
          "article_info": {
            "title": "Go 并发模式详解",
            "brief_content": "详细介绍 Go 语言的并发模型",
            "view_count": 8000,
            "digg_count": 350,
            "comment_count": 60,
            "collect_count": 120
          },
          "author_user_info": {
            "user_name": "gopherdev",
            "user_id": "uid999"
          },
          "tags": [
            {"tag_name": "Go"},
            {"tag_name": "并发"},
            {"tag_name": "goroutine"},
            {"tag_name": "extra"}
          ],
          "category": {
            "category_name": "后端"
          }
        }
      },
      {
        "result_model": {
          "article_id": "7222222222222222222",
          "article_info": {
            "title": "Go channel 使用指南",
            "brief_content": "channel 的正确使用方式",
            "view_count": 4000,
            "digg_count": 150,
            "comment_count": 30,
            "collect_count": 60
          },
          "author_user_info": {
            "user_name": "chandev",
            "user_id": "uid888"
          },
          "tags": [
            {"tag_name": "Go"},
            {"tag_name": "channel"}
          ],
          "category": {
            "category_name": "后端"
          }
        }
      }
    ]
  }
}`

const mockSearchErrorResponse = `{"err_no":403,"err_msg":"permission denied","data":{"count":0,"has_more":false,"cursor":"0","items":[]}}`

const mockSearchEmptyResponse = `{"err_no":0,"err_msg":"success","data":{"count":0,"has_more":false,"cursor":"0","items":[]}}`

func TestSearchUsesSearchEndpoint(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(mockSearchResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Search(context.Background(), "go", "0", 10)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/search_api/v1/search" {
		t.Errorf("path = %q, want /search_api/v1/search", gotPath)
	}
}

func TestSearchSendsKeywords(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		_, _ = w.Write([]byte(mockSearchResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Search(context.Background(), "go concurrency", "0", 5)
	if err != nil {
		t.Fatal(err)
	}
	if gotBody["keywords"] != "go concurrency" {
		t.Errorf("keywords = %v, want 'go concurrency'", gotBody["keywords"])
	}
}

func TestSearchParsesResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockSearchResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	results, err := c.Search(context.Background(), "go", "0", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}

	r := results[0]
	if r.Rank != 1 {
		t.Errorf("rank = %d, want 1", r.Rank)
	}
	if r.Title != "Go 并发模式详解" {
		t.Errorf("title = %q", r.Title)
	}
	if r.Author != "gopherdev" {
		t.Errorf("author = %q, want gopherdev", r.Author)
	}
	if r.Category != "后端" {
		t.Errorf("category = %q, want 后端", r.Category)
	}
	if r.Views != 8000 {
		t.Errorf("views = %d, want 8000", r.Views)
	}
	if r.Diggs != 350 {
		t.Errorf("diggs = %d, want 350", r.Diggs)
	}
	if r.URL != "https://juejin.cn/post/7111111111111111111" {
		t.Errorf("url = %q", r.URL)
	}
	if r.Query != "go" {
		t.Errorf("query = %q, want 'go'", r.Query)
	}
	// tags capped at 3
	if strings.Contains(r.Tags, "extra") {
		t.Errorf("tags should be capped at 3, got %q", r.Tags)
	}
	expected := "Go, 并发, goroutine"
	if r.Tags != expected {
		t.Errorf("tags = %q, want %q", r.Tags, expected)
	}
}

func TestSearchAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockSearchErrorResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Search(context.Background(), "go", "0", 5)
	if err == nil {
		t.Fatal("expected error from api err_no, got nil")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("error = %v, want 403", err)
	}
}

func TestSearchEmptyResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockSearchEmptyResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	results, err := c.Search(context.Background(), "xyznonexistent", "0", 5)
	if err != nil {
		t.Fatalf("expected no error on empty results, got %v", err)
	}
	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
}
