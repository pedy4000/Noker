package service

import (
	"context"
	"fmt"

	"github.com/pedy4000/noker/internal/models"
	"github.com/pedy4000/noker/internal/repository"
	"github.com/pedy4000/noker/pkg/utils"

	"github.com/google/uuid"
)

type OpportunityService struct {
	q *repository.Queries
}

func NewOpportunityService(q *repository.Queries) *OpportunityService {
	return &OpportunityService{q: q}
}

func (s *OpportunityService) ProcessExtractedOpportunities(
	ctx context.Context,
	meetingID uuid.UUID,
	extracted []models.ExtractedOpportunity,
) error {
	for _, opp := range extracted {
		if opp.Type == "new" {
			if err := s.createOpportunity(ctx, meetingID, opp); err != nil {
				return err
			}
		} else {
			if err := s.updateOpportunity(ctx, meetingID, opp); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *OpportunityService) createOpportunity(
	ctx context.Context,
	meetingID uuid.UUID,
	ext models.ExtractedOpportunity,
) error {
	var themeID uuid.NullUUID
	ext.Theme = utils.FormatTheme(ext.Theme)

	if ext.Theme != "" {
		theme, err := s.q.GetThemeByName(ctx, ext.Theme)
		if err != nil {
			// Create new theme
			newTheme, err := s.q.CreateTheme(ctx, ext.Theme)
			if err != nil {
				return err
			}
			themeID = utils.ToNullUUID(newTheme.ID)
		} else {
			themeID = utils.ToNullUUID(theme.ID)
		}
	}

	// New opportunity
	opp, err := s.q.CreateOpportunity(ctx, repository.CreateOpportunityParams{
		UserSegment:  ext.UserSegment,
		Struggle:     ext.Struggle,
		WhyItMatters: utils.ToNullString(ext.WhyItMatters),
		Workaround:   utils.ToNullString(ext.Workaround),
		ThemeID:      themeID,
	})
	if err != nil {
		return err
	}

	return s.addEvidence(ctx, opp.ID, meetingID, ext.EvidenceQuotes)
}

func (s *OpportunityService) updateOpportunity(
	ctx context.Context,
	meetingID uuid.UUID,
	ext models.ExtractedOpportunity,
) error {
	oppID, err := uuid.Parse(ext.ExistingOpportunityID)
	if err != nil {
		return fmt.Errorf("invalid existing opportunity ID '%s': %w", ext.ExistingOpportunityID, err)
	}

	return s.addEvidence(ctx, oppID, meetingID, ext.EvidenceQuotes)
}

func (s *OpportunityService) addEvidence(
	ctx context.Context,
	oppID, meetingID uuid.UUID,
	quotes []models.EvidenceQuote,
) error {
	for _, q := range quotes {
		if q.Quote == "" {
			continue
		}
		err := s.q.AddEvidence(ctx, repository.AddEvidenceParams{
			OpportunityID: oppID,
			MeetingID:     meetingID,
			Quote:         q.Quote,
			Context:       utils.ToNullString(q.Context),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
