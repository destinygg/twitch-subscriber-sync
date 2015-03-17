package main

import (
	"github.com/sorcix/irc"
	"testing"
)

func TestSubMatch(t *testing.T) {
	tests := map[string]string{
		":twitchnotify!twitchnotify@twitchnotify.tmi.twitch.tv PRIVMSG #itmejp :Wortavin just subscribed!":                     "Wortavin",
		":twitchnotify!twitchnotify@twitchnotify.tmi.twitch.tv PRIVMSG #itmejp :xanctius subscribed for 2 months in a row!":    "xanctius",
		":twitchnotify!twitchnotify@twitchnotify.tmi.twitch.tv PRIVMSG #itmejp :gruffalo50 subscribed for 19 months in a row!": "gruffalo50",
		":twitchnotify!twitchnotify@twitchnotify.tmi.twitch.tv PRIVMSG #itmejp :something invalid!":                            "",
	}

	for raw, expected := range tests {
		m := irc.ParseMessage(r)
		nick := getNewSubNick(m)

		if nick != expected {
			t.Errorf("Expected nick %s, got %s", expected, nick)
		}
	}
}
