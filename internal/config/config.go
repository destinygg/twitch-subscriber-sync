package config

import (
	"flag"
	"io"
	"os"

	"code.google.com/p/gcfg"
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
	Environment string
	MerchantID  string
	PublicKey   string
	PrivateKey  string
}

type SMTP struct {
	Addr      string
	Username  string
	Password  string
	FromEmail string
	LogEmail  []string
}

type AppConfig struct {
	Website
	Debug
	Database
	Redis
	Braintree
	SMTP
}

var settingsFile = flag.String("config", "settings.cfg", `path to the config file, it it doesn't exist it will
		be created with default values`)

const sampleconf = `[website]
addr=:80
basehost=www.destiny.gg
cdnhost=cdn.destiny.gg

[debug]
debug=no
logfile=logs/debug.txt

[database]
dsn=user:password@tcp(localhost:3306)/destinygg?loc=UTC&parseTime=true&strict=true&timeout=1s&time_zone='+00:00'
maxidleconnections=128
maxconnections=256

[redis]
addr=localhost:6379
dbindex=0
password=
poolsize=128

[braintree]
environment=production
merchantid=
publickey=
privatekey=

[smtp]
addr=
username=
password=
fromemail=
# where to send error emails to, if there are multiple logemail= lines every one
# of them will receive the emails
logemail=
`

func Init(ctx context.Context) context.Context {
	flag.Parse()
	f, err := os.OpenFile(*settingsFile, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		panic("Could not open " + *settingsFile + " err: " + err.Error())
	}
	defer f.Close()

	// empty? initialize it
	if info, err := f.Stat(); err == nil && info.Size() == 0 {
		io.WriteString(f, sampleconf)
		f.Seek(0, 0)
	}

	cfg := ReadConfig(f)
	return context.WithValue(ctx, "appconfig", cfg)
}

func ReadConfig(f *os.File) *AppConfig {
	ret := &AppConfig{}
	if err := gcfg.ReadInto(ret, f); err != nil {
		panic("Failed to parse config file, err: " + err.Error())
	}

	return ret
}

func GetFromContext(ctx context.Context) *AppConfig {
	cfg, _ := ctx.Value("appconfig").(*AppConfig)
	return cfg
}
