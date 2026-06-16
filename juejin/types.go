package juejin

import "strings"

// wireResponse is the top-level API response for feed endpoints.
type wireResponse struct {
	ErrNo  int        `json:"err_no"`
	ErrMsg string     `json:"err_msg"`
	Data   []wireItem `json:"data"`
}

type wireItem struct {
	ItemType int          `json:"item_type"`
	ItemInfo wireItemInfo `json:"item_info"`
}

type wireItemInfo struct {
	ArticleID      string          `json:"article_id"`
	ArticleInfo    wireArticleInfo `json:"article_info"`
	AuthorUserInfo struct {
		UserName string `json:"user_name"`
		UserID   string `json:"user_id"`
	} `json:"author_user_info"`
	Tags     []struct{ TagName string `json:"tag_name"` } `json:"tags"`
	Category struct{ CategoryName string `json:"category_name"` } `json:"category"`
}

type wireArticleInfo struct {
	Title        string `json:"title"`
	BriefContent string `json:"brief_content"`
	ViewCount    int    `json:"view_count"`
	DiggCount    int    `json:"digg_count"`
	CommentCount int    `json:"comment_count"`
	CollectCount int    `json:"collect_count"`
}

// wireSearchResponse is the top-level API response for the search endpoint.
type wireSearchResponse struct {
	ErrNo  int             `json:"err_no"`
	ErrMsg string          `json:"err_msg"`
	Data   wireSearchData  `json:"data"`
}

type wireSearchData struct {
	Count   int              `json:"count"`
	HasMore bool             `json:"has_more"`
	Cursor  string           `json:"cursor"`
	Items   []wireSearchItem `json:"items"`
}

type wireSearchItem struct {
	ResultModel wireSearchModel `json:"result_model"`
}

type wireSearchModel struct {
	ArticleID      string          `json:"article_id"`
	ArticleInfo    wireArticleInfo `json:"article_info"`
	AuthorUserInfo struct {
		UserName string `json:"user_name"`
		UserID   string `json:"user_id"`
	} `json:"author_user_info"`
	Tags     []struct{ TagName string `json:"tag_name"` } `json:"tags"`
	Category struct{ CategoryName string `json:"category_name"` } `json:"category"`
}

// Article is the record emitted for article list commands (hot, latest).
type Article struct {
	Rank     int    `json:"rank"`
	ID       string `json:"id"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	Category string `json:"category"`
	Tags     string `json:"tags"`
	Views    int    `json:"views"`
	Diggs    int    `json:"diggs"`
	Comments int    `json:"comments"`
	Brief    string `json:"brief"`
	URL      string `json:"url"`
}

// SearchResult is the record emitted for the search command.
type SearchResult struct {
	Rank     int    `json:"rank"`
	ID       string `json:"id"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	Category string `json:"category"`
	Tags     string `json:"tags"`
	Views    int    `json:"views"`
	Diggs    int    `json:"diggs"`
	Comments int    `json:"comments"`
	Brief    string `json:"brief"`
	URL      string `json:"url"`
	Query    string `json:"query"`
}

func wireToArticle(item wireItem, rank int) Article {
	info := item.ItemInfo
	ai := info.ArticleInfo

	var tagNames []string
	for i, t := range info.Tags {
		if i >= 3 {
			break
		}
		tagNames = append(tagNames, t.TagName)
	}

	return Article{
		Rank:     rank,
		ID:       info.ArticleID,
		Title:    ai.Title,
		Author:   info.AuthorUserInfo.UserName,
		Category: info.Category.CategoryName,
		Tags:     strings.Join(tagNames, ", "),
		Views:    ai.ViewCount,
		Diggs:    ai.DiggCount,
		Comments: ai.CommentCount,
		Brief:    ai.BriefContent,
		URL:      "https://juejin.cn/post/" + info.ArticleID,
	}
}

func wireSearchModelToResult(m wireSearchModel, rank int, query string) SearchResult {
	ai := m.ArticleInfo

	var tagNames []string
	for i, t := range m.Tags {
		if i >= 3 {
			break
		}
		tagNames = append(tagNames, t.TagName)
	}

	return SearchResult{
		Rank:     rank,
		ID:       m.ArticleID,
		Title:    ai.Title,
		Author:   m.AuthorUserInfo.UserName,
		Category: m.Category.CategoryName,
		Tags:     strings.Join(tagNames, ", "),
		Views:    ai.ViewCount,
		Diggs:    ai.DiggCount,
		Comments: ai.CommentCount,
		Brief:    ai.BriefContent,
		URL:      "https://juejin.cn/post/" + m.ArticleID,
		Query:    query,
	}
}
