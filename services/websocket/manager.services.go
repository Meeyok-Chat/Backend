package Websocket

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Meeyok-Chat/backend/models"
	"github.com/Meeyok-Chat/backend/repository/cache"
	"github.com/Meeyok-Chat/backend/repository/database"
	"github.com/Meeyok-Chat/backend/repository/queue/queuePublisher"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Event between client and server
const (
	EventSendMessage = "send_message"
	EventNewMessage  = "new_message"

	EventUpdateFeedback = "update_feedback"
	EventNewFeedback    = "new_feedback"

	EventSendStatus = "send_status"
	EventNewStatus  = "new_status"

	EventNewSummarySheet = "new_summary_sheet"
)

// Event between server and AI-Service
const (
	EventSendMessageToPrompt  = "send_message_to_prompt"
	EventGetMessageFromPrompt = "new_message_from_prompt"

	EventSendSummaryRequestToPrompt  = "send_summary_request"
	EventGetSummaryRequestFromPrompt = "new_summary_request"
)

// Status from client to server
const (
	ClientLoadingStatus = "client_loading"
	ClientSuccessStatus = "client_success"
	ClientErrorStatus   = "client_error"
)

// Status from server and client
const (
	ServerLoadingStatus = "server_loading"
	ServerSuccessStatus = "server_success"
	ServerErrorStatus   = "server_error"
)

type managerService struct {
	clients        models.ClientList
	cacheRepo      cache.CacheRepo
	chatRepo       database.ChatRepo
	queuePublisher queuePublisher.QueuePublisher
	// Using a syncMutex here to be able to lcok state before editing clients
	// Could also use Channels to block
	sync.RWMutex
	// handlers are functions that are used to handle Events
	handlers map[string]models.EventHandler
}

// Manager is used to hold references to all Clients Registered, and Broadcasting etc
type ManagerService interface {
	AddClient(conn *websocket.Conn, c *gin.Context, chatId primitive.ObjectID)
	RemoveClient(client *models.Client)
	RouteEvent(event models.Event, c *models.Client) error
	CheckOldClient(chatId string) error
	SendMessageToOfflineClient(event models.Event, promptData models.PromptData, chatId string, dataError bool)
	SummaryChat()
	GetClients() models.ClientList
	NewStatusHandler(status string, c *models.Client) error
	SendMessage(event models.Event, c *models.Client) error
	NewSummarySheetHandler(chatId string) error
}

// NewManager is used to initalize all the values inside the manager
func NewManagerService(cacheRepo cache.CacheRepo, chatRepo database.ChatRepo, queuePublisher queuePublisher.QueuePublisher) ManagerService {
	m := &managerService{
		clients:        make(models.ClientList),
		cacheRepo:      cacheRepo,
		chatRepo:       chatRepo,
		queuePublisher: queuePublisher,
		handlers:       make(map[string]models.EventHandler),
	}
	m.setupEventHandlers()
	return m
}

// setupEventHandlers configures and adds all handlers
func (ms *managerService) setupEventHandlers() {
	ms.handlers[EventSendMessage] = ms.sendMessageHandler
	ms.handlers[EventUpdateFeedback] = ms.updateFeedbackHandler
	ms.handlers[EventSendStatus] = ms.sendStatusHandler
}

// routeEvent is used to make sure the correct event goes into the correct handler
func (ms *managerService) RouteEvent(event models.Event, c *models.Client) error {
	// Check if Handler is present in Map
	if handler, ok := ms.handlers[event.Type]; ok {
		// Execute the handler and return any err
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("this event type is not supported")
	}
}

func (ms *managerService) CheckOldClient(chatId string) error {
	// Check if there is an existing client for the same chat and wait for it to be removed
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for {
		existingClientFound := false

		for client := range ms.clients {
			if client.Chat.ID.Hex() == chatId {
				existingClientFound = true
				break
			}
		}

		if !existingClientFound {
			break
		}

		select {
		case <-ctx.Done():
			// Timeout after 10 seconds
			return fmt.Errorf("failed to remove existing client for chat: %s within timeout", chatId)
		default:
			log.Println("Waiting for existing client to be removed")
			time.Sleep(500 * time.Millisecond)
		}
	}
	return nil
}

