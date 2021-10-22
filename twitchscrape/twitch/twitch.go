/***
  This file is part of twitchscrape.

  Copyright (c) 2015 Peter Sztan <sztanpet@gmail.com>

  twitchscrape is free software; you can redistribute it and/or modify it
  under the terms of the GNU Lesser General Public License as published by
  the Free Software Foundation; either version 3 of the License, or
  (at your option) any later version.

  twitchscrape is distributed in the hope that it will be useful, but
  WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
  Lesser General Public License for more details.

  You should have received a copy of the GNU Lesser General Public License
  along with twitchscrape; If not, see <http://www.gnu.org/licenses/>.
***/

package twitch

import (
	_ "crypto/sha512"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
	"io/ioutil"

	"github.com/destinygg/twitch-subscriber-sync/internal/config"
	"github.com/destinygg/twitch-subscriber-sync/internal/debug"
	"golang.org/x/net/context"
	"strconv"
)

type Twitch struct {
	cfg         *config.TwitchScrape
	apibase     string
	authapibase string
}

type User struct {
	ID      string
	Name    string
}

type TokenStruct struct {
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Scope []string `json:"scope"`
}

var client = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig:       &tls.Config{},
		ResponseHeaderTimeout: 30 * time.Second,
	},
}

func Init(ctx context.Context) context.Context {
	tw := &Twitch{
		cfg:     &config.FromContext(ctx).TwitchScrape,
		apibase: "https://api.twitch.tv/helix/",
		authapibase: "https://id.twitch.tv/oauth2/",
	}
	return context.WithValue(ctx, "twitch", tw)
}

func FromContext(ctx context.Context) *Twitch {
	cfg, _ := ctx.Value("twitch").(*Twitch)
	return cfg
}

func (t *Twitch) GetSubs() ([]User, error) {
	// https://dev.twitch.tv/docs/api/reference#get-broadcaster-subscriptions
	var users []User
	var js struct {
		Subs []struct {
			Name string `json:"user_login"`
			ID   string `json:"user_id"`
		} `json:"data"`

		Pagination struct {
			Cursor string `json:"cursor"`	
		} `json:"pagination"`

		Total int `json:"total"`
	}

	cursor := ""
	limit := 100
	urlBase := t.apibase + "subscriptions"

	headers := http.Header{
		"Authorization": []string{"Bearer " + t.cfg.AccessToken},
		"Client-ID":     []string{t.cfg.ClientID},
	}

	for {
		urlStr := urlBase + "?broadcaster_id=" + t.cfg.ChannelID + "&first=" + strconv.Itoa(limit) + "&after=" + cursor

		var err error
		var res *http.Response
		{
			u, err := url.Parse(urlStr)
			if err != nil {
				d.P("could not parse url", urlStr)
				return nil, err
			}
			d.DF(1,"Calling %s", u)
			res, err = client.Do(&http.Request{
				Method:     "GET",
				URL:        u,
				Proto:      "HTTP/1.1",
				ProtoMajor: 1,
				ProtoMinor: 1,
				Header:     headers,
				Body:       nil,
				Host:       u.Host,
			})
		}

		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err == nil {
			d.DF(1, "%s - %s", res.Status, bodyBytes)
		}

		if res != nil && res.StatusCode == 401 {
			t.Auth()
			return nil, fmt.Errorf("bad auth response")
		}
		if err != nil || res == nil || res.StatusCode != 200 {
			if err == nil {
				err = fmt.Errorf("non-200 statuscode received from twitch")
			}
			d.P("Failed to GET the subscribers, url, res, err", urlStr, res, err)
			return nil, err
		}

		err = json.Unmarshal(bodyBytes, &js)
		res.Body.Close()
		if err != nil {
			d.P("Failed to decode json, err", err)
			return nil, err
		}
		if users == nil {
			users = make([]User, 0, js.Total)
		}
		d.DF(1, "Successful response. Returned records [%v] Total users [%v]", len(js.Subs), len(users))

		for _, u := range js.Subs {
			users = append(users, User{
				ID:      fmt.Sprintf("%v", u.ID),
				Name:    u.Name,
			})
		}

		cursor = js.Pagination.Cursor

		// Finished when no subs are returned, which indicates the last page.
		if len(js.Subs) == 0 {
			return users, nil
		}
	}
}

func (t *Twitch) Auth() error {
	d.DF(1, "renewing access token")
	u, _ := url.Parse(t.authapibase + "token")
	q := u.Query()
	q.Add("grant_type", "refresh_token")
	q.Add("refresh_token", t.cfg.RefreshToken)
	q.Add("client_id", t.cfg.ClientID)
	q.Add("client_secret", t.cfg.ClientSecret)
	u.RawQuery = q.Encode()
	var res *http.Response
	{
		d.DF(1, "Calling %s", u)
		res, err := client.Do(&http.Request{
			Method:     "POST",
			URL:        u,
			Proto:      "HTTP/1.1",
			ProtoMinor: 1,
			Header:     nil,
			Body:       nil,
			Host:       u.Host,
		})
		if err != nil || res == nil || res.StatusCode != 200 {
			if res != nil && res.StatusCode != 200 {
				err = fmt.Errorf("non-200 statuscode received from twitch %v", res)
			} else if err == nil {
				err = fmt.Errorf("non-200 statuscode received from twitch")
			}
			d.P("Failed to GET the auth token, url, res, err", u, res, err)
			return err
		}
		tokens := &TokenStruct{}
		err = json.NewDecoder(res.Body).Decode(tokens)
		res.Body.Close()
		if err != nil {
			d.P("Failed to decode twitch response %v", err)
			return err
		}
		d.DF(1, "Updated OAuth Tokens")
		t.cfg.RefreshToken = tokens.RefreshToken
		t.cfg.AccessToken = tokens.AccessToken
		config.ReadTokensFile(t.cfg, true)
	}
	d.DF(1, "Response %s", res)
	return nil
}
