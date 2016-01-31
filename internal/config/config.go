/***
  This file is part of destinygg/config.

  Copyright (c) 2015 Peter Sztan <sztanpet@gmail.com>

  destinygg/config is free software; you can redistribute it and/or modify it
  under the terms of the GNU Lesser General Public License as published by
  the Free Software Foundation; either version 3 of the License, or
  (at your option) any later version.

  destinygg/config is distributed in the hope that it will be useful, but
  WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
  Lesser General Public License for more details.

  You should have received a copy of the GNU Lesser General Public License
  along with destinygg/config; If not, see <http://www.gnu.org/licenses/>.
***/

package config

import (
	"flag"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/naoina/toml"
	"golang.org/x/net/context"
)

type Website struct {
	Addr          string `toml:"addr"`
	BaseHost      string `toml:"basehost"`
	CDNHost       string `toml:"cdnhost"`
	PrivateAPIKey string `toml:"privateapikey"`
}

type Debug struct {
	Debug   bool   `toml:"debug"`
	Logfile string `toml:"logfile"`
}

type Database struct {
	DSN                string `toml:"dsn"`
	MaxIdleConnections int    `toml:"maxidleconnections"`
	MaxConnections     int    `toml:"maxconnections"`
}

type Redis struct {
	Addr     string `toml:"addr"`
	Password string `toml:"password"`
	DBIndex  int    `toml:"dbindex"`
	PoolSize int    `toml:"poolsize"`
}

type Braintree struct {
	Environment string `toml:"environment"`
	MerchantID  string `toml:"merchantid"`
	PublicKey   string `toml:"publickey"`
	PrivateKey  string `toml:"privatekey"`
}

type SMTP struct {
	Addr      string   `toml:"addr"`
	Username  string   `toml:"username"`
	Password  string   `toml:"password"`
	FromEmail string   `toml:"fromemail"`
	LogEmail  []string `toml:"logemail"`
}

type Metrics struct {
	URL      string `toml:"url"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type TwitchScrape struct {
	OAuthToken  string `toml:"oauthtoken"`
	GetSubURL   string `toml:"getsuburl"`
	AddSubURL   string `toml:"addsuburl"`
	ModSubURL   string `toml:"modsuburl"`
	ReSubURL    string `toml:"resuburl"`
	PollMinutes int64  `toml:"pollminutes"`
	Addr        string `toml:"addr"`
	Nick        string `toml:"nick"`
	Password    string `toml:"password"`
	Channel     string `toml:"channel"`
}

type AppConfig struct {
	Website      `toml:"website"`
	Debug        `toml:"debug"`
	Database     `toml:"database"`
	Redis        `toml:"redis"`
	Braintree    `toml:"braintree"`
	SMTP         `toml:"smtp"`
	Metrics      `toml:"metrics"`
	TwitchScrape `toml:"twitchscrape"`
}

var settingsFile *string

const sampleconf = `[website]
addr=":80"
basehost="www.destiny.gg"
cdnhost="cdn.destiny.gg"
privateapikey="keepitsecret"

[debug]
debug=false
logfile="logs/debug.txt"

[database]
dsn="user:password@tcp(localhost:3306)/destinygg?loc=UTC&parseTime=true&strict=true&timeout=1s&time_zone='+00:00'"
maxidleconnections=128
maxconnections=256

[redis]
addr="localhost:6379"
dbindex=0
password=""
poolsize=128

[braintree]
environment="production"
merchantid=""
publickey=""
privatekey=""

[smtp]
addr=""
username=""
password=""
fromemail=""
# where to send error emails to, if there are multiple, every one  of them will
# receive the emails, use array notation aka
# logemail=["firstemail@domain.tld", "secondemail@domain.tld"]
logemail=[]

[metrics]
url="http://localhost:8083"
username=""
password=""

[twitchscrape]
# oauthtoken is used to request the subs from the twitch api and for the
# password to the twitch irc server,
# requires scopes: channel_subscriptions channel_check_subscription chat_login
oauthtoken="generateone"
getsuburl="http://127.0.0.1/api/twitchsubscriptions"
addsuburl="http://127.0.0.1/api/addtwitchsubscription"
modsuburl="http://127.0.0.1/api/twitchsubscriptions"
resuburl="http://127.0.0.1/api/twitchresubscription"
# how many minutes between syncing the subs over
pollminutes=60
addr="irc.twitch.tv:6667"
nick="mytwitchuser"
channel="destiny"
`

func Init(ctx context.Context) context.Context {
	settingsFile = flag.String("config", "settings.cfg", `path to the config file, it it doesn't exist it will
			be created with default values`)
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

	cfg := &AppConfig{}
	if err := ReadConfig(f, cfg); err != nil {
		panic("Failed to parse config file, err: " + err.Error())
	}

	return context.WithValue(ctx, "appconfig", cfg)
}

func ReadConfig(r io.Reader, d interface{}) error {
	dec := toml.NewDecoder(r)
	return dec.Decode(d)
}

func WriteConfig(w io.Writer, d interface{}) error {
	enc := toml.NewEncoder(w)
	return enc.Encode(d)
}

func Save(ctx context.Context) error {
	return SafeSave(*settingsFile, *FromContext(ctx))
}

func SafeSave(file string, data interface{}) error {
	dir, err := filepath.Abs(filepath.Dir(file))
	if err != nil {
		return err
	}

	f, err := ioutil.TempFile(dir, "tmpconf-")
	if err != nil {
		return err
	}

	err = WriteConfig(f, data)
	if err != nil {
		return err
	}
	_ = f.Close()

	return os.Rename(f.Name(), file)
}

func FromContext(ctx context.Context) *AppConfig {
	cfg, _ := ctx.Value("appconfig").(*AppConfig)
	return cfg
}
