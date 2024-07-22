package main

import (
	// "context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"

	forum "forum/functions"

)

var database *sql.DB

func main() {
	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("temp"))))
	http.Handle("/main.css", http.FileServer(http.Dir("temp")))
	http.Handle("/login.css", http.FileServer(http.Dir("temp")))
	http.Handle("/com.css", http.FileServer(http.Dir("temp")))
	http.Handle("/reg.css", http.FileServer(http.Dir("temp")))
	http.Handle("/home.css", http.FileServer(http.Dir("temp")))
	http.Handle("/error.css", http.FileServer(http.Dir("temp")))

	// Handle dynamic requests
	http.HandleFunc("/WebServer", forum.WebServer)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		forum.Mainpage(w, r, database)
	})
	http.HandleFunc("/main", func(w http.ResponseWriter, r *http.Request) {
		forum.ParseMain(w, r, database)
	})
	http.HandleFunc("/doRegister", func(w http.ResponseWriter, r *http.Request) {
		forum.HandleReg(w, r, database)
	})
	http.HandleFunc("/doLogin", func(w http.ResponseWriter, r *http.Request) {
			forum.HandleLog(w, r, database)
		})

	http.HandleFunc("/doLogout", func(w http.ResponseWriter, r *http.Request) {
		forum.Logout(w, r, database)
	})

	http.HandleFunc("/createP", func(w http.ResponseWriter, r *http.Request) {
		forum.CreatePost(w, r, database)
	})
	http.HandleFunc("/createC", func(w http.ResponseWriter, r *http.Request) {
		forum.CreateComment(w, r, database)
	})
	http.HandleFunc("/feedback", func(w http.ResponseWriter, r *http.Request) {
		forum.FeedbackHandler(w, r, database)
	})
	http.HandleFunc("/like-post", func(w http.ResponseWriter, r *http.Request) {
		forum.HandleLikePost(w, r, database)
	})
	http.HandleFunc("/dislike-post", func(w http.ResponseWriter, r *http.Request) {
		forum.HandleDislikePost(w, r, database)
	})
	http.HandleFunc("/like-comment", func(w http.ResponseWriter, r *http.Request) {
		forum.HandleLikeComment(w, r, database)
	})
	http.HandleFunc("/dislike-comment", func(w http.ResponseWriter, r *http.Request) {
		forum.HandleDislikeComment(w, r, database)
	})
	

	var err error
	database, err = sql.Open("sqlite3", "./temp/forum.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// Execute the schema SQL file
	err = forum.ExecuteSQLFile(database, "functions/schema.sql")
	if err != nil {
		log.Fatalf("Error executing SQL file: %v", err)
	}

	// Open or create the log file in append mode
    logFile, err := os.Create("log.txt")
    if err != nil {
        log.Fatal("Error opening log file:", err)
    }
    defer logFile.Close()
    log.SetOutput(logFile)

	// Start the web server
	log.Println("Starting server on :8800")
	fmt.Println("Starting server on :8800")
	err = http.ListenAndServe(":8800", nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	
}
