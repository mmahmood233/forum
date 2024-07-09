package main

import (
	// "context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "modernc.org/sqlite"

	forum "forum/functions"

	"github.com/gorilla/sessions"
)

var database *sql.DB
var sessionStore = sessions.NewCookieStore([]byte("your-secret-key-here"))
var session = make(map[string]*forum.Session)

func main() {
	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("temp"))))
	http.Handle("/registered.css", http.FileServer(http.Dir("temp")))
	http.Handle("/login.css", http.FileServer(http.Dir("temp")))
	http.Handle("/com.css", http.FileServer(http.Dir("temp")))
    http.Handle("/reg.css", http.FileServer(http.Dir("temp")))
	http.Handle("/main.css", http.FileServer(http.Dir("temp")))
    http.Handle("/error.css", http.FileServer(http.Dir("temp")))


	// Handle dynamic requests
	http.HandleFunc("/WebServer", forum.WebServer)

	http.HandleFunc("/", parseMain)
	// http.HandleFunc("/registered", parseReg)

	http.HandleFunc("/doRegister", handleReg)
	http.HandleFunc("/doLogin", handleLog)
	// http.Handle("/doLogin", sessionMiddleware(http.HandlerFunc(handleLog)))
	http.HandleFunc("/doLogout", logout)
	http.HandleFunc("/createP", createPost)
	// http.HandleFunc("createP", ChooseCategory)
	http.HandleFunc("/createC", createComment)
	http.HandleFunc("/feedback", feedbackHandler)

	http.HandleFunc("/like-post", handleLikePost)
	http.HandleFunc("/dislike-post", handleDislikePost)
	http.HandleFunc("/like-comment", handleLikeComment)
	http.HandleFunc("/dislike-comment", handleDislikeComment)

	// http.HandleFunc("/createP", createPost)

	// Initialize the database
	// session = make(map[string]*forum.Session)

	var err error
	database, err = sql.Open("sqlite", "./temp/forum.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// Execute the schema SQL file
	err = executeSQLFile(database, "schema.sql")
	if err != nil {
		log.Fatalf("Error executing SQL file: %v", err)
	}

	// Start the web server
	log.Println("Starting server on :8800")
	err = http.ListenAndServe(":8800", nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func executeSQLFile(db *sql.DB, filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	sqlBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	sqlCommands := string(sqlBytes)
	_, err = db.Exec(sqlCommands)
	if err != nil {
		return err
	}

	return nil
}

// new
type Feedback struct {
	FeedbackType string `json:"forum"`
}

// var db *sql.DB

func feedbackHandler(w http.ResponseWriter, r *http.Request) {
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
				postLike := &forum.PostLike{
					UserID: feedback.UserID,
					PostID: feedback.ID,
					IsLike: true,
				}
				err = InsertPostLike(database, postLike)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				commentLike := &forum.CommentLike{
					UserID: feedback.UserID,
					CommentID: feedback.ID,
					IsLike: true,
				}
				err = InsertCommentLike(database, commentLike)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		case "dislike":
			if feedback.IsPost {
				postdisLike := &forum.PostDislike{
					UserID: feedback.UserID,
					PostID: feedback.ID,
					IsDislike: true,
				}
				err = InsertPostDislike(database, postdisLike)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				commentDisLike := &forum.CommentDislike{
					UserID: feedback.UserID,
					CommentID: feedback.ID,
					IsDislike: true,
				}
				err = InsertCommentDislike(database, commentDisLike)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func InsertCommentLike(db *sql.DB, commentLike *forum.CommentLike) error {
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
	// Check if the user has already liked the comment
	var existingLike bool
	wewe := db.QueryRow("SELECT EXISTS(SELECT 1 FROM comment_likes WHERE user_id = ? AND comment_id = ?)", commentLike.UserID, commentLike.CommentID).Scan(&existingLike)
	if wewe != nil {
		return err
	}

	if existingLike {
		// Delete the existing like
		_, wewe = db.Exec("DELETE FROM comment_likes WHERE user_id = ? AND comment_id = ?", commentLike.UserID, commentLike.CommentID)
		if wewe != nil {
			return err
		}
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
	SQL := `INSERT INTO comment_likes(user_id, comment_id, comment_is_like) VALUES (?, ?, ?)`
	_, err = db.Exec(SQL, commentLike.UserID, commentLike.CommentID, commentLike.IsLike)
	return err
}

func InsertCommentDislike(db *sql.DB, commentDislike *forum.CommentDislike) error {
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
	// Check if the user has already disliked the comment
	var existingDislike bool
	wewe := db.QueryRow("SELECT EXISTS(SELECT 1 FROM comment_dislikes WHERE user_id = ? AND comment_id = ?)", commentDislike.UserID, commentDislike.CommentID).Scan(&existingDislike)
	if wewe != nil {
		return err
	}

	if existingDislike {
		// Delete the existing dislike
		_, err = db.Exec("DELETE FROM comment_dislikes WHERE user_id = ? AND comment_id = ?", commentDislike.UserID, commentDislike.CommentID)
		if err != nil {
			return err
		}
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
	SQL := `INSERT INTO comment_dislikes(user_id, comment_id, comment_is_dislike) VALUES (?, ?, ?)`
	_, err = db.Exec(SQL, commentDislike.UserID, commentDislike.CommentID, commentDislike.IsDislike)
	return err
}

func InsertPostDislike(db *sql.DB, postDislike *forum.PostDislike) error {
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
	// Check if the user has already disliked the post
	var existingDislike bool
	wewe := db.QueryRow("SELECT EXISTS(SELECT 1 FROM post_dislikes WHERE user_id = ? AND post_id = ?)", postDislike.UserID, postDislike.PostID).Scan(&existingDislike)
	if wewe != nil {
		return err
	}

	if existingDislike {
		// Delete the existing dislike
		_, err = db.Exec("DELETE FROM post_dislikes WHERE user_id = ? AND post_id = ?", postDislike.UserID, postDislike.PostID)
		if err != nil {
			return err
		}
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
	SQL := `INSERT INTO post_dislikes(user_id, post_id, post_is_dislike) VALUES (?, ?, ?)`
	_, err = db.Exec(SQL, postDislike.UserID, postDislike.PostID, postDislike.IsDislike)
	return err
}

func InsertPostLike(db *sql.DB, postLike *forum.PostLike) error {
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
	// Check if the user has already liked the post
	var existingLike bool
	wewe := db.QueryRow("SELECT EXISTS(SELECT 1 FROM post_likes WHERE user_id = ? AND post_id = ?)", postLike.UserID, postLike.PostID).Scan(&existingLike)
	if wewe != nil {
		return err
	}

	if existingLike {
		// Delete the existing like
		_, err = db.Exec("DELETE FROM post_likes WHERE user_id = ? AND post_id = ?", postLike.UserID, postLike.PostID)
		if err != nil {
			return err
		}
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
	SQL := `INSERT INTO post_likes(user_id, post_id, post_is_like) VALUES (?, ?, ?)`
	_, err = db.Exec(SQL, postLike.UserID, postLike.PostID, postLike.IsLike)
	return err
}

//---

func handleReg(w http.ResponseWriter, r *http.Request) {
	var successMessage string

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")

		log.Printf("Received form data: email=%s, username=%s, password=%s\n", email, username, password)

		// Populate the User struct with form data
		user := &forum.User{
			Email:    email,
			Username: username,
			Password: password,
		}

		// Insert the new user into the database
		err := forum.InsertUser(database, user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Print the user struct to verify the ID has been updated
		fmt.Printf("User Struct after insertion: %+v\n", user)

		successMessage = "Registration successful!"
	}

	// Parse the HTML template file
	tmpl, err := template.ParseFiles("temp/regPage.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Define and initialize the anonymous struct
	data := struct {
		SuccessMessage string
	}{
		SuccessMessage: successMessage,
	}

	// Render the template with the data
	tmpl.Execute(w, data)
}

func handleLog(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email2")
		password := r.FormValue("password2")

		log.Printf("Received form data: email=%s, password=%s\n", email, password)

		// Authenticate the user (e.g., check the email and password)
		user, err := forum.ValByEmail(database, email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if user == nil || user.Password != password {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// Create a new session for the user
		sessionID := createSession(w, user.UserID)

		// Store the user ID in the session
		session[sessionID] = &forum.Session{
			UserID:    user.UserID,
			ExpiresAt: time.Now().Add(time.Hour * 24),
		}

		// Redirect the user to the home page or another page
		http.Redirect(w, r, "/registered", http.StatusSeeOther)
		return
	}

	// Parse the HTML template file
	tmpl, err := template.ParseFiles("temp/loginPage.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the template
	tmpl.Execute(w, nil)
}

func logout(w http.ResponseWriter, r *http.Request) {
	// Retrieve session ID from the cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Printf("No session cookie found: %v", err)
	} else {
		sessionID := cookie.Value
		log.Printf("Session ID to delete: %s", sessionID)

		if sessionID != "" {
			// Attempt to delete session from in-memory store
			delete(session, sessionID)
			log.Printf("Session deleted from memory: %s", sessionID)

			// Attempt to delete session from the database
			err := removeSessionDB(sessionID)
			if err != nil {
				log.Printf("Error deleting session from database: %v", err)
			} else {
				log.Printf("Session deleted from database: %s", sessionID)
			}
		}

		// Clear the session cookie
		http.SetCookie(w, &http.Cookie{
			Name:    "session_id",
			Value:   "",
			Expires: time.Unix(0, 0),
		})
		log.Printf("Session cookie cleared: %s", sessionID)
	}

	// Redirect to the login page
	http.Redirect(w, r, "/doLogin", http.StatusSeeOther)
}

func removeSessionDB(sessionID string) error {
	log.Printf("Attempting to remove session from database: %s", sessionID)

	// Prepare the SQL query
	stmt, err := database.Prepare("DELETE FROM sessions WHERE session_id = ?")
	if err != nil {
		log.Printf("Error preparing delete statement: %v", err)
		return err
	}
	defer stmt.Close()

	// Execute the query
	res, err := stmt.Exec(sessionID)
	if err != nil {
		log.Printf("Error executing delete statement: %v", err)
		return err
	}

	// Check how many rows were affected
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error fetching rows affected: %v", err)
		return err
	}

	if rowsAffected == 0 {
		log.Printf("No rows affected, session ID may not exist: %s", sessionID)
	} else {
		log.Printf("Rows affected: %d", rowsAffected)
	}

	return nil
}

func isLoggedIn(r *http.Request) bool {
	// Get the session from the request
	session, err := getSession(r)
	if err != nil {
		return false
	}

	// Check if the session is valid
	if session == nil || session.ExpiresAt.Before(time.Now()) {
		return false
	}

	return true

}

func parseMain(w http.ResponseWriter, r *http.Request) {
	// Get the selected category from the form value
	selectedCategory := r.FormValue("catCont2")

	// Retrieve posts from the database
	postsWithUsers, err := getPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if the user is logged in
	isLoggedIn := isLoggedIn(r)

	// Filter posts based on the selected category
	var filteredPosts []struct {
		Post       forum.Post
		User       forum.User
		Comments   []forum.Comment
		Categories []forum.Category
	}

	for _, postWithUser := range postsWithUsers {
		if selectedCategory == "" || categoryMatches(postWithUser.Categories, selectedCategory) {
			filteredPosts = append(filteredPosts, postWithUser)
		}
	}

	// Parse the HTML template file
	tmpl, err := template.ParseFiles("temp/registered.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Define and initialize the anonymous struct
	templateData := struct {
		Posts []struct {
			Post       forum.Post
			User       forum.User
			Comments   []forum.Comment
			Categories []forum.Category
		}
		IsLoggedIn       bool
		SelectedCategory string
	}{
		Posts:            filteredPosts,
		IsLoggedIn:       isLoggedIn,
		SelectedCategory: selectedCategory,
	}

	// Render the template with the data
	tmpl.Execute(w, templateData)
}

func categoryMatches(categories []forum.Category, selectedCategory string) bool {
	if selectedCategory == "" {
		return true
	}
	for _, category := range categories {
		if category.CatName == selectedCategory {
			return true
		}
	}
	return false
}

// func parseReg(w http.ResponseWriter, r *http.Request) {
// 	// Retrieve posts from the database
// 	posts, err := getPosts()
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Parse the HTML template file
// 	tmpl, err := template.ParseFiles("temp/registered.html")
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Render the template with posts data
// 	tmpl.Execute(w, posts)
// }

func getSessionID() string {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(key)
}

var sessionData = make(map[string]*forum.Session)

func createSession(w http.ResponseWriter, userID int) string {
	sessionID := getSessionID()
	sessionObj := &forum.Session{
		SessionID: sessionID,
		UserID:    userID,
		ExpiresAt: time.Now().Add(24 * time.Hour), // Set the expiration time (e.g., 24 hours)
	}
	sessionData[sessionID] = sessionObj

	cookie := http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)

	// Insert the session data into the database
	_, err := database.Exec("INSERT INTO sessions (session_id, user_id, expires_at) VALUES (?, ?, ?)", sessionID, userID, sessionObj.ExpiresAt)
	if err != nil {
		log.Printf("Error inserting session data: %v", err)
		// You may want to handle the error more gracefully here
	}

	return sessionID
}

func getSession(r *http.Request) (*forum.Session, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return nil, err
	}
	sessionID := cookie.Value

	// Query the database for the session data
	var userID int
	var expiresAt time.Time
	err = database.QueryRow("SELECT user_id, expires_at FROM sessions WHERE session_id = ?", sessionID).Scan(&userID, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid session")
		}
		return nil, err
	}

	sessionObj := &forum.Session{
		SessionID: sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	return sessionObj, nil
}

func ChooseCategory(w http.ResponseWriter, r *http.Request) {
	// Retrieve posts from the database
	posts, err := getPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	choosenCat := r.FormValue("catCont")
	category := &forum.Category{
		CatName: choosenCat,
		PostID:  posts[0].Post.PostID,
	}
	err = InsertCategory(category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		http.Error(w, "No posts found", http.StatusBadRequest)
		return
	}
}

func InsertCategory(cat *forum.Category) error {
	tx, err := database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var catID int64
	err = tx.QueryRow("SELECT cat_id FROM categories WHERE cat_name = ?", cat.CatName).Scan(&catID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Category doesn't exist, insert it
			result, err := tx.Exec("INSERT INTO categories (cat_name) VALUES (?)", cat.CatName)
			if err != nil {
				return err
			}
			catID, err = result.LastInsertId()
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Insert the post-category relationship
	_, err = tx.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", cat.PostID, catID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func createPost(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		sessionObj, err := getSession(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		postContent := r.FormValue("postCont")
		categoryName := r.FormValue("catCont")

		// Create a new Post struct
		post := &forum.Post{
			UserID:      sessionObj.UserID,
			PostContent: postContent,
			CreatedAt:   time.Now(),
		}

		// Insert the post into the database
		lastInsertID, err := insertPost(post)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Insert the category for the post
		category := &forum.Category{
			CatName: categoryName,
			PostID:  int(lastInsertID),
		}
		err = InsertCategory(category)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/registered", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles("temp/comPage.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the template
	tmpl.Execute(w, nil)
}

func insertPost(post *forum.Post) (int64, error) {
	stmt, err := database.Prepare("INSERT INTO posts (user_id, post_content, post_created_at) VALUES (?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(post.UserID, post.PostContent, post.CreatedAt)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func insertComment(comment *forum.Comment) error {
	stmt, err := database.Prepare("INSERT INTO comments (user_id, post_id, comment_content, comment_created_at) VALUES (?, ?, ?, ?)")
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

func createComment(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		session, err := getSession(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userID := session.UserID

		comContent := r.FormValue("commentCont")
		postID := r.URL.Query().Get("postID")

		// Convert postID from string to int
		postIDInt, err := strconv.Atoi(postID)
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		comment := &forum.Comment{
			UserID:         userID,
			PostID:         postIDInt,
			CommentContent: comContent,
			CreatedAt:      time.Now(),
		}

		err = insertComment(comment)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/registered", http.StatusSeeOther)
		return
	}
}

func getPosts() ([]struct {
	Post       forum.Post
	User       forum.User
	Comments   []forum.Comment
	Categories []forum.Category
}, error) {
	query := `
        SELECT p.post_id, p.user_id, p.post_content, p.post_created_at, u.username
        FROM posts p
        JOIN users u ON p.user_id = u.user_id
    `

	rows, err := database.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var postsWithUsers []struct {
		Post       forum.Post
		User       forum.User
		Comments   []forum.Comment
		Categories []forum.Category
	}

	for rows.Next() {
		var p forum.Post
		var u forum.User

		if err := rows.Scan(&p.PostID, &p.UserID, &p.PostContent, &p.CreatedAt, &u.Username); err != nil {
			return nil, err
		}

		comments, err := getCommentsByPostID(p.PostID)
		if err != nil {
			return nil, err
		}

		categories, err := getCategoriesByPostID(p.PostID)
		if err != nil {
			return nil, err
		}

		postsWithUsers = append(postsWithUsers, struct {
			Post       forum.Post
			User       forum.User
			Comments   []forum.Comment
			Categories []forum.Category
		}{p, u, comments, categories})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return postsWithUsers, nil
}

func getCommentsByPostID(postID int) ([]forum.Comment, error) {
	query := `
        SELECT c.comment_content, c.comment_created_at, u.username
        FROM comments c
        JOIN users u ON c.user_id = u.user_id
        WHERE c.post_id = ?
    `
	rows, err := database.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []forum.Comment
	for rows.Next() {
		var c forum.Comment
		var u forum.User

		if err := rows.Scan(&c.CommentContent, &c.CreatedAt, &u.Username); err != nil {
			return nil, err
		}

		c.Username = u.Username
		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func getCategoriesByPostID(postID int) ([]forum.Category, error) {
	query := `
        SELECT c.cat_id, c.cat_name
        FROM categories c
        JOIN post_categories pc ON c.cat_id = pc.category_id
        WHERE pc.post_id = ?
    `
	rows, err := database.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []forum.Category
	for rows.Next() {
		var c forum.Category
		if err := rows.Scan(&c.CatID, &c.CatName); err != nil {
			return nil, err
		}
		c.PostID = postID
		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func handleLikePost(w http.ResponseWriter, r *http.Request) {
	// Get the session and post ID
	session, err := getSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	postID, err := strconv.Atoi(r.FormValue("postID"))
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Check if the user has already liked the post
	var existingLike bool
	err = database.QueryRow("SELECT EXISTS(SELECT 1 FROM post_likes WHERE user_id = ? AND post_id = ?)", session.UserID, postID).Scan(&existingLike)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if existingLike {
		// Delete the existing like
		_, err = database.Exec("DELETE FROM post_likes WHERE user_id = ? AND post_id = ?", session.UserID, postID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Redirect or return a response
		http.Redirect(w, r, "/registered", http.StatusSeeOther)
		return
	} else {
		// Delete any existing dislike for the post
		err = DeletePostDislike(database, session.UserID, postID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Insert the like
		postLike := &forum.PostLike{
			UserID: session.UserID,
			PostID: postID,
			IsLike: true,
		}
		err = InsertPostLike(database, postLike)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Redirect or return a response
	http.Redirect(w, r, "/registered", http.StatusSeeOther)
}

func handleDislikePost(w http.ResponseWriter, r *http.Request) {
	// Get the session and post ID
	session, err := getSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	postID, err := strconv.Atoi(r.FormValue("postID"))
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Check if the user has already disliked the post
	var existingDislike bool
	err = database.QueryRow("SELECT EXISTS(SELECT 1 FROM post_dislikes WHERE user_id = ? AND post_id = ?)", session.UserID, postID).Scan(&existingDislike)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if existingDislike {
		// Delete the existing dislike
		_, err = database.Exec("DELETE FROM post_dislikes WHERE user_id = ? AND post_id = ?", session.UserID, postID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Redirect or return a response
		http.Redirect(w, r, "/registered", http.StatusSeeOther)
		return
	} else {
		// Delete any existing like for the post
		err = DeletePostLike(database, session.UserID, postID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Insert the dislike
		postDislike := &forum.PostDislike{
			UserID:    session.UserID,
			PostID:    postID,
			IsDislike: true,
		}
		err = InsertPostDislike(database, postDislike)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Redirect or return a response
	http.Redirect(w, r, "/registered", http.StatusSeeOther)
}

func handleLikeComment(w http.ResponseWriter, r *http.Request) {
	// Get the session and comment ID
	session, err := getSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	commentID, err := strconv.Atoi(r.FormValue("commentID"))
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	// Check if the user has already liked the comment
	var existingLike bool
	err = database.QueryRow("SELECT EXISTS(SELECT 1 FROM comment_likes WHERE user_id = ? AND comment_id = ?)", session.UserID, commentID).Scan(&existingLike)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if existingLike {
		// Delete the existing like
		_, err = database.Exec("DELETE FROM comment_likes WHERE user_id = ? AND comment_id = ?", session.UserID, commentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Redirect or return a response
		http.Redirect(w, r, "/registered", http.StatusSeeOther)
		return
	} else {
		// Delete any existing dislike for the comment
		err = DeleteCommentDislike(database, session.UserID, commentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Insert the like
		commentLike := &forum.CommentLike{
			UserID:    session.UserID,
			CommentID: commentID,
			IsLike:    true,
		}
		err = InsertCommentLike(database, commentLike)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Redirect or return a response
	http.Redirect(w, r, "/registered", http.StatusSeeOther)
}

func handleDislikeComment(w http.ResponseWriter, r *http.Request) {
	// Get the session and comment ID
	session, err := getSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	commentID, err := strconv.Atoi(r.FormValue("commentID"))
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	// Check if the user has already disliked the comment
	var existingDislike bool
	err = database.QueryRow("SELECT EXISTS(SELECT 1 FROM comment_dislikes WHERE user_id = ? AND comment_id = ?)", session.UserID, commentID).Scan(&existingDislike)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if existingDislike {
		// Delete the existing dislike
		_, err = database.Exec("DELETE FROM comment_dislikes WHERE user_id = ? AND comment_id = ?", session.UserID, commentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Redirect or return a response
		http.Redirect(w, r, "/registered", http.StatusSeeOther)
		return
	} else {
		// Delete any existing like for the comment
		err = DeleteCommentLike(database, session.UserID, commentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Insert the dislike
		commentDislike := &forum.CommentDislike{
			UserID:    session.UserID,
			CommentID: commentID,
			IsDislike: true,
		}
		err = InsertCommentDislike(database, commentDislike)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Redirect or return a response
	http.Redirect(w, r, "/registered", http.StatusSeeOther)
}

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