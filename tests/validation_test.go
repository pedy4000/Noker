package tests

import (
	"net/http"
	"testing"

	"github.com/pedy4000/noker/internal/api"
	"github.com/pedy4000/noker/internal/queue"
	"github.com/pedy4000/noker/internal/repository"
	"github.com/pedy4000/noker/pkg/config"
	"github.com/pedy4000/noker/pkg/db"

	"github.com/stretchr/testify/assert"
)

func TestCreateMeetingValidationWithoutTitle(t *testing.T) {
	cfg, _ := config.Load("../config.yaml")
	cfg.Database.URL = testDB
	dbConn, _ := db.Connect(testDB, cfg)
	queries := repository.New(dbConn)
	processor := queue.NewProcessor(queries, cfg)
	handler := api.NewHandler(queries, processor)
	router := api.NewRouter(handler, cfg)

	payload := map[string]any{
		"notes": "Some notes but no title",
	}

	resp := postMeeting(t, router, payload)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, resp.Body.String(), "'title' is required")

}

func TestCreateMeetingValidationShortTitle(t *testing.T) {
	cfg, _ := config.Load("../config.yaml")
	cfg.Database.URL = testDB
	dbConn, _ := db.Connect(testDB, cfg)
	queries := repository.New(dbConn)
	processor := queue.NewProcessor(queries, cfg)
	handler := api.NewHandler(queries, processor)
	router := api.NewRouter(handler, cfg)

	payload := map[string]any{
		"title": "ti",
		"notes": "Some notes but no title",
	}

	resp := postMeeting(t, router, payload)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, resp.Body.String(), "'title' must be at least 3 ")
}

func TestCreateMeetingValidationInValidSource(t *testing.T) {
	cfg, _ := config.Load("../config.yaml")
	cfg.Database.URL = testDB
	dbConn, _ := db.Connect(testDB, cfg)
	queries := repository.New(dbConn)
	processor := queue.NewProcessor(queries, cfg)
	handler := api.NewHandler(queries, processor)
	router := api.NewRouter(handler, cfg)

	payload := map[string]any{
		"title":  "title",
		"notes":  "Some notes but no title",
		"source": "internet",
	}

	resp := postMeeting(t, router, payload)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, resp.Body.String(), "'source' must be one of: manual, notion, file, upload")
}
