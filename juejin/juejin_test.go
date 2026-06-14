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
