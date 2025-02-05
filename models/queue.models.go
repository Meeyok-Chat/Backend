package models

type QueueReceiverPayload struct {
	PromptData PromptData `json:"promptData"`
}

type PromptData struct {
	Message                  string `json:"message"`
	From                     string `json:"from"`
	Phase                    string `json:"phase"`
	Reasoning                string `json:"reasoning"`
	NumberOfSelectedQuestion int    `json:"numberOfSelectedQuestion"`
}
