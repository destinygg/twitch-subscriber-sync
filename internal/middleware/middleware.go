/***
  This file is part of destinygg/middleware.

  Copyright (c) 2015 Peter Sztan <sztanpet@gmail.com>

  destinygg/middleware is free software; you can redistribute it and/or modify it
  under the terms of the GNU Lesser General Public License as published by
  the Free Software Foundation; either version 3 of the License, or
  (at your option) any later version.

  destinygg/middleware is distributed in the hope that it will be useful, but
  WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
  Lesser General Public License for more details.

  You should have received a copy of the GNU Lesser General Public License
  along with destinygg/middleware; If not, see <http://www.gnu.org/licenses/>.
***/

package middleware

import (
	"net/http"

	"github.com/destinygg/website2/internal/debug"
	"github.com/destinygg/website2/internal/errorpages"
	"github.com/destinygg/website2/internal/user"
	"golang.org/x/net/context"
)

func Auth(n chain.Handler) chain.Handler {
	return chain.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		u, err := user.GetFromRequest(ctx, r)
		if u == nil {
			// we do not care about the session not found error, all others
			// are interesting though, so log them
			if err != nil && err != user.ErrNotFound {
				d.D(err)
				errorpages.InternalError(ctx, w, r, struct {
					Err error
				}{
					Err: err,
				})
			} else {
				errorpages.AuthRequired(ctx, w, r)
			}
		} else {
			ctx = context.WithValue(ctx, user.ContextKey, u)
			n.ServeHTTPContext(ctx, w, r)
		}
	})
}

func AdminAuth(n chain.Handler) chain.Handler {
	return chain.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		u, err := user.GetAdminFromRequest(ctx, r)
		if u == nil {
			if err != nil {
				d.D(err)
			}

			errorpages.AdminAuthRequired(ctx, w, r)
		} else {
			ctx = u.StoreInContext(ctx)
			n.ServeHTTPContext(ctx, w, r)
		}
	})
}

func Panic(n chain.Handler) chain.Handler {
	return chain.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		defer func() {
			reason := recover()
			if reason == nil {
				return
			}

			errorpages.Recover(ctx, w, r, reason)
		}()

		n.ServeHTTPContext(ctx, w, r)
	})
}
