package forum

import (
    "database/sql"
    "errors"
    "log"
)

func InsertUser(db *sql.DB, user *User) error {
    // Check if the user already exists
    existingUser, err := ValByEmail(db, user.Email)
    if err != nil {
        return err
    }
    if existingUser != nil {
        return errors.New("user with this email already exists")
    }

    insertUserSQL := `INSERT INTO users(email, username, password) VALUES (?, ?, ?)`
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

    user.ID = int(userID) //to update struct with the id

    log.Printf("New user registered with ID: %d", user.ID)

    return nil
}

func ValByEmail(db *sql.DB, email string) (*User, error) {
    user := &User{}
    query := `SELECT id, email, username, password FROM users WHERE email = ?`
    row := db.QueryRow(query, email)
    err := row.Scan(&user.ID, &user.Email, &user.Username, &user.Password)
    if err != nil {
        if err == sql.ErrNoRows { //when no rows returned
            return nil, nil // no user found with the email
        }
        return nil, err //something else
    }
    return user, nil
}
