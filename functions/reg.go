package forum

import (
	"net/http"
	"strings"
    "log"
    "html/template"
    "database/sql"
)

func HandleReg(w http.ResponseWriter, r *http.Request, db *sql.DB) {
    var errorMessage string

    if r.Method == http.MethodPost {
        email := r.FormValue("email")
        username := r.FormValue("username")
        password := r.FormValue("password")

		//Check for Non-ASCII characters in username
		eng := Ascii(username)
		if eng != nil {
			HandleError(w, &Error{Err: 400, ErrStr: "Error 400 found"})
			return
		}

		//Check for Non-ASCII characters in email
		eng = Ascii(email)
		if eng != nil {
			HandleError(w, &Error{Err: 400, ErrStr: "Error 400 found"})
			return
		}

		//Check for Non-ASCII characters in password
		eng = Ascii(password)
		if eng != nil {
			HandleError(w, &Error{Err: 400, ErrStr: "Error 400 found"})
			return
		}

        if strings.Contains(email, " ") {
            errorMessage = "Email cannot contain spaces"
        } else if strings.Contains(username, " ") {
            errorMessage = "Username cannot contain spaces"
        } else if strings.Contains(password, " ") {
            errorMessage = "Password cannot contain spaces"
        } else if !strings.Contains(email, ".") {
            errorMessage = "Invalid email format"			
		}

        if errorMessage == "" {
            log.Printf("Received form data: email=%s, username=%s, password=%s\n", email, username, password)

			
            // Populate the User struct with form data
            user := &User{
                Email:    email,
                Username: username,
                Password: password,
            }

            // Insert the new user into the database
            err := InsertUser(db, user)
            if err != nil {
                if err.Error() == "user with this email already exists" {
                    errorMessage = "This email is already taken!"
                } else {
                    errorMessage = "This username is already taken!"
                }
            } else {
                http.Redirect(w, r, "/doLogin", http.StatusSeeOther)
            }
        }
    }

    // Parse the HTML template file
    tmpl, err := template.ParseFiles("temp/regPage.html")
    if err != nil {
        HandleError(w, &Error{Err: 500, ErrStr: "Error 500 found"})
        return
    }

    data := struct {
        ErrorMessage   string
    }{
        ErrorMessage:   errorMessage,
    }

    // Render the template with the data
    tmpl.Execute(w, data)
}
