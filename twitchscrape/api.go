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
