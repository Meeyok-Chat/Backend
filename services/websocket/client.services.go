package Websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Meeyok-Chat/backend/models"
	"github.com/gorilla/websocket"
)

var (
	// pongWait is how long we will await a pong response from client
	pongWait = 10 * time.Second
	// pingInterval has to be less than pongWait, We cant multiply by 0.9 to get 90% of time
	// Because that can make decimals, so instead *9 / 10 to get 90%
	// The reason why it has to be less than PingRequency is becuase otherwise it will send a new Ping before getting response
	pingInterval = (pongWait * 9) / 10
)

type clientService struct {
	client  *models.Client
	manager ManagerService
}

type ClientService interface {
	ReadMessages()
	WriteMessages()
	GetClient() *models.Client
}

// NewClient is used to initialize a new Client with all required values initialized
func NewClientService(user models.User, conn *websocket.Conn, manager ManagerService) ClientService {
	return &clientService{
		client: &models.Client{
			User:       user,
			ClientData: models.ClientData{},
			Connection: conn,
			Egress:     make(chan models.Event),
		},
		manager: manager,
	}
}

// readMessages will start the client to read messages and handle them
// appropriatly.
// This is suppose to be ran as a goroutine
func (cs *clientService) ReadMessages() {
	defer func() {
		// Graceful Close the Connection once this
		// function is done
		cs.manager.RemoveClient(cs.client)
	}()
	// Set Max Size of Messages in Bytes
	cs.client.Connection.SetReadLimit(512)
	// Configure Wait time for Pong response, use Current time + pongWait
	// This has to be done here to set the first initial timer.
	if err := cs.client.Connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println(err)
		return
	}
	// Configure how to handle Pong responses
	cs.client.Connection.SetPongHandler(cs.pongHandler)

	// Loop Forever
	for {
		// ReadMessage is used to read the next message in queue
		// in the connection
		_, payload, err := cs.client.Connection.ReadMessage()

		if err != nil {
			// If Connection is closed, we will Recieve an error here
			// We only want to log Strange errors, but simple Disconnection
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break // Break the loop to close conn & Cleanup
		}
		// Marshal incoming data into a Event struct
		var request models.Event
		if err := json.Unmarshal(payload, &request); err != nil {
			log.Printf("error marshalling message: %v", err)
			break // Breaking the connection here might be harsh xD
		}
		// Route the Event
		if err := cs.manager.RouteEvent(request, cs.client); err != nil {
			log.Println("Error handeling Message: ", err)
		}
	}
}

// pongHandler is used to handle PongMessages for the Client
func (cs *clientService) pongHandler(pongMsg string) error {
	// Current time + Pong Wait time
	log.Println("pong")
	return cs.client.Connection.SetReadDeadline(time.Now().Add(pongWait))
}

// writeMessages is a process that listens for new messages to output to the Client
func (cs *clientService) WriteMessages() {
	// Create a ticker that triggers a ping at given interval
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		// Graceful close if this triggers a closing
		cs.manager.RemoveClient(cs.client)
	}()

	for {
		select {
		case message, ok := <-cs.client.Egress:
			// Ok will be false Incase the egress channel is closed
			if !ok {
				// Manager has closed this connection channel, so communicate that to frontend
				if err := cs.client.Connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					// Log that the connection is closed and the reason
					log.Println("connection closed: ", err)
				}
				// Return to close the goroutine
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Println(err)
				return // closes the connection, should we really
			}
			// Write a Regular text message to the connection
			if err := cs.client.Connection.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println(err)
			}

			log.Println("sent message")
		case <-ticker.C:
			log.Println("ping")
			// Send the Ping
			if err := cs.client.Connection.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Println("writemsg: ", err)
				return // return to break this goroutine triggeing cleanup
			}
		}

	}
}

func (cs *clientService) GetClient() *models.Client {
	return cs.client
}