// addClient will add clients to our clientList
func (ms *managerService) AddClient(conn *websocket.Conn, c *gin.Context, chatId primitive.ObjectID) {
	chat, err := ms.chatRepo.GetChatById(chatId)
	if err != nil {
		log.Println(err)
	}

	log.Println("New connection with chatId : " + string(chatId.Hex()))

	cacheData, cacheErr := ms.cacheRepo.CheckCache(chatId.Hex())

	log.Println("Key :", chatId.Hex())
	// get chat from database and set that chat in cache. (UpdateChatCache)
	ms.cacheRepo.UpdateChatCache(chatId.Hex(), chat)

	chat.Messages = []models.Message{}
	// Create New Client
	clientService := NewClientService(chat, conn, ms)
	if cacheErr == nil {
		clientService.GetClient().ClientData = cacheData.ClientData
	}

	go clientService.ReadMessages()
	go clientService.WriteMessages()

	ms.initMessageBot(clientService.GetClient())

	if clientService.GetClient().ClientData != (models.ClientData{}) {
		log.Println("send server status")
		ms.NewStatusHandler(clientService.GetClient().ClientData.ServerStatus, clientService.GetClient())
	}

	// Lock so we can manipulate
	ms.Lock()
	defer ms.Unlock()

	// Add Client
	ms.clients[clientService.GetClient()] = true
}

func (ms *managerService) initMessageBot(c *models.Client) {
	chat, err := ms.chatRepo.GetChatById(c.Chat.ID)
	if err != nil {
		log.Println(err)
	}
	if len(chat.Messages) != 0 {
		return
	}
	ms.serverSendMessage(c, "สวัสดีค่ะ มีอะไรอยากให้ช่วยเหลือ สามารถเล่าให้ฟังได้เลยนะคะ\nทุกอย่างที่คุยกันในวันนี้จะถูกเก็บเป็นความลับ\nแต่ว่าฉันเป็น AI อาจมีข้อจำกัดบางอย่างที่ยังไม่สามารถทำได้\nขอให้คุณใช้วิจารญาณของคุณพิจารณาสิ่งที่ฉันพูด และตัดสินใจด้วยตัวคุณเองค่ะ")
	ms.serverSendMessage(c, "ก่อนที่จะเริ่มบทสนทนา ฉันขอทราบชื่อและอายุ หรือข้อมูลเบื้องต้น\nเพื่อเป็นข้อมูลในการประกอบการให้การปรึกษา\nแต่ข้อมูลเหล่านี้จะไม่ถูกเผยแพร่แน่นอนค่ะ")
}

func (ms *managerService) serverSendMessage(c *models.Client, message string) {
	sendMessageEventSystem := models.SendMessageEvent{
		Message:   message,
		From:      "system",
		Phase:     models.RapportPhase,
		Reasoning: "-",
	}
	payload, err := json.Marshal(sendMessageEventSystem)
	if err != nil {
		log.Fatalf("error marshalling sendMessageEvent: %v", err)
	}

	var eventSystem models.Event
	eventSystem.Type = EventSendMessage
	eventSystem.Payload = json.RawMessage(payload)
	ms.SendMessage(eventSystem, c)
}

// removeClient will remove the client and clean up
func (ms *managerService) RemoveClient(client *models.Client) {
	ms.Lock()
	defer ms.Unlock()

	// Check if Client exists, then delete it
	if _, ok := ms.clients[client]; ok {
		// Force send message
		if client.ClientData.ClientStatus == ClientLoadingStatus {
			ms.sendSuccessStatus(client)
		}
		// Store data into database
		ms.chatRepo.UploadChat(client.Chat)
		// close connection
		if client.Connection != nil {
			client.Connection.Close()
		}
		log.Println("close connection for :", client.Chat.ID.Hex())
		// remove
		delete(ms.clients, client)
		log.Println("delete client for :", client.Chat.ID.Hex())
	}
}

