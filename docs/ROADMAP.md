# llmstatus.io — Project Plan

**Version**: 0.1 (Initial)
**Last updated**: 2026-04-17
**Owner**: llmstatus.io maintainers
**Primary developer**: Claude Code + human review

---

## 0. How to use this document

This document is the **single source of truth** for building llmstatus.io V1.
It is written to be handed to Claude Code (or any AI coding agent) as project context.

**Reading order for Claude Code**:
1. Read Section 1 (Goals & Principles) first — this is non-negotiable
2. Read Section 2 (Architecture) to understand the big picture
3. Then proceed module by module according to priority in Section 11

**When in doubt, ask the human operator before making architectural decisions.**

---

## 1. Goals & Principles

### 1.1 What this is

llmstatus.io is an **independent, real-time monitoring service** for AI/LLM API
providers. It measures uptime, latency, and reliability of 20+ providers by
making real API calls from multiple geographic locations — not by scraping
official status pages.

### 1.2 What this is NOT

- **Not** a commercial product (at least not V1). No billing, no user tiers.
- **Not** a Helicone-style observability tool for users' own API calls.
- **Not** a scraper of status.openai.com / status.anthropic.com etc.
- **Not** a crypto-style "real-time leaderboard" with vanity metrics.

### 1.3 Non-negotiable principles

1. **Honesty over marketing**. If a service operated by the maintainers is having issues, display it.
2. **Transparency of methodology**. Every metric must be reproducible.
3. **Independence of data sources**. Real API calls, not scraped status pages.
4. **Data permanence**. Historical data is the moat. Never delete.
5. **Quiet by default**. No popups, no email captures blocking content, no ads.

### 1.4 Success criteria for V1

V1 is successful if, 3 months after launch:
- Data for 20+ providers is continuously collected with <1% gap
- A developer can land on the homepage and understand "who's up, who's down"
  in under 5 seconds
- The public API can be called by third-party tools
- At least 3 independent websites have embedded our status badges

---

## 2. System Architecture

### 2.1 High-level diagram

```
┌──────────────────────────────────────────────────────────────┐
│  Monitoring Nodes (7 global regions)                          │
│  - AWS us-west-2, us-east-1                                   │
│  - Hetzner DE                                                  │
│  - AWS ap-northeast-1 (Tokyo)                                 │
│  - AWS ap-southeast-1 (Singapore)                             │
│  - Aliyun cn-shanghai                                         │
│  - Tencent Cloud cn-guangzhou                                 │
└────────────────────┬─────────────────────────────────────────┘
                     │ (push probe results every 30-60s)
                     ▼
┌──────────────────────────────────────────────────────────────┐
│  Ingestion API (Go or Python FastAPI)                         │
│  - Receives probe results                                      │
│  - Validates + writes to TimescaleDB                          │
│  - Triggers event detection                                   │
└────────────────────┬─────────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────────┐
│  TimescaleDB (PostgreSQL 15 + timescaledb extension)          │
│  - probes (hypertable, raw data, 90d retention)               │
│  - probes_hourly (continuous aggregate, permanent)            │
│  - probes_daily (continuous aggregate, permanent)             │
│  - incidents (events detected by logic layer)                 │
│  - providers (config)                                         │
└────────────────────┬─────────────────────────────────────────┘
                     │
         ┌───────────┴────────────┐
         ▼                        ▼
┌────────────────────┐   ┌────────────────────┐
│  Public Read API    │   │  Event Detector     │
│  (FastAPI)          │   │  (cron every 60s)   │
│  - /providers       │   │  - applies rules    │
│  - /incidents       │   │  - creates events   │
│  - /badge/X.svg     │   │  - notifies agents  │
└──────────┬──────────┘   └─────────────────────┘
           │
           ▼
┌──────────────────────────────────────────────────────────────┐
│  Frontend (Next.js 14, static-export where possible)          │
│  - Homepage (SSG + revalidate every 30s)                      │
│  - Provider detail pages (SSG)                                │
│  - Incident pages (SSG)                                       │
│  - About / Methodology (fully static)                         │
└──────────────────────────────────────────────────────────────┘
```

### 2.2 Tech stack (opinionated choices)

