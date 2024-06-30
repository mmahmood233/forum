package forum

import (
	"fmt"
	"net/http"
)

func WebServer(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Server started on http://localhost:8800")
}