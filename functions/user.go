package forum

import (
	"database/sql"
	"errors"
	"log"
)
// var database *sql.DB


// func InsertCommentLike(commentID int, userID int, isLike bool) error {
// 	stmt, err := database.Prepare("INSERT INTO comment_likes (user_id, comment_id, comment_is_like) VALUES (?, ?, ?)")
// 	if err != nil {
// 		return err
// 	}
// 	defer stmt.Close()

// 	_, err = stmt.Exec(userID, commentID, isLike)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func InsertCommentDislike(commentID int, userID int, isDislike bool) error {
// 	stmt, err := database.Prepare("INSERT INTO comment_dislikes (user_id, comment_id, comment_is_dislike) VALUES (?, ?, ?)")
// 	if err != nil {
// 		return err
// 	}
// 	defer stmt.Close()

// 	_, err = stmt.Exec(userID, commentID, isDislike)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func InsertPostLike(postID int, userID int, isLike bool) error {
// 	stmt, err := database.Prepare("INSERT INTO post_likes (user_id, post_id, post_is_like) VALUES (?, ?, ?)")
// 	if err != nil {
// 		return err
// 	}
// 	defer stmt.Close()

// 	_, err = stmt.Exec(userID, postID, isLike)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func InsertPostDislike(postID int, userID int, isDislike bool) error {
// 	stmt, err := database.Prepare("INSERT INTO post_dislikes (user_id, post_id, post_is_dislike) VALUES (?, ?, ?)")
// 	if err != nil {
// 		return err
// 	}
// 	defer stmt.Close()

// 	_, err = stmt.Exec(userID, postID, isDislike)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

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
