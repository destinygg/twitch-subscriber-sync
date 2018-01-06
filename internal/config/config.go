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
	"os"
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

type Metrics struct {
	URL      string `toml:"url"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type TwitchScrape struct {
	OAuthToken  string `toml:"oauthtoken"`
	ClientID    string `toml:"clientid"`
	GetSubURL   string `toml:"getsuburl"`
	AddSubURL   string `toml:"addsuburl"`
	ModSubURL   string `toml:"modsuburl"`
	ReSubURL    string `toml:"resuburl"`
	SubURL      string `toml:"suburl"`
	PollMinutes int64  `toml:"pollminutes"`
	Password    string `toml:"password"`
	Channel     string `toml:"channel"`
	ChannelID 	string `toml:"channelid"`
}

type AppConfig struct {
	Website      `toml:"website"`
	Debug        `toml:"debug"`
	Database     `toml:"database"`
	Redis        `toml:"redis"`
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

[metrics]
url="http://localhost:8083"
username=""
password=""

[twitchscrape]
# oauthtoken is used to request the subs from the twitch api and for the
# password to the twitch irc server,
# requires scopes: channel_subscriptions channel_check_subscription chat_login
oauthtoken="generateone"
clientid="generateone"
getsuburl="http://127.0.0.1/api/twitchsubscriptions"
addsuburl="http://127.0.0.1/api/addtwitchsubscription"
modsuburl="http://127.0.0.1/api/twitchsubscriptions"
resuburl="http://127.0.0.1/api/twitchresubscription"
suburl="http://127.0.0.1/api/twitch/subscribe"
# how many minutes between syncing the subs over
pollminutes=60
channel="destiny"
channelid="18074328"
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

func FromContext(ctx context.Context) *AppConfig {
	cfg, _ := ctx.Value("appconfig").(*AppConfig)
	return cfg
}
