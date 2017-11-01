package twitch

import (
	"time"
	"github.com/gorilla/websocket"
	"os"
	"os/signal"
	"net/url"
	"bytes"
	"encoding/json"
	"math"
	"strings"
	"github.com/destinygg/website2/internal/config"
	"github.com/destinygg/website2/internal/debug"
	"github.com/destinygg/website2/twitchpubsub/api"
	"golang.org/x/net/context"
)

const (
	// Time allowed to write a message to the peer.
	maxMessageSize = 2048

	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to write a close message to the peer.
	closeWait = 3 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 5) / 10

	// twitch pub/sub
	webSocketUri    = "wss://pubsub-edge.twitch.tv/"
	msgEventPrefix  = "channel-subscribe-events-v1"
	msgTypePing     = "PING"
	msgTypePong     = "PONG"
	msgTypeListen   = "LISTEN"
	msgTypeResponse = "RESPONSE"
)

type IConn struct {
	conn *websocket.Conn
	cfg *config.TwitchScrape
	tries float64
	closing bool
}
type Message struct {
	Type  string `json:"type"`
	Error string `json:"error,omitempty"`
	Data MessageData `json:"data,omitempty"`
}
type MessageData struct {
	Topic string `json:"topic"`
	Message string `json:"message"`
}
type SubscribePayload struct {
	Type string `json:"type"`
	Nonce string `json:"nonce,omitempty"`
	Data SubscribePayloadData `json:"data"`
}
type SubscribePayloadData struct {
	Topics []string `json:"topics"`
	AuthToken string `json:"auth_token,omitempty"`
}

/*type SubscribeMessageData struct {
	UserName string 	`json:"user_name"`
	DisplayName string 	`json:"display_name"`
	ChannelName string 	`json:"channel_name"`
	UserID string 		`json:"user_id"`
	ChannelID string 	`json:"channel_id"`
	Time string 		`json:"time"`
	SubPlan string 		`json:"sub_plan"`
	SubPlanName string 	`json:"sub_plan_name"`
	Months string		`json:"months"`
	Context string 		`json:"context"`
	SubMessage map[string]interface{} `json:"sub_message"`
}*/

func Init(ctx context.Context) context.Context {
	c := &IConn{
		cfg: &config.FromContext(ctx).TwitchScrape,
		closing: false,
		tries: 0,
	}
	c.run(api.FromContext(ctx))
	return context.WithValue(ctx, "twitch", c)
}

func (c *IConn) run(a *api.Api) {
	time.Local = time.UTC

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	done := make(chan struct{})

	c.Reconnect()
	defer c.conn.Close()

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	go func() {
		defer c.conn.Close()
		defer close(done)
		for {
			m := c.Read()
			if m == nil {
				continue
			}
			switch m.Type {
			case msgTypeResponse:
				// TODO this is a response to the subscribe frame
				// Currently not looking at the response and assume the subscribe worked.
				break
			default:
				p := strings.Split(m.Data.Topic, ".")[0]
				switch p {
				case msgEventPrefix:
					// https://dev.twitch.tv/docs/pubsub#example-channel-subscriptions-event-message
					// msg := &SubscribeMessageData{}
					// json.Unmarshal([]byte(m.Data.Message), &msg)
					d.DF(1, "Data %+v", m.Data.Message)
					a.SendSubDataToApi(strings.NewReader(m.Data.Message))
				default:
					d.DF(1, "Unsupported message.", m)
				}
			}
		}
	}()

	for {
		select {
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			c.conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"`+msgTypePing+`"}`))
		case <-interrupt:
			c.closing = true
			d.DF(1, "interrupted")
			if err := c.SendCloseFrame(); err != nil {
				d.DF(1,"write close: %v", err)
				return
			}
			select {
			case <-done:
			case <-time.After(closeWait):
			}
			c.conn.Close()
			c.closing = false
			d.DF(1, "connection closed")
			return
		}
	}
}

func (c *IConn) ReconnectAfterError(err error) {
	if c.tries > 10.0 {
		c.tries = 10.0
	}
	dur := time.Duration(math.Pow(2.0, c.tries)*300) * time.Millisecond
	d.DF(1,"reconnecting in %s", dur)
	time.Sleep(dur)
	c.tries++
	c.Reconnect()
}

func (c *IConn) Reconnect() {
	if c.conn != nil {
		c.conn.Close()
	}
	u, err := url.Parse(webSocketUri)
	d.DF(1,"connecting: %s", u.String())
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		d.DF(1, "conn error: %+v", err)
		c.ReconnectAfterError(err)
		return
	}
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	c.conn = conn

	m := &SubscribePayload{
		Type: msgTypeListen,
		Data: SubscribePayloadData{
			AuthToken: c.cfg.OAuthToken,
			Topics: []string{msgEventPrefix + "." + c.cfg.ChannelID},
		}}

	buf := &bytes.Buffer{}
	json.NewEncoder(buf).Encode(m)
	c.Write(websocket.TextMessage, buf.Bytes())
}

func (c *IConn) SendCloseFrame() error {
	c.conn.SetWriteDeadline(time.Now().Add(closeWait))
	return c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

func (c *IConn) Write(messageType int, data []byte) {
	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	data = bytes.TrimSpace(data)
	d.DF(1,"-> %s", data)
	if err := c.conn.WriteMessage(messageType, data); err != nil {
		d.DF(1,"write error: %+v", err)
		c.ReconnectAfterError(err)
	}
}

func (c *IConn) Read() (*Message) {
	_, message, err := c.conn.ReadMessage()
	if err != nil {
		if !c.closing {
			// TODO we aren't getting a valid close status after sending the close frame to twitch
			//if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure)
			d.DF(1, "read error: %+v", err)
			c.ReconnectAfterError(err)
		}
		return nil
	}
	m := &Message{}
	if err := json.Unmarshal(message, &m); err != nil {
		d.DF(1,"parse error: %v", err)
		return nil
	}
	if m.Type == msgTypePong {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	}
	d.DF(1,"<- %s", m)
	return m
}