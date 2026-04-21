package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"
	"encoding/xml"

	"github.com/jackc/pgx/v5"

	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

const feedMaxItems = 50

// feedTmpl emits RSS 2.0 with atom:link rel="self" and lastBuildDate.
// All user-supplied strings pass through xmlescape to prevent injection.
var feedTmpl = template.Must(template.New("rss").Funcs(template.FuncMap{
	"xmlescape": func(s string) string {
		var b strings.Builder
		_ = xml.EscapeText(&b, []byte(s))
		return b.String()
	},
}).Parse(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title>{{xmlescape .Title}}</title>
    <link>{{xmlescape .Link}}</link>
    <description>{{xmlescape .Description}}</description>
    <language>{{.Language}}</language>
    <lastBuildDate>{{.LastBuildDate}}</lastBuildDate>
    <ttl>{{.TTL}}</ttl>
    <atom:link href="{{xmlescape .SelfURL}}" rel="self" type="application/rss+xml"/>
    {{- range .Items}}
    <item>
      <title>{{xmlescape .Title}}</title>
      <link>{{xmlescape .Link}}</link>
      <description>{{xmlescape .Description}}</description>
      <pubDate>{{.PubDate}}</pubDate>
      <guid isPermaLink="true">{{xmlescape .GUID}}</guid>
    </item>
    {{- end}}
  </channel>
</rss>`))

type feedData struct {
	Title         string
	Link          string
	Description   string
	Language      string
	LastBuildDate string
	TTL           int
	SelfURL       string
	Items         []feedItem
}

type feedItem struct {
	Title       string
	Link        string
	Description string
	PubDate     string
	GUID        string
}

// ----- handlers -----------------------------------------------------------------

func (s *Server) getGlobalFeed(w http.ResponseWriter, r *http.Request) {
	incidents, err := s.store.ListIncidents(r.Context(), pgstore.ListIncidentsParams{
		Limit:  feedMaxItems,
		Offset: 0,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load incidents")
		return
	}

	providers, _ := s.store.ListActiveProviders(r.Context())
	names := providerNameMap(providers)

	base := siteBase(r)
	writeFeed(w, feedData{
		Title:         "llmstatus.io — All Incidents",
		Link:          base,
		Description:   "Real-time incident feed for all AI API providers monitored by llmstatus.io",
		Language:      "en",
		LastBuildDate: time.Now().UTC().Format(time.RFC1123Z),
		TTL:           1,
		SelfURL:       base + "/feed.xml",
		Items:         incidentsToItems(incidents, names, base),
	})
}

func (s *Server) getProviderFeed(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	p, err := s.store.GetProvider(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "provider not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch provider")
		return
	}

	incidents, err := s.store.ListIncidentsByProvider(r.Context(), pgstore.ListIncidentsByProviderParams{
		ProviderID: id,
		Limit:      feedMaxItems,
		Offset:     0,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load incidents")
		return
	}

	base := siteBase(r)
	writeFeed(w, feedData{
		Title:         p.Name + " — Incidents | llmstatus.io",
		Link:          base + "/providers/" + p.ID,
		Description:   "Incident feed for " + p.Name + " — llmstatus.io",
		Language:      "en",
		LastBuildDate: time.Now().UTC().Format(time.RFC1123Z),
		TTL:           1,
		SelfURL:       base + "/v1/providers/" + p.ID + "/feed.xml",
		Items:         incidentsToItems(incidents, map[string]string{p.ID: p.Name}, base),
	})
}

// ----- helpers ------------------------------------------------------------------

func writeFeed(w http.ResponseWriter, data feedData) {
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "max-age=60, s-maxage=60")
	_ = feedTmpl.Execute(w, data)
}

func incidentsToItems(incidents []pgstore.Incident, names map[string]string, base string) []feedItem {
	items := make([]feedItem, 0, len(incidents))
	for _, inc := range incidents {
		name := inc.ProviderID
		if n, ok := names[inc.ProviderID]; ok {
			name = n
		}
		link := base + "/incidents/" + inc.Slug
		desc := fmt.Sprintf("Provider: %s | Severity: %s | Status: %s | Started: %s",
			name, inc.Severity, inc.Status,
			mustTime(inc.StartedAt).Format("Mon, 02 Jan 2006 15:04:05 UTC"),
		)
		if inc.Description.Valid && inc.Description.String != "" {
			desc += "\n\n" + inc.Description.String
		}
		items = append(items, feedItem{
			Title:       fmt.Sprintf("[%s] %s: %s", name, inc.Severity, inc.Title),
			Link:        link,
			Description: desc,
			PubDate:     mustTime(inc.StartedAt).UTC().Format(time.RFC1123Z),
			GUID:        link,
		})
	}
	return items
}

func providerNameMap(providers []pgstore.Provider) map[string]string {
	m := make(map[string]string, len(providers))
	for _, p := range providers {
		m[p.ID] = p.Name
	}
	return m
}

// siteBase constructs the scheme+host base URL, honouring X-Forwarded-Proto
// set by nginx in production.
func siteBase(r *http.Request) string {
	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	return scheme + "://" + r.Host
}
