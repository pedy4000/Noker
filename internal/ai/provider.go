package ai

import (
	"context"

	"github.com/pedy4000/noker/internal/models"
	"github.com/pedy4000/noker/internal/repository"
	"github.com/pedy4000/noker/pkg/config"
)

type Provider interface {
	Extract(ctx context.Context, meeting *models.Meeting, existingOpps []repository.ListAllOpportunitiesForDeduplicationRow) ([]models.ExtractedOpportunity, error)
}

func NewExtractor(cfg *config.Config) Provider {
	var p Provider

	switch cfg.AI.Provider {
	case "openai":
		p = NewOpenAIProvider(cfg)
	default:
		p = NewOpenAIProvider(cfg)
	}

	return p
}

// TODO: we can easily add other ai provider like grok in case of need
