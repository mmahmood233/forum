package forum

import (
	"database/sql"
	"log"
)

func InsertCommentLike(db *sql.DB, commentLike *CommentLike) error {
	// Check if the user has already liked the comment
	var existingLike bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM comment_likes WHERE user_id = ? AND comment_id = ?)", commentLike.UserID, commentLike.CommentID).Scan(&existingLike)
	if err != nil {
		return err
	}

	if existingLike {
		// Delete the existing like
		_, err = db.Exec("DELETE FROM comment_likes WHERE user_id = ? AND comment_id = ?", commentLike.UserID, commentLike.CommentID)
		if err != nil {
			return err
		}
		return nil
	}

	// Check if the user has already disliked the comment
	var existingDislike bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM comment_dislikes WHERE user_id = ? AND comment_id = ?)", commentLike.UserID, commentLike.CommentID).Scan(&existingDislike)
	if err != nil {
		return err
	}

	if existingDislike {
		// Delete the existing dislike
		_, err = db.Exec("DELETE FROM comment_dislikes WHERE user_id = ? AND comment_id = ?", commentLike.UserID, commentLike.CommentID)
		if err != nil {
			return err
		}
	}

	// Insert the new like
	insertCommentLikeSQL := `INSERT INTO comment_likes(user_id, comment_id, comment_is_like) VALUES (?, ?, ?)`
	statement, err := db.Prepare(insertCommentLikeSQL)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(commentLike.UserID, commentLike.CommentID, commentLike.IsLike)
	if err != nil {
		log.Printf("Error executing statement: %v", err)
		return err
	}

	return nil
}

func InsertPostLike(db *sql.DB, postLike *PostLike) error {
	// Check if the user has already liked the post
	var existingLike bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM post_likes WHERE user_id = ? AND post_id = ?)", postLike.UserID, postLike.PostID).Scan(&existingLike)
	if err != nil {
		return err
	}

	if existingLike {
		// Delete the existing like
		_, err = db.Exec("DELETE FROM post_likes WHERE user_id = ? AND post_id = ?", postLike.UserID, postLike.PostID)
		if err != nil {
			return err
		}
		return nil
	}

	// Check if the user has already disliked the post
	var existingDislike bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM post_dislikes WHERE user_id = ? AND post_id = ?)", postLike.UserID, postLike.PostID).Scan(&existingDislike)
	if err != nil {
		return err
	}

	if existingDislike {
		// Delete the existing dislike
		_, err = db.Exec("DELETE FROM post_dislikes WHERE user_id = ? AND post_id = ?", postLike.UserID, postLike.PostID)
		if err != nil {
			return err
		}
	}

	// Insert the new like
	insertPostLikeSQL := `INSERT INTO post_likes(user_id, post_id, post_is_like) VALUES (?, ?, ?)`
	statement, err := db.Prepare(insertPostLikeSQL)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(postLike.UserID, postLike.PostID, postLike.IsLike)
	if err != nil {
		log.Printf("Error executing statement: %v", err)
		return err
	}

	return nil
}

