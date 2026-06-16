package juejin

import "strings"

// wireResponse is the top-level API response for list endpoints.
type wireResponse struct {
	ErrNo  int        `json:"err_no"`
	ErrMsg string     `json:"err_msg"`
	Data   []wireItem `json:"data"`
}

// wireObjectResponse is the top-level API response for single-object endpoints.
type wireObjectResponse[T any] struct {
	ErrNo  int    `json:"err_no"`
	ErrMsg string `json:"err_msg"`
	Data   T      `json:"data"`
}

// wireTagListResponse is the top-level response for the tag list endpoint.
type wireTagListResponse struct {
	ErrNo  int       `json:"err_no"`
	ErrMsg string    `json:"err_msg"`
	Data   []wireTag `json:"data"`
}

// wirePinListResponse is the top-level response for the moment/pin list endpoint.
type wirePinListResponse struct {
	ErrNo   int       `json:"err_no"`
	ErrMsg  string    `json:"err_msg"`
	Data    []wirePin `json:"data"`
	HasMore bool      `json:"has_more"`
	Cursor  string    `json:"cursor"`
}

// wireCourseListResponse is the top-level response for the course/column list endpoint.
type wireCourseListResponse struct {
	ErrNo   int          `json:"err_no"`
	ErrMsg  string       `json:"err_msg"`
	Data    []wireCourse `json:"data"`
	HasMore bool         `json:"has_more"`
	Cursor  string       `json:"cursor"`
}

// wireUserData holds the user profile fields returned by user_api.
type wireUserData struct {
	UserID        string `json:"user_id"`
	UserName      string `json:"user_name"`
	Description   string `json:"description"`
	AvatarLarge   string `json:"avatar_large"`
	Company       string `json:"company"`
	JobTitle      string `json:"job_title"`
	Level         int    `json:"level"`
	GotDiggCount  int64  `json:"got_digg_count"`
	GotViewCount  int64  `json:"got_view_count"`
	ArticleCount  int    `json:"article_count"`
	FollowCount   int64  `json:"follow_count"`
	FollowerCount int64  `json:"follower_count"`
}

// wireTag holds the tag fields returned by tag_api.
type wireTag struct {
	TagID        string `json:"tag_id"`
	TagName      string `json:"tag_name"`
	Icon         string `json:"icon"`
	FollowCount  int64  `json:"follow_count"`
	ArticleCount int64  `json:"article_count"`
}

// wirePin holds pin (沸点 / moment) fields.
type wirePin struct {
	MsgID        string       `json:"msg_id"`
	Content      string       `json:"content"`
	DiggCount    int64        `json:"digg_count"`
	CommentCount int          `json:"comment_count"`
	UserInfo     wirePinUser  `json:"user_info"`
	// ctime is a Unix timestamp string (seconds since epoch).
	Ctime        string       `json:"ctime"`
}

type wirePinUser struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

// wireCourse holds course/column fields.
type wireCourse struct {
	ColumnID        string          `json:"column_id"`
	Title           string          `json:"title"`
	BriefContent    string          `json:"brief_content"`
	Cover           string          `json:"cover"`
	Price           int64           `json:"price"`
	SubscribeCount  int64           `json:"subscribe_count"`
	IsFree          int             `json:"is_free"`
	AuthorUserInfo  wireCourseAuthor `json:"author_user_info"`
}

type wireCourseAuthor struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

// UserProfile is the record emitted by the user command.
type UserProfile struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Avatar        string `json:"avatar"`
	Company       string `json:"company"`
	JobTitle      string `json:"job_title"`
	Level         int    `json:"level"`
	GotDiggs      int64  `json:"got_diggs"`
	GotViews      int64  `json:"got_views"`
	ArticleCount  int    `json:"article_count"`
	FollowCount   int64  `json:"follow_count"`
	FollowerCount int64  `json:"follower_count"`
	URL           string `json:"url"`
}

// Tag is the record emitted by the tags command.
type Tag struct {
	Rank         int    `json:"rank"`
	ID           string `json:"id"`
	Name         string `json:"name"`
	Icon         string `json:"icon"`
	FollowCount  int64  `json:"follow_count"`
	ArticleCount int64  `json:"article_count"`
	URL          string `json:"url"`
}

// Pin is the record emitted by the pins command.
type Pin struct {
	Rank         int    `json:"rank"`
	ID           string `json:"id"`
	Content      string `json:"content"`
	Author       string `json:"author"`
	AuthorID     string `json:"author_id"`
	DiggCount    int64  `json:"digg_count"`
	CommentCount int    `json:"comment_count"`
	CreatedAt    string `json:"created_at"`
	URL          string `json:"url"`
}

type wireItem struct {
	ItemType int          `json:"item_type"`
	ItemInfo wireItemInfo `json:"item_info"`
}

type wireItemInfo struct {
	ArticleID   string          `json:"article_id"`
	ArticleInfo wireArticleInfo `json:"article_info"`
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

// Article is the record emitted for article list commands.
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
