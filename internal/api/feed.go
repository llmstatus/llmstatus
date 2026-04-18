package api

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"

	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

const feedMaxItems = 50

// ----- RSS 2.0 structures -------------------------------------------------------

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language"`
	TTL         int       `xml:"ttl"`
	Items       []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string  `xml:"title"`
	Link        string  `xml:"link"`
	Description string  `xml:"description"`
	PubDate     string  `xml:"pubDate"`
	GUID        rssGUID `xml:"guid"`
}

// rssGUID marks the guid as a permalink when the value is a stable URL.
type rssGUID struct {
	Value       string `xml:",chardata"`
	IsPermaLink string `xml:"isPermaLink,attr"`
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
	writeFeed(w, rssChannel{
		Title:       "llmstatus.io — All Incidents",
		Link:        base,
		Description: "Real-time incident feed for all AI API providers monitored by llmstatus.io",
		Language:    "en",
		TTL:         1,
		Items:       incidentsToItems(incidents, names, base),
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
	writeFeed(w, rssChannel{
		Title:       p.Name + " — Incidents | llmstatus.io",
		Link:        base + "/providers/" + p.ID,
		Description: "Incident feed for " + p.Name + " — llmstatus.io",
		Language:    "en",
		TTL:         1,
		Items:       incidentsToItems(incidents, map[string]string{p.ID: p.Name}, base),
	})
}

// ----- helpers ------------------------------------------------------------------

func writeFeed(w http.ResponseWriter, ch rssChannel) {
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "max-age=60, s-maxage=60")

	_, _ = fmt.Fprint(w, xml.Header)
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	_ = enc.Encode(rssFeed{Version: "2.0", Channel: ch})
}

func incidentsToItems(incidents []pgstore.Incident, names map[string]string, base string) []rssItem {
	items := make([]rssItem, 0, len(incidents))
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
		items = append(items, rssItem{
			Title:       fmt.Sprintf("[%s] %s: %s", name, inc.Severity, inc.Title),
			Link:        link,
			Description: desc,
			PubDate:     mustTime(inc.StartedAt).UTC().Format("Mon, 02 Jan 2006 15:04:05 -0700"),
			GUID:        rssGUID{Value: link, IsPermaLink: "true"},
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
