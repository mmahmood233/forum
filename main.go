package main

import (
	// "context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
	"errors"
	"crypto/rand"
	"strconv"

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

	// Handle dynamic requests
	http.HandleFunc("/WebServer", forum.WebServer)

	http.HandleFunc("/", parseMain)
	http.HandleFunc("/registered", parseReg)

	http.HandleFunc("/doRegister", handleReg)
	http.HandleFunc("/doLogin", handleLog)
	// http.Handle("/doLogin", sessionMiddleware(http.HandlerFunc(handleLog)))
	http.HandleFunc("/createP", createPost)
	http.HandleFunc("/createC", createComment)
	http.HandleFunc("/feedback", feedbackHandler)

	// http.HandleFunc("/like-post", handleLikePost)
	// http.HandleFunc("/dislike-post", handleDislikePost)
	// http.HandleFunc("/like-comment", handleLikeComment)
	// http.HandleFunc("/dislike-comment", handleDislikeComment)

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
            Type    string `json:"type"`
            ID      int    `json:"id"`
            IsPost  bool   `json:"isPost"`
            UserID  int    `json:"userID"`
        }
        err := json.NewDecoder(r.Body).Decode(&feedback)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        switch feedback.Type {
        case "like":
            if feedback.IsPost {
                InsertPostLike(feedback.UserID, feedback.ID)
            } else {
                InsertCommentLike(feedback.UserID, feedback.ID)
            }
        case "dislike":
            if feedback.IsPost {
                InsertPostDislike(feedback.UserID, feedback.ID)
            } else {
                InsertCommentDislike(feedback.UserID, feedback.ID)
            }
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"status": "success"})
    } else {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
    }
}

func InsertCommentLike(userID int, commentID int) {
    _, err := database.Exec(`INSERT INTO comment_likes (user_id, comment_id, comment_is_like) VALUES (?, ?, ?)`, userID, commentID, true)
    if err != nil {
        log.Fatal(err)
    }
}

func InsertCommentDislike(userID int, commentID int) {
    _, err := database.Exec(`INSERT INTO comment_dislikes (user_id, comment_id, comment_is_dislike) VALUES (?, ?, ?)`, userID, commentID, true)
    if err != nil {
        log.Fatal(err)
    }
}

func InsertPostLike(userID int, postID int) {
    _, err := database.Exec(`INSERT INTO post_likes (user_id, post_id, post_is_like) VALUES (?, ?, ?)`, userID, postID, true)
    if err != nil {
        log.Fatal(err)
    }
}

func InsertPostDislike(userID int, postID int) {
    _, err := database.Exec(`INSERT INTO post_dislikes (user_id, post_id, post_is_dislike) VALUES (?, ?, ?)`, userID, postID, true)
    if err != nil {
        log.Fatal(err)
    }
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
			UserID: user.UserID,
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




func parseMain(w http.ResponseWriter, r *http.Request) {
	// Retrieve posts from the database
	posts, err := getPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse the HTML template file
	tmpl, err := template.ParseFiles("temp/registered.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the template with posts data
	tmpl.Execute(w, posts)
}

func parseReg(w http.ResponseWriter, r *http.Request) {
	// Retrieve posts from the database
	posts, err := getPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse the HTML template file
	tmpl, err := template.ParseFiles("temp/registered.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the template with posts data
	tmpl.Execute(w, posts)
}

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


func createPost(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        sessionObj, err := getSession(r)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        postContent := r.FormValue("postCont")

        // Create a new Post struct
        post := &forum.Post{
            UserID:      sessionObj.UserID,
            PostContent: postContent,
            CreatedAt:   time.Now(),
        }

        // Insert the post into the database
        err = insertPost(post)
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

func insertPost(post *forum.Post) error {
	stmt, err := database.Prepare("INSERT INTO posts (user_id, post_content, post_created_at) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(post.UserID, post.PostContent, post.CreatedAt)
	if err != nil {
		return err
	}

	return nil
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
	Post     forum.Post
	User     forum.User
	Comments []forum.Comment
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
		Post     forum.Post
		User     forum.User
		Comments []forum.Comment
	}

	for rows.Next() { //for loop until no more rows
		var p forum.Post
		var u forum.User

		if err := rows.Scan(&p.PostID, &p.UserID, &p.PostContent, &p.CreatedAt, &u.Username); err != nil {
			return nil, err
		}

		comments, err := getCommentsByPostID(p.PostID)
		if err != nil {
			return nil, err
		}

		postsWithUsers = append(postsWithUsers, struct {
			Post     forum.Post
			User     forum.User
			Comments []forum.Comment
		}{p, u, comments})
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
