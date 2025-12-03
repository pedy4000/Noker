package api

import (
	"github.com/pedy4000/noker/internal/models"

	"github.com/google/uuid"
)

// Request models
type CreateMeetingRequest struct {
	Title    string               `json:"title" validate:"required,min=3,max=100"`
	Notes    string               `json:"notes" validate:"required,min=10,max=10000"`
	Source   models.MeetingSource `json:"source,omitempty" validate:"omitempty,oneof=manual notion file upload"`
	Metadata map[string]any       `json:"metadata,omitempty"`
}

// Response models
type CreateMeetingResponse struct {
	Status      string    `json:"status"`
	MeetingID   uuid.UUID `json:"meeting_id"`
	Message     string    `json:"message"`
	ProcessedIn string    `json:"processed_in"`
}

type MeetingResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Source    string    `json:"source"`
	Metadata  string    `json:"metadata,omitempty"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Created   string    `json:"created"`
	Processed string    `json:"processed"`
	Updated   string    `json:"updated"`
}

type OpportunityResponse struct {
	ID            uuid.UUID          `json:"id"`
	UserSegment   string             `json:"user_segment"`
	Struggle      string             `json:"struggle"`
	Why           string             `json:"why,omitempty"`
	Workaround    string             `json:"workaround,omitempty"`
	Theme         string             `json:"theme,omitempty"`
	EvidenceCount int                `json:"evidence_count"`
	Evidences     []EvidenceResponse `json:"evidence,omitempty"`
	Created       string             `json:"created"`
}

type EvidenceResponse struct {
	ID      uuid.UUID `json:"id"`
	Quote   string    `json:"quote"`
	Context string    `json:"context"`
	Created string    `json:"created"`

	MeetingID    uuid.UUID `json:"meeting_id"`
	MeetingTitle string    `json:"meeting_title"`
	MeetingDate  string    `json:"meeting_date"`
}

type ThemeResponse struct {
	Name          string                `json:"theme_name"`
	Opportunities []OpportunityResponse `json:"opportunities"`
}
