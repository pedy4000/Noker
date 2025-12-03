package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestSlackCommandShowNewOpportunities(t *testing.T) {
	cfg, _ := config.Load("../config.yaml")
	cfg.Database.URL = testDB
	dbConn, _ := db.Connect(testDB, cfg)
	queries := repository.New(dbConn)
	processor := queue.NewProcessor(queries, cfg)
	processor.Start()
	handler := api.NewHandler(queries, processor)
	router := api.NewRouter(handler, cfg)

	dbConn.Exec("TRUNCATE TABLE opportunity_evidence, meetings, opportunities CASCADE")

	notesList := []string{
		"CSV export fails2, dates are wrong.",
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

	require.Eventually(t, func() bool {
		var count int
		dbConn.QueryRow("SELECT COUNT(*) FROM opportunities").Scan(&count)
		return count == 1
	}, 10*time.Second, 200*time.Millisecond, "testing failed")

	req := httptest.NewRequest("GET", "/api/cmd?text=show_new_opportunities", nil)
	req.Header.Set("X-API-Key", "noker-dev-key-2025")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	body := w.Body.String()
	assert.True(t, strings.Contains(body, "Found 1 new opportunities"))
}
