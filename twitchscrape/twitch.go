package main

import (
	_ "crypto/sha512"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/destinygg/website2/internal/config"
	"github.com/destinygg/website2/internal/debug"
)

var twitch *Twitch

type Twitch struct {
	cfg     *config.TwitchScrape
	apibase string
}

func InitTwitch(cfg *config.TwitchScrape) {
	twitch = &Twitch{
		cfg:     cfg,
		apibase: "https://api.twitch.tv/kraken/",
	}
}

func (t *Twitch) GetSubs() map[string]time.Time {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:       &tls.Config{},
			ResponseHeaderTimeout: 5 * time.Second,
		},
	}

	var err error
	var users map[string]time.Time

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
		resp, err := client.Do(req)
		if err != nil {
			d.F("Failed to GET the subscribers, req: %+v, err: %+v", req, err)
		}

		dec := json.NewDecoder(resp.Body)
		err = dec.Decode(&js)
		if err != nil {
			d.F("Failed to decode json, err: %+v", err)
		}

		_ = resp.Body.Close()

		if users == nil {
			users = make(map[string]time.Time, js.Total)
		}

		if len(js.Subs) == 0 {
			return users
		}

		for _, s := range js.Subs {
			t, _ := time.ParseInLocation("2006-01-02T15:04:05Z", s.Created, time.UTC)
			users[s.User.Name] = t
		}

		u, err := url.Parse(js.Links.Next)
		if err != nil {
			d.F("Failed to parse url, url was: %s, err: %+v", js.Links.Next, err)
		}

		req.URL = u
	}
}