| Layer | Choice | Reason |
|---|---|---|
| Monitoring node | **Python 3.12 + httpx + asyncio** | Readable, easy to add providers |
| Ingestion API | **Python FastAPI** | Same language as probes, simple |
| Database | **PostgreSQL 15 + TimescaleDB** | Time-series + relational in one |
| Event detection | **Python cron job** | Simple, no need for Kafka/Flink at V1 |
| Public API | **FastAPI + Redis cache** | 30s cache on all endpoints |
| Frontend | **Next.js 14 (App Router)** | SSG + ISR, good SEO |
| Styling | **Tailwind + CSS variables** | Fast iteration |
| Deployment | **Hetzner CX21 VPS + Docker Compose** | $10/mo, sufficient for V1 |
| Monitoring of us | **Better Stack or UptimeRobot** | Who watches the watchmen |

**Explicitly avoided** (for V1):
- Kubernetes (overkill)
- Kafka / Redpanda (overkill)
- Serverless for probes (we need stable IPs + geographic diversity)
- tRPC / GraphQL (REST is simpler and more cacheable)

### 2.3 Geographic probe distribution

| Region | Provider | Cost/mo | Purpose |
|---|---|---|---|
| us-west-2 | AWS EC2 t3.micro | $8 | Near OpenAI/Anthropic origins |
| us-east-1 | AWS EC2 t3.micro | $8 | Near Google/AWS Bedrock |
| Germany | Hetzner CX11 | $4 | EU coverage, cheap |
| Tokyo | AWS t3.micro | $9 | APAC |
| Singapore | AWS t3.micro | $9 | SEA |
| Shanghai | Aliyun ecs.t6-c1m1 | ~$6 | **China view** |
| Guangzhou | Tencent Cloud S5 | ~$6 | **China view backup** |

**Total probe infra: ~$50/mo**

---

## 3. Database Schema

### 3.1 `providers` table

```sql
CREATE TABLE providers (
    id              TEXT PRIMARY KEY,              -- e.g. 'openai'
    name            TEXT NOT NULL,                 -- e.g. 'OpenAI'
    category        TEXT NOT NULL,                 -- 'official' | 'aggregator' | 'chinese_official'
    base_url        TEXT NOT NULL,
    auth_type       TEXT NOT NULL,                 -- 'bearer' | 'api_key_header' | 'custom'
    status_page_url TEXT,
    documentation_url TEXT,
    region          TEXT,                          -- 'global' | 'us' | 'cn' | 'eu'
    added_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    active          BOOLEAN NOT NULL DEFAULT TRUE,
    config          JSONB NOT NULL DEFAULT '{}'::JSONB  -- provider-specific config
);
```

### 3.2 `models` table

```sql
CREATE TABLE models (
    id              SERIAL PRIMARY KEY,
    provider_id     TEXT NOT NULL REFERENCES providers(id),
    model_id        TEXT NOT NULL,                 -- e.g. 'gpt-4o'
    display_name    TEXT NOT NULL,
    model_type      TEXT NOT NULL,                 -- 'chat' | 'embedding' | 'image'
    active          BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE (provider_id, model_id)
);
```

### 3.3 `probes` table (hypertable)

```sql
CREATE TABLE probes (
    time            TIMESTAMPTZ NOT NULL,
    provider_id     TEXT NOT NULL,
    model_id        TEXT NOT NULL,
    node_region     TEXT NOT NULL,                 -- probe origin
    probe_type      TEXT NOT NULL,                 -- 'ping' | 'light_inference' | 'medium_inference' | 'streaming'
    success         BOOLEAN NOT NULL,
    http_status     INT,
    error_type      TEXT,                          -- 'timeout' | 'rate_limit' | '5xx' | 'auth' | 'network' | 'content_policy' | 'unknown'
    error_message   TEXT,
    latency_ms      INT,                           -- total wall time
    ttft_ms         INT,                           -- time to first token (streaming only)
    output_tokens   INT,
    tokens_per_sec  NUMERIC(8,2)
);

SELECT create_hypertable('probes', 'time');
CREATE INDEX idx_probes_provider_time ON probes (provider_id, time DESC);
CREATE INDEX idx_probes_node_time ON probes (node_region, time DESC);
```

### 3.4 Continuous aggregates

```sql
-- Hourly rollup
CREATE MATERIALIZED VIEW probes_hourly
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', time) AS bucket,
    provider_id,
    model_id,
    node_region,
    probe_type,
    COUNT(*) AS total,
    COUNT(*) FILTER (WHERE success) AS successes,
    percentile_cont(0.50) WITHIN GROUP (ORDER BY latency_ms) AS p50,
    percentile_cont(0.95) WITHIN GROUP (ORDER BY latency_ms) AS p95,
    percentile_cont(0.99) WITHIN GROUP (ORDER BY latency_ms) AS p99,
    AVG(tokens_per_sec) AS avg_tps
FROM probes
GROUP BY bucket, provider_id, model_id, node_region, probe_type;

-- Daily rollup (similar structure, bucket = '1 day')
```

