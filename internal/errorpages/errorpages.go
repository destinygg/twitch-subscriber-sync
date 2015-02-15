package erp

import (
	"fmt"
	"net/http"

	"github.com/destinygg/website2/internal/debug"
)

func AuthRequired(w http.ResponseWriter, r *http.Request) {
	// TODO with a proper template and shit
	http.Error(w, "authentication required", http.StatusForbidden)
}

func Recover(reason interface{}, w http.ResponseWriter, r *http.Request) {
	// TODO get a stack trace and as much info as possible, save it under some
	// key and show the user only that key
	// also possibly email about the error
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)

	err := d.NewErrorTrace(5, reason)
	fmt.Fprint(w, err.Error())
}
