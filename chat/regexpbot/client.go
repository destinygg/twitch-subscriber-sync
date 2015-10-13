package main

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/destinygg/website2/internal/debug"
	"github.com/gorilla/websocket"
)

type inMessage struct {
	Nick      string
	Features  []string
	Data      string
	Timestamp uint64
}

func parseMessage(msg []byte) (string, *inMessage) {
	index := bytes.IndexByte(msg, ' ')
	if index == -1 || len(msg) < index+1 {
		d.P("invalid message", string(msg))
		return "", nil
	}

	command := string(msg[:index])
	data := &inMessage{}
	if err := json.Unmarshal(msg[index+1:], data); err != nil {
		return "", nil
	}

	return command, data
}

func sendMessage(conn *websocket.Conn, str string) error {
	m := &struct {
		Data string `json:"data"`
	}{
		Data: str,
	}

	return send(conn, "MSG", m)
}

func sendMute(conn *websocket.Conn, target string, dur uint64) error {
	m := &struct {
		Data     string `json:"data"`
		Duration uint64 `json:"duration"`
	}{
		Data:     target,
		Duration: dur,
	}

	return send(conn, "MUTE", m)
}

func sendBan(conn *websocket.Conn, target, reason string, dur uint64, ipban bool) error {
	m := &struct {
		Nick     string `json:"nick"`
		Reason   string `json:"reason"`
		Duration uint64 `json:"duration"`
		IPBan    bool   `json:"banip"`
	}{
		Nick:     target,
		Reason:   reason,
		Duration: dur,
		IPBan:    ipban,
	}

	return send(conn, "BAN", m)
}

func send(conn *websocket.Conn, cmd string, m interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		panic("Could not marshal send message, err: " + err.Error())
	}

	buf := bytes.NewBuffer(make([]byte, 0, len(data)+len(cmd)+1))
	buf.WriteString(cmd)
	buf.WriteString(" ")
	buf.Write(data)

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return conn.WriteMessage(websocket.TextMessage, buf.Bytes())
}