### 3.5 `incidents` table

```sql
CREATE TABLE incidents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug            TEXT UNIQUE NOT NULL,          -- e.g. '2026-04-15-openai-elevated-errors'
    provider_id     TEXT NOT NULL REFERENCES providers(id),
    severity        TEXT NOT NULL,                 -- 'minor' | 'major' | 'critical'
    title           TEXT NOT NULL,
    description     TEXT,                          -- agent-generated, human-editable
    status          TEXT NOT NULL,                 -- 'ongoing' | 'monitoring' | 'resolved'
    affected_models TEXT[],                        -- array of model_ids
    affected_regions TEXT[],                       -- array of node_regions
    started_at      TIMESTAMPTZ NOT NULL,
    resolved_at     TIMESTAMPTZ,
    detection_method TEXT,                         -- 'auto' | 'manual'
    detection_rule  TEXT,                          -- which rule triggered it
    metrics_snapshot JSONB,                        -- error rate, latency at peak
    human_reviewed  BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## 4. Provider Adapter Specification

Every provider is implemented as a Python class inheriting from `BaseProvider`.

### 4.1 Base interface

```python
class BaseProvider(ABC):
    """Every provider adapter must implement this interface."""

    @property
    @abstractmethod
    def id(self) -> str: ...  # e.g. 'openai'

    @property
    @abstractmethod
    def category(self) -> str: ...  # 'official' | 'aggregator' | 'chinese_official'

    @abstractmethod
    async def probe_ping(self) -> ProbeResult: ...
    """Lightest possible check: is the endpoint reachable + authenticates?"""

    @abstractmethod
    async def probe_light_inference(self, model: str) -> ProbeResult: ...
    """Send minimal prompt 'Reply with OK', expect '<=10 token response."""

    @abstractmethod
    async def probe_medium_inference(self, model: str) -> ProbeResult: ...
    """~200 token input, request 100 token output."""

    @abstractmethod
    async def probe_streaming(self, model: str) -> ProbeResult: ...
    """Measure TTFT and tokens/sec over SSE."""

    @abstractmethod
    def list_probe_models(self) -> list[str]: ...
    """Which models to probe regularly."""
```

### 4.2 `ProbeResult` schema

```python
@dataclass
class ProbeResult:
    success: bool
    http_status: int | None
    error_type: str | None       # enumerated
    error_message: str | None    # first 500 chars
    latency_ms: int | None
    ttft_ms: int | None          # streaming only
    output_tokens: int | None
    tokens_per_sec: float | None
    raw_response: dict | None    # for debugging, not stored in DB
```

### 4.3 Error type taxonomy (strictly enumerated)

```python
class ErrorType(str, Enum):
    TIMEOUT = "timeout"                    # >30s no response
    NETWORK = "network"                    # DNS, TCP, TLS failure
    RATE_LIMIT = "rate_limit"              # 429 or equivalent
    AUTH = "auth"                          # 401, 403
    SERVER_ERROR = "5xx"                   # 500-599
    BAD_REQUEST = "4xx"                    # 400-499 except rate_limit/auth
    CONTENT_POLICY = "content_policy"      # blocked by safety filter
    MODEL_OVERLOADED = "model_overloaded"  # explicit overload signals
    EMPTY_RESPONSE = "empty_response"      # 200 OK but no content
    MALFORMED = "malformed"                # response doesn't parse
    UNKNOWN = "unknown"                    # anything else
