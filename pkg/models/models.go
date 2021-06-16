package models

type EventBridgeEvent struct {
        Detail EventBridgeEventDetail `json:"detail"`
}

type EventBridgeEventDetail struct {
	InstanceId string `json:"instance-id"`
}
