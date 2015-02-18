package middleware

import (
	"net/http"

	"github.com/destinygg/website2/internal/debug"
	"github.com/destinygg/website2/internal/errorpages"
	"github.com/destinygg/website2/internal/user"
	"golang.org/x/net/context"
)

func Auth(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	u, err := user.GetFromRequest(ctx, r)
	if u == nil {
		if err != nil {
			d.D(err)
		}

		erp.AuthRequired(w, r)
		return ctx
	} else {
		return context.WithValue(ctx, "user", u)
	}
}

func AdminAuth(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	// TODO
	return ctx
}

func Recover(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	reason := kami.Exception(ctx)
	erp.Recover(reason, w, r)
}