func (ms *managerService) SendMessageToOfflineClient(event models.Event, promptData models.PromptData, chatId string, dataError bool) {
	// Get Chat
	chatObjectID, _ := primitive.ObjectIDFromHex(chatId)
	chat, err := ms.chatRepo.GetChatById(chatObjectID)
	if err != nil {
		log.Println("send message to offline client failed : There is no chat")
		return
	}
	chat.Messages = []models.Message{}

	// New Client
	log.Println("new client for : ", chat.ID.Hex())
	clientService := NewClientService(chat, nil, ms)
	ms.clients[clientService.GetClient()] = true

	// Get cache data
	cacheData, err := ms.cacheRepo.CheckCache(chat.ID.Hex())
	if err == nil {
		clientService.GetClient().ClientData = cacheData.ClientData
		log.Println("get client data from cache")
	}

	if dataError {
		clientService.GetClient().ClientData.ServerStatus = ServerErrorStatus
		ms.cacheRepo.UpdateClientCache(clientService.GetClient().Chat.ID.Hex(), clientService.GetClient().ClientData)
	} else {
		clientService.GetClient().ClientData.ServerStatus = ServerSuccessStatus
		ms.cacheRepo.UpdateClientCache(clientService.GetClient().Chat.ID.Hex(), clientService.GetClient().ClientData)

		var chatevent models.SendMessageEvent
		if err := json.Unmarshal(event.Payload, &chatevent); err != nil {
			log.Fatalf("bad payload in request: %v", err)
			return
		}

		// Check if the message is base64 encoded and decode if it is
		var message string
		if _, err := base64.StdEncoding.DecodeString(chatevent.Message); err == nil {
			decodedMessage, err := base64.StdEncoding.DecodeString(chatevent.Message)
			if err != nil {
				log.Fatalf("failed to decode message: %v", err)
				return
			}
			message = string(decodedMessage)
		} else {
			message = chatevent.Message
		}

		// Store the message
		newMessage := ms.chatRepo.NewMessage(chatevent.From, message, chatevent.Phase, chatevent.Reasoning)

		// manage new message
		clientService.GetClient().Chat.ProvidePromptData.NumberOfSelectedQuestion = promptData.NumberOfSelectedQuestion
		clientService.GetClient().Chat.Messages = append(clientService.GetClient().Chat.Messages, newMessage)
		ms.cacheRepo.AppendMessageCache(clientService.GetClient().Chat.ID.Hex(), newMessage)

		// manage provide prompt data
		clientService.GetClient().Chat.ProvidePromptData.TimeLastMessage = newMessage.CreatedAt
		ms.cacheRepo.UpdateProvidePromptDataCache(clientService.GetClient().Chat.ID.Hex(), clientService.GetClient().Chat.ProvidePromptData)
	}
	ms.RemoveClient(clientService.GetClient())
}

