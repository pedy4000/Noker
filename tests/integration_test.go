package tests

import (
	"bytes"
	"encoding/json"
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

var (
	testDB = "postgres://noker:noker@localhost:5432/noker_test?sslmode=disable"
)

// TestFullOSTFlow tests the entire real-world path:
// 1. Create meeting → queued
// 2. Worker processes it → calls OST-GPT → creates opportunities + evidence
// 3. Fetch opportunities → see correct deduplication and structure
func TestFullOSTFlow(t *testing.T) {
	// 1. Setup test database + router with real dependencies
	cfg, err := config.Load("../config.yaml")
	require.NoError(t, err)
	cfg.Database.URL = testDB

	testDB, err := db.Connect(testDB, cfg)
	require.NoError(t, err)

	queries := repository.New(testDB)
	processor := queue.NewProcessor(queries, cfg)
	handler := api.NewHandler(queries, processor)
	processor.Start()

	router := api.NewRouter(handler, cfg)

	// Clean slate
	_, err = testDB.Exec("TRUNCATE TABLE opportunity_evidence, meetings, opportunities, themes CASCADE")
	require.NoError(t, err)

	// ——————————————————————————————————————————————————————————
	// Step 1: Send first meeting — should create NEW opportunity
	// ——————————————————————————————————————————————————————————
	payload1 := map[string]any{
		"title":    "Acme Corp – Export Hell",
		"notes":    "Every Monday we spend 3 hours cleaning exported CSVs. Persian text becomes garbage, dates are wrong, columns missing.",
		"source":   "manual",
		"metadata": map[string]any{"customer": "Acme Corp", "plan": "enterprise"},
	}

	resp1 := postMeeting(t, router, payload1)
	assert.Equal(t, http.StatusAccepted, resp1.Code)
	assert.Contains(t, resp1.Body.String(), "queued")

	// ——————————————————————————————————————————————————————————
	// Step 2: Wait for background worker to process (real async!)
	// ——————————————————————————————————————————————————————————
	var resp api.CreateMeetingResponse
	err = json.Unmarshal(resp1.Body.Bytes(), &resp)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		var count int
		testDB.QueryRow("SELECT COUNT(*) FROM opportunities").Scan(&count)
		return count > 0
	}, 10*time.Second, 200*time.Millisecond, "OST-GPT job never created opportunity")

	// ——————————————————————————————————————————————————————————
	// Step 3: Send SECOND meeting with SAME root problem (should MATCH, not create new)
	// ——————————————————————————————————————————————————————————
	payload2 := map[string]any{
		"title":  "Acme Follow-up – Still Fixing CSVs",
		"notes":  "Our finance team wasted another 4 hours this week fixing exported reports. Fonts broken again. We’re losing patience.",
		"source": "manual",
	}

	resp2 := postMeeting(t, router, payload2)
	assert.Equal(t, http.StatusAccepted, resp2.Code)

	// Wait for second job
	require.Eventually(t, func() bool {
		var count int
		testDB.QueryRow("SELECT COUNT(*) FROM opportunity_evidence").Scan(&count)
		return count >= 3 // at least 3 evidence quotes total
	}, 10*time.Second, 200*time.Millisecond)

	// ——————————————————————————————————————————————————————————
	// Step 4: Verify FINAL state — only ONE opportunity, rich evidence
	// ——————————————————————————————————————————————————————————
	req := httptest.NewRequest("GET", "/api/opportunities/recent?include_evidence=true", nil)
	req.Header.Set("X-API-Key", "noker-dev-key-2025")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var opps []api.OpportunityResponse
	err = json.Unmarshal(w.Body.Bytes(), &opps)
	require.NoError(t, err)

	// Must have EXACTLY 1 opportunity (perfect deduplication)
	assert.Len(t, opps, 1)
	opp := opps[0]

	// Struggle must be clean and generalized
	assert.Contains(t, opp.Struggle, "CSV export") // or "export corrupts", etc.
	assert.NotContains(t, opp.Struggle, "Acme")
	assert.NotContains(t, opp.Struggle, "3 hours")
	assert.True(t, len(opp.Struggle) <= 60, "struggle too long")

	// Must have multiple evidence from different meetings
	assert.True(t, len(opp.Evidences) >= 2, "deduplication failed — should have merged evidence")

	// Evidence quotes must be preserved
	quotes := make([]string, len(opp.Evidences))
	for i, ev := range opp.Evidences {
		quotes[i] = ev.Quote
	}

	found := false
	for _, q := range quotes {
		if strings.Contains(q, "Persian text becomes garbage") {
			found = true
			break
		}
	}
	require.True(t, found, "none of the quotes contain expected text")

	found = false
	for _, q := range quotes {
		if strings.Contains(q, "Fonts broken again") {
			found = true
			break
		}
	}
	require.True(t, found, "none of the quotes contain expected text")

	// Theme must be correct and short
	assert.Contains(t, []string{
		"export-issues", "export", "reporting-issues", "csv-export",
	}, opp.Theme)
}

// Helper: POST a meeting and return response recorder
func postMeeting(t *testing.T, router http.Handler, payload map[string]any) *httptest.ResponseRecorder {
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/meetings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "noker-dev-key-2025")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}
