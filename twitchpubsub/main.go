package main

import (
	"time"
	"github.com/destinygg/website2/internal/config"
	"github.com/destinygg/website2/internal/debug"
	"github.com/destinygg/website2/twitchpubsub/api"
	"github.com/destinygg/website2/twitchpubsub/twitch"
	"golang.org/x/net/context"
)

func main() {
	time.Local = time.UTC
	ctx := context.Background()
	ctx = config.Init(ctx)
	ctx = d.Init(ctx)
	ctx = api.Init(ctx)
	ctx = twitch.Init(ctx)
}