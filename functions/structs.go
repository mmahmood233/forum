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
	LikeCount   int
    DislikeCount int
}

type PostLike struct {
	PostLikeID int  
	UserID     int  
	PostID     int  
	IsLike     bool 
}

type PostDislike struct {
	PostDislikeID int
	UserID        int
	PostID        int
	IsDislike     bool
}

type Comment struct {
	CommentID      int
	PostID         int
	UserID         int
	CommentContent string
	CreatedAt      time.Time
	Username       string
	LikeCount   int
    DislikeCount int
}

type CommentLike struct {
	CommentLikeID int
	UserID        int
	CommentID     int
	IsLike        bool
}

type CommentDislike struct {
	CommentDislikeID int
	UserID           int
	CommentID        int
	IsDislike        bool
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

type Error struct {
	Err int
	ErrStr string
}

