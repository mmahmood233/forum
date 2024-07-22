package forum

import (
    "net/http"
    "strconv"
    "log"
    "encoding/json"
    "database/sql"
)

func HandleLikePost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
        log.Printf("User %s (ID: %d) liked post %d", user.Username, session.UserID, postID)
    }

    var existingLike bool
    err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM post_likes WHERE user_id = ? AND post_id = ?)", session.UserID, postID).Scan(&existingLike)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if existingLike {
        _, err = db.Exec("DELETE FROM post_likes WHERE user_id = ? AND post_id = ?", session.UserID, postID)
    } else {
        err = DeletePostDislike(db, session.UserID, postID)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        postLike := &PostLike{
            UserID: session.UserID,
            PostID: postID,
            IsLike: true,
        }
        err = InsertPostLike(db, postLike)
    }

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    var likeCount int
    err = db.QueryRow("SELECT COUNT(*) FROM post_likes WHERE post_id = ?", postID).Scan(&likeCount)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

	// http.Redirect(w, r, "/registered", http.StatusSeeOther)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "count":   likeCount,
    })

	// Redirect or return a response
	// http.Redirect(w, r, "/registered", http.StatusSeeOther)
}

func HandleLikeComment(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
        log.Printf("User %s (ID: %d) liked comment %d", user.Username, session.UserID, commentID)
    }

	// Check if the user has already liked the comment
	var existingLike bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM comment_likes WHERE user_id = ? AND comment_id = ?)", session.UserID, commentID).Scan(&existingLike)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if existingLike {
		// Delete the existing like
		_, err = db.Exec("DELETE FROM comment_likes WHERE user_id = ? AND comment_id = ?", session.UserID, commentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	} else {
		// Delete any existing dislike for the comment
		err = DeleteCommentDislike(db, session.UserID, commentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Insert the like
		commentLike := &CommentLike{
			UserID:    session.UserID,
			CommentID: commentID,
			IsLike:    true,
		}
		err = InsertCommentLike(db, commentLike)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	var likeCount int
    err = db.QueryRow("SELECT COUNT(*) FROM comment_likes WHERE comment_id = ?", commentID).Scan(&likeCount)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

	// http.Redirect(w, r, "/registered", http.StatusSeeOther)


    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "count":   likeCount,
    })

	// Redirect or return a response
	// http.Redirect(w, r, "/registered", http.StatusSeeOther)
}