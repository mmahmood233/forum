package forum

import (
    "net/http"
    "strings"
    "time"
	"log"
	"html/template"
    "database/sql"
)

func CreatePost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
    var errorMessage string

    if r.Method == http.MethodPost {
        sessionObj, err := GetSession(r, db)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        postContent := strings.TrimSpace(r.FormValue("postCont"))
        if postContent == "" {
            errorMessage = "No post content"
        } else {
            categoryNames := r.Form["catCont"]
            if len(categoryNames) == 0 {
                categoryNames = []string{"None"}
            }

            // Create a new Post struct
            post := &Post{
                UserID:      sessionObj.UserID,
                PostContent: postContent,
				CreatedAt:   time.Now(),
            }

            // Insert the post into the database
            lastInsertID, err := InsertPost(post, db)
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }

			log.Printf("New post created - ID: %d, Content: %s, Categories: %v", lastInsertID, postContent, categoryNames)

            // Insert categories for the post
            for _, categoryName := range categoryNames {
                category := &Category{
                    CatName: categoryName,
                    PostID:  int(lastInsertID),
                }
                err = InsertCategory(category, db)
                if err != nil {
                    http.Error(w, err.Error(), http.StatusInternalServerError)
                    return
                }
            }

            http.Redirect(w, r, "/main", http.StatusSeeOther)
            return
        }
    }

    tmpl, err := template.ParseFiles("temp/comPage.html")
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


func InsertPost(post *Post, db *sql.DB) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO posts (user_id, post_content, post_created_at) VALUES (?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(post.UserID, post.PostContent, post.CreatedAt)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func GetPosts(db *sql.DB) ([]struct {
    Post       Post
    User       User
    Comments   []Comment
    Categories []Category
}, error) {
    query := `
        SELECT p.post_id, p.user_id, p.post_content, p.post_created_at, u.username,
               (SELECT COUNT(*) FROM post_likes WHERE post_id = p.post_id) as like_count,
               (SELECT COUNT(*) FROM post_dislikes WHERE post_id = p.post_id) as dislike_count
        FROM posts p
        JOIN users u ON p.user_id = u.user_id
    `

    rows, err := db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var postsWithUsers []struct {
        Post       Post
        User       User
        Comments   []Comment
        Categories []Category
    }

    for rows.Next() {
        var p Post
        var u User

        var createdAtStr string
		if err := rows.Scan(&p.PostID, &p.UserID, &p.PostContent, &createdAtStr, &u.Username, &p.LikeCount, &p.DislikeCount); err != nil {
			return nil, err
		}
		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		p.CreatedAt = createdAt



        comments, err := GetCommentsByPostID(p.PostID, db)
        if err != nil {
            return nil, err
        }

        categories, err := GetCategoriesByPostID(p.PostID, db)
        if err != nil {
            return nil, err
        }
		if len(categories) == 0 {
            categories = append(categories, Category{CatName: "None"})
        }

        postsWithUsers = append(postsWithUsers, struct {
            Post       Post
            User       User
            Comments   []Comment
            Categories []Category
        }{p, u, comments, categories})
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return postsWithUsers, nil
}