```

### 4.4 Required providers for V1 launch (20 total)

**Official / Frontier (10)**:
1. OpenAI (`openai`) — models: gpt-4o, gpt-4o-mini, o1-mini, o3
2. Anthropic (`anthropic`) — models: claude-opus-4-7, claude-sonnet-4-6, claude-haiku-4-5
3. Google Gemini (`google_gemini`) — models: gemini-2.5-pro, gemini-2.0-flash
4. AWS Bedrock (`aws_bedrock`) — models: claude-sonnet via bedrock, llama variants
5. Azure OpenAI (`azure_openai`) — same models as OpenAI, separate endpoint
6. Mistral AI (`mistral`) — models: mistral-large-latest
7. Cohere (`cohere`) — models: command-r-plus
8. Groq (`groq`) — models: llama-3.3-70b
9. Together AI (`together`) — models: deepseek-v3
10. Fireworks AI (`fireworks`) — models: llama-3.3-70b

**Chinese official (6)**:
11. DeepSeek (`deepseek`) — models: deepseek-chat, deepseek-reasoner
12. Moonshot / Kimi (`moonshot`) — models: moonshot-v1-8k, moonshot-v1-32k
13. Zhipu AI / GLM (`zhipu`) — models: glm-4-plus
14. Alibaba Tongyi (`qwen`) — models: qwen-max, qwen-plus
15. Baidu ERNIE (`ernie`) — models: ernie-4.0
16. ByteDance Doubao (`doubao`) — models: doubao-pro

**Aggregators (4)**:
17. OpenRouter (`openrouter`) — probe: gpt-4o, claude-sonnet
18. _(reserved for a self-monitored service; see METHODOLOGY.md §11)_
19. One additional reputable CN aggregator (TBD)
20. LiteLLM self-hosted (if we deploy one) OR skip for V1

**Implementation priority**: implement OpenAI first, get the full pipeline
end-to-end working (probe → DB → API → frontend render), THEN add more.

### 4.5 Adapter implementation checklist

For each new provider, Claude Code should:
- [ ] Create `adapters/{id}.py` inheriting `BaseProvider`
- [ ] Write at least 3 happy-path tests (mock HTTP)
- [ ] Write 3 error-path tests (timeout, 429, 500)
- [ ] Add to `adapters/__init__.py` registry
- [ ] Insert provider + models into DB via migration
- [ ] Manually test from dev machine: run `python -m probes.run --provider={id}`
- [ ] Verify data appears in DB
- [ ] Verify frontend renders the provider
- [ ] **Human review required** before enabling in production

---

## 5. Probe Scheduling

### 5.1 Frequency matrix

| Probe type | Frequency | Rationale |
|---|---|---|
| Ping | every 30s | Cheap, catches hard outages fast |
| Light inference | every 60s | Real end-to-end check |
| Medium inference | every 5 min | Latency measurement quality |
| Streaming | every 5 min | TTFT measurement |

All probe types run from **all 7 nodes independently**.

Volume math: 20 providers × ~2 models/provider × 7 nodes × (120 ping + 60 light + 12 medium + 12 streaming) per hour = ~67,000 probes/hour = ~48M/month.

### 5.2 Cost estimation

Real token spend: medium_inference averages ~300 input + ~100 output tokens.
- 20 providers × 2 models × 7 nodes × 288 (5-min intervals per day) = ~80,640 medium calls/day
- × 30 days = ~2.4M calls/month
- Average cost: ~$0.0005/call blended = **~$1200/month in API costs**

**This is higher than our initial estimate of $400/mo. Adjust:**
- Either reduce medium_inference frequency to every 10 min (halves cost)
- OR reduce number of models per provider (probe only 1 cheapest model)
- OR accept the $1200/month cost as infrastructure investment

Recommend: **1 model per provider for V1**, adds more models in V2.

### 5.3 Rate limiting and ethics

- Respect provider rate limits. Implement exponential backoff.
- Never probe at a frequency that could be interpreted as abuse.
- Use dedicated paid API keys for monitoring (so we're paying customers).
- Publish our methodology publicly so providers know what we do.

---

## 6. Event Detection Rules

The event detector runs every 60 seconds. It looks at the last 10 minutes
of probe data and applies rules.

### 6.1 Rule: "Provider is DOWN"

Trigger: For provider P, in the last 5 minutes, across all nodes and all
probe types, error rate > 50%.

Severity: `critical`
Title template: `{Provider} is experiencing major disruption`

### 6.2 Rule: "Elevated error rate"

Trigger: For provider P, error rate in last 10 min > 5% AND error rate in
baseline (same hour, past 7 days) < 1%.

Severity: `major`
Title template: `{Provider} elevated errors detected`

### 6.3 Rule: "Degraded latency"

Trigger: For provider P, p95 latency in last 10 min > 3x baseline
(same hour, past 7 days), sustained for 5+ minutes.

Severity: `minor`
Title template: `{Provider} elevated latency`

### 6.4 Rule: "Regional outage"

Trigger: A provider works fine from some nodes but error rate > 50% from
one specific node for 5+ minutes.

Severity: `minor`
Title template: `{Provider} unreachable from {region}`

### 6.5 Deduplication

If an event of the same rule + provider is ongoing, don't create a new one.
Update the existing one.

### 6.6 Resolution

An event auto-resolves if the triggering condition is no longer met for
10 consecutive minutes.

### 6.7 Human review queue

Every new event lands in a "pending review" state. A human (or Layer 4 agent)
has 2 hours to approve/reject/edit before it's published publicly.

**Exception**: `critical` events are published immediately. Humans can retroactively edit.

---

## 7. Public API Specification

All endpoints are GET, no auth required for V1, 30s cache via Redis.

Base: `https://llmstatus.io/api/v1`

