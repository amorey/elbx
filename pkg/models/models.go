package models

import (
	"encoding/json"
)

// TODO:
// Spot notice schema: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/spot-interruptions.html
// ASG event schema: https://docs.aws.amazon.com/autoscaling/ec2/userguide/cloud-watch-events.html#terminate-lifecycle-action

type EventBridgeEvent struct {
	Version    string          `json:"version"`
        ID         string          `json:"id"`
        DetailType string          `json:"detail-type"`
        Source     string          `json:"source"`
        Account    string          `json:"account"`
        Time       string          `json:"time"`
        Region     string          `json:"region"`
        Resources  []string        `json:"resources"`
        Detail     json.RawMessage `json:"detail"`
}

type LifecycleDetail struct {
        LifecycleActionToken string `json:"LifecycleActionToken"`
        AutoScalingGroupName string `json:"AutoScalingGroupName"`
        LifecycleHookName    string `json:"LifecycleHookName"`
        EC2InstanceID        string `json:"EC2InstanceId"`
        LifecycleTransition  string `json:"LifecycleTransition"`
}

type SpotInterruptionDetail struct {
        InstanceID     string `json:"instance-id"`
        InstanceAction string `json:"instance-action"`
}
