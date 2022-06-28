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
	"github.com/destinygg/twitch-subscriber-sync/internal/config"
	"github.com/destinygg/twitch-subscriber-sync/internal/debug"
	"github.com/destinygg/twitch-subscriber-sync/twitchscrape/api"
	"github.com/destinygg/twitch-subscriber-sync/twitchscrape/twitch"
	"golang.org/x/net/context"
	"time"
)

func main() {
	time.Local = time.UTC
	ctx := context.Background()
	ctx = config.Init(ctx)
	ctx = d.Init(ctx)
	ctx = twitch.Init(ctx)
	ctx = api.Init(ctx)
}
