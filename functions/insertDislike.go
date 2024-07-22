package forum

import (
	"database/sql"
	"log"
)

func InsertCommentDislike(db *sql.DB, commentDislike *CommentDislike) error {
	// Check if the user has already disliked the comment
	var existingDislike bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM comment_dislikes WHERE user_id = ? AND comment_id = ?)", commentDislike.UserID, commentDislike.CommentID).Scan(&existingDislike)
	if err != nil {
		return err
	}

	if existingDislike {
		// Delete the existing dislike
		_, err = db.Exec("DELETE FROM comment_dislikes WHERE user_id = ? AND comment_id = ?", commentDislike.UserID, commentDislike.CommentID)
		if err != nil {
			return err
		}
		return nil
	}

	// Check if the user has already liked the comment
	var existingLike bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM comment_likes WHERE user_id = ? AND comment_id = ?)", commentDislike.UserID, commentDislike.CommentID).Scan(&existingLike)
	if err != nil {
		return err
	}

	if existingLike {
		// Delete the existing like
		_, err = db.Exec("DELETE FROM comment_likes WHERE user_id = ? AND comment_id = ?", commentDislike.UserID, commentDislike.CommentID)
		if err != nil {
			return err
		}
	}

	// Insert the new dislike
	insertCommentDislikeSQL := `INSERT INTO comment_dislikes(user_id, comment_id, comment_is_dislike) VALUES (?, ?, ?)`
	statement, err := db.Prepare(insertCommentDislikeSQL)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(commentDislike.UserID, commentDislike.CommentID, commentDislike.IsDislike)
	if err != nil {
		log.Printf("Error executing statement: %v", err)
		return err
	}

	return nil
}

func InsertPostDislike(db *sql.DB, postDislike *PostDislike) error {
	// Check if the user has already disliked the post
	var existingDislike bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM post_dislikes WHERE user_id = ? AND post_id = ?)", postDislike.UserID, postDislike.PostID).Scan(&existingDislike)
	if err != nil {
		return err
	}

	if existingDislike {
		// Delete the existing dislike
		_, err = db.Exec("DELETE FROM post_dislikes WHERE user_id = ? AND post_id = ?", postDislike.UserID, postDislike.PostID)
		if err != nil {
			return err
		}
		return nil
	}

	// Check if the user has already liked the post
	var existingLike bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM post_likes WHERE user_id = ? AND post_id = ?)", postDislike.UserID, postDislike.PostID).Scan(&existingLike)
	if err != nil {
		return err
	}

	if existingLike {
		// Delete the existing like
		_, err = db.Exec("DELETE FROM post_likes WHERE user_id = ? AND post_id = ?", postDislike.UserID, postDislike.PostID)
		if err != nil {
			return err
		}
	}

	// Insert the new dislike
	insertPostDislikeSQL := `INSERT INTO post_dislikes(user_id, post_id, post_is_dislike) VALUES (?, ?, ?)`
	statement, err := db.Prepare(insertPostDislikeSQL)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(postDislike.UserID, postDislike.PostID, postDislike.IsDislike)
	if err != nil {
		log.Printf("Error executing statement: %v", err)
		return err
	}

	return nil
}