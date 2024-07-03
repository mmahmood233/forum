package forum

import "time"

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
    PostLikeID int
    UserID     int
    PostID     int
    IsLike     bool
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
    CommentLikeID int
    UserID        int
    CommentID     int
    IsLike        bool
}

type Category struct {
    ID      int
    CatName string
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
