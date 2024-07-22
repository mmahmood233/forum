package forum

import (
	"html/template"
	"log"
	"net/http"
	"database/sql"
)

func Mainpage(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if IsLoggedIn(r, db) {
        http.Redirect(w, r, "/main", http.StatusSeeOther)
        return
    }
	if r.URL.Path!= "/" {
		HandleError(w, &Error{Err: 404, ErrStr: "Error 404 found"})
		return
	}
	tmp1, err := template.ParseFiles("temp/home.html")
	if err != nil {
		HandleError(w, &Error{Err: 500, ErrStr: "Error 500 found"})
		return
	}
	tmp1.Execute(w, nil)
}

func ParseMain(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get the selected category from the form value
	selectedCategory := r.FormValue("catCont2")
	filter := r.FormValue("filter")


	// Retrieve posts from the database
	postsWithUsers, err := GetPosts(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if the user is logged in
	isLoggedIn := IsLoggedIn(r, db)
	
	var loggedInUser *User
	if isLoggedIn {
        session, _ := GetSession(r, db)
        loggedInUser, _ = GetUserByID(db, session.UserID)
    }

	// Filter posts based on the selected category
	var filteredPosts []struct {
		Post       Post
		User       User
		Comments   []Comment
		Categories []Category
	}

	for _, postWithUser := range postsWithUsers {
        if filter == "myCreatedPosts" {
            if isLoggedIn && loggedInUser != nil && postWithUser.Post.UserID == loggedInUser.UserID {
                filteredPosts = append(filteredPosts, postWithUser)
            }
        } else if filter == "myLikedPosts" {
            if isLoggedIn && loggedInUser != nil {
                var userLiked bool
                err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM post_likes WHERE user_id = ? AND post_id = ?)", loggedInUser.UserID, postWithUser.Post.PostID).Scan(&userLiked)
                if err != nil {
                    log.Printf("Error checking if user liked post: %v", err)
                    continue
                }
                if userLiked {
                    filteredPosts = append(filteredPosts, postWithUser)
                }
            }
        } else if selectedCategory == "None" {
            if len(postWithUser.Categories) == 0 || (len(postWithUser.Categories) == 1 && postWithUser.Categories[0].CatName == "None") {
                filteredPosts = append(filteredPosts, postWithUser)
            }
        } else if selectedCategory == "" || CategoryMatches(postWithUser.Categories, selectedCategory) {
            filteredPosts = append(filteredPosts, postWithUser)
        }
    }

	// Parse the HTML template file
	tmpl, err := template.ParseFiles("temp/main.html")
	if err != nil {
		HandleError(w, &Error{Err: 500, ErrStr: "Error 500 found"})
		return
	}

 // Define and initialize the anonymous struct
	templateData := struct {
		Posts []struct {
			Post       Post
			User       User
			Comments   []Comment
			Categories []Category
		}
		IsLoggedIn       bool
		SelectedCategory string
		Filter           string
		LoggedInUser     *User
	}{
		Posts:            filteredPosts,
		IsLoggedIn:       isLoggedIn,
		SelectedCategory: selectedCategory,
		Filter:           filter,
		LoggedInUser:     loggedInUser,
	}

	// Render the template with the data
	tmpl.Execute(w, templateData)
}
