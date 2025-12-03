package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/pedy4000/noker/internal/api/response"
	"github.com/pedy4000/noker/internal/models"
	"github.com/pedy4000/noker/internal/queue"
	"github.com/pedy4000/noker/internal/repository"
	"github.com/pedy4000/noker/pkg/logger"
	"github.com/pedy4000/noker/pkg/utils"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type Handler struct {
	q      *repository.Queries
	worker queue.Processor
}

func NewHandler(q *repository.Queries, worker queue.Processor) *Handler {
	return &Handler{q: q, worker: worker}
}

// POST /api/meetings
func (h *Handler) CreateMeeting(w http.ResponseWriter, r *http.Request) {
	input := r.Context().Value("Body").(CreateMeetingRequest)

	if input.Source == "" {
		input.Source = models.SourceManual
	}

	var Notes string
	switch input.Source {
	case models.SourceManual:
		Notes = input.Notes
	default:
		// TODO: we can read from all other sources (zoom, upload, notion, etc.)
		Notes = input.Notes
	}

	metadataJSON, err := json.Marshal(input.Metadata)
	if err != nil {
		response.Error(w, "Invalid metadata", http.StatusBadRequest)
		return
	}

	result, err := h.q.CreateMeeting(r.Context(), repository.CreateMeetingParams{
		Title:    input.Title,
		RawNotes: Notes,
		Source:   string(input.Source),
		Metadata: pqtype.NullRawMessage{
			RawMessage: json.RawMessage(metadataJSON),
			Valid:      len(metadataJSON) > 0 && string(metadataJSON) != "null",
		},
	})
	if err != nil {
		logger.Error("CreateMeeting DB error:", err)
		response.Error(w, "Failed to save meeting", http.StatusInternalServerError)
		return
	}

	h.worker.Enqueue(result.ID)

	resp := CreateMeetingResponse{
		Status:      "queued",
		MeetingID:   result.ID,
		Message:     "Notes received! AI is extracting opportunities in background...",
		ProcessedIn: "a few seconds",
	}

	response.JSON(w, http.StatusAccepted, resp)
}

// GET /api/meetings/{id}/status
func (h *Handler) MeetingStatus(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, "Invalid meeting ID", http.StatusBadRequest)
		return
	}

	meeting, err := h.q.GetMeeting(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.Error(w, "Meeting not found", http.StatusNotFound)
		} else {
			response.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	resp := MeetingResponse{
		ID:        meeting.ID,
		Title:     meeting.Title,
		Source:    meeting.Source,
		Metadata:  string(meeting.Metadata.RawMessage),
		Status:    meeting.ProcessingStatus,
		Message:   meeting.ProcessingError.String,
		Created:   utils.FormatTime(meeting.CreatedAt, "never"),
		Processed: utils.FormatTime(meeting.ProcessedAt, "never"),
		Updated:   utils.FormatTime(meeting.UpdatedAt, "never"),
	}

	response.JSON(w, http.StatusOK, resp)
}

// GET /api/opportunities/{id}
func (h *Handler) GetOpportunity(w http.ResponseWriter, r *http.Request, idStr string) {
	includeEvidence := r.URL.Query().Get("include_evidence") == "true"

	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, "Invalid opportunity ID", http.StatusBadRequest)
		return
	}

	opp, err := h.q.GetOpportunity(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.Error(w, "Opportunity not found", http.StatusNotFound)
			return
		}
		response.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	var evidences []EvidenceResponse
	if includeEvidence {
		evidences, err = h.getEvidences(r, opp.ID)
		if err != nil {
			logger.Error("ListEvidenceByOpportunity", err)
			response.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}

	resp := OpportunityResponse{
		ID:            opp.ID,
		UserSegment:   opp.UserSegment,
		Struggle:      opp.Struggle,
		Why:           opp.WhyItMatters.String,
		Workaround:    opp.Workaround.String,
		Theme:         opp.ThemeName.String,
		EvidenceCount: int(opp.EvidenceCount),
		Evidences:     evidences,
		Created:       utils.FormatTime(opp.CreatedAt, "never"),
	}

	response.JSON(w, http.StatusOK, resp)
}

