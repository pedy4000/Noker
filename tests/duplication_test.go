package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/pedy4000/noker/internal/api"
	"github.com/pedy4000/noker/internal/queue"
	"github.com/pedy4000/noker/internal/repository"
	"github.com/pedy4000/noker/pkg/config"
	"github.com/pedy4000/noker/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpportunityDeduplication(t *testing.T) {
	cfg, _ := config.Load("../config.yaml")
	cfg.Database.URL = testDB
	dbConn, _ := db.Connect(testDB, cfg)
	queries := repository.New(dbConn)
	processor := queue.NewProcessor(queries, cfg)
	processor.Start()
	handler := api.NewHandler(queries, processor)
	router := api.NewRouter(handler, cfg)

	// Clear tables
	dbConn.Exec("TRUNCATE TABLE opportunity_evidence, meetings, opportunities CASCADE")

	notesList := []string{
		"CSV export fails, dates are wrong.",
		"Exported CSVs are corrupted again.",
		"Our finance team wasted hours fixing CSVs.",
	}

	for _, n := range notesList {
		payload := map[string]any{
			"title":  "Export Issue",
			"notes":  n,
			"source": "manual",
		}
		resp := postMeeting(t, router, payload)
		assert.Equal(t, http.StatusAccepted, resp.Code)
	}

	// Wait for processing
	require.Eventually(t, func() bool {
		var count int
		dbConn.QueryRow("SELECT COUNT(*) FROM meetings WHERE processing_status != 'done'").Scan(&count)
		return count == 0
	}, 10*time.Second, 200*time.Millisecond, "process failed")

	// opportunities == 1
	var count int
	dbConn.QueryRow("SELECT COUNT(*) FROM opportunities").Scan(&count)
	assert.Equal(t, count, 1)

	// Evidence count >= 3
	dbConn.QueryRow("SELECT COUNT(*) FROM opportunity_evidence").Scan(&count)
	assert.Equal(t, count, 3)
}
