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

type User struct {
	Username string
	Userid   int64
	Features []string
}

var sessioncookie = regexp.MustCompile(`^[a-z0-9]{10, 30}$`)

func GetFromContext(ctx context.Context) (*User, bool) {
	u, ok := ctx.Value("user").(*User)
	if !ok {
		return nil, ok
	}

	return u, ok
}

func GetFromRequest(ctx context.Context, r *http.Request) (*User, error) {
	sessionid, err := r.Cookie("sid")
	if err != nil || !sessioncookie.MatchString(sessionid.Value) {
		// not an error, dont log it
		return nil, nil
	}

	conn := rds.GetRedisConnFromContext(ctx)
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