// GET /api/opportunities/recent
func (h *Handler) RecentOpportunities(w http.ResponseWriter, r *http.Request) {
	includeEvidence := r.URL.Query().Get("include_evidence") == "true"

	ops, err := h.q.ListRecentOpportunities(r.Context())
	if err != nil {
		logger.Error("ListRecentOpportunities:", err)
		response.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	result := make([]OpportunityResponse, len(ops))
	for i, op := range ops {
		var evidences []EvidenceResponse
		if includeEvidence {
			evidences, err = h.getEvidences(r, op.ID)
			if err != nil {
				logger.Error("ListEvidenceByOpportunity", err)
				response.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
		}

		result[i] = OpportunityResponse{
			ID:            op.ID,
			UserSegment:   op.UserSegment,
			Struggle:      op.Struggle,
			Why:           op.WhyItMatters.String,
			Workaround:    op.Workaround.String,
			Theme:         op.ThemeName.String,
			EvidenceCount: int(op.EvidenceCount),
			Evidences:     evidences,
			Created:       utils.FormatTime(op.CreatedAt, "never"),
		}
	}

	response.JSON(w, http.StatusOK, result)
}

// GET /api/themes/{theme}/top-opportunities?limit=10
func (h *Handler) TopOpportunitiesByTheme(w http.ResponseWriter, r *http.Request, theme string) {
	includeEvidence := r.URL.Query().Get("include_evidence") == "true"

	if theme == "" {
		response.Error(w, "Theme name is required", http.StatusBadRequest)
		return
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}

	opps, err := h.q.ListTopOpportunitiesByTheme(r.Context(), repository.ListTopOpportunitiesByThemeParams{
		Lower: theme,
		Limit: int32(limit),
	})

	if err != nil {
		response.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if len(opps) == 0 {
		response.Error(w, "Theme not found", http.StatusNotFound)
		return
	}

	ops := make([]OpportunityResponse, len(opps))
	for i, op := range opps {
		var evidences []EvidenceResponse
		if includeEvidence {
			evidences, err = h.getEvidences(r, op.ID)
			if err != nil {
				logger.Error("ListEvidenceByOpportunity", err)
				response.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
		}

		ops[i] = OpportunityResponse{
			ID:            op.ID,
			UserSegment:   op.UserSegment,
			Struggle:      op.Struggle,
			Why:           op.WhyItMatters.String,
			Workaround:    op.WhyItMatters.String,
			EvidenceCount: int(op.EvidenceCount),
			Evidences:     evidences,
			Created:       utils.FormatTime(op.CreatedAt, "never"),
		}
	}
	result := ThemeResponse{
		Name:          theme,
		Opportunities: ops,
	}

	response.JSON(w, http.StatusOK, result)
}

// GET /api/cmd
func (h *Handler) SlackCommand(w http.ResponseWriter, r *http.Request) {
	cmd := strings.TrimSpace(r.URL.Query().Get("text"))
	if cmd == "" {
		cmd = "help"
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	switch strings.ToLower(cmd) {
	case "show_new_opportunities":
		ops, err := h.q.ListRecentOpportunities(r.Context())
		if err != nil {
			w.Write([]byte("Error fetching opportunities"))
			return
		}
		if len(ops) == 0 {
			w.Write([]byte("No new opportunities in the last 24 hours. Keep discovering!"))
			return
		}
		fmt.Fprintf(w, "Found %d new opportunities in the last 24h:\n\n", len(ops))
		for _, op := range ops {
			theme := ""
			if op.ThemeName.Valid && op.ThemeName.String != "" {
				theme = fmt.Sprintf(" [%s]", op.ThemeName.String)
			}
			fmt.Fprintf(w, "• *%s*\n  %s%s\n\n", op.UserSegment, op.Struggle, theme)
		}
	case "themes_this_week":
		themes, err := h.q.ListTopThemesThisWeek(r.Context())
		if err != nil {
			w.Write([]byte("Error fetching themes"))
			return
		}

		if len(themes) == 0 {
			w.Write([]byte("No opportunities this week yet. Keep discovering!"))
			return
		}

		w.Write([]byte("*Top opportunity themes this week:*\n\n"))

		for i, t := range themes {
			rank := i + 1
			fmt.Fprintf(w, "%d. *%s* — %d opportunities\n", rank, t.ThemeName, t.OpportunityCount)
		}

		w.Write([]byte("\nUse `/cmd?text=show_new_opportunities` to see the latest pains"))
	case "help":
		fallthrough
	default:
		w.Write([]byte(`Noker Commands

• show_new_opportunities → Latest customer pains
• themes_this_week       → Trending opportunity themes
• help                   → This message

Example: /cmd?text=show_new_opportunities`))
	}
}

func (h *Handler) ServeGraph(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/index.html")
}

func (h *Handler) getEvidences(r *http.Request, id uuid.UUID) ([]EvidenceResponse, error) {
	evidences, err := h.q.ListEvidenceByOpportunity(r.Context(), id)
	if err != nil {
		return nil, err
	}

	result := make([]EvidenceResponse, len(evidences))
	for i, ev := range evidences {
		result[i] = EvidenceResponse{
			ID:           ev.ID,
			Quote:        ev.Quote,
			Context:      ev.Context.String,
			Created:      utils.FormatTime(ev.CreatedAt, "never"),
			MeetingID:    ev.MeetingID,
			MeetingTitle: ev.MeetingTitle,
			MeetingDate:  utils.FormatTime(ev.MeetingDate, "never"),
		}
	}
	return result, nil
}
