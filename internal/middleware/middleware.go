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
	"github.com/guregu/kami"
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

func RegisterPanicHandler() {
	kami.PanicHandler = recover
}

func recover(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	reason := kami.Exception(ctx)
	erp.Recover(reason, w, r)
}
