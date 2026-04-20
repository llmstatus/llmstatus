package notifier

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/llmstatus/llmstatus/internal/email"
	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// DigestStore is the DB subset needed by the digest goroutine.
type DigestStore interface {
	ListUsersForDigest(ctx context.Context) ([]pgstore.User, error)
	ListDigestSubscriptions(ctx context.Context, userID int64) ([]pgstore.ListDigestSubscriptionsRow, error)
	ListRecentIncidentsByProvider(ctx context.Context, providerID string) ([]pgstore.Incident, error)
	IsDigestSent(ctx context.Context, arg pgstore.IsDigestSentParams) (bool, error)
	LogDigest(ctx context.Context, arg pgstore.LogDigestParams) error
}

// runDigest starts the hourly digest goroutine; it returns when ctx is cancelled.
func (n *Notifier) runDigest(ctx context.Context) {
	// Fire on every clock hour rather than every 60 minutes to stay aligned.
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			n.sendDigests(ctx, t.UTC())
		}
	}
}

func (n *Notifier) sendDigests(ctx context.Context, now time.Time) {
	ds, ok := n.cfg.Store.(DigestStore)
	if !ok {
		return // digest store not available
	}
	users, err := ds.ListUsersForDigest(ctx)
	if err != nil {
		slog.Error("digest: list users", "err", err)
		return
	}
	for _, u := range users {
		n.maybeDigest(ctx, ds, u, now)
	}
}

func (n *Notifier) maybeDigest(ctx context.Context, ds DigestStore, u pgstore.User, now time.Time) {
	loc, err := time.LoadLocation(u.Timezone)
	if err != nil {
		loc = time.UTC
	}
	local := now.In(loc)
	if int(local.Hour()) != int(u.DigestHour) {
		return
	}

	localDate := pgtype.Date{Time: local.Truncate(24 * time.Hour), Valid: true}
	sent, err := ds.IsDigestSent(ctx, pgstore.IsDigestSentParams{UserID: u.ID, SentDate: localDate})
	if err != nil {
		slog.Error("digest: check sent", "user", u.ID, "err", err)
		return
	}
	if sent {
		return
	}

	subs, err := ds.ListDigestSubscriptions(ctx, u.ID)
	if err != nil || len(subs) == 0 {
		return
	}

	if err := n.sendDigestEmail(ctx, ds, u, subs, local); err != nil {
		slog.Error("digest: send email", "user", u.ID, "err", err)
		return
	}

	if err := ds.LogDigest(ctx, pgstore.LogDigestParams{UserID: u.ID, SentDate: localDate}); err != nil {
		slog.Error("digest: log", "user", u.ID, "err", err)
	}
}

func (n *Notifier) sendDigestEmail(ctx context.Context, ds DigestStore, u pgstore.User, subs []pgstore.ListDigestSubscriptionsRow, localNow time.Time) error {
	subject := fmt.Sprintf("[llmstatus] Daily digest — %s", localNow.Format("2006-01-02"))
	text, html := buildDigestEmail(ctx, ds, subs, localNow, n.cfg.SiteURL)
	return n.cfg.Email.Send(ctx, email.Message{
		To:      u.Email,
		Subject: subject,
		Text:    text,
		HTML:    html,
	})
}

func buildDigestEmail(ctx context.Context, ds DigestStore, subs []pgstore.ListDigestSubscriptionsRow, localNow time.Time, siteURL string) (text, html string) {
	date := localNow.Format("2006-01-02")
	unsubURL := siteURL + "/account/subscriptions"

	var sb strings.Builder
	var hb strings.Builder

	sb.WriteString(fmt.Sprintf("llmstatus.io daily digest — %s\n\n", date))
	hb.WriteString(fmt.Sprintf(`<!DOCTYPE html><html><body style="font-family:monospace;background:#0d0d0d;color:#e0e0e0;padding:40px;max-width:640px">
<h2 style="color:#e0e0e0">[llmstatus] daily digest</h2>
<p style="color:#888;font-size:12px;margin-bottom:24px">%s</p>`, date))

	for _, sub := range subs {
		incidents, _ := ds.ListRecentIncidentsByProvider(ctx, sub.ProviderID)
		sb.WriteString(fmt.Sprintf("## %s\n", sub.ProviderName))
		hb.WriteString(fmt.Sprintf(`<div style="border:1px solid #333;border-radius:6px;padding:16px;margin-bottom:16px">
<h3 style="margin:0 0 8px;color:#e0e0e0">%s</h3>`, sub.ProviderName))

		if len(incidents) == 0 {
			sb.WriteString("  No incidents in the last 24h.\n\n")
			hb.WriteString(`<p style="margin:0;color:#4caf50;font-size:13px">✓ No incidents in the last 24h</p>`)
		} else {
			for _, inc := range incidents {
				severity := inc.Severity
				status := inc.Status
				sb.WriteString(fmt.Sprintf("  [%s] %s (%s)\n", severity, inc.Title, status))
				color := "#f5a623"
				if status == "resolved" {
					color = "#4caf50"
				}
				hb.WriteString(fmt.Sprintf(
					`<p style="margin:4px 0;font-size:13px"><span style="color:%s">[%s]</span> %s <span style="color:#888">(%s)</span></p>`,
					color, severity, inc.Title, status))
			}
			sb.WriteString("\n")
		}

		hb.WriteString(fmt.Sprintf(`<p style="margin:8px 0 0;font-size:11px"><a href="%s/providers/%s" style="color:#888">View provider →</a></p></div>`,
			siteURL, sub.ProviderID))
	}

	sb.WriteString(fmt.Sprintf("\nManage your subscriptions: %s\n", unsubURL))
	hb.WriteString(fmt.Sprintf(`<p style="margin-top:32px;font-size:11px;color:#555">
<a href="%s" style="color:#555">Manage subscriptions</a>
</p></body></html>`, unsubURL))

	return sb.String(), hb.String()
}
