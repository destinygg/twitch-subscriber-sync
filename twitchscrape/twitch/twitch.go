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
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/destinygg/website2/internal/config"
	"github.com/destinygg/website2/internal/debug"
	"golang.org/x/net/context"
)

type Twitch struct {
	cfg     *config.TwitchScrape
	apibase string
}
type User struct {
	ID      string
	Name    string
	Created time.Time
}

func Init(ctx context.Context) context.Context {
	tw := &Twitch{
		cfg:     &config.GetFromContext(ctx).TwitchScrape,
		apibase: "https://api.twitch.tv/kraken/",
	}

	return context.WithValue(ctx, "twitch", tw)
}

func GetFromContext(ctx context.Context) *Twitch {
	cfg, _ := ctx.Value("twitch").(*Twitch)
	return cfg
}

func (t *Twitch) GetSubs() []User {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:       &tls.Config{},
			ResponseHeaderTimeout: 5 * time.Second,
		},
	}

	var err error
	var users []User

	/*
	  {
	    "_total": 286,
	    "_links": {
	      "self": "https://api.twitch.tv/kraken/channels/destiny/subscriptions?direction=ASC&limit=25&offset=0",
	      "next": "https://api.twitch.tv/kraken/channels/destiny/subscriptions?direction=ASC&limit=25&offset=25"
	    },
	    "subscriptions": [
	      {
	        "created_at": "2011-11-23T02:53:17Z",
	        "_id": "c4407b3d0b1d71ec6a2943950cc15e135d092391",
	        "_links": {
	          "self": "https://api.twitch.tv/kraken/channels/destiny/subscriptions/snowythedog"
	        },
	        "user": {
	          "display_name": "Snowythedog",
	          "_id": 22981482,
	          "name": "snowythedog",
	          "staff": false,
	          "created_at": "2011-06-16T18:23:11Z",
	          "updated_at": "2014-10-23T02:20:51Z",
	          "logo": null,
	          "_links": {
	            "self": "https://api.twitch.tv/kraken/users/snowythedog"
	          }
	        }
	      }
	    ]
	  }
	*/
	var js struct {
		Total int `json:"_total"`
		Links struct {
			Next string `json:"next"`
		} `json:"_links"`
		Subs []struct {
			Created string `json:"created_at"`
			User    struct {
				Name string `json:"name"`
				ID   int    `json:"_id"`
			} `json:"user"`
		} `json:"subscriptions"`
	}

	urlStr := t.apibase + "channels/" + t.cfg.Channel + "/subscriptions?limit=100"
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		d.F("Failed to create request, err: %+v", req, err)
	}
	req.Header.Add("Accept", "application/vnd.twitchtv.v3+json")
	req.Header.Add("Authorization", "OAuth "+t.cfg.OAuthToken)

	for {
		res, err := client.Do(req)
		if err != nil {
			d.P("Failed to GET the subscribers, req, err", req, err)
			return users
		}

		dec := json.NewDecoder(res.Body)
		err = dec.Decode(&js)
		defer res.Body.Close()
		if err != nil {
			body, _ := ioutil.ReadAll(res.Body)
			d.P("Failed to decode json, err", err, string(body))
			return users
		}

		if users == nil {
			users = make([]User, 0, js.Total)
		}

		if len(js.Subs) == 0 {
			return users
		}

		for _, s := range js.Subs {
			t, _ := time.ParseInLocation("2006-01-02T15:04:05Z", s.Created, time.UTC)
			users = append(users, User{
				ID:      fmt.Sprintf("%v", s.User.ID),
				Name:    s.User.Name,
				Created: t,
			})
		}

		u, err := url.Parse(js.Links.Next)
		if err != nil {
			d.F("Failed to parse url, url was: %s, err: %+v", js.Links.Next, err)
		}

		req.URL = u
	}
}
