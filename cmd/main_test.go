//go:build integration
// +build integration

package main_test

import (
	"DebtEase/internal/api"
	"DebtEase/internal/database"
	"DebtEase/internal/middleware"
	"DebtEase/internal/stress"
	webapp_handlers "DebtEase/internal/webapp_specifics/handlers"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func initTestServer(t *testing.T) (http.Handler, *database.Queries, func()) {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		t.Skip("DB_URL not set, skipping integration tests")
		return nil, nil, func() {}
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Skipf("open db: %v (is Postgres running?)", err)
		return nil, nil, func() {}
	}
	if err := db.Ping(); err != nil {
		db.Close()
		t.Skipf("cannot connect to DB: %v (start Postgres and run migrations)", err)
		return nil, nil, func() {}
	}
	queries := database.New(db)
	cfg := &api.Config{DB: queries}
	mux := http.NewServeMux()
	mux.Handle("GET /api/v1/session", middleware.Logging(webapp_handlers.HandleGetSession(cfg)))
	mux.Handle("POST /api/v1/finance-steps", middleware.Logging(webapp_handlers.HandlePostFinanceSteps(cfg)))
	mux.Handle("PUT /api/v1/finance-steps", middleware.Logging(webapp_handlers.HandlePutFinanceSteps(cfg)))
	mux.Handle("PATCH /api/v1/finance-steps", middleware.Logging(webapp_handlers.HandlePatchFinanceSteps(cfg)))
	mux.Handle("GET /api/v1/finance-steps", middleware.Logging(webapp_handlers.HandleGetFinanceSteps(cfg)))
	mux.Handle("DELETE /api/v1/finance-steps", middleware.Logging(webapp_handlers.HandleDeleteFinanceSteps(cfg)))
	mux.Handle("GET /api/v1/debt-health", middleware.Logging(webapp_handlers.HandleGetDebtHealth(cfg)))
	mux.Handle("POST /api/v1/classification", middleware.Logging(stress.HandlePostClassification(cfg)))
	mux.Handle("POST /api/v1/stress/metrics", middleware.Logging(stress.HandlePostStressMetrics(cfg)))
	return mux, queries, func() { db.Close() }
}

func TestGetSession_MissingSession(t *testing.T) {
	t.Parallel()
	h, _, cleanup := initTestServer(t)
	defer cleanup()
	if h == nil {
		return
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/session", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("GET /api/v1/session without session: got status %d, want 400", rec.Code)
	}
}

