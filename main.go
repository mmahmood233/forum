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
    http.Handle("/mainPage.html", http.FileServer(http.Dir("temp")))

    http.HandleFunc("/WebServer", forum.WebServer)

    http.HandleFunc("/", handleReg)

    database, err := sql.Open("sqlite", "./temp/forum.db")
    if err != nil {
        log.Fatal(err)
    }
    defer database.Close()

    err = executeSQL(database, "schema.sql")
    if err != nil {
        log.Fatalf("Error executing SQL file: %v", err)
    }

    log.Println("Starting server on :8800")
    err = http.ListenAndServe(":8800", nil)
    if err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}

func executeSQL(db *sql.DB, filepath string) error {
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
    var successMessage string //for testing

    if r.Method == http.MethodPost { 
        //getting data from html
        email := r.FormValue("email")
        username := r.FormValue("username")
        password := r.FormValue("password")

        log.Printf("Received form data: email=%s, username=%s, password=%s\n", email, username, password)

        //updating the struct
        user := &forum.User{
            Email:    email,
            Username: username,
            Password: password,
        }

        err := forum.InsertUser(database, user)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        fmt.Printf("User Struct: %+v\n", user)

        successMessage = "Registration successful!"
    }

    tmpl, err := template.ParseFiles("temp/mainPage.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    //used for testing for now only
    data := struct {
        SuccessMessage string
    }{
        SuccessMessage: successMessage,
    }

    tmpl.Execute(w, data)
}
