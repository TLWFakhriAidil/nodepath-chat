package broadcast

import (
	domainBroadcast "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/broadcast"
)

// IBroadcastManager defines the interface for broadcast managers
type IBroadcastManager interface {
	SendMessage(msg domainBroadcast.BroadcastMessage) error
	GetOrCreateWorker(deviceID string) *DeviceWorker
	GetWorkerStatus(deviceID string) (domainBroadcast.WorkerStatus, bool)
	GetAllWorkerStatus() []domainBroadcast.WorkerStatus
	StopAllWorkers() error
	StopWorker(deviceID string) error
	ResumeFailedWorkers() error
	CheckWorkerHealth()
}