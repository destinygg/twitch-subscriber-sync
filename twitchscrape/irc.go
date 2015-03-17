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
	"math"
	"net"
	"regexp"
	"time"

	"github.com/destinygg/website2/internal/config"
	"github.com/destinygg/website2/internal/debug"
	"github.com/sorcix/irc"
	"golang.org/x/net/context"
)

// IConn represents the IRC connection to twitch,
// it is purely "single-threaded", methods are not safe to call concurrently
type IConn struct {
	conn net.Conn
	*irc.Decoder
	*irc.Encoder
	cfg *config.TwitchScrape
	// exponentially increase the time we sleep based on the number of tries
	// only resets when successfully connected to the server
	tries float64
	// the number of pings that were sent but not yet answered, should never go
	// beyond 2
	pendingPings int
}

func (c *IConn) Reconnect() {
	if c.conn != nil {
		_ = c.conn.Close()
	}

	conn, err := net.DialTimeout("tcp", c.cfg.Addr, 5*time.Second)
	if err != nil {
		c.delayAndLog("conn error: %+v", err)
		c.Reconnect()
		return
	}

	c.pendingPings = 0
	c.conn = conn
	c.Decoder = irc.NewDecoder(conn)
	c.Encoder = irc.NewEncoder(conn)

	pw := "oauth:" + c.cfg.OAuthToken
	c.Write(&irc.Message{Command: irc.PASS, Params: []string{pw}})
	c.Write(&irc.Message{Command: irc.NICK, Params: []string{c.cfg.Nick}})
	// sending irc.USER isn't even required, so just skip it
}

func (c *IConn) delayAndLog(format string, args ...interface{}) time.Duration {
	// clamp tries, so that the maximum amount of time we wait is ~5 minutes
	if c.tries > 40.0 {
		c.tries = 40.0
	}

	d := time.Duration(math.Pow(2.0, c.tries)*200) * time.Millisecond
	c.logWithDuration(format, d, args...)
	time.Sleep(d)
	c.tries++
	return d
}

func (c *IConn) logWithDuration(format string, dur time.Duration, args ...interface{}) {
	newargs := make([]interface{}, 0, len(args)+1)
	newargs = append(newargs, args...)
	newargs = append(newargs, dur)
	d.PF(2, format+", reconnecting in %s", newargs...)
}

// Write handles sending messages, it reconnects if there are problems
func (c *IConn) Write(m *irc.Message) {
	_ = c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	d.DF(2, "\t> %+v", m)
	if err := c.Encode(m); err != nil {
		c.delayAndLog("write error: %+v", err)
		c.Reconnect()
	}
}

// Read handles parsing messages from IRC and reconnects if there are problems
// returns nil on error
func (c *IConn) Read() *irc.Message {
	// if there are pending pings, lower the timeout duration to speed up
	// the disconnection
	if c.pendingPings > 0 {
		_ = c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	} else {
		_ = c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	}

	m, err := c.Decode()
	if err == nil {
		// we do not actually care about the type of the message the server sends us,
		// as long as it sends something it signals that its alive
		if c.pendingPings > 0 {
			c.pendingPings--
		}

		d.DF(2, "\t< %+v", m)
		return m
	}

	// if we hit the timeout and there are no outstanding pings, send one
	if e, ok := err.(net.Error); ok && e.Timeout() && c.pendingPings < 1 {
		c.pendingPings++
		c.Write(&irc.Message{
			Command: "PING",
			Params:  []string{"destinygg-subscription-notifier"},
		})
		return nil
	}

	// otherwise there either was an error or we did not get a reply for our ping
	c.delayAndLog("read error: %+v", err)
	c.Reconnect()
	return nil
}

func InitIRC(ctx context.Context) {
	// TODO implement metrics for emote usage, lines per sec, etc
	cfg := &config.GetFromContext(ctx).TwitchScrape
	c := &IConn{cfg: cfg}
	c.Reconnect()

	for {
		m := c.Read()
		// if we read nil, there was an error/timeout, Read will handle reconnection
		if m == nil {
			continue
		}

		switch m.Command {
		case irc.PING:
			c.Write(&irc.Message{Command: irc.PONG, Params: m.Params, Trailing: m.Trailing})
		case irc.RPL_WELCOME: // successfully connected
			c.tries = 0
			c.Write(&irc.Message{Command: irc.JOIN, Params: []string{"#" + cfg.Channel}})
		case irc.PRIVMSG:
			nick := getNewSubNick(m)
			if nick != "" {
				api.AddSubs([]string{nick})
			}
		}
	}
}

var subRe = regexp.MustCompile(`^([^ ]+) (?:just subscribed|subscribed for \d+ months in a row)!$`)

func getNewSubNick(m *irc.Message) string {
	if m.Prefix.Name == "twitchnotify" {
		match := subRe.FindStringSubmatch(m.Trailing)
		d.DF(1, "< MATCHED %+v, %+v", match, m)
		if len(match) == 2 {
			return match[1]
		}
	}

	return ""
}
