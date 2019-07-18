package worker

import (
	"context"
	"time"

	"gitlab.benzinga.io/benzinga/content-models/models"
)

type FTPDeliveryRecord struct {
	NodeID          int64
	EventID         int64
	EventType       models.EventType
	ConsumerGroupID string
	FTPHost         string
	FTPUsername     string
	FTPPath         string
	Filename        string
	SHA256Checksum  string
	Timestamp       time.Time
	SizeBytes       int
}

type Worker interface {
	Work(ctx context.Context)
}
