package main

import (
	"github.com/destinygg/website2/internal/config"
	"github.com/destinygg/website2/internal/debug"
	"golang.org/x/net/context"
)

func main() {
	ctx := context.Background()
	ctx = config.Init(ctx)
	cfg := config.GetFromContext(ctx)

	d.Init(ctx)
	InitTwitch(&cfg.TwitchScrape)
	InitApi(cfg) // api will require more than just the scraper config
	InitIRC(ctx)
}
