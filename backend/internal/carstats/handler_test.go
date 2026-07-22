package carstats

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/KovacevicAleksa/rentcar/backend/internal/auth"
)

func newEventsRouter(svc *Service) (*gin.Engine, *auth.TokenService) {
	gin.SetMode(gin.TestMode)
	tokens := auth.NewTokenService("test-secret", time.Hour)
	r := gin.New()
	RegisterRoutes(r, svc, tokens)
	return r, tokens
}

func TestGetCarEventStatsRequiresAuth(t *testing.T) {
	r, _ := newEventsRouter(NewService(&stubRepo{}))

	req := httptest.NewRequest(http.MethodGet, "/carstats/car/CAR001/events", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}

func TestGetCarEventStatsReturnsBuckets(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	svc := NewService(&stubRepo{events: []Event{
		{CarID: "CAR001", Type: EventOverheat, OccurredAt: now.Add(-1 * time.Hour)},
		{CarID: "CAR001", Type: EventCheckEngine, OccurredAt: now.Add(-2 * time.Hour)},
	}})
	svc.now = func() time.Time { return now }
	r, tokens := newEventsRouter(svc)

	token, _ := tokens.Generate("u-1", "a@rentcar.com", auth.RoleUser)
	req := httptest.NewRequest(http.MethodGet, "/carstats/car/CAR001/events", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
	var resp EventStatsResult
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp.OverheatTotal != 1 || resp.CheckEngineTotal != 1 || len(resp.Buckets) != 1 {
		t.Errorf("unexpected stats: %+v", resp)
	}
}
