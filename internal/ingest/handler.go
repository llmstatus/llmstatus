// Package ingest implements the HTTP handler for the probe result ingestion
// endpoint. It validates incoming ProbeResult JSON and writes each record to
// InfluxDB via the influx.Writer interface.
package ingest

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/llmstatus/llmstatus/internal/probes"
	"github.com/llmstatus/llmstatus/internal/store/influx"
)

// Handler handles POST /v1/probe requests.
type Handler struct {
	writer influx.Writer
}

// NewHandler returns a Handler backed by the supplied Writer.
func NewHandler(w influx.Writer) *Handler {
	return &Handler{writer: w}
}

// ServeHTTP accepts a single ProbeResult as a JSON body, validates it, and
// writes it to InfluxDB. Returns 204 on success, 400 on validation failure,
// 405 for non-POST methods, or 500 on write errors.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var result probes.ProbeResult
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	if err := validate(result); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.writer.WriteProbeResult(r.Context(), result); err != nil {
		slog.Error("ingest: write failed",
			"provider", result.ProviderID,
			"model", result.Model,
			"err", err,
		)
		writeError(w, http.StatusInternalServerError, "write failed")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func validate(r probes.ProbeResult) error {
	switch {
	case r.ProviderID == "":
		return errMissing("provider_id")
	case r.Model == "":
		return errMissing("model")
	case r.ProbeType == "":
		return errMissing("probe_type")
	case r.RegionID == "":
		return errMissing("region_id")
	case r.StartedAt.IsZero():
		return errMissing("started_at")
	}
	return nil
}

type validationError struct{ field string }

func (e *validationError) Error() string { return "missing required field: " + e.field }

func errMissing(field string) error { return &validationError{field} }

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
