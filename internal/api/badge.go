package api

import (
	"fmt"
	"html"
	"net/http"
	"strings"

	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

func (s *Server) getBadge(w http.ResponseWriter, r *http.Request) {
	// Strip optional .svg extension — Go mux does not support mid-segment wildcards.
	id := strings.TrimSuffix(r.PathValue("id"), ".svg")

	p, err := s.store.GetProvider(r.Context(), id)
	if err != nil {
		writeBadge(w, "llmstatus", "unknown", "#9E9E9E")
		return
	}

	activeIncs, err := s.store.ListIncidentsByProvider(r.Context(), pgstore.ListIncidentsByProviderParams{
		ProviderID: id,
		Limit:      5,
		Offset:     0,
	})
	if err != nil {
		activeIncs = nil
	}

	incMap := map[string]pgstore.Incident{}
	for _, inc := range activeIncs {
		if inc.Status == "ongoing" {
			existing, seen := incMap[inc.ProviderID]
			if !seen || severityRank(inc.Severity) > severityRank(existing.Severity) {
				incMap[inc.ProviderID] = inc
			}
		}
	}

	status, _ := deriveStatus(p.ID, incMap)
	message := status

	// ?style=detailed appends live uptime when available.
	if r.URL.Query().Get("style") == "detailed" && s.liveStats != nil {
		if stats, err := s.liveStats.AllProviderLiveStats(r.Context()); err == nil {
			for _, st := range stats {
				if st.ProviderID == id && st.Uptime24h >= 0 {
					message = fmt.Sprintf("%s · %.1f%%", status, st.Uptime24h*100)
					break
				}
			}
		}
	}

	writeBadge(w, p.Name, message, statusColor(status))
}

func writeBadge(w http.ResponseWriter, label, message, color string) {
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, max-age=30")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = fmt.Fprint(w, renderBadge(label, message, color))
}

func statusColor(status string) string {
	switch status {
	case "operational":
		return "#4CAF50"
	case "degraded":
		return "#FF9800"
	case "down":
		return "#F44336"
	default:
		return "#9E9E9E"
	}
}

// renderBadge returns a shields.io-style flat SVG badge.
//
// Coordinate space: SVG uses font-size="110" with transform="scale(.1)" so all
// x/y/width values here are in 10× units (1 SVG unit = 10 displayed pixels).
func renderBadge(label, message, color string) string {
	const (
		hPx  = 20  // badge height in display pixels
		padU = 50  // horizontal padding in 10× units (= 5px each side)
	)

	labelU := badgeTextWidth(label) + padU*2
	msgU := badgeTextWidth(message) + padU*2
	totalU := labelU + msgU

	// Total width in display pixels (round up).
	totalPx := (totalU + 9) / 10

	// Text x-centres in 10× units.
	labelCX := labelU/2 + 10
	msgCX := labelU + msgU/2

	esc := html.EscapeString
	l, m := esc(label), esc(message)

	return fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" role="img" aria-label="%s: %s">`+
			`<title>%s: %s</title>`+
			`<g shape-rendering="crispEdges">`+
			`<rect width="%d" height="%d" fill="#555"/>`+
			`<rect x="%d" width="%d" height="%d" fill="%s"/>`+
			`</g>`+
			`<g fill="#fff" text-anchor="middle" font-family="Verdana,DejaVu Sans,Geneva,sans-serif" font-size="110">`+
			`<text x="%d" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="%d" lengthAdjust="spacing">%s</text>`+
			`<text x="%d" y="140" transform="scale(.1)" textLength="%d" lengthAdjust="spacing">%s</text>`+
			`<text x="%d" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="%d" lengthAdjust="spacing">%s</text>`+
			`<text x="%d" y="140" transform="scale(.1)" textLength="%d" lengthAdjust="spacing">%s</text>`+
			`</g>`+
			`</svg>`,
		totalPx, hPx,
		l, m,
		l, m,
		totalPx, hPx,
		(labelU)/10, (msgU)/10, hPx, color,
		labelCX, labelU-padU*2, l,
		labelCX, labelU-padU*2, l,
		msgCX, msgU-padU*2, m,
		msgCX, msgU-padU*2, m,
	)
}

// badgeTextWidth returns the approximate text width in 10× SVG units for
// Verdana 11px. Values are derived from measured Verdana advance widths.
func badgeTextWidth(s string) int {
	w := 0
	for _, c := range s {
		w += verdanaWidth(c)
	}
	return w
}

func verdanaWidth(c rune) int {
	switch c {
	case ' ':
		return 33
	case 'f', 'i', 'j', 'l', 'r', 't':
		return 45
	case 'I':
		return 40
	case 'm', 'w':
		return 90
	case 'M', 'W':
		return 100
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'J', 'K', 'L', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'X', 'Y', 'Z':
		return 80
	default:
		// Covers a-z (except narrow ones above) and 0-9.
		return 65
	}
}
