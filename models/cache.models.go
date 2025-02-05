package models

type CacheData struct {
	ChatData   Chat       `json:"chat,omitempty"`
	ClientData ClientData `json:"clientData,omitempty"`
}
