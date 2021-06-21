package models

// TODO:
// Spot notice schema: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/spot-interruptions.html
// ASG event schema: https://docs.aws.amazon.com/autoscaling/ec2/userguide/cloud-watch-events.html#terminate-lifecycle-action

type EventBridgeEvent struct {
        Detail EventBridgeEventDetail `json:"detail"`
}

type EventBridgeEventDetail struct {
	InstanceId string `json:"EC2InstanceId"`
}
