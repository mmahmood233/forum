package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	// "fmt"

	// "net/http"
	// "html/template"

	_ "modernc.org/sqlite"

	forum "forum/functions"
)

var database *sql.DB


func main() {
    // Serve static files
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("temp"))))

    // Handle dynamic requests
    http.HandleFunc("/WebServer", forum.WebServer)

    http.HandleFunc("/", ParseF)

    http.HandleFunc("/register", handleRegister)


    // Initialize the database
    database, err := sql.Open("sqlite", "./forum.db")
    if err != nil {
        log.Fatal(err)
    }
    defer database.Close()

    // Execute the schema SQL file
    err = executeSQLFile(database, "schema.sql")
    if err != nil {
        log.Fatalf("Error executing SQL file: %v", err)
    }

    // Insert a test user
    // forum.InsertUser(database, getEmail(), getName(), getPass())

    // Retrieve the test user
    // user, err := forum.GetUserByEmail(database, "user@example.com")
    // if err != nil {
    //     log.Fatalf("Error retrieving user: %v", err)
    // }
    // if user != nil {
    //     log.Printf("Retrieved user: %+v", user)
    // }

    // Start the web server
    log.Println("Starting server on :8800")
    err = http.ListenAndServe(":8800", nil)
    if err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
    // fmt.Println(&User)
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

func ParseF(w http.ResponseWriter, r *http.Request) {
    tmpl, err := template.ParseFiles("temp/mainPage.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    tmpl.Execute(w, nil)
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
    log.Println("handleRegister called")

    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    email := r.FormValue("email")
    username := r.FormValue("username")
    password := r.FormValue("password")

    log.Printf("Received form data: email=%s, username=%s, password=%s\n", email, username, password)

    // Populate the User struct with form data
    user := forum.User{
        Email:    email,
        Username: username,
        Password: password,
    }

    // Print the user struct to verify the data
    fmt.Printf("User Struct: %+v\n", user)

    // Insert the new user into the database
    err := forum.InsertUser(database, user.Email, user.Username, user.Password)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/", http.StatusSeeOther)
}