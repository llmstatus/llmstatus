import type { Metadata } from "next";
import { CopyButton } from "@/components/CopyButton";

export const dynamic = "force-static";

export const metadata: Metadata = {
  title: "API Documentation",
  description:
    "Public REST API for llmstatus.io — access real-time AI provider status, " +
    "uptime history, incident data, SVG badges, and RSS feeds.",
  openGraph: {
    title: "API Documentation — llmstatus.io",
    description:
      "Public REST API for AI provider status, uptime history, incidents, badges, and RSS.",
  },
};

const BASE = process.env.NEXT_PUBLIC_SITE_URL ?? "https://llmstatus.io";
const API_BASE = `${BASE}/v1`;

interface EndpointProps {
  method: string;
  path: string;
  description: string;
  params?: { name: string; description: string }[];
  example: string;
  note?: string;
}

function Endpoint({ method, path, description, params, example, note }: EndpointProps) {
  return (
    <div className="rounded-lg border border-[var(--ink-600)] overflow-hidden mb-6">
      {/* Method + path header */}
      <div className="flex items-center gap-3 px-4 py-3 border-b border-[var(--ink-600)] bg-[var(--canvas-sunken)]">
        <span className="font-mono text-[11px] font-bold uppercase tracking-widest text-[var(--signal-ok)]">
          {method}
        </span>
        <code className="font-mono text-sm text-[var(--ink-100)]">{path}</code>
      </div>

      <div className="px-4 py-4 space-y-4">
        <p className="text-sm text-[var(--ink-300)] leading-relaxed">{description}</p>

        {params && params.length > 0 && (
          <div>
            <p className="text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--ink-400)] mb-2">
              Parameters
            </p>
            <div className="space-y-1">
              {params.map((p) => (
                <div key={p.name} className="flex gap-3 text-sm">
                  <code className="font-mono text-[var(--signal-amber)] shrink-0">{p.name}</code>
                  <span className="text-[var(--ink-400)]">{p.description}</span>
                </div>
              ))}
            </div>
          </div>
        )}

        <div>
          <div className="flex items-center justify-between mb-2">
            <p className="text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--ink-400)]">
              Example
            </p>
            <CopyButton text={example} />
          </div>
          <pre className="rounded bg-[var(--canvas-sunken)] border border-[var(--ink-600)] px-3 py-3 text-[11px] font-mono text-[var(--ink-200)] overflow-x-auto leading-relaxed">
            {example}
          </pre>
        </div>

        {note && (
          <p className="text-xs text-[var(--ink-400)] leading-relaxed border-l-2 border-[var(--ink-600)] pl-3">
            {note}
          </p>
        )}
      </div>
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="mb-12">
      <h2 className="text-[11px] font-semibold uppercase tracking-[0.12em] text-[var(--signal-amber)] mb-4">
        {title}
      </h2>
      {children}
    </section>
  );
}

