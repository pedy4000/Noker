package queue

import (
	"github.com/pedy4000/noker/internal/repository"
	"github.com/pedy4000/noker/pkg/config"

	"github.com/google/uuid"
)

type Job struct {
	MeetingID uuid.UUID
}

type Processor interface {
	Enqueue(meetingID uuid.UUID)
	Start()
	Stop()
}

func NewProcessor(queries *repository.Queries, cfg *config.Config) Processor {
	var p Processor

	switch cfg.Queue.Type {
	case "inmemory":
		p = NewInMemoryWorker(queries, cfg)
	default:
		p = NewInMemoryWorker(queries, cfg)
	}
	return p
}

// TODO: we can easily add other processor like kafka in case of need
