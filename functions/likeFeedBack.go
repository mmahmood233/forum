package forum

import (
	"encoding/json"
	"net/http"
	"database/sql"
)

type Feedback struct {
	FeedbackType string `json:"forum"`
}

// var db *sql.DB

func FeedbackHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method == http.MethodPost {
		var feedback struct {
			Type   string `json:"type"`
			ID     int    `json:"id"`
			IsPost bool   `json:"isPost"`
			UserID int    `json:"userID"`
		}
		err := json.NewDecoder(r.Body).Decode(&feedback)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch feedback.Type {
		case "like":
			if feedback.IsPost {
				postLike := &PostLike{
					UserID: feedback.UserID,
					PostID: feedback.ID,
					IsLike: true,
				}
				err = InsertPostLike(db, postLike)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				commentLike := &CommentLike{
					UserID:    feedback.UserID,
					CommentID: feedback.ID,
					IsLike:    true,
				}
				err = InsertCommentLike(db, commentLike)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		case "dislike":
			if feedback.IsPost {
				postdisLike := &PostDislike{
					UserID:    feedback.UserID,
					PostID:    feedback.ID,
					IsDislike: true,
				}
				err = InsertPostDislike(db, postdisLike)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				commentDisLike := &CommentDislike{
					UserID:    feedback.UserID,
					CommentID: feedback.ID,
					IsDislike: true,
				}
			
				err = InsertCommentDislike(db, commentDisLike)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				
			}
		}	
	}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	} 
}