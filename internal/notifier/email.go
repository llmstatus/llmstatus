package notifier

import (
	"context"
	"fmt"
	"time"

	"github.com/llmstatus/llmstatus/internal/email"
	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

func (n *Notifier) sendEmail(ctx context.Context, to string, inc pgstore.Incident, event string) error {
	subj, text, html := buildEmail(inc, event, n.cfg.SiteURL)
	return n.cfg.Email.Send(ctx, email.Message{
		To:      to,
		Subject: subj,
		Text:    text,
		HTML:    html,
	})
}

func buildEmail(inc pgstore.Incident, event, siteURL string) (subject, text, html string) {
	incURL := siteURL + "/incidents/" + inc.Slug
	unsubURL := siteURL + "/account/subscriptions"
	ts := inc.StartedAt.Time.UTC().Format(time.RFC3339)

	if event == eventResolved {
		subject = fmt.Sprintf("[llmstatus] Resolved: %s %s incident — %s", inc.ProviderID, inc.Severity, inc.Title)
		text = fmt.Sprintf("The following incident has been resolved:\n\n%s\nSeverity: %s\nStarted: %s\n\nDetails: %s\n\nManage alerts: %s",
			inc.Title, inc.Severity, ts, incURL, unsubURL)
		html = incidentEmailHTML(inc, event, incURL, unsubURL, ts)
		return
	}

	subject = fmt.Sprintf("[llmstatus] %s %s incident — %s", inc.ProviderID, inc.Severity, inc.Title)
	text = fmt.Sprintf("A new incident has been detected:\n\n%s\nProvider: %s\nSeverity: %s\nStarted: %s\n\nDetails: %s\n\nManage alerts: %s",
		inc.Title, inc.ProviderID, inc.Severity, ts, incURL, unsubURL)
	html = incidentEmailHTML(inc, event, incURL, unsubURL, ts)
	return
}

func incidentEmailHTML(inc pgstore.Incident, event, incURL, unsubURL, ts string) string {
	badge := `<span style="color:#f5a623">` + inc.Severity + `</span>`
	title := inc.Title
	statusLine := "Status: <strong>ongoing</strong>"
	if event == eventResolved {
		badge = `<span style="color:#4caf50">resolved</span>`
		statusLine = "Status: <strong>resolved</strong>"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html><body style="font-family:monospace;background:#0d0d0d;color:#e0e0e0;padding:40px;max-width:600px">
<h2 style="color:#e0e0e0;margin-bottom:4px">[llmstatus] incident alert</h2>
<p style="margin:0 0 20px;color:#888;font-size:12px">%s · %s</p>

<div style="border:1px solid #333;border-radius:6px;padding:20px;margin-bottom:20px">
  <p style="margin:0 0 8px;font-size:18px;color:#e0e0e0">%s</p>
  <p style="margin:0 0 4px;color:#888;font-size:13px">Provider: <strong style="color:#e0e0e0">%s</strong></p>
  <p style="margin:0 0 4px;color:#888;font-size:13px">Severity: %s</p>
  <p style="margin:0 0 4px;color:#888;font-size:13px">%s</p>
  <p style="margin:0;color:#888;font-size:13px">Started: <strong style="color:#e0e0e0">%s</strong></p>
</div>

<p><a href="%s" style="color:#f5a623;text-decoration:none">View incident details →</a></p>

<p style="margin-top:40px;font-size:11px;color:#555">
  You received this because you subscribed to %s alerts on llmstatus.io.<br>
  <a href="%s" style="color:#555">Manage subscriptions</a>
</p>
</body></html>`,
		ts, inc.ProviderID,
		title,
		inc.ProviderID, badge, statusLine, ts,
		incURL,
		inc.ProviderID, unsubURL)
}
