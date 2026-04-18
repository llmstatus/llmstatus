package ingest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/llmstatus/llmstatus/internal/probes"
	"github.com/llmstatus/llmstatus/internal/store/influx"
)

// ---- test double ------------------------------------------------------------

type mockWriter struct {
	called []probes.ProbeResult
	err    error
}

func (m *mockWriter) WriteProbeResult(_ context.Context, r probes.ProbeResult) error {
	m.called = append(m.called, r)
	return m.err
}
func (m *mockWriter) Close() error { return nil }

var _ influx.Writer = (*mockWriter)(nil)

// ---- helpers ----------------------------------------------------------------

func validResult() probes.ProbeResult {
	return probes.ProbeResult{
		ProviderID: "openai",
		Model:      "gpt-4o-mini",
		ProbeType:  "light_inference",
		RegionID:   "us-west-2",
		StartedAt:  time.Now().UTC(),
		Success:    true,
		DurationMs: 312,
		HTTPStatus: 200,
		TokensIn:   8,
		TokensOut:  1,
	}
}

func postJSON(t *testing.T, h http.Handler, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/probe", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// ---- tests ------------------------------------------------------------------

func TestHandler_ValidProbe(t *testing.T) {
	w := &mockWriter{}
	h := NewHandler(w)
	r := validResult()

	rec := postJSON(t, h, r)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status: got %d, want 204", rec.Code)
	}
	if len(w.called) != 1 {
		t.Fatalf("writer called %d times, want 1", len(w.called))
	}
	got := w.called[0]
	if got.ProviderID != r.ProviderID {
		t.Errorf("ProviderID: got %q, want %q", got.ProviderID, r.ProviderID)
	}
	if got.DurationMs != r.DurationMs {
		t.Errorf("DurationMs: got %d, want %d", got.DurationMs, r.DurationMs)
	}
}

func TestHandler_MissingFields(t *testing.T) {
	cases := []struct {
		name  string
		mutate func(*probes.ProbeResult)
	}{
		{"missing provider_id", func(r *probes.ProbeResult) { r.ProviderID = "" }},
		{"missing model", func(r *probes.ProbeResult) { r.Model = "" }},
		{"missing probe_type", func(r *probes.ProbeResult) { r.ProbeType = "" }},
		{"missing region_id", func(r *probes.ProbeResult) { r.RegionID = "" }},
		{"zero started_at", func(r *probes.ProbeResult) { r.StartedAt = time.Time{} }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := &mockWriter{}
			h := NewHandler(w)
			r := validResult()
			tc.mutate(&r)

			rec := postJSON(t, h, r)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status: got %d, want 400", rec.Code)
			}
			if len(w.called) != 0 {
				t.Error("writer should not be called on validation failure")
			}
		})
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	h := NewHandler(&mockWriter{})
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		req := httptest.NewRequest(method, "/v1/probe", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: got %d, want 405", method, rec.Code)
		}
	}
}

func TestHandler_MalformedJSON(t *testing.T) {
	h := NewHandler(&mockWriter{})
	req := httptest.NewRequest(http.MethodPost, "/v1/probe", bytes.NewBufferString("{not json}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rec.Code)
	}
}

func TestHandler_WriterError(t *testing.T) {
	w := &mockWriter{err: &influxWriteError{}}
	h := NewHandler(w)

	rec := postJSON(t, h, validResult())

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", rec.Code)
	}
}

type influxWriteError struct{}

func (e *influxWriteError) Error() string { return "influx: write status 503: service unavailable" }