func (ms *managerService) SummaryChat() {
	for {
		for c := range ms.clients {
			if c.Chat.ProvidePromptData.SummaryState || c.Chat.ProvidePromptData.Turn < 2 {
				continue
			}

			lastMessageTime := c.Chat.ProvidePromptData.TimeLastMessage.Add(5 * time.Minute)

			if time.Now().After(lastMessageTime) {
				fmt.Println("chat " + c.Chat.ID.Hex() + " is sent summary request to prompt")
				ms.sendEventToQueue("สรุปบทสนทนาให้หน่อยได้ไหม", EventSendSummaryRequestToPrompt, c)
				c.Chat.ProvidePromptData.SummaryState = true
				ms.cacheRepo.UpdateProvidePromptDataCache(c.Chat.ID.Hex(), c.Chat.ProvidePromptData)
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (ms *managerService) GetClients() models.ClientList {
	return ms.clients
}

// Event
func (es *managerService) sendMessageHandler(event models.Event, c *models.Client) error {
	// chat event
	var sendMessageEventChat models.SendMessageEvent
	if err := json.Unmarshal(event.Payload, &sendMessageEventChat); err != nil {
		log.Printf("error unmarshalling payload: %v", err)
		return err
	}

	sendMessageEventChat.From = "user"

	// Phase
	if c.Chat.ProvidePromptData.NumberOfSelectedQuestion == 0 {
		sendMessageEventChat.Phase = models.RapportPhase
	} else if c.Chat.ProvidePromptData.NumberOfSelectedQuestion < 12 {
		sendMessageEventChat.Phase = models.ExplorePhase
	} else {
		sendMessageEventChat.Phase = models.EndPhase
	}

	sendMessageEventChat.Reasoning = ""

	updatedPayload, err := json.Marshal(sendMessageEventChat)
	if err != nil {
		log.Printf("error marshalling updated payload: %v", err)
		return err
	}
	event.Payload = updatedPayload
	es.SendMessage(event, c)

	c.ClientData.ClientStatus = ClientLoadingStatus
	c.ClientData.Message = c.ClientData.Message + " " + sendMessageEventChat.Message
	es.cacheRepo.UpdateClientCache(c.Chat.ID.Hex(), c.ClientData)
	return nil
}

// SendMessageHandler will send out a message to all other participants in the chat
func (es *managerService) SendMessage(event models.Event, c *models.Client) error {
	// Marshal Payload into wanted format
	var chatevent models.SendMessageEvent
	if err := json.Unmarshal(event.Payload, &chatevent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	// Check if the message is base64 encoded and decode if it is
	var message string
	if _, err := base64.StdEncoding.DecodeString(chatevent.Message); err == nil {
		decodedMessage, err := base64.StdEncoding.DecodeString(chatevent.Message)
		if err != nil {
			return fmt.Errorf("failed to decode message: %v", err)
		}
		message = string(decodedMessage)
	} else {
		message = chatevent.Message
	}

	// Store the message
	newMessage := es.chatRepo.NewMessage(chatevent.From, message, chatevent.Phase, chatevent.Reasoning)

	// manage new message
	c.Chat.Messages = append(c.Chat.Messages, newMessage)
	es.cacheRepo.AppendMessageCache(c.Chat.ID.Hex(), newMessage)

	// manage provide prompt data
	c.Chat.ProvidePromptData.TimeLastMessage = newMessage.CreatedAt
	es.cacheRepo.UpdateProvidePromptDataCache(c.Chat.ID.Hex(), c.Chat.ProvidePromptData)

	data, err := json.Marshal(newMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %v", err)
	}

	// Place payload into an Event
	var outgoingEvent models.Event
	outgoingEvent.Payload = data
	outgoingEvent.Type = EventNewMessage

	// Broadcast to client
	c.Egress <- outgoingEvent
	return nil
}

func (es *managerService) updateFeedbackHandler(event models.Event, c *models.Client) error {
	// Marshal Payload into wanted format
	var feedbackevent models.UpdateFeedbackEvent
	if err := json.Unmarshal(event.Payload, &feedbackevent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	// update feedback in clinet
	feedbackObjectID, err := primitive.ObjectIDFromHex(feedbackevent.ID)
	if err != nil {
		return fmt.Errorf("failed to transform Id to objectID: %v", err)
	}
	feedback := es.chatRepo.CreateFeedback(feedbackObjectID, feedbackevent.Message, feedbackevent.Score)
	log.Println(c.Chat)
	_, err = es.updateFeedback(c, feedback)
	if err != nil {
		return fmt.Errorf("failed to update feedback: %v", err)
	}
	log.Println(c.Chat)

	data, err := json.Marshal(feedback)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %v", err)
	}

	// Place payload into an Event
	var outgoingEvent models.Event
	outgoingEvent.Payload = data
	outgoingEvent.Type = EventNewFeedback

	// Broadcast to client
	c.Egress <- outgoingEvent
	return nil
}

func (es *managerService) updateFeedback(c *models.Client, feedback models.Feedback) (models.Feedback, error) {
	// Check the main Feedback in Chat
	if c.Chat.Feedback.ID == feedback.ID {
		c.Chat.Feedback = feedback
		if err := es.chatRepo.UpdateChatFeedback(c.Chat.ID, feedback); err != nil {
			return models.Feedback{}, err
		}
		return feedback, nil
	}

	// Check the Feedbacks in Messages
	for i := range c.Chat.Messages {
		if c.Chat.Messages[i].Feedback.ID == feedback.ID {
			c.Chat.Messages[i].Feedback = feedback
			return feedback, nil
		}
	}

	// Check the Feedback in database
	if err := es.chatRepo.UpdateMessageFeedback(c.Chat.ID, feedback); err != nil {
		return models.Feedback{}, err
	}
	return feedback, nil
}

func (es *managerService) sendStatusHandler(event models.Event, c *models.Client) error {
	var sendStatusEvent models.SendStatusEvent
	if err := json.Unmarshal(event.Payload, &sendStatusEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	if sendStatusEvent.Status == ClientSuccessStatus && c.ClientData.Message != "" {
		es.sendSuccessStatus(c)
		return nil
	}
	return fmt.Errorf("bad payload in request")
}

func (es *managerService) sendSuccessStatus(c *models.Client) {
	msg := c.ClientData.Message
	c.Chat.ProvidePromptData.Turn += 1
	es.cacheRepo.UpdateProvidePromptDataCache(c.Chat.ID.Hex(), c.Chat.ProvidePromptData)

	c.ClientData.Message = ""
	c.ClientData.ClientStatus = ClientSuccessStatus
	c.ClientData.ServerStatus = ServerLoadingStatus
	es.cacheRepo.UpdateClientCache(c.Chat.ID.Hex(), c.ClientData)

	es.sendEventToQueue(msg, EventSendMessageToPrompt, c)
}

func (es *managerService) sendEventToQueue(msg string, eventType string, c *models.Client) {
	// system event for ai service test
	sendMessageEventPrompt := models.SendMessageEvent{
		Message: msg,
		From:    c.Chat.ID.Hex(),
	}
	payload, err := json.Marshal(sendMessageEventPrompt)
	if err != nil {
		log.Fatalf("error marshalling sendMessageEvent: %v", err)
	}

	var event models.Event
	event.Payload = payload
	event.Type = eventType

	outgoingEvent, err := json.Marshal(event)
	if err != nil {
		log.Fatalf("error marshalling sendMessageEvent: %v", err)
	}

	es.queuePublisher.SQSSendMessage(outgoingEvent)
	es.NewStatusHandler(ServerLoadingStatus, c)
}

// ---- Event of AI-Service ----
func (es *managerService) NewStatusHandler(status string, c *models.Client) error {
	statusMessageEvent := models.StatusMessageEvent{
		Status: status,
	}

	statusPayload, err := json.Marshal(statusMessageEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %v", err)
	}

	c.ClientData.ServerStatus = status
	es.cacheRepo.UpdateClientCache(c.Chat.ID.Hex(), c.ClientData)

	// Place payload into an Event
	var outgoingEvent models.Event
	outgoingEvent.Payload = statusPayload
	outgoingEvent.Type = EventNewStatus

	// Broadcast to client
	c.Egress <- outgoingEvent
	return nil
}

func (es *managerService) NewSummarySheetHandler(chatId string) error {
	for client := range es.clients {
		if client.Chat.ID.Hex() == chatId {
			// Place payload into an Event
			var outgoingEvent models.Event
			outgoingEvent.Type = EventNewSummarySheet

			// Broadcast to client
			client.Egress <- outgoingEvent
			return nil
		}
	}

	return fmt.Errorf("failed to find client with chatId: %s", chatId)
}
