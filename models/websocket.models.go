package models

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type ClientList map[*Client]bool

type Client struct {
	Chat       Chat
	ClientData ClientData
	Connection *websocket.Conn
	Egress     chan Event
	Chatroom   string
}

type ClientData struct {
	Message      string `json:"message,omitempty"`
	ClientStatus string `json:"clientStatus,omitempty"`
	ServerStatus string `json:"serverStatus,omitempty"`
}

type EventHandler func(event Event, c *Client) error

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type StatusMessageEvent struct {
	Status string `json:"status"`
}

// SendMessageEvent is the payload sent in the
// send_message event
type SendMessageEvent struct {
	Message   string `json:"message"`
	From      string `json:"from"`
	Phase     string `json:"phase"`
	Reasoning string `json:"reasoning"`
}

type UpdateFeedbackEvent struct {
	ID      string `json:"id,omitempty"`
	Message string `json:"message"`
	Score   int    `json:"score"`
}

type SendStatusEvent struct {
	Status string `json:"status"`
}

type ChangeRoomEvent struct {
	Name string `json:"name"`
}
