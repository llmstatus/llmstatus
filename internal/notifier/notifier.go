// Package notifier polls for incident changes and delivers alerts via email and webhook.
package notifier

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/llmstatus/llmstatus/internal/email"
	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// compile-time check that *email.Client satisfies email.Sender.
var _ email.Sender = (*email.Client)(nil)

const (
	eventCreated  = "incident.created"
	eventResolved = "incident.resolved"
)

var severityLevel = map[string]int{"minor": 1, "major": 2, "critical": 3}

// Store is the DB subset required by the notifier.
type Store interface {
	ListIncidentsUpdatedSince(ctx context.Context, updatedAt pgtype.Timestamptz) ([]pgstore.Incident, error)
	ListSubscriptionsForProvider(ctx context.Context, providerID string) ([]pgstore.ListSubscriptionsForProviderRow, error)
	IsAlertSent(ctx context.Context, arg pgstore.IsAlertSentParams) (bool, error)
	LogAlert(ctx context.Context, arg pgstore.LogAlertParams) error
}

// Config holds all dependencies for the Notifier.
type Config struct {
	Store    Store
	Email    email.Sender
	SiteURL  string
	Interval time.Duration
}

// Notifier polls for incident changes and sends alerts.
type Notifier struct {
	cfg Config
}

func New(cfg Config) *Notifier {
	if cfg.Interval <= 0 {
		cfg.Interval = 30 * time.Second
	}
	if cfg.SiteURL == "" {
		cfg.SiteURL = "https://llmstatus.io"
	}
	return &Notifier{cfg: cfg}
}

// Run blocks until ctx is cancelled, polling on cfg.Interval.
// It also starts the hourly digest goroutine.
func (n *Notifier) Run(ctx context.Context) {
	go n.runDigest(ctx)

	// Use a small lookback so events at tick boundaries aren't missed.
	const lookback = 10 * time.Second
	lastCheck := time.Now().UTC().Add(-lookback)

	ticker := time.NewTicker(n.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			since := lastCheck.Add(-lookback)
			lastCheck = t.UTC()
			n.poll(ctx, since)
		}
	}
}

func (n *Notifier) poll(ctx context.Context, since time.Time) {
	incidents, err := n.cfg.Store.ListIncidentsUpdatedSince(ctx, pgtype.Timestamptz{Time: since, Valid: true})
	if err != nil {
		slog.Error("notifier: list incidents", "err", err)
		return
	}
	for _, inc := range incidents {
		n.processIncident(ctx, inc)
	}
}

func (n *Notifier) processIncident(ctx context.Context, inc pgstore.Incident) {
	event := incidentEvent(inc.Status)
	if event == "" {
		return
	}

	subs, err := n.cfg.Store.ListSubscriptionsForProvider(ctx, inc.ProviderID)
	if err != nil {
		slog.Error("notifier: list subscriptions", "provider", inc.ProviderID, "err", err)
		return
	}

	for _, sub := range subs {
		if severityLevel[inc.Severity] < severityLevel[sub.MinSeverity] {
			continue
		}
		if sub.EmailAlerts {
			n.deliver(ctx, sub, inc, "email", event)
		}
		if sub.WebhookUrl.Valid && sub.WebhookUrl.String != "" {
			n.deliver(ctx, sub, inc, "webhook", event)
		}
	}
}

func (n *Notifier) deliver(ctx context.Context, sub pgstore.ListSubscriptionsForProviderRow, inc pgstore.Incident, channel, event string) {
	sent, err := n.cfg.Store.IsAlertSent(ctx, pgstore.IsAlertSentParams{
		SubscriptionID: sub.ID,
		IncidentID:     inc.ID,
		Channel:        channel,
		Event:          event,
	})
	if err != nil {
		slog.Error("notifier: check alert_log", "err", err)
		return
	}
	if sent {
		return
	}

	var sendErr error
	switch channel {
	case "email":
		sendErr = n.sendEmail(ctx, sub.UserEmail, inc, event)
	case "webhook":
		sendErr = deliverWebhook(ctx, sub.WebhookUrl.String, buildPayload(inc, event, n.cfg.SiteURL))
	}
	if sendErr != nil {
		slog.Error("notifier: send failed", "channel", channel, "event", event,
			"incident", inc.ID, "err", sendErr)
		return
	}

	if err := n.cfg.Store.LogAlert(ctx, pgstore.LogAlertParams{
		SubscriptionID: sub.ID,
		IncidentID:     inc.ID,
		Channel:        channel,
		Event:          event,
	}); err != nil {
		slog.Error("notifier: log alert", "err", err)
	}
}

// incidentEvent maps incident status to the event name we send.
// Returns "" for statuses we don't notify on.
func incidentEvent(status string) string {
	switch status {
	case "ongoing":
		return eventCreated
	case "resolved":
		return eventResolved
	default:
		return ""
	}
}

type webhookPayload struct {
	Event      string  `json:"event"`
	ProviderID string  `json:"provider_id"`
	IncidentID uuid.UUID `json:"incident_id"`
	Severity   string  `json:"severity"`
	Title      string  `json:"title"`
	Status     string  `json:"status"`
	StartedAt  string  `json:"started_at"`
	ResolvedAt *string `json:"resolved_at,omitempty"`
	URL        string  `json:"url"`
}

func buildPayload(inc pgstore.Incident, event, siteURL string) webhookPayload {
	p := webhookPayload{
		Event:      event,
		ProviderID: inc.ProviderID,
		IncidentID: inc.ID,
		Severity:   inc.Severity,
		Title:      inc.Title,
		Status:     inc.Status,
		StartedAt:  inc.StartedAt.Time.UTC().Format(time.RFC3339),
		URL:        siteURL + "/incidents/" + inc.Slug,
	}
	if inc.ResolvedAt.Valid {
		s := inc.ResolvedAt.Time.UTC().Format(time.RFC3339)
		p.ResolvedAt = &s
	}
	return p
}
