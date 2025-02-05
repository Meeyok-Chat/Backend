package queueReceiver

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	"github.com/Meeyok-Chat/backend/configs"
	"github.com/Meeyok-Chat/backend/models"
	"github.com/Meeyok-Chat/backend/repository/cache"
	Websocket "github.com/Meeyok-Chat/backend/services/websocket"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type QueueReceiver struct {
	sqsSvc *sqs.SQS

	sync.RWMutex

	managerService Websocket.ManagerService
	cacheRepo      cache.CacheRepo
}

const (
	ResultQueue = "AWS_SQS_RESULT_URL"
	DLQueue     = "AWS_SQS_DEADLETTERQUEUE_URL"
)

// AWS SQS
func ConnectSQS() *sqs.SQS {
	accessKeyID := configs.GetEnv("AWS_ACCESS_KEY_ID")
	secretAccessKey := configs.GetEnv("AWS_SECRET_ACCESS_KEY")

	sess, _ := session.NewSession(&aws.Config{
		Region:      aws.String(configs.GetEnv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
		MaxRetries:  aws.Int(5),
	})

	sqsSvc := sqs.New(sess)

	return sqsSvc
}

func NewConsumerManager(managerService Websocket.ManagerService, cacheRepo cache.CacheRepo) *QueueReceiver {
	cm := &QueueReceiver{
		sqsSvc:         ConnectSQS(),
		managerService: managerService,
		cacheRepo:      cacheRepo,
	}
	return cm
}

func (cm *QueueReceiver) SQSSendMessage(message []byte, QueueUrl string) {
	_, err := cm.sqsSvc.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(string(message)),
		QueueUrl:    aws.String(QueueUrl),
	})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Sending to publisher queue message successfully")
}

func (cm *QueueReceiver) ReadResult() {
	for {
		cm.Read(ResultQueue)
	}
}

func (cm *QueueReceiver) ReadDLQ() {
	for {
		cm.Read(DLQueue)
	}
}

func (cm *QueueReceiver) Read(queueUrl string) {
	chnMessages := make(chan *sqs.Message, 2)
	go cm.PollMessages(chnMessages, queueUrl)

	for message := range chnMessages {
		var parsedEvent models.Event
		if err := json.Unmarshal([]byte(*message.Body), &parsedEvent); err != nil {
			log.Println(err)
			continue
		}
		var payload models.QueueReceiverPayload
		if err := json.Unmarshal(parsedEvent.Payload, &payload); err != nil {
			log.Println(err)
			continue
		}

		chatId := payload.PromptData.From

		var messageData models.SendMessageEvent
		messageData.Message = payload.PromptData.Message
		messageData.From = "system"
		messageData.Phase = payload.PromptData.Phase
		messageData.Reasoning = payload.PromptData.Reasoning

		parsedEvent.Payload, _ = json.Marshal(messageData)

		cm.SendMessageToClient(parsedEvent, payload.PromptData, chatId, queueUrl)
		cm.SQSDeleteMessage(message, configs.GetEnv(queueUrl))
	}
}

func (cm *QueueReceiver) PollMessages(chn chan<- *sqs.Message, queueUrl string) {
	sqsSvc := ConnectSQS()
	for {
		output, err := sqsSvc.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(os.Getenv(queueUrl)),
			MaxNumberOfMessages: aws.Int64(2),
			WaitTimeSeconds:     aws.Int64(15),
		})

		if err != nil {
			log.Println(err)
		}

		for _, message := range output.Messages {
			chn <- message
		}
	}
}

func (cm *QueueReceiver) SendMessageToClient(event models.Event, promptData models.PromptData, from string, queueUrl string) {
	clients := cm.managerService.GetClients()

	if clients == nil {
		return
	}

	for client := range clients {
		if client.Chat.ID.Hex() == from {
			log.Println("Bot replied to client")
			if queueUrl == DLQueue {
				cm.managerService.NewStatusHandler(Websocket.ServerErrorStatus, client)
			} else if queueUrl == ResultQueue {
				cm.managerService.NewStatusHandler(Websocket.ServerSuccessStatus, client)

				// update NumberOfSelectedQuestion in cache
				client.Chat.ProvidePromptData.NumberOfSelectedQuestion = promptData.NumberOfSelectedQuestion
				cm.cacheRepo.UpdateProvidePromptDataCache(client.Chat.ID.Hex(), client.Chat.ProvidePromptData)

				cm.managerService.SendMessage(event, client)
			}
			clients = nil
			return
		}
	}
	log.Println("Bot replied to offline client")
	if queueUrl == DLQueue {
		cm.managerService.SendMessageToOfflineClient(event, promptData, from, true)
	} else {
		cm.managerService.SendMessageToOfflineClient(event, promptData, from, false)
	}
	clients = nil
}

func (cm *QueueReceiver) SQSDeleteMessage(msg *sqs.Message, queueUrl string) {
	cm.sqsSvc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueUrl),
		ReceiptHandle: msg.ReceiptHandle,
	})
}
