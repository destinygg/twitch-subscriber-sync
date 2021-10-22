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

package api

import (
	"bytes"
	_ "crypto/sha512"
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/destinygg/twitch-subscriber-sync/internal/config"
	"github.com/destinygg/twitch-subscriber-sync/internal/debug"
	"github.com/destinygg/twitch-subscriber-sync/twitchscrape/twitch"
	"golang.org/x/net/context"
)

type Api struct {
	cfg *config.AppConfig

	mu sync.Mutex
	// subs are keyed by ids that are alphanumeric but not necessarily only digits
	subs       map[string]int
	client     http.Client
}

func Init(ctx context.Context) context.Context {
	api := &Api{
		cfg:        config.FromContext(ctx),
		subs:       map[string]int{},
		client: http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig:       &tls.Config{},
				ResponseHeaderTimeout: 5 * time.Second,
			},
		},
	}

	api.run(twitch.FromContext(ctx))
	return context.WithValue(ctx, "dggapi", api)
}

func FromContext(ctx context.Context) *Api {
	api, _ := ctx.Value("dggapi").(*Api)
	return api
}

func (a *Api) call(method, url string, body io.Reader) ([]byte, error) {
	u := url + "?privatekey=" + a.cfg.Website.PrivateAPIKey
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		d.PF(2, "Could not create request: %#v", err)
		return nil, err
	}

	res, err := a.client.Do(req)
	if res == nil || res.Body == nil {
		return nil, nil
	}
	defer res.Body.Close()

	if err != nil || res.StatusCode < 200 || res.StatusCode >= 300 {
		data, _ := ioutil.ReadAll(res.Body)
		d.PF(2, "Request failed: %#v, body was \n%v", err, string(data))
		return nil, err
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		d.PF(2, "Could not read body: %#v", err)
		return nil, err
	}

	return data, nil
}

func (a *Api) getSubsLocked() error {
	userids := struct {
		Authids []string `json:"authids"`
	}{}

	data, err := a.call("GET", a.cfg.TwitchScrape.GetSubURL, nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &userids)
	if err != nil {
		return err
	}

	for _, id := range userids.Authids {
		a.subs[id] = 1
	}
	return nil
}

// separate url parameter so that we can differentiate between resubs and
// fresh subs
func (a Api) syncSubs(subs map[string]int, url string) error {
	buf := &bytes.Buffer{}
	json.NewEncoder(buf).Encode(subs)
	_, err := a.call("POST", url, buf)
	return err
}

func (a *Api) run(tw *twitch.Twitch) {
	t := time.NewTicker(time.Duration(a.cfg.PollMinutes) * time.Minute)

	for {
		err := a.syncFromTwitch(tw)
		// retry on error
		if err != nil {
			d.P("syncFromTwitch failed, retrying: ", err)
			time.Sleep(30 * time.Second)
			continue
		}

		_ = <-t.C
	}
}

func (a *Api) syncFromTwitch(tw *twitch.Twitch) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	err := a.getSubsLocked()
	if err != nil {
		d.P("Could not get subs: ", err)
		return err
	}

	users, err := tw.GetSubs()
	if err != nil {
		return err
	}

	diff := make(map[string]int)
	visited := make(map[string]struct{}, len(users))

	for _, u := range users {
		visited[u.ID] = struct{}{}
		wassub, ok := a.subs[u.ID]
		if wassub != 1 && ok { // was not a sub before, but is now
			a.subs[u.ID] = 1
			diff[u.ID] = 1
		} else if !ok { // was not found at all, but could have registered since
			diff[u.ID] = 1
		}
	}

	// now check for expired subs, expired var is purely for logging reasons
	var expired int
	for id, wassub := range a.subs {
		if _, ok := visited[id]; ok { // already seen, has to be a sub
			continue
		}
		if wassub == 1 { // was a sub, but is no longer
			a.subs[id] = 0
			diff[id] = 0
			expired++
		}
	}

	// report the difference from the known d.gg subs always
	d.DF(1, "Found %v subs, syncing: %v, number of expired/notfound subs: %v", len(users), len(diff), expired)
	err = a.syncSubs(diff, a.cfg.TwitchScrape.ModSubURL)

	return err
}
