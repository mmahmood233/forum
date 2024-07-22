package forum

import (
	"html/template"
	"net/http"
)

func HandleError(w http.ResponseWriter, data *Error) {
	tmpl, err := template.ParseFiles("temp/error.html")
	if err != nil {
		// Render a generic error page if template parsing fails
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if data.Err == 400 {
		w.WriteHeader(http.StatusBadRequest)
	} else if data.Err == 404 {
		w.WriteHeader(http.StatusNotFound)
	} else if data.Err == 500 {
		w.WriteHeader(http.StatusInternalServerError)
	}
	err = tmpl.Execute(w, data)
}