package forum

import (
    "database/sql"
    "errors"
    // "log"
)

func InsertUser(db *sql.DB, email, username, password string) error {
    // Check if the user already exists
    existingUser, err := GetUserByEmail(db, email)
    if err != nil {
        return err
    }
    if existingUser != nil {
        return errors.New("user with this email already exists")
    }

    // Insert the new user
    insertUserSQL := `INSERT INTO users(email, username, password) VALUES (?, ?, ?)`
    statement, err := db.Prepare(insertUserSQL)
    if err != nil {
        return err
    }
    defer statement.Close()

    _, err = statement.Exec(email, username, password)
    if err != nil {
        return err
    }

    return nil
}

func GetUserByEmail(db *sql.DB, email string) (*User, error) {
    user := &User{}
    query := `SELECT id, email, username, password FROM users WHERE email = ?`
    row := db.QueryRow(query, email)
    err := row.Scan(&user.ID, &user.Email, &user.Username, &user.Password)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil // No user found with the given email
        }
        return nil, err // Some other error occurred
    }
    return user, nil
}
