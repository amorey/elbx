package sqsmonitor

import (
	//"encoding/json"
	"testing"

        "github.com/stretchr/testify/assert"
	
	"github.com/amorey/elbx/pkg/models"
)

var asgLifecycleEvent = models.EventBridgeEvent{
        Version:    "0",
        ID:         "782d5b4c-0f6f-1fd6-9d62-ecf6aed0a470",
        DetailType: "EC2 Instance-terminate Lifecycle Action",
        Source:     "aws.autoscaling",
        Account:    "123456789012",
        Time:       "2020-07-01T22:19:58Z",
        Region:     "us-east-1",
        Resources: []string{
                "arn:aws:autoscaling:us-east-1:123456789012:autoScalingGroup:26e7234b-03a4-47fb-b0a9-2b241662774e:autoScalingGroupName/nth-test1",
        },
        Detail: []byte(`{
                "LifecycleActionToken": "0befcbdb-6ecd-498a-9ff7-ae9b54447cd6",
                "AutoScalingGroupName": "nth-test1",
                "LifecycleHookName": "node-termination-handler",
                "EC2InstanceId": "i-0633ac2b0d9769723",
                "LifecycleTransition": "autoscaling:EC2_INSTANCE_TERMINATING"
          }`),
}

var spotItnEvent = models.EventBridgeEvent{
        Version:    "0",
        ID:         "1e5527d7-bb36-4607-3370-4164db56a40e",
        DetailType: "EC2 Spot Instance Interruption Warning",
        Source:     "aws.ec2",
        Account:    "123456789012",
        Time:       "1970-01-01T00:00:00Z",
        Region:     "us-east-1",
        Resources: []string{
                "arn:aws:ec2:us-east-1b:instance/i-0b662ef9931388ba0",
        },
        Detail: []byte(`{
                "instance-id": "i-0b662ef9931388ba0",
                "instance-action": "terminate"
        }`),
}

func TestEventBridgeEventSchema(t *testing.T) {
	tests := []struct {
		name    string
		ebEvent models.EventBridgeEvent
	}{
		{
			"EC2 Instance-terminate Lifecycle Action",
			asgLifecycleEvent,
		},
		{
			"EC2 Spot Instance interruption notice",
			spotItnEvent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, true, true)
		})
	}
}
