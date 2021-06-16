package sqsmonitor

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/rs/zerolog/log"

	"github.com/amorey/elbx/pkg/models"
)

var queueName = "elbx-test"

type Monitor struct {
	queueUrl  *string
	sqsClient *sqs.Client
}

func (m *Monitor) receiveSQSMessages(ctx context.Context) (*[]types.Message, error) {
	log.Debug().Msg("Starting SQS long-poll")

	// poll for messages
	result, err := m.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		MessageAttributeNames: []string{
			string(types.QueueAttributeNameAll),
		},
		QueueUrl:            m.queueUrl,
		MaxNumberOfMessages: 5,
		VisibilityTimeout:   int32(20),
		WaitTimeSeconds:     int32(10),
	})
	if err != nil {
		return nil, err
	}

	log.Debug().Msg(fmt.Sprintf("Finished SQS long-poll: found %d messages", len(result.Messages)))
	
	return &result.Messages, nil
}

func (m *Monitor) processSQSMessage(ctx context.Context, eventChan chan<- models.EventBridgeEvent, message *types.Message) error {
	log.Info().Msg(fmt.Sprintf("Processing SQS message (%s)", *message.MessageId))

	// init event instance
	event := models.EventBridgeEvent{}
        err := json.Unmarshal([]byte(*message.Body), &event)
        if err != nil {
                return err
        }

	// broadcast event
	eventChan <- event
	
	// delete message
	_, err = m.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		ReceiptHandle: message.ReceiptHandle,
		QueueUrl: m.queueUrl,
	})
	if err != nil {
		return err
	}

	log.Debug().Msg(fmt.Sprintf("Successfully deleted SQS message (%s)", *message.MessageId))

	return nil
}

func (m *Monitor) WatchForSQSMessages(ctx context.Context, eventChan chan<- models.EventBridgeEvent, wg *sync.WaitGroup) {
	defer wg.Done()

Loop:
	for {
		// get messages
		messages, err := m.receiveSQSMessages(ctx)
		if ctx.Err() != nil {
			break Loop
		} else if err != nil {
			log.Error().Msg(err.Error())
			time.Sleep(10 * time.Second)
			continue
		}

		// process messages
		for _, message := range *messages {
			err := m.processSQSMessage(ctx, eventChan, &message)
			if ctx.Err() != nil {
				break Loop
			} else if err != nil {
				log.Error().Msg(err.Error())
			}
		}
	}

	log.Debug().Msg("WatchForSQSMessages exiting")
}

func New(queueUrl *string) (*Monitor, error) {
	// TODO: should context use a timeout?
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	// init sqs client and fetch queue url
	m := Monitor{
		queueUrl: queueUrl,
		sqsClient: sqs.NewFromConfig(cfg),
	}

	return &m, nil
}
