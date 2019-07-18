package sender

import (
	"context"

	"gitlab.benzinga.io/benzinga/ftp-engine/process"
)

// Sender describes a data sender
type Sender interface {
	Send(ctx context.Context, data *process.Output) error
	Status() error
	Close() error
}
