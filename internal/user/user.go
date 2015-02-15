package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/destinygg/website2/internal/redis"
	"github.com/gorilla/context"
)

type User struct {
	Username string
	Userid   int64
	Features []string
}

var sessioncookie = regexp.MustCompile(`^[a-z0-9]{10, 30}$`)

func GetFromContext(r *http.Request) (*User, bool) {
	u, ok := context.GetOk(r, "user")
	if !ok {
		return nil, ok
	}

	ret, ok := u.(*User)
	return ret, ok
}

func GetFromRequest(r *http.Request) (*User, error) {
	sessionid, err := r.Cookie("sid")
	if err != nil || !sessioncookie.MatchString(sessionid.Value) {
		// not an error, dont log it
		return nil, nil
	}

	conn := rds.GetRedisConnFromContext(r)
	authdata, err := rds.GetBytes(conn, fmt.Sprintf("CHAT:session-%v", sessionid.Value))
	if err != nil || len(authdata) == 0 {
		return nil, fmt.Errorf("No sessiondata found |%v| |%v|", err, authdata)
	}

	var tu struct {
		Username string   `json:"username"`
		Userid   string   `json:"userId"`
		Features []string `json:"features"`
	}

	err = json.Unmarshal(authdata, &tu)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal sessiondata |%v| |%v|", err, authdata)
	}

	uid, err := strconv.ParseInt(tu.Userid, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse number |%v| |%v|", err, tu.Userid)
	}

	u := &User{
		Username: tu.Username,
		Userid:   uid,
		Features: tu.Features,
	}

	return u, nil
}
