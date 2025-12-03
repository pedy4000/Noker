package models

import (
	"time"

	"github.com/google/uuid"
)

// Source of the meeting notes
type MeetingSource string

const (
	SourceManual MeetingSource = "manual"
	SourceNotion MeetingSource = "notion"
	SourceFile   MeetingSource = "file"
)

// Meeting represents a raw discovery/demo call
type Meeting struct {
	ID        uuid.UUID      `json:"id"`
	Title     string         `json:"title,omitempty"`
	RawNotes  string         `json:"raw_notes"`
	Source    MeetingSource  `json:"source"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// Opportunity — the core OST (Opportunity Solution Tree) node
type Opportunity struct {
	ID            uuid.UUID  `json:"id"`
	UserSegment   string     `json:"user_segment"`
	Struggle      string     `json:"struggle"` // Job To Be Done
	WhyItMatters  string     `json:"why_it_matters,omitempty"`
	Workaround    string     `json:"workaround,omitempty"`
	ThemeID       *uuid.UUID `json:"theme_id,omitempty"`
	ThemeName     *string    `json:"theme,omitempty"` // joined for API convenience
	EvidenceCount int        `json:"evidence_count"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Evidence — direct quote + traceability
type Evidence struct {
	ID            uuid.UUID `json:"id"`
	OpportunityID uuid.UUID `json:"opportunity_id"`
	MeetingID     uuid.UUID `json:"meeting_id"`
	Quote         string    `json:"quote"`
	Context       string    `json:"context,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// Theme — optional grouping
type Theme struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// For AI extraction — structured output
type ExtractedOpportunity struct {
	Type                  string          `json:"type"`
	ExistingOpportunityID string          `json:"existing_opportunity_id"`
	UserSegment           string          `json:"user_segment"`
	Struggle              string          `json:"struggle"`
	WhyItMatters          string          `json:"why_it_matters,omitempty"`
	Workaround            string          `json:"workaround,omitempty"`
	Theme                 string          `json:"theme,omitempty"`
	EvidenceQuotes        []EvidenceQuote `json:"evidence_quotes"`
}

type EvidenceQuote struct {
	Quote   string `json:"quote"`
	Context string `json:"context,omitempty"`
}
