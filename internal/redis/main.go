package r

import (
	"fmt"
	"time"

	"github.com/destinygg/website2/internal/debug"
	"github.com/tideland/godm/v3/redis"
	"golang.org/x/net/context"
)

// GetRedisConnFromContext tries to get a redis connection from the redis
// database assigned to the context with the "redisdb" key
// It panics if it cannot get a connection after 3 tries
// The returned connection has to be .Return()-ed after use
func GetRedisConnFromContext(ctx context.Context) *redis.Connection {
	db, ok := ctx.Value("redisdb").(*redis.Database)
	if !ok {
		panic("Redis database not found in the context")
	}

	return GetConn(db)
}

// New sets up the redis database with the given arguments, panics if it cannot
func New(addr string, dbnum int64, pw string, poolsize int) *redis.Database {
	db, err := redis.Open(
		redis.TcpConnection(addr, 1*time.Second),
		redis.Index(int(dbnum), pw),
		redis.PoolSize(poolsize),
	)
	if err != nil {
		d.FBT("Error making the redis pool", err)
	}

	return db
}

// GetConn gets a connection from the database, it retries 3 times before panicing
func GetConn(db *redis.Database) *redis.Connection {
	var err error
	var conn *redis.Connection

	for i := 0; i < 3; i++ {
		conn, err = db.Connection()
		if err == nil {
			return conn
		}
	}

	panic(fmt.Sprintf("Unable to get redis connection, err: %+v", err))
}

// RegisterScript register a script to be run
func RegisterScript(db *redis.Database, script string) string {
	conn := GetConn(db)
	defer conn.Return()

	hash, err := conn.DoString("SCRIPT", "LOAD", script)
	if err != nil {
		d.FBT("Script loading error", err, script)
	}

	return hash
}

// RunScript runs a script identified by the hash returned from RegisterScript
func RunScript(conn *redis.Connection, hash string, args ...interface{}) (*redis.ResultSet, error) {
	t := make([]interface{}, 0, len(args)+2)
	t = append(t, hash)
	t = append(t, 1)
	for _, v := range args {
		t = append(t, v)
	}

	return conn.Do("EVALSHA", t...)
}

func pubsubchan(channel string, dbnum int) string {
	return fmt.Sprintf("%s-%d", channel, dbnum)
}

// SetupSubscribe will set up a redis subscription on the channel suffixed by
// the database number, so given a channel name "foo" and db index 1, the
// channel name will be "foo-1" (because channels are not db local in redis)
// The callback will only be called for non-nil values received on the channel
// It should be run in a go routine
func SetupSubscribe(db *redis.Database, channel string, cb func(*redis.PublishedValue)) {
again:
	sub, err := db.Subscription()
	if err != nil {
		goto again
	}

	//dbnum := redis.GetIndex(db)
	dbnum := 0
	err = sub.Subscribe(pubsubchan(channel, dbnum))
	if err != nil {
		goto again
	}

	for {
		result, err := sub.Pop()
		if err != nil {
			goto again
		}

		if result.Value.IsNil() {
			continue
		}

		cb(result)
	}
}

// Publish sends to the channel
func Publish(db *redis.Database, channel string, msg []byte) error {
	//dbnum := redis.GetDatabase(db)
	dbnum := 0
	conn := GetConn(db)

	_, err := conn.Do("PUBLISH", pubsubchan(channel, dbnum), msg)
	return err
}

// GetBytes gets the data stored at the given key
func GetBytes(conn *redis.Connection, key string) ([]byte, error) {
	result, err := conn.Do("GET", key)
	if err != nil {
		return nil, err
	}

	value, err := result.ValueAt(0)
	if err != nil {
		return nil, err
	}

	return value.Bytes(), err
}
