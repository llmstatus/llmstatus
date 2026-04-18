package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/llmstatus/llmstatus/internal/api"
	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

func TestGetBadge_Operational(t *testing.T) {
	store := &fakeStore{
		providers: []pgstore.Provider{fixtureProvider("openai", "OpenAI")},
	}
	srv := api.New(store)

	req := httptest.NewRequest(http.MethodGet, "/badge/openai.svg", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want 200", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "image/svg+xml") {
		t.Fatalf("Content-Type=%q want image/svg+xml", ct)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "OpenAI") {
		t.Error("badge SVG missing provider name")
	}
	if !strings.Contains(body, "operational") {
		t.Error("badge SVG missing status")
	}
	if !strings.Contains(body, "#4CAF50") {
		t.Error("badge SVG missing green color for operational status")
	}
}

func TestGetBadge_Degraded(t *testing.T) {
	store := &fakeStore{
		providers: []pgstore.Provider{fixtureProvider("anthropic", "Anthropic")},
		incidents: []pgstore.Incident{fixtureIncident("anthropic", "inc-001", "ongoing", "major")},
	}
	srv := api.New(store)

	req := httptest.NewRequest(http.MethodGet, "/badge/anthropic.svg", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want 200", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "degraded") {
		t.Error("badge SVG missing degraded status")
	}
	if !strings.Contains(body, "#FF9800") {
		t.Error("badge SVG missing amber color for degraded status")
	}
}

func TestGetBadge_Down(t *testing.T) {
	store := &fakeStore{
		providers: []pgstore.Provider{fixtureProvider("openai", "OpenAI")},
		incidents: []pgstore.Incident{fixtureIncident("openai", "inc-down", "ongoing", "critical")},
	}
	srv := api.New(store)

	req := httptest.NewRequest(http.MethodGet, "/badge/openai.svg", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "down") {
		t.Error("badge SVG missing down status")
	}
	if !strings.Contains(body, "#F44336") {
		t.Error("badge SVG missing red color for down status")
	}
}

func TestGetBadge_UnknownProvider(t *testing.T) {
	store := &fakeStore{}
	srv := api.New(store)

	req := httptest.NewRequest(http.MethodGet, "/badge/nonexistent.svg", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	// Returns 200 with an "unknown" badge, not a JSON 404.
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want 200 (unknown badge)", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "unknown") {
		t.Error("badge SVG missing 'unknown' for missing provider")
	}
	if !strings.Contains(body, "#9E9E9E") {
		t.Error("badge SVG missing gray color for unknown status")
	}
}

func TestGetBadge_CacheHeaders(t *testing.T) {
	store := &fakeStore{
		providers: []pgstore.Provider{fixtureProvider("openai", "OpenAI")},
	}
	srv := api.New(store)

	req := httptest.NewRequest(http.MethodGet, "/badge/openai.svg", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	cc := rr.Header().Get("Cache-Control")
	if !strings.Contains(cc, "max-age=30") {
		t.Fatalf("Cache-Control=%q want max-age=30", cc)
	}
	if rr.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("X-Content-Type-Options header missing or wrong")
	}
}

func TestGetBadge_SVGStructure(t *testing.T) {
	store := &fakeStore{
		providers: []pgstore.Provider{fixtureProvider("openai", "OpenAI")},
	}
	srv := api.New(store)

	req := httptest.NewRequest(http.MethodGet, "/badge/openai.svg", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	body := rr.Body.String()
	for _, want := range []string{
		`<svg `, `xmlns="http://www.w3.org/2000/svg"`,
		`<title>`, `</title>`,
		`role="img"`, `aria-label=`,
		`</svg>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("SVG missing %q", want)
		}
	}
}

func TestGetBadge_XSSEscaping(t *testing.T) {
	store := &fakeStore{
		providers: []pgstore.Provider{fixtureProvider("xss-test", `<script>alert(1)</script>`)},
	}
	srv := api.New(store)

	req := httptest.NewRequest(http.MethodGet, "/badge/xss-test.svg", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	body := rr.Body.String()
	if strings.Contains(body, "<script>") {
		t.Error("SVG contains unescaped <script> tag — XSS vulnerability")
	}
	if !strings.Contains(body, "&lt;script&gt;") {
		t.Error("expected HTML-escaped script tag in SVG")
	}
}
