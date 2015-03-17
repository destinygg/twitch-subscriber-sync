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

package main

import (
	//"net/http"
	"github.com/destinygg/website2/internal/config"
	"time"
)

var api *Api

type Api struct {
	cfg  *config.AppConfig
	subs map[string]struct{}
}

func InitApi(cfg *config.AppConfig) {
	api = &Api{
		cfg: cfg,
	}

	go api.run()
}

func (a *Api) AddSubs(nicks []string) {
	// TODO
}

func (a *Api) DelSubs(nicks []string) {
	// TODO
}

func (a *Api) run() {
	t := time.NewTicker(time.Duration(a.cfg.PollMinutes) * time.Minute)
	for range t.C {
		// TODO get subs from the website
		// get subs from twitch
		// get the difference and act accordingly
	}
}
