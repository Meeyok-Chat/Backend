package models

type QueuePublisherPayload struct {
	From string `json:"from"`
}

type QueueReceiverPayload struct {
	From    string `json:"from"`
	Message string `json:"message"`
}
