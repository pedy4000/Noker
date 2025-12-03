package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pedy4000/noker/internal/models"
	"github.com/pedy4000/noker/internal/repository"
	"github.com/pedy4000/noker/pkg/config"
	"github.com/pedy4000/noker/pkg/logger"
)

type OpenAIExtractor struct {
	client *http.Client
	cfg    *config.Config
}

func NewOpenAIProvider(cfg *config.Config) *OpenAIExtractor {
	return &OpenAIExtractor{
		client: &http.Client{Timeout: 90 * time.Second},
		cfg:    cfg,
	}
}

type extractionResponse struct {
	Results []models.ExtractedOpportunity `json:"results"`
}

func (e *OpenAIExtractor) Extract(ctx context.Context, meeting *models.Meeting, opps []repository.ListAllOpportunitiesForDeduplicationRow) ([]models.ExtractedOpportunity, error) {
	userPrompt := fmt.Sprintf(UserPromptTemplate, meeting.Title, meeting.Source, meeting.RawNotes, formatExistingForAI(opps))

	reqBody := map[string]any{
		"model":       e.cfg.AI.Model,
		"temperature": e.cfg.AI.Temperature,
		"messages": []map[string]string{
			{"role": "system", "content": SystemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"response_format": map[string]string{"type": "json_object"},
	}

	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+e.cfg.AI.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		logger.Error("OpenAI error:", resp.StatusCode, string(body))
		return nil, fmt.Errorf("openai error %d", resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	var extracted extractionResponse
	if err := json.Unmarshal([]byte(result.Choices[0].Message.Content), &extracted); err != nil {
		logger.Error("Failed to parse LLM JSON:", err, result.Choices[0].Message.Content)
		return nil, err
	}

	return extracted.Results, nil
}

func formatExistingForAI(opps []repository.ListAllOpportunitiesForDeduplicationRow) string {
	if len(opps) == 0 {
		return "None — this is the first meeting"
	}

	var lines []string
	for _, o := range opps {
		id := o.OpportunityID
		struggle := o.Struggle
		if len(struggle) > 100 {
			struggle = struggle[:97] + "..."
		}
		theme := o.ThemeName

		lines = append(lines, fmt.Sprintf("%s | %s | %s", id, struggle, theme))
	}
	return strings.Join(lines, "\n")
}
