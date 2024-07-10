package forum

import (
	"time"
)

type User struct {
	UserID   int
	Email    string
	Username string
	Password string
}

type Post struct {
	PostID      int
	UserID      int
	PostContent string
	CreatedAt   time.Time
}

type PostLike struct {
	PostLikeID int  `json:"post_like_id"`
	UserID     int  `json:"user_id"`
	PostID     int  `json:"post_id"`
	IsLike     bool `json:"is_like"`
}

type PostDislike struct {
	PostDislikeID int  `json:"post_dislike_id"`
	UserID        int  `json:"user_id"`
	PostID        int  `json:"post_id"`
	IsDislike     bool `json:"is_dislike"`
}

type Comment struct {
	CommentID      int
	PostID         int
	UserID         int
	CommentContent string
	CreatedAt      time.Time
	Username       string
}

type CommentLike struct {
	CommentLikeID int  `json:"comment_like_id"`
	UserID        int  `json:"user_id"`
	CommentID     int  `json:"comment_id"`
	IsLike        bool `json:"is_like"`
}

type CommentDislike struct {
	CommentDislikeID int  `json:"comment_dislike_id"`
	UserID           int  `json:"user_id"`
	CommentID        int  `json:"comment_id"`
	IsDislike        bool `json:"is_dislike"`
}

type Category struct {
	CatID   int
	CatName string
	PostID  int
}

type Tag struct {
	TagID   int
	PostID  int
	TagName string
}

type PostCategory struct {
	PostID     int
	CategoryID int
}

type Session struct {
	SessionID string
	UserID    int
	ExpiresAt time.Time
}
