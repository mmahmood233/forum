package forum

import (
	"net/http"
	"strconv"
	"strings"
	"time"
	"log"
	"database/sql"
)

// var database *sql.DB

func CreateComment(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method == http.MethodPost {
		session, err := GetSession(r, db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userID := session.UserID

        comContent := strings.TrimSpace(r.FormValue("commentCont"))
		if comContent == "" {
			errorData := &Error{
				Err:     400,
				ErrStr: "Comment content cannot be empty",
			}
			HandleError(w, errorData)
			return
		}
		postID := r.URL.Query().Get("postID")

		// Convert postID from string to int
		postIDInt, err := strconv.Atoi(postID)
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		comment := &Comment{
			UserID:         userID,
			PostID:         postIDInt,
			CommentContent: comContent,
			CreatedAt:      time.Now(),
		}

		err = InsertComment(comment, db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

        // Log the comment information
        user, err := GetUserByID(db, userID)
        if err != nil {
            log.Printf("Error getting user info: %v", err)
        } else {
            log.Printf("New comment added - User: %s (ID: %d), Post ID: %d, Content: %s", user.Username, userID, postIDInt, comContent)
        }

		http.Redirect(w, r, "/main", http.StatusSeeOther)
		return
	}
}

func InsertComment(comment *Comment, db *sql.DB) error {
	stmt, err := db.Prepare("INSERT INTO comments (user_id, post_id, comment_content, comment_created_at) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(comment.UserID, comment.PostID, comment.CommentContent, comment.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

func GetCommentsByPostID(postID int, db *sql.DB) ([]Comment, error) {
    query := `
        SELECT c.comment_id, c.comment_content, c.comment_created_at, u.username,
               (SELECT COUNT(*) FROM comment_likes WHERE comment_id = c.comment_id) as like_count,
               (SELECT COUNT(*) FROM comment_dislikes WHERE comment_id = c.comment_id) as dislike_count
        FROM comments c
        JOIN users u ON c.user_id = u.user_id
        WHERE c.post_id = ?
    `
    rows, err := db.Query(query, postID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var comments []Comment
    for rows.Next() {
        var c Comment

        var createdAtStr string
		if err := rows.Scan(&c.CommentID, &c.CommentContent, &createdAtStr, &c.Username, &c.LikeCount, &c.DislikeCount); err != nil {
			return nil, err
		}
		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		c.CreatedAt = createdAt


        comments = append(comments, c)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return comments, nil
}