package forum

import (
	"golang.org/x/crypto/bcrypt"
	"fmt"
	"encoding/base64"
	"net/http"
	"time"
	"log"
	"errors"
	"database/sql"
)

func GetSessionID() string {
    timestamp := time.Now().UnixNano()
    data := []byte(fmt.Sprintf("%d", timestamp))
    hashedBytes, err := bcrypt.GenerateFromPassword(data, bcrypt.DefaultCost)
    if err != nil {
        return ""
    }
    return base64.URLEncoding.EncodeToString(hashedBytes)
}

var sessionData = make(map[string]*Session)

func CreateSession(w http.ResponseWriter, userID int, db *sql.DB) string {
    // Delete any existing sessions for this user
    err := DeleteExistingSessionsForUser(userID, db)
    if err != nil {
        log.Printf("Error deleting existing sessions: %v", err)
        // You may want to handle this error more gracefully
    }

    sessionID := GetSessionID()
    sessionObj := &Session{
        SessionID: sessionID,
        UserID:    userID,
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }
    sessionData[sessionID] = sessionObj

    cookie := http.Cookie{
        Name:     "session_id",
        Value:    sessionID,
        Path:     "/",
        HttpOnly: true,
    }

    http.SetCookie(w, &cookie)

    // Insert the new session data into the database
    _, err = db.Exec("INSERT INTO sessions (session_id, user_id, expires_at) VALUES (?, ?, ?)", sessionID, userID, sessionObj.ExpiresAt)
    if err != nil {
        log.Printf("Error inserting session data: %v", err)
        // You may want to handle this error more gracefully
    }	

    return sessionID
}


func DeleteExistingSessionsForUser(userID int, db *sql.DB) error {
    _, err := db.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
    return err
}


func GetSession(r *http.Request, db *sql.DB) (*Session, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		// return nil, err
		return nil, errors.New("invalid session")
	}
	sessionID := cookie.Value

	// Query the database for the session data
	var userID int
	var expiresAt time.Time
	err = db.QueryRow("SELECT user_id, expires_at FROM sessions WHERE session_id = ?", sessionID).Scan(&userID, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid session")
		}
		return nil, err
	}

	sessionObj := &Session{
		SessionID: sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	return sessionObj, nil
}

func RemoveSessionDB(sessionID string, db *sql.DB) error {
	log.Printf("Attempting to remove session from database: %s", sessionID)

	// Prepare the SQL query
	stmt, err := db.Prepare("DELETE FROM sessions WHERE session_id = ?")
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