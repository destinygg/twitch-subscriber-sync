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
	"github.com/sorcix/irc"
	"testing"
)

func TestSubMatch(t *testing.T) {
	tests := map[string]struct {
		nick  string
		resub bool
	}{
		":twitchnotify!twitchnotify@twitchnotify.tmi.twitch.tv PRIVMSG #itmejp :Wortavin just subscribed!":                     {"Wortavin", false},
		":twitchnotify!twitchnotify@twitchnotify.tmi.twitch.tv PRIVMSG #itmejp :xanctius subscribed for 2 months in a row!":    {"xanctius", true},
		":twitchnotify!twitchnotify@twitchnotify.tmi.twitch.tv PRIVMSG #itmejp :gruffalo50 subscribed for 19 months in a row!": {"gruffalo50", true},
		":twitchnotify!twitchnotify@twitchnotify.tmi.twitch.tv PRIVMSG #itmejp :something invalid!":                            {"", false},
	}

	for raw, expected := range tests {
		m := irc.ParseMessage(raw)
		nick, resub := getNewSubNick(m)

		if nick != expected.nick && resub != expected.resub {
			t.Errorf("Expected nick %s resub: %v, got %s %v", expected.nick, expected.resub, nick, resub)
		}
	}
}
