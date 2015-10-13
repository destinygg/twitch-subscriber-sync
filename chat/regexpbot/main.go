package main

import (
	"bytes"
	"flag"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/destinygg/website2/internal/debug"
	"github.com/gorilla/websocket"
)

var settingsFile = flag.String("config", "settings.cfg", `path to the config file, it it doesn't exist it will
		be created with default values`)

func startWebsocket(s *state) {
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

again:
	c, _, err := dialer.Dial(s.Chat.URL, s.headers)
	if err != nil {
		d.P("Failed to dial websocket", err)
		time.Sleep(5 * time.Second)
		goto again
	}

	var lastSeenNuke time.Time
	c.SetPingHandler(func(m string) error {
		now := time.Now()
		if s.currentNuke != nil {
			if lastSeenNuke.IsZero() {
				lastSeenNuke = now
			} else if now.After(lastSeenNuke.Add(10 * time.Minute)) {
				s.currentNuke = nil
				lastSeenNuke = time.Time{}
			}
		}

		// clean up the nuked nicks here
		for n, t := range s.nukedNicks {
			if now.After(t.Add(1 * time.Hour)) {
				delete(s.nukedNicks, n)
			}
		}

		c.SetWriteDeadline(now.Add(5 * time.Second))
		return c.WriteMessage(websocket.PongMessage, []byte(m))
	})
	c.SetPongHandler(func(m string) error {
		c.SetReadDeadline(time.Now().Add(120 * time.Second))
		return nil
	})

	s.conn = c
	handleWebsocket(s)
	_ = c.Close()
	time.Sleep(5 * time.Second)
	goto again
}

func handleWebsocket(s *state) {
	for {
		s.conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		msgtype, msg, err := s.conn.ReadMessage()
		if err != nil {
			d.P("Read error, reconnecting", err)
			return
		}

		if msgtype != websocket.TextMessage {
			continue
		}

		// parse message, decide what to do
		cmd, data := parseMessage(msg)
		if data == nil || len(data.Nick) == 0 || cmd != "MSG" {
			continue
		}

		adminFound := false
		for _, nick := range s.Admins {
			if data.Nick == nick {
				adminFound = true
				break
			}
		}

		if adminFound {
			err = handleAdminMessage(s, []byte(data.Data))
		} else {
			err = handleMessage(s, data)
		}

		if err != nil {
			return
		}
	}
}

func handleMessage(s *state, msg *inMessage) error {
	for _, v := range msg.Features {
		if v == "admin" || v == "protected" {
			return nil
		}
	}

	if s.currentNuke != nil && s.currentNuke.MatchString(msg.Data) {
		err := sendMute(s.conn, msg.Nick, s.currentNukeDur)
		if err != nil {
			return err
		}

		s.nukedNicks[msg.Nick] = time.Now()
	} else {
		s.logChatMsg(msg.Nick, msg.Data)
	}

	for re, dur := range s.blacklist {
		if re.MatchString(msg.Data) {
			// make the offenses scale
			s.numOffenses[msg.Nick]++
			n := s.numOffenses[msg.Nick]
			dur *= uint64(math.Pow(2, float64(n-1)))

			err := sendMute(s.conn, msg.Nick, dur)
			if err != nil {
				return err
			}
			s.save()
			return sendMessage(s.conn, fmt.Sprintf("%s: blacklisted phrase, muted for %s Nappa", msg.Nick, humanizeDuration(dur)))
		}
	}

	return nil
}

func handleAdminMessage(s *state, msg []byte) error {
	var command string
	var arg []byte
	if len(msg) == 0 || msg[0] != '!' {
		return nil
	}

	index := bytes.IndexByte(msg, ' ')
	if index == -1 {
		command = string(msg[1:])
	} else {
		command = string(msg[1:index])
	}

	command = strings.ToLower(command)
	if len(msg) > index+1 {
		arg = msg[index+1:]
	}

	if fn, ok := adminCommands[command]; ok {
		return fn(s, arg)
	}

	return nil
}

