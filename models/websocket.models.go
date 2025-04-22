package models

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

// Client
type ClientList map[*Client]bool

type Client struct {
	User       User
	ClientData ClientData
	Connection *websocket.Conn
	Egress     chan Event
}

type ClientData struct {
	Message      string `json:"message,omitempty"`
	ClientStatus string `json:"clientStatus,omitempty"`
	ServerStatus string `json:"serverStatus,omitempty"`
}

// Event
const (
	EventSendMessage = "send_message"
	EventNewMessage  = "new_message"

	EventSendMessageToMeeyok = "send_message_to_meeyok"

	EventNewUser   = "new_user"
	EventLeaveUser = "leave_user"
)

type EventHandler func(event Event, c *Client) error

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type SendMessageEvent struct {
	ChatID    string    `json:"chat_id"`
	Message   string    `json:"message"`
	From      string    `json:"from"`
	CreatedAt time.Time `json:"createAt"`
}
