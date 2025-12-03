package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pedy4000/noker/internal/ai"
	"github.com/pedy4000/noker/internal/models"
	"github.com/pedy4000/noker/internal/repository"
	"github.com/pedy4000/noker/internal/service"
	"github.com/pedy4000/noker/pkg/config"
	"github.com/pedy4000/noker/pkg/logger"

	"github.com/google/uuid"
)

type InMemoryWorker struct {
	queries      *repository.Queries
	extractor    ai.Provider
	service      *service.OpportunityService
	cfg          *config.Config
	jobs         chan Job
	wg           sync.WaitGroup
	shutdown     chan struct{}
	shutdownOnce sync.Once
}

func NewInMemoryWorker(queries *repository.Queries, cfg *config.Config) *InMemoryWorker {
	return &InMemoryWorker{
		queries:   queries,
		extractor: ai.NewExtractor(cfg),
		service:   service.NewOpportunityService(queries),
		cfg:       cfg,
		jobs:      make(chan Job, cfg.Queue.BufferSize),
		shutdown:  make(chan struct{}),
	}
}

func (w *InMemoryWorker) Enqueue(meetingID uuid.UUID) {
	select {
	case w.jobs <- Job{MeetingID: meetingID}:
		logger.Debug("Job enqueued for meeting:", meetingID)
	default:
		logger.Error("Job queue is full! Dropping job for meeting:", meetingID)
	}
}

func (w *InMemoryWorker) Start() {
	for i := 0; i < w.cfg.Queue.WorkerCount; i++ {
		w.wg.Add(1)
		go func(id int) {
			defer w.wg.Done()
			logger.Info("Worker", id, "started")
			for {
				select {
				case <-w.shutdown:
					logger.Info("Worker", id, "shutting down")
					return
				case job := <-w.jobs:
					w.processJob(job)
				}
			}
		}(i)
	}

	// Health ping
	go func() {
		ticker := time.NewTicker(time.Duration(w.cfg.Queue.PollIntervalMs) * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-w.shutdown:
				return
			case <-ticker.C:
				// Keep connection alive
			}
		}
	}()
}

func (w *InMemoryWorker) processJob(job Job) {
	ctx := context.Background()
	meeting, err := w.queries.GetMeeting(ctx, job.MeetingID)
	if err != nil {
		logger.Error("Failed to fetch meeting:", job.MeetingID, err)
		return
	}

	w.setMeetingStatus(meeting.ID, "processing")

	m := &models.Meeting{
		ID:       meeting.ID,
		Title:    meeting.Title,
		RawNotes: meeting.RawNotes,
		Source:   models.MeetingSource(meeting.Source),
	}

	opps, err := w.queries.ListAllOpportunitiesForDeduplication(ctx)
	if err != nil {
		logger.Error("Failed to fetch opportunities:", job.MeetingID, err)
		return
	}

	extracted, err := w.extractor.Extract(ctx, m, opps)
	if err != nil {
		w.setMeetingStatus(meeting.ID, "failed",
			fmt.Sprintf("AI extraction failed: %v", err))
		logger.Error("AI extraction failed for meeting", job.MeetingID, err)
		return
	}

	if err := w.service.ProcessExtractedOpportunities(ctx, job.MeetingID, extracted); err != nil {
		w.setMeetingStatus(meeting.ID, "failed",
			fmt.Sprintf("Failed to save opportunities: %v", err))
		logger.Error("Failed to save opportunities:", err)
		return
	}

	w.setMeetingStatus(meeting.ID, "done")
	logger.Debug("Successfully processed meeting:", job.MeetingID, "→", len(extracted), "opportunities")
}

func (w *InMemoryWorker) Stop() {
	w.shutdownOnce.Do(func() {
		close(w.shutdown)
		close(w.jobs)
		w.wg.Wait()
		logger.Info("All workers stopped")
	})
}

func (w *InMemoryWorker) setMeetingStatus(meetingID uuid.UUID, status string, errMsg ...string) {
	ctx := context.Background()

	params := repository.UpdateMeetingStatusParams{
		ID:               meetingID,
		ProcessingStatus: status,
		Column3:          "",
	}

	if len(errMsg) > 0 {
		message := errMsg[0]
		if len(message) > 200 {
			message = message[:197] + "..." // Database limitation
		}
		params.Column3 = message
	}

	if err := w.queries.UpdateMeetingStatus(ctx, params); err != nil {
		logger.Error("Failed to update meeting status", "meeting_id", meetingID, "error", err)
	}
}
