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
	"github.com/Meeyok-Chat/backend/repository/database"
	"github.com/Meeyok-Chat/backend/repository/queue/queuePublisher"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type managerService struct {
	clients  models.ClientList
	chatRepo database.ChatRepo
	userRepo database.UserRepo

	queuePublisher queuePublisher.QueuePublisher
	// Using a syncMutex here to be able to lcok state before editing clients
	// Could also use Channels to block
	sync.RWMutex
	// handlers are functions that are used to handle Events
	handlers map[string]models.EventHandler
}

// Manager is used to hold references to all Clients Registered, and Broadcasting etc
type ManagerService interface {
	AddClient(conn *websocket.Conn, c *gin.Context, userID string)
	RemoveClient(client *models.Client)
	RouteEvent(event models.Event, c *models.Client) error
	CheckOldClient(userID string) error
	GetClients() []models.User

	SendMessageHandler(event models.Event, c *models.Client) error
}

// NewManager is used to initalize all the values inside the manager
func NewManagerService(queuePublisher queuePublisher.QueuePublisher, chatRepo database.ChatRepo, userRepo database.UserRepo) ManagerService {
	m := &managerService{
		clients:        make(models.ClientList),
		chatRepo:       chatRepo,
		userRepo:       userRepo,
		queuePublisher: queuePublisher,
		handlers:       make(map[string]models.EventHandler),
	}
	m.setupEventHandlers()
	return m
}

// setupEventHandlers configures and adds all handlers
func (ms *managerService) setupEventHandlers() {
	ms.handlers[models.EventSendMessage] = ms.SendMessageHandler
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

func (ms *managerService) CheckOldClient(userID string) error {
	// Check if there is an existing client for the same chat and wait for it to be removed
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for {
		existingClientFound := false

		for client := range ms.clients {
			if client.User.ID.Hex() == userID {
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
			return fmt.Errorf("failed to remove existing client for chat: %s within timeout", userID)
		default:
			log.Println("Waiting for existing client to be removed")
			time.Sleep(500 * time.Millisecond)
		}
	}
	return nil
}

// addClient will add clients to our clientList
func (ms *managerService) AddClient(conn *websocket.Conn, c *gin.Context, userID string) {
	user, err := ms.userRepo.GetUserByID(userID)
	if err != nil {
		log.Println(err)
	}
	log.Println("New connection with userID : " + user.ID.Hex())

	// Create New Client
	clientService := NewClientService(user, conn, ms)

	go clientService.ReadMessages()
	go clientService.WriteMessages()

	// Lock so we can manipulate
	ms.Lock()
	defer ms.Unlock()

	// Add Client
	ms.clients[clientService.GetClient()] = true

	ms.SendUserStatusHandler(user.ID.Hex(), models.EventNewUser)
}

func (ms *managerService) GetClients() []models.User {
	var users []models.User
	for client := range ms.clients {
		users = append(users, client.User)
	}

	if users == nil {
		return []models.User{}
	}
	return users
}

// removeClient will remove the client and clean up
func (ms *managerService) RemoveClient(client *models.Client) {
	ms.Lock()
	defer ms.Unlock()

	// Check if Client exists, then delete it
	if _, ok := ms.clients[client]; ok {
		// Store data into database
		// ms.chatRepo.UploadChat(client.User)
		// close connection
		if client.Connection != nil {
			client.Connection.Close()
		}
		log.Println("close connection for :", client.User.ID.Hex())
		// remove
		delete(ms.clients, client)
		log.Println("delete client for :", client.User.ID.Hex())

		ms.SendUserStatusHandler(client.User.ID.Hex(), models.EventLeaveUser)
	}
}

// Event
func (ms *managerService) SendMessageHandler(event models.Event, c *models.Client) error {
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

	chatevent.CreatedAt = time.Now()

	// Get Chat
	chat, err := ms.chatRepo.GetChatByID(chatevent.ChatID)
	if err != nil {
		return fmt.Errorf("failed to get chat")
	}

	// Store the message
	newMessage := ms.chatRepo.NewMessage(message, chatevent.From)

	// manage new message
	chat.Messages = append(chat.Messages, newMessage)
	err = ms.chatRepo.AppendMessage(chat.ID.Hex(), newMessage)
	if err != nil {
		return fmt.Errorf("failed to append new message to database: %v", err)
	}

	data, err := json.Marshal(chatevent)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %v", err)
	}

	// Send message to Meeyok AI
	if message == "@Meeyok AI" {
		ms.sendEventToQueue(chatevent.ChatID)
	}

	// Place payload into an Event
	var outgoingEvent models.Event
	outgoingEvent.Payload = data
	outgoingEvent.Type = models.EventNewMessage

	for client := range ms.clients {
		for _, userID := range chat.Users {
			if client.User.ID.Hex() == userID {
				client.Egress <- outgoingEvent
			}
		}
	}
	return nil
}

func (ms *managerService) sendEventToQueue(chatID string) {
	// system event for ai service test
	queuePublisherPayload := models.QueuePublisherPayload{
		From: chatID,
	}
	payload, err := json.Marshal(queuePublisherPayload)
	if err != nil {
		log.Fatalf("error marshalling sendMessageEvent: %v", err)
	}

	var event models.Event
	event.Payload = payload
	event.Type = models.EventSendMessageToMeeyok

	outgoingEvent, err := json.Marshal(event)
	if err != nil {
		log.Fatalf("error marshalling sendMessageEvent: %v", err)
	}

	ms.queuePublisher.SQSSendMessage(outgoingEvent)
}

// Event
func (ms *managerService) SendUserStatusHandler(userId string, eventType string) error {
	payload := models.NewUserStatusEvent{
		UserID: userId,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %v", err)
	}

	// Place payload into an Event
	var outgoingEvent models.Event
	outgoingEvent.Payload = data
	outgoingEvent.Type = eventType

	for client := range ms.clients {
		client.Egress <- outgoingEvent
	}
	return nil
}
