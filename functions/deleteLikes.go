package forum

import (
	"database/sql"
)

func DeletePostLike(db *sql.DB, userID, postID int) error {
	_, err := db.Exec("DELETE FROM post_likes WHERE user_id = ? AND post_id = ?", userID, postID)
	return err
}

func DeletePostDislike(db *sql.DB, userID, postID int) error {
	_, err := db.Exec("DELETE FROM post_dislikes WHERE user_id = ? AND post_id = ?", userID, postID)
	return err
}

func DeleteCommentLike(db *sql.DB, userID, commentID int) error {
	_, err := db.Exec("DELETE FROM comment_likes WHERE user_id = ? AND comment_id = ?", userID, commentID)
	return err
}

func DeleteCommentDislike(db *sql.DB, userID, commentID int) error {
	_, err := db.Exec("DELETE FROM comment_dislikes WHERE user_id = ? AND comment_id = ?", userID, commentID)
	return err
}