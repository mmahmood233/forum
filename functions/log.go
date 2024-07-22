package forum

import (
    "database/sql"
    "net/http"
    "log"
    "time"
	"html/template"
)

var session = make(map[string]*Session)

func HandleLog(w http.ResponseWriter, r *http.Request, db *sql.DB) {
    var errorMessage string
    if r.Method == http.MethodPost {
        identifier := r.FormValue("identifier")
        password := r.FormValue("password2")

        log.Printf("Received form data: user=%s, password=%s\n", identifier, password)

        // Authenticate the user (e.g., check the email/username and password)
        user, err := ValByEmailOrUsername(db, identifier)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        if user == nil || user.Password != password {
            errorMessage = "Invalid username/email or password"
        } else {
            // Create a new session for the user
            sessionID := CreateSession(w, user.UserID, db)

            // Store the user ID in the session
            session[sessionID] = &Session{
                UserID:    user.UserID,
                ExpiresAt: time.Now().Add(time.Hour * 24),
            }

            // Redirect the user to the home page or another page
            http.Redirect(w, r, "/main", http.StatusSeeOther)
            return
        }
    }

    // Parse the HTML template file
    tmpl, err := template.ParseFiles("temp/loginPage.html")
    if err != nil {
        HandleError(w, &Error{Err: 500, ErrStr: "Error 500 found"})
        return
    }

    data := struct {
        ErrorMessage string
    }{
        ErrorMessage: errorMessage,
    }

    // Render the template with the data
    tmpl.Execute(w, data)
}

func Logout(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
			err := RemoveSessionDB(sessionID, db)
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

func IsLoggedIn(r *http.Request, db *sql.DB) bool {
	// Get the session from the request
	session, err := GetSession(r, db)
	if err != nil {
		return false
	}

	// Check if the session is valid
	if session == nil || session.ExpiresAt.Before(time.Now()) {
		return false
	}

	return true

}

func GetUserByID(db *sql.DB, userID int) (*User, error) {
    user := &User{}
    query := `SELECT user_id, email, username FROM users WHERE user_id = ?`
    err := db.QueryRow(query, userID).Scan(&user.UserID, &user.Email, &user.Username)
    if err != nil {
        return nil, err
    }
    return user, nil
}