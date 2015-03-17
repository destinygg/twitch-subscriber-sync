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

type IConn struct {
	conn net.Conn
	*irc.Decoder
	*irc.Encoder
	tries        float64
	cfg          *config.TwitchScrape
	pendingpings int
}

func (c *IConn) Reconnect() {
	conn, err := net.DialTimeout("tcp", c.cfg.Addr, 5*time.Second)
	if err != nil {
		c.delayAndLog("conn error: %+v", err)
		c.Reconnect()
		return
	}

	c.pendingpings = 0
	c.conn = conn
	c.Decoder = irc.NewDecoder(conn)
	c.Encoder = irc.NewEncoder(conn)

	pw := "oauth:" + c.cfg.OAuthToken
	_ = c.Write(&irc.Message{Command: irc.PASS, Params: []string{pw}})
	_ = c.Write(&irc.Message{Command: irc.NICK, Params: []string{c.cfg.Nick}})
	// nothing matters in this message
	_ = c.Write(&irc.Message{Command: irc.USER, Params: []string{"a", "b", "c"}, Trailing: "d"})
}

func (c *IConn) delayAndLog(format string, args ...interface{}) time.Duration {
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
	d.PF(2, "irc: "+format+", reconnecting in %s", newargs...)
}

func (c *IConn) Write(m *irc.Message) error {
	var err error
	c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	d.DF(2, "\t> %+v", m)
	if err = c.Encode(m); err != nil {
		c.delayAndLog("write error: %+v", err)
		c.Reconnect()
	}

	return err
}

func (c *IConn) Read() *irc.Message {
	// if there are pending pings, lower the timeout duration to speed up
	// the disconnection
	if c.pendingpings > 0 {
		c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	} else {
		c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	}

	m, err := c.Decode()
	if err == nil {
		// we do not actually care about the type of the message the server sends us,
		// as long as it sends something it signals that its alive
		if c.pendingpings > 0 {
			c.pendingpings--
		}

		d.DF(2, "\t< %+v", m)
		return m
	}

	// if we hit the timeout and there are no outstanding pings, send one
	if e, ok := err.(net.Error); ok && e.Timeout() && c.pendingpings < 1 {
		c.pendingpings++
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
	cfg := config.GetFromContext(ctx).TwitchScrape
	c := &IConn{cfg: cfg}
	c.Reconnect()

	for {
		m := c.Read()
		if m == nil {
			continue
		}

		switch m.Command {

		case irc.PING:
			_ = c.Write(&irc.Message{Command: irc.PONG, Params: m.Params, Trailing: m.Trailing})
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