export default function ApiPage() {
  return (
    <main className="flex-1 mx-auto w-full max-w-4xl px-6 py-10">
      {/* Header */}
      <div className="mb-10">
        <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-[var(--signal-amber)] mb-4">
          API
        </p>
        <h1 className="text-2xl font-semibold text-[var(--ink-100)] mb-3">
          Public REST API
        </h1>
        <p className="text-sm text-[var(--ink-400)] leading-relaxed max-w-xl">
          Free, unauthenticated, read-only. Returns real-time AI provider status,
          uptime history, and incident data.
        </p>
      </div>

      {/* Overview */}
      <Section title="Overview">
        <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-4 py-4 space-y-3 text-sm mb-6">
          <div className="flex gap-4">
            <span className="text-[var(--ink-400)] w-28 shrink-0">Base URL</span>
            <code className="font-mono text-[var(--ink-200)]">{API_BASE}</code>
          </div>
          <div className="flex gap-4">
            <span className="text-[var(--ink-400)] w-28 shrink-0">Auth</span>
            <span className="text-[var(--ink-300)]">None required</span>
          </div>
          <div className="flex gap-4">
            <span className="text-[var(--ink-400)] w-28 shrink-0">Rate limit</span>
            <span className="text-[var(--ink-300)]">60 requests / minute per IP</span>
          </div>
          <div className="flex gap-4">
            <span className="text-[var(--ink-400)] w-28 shrink-0">Cache TTL</span>
            <span className="text-[var(--ink-300)]">30 seconds on all endpoints</span>
          </div>
          <div className="flex gap-4">
            <span className="text-[var(--ink-400)] w-28 shrink-0">Format</span>
            <span className="text-[var(--ink-300)]">JSON (UTF-8)</span>
          </div>
        </div>

        <div>
          <p className="text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--ink-400)] mb-2">
            Response envelope
          </p>
          <pre className="rounded bg-[var(--canvas-sunken)] border border-[var(--ink-600)] px-3 py-3 text-[11px] font-mono text-[var(--ink-200)] overflow-x-auto leading-relaxed">{`{
  "data": { ... },
  "meta": {
    "generated_at": "2026-04-18T10:00:00Z",
    "cache_ttl_s": 30
  }
}`}</pre>
        </div>

        <div className="mt-4">
          <p className="text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--ink-400)] mb-2">
            Rate limit headers
          </p>
          <pre className="rounded bg-[var(--canvas-sunken)] border border-[var(--ink-600)] px-3 py-3 text-[11px] font-mono text-[var(--ink-200)] overflow-x-auto leading-relaxed">{`X-RateLimit-Limit: 60
X-RateLimit-Remaining: 58
X-RateLimit-Reset: 1745136060
# 429 response includes:
Retry-After: 42`}</pre>
        </div>
      </Section>

      {/* Status */}
      <Section title="System status">
        <Endpoint
          method="GET"
          path="/v1/status"
          description="Overall system status derived from active providers and ongoing incidents. Returns the worst-case status across all providers."
          example={`curl ${API_BASE}/status

# Response
{
  "data": {
    "status": "operational",   // "operational" | "degraded" | "down"
    "counts": {
      "operational": 18,
      "degraded": 1,
      "down": 0
    }
  }
}`}
        />
      </Section>

      {/* Providers */}
      <Section title="Providers">
        <Endpoint
          method="GET"
          path="/v1/providers"
          description="List all active providers with current status, 24-hour uptime, and p95 latency."
          example={`curl ${API_BASE}/providers

# Response — array of provider summaries
{
  "data": [
    {
      "id": "openai",
      "name": "OpenAI",
      "category": "official",
      "region": "global",
      "current_status": "operational",
      "uptime_24h": 0.9997,
      "p95_ms": 1240,
      "active_incident_id": null
    }
  ]
}`}
        />

        <Endpoint
          method="GET"
          path="/v1/providers/{id}"
          description="Full detail for a single provider including models, active incidents, and links."
          params={[
            { name: "{id}", description: "Provider identifier, e.g. openai, anthropic, deepseek" },
          ]}
          example={`curl ${API_BASE}/providers/openai

# Response
{
  "data": {
    "id": "openai",
    "name": "OpenAI",
    "category": "official",
    "region": "global",
    "current_status": "operational",
    "status_page_url": "https://status.openai.com",
    "models": [
      { "model_id": "gpt-4o", "display_name": "GPT-4o", "model_type": "chat", "active": true }
    ],
    "active_incidents": []
  }
}`}
        />

        <Endpoint
          method="GET"
          path="/v1/providers/{id}/history"
          description="Time-series uptime and latency data for a provider. Returns one bucket per hour (24h), per day (7d), or per day (30d)."
          params={[
            { name: "{id}", description: "Provider identifier" },
            { name: "window", description: "Time window: 24h | 7d | 30d (default: 30d)" },
          ]}
          example={`curl "${API_BASE}/providers/openai/history?window=7d"

# Response — array of time buckets
{
  "data": [
    {
      "timestamp": "2026-04-11T00:00:00Z",
      "total": 1440,
      "errors": 2,
      "uptime": 0.9986,
      "p95_ms": 1380
    }
  ]
}`}
        />
      </Section>

      {/* Incidents */}
      <Section title="Incidents">
        <Endpoint
          method="GET"
          path="/v1/incidents"
          description="List incidents. Defaults to all statuses, newest first."
          params={[
            { name: "status", description: "Filter by status: ongoing | monitoring | resolved | all (default: all)" },
            { name: "limit", description: "Maximum number of results, 1–200 (default: 20)" },
          ]}
          example={`curl "${API_BASE}/incidents?status=ongoing&limit=5"

# Response
{
  "data": [
    {
      "id": "018e...",
      "slug": "2026-04-15-openai-elevated-errors",
      "provider_id": "openai",
      "severity": "major",
      "title": "OpenAI elevated errors detected",
      "status": "ongoing",
      "started_at": "2026-04-15T14:22:00Z"
    }
  ]
}`}
        />

        <Endpoint
          method="GET"
          path="/v1/incidents/{id}"
          description="Full detail for one incident. The {id} parameter accepts either the UUID or the human-readable slug."
          params={[
            { name: "{id}", description: "Incident UUID or slug (e.g. 2026-04-15-openai-elevated-errors)" },
          ]}
          example={`curl ${API_BASE}/incidents/2026-04-15-openai-elevated-errors`}
        />
      </Section>

      {/* Badges */}
      <Section title="SVG badges">
        <Endpoint
          method="GET"
          path="/badge/{id}.svg"
          description="Shields.io-style SVG status badge. Suitable for embedding in README files and websites."
          params={[
            { name: "{id}", description: "Provider identifier, e.g. openai" },
            { name: "style", description: "Badge style: simple (default) | detailed (adds 24h uptime %)" },
          ]}
          example={`# Simple badge
curl ${BASE}/badge/openai.svg

# Detailed badge with uptime
curl "${BASE}/badge/openai.svg?style=detailed"

# Markdown embed
[![OpenAI status](${BASE}/badge/openai.svg)](${BASE}/providers/openai)`}
          note="Badges are served with Cache-Control: max-age=30. The badge endpoint returns 200 with an 'unknown' gray badge for unrecognised provider IDs rather than a 404."
        />
      </Section>

      {/* Feeds */}
      <Section title="RSS feeds">
        <Endpoint
          method="GET"
          path="/feed.xml"
          description="RSS 2.0 feed of all incidents across all providers, newest first."
          example={`curl ${BASE}/feed.xml`}
        />

        <Endpoint
          method="GET"
          path="/v1/providers/{id}/feed.xml"
          description="RSS 2.0 feed of incidents for a single provider."
          params={[
            { name: "{id}", description: "Provider identifier" },
          ]}
          example={`curl ${API_BASE}/providers/openai/feed.xml`}
        />
      </Section>

      {/* Errors */}
      <Section title="Error responses">
        <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-4 py-4 space-y-4">
          <pre className="text-[11px] font-mono text-[var(--ink-200)] overflow-x-auto leading-relaxed">{`# 404 Not Found
{ "error": "not found" }

# 429 Too Many Requests
{ "error": "rate limit exceeded" }
# + header: Retry-After: <seconds>

# 503 Service Unavailable
{ "error": "database unavailable" }`}</pre>
          <p className="text-sm text-[var(--ink-400)] leading-relaxed">
            All error responses use the same{" "}
            <code className="font-mono text-[var(--ink-300)]">{"{ \"error\": \"...\" }"}</code>{" "}
            format. HTTP status codes follow standard semantics.
          </p>
        </div>
      </Section>
    </main>
  );
}
