package main

import (
	"database/sql"

	"github.com/destinygg/website2/internal/db"
	"github.com/destinygg/website2/internal/debug"
	"github.com/destinygg/website2/internal/redis"
	"github.com/tideland/godm/v3/redis"
	"golang.org/x/net/context"
)

func initBans(ctx context.Context, unbanchan chan string) {
	db := db.GetFromContext(ctx)
	rdb := rds.GetFromContext(ctx)
	rds.SetupSubscribe(rdb, "unbanuserid", func(result *redis.PublishedValue) {
		userid, err := result.Value.Uint64()
		if err != nil {
			d.D("Error parsing message as uint64:", userid, err)
			return
		}

		stmt, err := db.Prepare(`
			SELECT authDetail
			FROM dfl_users_auth
			WHERE
				userId       = ? AND
				authProvider = 'twitch'
			ORDER BY modifiedDate DESC
			LIMIT 1
		`)

		if err != nil {
			d.DF(1, "err: %v userid: %v", err, userid)
		}
		defer stmt.Close()

		var nick sql.NullString
		err = stmt.QueryRow(userid).Scan(&nick)
		if err != nil || !nick.Valid {
			d.DF(1, "err: %v nick valid: %v", err, nick.Valid)
			return
		}

		go (func(s string) {
			unbanchan <- s
		})(nick.String)
	})
}
