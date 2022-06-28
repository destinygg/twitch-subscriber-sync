package main

import (
	"github.com/destinygg/twitch-subscriber-sync/internal/config"
	"github.com/destinygg/twitch-subscriber-sync/internal/debug"
	"github.com/destinygg/twitch-subscriber-sync/twitchpubsub/api"
	"github.com/destinygg/twitch-subscriber-sync/twitchpubsub/twitch"
	"golang.org/x/net/context"
	"time"
)

func main() {
	time.Local = time.UTC
	ctx := context.Background()
	ctx = config.Init(ctx)
	ctx = d.Init(ctx)
	ctx = api.Init(ctx)
	ctx = twitch.Init(ctx)
}
