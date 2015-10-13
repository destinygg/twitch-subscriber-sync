package main

import (
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/destinygg/website2/internal/config"
	"github.com/gorilla/websocket"
)

type Item struct {
	Regexp   string `toml:"regexp"`
	Duration uint64 `toml:"duration"`
}
type Offense struct {
	Nick  string `toml:"nick"`
	Count uint64 `toml:"count"`
}
type chatMsg struct {
	nick string
	msg  string
}

type state struct {
	headers   http.Header
	blacklist map[*regexp.Regexp]uint64
	conn      *websocket.Conn

	// a map of nicks and the number of offenses they did
	numOffenses map[string]uint64

	// a circular buffer of chat messages to check nukes against
	lastMsgs  [300]chatMsg
	lastMsgIx int

	// the current nuke state
	currentNuke    *regexp.Regexp
	currentNukeDur uint64
	nukedNicks     map[string]time.Time

	Admins []string `toml:"admins"`
	Chat   struct {
		URL       string `toml:"URL"`
		AuthToken string `toml:"authtoken"`
	}
	Blacklist struct {
		DefaultDuration string `toml:"defaultduration"`
		Item            []Item `toml:"item"`
	}
	Offenses []Offense `toml:"offense"`
}

func loadState() *state {
	f, err := os.OpenFile(*settingsFile, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		panic("Could not open " + *settingsFile + " err: " + err.Error())
	}
	defer f.Close()

	s := &state{
		blacklist:   map[*regexp.Regexp]uint64{},
		numOffenses: map[string]uint64{},
		nukedNicks:  map[string]time.Time{},
		lastMsgIx:   -1, // because logChatMsg increments it first
	}

	if info, err := f.Stat(); err == nil && info.Size() == 0 {
		_ = config.WriteConfig(f, *s)
		panic("Default config written, please edit and re-run")
	}

	err = config.ReadConfig(f, s)
	if err != nil {
		panic("Failed to read config, err:" + err.Error())
	}

	return s
}

func (s *state) init() {
	if len(s.Blacklist.DefaultDuration) > 0 {
		dur := parseDuration(s.Blacklist.DefaultDuration, "m")
		if dur != 0 && dur != defaultDuration {
			defaultDuration = dur
		}
	}

	s.headers = http.Header{
		"Origin": []string{"http://localhost"},
		"Cookie": []string{"authtoken=" + s.Chat.AuthToken},
	}

	for _, v := range s.Blacklist.Item {
		re, _, err := compileRegexp(s, []byte(`"`+v.Regexp+`"`))
		if err != nil {
			panic("Unable to compile regexp, err: " + err.Error())
		}

		s.addBlacklist(re, v.Duration)
	}
	for _, v := range s.Offenses {
		s.numOffenses[v.Nick] = v.Count
	}
}

func (s *state) save() {
	of := make([]Offense, 0, len(s.numOffenses))
	for n, c := range s.numOffenses {
		of = append(of, Offense{
			Nick:  n,
			Count: c,
		})
	}
	s.Offenses = of

	bl := make([]Item, 0, len(s.blacklist))
	for r, d := range s.blacklist {
		bl = append(bl, Item{
			Regexp:   r.String(),
			Duration: d,
		})
	}
	s.Blacklist.Item = bl

	err := config.SafeSave(*settingsFile, *s)
	if err != nil {
		panic("Failed to save config, err:" + err.Error())
	}
}

func (s *state) addBlacklist(re *regexp.Regexp, d interface{}) {
	var dur uint64
	var shouldSave bool

	switch d.(type) {
	case uint64:
		dur = d.(uint64)
		// if dur's type is uint64 -> comfing from init, do not save
	case string:
		dur = parseDuration(d.(string), "min")
		// if dur's type is string -> parse the duration, save
		shouldSave = true
	}

	s.blacklist[re] = dur
	if shouldSave {
		s.save()
	}
}

func (s *state) logChatMsg(nick, msg string) {
	// only log the last N msgs, use the array as a circular buffer
	if s.lastMsgIx == cap(s.lastMsgs)-1 {
		s.lastMsgIx = -1 // start over from the beginning
	}

	s.lastMsgIx++
	s.lastMsgs[s.lastMsgIx] = chatMsg{nick: nick, msg: msg}
}