var adminRE = regexp.MustCompile("(?i)^.*[\"`](.+)[\"`].*?([\\da-z]+)?")

func compileRegexp(s *state, line []byte) (*regexp.Regexp, interface{}, error) {
	m := adminRE.FindSubmatch(line)
	if len(m) == 0 {
		return nil, 0, nil
	}

	rs := string(m[1])
	re, err := regexp.Compile(rs)
	if err != nil {
		return nil, 0, nil
	}

	var dur string
	if len(m[2]) > 0 {
		dur = string(m[2])
	} else {
		dur = s.Blacklist.DefaultDuration
	}

	re.Longest()
	return re, dur, nil
}

var adminCommands = map[string]func(*state, []byte) error{
	"addregexp": func(s *state, arg []byte) error {
		re, dur, err := compileRegexp(s, arg)
		if err != nil {
			return sendMessage(s.conn, "Unable to parse regexp, err: "+err.Error())
		}
		s.addBlacklist(re, dur)
		s.save()
		return sendMessage(s.conn, "Done")
	},
	"listregexp": func(s *state, _ []byte) error {
		res := make([]string, 0, len(s.blacklist))
		for k, v := range s.blacklist {
			res = append(res, fmt.Sprintf(`"%s" mute duration: %s`, k.String(), humanizeDuration(v)))
		}

		sort.Sort(sort.StringSlice(res))
		for _, v := range res {
			err := sendMessage(s.conn, v)
			if err != nil {
				return err
			}
		}

		return nil
	},
	"delregexp": func(s *state, arg []byte) error {
		m := adminRE.FindSubmatch(arg)
		if len(m) == 0 {
			return nil
		}

		restr := string(m[2])
		for k := range s.blacklist {
			if k.String() == restr {
				delete(s.blacklist, k)
				for k, v := range s.Blacklist.Item {
					if v.Regexp == restr { // not preserving the order, dont care
						s.Blacklist.Item[k] = s.Blacklist.Item[len(s.Blacklist.Item)-1]
						s.Blacklist.Item = s.Blacklist.Item[:len(s.Blacklist.Item)-1]
						break
					}
				}
				s.save()
				return sendMessage(s.conn, "Deleted")
			}
		}
		return nil
	},
	"resetoffenses": func(s *state, arg []byte) error {
		nick := strings.TrimSpace(string(arg))
		if len(nick) == 0 {
			return nil
		}

		if n, ok := s.numOffenses[nick]; ok {
			s.numOffenses[nick] = 0
			s.save()
			return sendMessage(s.conn, fmt.Sprintf("Reset %s, had %d offenses Hhhehhehe", nick, n))
		}

		return sendMessage(s.conn, "Not found")
	},
	"regexpnuke": func(s *state, arg []byte) error {
		re, dur, err := compileRegexp(s, arg)
		if err != nil {
			return sendMessage(s.conn, "Unable to parse regexp, err: "+err.Error())
		}

		d := dur.(uint64)
		s.currentNukeDur = d
		now := time.Now()
		for _, v := range s.lastMsgs {
			if len(v.nick) == 0 || !re.MatchString(v.msg) {
				continue
			}

			err := sendMute(s.conn, v.nick, d)
			if err != nil {
				return err
			}

			s.nukedNicks[v.nick] = now
		}
		s.currentNuke = re
		return nil
	},
	"regexpaegis": func(s *state, arg []byte) error {
		s.currentNuke = nil
		for nick := range s.nukedNicks {
			err := sendUnmute(s.conn, nick)
			if err != nil {
				return err
			}
			delete(s.nukedNicks, nick)
		}
		return nil
	},
}

func main() {
	flag.Parse()
	adminRE.Longest()

	s := loadState()
	s.init()

	startWebsocket(s)
}
