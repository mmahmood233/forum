package forum

import (
	"database/sql"
	"errors"
	"log"
)

func InsertUser(db *sql.DB, user *User) error {
	// Check if the user already exists
	existingUser, err := ValByEmailOrUsername(db, user.Email)
    if err != nil {
        return err
    }
    if existingUser != nil {
        return errors.New("user with this email already exists")
    }

    existingUser, err = ValByEmailOrUsername(db, user.Username)
    if err != nil {
        return err
    }
    if existingUser != nil {
        return errors.New("user with this username already exists")
    }

	insertUserSQL := `INSERT INTO users(email, username, password1) VALUES (?, ?, ?)`
	statement, err := db.Prepare(insertUserSQL)
	if err != nil {
		return err
	}
	defer statement.Close()

	result, err := statement.Exec(user.Email, user.Username, user.Password)
	if err != nil {
		return err
	}

	userID, err := result.LastInsertId() //to incremnet id automatically
	if err != nil {
		return err
	}

	user.UserID = int(userID) //to update struct with the id

	log.Printf("New user registered with ID: %d", user.UserID)

	return nil
}

func ValByEmailOrUsername(db *sql.DB, input string) (*User, error) {
    user := &User{}
    query := `SELECT user_id, email, username, password1 FROM users WHERE email = ? OR username = ?`
    row := db.QueryRow(query, input, input)
    err := row.Scan(&user.UserID, &user.Email, &user.Username, &user.Password)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil // no user found
        }
        return nil, err
    }
    return user, nil
}