### 7.1 Endpoints

```
GET /providers
→ Array of all providers with current status summary
  [{
    id, name, category, region,
    current_status: 'operational' | 'degraded' | 'down',
    uptime_24h: 0.9998,
    uptime_7d: 0.9995,
    latency_p95_ms: 680,
    active_incident_id: null | uuid
  }]

GET /providers/{id}
→ Detailed current info for one provider
  Includes per-model breakdown, per-region breakdown

GET /providers/{id}/history?window=24h|7d|30d&metric=uptime|latency|errors
→ Time-series data
  [{ timestamp, value }, ...]

GET /incidents?status=ongoing|resolved|all&limit=20
→ List of recent incidents

GET /incidents/{id}
→ Full detail of one incident

GET /badge/{provider_id}.svg?style=simple|detailed
→ SVG badge that can be embedded

GET /feed.xml
→ RSS feed of all incidents

GET /providers/{id}/feed.xml
→ RSS feed for one provider's incidents
```

### 7.2 Response format

Always JSON (except badges and feeds):

```json
{
  "data": { ... },
  "meta": {
    "generated_at": "2026-04-17T10:23:45Z",
    "cache_ttl_s": 30
  }
}
```

### 7.3 Rate limiting

- 60 requests/minute per IP for V1
- Return standard `X-RateLimit-*` headers
- 429 with `Retry-After` when exceeded

---

## 8. Frontend Structure

### 8.1 Routes

```
/                        Homepage — live status overview
/providers               Full provider list with filters
/providers/{id}          Single provider detail page
/incidents               All incidents archive
/incidents/{slug}        Single incident detail page
/china                   China-focused view (special page)
/compare                 Provider comparison tool
/api                     API documentation
/about                   About + methodology + team
/methodology             Detailed methodology (for credibility)
/badges                  Badge showcase + embed code
```

### 8.2 Key components

- `<StatusPill>`: green/yellow/red dot with label
- `<LatencyBar>`: horizontal bar with p50/p95/p99 markers
- `<UptimeSparkline>`: 30-day bar chart, one bar per day
- `<IncidentCard>`: summary card for an incident
- `<ProbeTimestamp>`: shows "last checked Xs ago" with live update
- `<ProviderTable>`: sortable table of all providers

### 8.3 SEO requirements (critical)

Every provider page MUST have:
- `<title>`: `Is {Provider} API down? — LLM Status`
- `<meta description>`: data-driven, e.g. "OpenAI API is currently operational. 99.98% uptime over the past 24 hours."
- `<meta og:image>`: dynamically generated status card
- JSON-LD `Service` schema
- `<link rel="alternate" type="application/rss+xml">` pointing to provider feed

Every incident page MUST have:
- `<title>`: `{Provider} incident on {date}: {title}`
- JSON-LD `Event` schema
- Permanent URL (never change)

### 8.4 Build strategy

- Use Next.js 14 App Router
- Static-export homepage with 30s revalidate (ISR)
- Provider pages: ISR with 60s revalidate
- Incident pages: static after the incident resolves (essentially immutable)
- About / methodology: fully static

---

## 9. Deployment & Operations

### 9.1 Infrastructure

- **Main app**: 1x Hetzner CX31 (4 vCPU, 8GB) = €13/mo
  - Runs ingestion API, public API, event detector, frontend, nginx
- **Database**: Same machine for V1 (TimescaleDB on same box)
  - Upgrade to separate DB server when DB > 100GB
- **Probe nodes**: 7 small VMs as listed in 2.3
- **CDN**: Cloudflare (free tier)
- **DNS**: Cloudflare

### 9.2 Deployment method

- Dockerfile for each service (ingestion, api, detector, frontend)
- Docker Compose for single-machine orchestration
- GitHub Actions for CI/CD (build → push to registry → SSH deploy)

### 9.3 Observability of ourselves

