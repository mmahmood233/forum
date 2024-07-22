package forum

import (
	"log"
	"net/http"
	"strconv"
	"encoding/json"
	"database/sql"
)

// var database *sql.DB

func HandleDislikePost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get the session and post ID
	session, err := GetSession(r, db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	postID, err := strconv.Atoi(r.FormValue("postID"))
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	user, err := GetUserByID(db, session.UserID)
    if err != nil {
        log.Printf("Error getting user info: %v", err)
    } else {
        log.Printf("User %s (ID: %d) disliked post %d", user.Username, session.UserID, postID)
    }

	// Check if the user has already disliked the post
	var existingDislike bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM post_dislikes WHERE user_id = ? AND post_id = ?)", session.UserID, postID).Scan(&existingDislike)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if existingDislike {
		// Delete the existing dislike
		_, err = db.Exec("DELETE FROM post_dislikes WHERE user_id = ? AND post_id = ?", session.UserID, postID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Redirect or return a response
	} else {
		// Delete any existing like for the post
		err = DeletePostLike(db, session.UserID, postID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Insert the dislike
		postDislike := &PostDislike{
			UserID:    session.UserID,
			PostID:    postID,
			IsDislike: true,
		}
		err = InsertPostDislike(db, postDislike)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	var dislikeCount int
    err = db.QueryRow("SELECT COUNT(*) FROM post_dislikes WHERE post_id = ?", postID).Scan(&dislikeCount)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

	// http.Redirect(w, r, "/registered", http.StatusSeeOther)


    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "count":   dislikeCount,
    })

	// Redirect or return a response
	// http.Redirect(w, r, "/registered", http.StatusSeeOther)
}



func HandleDislikeComment(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get the session and comment ID
	session, err := GetSession(r, db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	commentID, err := strconv.Atoi(r.FormValue("commentID"))
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	user, err := GetUserByID(db, session.UserID)
    if err != nil {
        log.Printf("Error getting user info: %v", err)
    } else {
        log.Printf("User %s (ID: %d) disliked comment %d", user.Username, session.UserID, commentID)
    }

	// Check if the user has already disliked the comment
	var existingDislike bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM comment_dislikes WHERE user_id = ? AND comment_id = ?)", session.UserID, commentID).Scan(&existingDislike)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if existingDislike {
		// Delete the existing dislike
		_, err = db.Exec("DELETE FROM comment_dislikes WHERE user_id = ? AND comment_id = ?", session.UserID, commentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Redirect or return a response

	} else {
		// Delete any existing like for the comment
		err = DeleteCommentLike(db, session.UserID, commentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Insert the dislike
		commentDislike := &CommentDislike{
			UserID:    session.UserID,
			CommentID: commentID,
			IsDislike: true,
		}
		err = InsertCommentDislike(db, commentDislike)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	var dislikeCount int
    err = db.QueryRow("SELECT COUNT(*) FROM comment_dislikes WHERE comment_id = ?", commentID).Scan(&dislikeCount)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

	// http.Redirect(w, r, "/registered", http.StatusSeeOther)


    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "count":   dislikeCount,
    })

	// Redirect or return a response
	// http.Redirect(w, r, "/registered", http.StatusSeeOther)
}