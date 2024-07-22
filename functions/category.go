package forum

import (
    "database/sql"
    "net/http"
)

// var database *sql.DB

func ChooseCategory(w http.ResponseWriter, r *http.Request, db *sql.DB) {
    posts, err := GetPosts(db)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if len(posts) == 0 {
        http.Error(w, "No posts found", http.StatusBadRequest)
        return
    }

    choosenCats := r.Form["catCont[]"]
    if len(choosenCats) == 0 {
        http.Error(w, "No categories selected", http.StatusBadRequest)
        return
    }

    for _, choosenCat := range choosenCats {
        category := &Category{
            CatName: choosenCat,
            PostID:  posts[0].Post.PostID,
        }
        err = InsertCategory(category, db)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    }

    // Redirect or respond with success message
    http.Redirect(w, r, "/main", http.StatusSeeOther)
}

func InsertCategory(cat *Category, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var catID int64
	err = tx.QueryRow("SELECT cat_id FROM categories WHERE cat_name = ?", cat.CatName).Scan(&catID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Category doesn't exist, insert it
			result, err := tx.Exec("INSERT INTO categories (cat_name) VALUES (?)", cat.CatName)
			if err != nil {
				return err
			}
			catID, err = result.LastInsertId()
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Insert the post-category relationship
	_, err = tx.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", cat.PostID, catID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func CategoryMatches(categories []Category, selectedCategory string) bool {
	if selectedCategory == "" {
		return true
	}
	for _, category := range categories {
		if category.CatName == selectedCategory {
			return true
		}
	}
	return false
}

func GetCategoriesByPostID(postID int, db *sql.DB) ([]Category, error) {
	query := `
        SELECT c.cat_id, c.cat_name
        FROM categories c
        JOIN post_categories pc ON c.cat_id = pc.category_id
        WHERE pc.post_id = ?
    `
	rows, err := db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.CatID, &c.CatName); err != nil {
			return nil, err
		}
		c.PostID = postID
		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}
