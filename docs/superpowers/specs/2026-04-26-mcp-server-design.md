# MCP Server Design — @llmstatus/mcp

**Date**: 2026-04-26  
**Status**: Approved  
**Issue**: TBD (create LLMS- issue before implementation)

---

## 1. Goal

Publish an npm package (`@llmstatus/mcp`) that exposes llmstatus.io data as
MCP tools. Any Claude or Cursor user can add it to their config and then ask
"Is OpenAI down right now?" — the AI calls the tool, gets a plain-text answer,
and responds inline. Zero configuration required.

---

## 2. Architecture

A thin stdio MCP server written in TypeScript. It is a pure HTTP client: it
calls the public `https://api.llmstatus.io` REST API and formats responses as
AI-readable plain text. No business logic is duplicated; no database access.

```
User question
    │
    ▼
Claude / Cursor (MCP host)
    │  calls tool via stdio
    ▼
@llmstatus/mcp (npx, ephemeral process)
    │  GET /v1/...
    ▼
api.llmstatus.io (public, unauthenticated, 30 s Redis cache)
    │
    ▼
Plain-text summary returned to AI
```

---

## 3. Directory Structure

Inside `llmstatus/mcp/` (new directory in the open-source repo):

```
mcp/
├── package.json          # name: @llmstatus/mcp, bin: llmstatus-mcp
├── tsconfig.json
├── src/
│   ├── index.ts          # entry point: registers tools, starts stdio server
│   ├── client.ts         # fetchJSON() with base URL + error handling
│   ├── types.ts          # TypeScript types mirroring Go API response shapes
│   └── tools/
│       ├── list_providers.ts
│       ├── get_provider_status.ts
│       ├── list_active_incidents.ts
│       ├── get_incident_detail.ts
│       ├── get_provider_history.ts
│       └── compare_providers.ts
└── dist/                 # compiled output (.gitignored, included in npm publish)
```

**Runtime dependencies**: `@modelcontextprotocol/sdk` only.  
**Build**: `tsc` (no bundler).  
**Node version**: ≥ 18 (for `fetch` built-in).

---

## 4. Tools

All tools are read-only. Each returns a plain-text summary, not raw JSON.

### 4.1 `list_providers`

- **Endpoint**: `GET /v1/providers`
- **Parameters**: none
- **Output example**:
  ```
  20 providers monitored. Current status:
  ✓ OpenAI — operational (99.97% / 342ms p95)
  ✓ Anthropic — operational (99.95% / 410ms p95)
  ⚠ Groq — degraded (active incident)
  ...
  ```

### 4.2 `get_provider_status`

- **Endpoint**: `GET /v1/providers/{id}`
- **Parameters**: `id: string` (provider ID, e.g. `openai`)
- **Output example**:
  ```
  OpenAI — operational
  Uptime (24h): 99.97% | P95 latency: 342ms
  Models: gpt-4o ✓, gpt-4o-mini ✓, o3 ✓
  No active incidents.
  ```

### 4.3 `list_active_incidents`

- **Endpoint**: `GET /v1/incidents?status=ongoing`
- **Parameters**: `provider_id?: string` (optional filter)
- **Output example**:
  ```
  2 active incidents:
  [major] Groq — elevated error rate (started 14 min ago)
  [minor] Azure OpenAI — degraded latency in us-east (started 2h ago)
  ```

### 4.4 `get_incident_detail`

- **Endpoint**: `GET /v1/incidents/{id}`
- **Parameters**: `id: string` (UUID or slug)
- **Output example**:
  ```
  Incident: Groq elevated error rate
  Status: ongoing | Severity: major
  Provider: groq | Started: 2026-04-26T08:12:00Z
  Affected models: llama-3.3-70b
  Affected regions: us-west-2, us-east-1
  Description: Error rate rose above 15% across US nodes. Rate-limit errors
  are the primary failure type.
  ```

### 4.5 `get_provider_history`

- **Endpoint**: `GET /v1/providers/{id}/history`
- **Parameters**: `id: string`, `window?: "24h" | "7d" | "30d"` (default `"24h"`)
- **Output example**:
  ```
  Anthropic — past 7 days
  Uptime: 99.92% | P95: 395ms avg | P99: 1240ms avg
  Incidents: 1 minor (resolved)
  ```

### 4.6 `compare_providers`

- **Endpoint**: concurrent `GET /v1/providers/{id}` for each ID
- **Parameters**: `ids: string[]` (2–5 provider IDs; capped at 5 to bound concurrent requests)
- **Output example**:
  ```
  Provider comparison (24h):
  Provider       Status       Uptime   P95 (ms)
  ─────────────────────────────────────────────
  openai         operational  99.97%   342
  anthropic      operational  99.95%   410
  google_gemini  operational  99.88%   520
  ```

---

## 5. Error Handling

MCP tools must never throw unhandled errors — they return readable text so the
AI can relay the issue to the user.

| Error | Response text |
|---|---|
| 404 — provider not found | `Provider "X" not found. Valid IDs: openai, anthropic, ...` |
| Network / timeout | `llmstatus.io is temporarily unreachable. Please try again shortly.` |
| API 5xx | `llmstatus.io returned an error (500). Please try again shortly.` |
| Invalid parameter (zod) | Rejected at schema layer; AI auto-corrects |

---

## 6. Configuration

One optional environment variable:

```
LLMSTATUS_API_BASE=https://api.llmstatus.io   # default; override for local dev
```

No API key required.

---

## 7. Installation

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "llmstatus": {
      "command": "npx",
      "args": ["-y", "@llmstatus/mcp"]
    }
  }
}
```

### Cursor

Add to `.cursor/mcp.json` in the project root (or global Cursor settings):

```json
{
  "mcpServers": {
    "llmstatus": {
      "command": "npx",
      "args": ["-y", "@llmstatus/mcp"]
    }
  }
}
```

### Local development override

```json
{
  "mcpServers": {
    "llmstatus": {
      "command": "npx",
      "args": ["-y", "@llmstatus/mcp"],
      "env": { "LLMSTATUS_API_BASE": "http://localhost:8080" }
    }
  }
}
```

---

## 8. Publishing

- Package name: `@llmstatus/mcp`
- Versioned in sync with main repo (semver; PATCH for tool fixes, MINOR for new tools)
- Published via `npm publish` from `llmstatus/mcp/` on tag
- GitHub Actions workflow: `.github/workflows/publish-mcp.yml`
  - Triggers on tags matching `mcp-v*`
  - Runs `tsc`, `npm publish --access public`

---

## 9. Out of Scope (V1)

- HTTP SSE transport (remote/team deployment)
- In-process caching (stdio processes are short-lived; low ROI)
- Authentication / private data access
- Write operations (subscribe, report)
- Resources or Prompts (MCP primitives beyond Tools)

---

## 10. Open Questions (operator decision required)

1. **npm org**: Does `@llmstatus` npm org exist? If not, fall back to unscoped `llmstatus-mcp`.
2. **Versioning coupling**: Should `mcp/` have its own version (`1.0.0`) independent of the Go
   services, or mirror the monorepo version?
3. **Rate limit carve-out**: `npx` installs trigger many concurrent users hitting the API.
   Should MCP tool calls get a dedicated rate-limit bucket in `internal/api/ratelimit.go`?
