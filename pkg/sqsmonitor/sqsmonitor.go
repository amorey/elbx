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

type Monitor struct {
	queueUrl  *string
	sqsClient *sqs.Client
}

// Execute SQS long-poll and return when timeout is reached or a message is found
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

// Process new SQS messages
func (m *Monitor) processSQSMessage(ctx context.Context, commsChan chan<- string, message *types.Message) error {
	log.Info().Msg(fmt.Sprintf("Processing SQS message (%s)", *message.MessageId))

	// init event instance
	event := models.EventBridgeEvent{}
        err := json.Unmarshal([]byte(*message.Body), &event)
        if err != nil {
                return err
        }

	var instanceId string
	
	// extract instance-id
	switch event.DetailType {
	case "EC2 Instance-terminate Lifecycle Action":
		detail := models.LifecycleDetail{}
		err = json.Unmarshal([]byte(event.Detail), &detail)
		if err != nil {
			return err
		}
		instanceId = detail.EC2InstanceID
	case "EC2 Spot Instance Interruption Warning":
		detail := models.SpotInterruptionDetail{}
		err = json.Unmarshal([]byte(event.Detail), &detail)
		if err != nil {
			return err
		}
		instanceId = detail.InstanceID
	}
	
	// broadcast event
	commsChan <- instanceId
	
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

// Continuously poll SQS for new messages until program is interrupted
func (m *Monitor) WatchForSQSMessages(ctx context.Context, commsChan chan<- string, wg *sync.WaitGroup) {
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
			err := m.processSQSMessage(ctx, commsChan, &message)
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
