package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	// "crypto/rand"
	"strconv"

	_ "modernc.org/sqlite"

	forum "forum/functions"

	"github.com/gorilla/sessions"
)

var database *sql.DB

func main() {
	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("temp"))))

	// Handle dynamic requests
	http.HandleFunc("/WebServer", forum.WebServer)

	http.HandleFunc("/", parseMain)
	http.HandleFunc("/registered", parseReg)

	http.HandleFunc("/doRegister", handleReg)
	// http.HandleFunc("/doLogin", handleLog)
	http.Handle("/doLogin", sessionMiddleware(http.HandlerFunc(handleLog)))
	http.HandleFunc("/createP", createPost)
	http.HandleFunc("/createC", createComment)

	// http.HandleFunc("/createP", createPost)

	// Initialize the database
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
		session, err := getSession(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set the user ID in the session data
		session.Values["user_id"] = user.UserID

		// Set the session expiration time (e.g., 24 hours)
		session.Options.MaxAge = 24 * 60 * 60

		// Save the session
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
	tmpl, err := template.ParseFiles("temp/mainPage.html")
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


// func getPosts() ([]struct {
//     Post forum.Post
//     User forum.User
// }, error) {
//     query := `
//         SELECT p.user_id, p.post_content, p.post_created_at, u.username
//         FROM posts p
//         JOIN users u ON p.user_id = u.user_id
//     `

//     rows, err := database.Query(query)
//     if err != nil {
//         return nil, err
//     }
//     defer rows.Close()

//     var postsWithUsers []struct {
//         Post forum.Post
//         User forum.User
//     }

//     for rows.Next() {
//         var p forum.Post
//         var u forum.User

//         if err := rows.Scan(&p.UserID, &p.PostContent, &p.CreatedAt, &u.Username); err != nil {
//             return nil, err
//         }

//         postsWithUsers = append(postsWithUsers, struct {
//             Post forum.Post
//             User forum.User
//         }{p, u})
//     }

//     if err := rows.Err(); err != nil {
//         return nil, err
//     }

//     return postsWithUsers, nil
// }

// func createKey() ([]byte, error) {
//     key := make([]byte, 32)
//     _, err := rand.Read(key)
//     if err != nil {
//         return nil, err
//     }
//     return key, nil
// }

var (
	// Create a new session store
	store       = sessions.NewCookieStore([]byte("your-secret-key"))
	sessionName = "forum-session"
)

func sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the session from the request
		session, err := store.Get(r, "session-name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Save the session data in the request context
		ctx := context.WithValue(r.Context(), "session", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func createPost(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Get the session data to identify the user
		session, err := getSession(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Get the form values
		postContent := r.FormValue("postCont")

		// Create a new Post struct
		post := &forum.Post{
			UserID:      session.Values["user_id"].(int),
			PostContent: postContent,
			CreatedAt:   time.Now(),
		}

		// Insert the post into the database
		err = insertPost(post)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Redirect to the main page
		http.Redirect(w, r, "/registered", http.StatusSeeOther)
		return
	}

	// Parse the HTML template file
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

func getSession(r *http.Request) (*sessions.Session, error) {
	// Get the session from the request
	session, err := store.Get(r, sessionName)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// func openDB() (*sql.DB, error) {
//     if database == nil {
//         var err error
//         database, err = sql.Open("sqlite", "./temp/forum.db")
//         if err != nil {
//             return nil, err
//         }
//     }
//     return database, nil
// }

// func createComment(w http.ResponseWriter, r *http.Request) {
//     if r.Method == http.MethodPost {
//         session, err := getSession(r)
//         if err != nil {
//             http.Error(w, err.Error(), http.StatusInternalServerError)
//             return
//         }
//         comContent := r.FormValue("commentCont")

//                // Print the session data
//                fmt.Printf("Session Data:\n")
//                fmt.Printf("  User ID: %v\n", session.Values["user_id"])
//                fmt.Printf("  Post ID: %v\n", session.Values["post_id"])

//                // Print the comment content
//                fmt.Printf("Comment Content: %s\n", comContent)

//         comment := &forum.Comment{
//             UserID:      session.Values["user_id"].(int),
//             PostID:      session.Values["post_id"].(int),
//             CommentContent: comContent,
//             CreatedAt:   time.Now(),
//     }

//     fmt.Printf("Comment Struct:\n")
//     fmt.Printf("  User ID: %d\n", comment.UserID)
//     fmt.Printf("  Post ID: %d\n", comment.PostID)
//     fmt.Printf("  Comment Content: %s\n", comment.CommentContent)
//     fmt.Printf("  Created At: %v\n", comment.CreatedAt)

//     err = insertComment(comment)
//     if err != nil {
//         http.Error(w, err.Error(), http.StatusInternalServerError)
//         return
//     }

//     http.Redirect(w, r, "/", http.StatusSeeOther)
//     return
// }
// }

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

		comContent := r.FormValue("commentCont")
		postID := r.URL.Query().Get("postID")

		// Convert postID from string to int
		postIDInt, err := strconv.Atoi(postID)
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		comment := &forum.Comment{
			UserID:         session.Values["user_id"].(int),
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
