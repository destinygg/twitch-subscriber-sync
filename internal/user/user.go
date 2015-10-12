/***
  This file is part of destinygg/user.

  Copyright (c) 2015 Peter Sztan <sztanpet@gmail.com>

  destinygg/user is free software; you can redistribute it and/or modify it
  under the terms of the GNU Lesser General Public License as published by
  the Free Software Foundation; either version 3 of the License, or
  (at your option) any later version.

  destinygg/user is distributed in the hope that it will be useful, but
  WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
  Lesser General Public License for more details.

  You should have received a copy of the GNU Lesser General Public License
  along with destinygg/user; If not, see <http://www.gnu.org/licenses/>.
***/

package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/destinygg/website2/internal/redis"
	"golang.org/x/net/context"
)

var sessioncookie = regexp.MustCompile(`^[a-z0-9]{10, 30}$`)
var contextKey *int
var ErrNotFound = fmt.Errorf("user: session not found")

func init() {
	contextKey = new(int)
}

type User struct {
	Username string
	Userid   int64
	Features []string
}

func (u *User) StoreInContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey, u)
}

func GetFromContext(ctx context.Context) (*User, bool) {
	u, ok := ctx.Value(contextKey).(*User)
	if !ok {
		return nil, ok
	}

	return u, ok
}

func GetFromRequest(ctx context.Context, r *http.Request) (*User, error) {
	return getFromRedis(ctx, r, "CHAT:session-%v")
}
func getAdminFromRequest(ctx context.Context, r *http.Request) (*User, error) {
	return getFromRedis(ctx, r, "CHAT:adminsession-%v")
}

func getFromRedis(ctx context.Context, r *http.Request, keyfmt string) {
	sessionid, err := r.Cookie("sid")
	if err != nil || !sessioncookie.MatchString(sessionid.Value) {
		// not an error, dont log it
		return nil, nil
	}

	conn := rds.GetRedisConnFromContext(ctx)
	defer conn.Return()

	rk := fmt.Sprintf(keyfmt, sessionid.Value)
	authdata, err := rds.GetBytes(conn, rk)
	if err != nil || len(authdata) == 0 {
		return nil, ErrNotFound
	}

	var tu struct {
		Username string   `json:"username"`
		Userid   string   `json:"userId"`
		Features []string `json:"features"`
	}

	err = json.Unmarshal(authdata, &tu)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal session %v |%v| |%v|", rk, err, authdata)
	}

	uid, err := strconv.ParseInt(tu.Userid, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse number for session %v |%v| |%v|", rk, err, tu.Userid)
	}

	u := &User{
		Username: tu.Username,
		Userid:   uid,
		Features: tu.Features,
	}

	return u, nil
}