- Better Stack (or UptimeRobot free tier) monitoring llmstatus.io itself
- Loki + Grafana for logs (optional for V1, can defer)
- Slack / Lark webhook alerts for: DB down, API down, detector hasn't run in 5min

### 9.4 Data backup

- Daily `pg_dump` to Backblaze B2
- Weekly full snapshot
- **Never delete raw probe data** (aggregate after 90 days to save space, but keep aggregates forever)

---

## 10. Security & Ethics

### 10.1 Do not

- Never log or store the raw content returned by provider APIs
  (there could be sensitive data from shared test prompts)
- Never publish individual probe records with raw response bodies
- Never attempt to probe providers we don't have a paid account with

### 10.2 Do

- Publish our source code (at minimum, the probe logic and adapters)
- Publish our methodology in detail
- Respect robots.txt and API terms of service
- If a provider explicitly requests we stop probing them, stop and discuss

### 10.3 Data privacy

No user accounts in V1 = no PII to worry about.
When email subscriptions are added in V2:
- Store only email + subscription preferences
- No tracking beyond standard server logs
- GDPR-compliant unsubscribe in every email

---

## 11. Development Priority (Week-by-week)

### Week 1: Foundation
- [ ] Set up monorepo structure
- [ ] Database schema + migrations
- [ ] `BaseProvider` abstract class + `ProbeResult` types
- [ ] OpenAI adapter (first end-to-end implementation)
- [ ] Ingestion API with minimal endpoint
- [ ] Deploy to 1 node (us-west) for testing

### Week 2: Scale out providers
- [ ] Anthropic, Google Gemini adapters
- [ ] DeepSeek, Moonshot adapters
- [ ] 2 more nodes online (us-east, germany)
- [ ] Basic event detection (rule 6.1, 6.2)
- [ ] Public API v1 endpoints

### Week 3: Frontend MVP
- [ ] Homepage with live status
- [ ] Provider detail page template
- [ ] Status pill + latency bar + uptime sparkline components
- [ ] Deploy frontend

### Week 4: Remaining providers + China view
- [ ] All 20 providers online
- [ ] China nodes deployed (Shanghai, Guangzhou)
- [ ] China view page
- [ ] RSS feeds
- [ ] Badges

### Week 5-6: Polish + content
- [ ] About / Methodology pages (HIGH PRIORITY for credibility)
- [ ] Incident detail pages with permanent URLs
- [ ] First "soft launch" to small audience
- [ ] Daily digest email (V2 feature, optional here)

### Week 7-9: Event detection refinement
- [ ] Rules 6.3, 6.4, 6.5
- [ ] Human review queue UI
- [ ] Webhook subscriptions
- [ ] Status badge variety

### Week 10-12: Launch preparation
- [ ] Aggregator tab (OpenRouter, LiteLLM, etc.)
- [ ] Compare page
- [ ] First monthly report (based on 3 months of data)
- [ ] Hacker News launch

---

## 12. Open Questions (to resolve before implementation)

These require human decision, not agent decision:

1. **Which infrastructure provider's paid APIs do we already have?**
   We need accounts with all 20 providers. Some require business verification.

2. **China nodes: Aliyun or Tencent Cloud for primary?**
   Aliyun has better international connectivity; Tencent better CN domestic.

3. **When a self-monitored service has issues, who approves the incident page going live?**
   Proposed: automatic publication if severity < critical. Critical incidents go to human review (operator) first.

4. **Domain/email setup**: what email is used for `contact@llmstatus.io`?
   Suggest a dedicated mailbox, not a forward to the maintainers' other products.

5. **Entity**: is llmstatus.io its own legal entity, or hosted under an
   existing one operated by the maintainers? This affects the About page
   and the legal footer.

---

## 13. Handoff notes for Claude Code

When a Claude Code instance reads this document to start work:

1. **Do not skip Section 1.3 (non-negotiable principles)**. They affect every decision.
2. **Ask for clarification** on Section 12 open questions before implementing affected parts.
3. **Implement OpenAI adapter first, full end-to-end, before writing any other adapter.**
4. **Write tests alongside code**, not after. Minimum coverage: happy path + 3 error paths.
5. **Commit frequently with descriptive messages** so human reviewer can catch issues early.
6. **When in doubt, prefer boring technology**. This is infrastructure, not research.
7. **If you find yourself needing to make an architectural decision not covered here**,
   document the options and ask the human operator. Do not decide silently.

---

End of ROADMAP.md — Version 0.1
