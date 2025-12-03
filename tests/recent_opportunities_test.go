package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestRecentOpportunities(t *testing.T) {
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

	req := httptest.NewRequest("GET", "/api/opportunities/recent?include_evidence=true", nil)
	req.Header.Set("X-API-Key", "noker-dev-key-2025")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var opps []api.OpportunityResponse
	json.Unmarshal(w.Body.Bytes(), &opps)
	for _, o := range opps {
		year := time.Now().Year()                           // or the expected year
		createdStr := fmt.Sprintf("%s %d", o.Created, year) // "Dec 03 05:57 2025"
		createdAt, err := time.Parse("Jan 02 15:04 2006", createdStr)
		require.NoError(t, err)
		assert.WithinDuration(t, time.Now(), createdAt, 24*time.Hour)
	}
}
