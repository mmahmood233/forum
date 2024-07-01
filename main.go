package main

import (
    "database/sql"
    "fmt"
    "html/template"
    "io/ioutil"
    "log"
    "net/http"
    "os"

    _ "modernc.org/sqlite"

    forum "forum/functions"
)

var database *sql.DB

func main() {
    // Serve static files
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("temp"))))

    // Handle dynamic requests
    http.HandleFunc("/WebServer", forum.WebServer)

    http.HandleFunc("/", parseMain)

    http.HandleFunc("/doRegister", handleReg)
    http.HandleFunc("/doLogin", handleLog)


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
    var loginMessage string

    if r.Method == http.MethodPost {
        email := r.FormValue("email2")
        password := r.FormValue("password2")

        log.Printf("Received form data: email=%s, password=%s\n", email, password)

        user, err := forum.ValByEmail(database, email)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        if user == nil {
            loginMessage = "No user found with this email"
        } else {
            // Check if the password matches
            if user.Password == password {
                loginMessage = "Login successful!"
            } else {
                loginMessage = "Invalid password"
            }
        }
    }

    // Parse the HTML template file
    tmpl, err := template.ParseFiles("temp/loginPage.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Define and initialize the anonymous struct
    data := struct {
        LoginMessage string
    }{
        LoginMessage: loginMessage,
    }

    // Render the template with the data
    tmpl.Execute(w, data)
}

func parseMain(w http.ResponseWriter, r *http.Request) {
    tmpl, err := template.ParseFiles("temp/mainPage.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    tmpl.Execute(w, nil)
}