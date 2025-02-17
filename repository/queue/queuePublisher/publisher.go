package queuePublisher

import (
	"log"

	"github.com/Meeyok-Chat/backend/configs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const (
	PublisherQueue = "AWS_SQS_PUBLISHER"
)

type queuePublisher struct {
	sqsSvc *sqs.SQS
}

type QueuePublisher interface {
	SQSSendMessage(message []byte)
}

func NewQueuePublisher() QueuePublisher {
	return &queuePublisher{
		sqsSvc: ConnectSQS(),
	}
}

// AWS SQS
func ConnectSQS() *sqs.SQS {
	accessKeyId := configs.GetEnv("AWS_ACCESS_KEY_ID")
	secretAccessKey := configs.GetEnv("AWS_SECRET_ACCESS_KEY")

	sess, _ := session.NewSession(&aws.Config{
		Region:      aws.String(configs.GetEnv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(accessKeyId, secretAccessKey, ""),
		MaxRetries:  aws.Int(5),
	})

	sqsSvc := sqs.New(sess)

	return sqsSvc
}

func (qb *queuePublisher) SQSSendMessage(message []byte) {
	_, err := qb.sqsSvc.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(string(message)),
		QueueUrl:    aws.String(configs.GetEnv(PublisherQueue)),
	})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Sending to publisher queue message successfully")
}
