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
	"regexp"

	"github.com/destinygg/website2/internal/config"
	"github.com/destinygg/website2/internal/debug"
	"github.com/destinygg/website2/twitchscrape/api"
	"github.com/destinygg/website2/twitchscrape/twirc"
	"github.com/destinygg/website2/twitchscrape/twitch"
	"github.com/sorcix/irc"
	"golang.org/x/net/context"
)

func main() {
	ctx := context.Background()
	ctx = config.Init(ctx)

	ctx = d.Init(ctx)
	ctx = twitch.Init(ctx)
	ctx = api.Init(ctx)
	twirc.Init(ctx, func(c *twirc.IConn, m *irc.Message) {
		// TODO handle syncing of bans
		// TODO creation date of subs (need to do it everywhere at the same time)
		a := api.GetFromContext(ctx)

		switch m.Command {
		case irc.PRIVMSG:
			if nick, resub := getNewSubNick(m); nick != "" {
				if resub {
					a.ReSub(nick)
				} else {
					a.AddSub(nick)
				}
			}
		}
	})
}

var subRe = regexp.MustCompile(`^([^ ]+) (?:just subscribed|subscribed for (\d+) months in a row)!$`)

func getNewSubNick(m *irc.Message) (nick string, resub bool) {
	if m.Prefix.Name != "twitchnotify" {
		return
	}

	match := subRe.FindStringSubmatch(m.Trailing)
	d.DF(1, "< MATCHED %+v, %+v", match, m)
	if len(match) < 2 {
		return
	}

	return match[1], match[2] != ""
}
