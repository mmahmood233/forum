package forum

import (
    "database/sql"
    "errors"
    "log"
)


// the funcs insertpostlike, dislike, insertcommentlike, dislike aren't working accordingly
//it corresponds to schema.sql  
func InsertPostLike(db *sql.DB, postLike *PostLike) error {
    insertPostLikeSQL := `INSERT INTO post_likes(user_id, post_id, post_is_like) VALUES (?, ?, ?)`
    statement, err := db.Prepare(insertPostLikeSQL)
    if err != nil {
        return err
    }
    defer statement.Close()

    _, err = statement.Exec(postLike.UserID, postLike.PostID, postLike.IsLike)
    if err != nil {
        return err
    }

    return nil
}

// InsertPostDislike inserts a new post dislike into the database
func InsertPostDislike(db *sql.DB, postDislike *PostDislike) error {
    insertPostDislikeSQL := `INSERT INTO post_dislikes(user_id, post_id, post_is_dislike) VALUES (?, ?, ?)`
    statement, err := db.Prepare(insertPostDislikeSQL)
    if err != nil {
        return err
    }
    defer statement.Close()

    _, err = statement.Exec(postDislike.UserID, postDislike.PostID, postDislike.IsDislike)
    if err != nil {
        return err
    }
    return nil
}

// InsertCommentLike inserts a new comment like into the database
func InsertCommentLike(db *sql.DB, commentLike *CommentLike) error {
    insertCommentLikeSQL := `INSERT INTO comment_likes(user_id, comment_id, comment_is_like) VALUES (?, ?, ?)`
    statement, err := db.Prepare(insertCommentLikeSQL)
    if err != nil {
        return err
    }
    defer statement.Close()

    _, err = statement.Exec(commentLike.UserID, commentLike.CommentID, commentLike.IsLike)
    if err != nil {
        return err
    }
    return nil
}

func InsertCommentDislike(db *sql.DB, commentDislike *CommentDislike) error {
    insertCommentDislikeSQL := `INSERT INTO comment_dislikes(user_id, comment_id, comment_is_dislike) VALUES (?, ?, ?)`
    statement, err := db.Prepare(insertCommentDislikeSQL)
    if err != nil {
        return err
    }
    defer statement.Close()

    _, err = statement.Exec(commentDislike.UserID, commentDislike.CommentID, commentDislike.IsDislike)
    if err != nil {
        return err
    }
    return nil
}

func InsertUser(db *sql.DB, user *User) error {
    // Check if the user already exists
    existingUser, err := ValByEmail(db, user.Email)
    if err != nil {
        return err
    }
    if existingUser != nil {
        return errors.New("user with this email already exists")
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

func ValByEmail(db *sql.DB, email string) (*User, error) {
    user := &User{}
    query := `SELECT user_id, email, username, password1 FROM users WHERE email = ?`
    row := db.QueryRow(query, email)
    err := row.Scan(&user.UserID, &user.Email, &user.Username, &user.Password)
    if err != nil {
        if err == sql.ErrNoRows { //when no rows returned
            return nil, nil // no user found with the email
        }
        return nil, err //something else
    }
    return user, nil
}
