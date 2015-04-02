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
	"strings"
	"sync"
	"time"

	"github.com/destinygg/website2/internal/config"
	"github.com/destinygg/website2/internal/debug"
	"github.com/destinygg/website2/twitchscrape/twitch"
	"golang.org/x/net/context"
)

type Api struct {
	cfg *config.AppConfig

	mu sync.Mutex
	// subs are keyed by ids that are alphanumeric but not necessarily only digits
	subs       map[string]int
	nicksToIDs map[string]string
	client     http.Client
}

func Init(ctx context.Context) context.Context {
	api := &Api{
		cfg:        config.GetFromContext(ctx),
		subs:       map[string]int{},
		nicksToIDs: map[string]string{},
		client: http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig:       &tls.Config{},
				ResponseHeaderTimeout: 5 * time.Second,
			},
		},
	}

	go api.run(twitch.GetFromContext(ctx))
	return context.WithValue(ctx, "dggapi", api)
}

func GetFromContext(ctx context.Context) *Api {
	api, _ := ctx.Value("dggapi").(*Api)
	return api
}

func (a *Api) call(method, url string, body io.Reader) (data []byte, err error) {
	u := url + "?privatekey=" + a.cfg.Website.PrivateAPIKey[0]
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		d.PF(2, "Could not create request: %#v", err)
		return nil, err
	}

	res, err := a.client.Do(req)
	d.DF(2, "Req: %#v\nRes: %#v\n err: %#v\n\n", req, res, err)
	if res.Body == nil {
		return nil, nil
	}
	defer res.Body.Close()

	if err != nil || res.StatusCode < 200 || res.StatusCode >= 300 {
		data, _ = ioutil.ReadAll(res.Body)
		d.PF(2, "Request failed: %#v, body was \n%v", err, string(data))
		return nil, err
	}

	data, err = ioutil.ReadAll(res.Body)
	if err != nil {
		d.PF(2, "Could not read body: %#v", err)
		return nil, err
	}

	return
}

func (a *Api) getSubs() {
	userids := struct {
		Authids []string
	}{}

	data, err := a.call("GET", a.cfg.TwitchScrape.GetSubURL, nil)
	if err != nil {
		time.Sleep(time.Second)
		a.getSubs()
	}

	err = json.Unmarshal(data, &userids)
	if err != nil {
		d.P("Could not unmarshal subs:", err, string(data))
		return
	}

	for _, id := range userids.Authids {
		a.subs[id] = 1
	}
	return
}

func (a *Api) fromNick(nick string) string {
	if id, ok := a.nicksToIDs[strings.ToLower(nick)]; ok {
		return id
	} else {
		d.DF(2, "Could not find the ID of the twitch user %v", nick)
		return ""
	}
}

// ReSub is safe to call concurrently
func (a *Api) ReSub(nick string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if id := a.fromNick(nick); id != "" {
		a.subs[id] = 1

		d := map[string]int{id: 1}
		a.syncSubs(d, a.cfg.TwitchScrape.ReSubURL)
	}
}

// AddSub is safe to call concurrently
func (a *Api) AddSub(nick string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	data := struct {
		Nick string `json:"nick"`
	}{Nick: nick}
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	_ = enc.Encode(data)
	a.call("POST", a.cfg.TwitchScrape.AddSubURL, &buf)
}

func (a Api) syncSubs(subs map[string]int, url string) {
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	_ = enc.Encode(subs)
	a.call("POST", url, &buf)
}

func (a *Api) run(tw *twitch.Twitch) {
	t := time.NewTicker(time.Duration(a.cfg.PollMinutes) * time.Minute)

	for {
		a.mu.Lock()
		a.getSubs()
		users, err := tw.GetSubs()
		// can only decide the list of unsubs if GetSubs returns no error and
		// as such, returns every single sub
		if err == nil {
			diff := make(map[string]int)
			visited := make(map[string]struct{}, len(users))

			for _, u := range users {
				a.nicksToIDs[strings.ToLower(u.Name)] = u.ID
				visited[u.ID] = struct{}{}

				wassub, ok := a.subs[u.ID]
				if wassub != 1 && ok { // was not a sub before, but is now
					a.subs[u.ID] = 1
					diff[u.ID] = 1
				} else if !ok {
					diff[u.ID] = 1
				}
			}

			// now check for expired subs
			for id, wassub := range a.subs {
				if _, ok := visited[id]; ok { // already seen, has to be a sub
					continue
				}

				if wassub == 1 { // was a sub, but is no longer
					a.subs[id] = 0
					diff[id] = 0
				}
			}

			a.syncSubs(diff, a.cfg.TwitchScrape.ModSubURL)
		}
		a.mu.Unlock()
		_ = <-t.C
	}
}