func TestGetSession_WithValidSession(t *testing.T) {
	t.Parallel()
	h, q, cleanup := initTestServer(t)
	defer cleanup()
	if h == nil {
		return
	}
	ctx := context.Background()
	sessionID := uuid.New()
	_, err := q.CreateCalculationSession(ctx, database.CreateCalculationSessionParams{
		SessionID:     sessionID,
		EngineVersion: "v1",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/session?session_id="+sessionID.String(), nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /api/v1/session: got status %d, want 200", rec.Code)
	}
	var out struct {
		SessionID     string `json:"session_id"`
		EngineVersion string `json:"engine_version"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.SessionID != sessionID.String() || out.EngineVersion != "v1" {
		t.Errorf("unexpected response: %+v", out)
	}
}

func TestPostGetFinanceSteps(t *testing.T) {
	t.Parallel()
	h, q, cleanup := initTestServer(t)
	defer cleanup()
	if h == nil {
		return
	}
	ctx := context.Background()
	sessionID := uuid.New()
	_, err := q.CreateCalculationSession(ctx, database.CreateCalculationSessionParams{
		SessionID:     sessionID,
		EngineVersion: "v1",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	body := map[string]interface{}{
		"session_id":         sessionID.String(),
		"monthly_income":     "100000",
		"fixed_obligations": "10000",
		"loans": []map[string]interface{}{
			{"id": "l1", "principal": "100000", "annual_rate": "12", "tenure_months": 24, "moratorium_months": 0},
		},
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/finance-steps", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("POST /api/v1/finance-steps: got status %d, want 200", rec.Code)
	}
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/finance-steps?session_id="+sessionID.String(), nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("GET /api/v1/finance-steps: got status %d, want 200", rec2.Code)
	}
	var getResp struct {
		SessionID    string `json:"session_id"`
		FinanceSteps struct {
			Loans []interface{} `json:"loans"`
		} `json:"finance_steps"`
	}
	if err := json.NewDecoder(rec2.Body).Decode(&getResp); err != nil {
		t.Fatalf("decode get: %v", err)
	}
	if getResp.SessionID != sessionID.String() || len(getResp.FinanceSteps.Loans) != 1 {
		t.Errorf("unexpected GET response: %+v", getResp)
	}
}

func TestPutFinanceSteps(t *testing.T) {
	t.Parallel()
	h, q, cleanup := initTestServer(t)
	defer cleanup()
	if h == nil {
		return
	}
	ctx := context.Background()
	sessionID := uuid.New()
	_, err := q.CreateCalculationSession(ctx, database.CreateCalculationSessionParams{
		SessionID:     sessionID,
		EngineVersion: "v1",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	body := map[string]interface{}{
		"session_id":         sessionID.String(),
		"monthly_income":     "75000",
		"fixed_obligations":  "5000",
		"loans":              []map[string]interface{}{{"id": "l1", "principal": "200000", "annual_rate": "10", "tenure_months": 36, "moratorium_months": 0}},
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/finance-steps", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("PUT /api/v1/finance-steps: got status %d, want 200", rec.Code)
	}
	// GET should return the replaced data
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/finance-steps?session_id="+sessionID.String(), nil)
	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Errorf("GET after PUT: got status %d", getRec.Code)
	}
}

func TestPatchFinanceSteps(t *testing.T) {
	t.Parallel()
	h, q, cleanup := initTestServer(t)
	defer cleanup()
	if h == nil {
		return
	}
	ctx := context.Background()
	sessionID := uuid.New()
	_, err := q.CreateCalculationSession(ctx, database.CreateCalculationSessionParams{
		SessionID:     sessionID,
		EngineVersion: "v1",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	// POST initial
	postBody := map[string]interface{}{
		"session_id": sessionID.String(), "monthly_income": "100000",
		"loans": []map[string]interface{}{{"id": "l1", "principal": "50000", "annual_rate": "12", "tenure_months": 12, "moratorium_months": 0}},
	}
	pb, _ := json.Marshal(postBody)
	postReq := httptest.NewRequest(http.MethodPost, "/api/v1/finance-steps", bytes.NewReader(pb))
	postReq.Header.Set("Content-Type", "application/json")
	postRec := httptest.NewRecorder()
	h.ServeHTTP(postRec, postReq)
	if postRec.Code != http.StatusOK {
		t.Fatalf("POST setup: got %d", postRec.Code)
	}
	// PATCH only fixed_obligations
	patchBody := map[string]interface{}{"session_id": sessionID.String(), "fixed_obligations": "15000"}
	patchB, _ := json.Marshal(patchBody)
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/finance-steps", bytes.NewReader(patchB))
	patchReq.Header.Set("Content-Type", "application/json")
	patchRec := httptest.NewRecorder()
	h.ServeHTTP(patchRec, patchReq)
	if patchRec.Code != http.StatusOK {
		t.Errorf("PATCH /api/v1/finance-steps: got status %d, want 200", patchRec.Code)
	}
}

func TestDeleteFinanceSteps(t *testing.T) {
	t.Parallel()
	h, q, cleanup := initTestServer(t)
	defer cleanup()
	if h == nil {
		return
	}
	ctx := context.Background()
	sessionID := uuid.New()
	_, err := q.CreateCalculationSession(ctx, database.CreateCalculationSessionParams{
		SessionID:     sessionID,
		EngineVersion: "v1",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	// POST then DELETE
	body := map[string]interface{}{
		"session_id": sessionID.String(), "monthly_income": "50000",
		"loans": []map[string]interface{}{{"id": "l1", "principal": "10000", "annual_rate": "8", "tenure_months": 6, "moratorium_months": 0}},
	}
	b, _ := json.Marshal(body)
	postReq := httptest.NewRequest(http.MethodPost, "/api/v1/finance-steps", bytes.NewReader(b))
	postReq.Header.Set("Content-Type", "application/json")
	postRec := httptest.NewRecorder()
	h.ServeHTTP(postRec, postReq)
	if postRec.Code != http.StatusOK {
		t.Fatalf("POST setup: got %d", postRec.Code)
	}
	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/finance-steps?session_id="+sessionID.String(), nil)
	delRec := httptest.NewRecorder()
	h.ServeHTTP(delRec, delReq)
	if delRec.Code != http.StatusNoContent {
		t.Errorf("DELETE /api/v1/finance-steps: got status %d, want 204", delRec.Code)
	}
	// GET should 404
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/finance-steps?session_id="+sessionID.String(), nil)
	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusNotFound {
		t.Errorf("GET after DELETE: got status %d, want 404", getRec.Code)
	}
}

func TestGetDebtHealth_NoSteps(t *testing.T) {
	t.Parallel()
	h, q, cleanup := initTestServer(t)
	defer cleanup()
	if h == nil {
		return
	}
	ctx := context.Background()
	sessionID := uuid.New()
	_, err := q.CreateCalculationSession(ctx, database.CreateCalculationSessionParams{
		SessionID:     sessionID,
		EngineVersion: "v1",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/debt-health?session_id="+sessionID.String(), nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /api/v1/debt-health: got status %d, want 200", rec.Code)
	}
	var out struct {
		OverallHealthScore int `json:"overall_health_score"`
		DebtCount          int `json:"debt_count"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.DebtCount != 0 || out.OverallHealthScore != 100 {
		t.Errorf("unexpected debt-health: %+v", out)
	}
}

func TestPostClassification(t *testing.T) {
	t.Parallel()
	h, _, cleanup := initTestServer(t)
	defer cleanup()
	if h == nil {
		return
	}
	body := map[string]interface{}{
		"loans": []map[string]interface{}{
			{"id": "l1", "principal": "100000", "annual_rate": "12", "tenure_months": 24, "moratorium_months": 0},
		},
		"monthly_income":     "150000",
		"fixed_obligations": "20000",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/classification", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("POST /api/v1/classification: got status %d, want 200", rec.Code)
	}
	var out stress.ClassificationResponse
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Classification == "" || out.Confidence == "" {
		t.Errorf("missing classification or confidence: %+v", out)
	}
}

func TestPostStressMetrics(t *testing.T) {
	t.Parallel()
	h, _, cleanup := initTestServer(t)
	defer cleanup()
	if h == nil {
		return
	}
	body := map[string]interface{}{
		"loans": []map[string]interface{}{
			{"id": "l1", "principal": "50000", "annual_rate": "10", "tenure_months": 12, "moratorium_months": 0},
		},
		"monthly_income": "80000",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stress/metrics", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("POST /api/v1/stress/metrics: got status %d, want 200", rec.Code)
	}
	var out stress.StressMetricsResponse
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Metrics.FragmentationIndex != 1 {
		t.Errorf("expected fragmentation_index 1, got %f", out.Metrics.FragmentationIndex)
	}
}
