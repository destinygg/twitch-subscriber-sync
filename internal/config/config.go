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
	ClientID     string `toml:"clientid"`
	ClientSecret string `toml:"clientsecret"`
	AccessToken  string `toml:"accesstoken"`
	RefreshToken string `toml:"refreshtoken"`
	GetSubURL    string `toml:"getsuburl"`
	AddSubURL    string `toml:"addsuburl"`
	ModSubURL    string `toml:"modsuburl"`
	ReSubURL     string `toml:"resuburl"`
	SubURL       string `toml:"suburl"`
	PollMinutes  int64  `toml:"pollminutes"`
	Password     string `toml:"password"`
	Channel      string `toml:"channel"`
	ChannelID    string `toml:"channelid"`
}

type AppConfig struct {
	Website      `toml:"website"`
	Debug        `toml:"debug"`
	Database     `toml:"database"`
	Redis        `toml:"redis"`
	Metrics      `toml:"metrics"`
	TwitchScrape `toml:"twitchscrape"`
}

type TwitchTokens struct {
	AccessToken  string `toml:"accesstoken"`
	RefreshToken string `toml:"refreshtoken"`
}

var settingsFile *string
var tokensFile *string

func Init(ctx context.Context) context.Context {
	settingsFile = flag.String("config", "settings.cfg", `path to the config file`)
	tokensFile = flag.String("tokens", "twitchtokens", `path to the tokens file`)
	flag.Parse()
	cfg := ReadSettingsFile()
	ReadTokensFile(&cfg.TwitchScrape, false)
	return context.WithValue(ctx, "appconfig", cfg)
}

func ReadSettingsFile() *AppConfig {
	f, err := os.OpenFile(*settingsFile, os.O_RDONLY, 0660)
	defer f.Close()
	if err != nil {
		panic("Could not open " + *settingsFile + " err: " + err.Error())
	}
	cfg := &AppConfig{}
	if err := ReadConfig(f, cfg); err != nil {
		panic("Failed to parse config file, err: " + err.Error())
	}
	return cfg
}

func ReadTokensFile(cfg *TwitchScrape, overwrite bool) {
	f, err := os.OpenFile(*tokensFile, os.O_CREATE|os.O_RDWR, 0660)
	defer f.Close()
	if err != nil {
		panic("Could not open " + *tokensFile + " err: " + err.Error())
	}
	if info, err := f.Stat(); err == nil && (info.Size() == 0 || overwrite) {
		var tokenStr string
		tokenStr = "accesstoken=\""+ cfg.AccessToken +"\"\r\n"
		tokenStr += "refreshtoken=\""+ cfg.RefreshToken +"\"\r\n"
		if info.Size() > 0 {
			f.Truncate(0)
			f.Seek(0, 0)
		}
		io.WriteString(f, tokenStr)
		f.Seek(0, 0)
	}
	tokens := &TwitchTokens{}
	if err := ReadConfig(f, tokens); err != nil {
		panic("Failed to parse config file, err: " + err.Error())
	}
	cfg.AccessToken = tokens.AccessToken
	cfg.RefreshToken = tokens.RefreshToken
}

func ReadConfig(r io.Reader, d interface{}) error {
	return toml.NewDecoder(r).Decode(d)
}

func FromContext(ctx context.Context) *AppConfig {
	cfg, _ := ctx.Value("appconfig").(*AppConfig)
	return cfg
}
