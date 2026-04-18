package api_test

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/llmstatus/llmstatus/internal/api"
	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// xmlFeed mirrors rssFeed for decoding in tests — kept internal to test file.
type xmlFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel xmlChannel `xml:"channel"`
}

type xmlChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language"`
	Items       []xmlItem `xml:"item"`
}

type xmlItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
	GUID    string `xml:"guid"`
}

// ---- global feed ---------------------------------------------------------------

func TestGetGlobalFeed_ContentType(t *testing.T) {
	srv := api.New(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/feed.xml", nil)
	req.Host = "llmstatus.io"
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want 200", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/rss+xml") {
		t.Fatalf("Content-Type=%q want application/rss+xml", ct)
	}
	cc := rr.Header().Get("Cache-Control")
	if !strings.Contains(cc, "max-age=60") {
		t.Fatalf("Cache-Control=%q want max-age=60", cc)
	}
}

func TestGetGlobalFeed_XMLDeclaration(t *testing.T) {
	srv := api.New(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/feed.xml", nil)
	req.Host = "llmstatus.io"
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if !strings.HasPrefix(rr.Body.String(), "<?xml") {
		t.Error("response does not start with XML declaration")
	}
}

func TestGetGlobalFeed_ValidRSS(t *testing.T) {
	store := &fakeStore{
		providers: []pgstore.Provider{fixtureProvider("openai", "OpenAI")},
		incidents: []pgstore.Incident{
			fixtureIncident("openai", "openai-outage-001", "ongoing", "critical"),
			fixtureIncident("openai", "openai-latency-002", "resolved", "minor"),
		},
	}
	srv := api.New(store)
	req := httptest.NewRequest(http.MethodGet, "/feed.xml", nil)
	req.Host = "llmstatus.io"
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	var feed xmlFeed
	if err := xml.NewDecoder(rr.Body).Decode(&feed); err != nil {
		t.Fatalf("XML decode error: %v", err)
	}
	if feed.Version != "2.0" {
		t.Errorf("RSS version=%q want 2.0", feed.Version)
	}
	if feed.Channel.Language != "en" {
		t.Errorf("language=%q want en", feed.Channel.Language)
	}
	if len(feed.Channel.Items) != 2 {
		t.Fatalf("items=%d want 2", len(feed.Channel.Items))
	}
}

func TestGetGlobalFeed_ItemFields(t *testing.T) {
	store := &fakeStore{
		providers: []pgstore.Provider{fixtureProvider("openai", "OpenAI")},
		incidents: []pgstore.Incident{fixtureIncident("openai", "openai-outage-001", "ongoing", "critical")},
	}
	srv := api.New(store)
	req := httptest.NewRequest(http.MethodGet, "/feed.xml", nil)
	req.Host = "example.com"
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	var feed xmlFeed
	if err := xml.NewDecoder(rr.Body).Decode(&feed); err != nil {
		t.Fatalf("XML decode: %v", err)
	}
	item := feed.Channel.Items[0]

	if !strings.Contains(item.Title, "OpenAI") {
		t.Errorf("item title %q missing provider name", item.Title)
	}
	if !strings.Contains(item.Title, "critical") {
		t.Errorf("item title %q missing severity", item.Title)
	}
	if !strings.Contains(item.Link, "example.com") {
		t.Errorf("item link %q should contain host", item.Link)
	}
	if !strings.Contains(item.Link, "openai-outage-001") {
		t.Errorf("item link %q should contain slug", item.Link)
	}
	if item.PubDate == "" {
		t.Error("item pubDate is empty")
	}
	if item.GUID == "" {
		t.Error("item guid is empty")
	}
}

func TestGetGlobalFeed_ProviderIDFallback(t *testing.T) {
	// Incident references a provider not in ListActiveProviders — falls back to ID.
	store := &fakeStore{
		providers: []pgstore.Provider{},
		incidents: []pgstore.Incident{fixtureIncident("deepseek", "ds-001", "ongoing", "major")},
	}
	srv := api.New(store)
	req := httptest.NewRequest(http.MethodGet, "/feed.xml", nil)
	req.Host = "llmstatus.io"
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "deepseek") {
		t.Error("feed should fall back to provider ID when name unavailable")
	}
}

func TestGetGlobalFeed_EmptyIncidents(t *testing.T) {
	srv := api.New(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/feed.xml", nil)
	req.Host = "llmstatus.io"
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	var feed xmlFeed
	if err := xml.NewDecoder(rr.Body).Decode(&feed); err != nil {
		t.Fatalf("XML decode: %v", err)
	}
	if len(feed.Channel.Items) != 0 {
		t.Errorf("items=%d want 0 for empty store", len(feed.Channel.Items))
	}
}

// ---- per-provider feed ---------------------------------------------------------

func TestGetProviderFeed_Success(t *testing.T) {
	store := &fakeStore{
		providers: []pgstore.Provider{fixtureProvider("anthropic", "Anthropic")},
		incidents: []pgstore.Incident{fixtureIncident("anthropic", "ant-001", "resolved", "minor")},
	}
	srv := api.New(store)
	req := httptest.NewRequest(http.MethodGet, "/v1/providers/anthropic/feed.xml", nil)
	req.Host = "llmstatus.io"
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want 200", rr.Code)
	}
	var feed xmlFeed
	if err := xml.NewDecoder(rr.Body).Decode(&feed); err != nil {
		t.Fatalf("XML decode: %v", err)
	}
	if !strings.Contains(feed.Channel.Title, "Anthropic") {
		t.Errorf("channel title %q missing provider name", feed.Channel.Title)
	}
	if len(feed.Channel.Items) != 1 {
		t.Fatalf("items=%d want 1", len(feed.Channel.Items))
	}
}

func TestGetProviderFeed_NotFound(t *testing.T) {
	srv := api.New(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/v1/providers/nonexistent/feed.xml", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status=%d want 404", rr.Code)
	}
}

func TestGetProviderFeed_XForwardedProto(t *testing.T) {
	store := &fakeStore{
		providers: []pgstore.Provider{fixtureProvider("openai", "OpenAI")},
	}
	srv := api.New(store)
	req := httptest.NewRequest(http.MethodGet, "/v1/providers/openai/feed.xml", nil)
	req.Host = "llmstatus.io"
	req.Header.Set("X-Forwarded-Proto", "https")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "https://llmstatus.io") {
		t.Errorf("feed links should use https from X-Forwarded-Proto, got: %s", body[:min(len(body), 300)])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
