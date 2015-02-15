package middleware

import (
	"net/http"

	"github.com/destinygg/website2/internal/debug"
	"github.com/destinygg/website2/internal/errorpages"
	"github.com/destinygg/website2/internal/user"
	"github.com/gorilla/context"
)

func Auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, err := user.GetFromRequest(r)
		if u == nil {
			if err != nil {
				d.D(err)
			}

			erp.AuthRequired(w, r)
		} else {
			context.Set(r, "user", u)
			h.ServeHTTP(w, r)
		}
	})
}

func Recover(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer (func() {
			if reason := recover(); reason != nil {
				erp.Recover(reason, w, r)
			}
		})()

		h.ServeHTTP(w, r)
	})
}
