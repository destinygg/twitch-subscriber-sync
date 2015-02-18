package config

import (
	"io"
	"os"

	//"code.google.com/p/gcfg"
	"golang.org/x/net/context"
)

// TODO tag fields
type Website struct {
	Addr     string
	BaseHost string
	CDNHost  string
}

type Debug struct {
	Debug   bool
	Logfile string
}

type Database struct {
	DSN                string
	MaxIdleConnections int
	MaxConnections     int
}

type Redis struct {
	Addr     string
	Password string
	DBIndex  int
	PoolSize int
}

type Braintree struct {
	MerchantID string
	PublicKey  string
	PrivateKey string
}

type AppConfig struct {
	Website
	Debug
	Database
	Redis
	Braintree
}

const sampleconf = `
[website]
addr=:80
basehost=www.destiny.gg
cdnhost=cdn.destiny.gg

[debug]
debug=no
logfile=logs/debug.txt

[database]
dsn=user:password@tcp(localhost:3306)/destinygg?loc=UTC&parseTime=true&strict=true&timeout=1s&time_zone="+00:00"
maxidleconnections=128
maxconnections=256

[redis]
addr=localhost:6379
dbindex=0
password=
poolsize=20

[braintree]
merchantid=
publickey=
privatekey=
`

func Init(ctx context.Context) context.Context {
	// TODO use a flag for getting the config file name?
	f, err := os.OpenFile("settings.cfg", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		panic("could not open settings.cfg")
	}

	cfg := ReadConfig(f)
	if cfg == nil {
		io.WriteString(f, sampleconf)
		f.Seek(0, 0)
		cfg = ReadConfig(f)
	}

	return context.WithValue(ctx, "appconfig", cfg)
}

func ReadConfig(f *os.File) *AppConfig {
	return nil
}
