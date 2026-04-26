# MCP Server Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish `@llmstatus/mcp` — a TypeScript MCP server that exposes llmstatus.io data as 6 AI-callable tools so Claude/Cursor users can ask "Is OpenAI down?" and get a live answer.

**Architecture:** Thin stdio MCP server in `llmstatus/mcp/`; calls the public `https://api.llmstatus.io` REST API; formats raw JSON into AI-readable plain text. No business logic, no database access, no auth required.

**Tech Stack:** TypeScript 5, `@modelcontextprotocol/sdk` ^1.0.0, `zod` ^3.23, `vitest` ^1.6 for tests, `tsc` for build, Node ≥ 18.

---

## File Map

| File | Responsibility |
|---|---|
| `mcp/package.json` | Package metadata, scripts, dependencies |
| `mcp/tsconfig.json` | TypeScript compiler options (ESM, NodeNext) |
| `mcp/src/types.ts` | TypeScript interfaces mirroring Go API JSON shapes; `ApiError` class |
| `mcp/src/client.ts` | `LLMStatusClient` — typed `fetch` wrapper for all API endpoints |
| `mcp/src/tools/list_providers.ts` | Format provider list; export tool name/description/schema/handler |
| `mcp/src/tools/get_provider_status.ts` | Format single provider detail |
| `mcp/src/tools/list_active_incidents.ts` | Format incident list (client-side provider filter) |
| `mcp/src/tools/get_incident_detail.ts` | Format single incident detail |
| `mcp/src/tools/get_provider_history.ts` | Format history bucket summary |
| `mcp/src/tools/compare_providers.ts` | Concurrent fetch + aligned table format |
| `mcp/src/index.ts` | Wire `McpServer`, register 6 tools, connect `StdioServerTransport` |
| `mcp/src/tools/__tests__/*.test.ts` | Vitest unit tests for each formatter |
| `.github/workflows/publish-mcp.yml` | CI: build + test + `npm publish` on `mcp-v*` tag |

---

## Task 1: Scaffold `mcp/` directory

**Files:**
- Create: `mcp/package.json`
- Create: `mcp/tsconfig.json`

- [ ] **Step 1: Create `mcp/package.json`**

```json
{
  "name": "@llmstatus/mcp",
  "version": "1.0.0",
  "description": "MCP server for llmstatus.io — query AI provider status from Claude and Cursor",
  "type": "module",
  "bin": {
    "llmstatus-mcp": "./dist/index.js"
  },
  "files": [
    "dist"
  ],
  "scripts": {
    "build": "tsc",
    "test": "vitest run",
    "test:watch": "vitest",
    "prepublishOnly": "npm run build && npm test"
  },
  "dependencies": {
    "@modelcontextprotocol/sdk": "^1.0.0",
    "zod": "^3.23.0"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "typescript": "^5.4.0",
    "vitest": "^1.6.0"
  },
  "engines": {
    "node": ">=18"
  },
  "license": "Apache-2.0",
  "repository": {
    "type": "git",
    "url": "https://github.com/llmstatus/llmstatus"
  }
}
```

- [ ] **Step 2: Create `mcp/tsconfig.json`**

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "declaration": true,
    "sourceMap": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}
```

- [ ] **Step 3: Install dependencies**

```bash
cd mcp && npm install
```

Expected: `node_modules/` created, `package-lock.json` written, no errors.

- [ ] **Step 4: Verify TypeScript compiler resolves**

```bash
cd mcp && npx tsc --version
```

Expected: `Version 5.x.x`

- [ ] **Step 5: Commit**

```bash
git add mcp/package.json mcp/tsconfig.json mcp/package-lock.json
git commit -m "chore(mcp): scaffold package"
```

---

## Task 2: Types and HTTP client

**Files:**
- Create: `mcp/src/types.ts`
- Create: `mcp/src/client.ts`
- Create: `mcp/src/tools/__tests__/client.test.ts`

- [ ] **Step 1: Create `mcp/src/types.ts`**

```typescript
export interface ApiEnvelope<T> {
  data: T;
  meta: {
    generated_at: string;
    cache_ttl_s: number;
  };
}

export interface ModelStat {
  model_id: string;
  display_name: string;
  uptime_24h: number;
  p95_ms: number;
  sparkline: number[];
}

export interface ProviderSummary {
  id: string;
  name: string;
  category: string;
  region: string;
  probe_scope: string;
  current_status: "operational" | "degraded" | "down";
  active_incident_id?: string;
  uptime_24h?: number;
  p95_ms?: number;
  model_stats: ModelStat[];
}

export interface ModelSummary {
  model_id: string;
  display_name: string;
  model_type: string;
  active: boolean;
}

export interface IncidentRef {
  id: string;
  slug: string;
  severity: string;
  title: string;
  status: string;
  started_at: string;
}

export interface RegionStat {
  region_id: string;
  uptime_24h: number;
  p95_ms: number;
}

export interface ProviderDetail extends ProviderSummary {
  status_page_url?: string;
  documentation_url?: string;
  models: ModelSummary[];
  active_incidents: IncidentRef[];
  region_stats: RegionStat[];
}

export interface IncidentResponse {
  id: string;
  slug: string;
  provider_id: string;
  severity: "critical" | "major" | "minor";
  title: string;
  description?: string;
  status: "ongoing" | "monitoring" | "resolved";
  affected_models: string[];
  affected_regions: string[];
  started_at: string;
  resolved_at?: string;
  detection_method: string;
  human_reviewed: boolean;
}

export interface HistoryBucket {
  timestamp: string;
  total: number;
  errors: number;
  uptime: number;
  p95_ms: number;
}

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    message: string
  ) {
    super(message);
    this.name = "ApiError";
  }
}
```

- [ ] **Step 2: Create `mcp/src/client.ts`**

```typescript
import type {
  ApiEnvelope,
  HistoryBucket,
  IncidentResponse,
  ProviderDetail,
  ProviderSummary,
} from "./types.js";
import { ApiError } from "./types.js";

const DEFAULT_BASE_URL = "https://api.llmstatus.io";

export class LLMStatusClient {
  private readonly baseUrl: string;

  constructor(baseUrl?: string) {
    this.baseUrl = (baseUrl ?? process.env["LLMSTATUS_API_BASE"] ?? DEFAULT_BASE_URL).replace(
      /\/$/,
      ""
    );
  }

  private async fetchJSON<T>(path: string): Promise<T> {
    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}${path}`, {
        headers: { "User-Agent": "@llmstatus/mcp/1.0.0" },
        signal: AbortSignal.timeout(10_000),
      });
    } catch {
      throw new ApiError(
        0,
        "llmstatus.io is temporarily unreachable. Please try again shortly."
      );
    }
    if (!response.ok) {
      throw new ApiError(
        response.status,
        `llmstatus.io returned HTTP ${response.status}.${response.status === 404 ? " Resource not found." : " Please try again shortly."}`
      );
    }
    const envelope = (await response.json()) as ApiEnvelope<T>;
    return envelope.data;
  }

  listProviders(): Promise<ProviderSummary[]> {
    return this.fetchJSON<ProviderSummary[]>("/v1/providers");
  }

  getProvider(id: string): Promise<ProviderDetail> {
    return this.fetchJSON<ProviderDetail>(`/v1/providers/${encodeURIComponent(id)}`);
  }

  listIncidents(params: { status?: string; limit?: number } = {}): Promise<IncidentResponse[]> {
    const qs = new URLSearchParams();
    if (params.status) qs.set("status", params.status);
    if (params.limit != null) qs.set("limit", String(params.limit));
    const suffix = qs.toString() ? `?${qs.toString()}` : "";
    return this.fetchJSON<IncidentResponse[]>(`/v1/incidents${suffix}`);
  }

  getIncident(id: string): Promise<IncidentResponse> {
    return this.fetchJSON<IncidentResponse>(`/v1/incidents/${encodeURIComponent(id)}`);
  }

  getProviderHistory(id: string, window = "30d"): Promise<HistoryBucket[]> {
    return this.fetchJSON<HistoryBucket[]>(
      `/v1/providers/${encodeURIComponent(id)}/history?window=${encodeURIComponent(window)}`
    );
  }
}
```

- [ ] **Step 3: Create `mcp/src/tools/__tests__/` directory and write client tests**

Create `mcp/src/tools/__tests__/client.test.ts`:

```typescript
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { LLMStatusClient } from "../../client.js";
import { ApiError } from "../../types.js";

const mockFetch = vi.fn();
beforeEach(() => {
  vi.stubGlobal("fetch", mockFetch);
});
afterEach(() => {
  vi.restoreAllMocks();
});

function makeEnvelope<T>(data: T) {
  return { data, meta: { generated_at: "2026-04-26T00:00:00Z", cache_ttl_s: 30 } };
}

describe("LLMStatusClient.listProviders", () => {
  it("returns parsed provider array on 200", async () => {
    const providers = [{ id: "openai", name: "OpenAI" }];
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => makeEnvelope(providers),
    });
    const client = new LLMStatusClient("http://localhost");
    const result = await client.listProviders();
    expect(result).toEqual(providers);
    expect(mockFetch).toHaveBeenCalledWith(
      "http://localhost/v1/providers",
      expect.objectContaining({ headers: expect.any(Object) })
    );
  });

  it("throws ApiError on 404", async () => {
    mockFetch.mockResolvedValueOnce({ ok: false, status: 404 });
    const client = new LLMStatusClient("http://localhost");
    await expect(client.listProviders()).rejects.toBeInstanceOf(ApiError);
    await expect(client.listProviders()).rejects.toMatchObject({ status: 404 });
  });

  it("throws ApiError on network failure", async () => {
    mockFetch.mockRejectedValueOnce(new TypeError("fetch failed"));
    const client = new LLMStatusClient("http://localhost");
    await expect(client.listProviders()).rejects.toBeInstanceOf(ApiError);
    await expect(client.listProviders()).rejects.toMatchObject({ status: 0 });
  });
});

describe("LLMStatusClient.listIncidents", () => {
  it("appends status query param when provided", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => makeEnvelope([]),
    });
    const client = new LLMStatusClient("http://localhost");
    await client.listIncidents({ status: "ongoing" });
    expect(mockFetch).toHaveBeenCalledWith(
      "http://localhost/v1/incidents?status=ongoing",
      expect.anything()
    );
  });
});
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd mcp && npm test
```

Expected: all client tests PASS.

- [ ] **Step 5: Commit**

```bash
git add mcp/src/types.ts mcp/src/client.ts "mcp/src/tools/__tests__/client.test.ts"
git commit -m "feat(mcp): add types and HTTP client"
```

---

## Task 3: `list_providers` tool

**Files:**
- Create: `mcp/src/tools/list_providers.ts`
- Create: `mcp/src/tools/__tests__/list_providers.test.ts`

- [ ] **Step 1: Write the failing formatter test**

Create `mcp/src/tools/__tests__/list_providers.test.ts`:

```typescript
import { describe, expect, it } from "vitest";
import { formatProviderList } from "../list_providers.js";
import type { ProviderSummary } from "../../types.js";

const base: ProviderSummary = {
  id: "openai",
  name: "OpenAI",
  category: "official",
  region: "us",
  probe_scope: "global",
  current_status: "operational",
  uptime_24h: 0.9997,
  p95_ms: 342,
  model_stats: [],
};

describe("formatProviderList", () => {
  it("shows count and operational tick", () => {
    const out = formatProviderList([base]);
    expect(out).toContain("1 provider");
    expect(out).toContain("✓");
    expect(out).toContain("OpenAI");
    expect(out).toContain("operational");
    expect(out).toContain("99.97%");
    expect(out).toContain("342ms");
  });

  it("shows warning icon for degraded provider", () => {
    const out = formatProviderList([{ ...base, current_status: "degraded" }]);
    expect(out).toContain("⚠");
  });

  it("shows cross icon for down provider", () => {
    const out = formatProviderList([{ ...base, current_status: "down" }]);
    expect(out).toContain("✗");
  });

  it("handles missing uptime/latency gracefully", () => {
    const minimal = { ...base, uptime_24h: undefined, p95_ms: undefined };
    expect(() => formatProviderList([minimal])).not.toThrow();
  });

  it("returns a message for empty list", () => {
    const out = formatProviderList([]);
    expect(out).toContain("0 provider");
  });
});
```

- [ ] **Step 2: Run to verify test fails**

```bash
cd mcp && npm test -- list_providers
```

Expected: FAIL — `formatProviderList` not found.

- [ ] **Step 3: Implement `mcp/src/tools/list_providers.ts`**

```typescript
import type { LLMStatusClient } from "../client.js";
import type { ProviderSummary } from "../types.js";

export const TOOL_NAME = "list_providers";
export const TOOL_DESCRIPTION =
  "List all AI providers monitored by llmstatus.io with their current operational status, 24-hour uptime, and latency.";
export const TOOL_SCHEMA = {};

export function formatProviderList(providers: ProviderSummary[]): string {
  const count = providers.length;
  if (count === 0) return "0 providers found.";

  const icon = (s: ProviderSummary["current_status"]) =>
    s === "operational" ? "✓" : s === "down" ? "✗" : "⚠";

  const lines = [`${count} provider${count !== 1 ? "s" : ""} monitored:\n`];
  for (const p of providers) {
    const uptime =
      p.uptime_24h != null ? ` (${(p.uptime_24h * 100).toFixed(2)}% uptime 24h)` : "";
    const p95 = p.p95_ms != null ? ` / ${Math.round(p.p95_ms)}ms p95` : "";
    lines.push(`${icon(p.current_status)} ${p.name} [${p.id}] — ${p.current_status}${uptime}${p95}`);
  }
  return lines.join("\n");
}

export async function handleListProviders(client: LLMStatusClient): Promise<string> {
  const providers = await client.listProviders();
  return formatProviderList(providers);
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd mcp && npm test -- list_providers
```

Expected: all 5 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add mcp/src/tools/list_providers.ts "mcp/src/tools/__tests__/list_providers.test.ts"
git commit -m "feat(mcp): add list_providers tool"
```

---

## Task 4: `get_provider_status` tool

**Files:**
- Create: `mcp/src/tools/get_provider_status.ts`
- Create: `mcp/src/tools/__tests__/get_provider_status.test.ts`

- [ ] **Step 1: Write the failing formatter test**

Create `mcp/src/tools/__tests__/get_provider_status.test.ts`:

```typescript
import { describe, expect, it } from "vitest";
import { formatProviderDetail } from "../get_provider_status.js";
import type { ProviderDetail } from "../../types.js";

const base: ProviderDetail = {
  id: "openai",
  name: "OpenAI",
  category: "official",
  region: "us",
  probe_scope: "global",
  current_status: "operational",
  uptime_24h: 0.9997,
  p95_ms: 342,
  model_stats: [
    { model_id: "gpt-4o", display_name: "GPT-4o", uptime_24h: 0.9998, p95_ms: 340, sparkline: [] },
    { model_id: "gpt-4o-mini", display_name: "GPT-4o mini", uptime_24h: 0.9999, p95_ms: 210, sparkline: [] },
  ],
  models: [],
  active_incidents: [],
  region_stats: [],
};

describe("formatProviderDetail", () => {
  it("shows name, status, uptime, and latency", () => {
    const out = formatProviderDetail(base);
    expect(out).toContain("OpenAI");
    expect(out).toContain("operational");
    expect(out).toContain("99.97%");
    expect(out).toContain("342ms");
  });

  it("shows model list", () => {
    const out = formatProviderDetail(base);
    expect(out).toContain("GPT-4o");
    expect(out).toContain("GPT-4o mini");
  });

  it("shows 'No active incidents' when clean", () => {
    const out = formatProviderDetail(base);
    expect(out).toContain("No active incidents");
  });

  it("shows active incident title when present", () => {
    const withInc: ProviderDetail = {
      ...base,
      current_status: "degraded",
      active_incidents: [
        { id: "abc", slug: "test", severity: "major", title: "Elevated errors", status: "ongoing", started_at: "2026-04-26T10:00:00Z" },
      ],
    };
    const out = formatProviderDetail(withInc);
    expect(out).toContain("Elevated errors");
    expect(out).toContain("major");
  });
});
```

- [ ] **Step 2: Run to verify test fails**

```bash
cd mcp && npm test -- get_provider_status
```

Expected: FAIL.

- [ ] **Step 3: Implement `mcp/src/tools/get_provider_status.ts`**

```typescript
import { z } from "zod";
import type { LLMStatusClient } from "../client.js";
import { ApiError } from "../types.js";
import type { ProviderDetail } from "../types.js";

export const TOOL_NAME = "get_provider_status";
export const TOOL_DESCRIPTION =
  "Get the current operational status, uptime, latency, and active incidents for a specific AI provider.";
export const TOOL_SCHEMA = {
  id: z.string().describe("Provider ID, e.g. 'openai', 'anthropic', 'google_gemini'"),
};

export function formatProviderDetail(p: ProviderDetail): string {
  const icon =
    p.current_status === "operational" ? "✓" : p.current_status === "down" ? "✗" : "⚠";
  const lines = [`${p.name} — ${icon} ${p.current_status}`];
  if (p.uptime_24h != null) lines.push(`Uptime (24h): ${(p.uptime_24h * 100).toFixed(2)}%`);
  if (p.p95_ms != null) lines.push(`P95 latency: ${Math.round(p.p95_ms)}ms`);

  const models = p.model_stats.filter((m) => m.uptime_24h > 0);
  if (models.length > 0) {
    lines.push(
      `Models: ${models.map((m) => `${m.display_name} (${(m.uptime_24h * 100).toFixed(1)}%)`).join(", ")}`
    );
  }

  if (p.active_incidents.length > 0) {
    lines.push("\nActive incidents:");
    for (const inc of p.active_incidents) {
      lines.push(`  [${inc.severity}] ${inc.title}`);
    }
  } else {
    lines.push("No active incidents.");
  }

  return lines.join("\n");
}

export async function handleGetProviderStatus(
  id: string,
  client: LLMStatusClient
): Promise<string> {
  try {
    const detail = await client.getProvider(id);
    return formatProviderDetail(detail);
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) {
      const all = await client.listProviders().catch(() => []);
      const ids = all.map((p) => p.id).join(", ");
      return `Provider "${id}" not found.${ids ? ` Valid IDs: ${ids}` : ""}`;
    }
    throw err;
  }
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd mcp && npm test -- get_provider_status
```

Expected: all 4 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add mcp/src/tools/get_provider_status.ts "mcp/src/tools/__tests__/get_provider_status.test.ts"
git commit -m "feat(mcp): add get_provider_status tool"
```

---

## Task 5: `list_active_incidents` tool

**Files:**
- Create: `mcp/src/tools/list_active_incidents.ts`
- Create: `mcp/src/tools/__tests__/list_active_incidents.test.ts`

- [ ] **Step 1: Write the failing formatter test**

Create `mcp/src/tools/__tests__/list_active_incidents.test.ts`:

```typescript
import { describe, expect, it } from "vitest";
import { formatIncidentList } from "../list_active_incidents.js";
import type { IncidentResponse } from "../../types.js";

const makeInc = (overrides: Partial<IncidentResponse> = {}): IncidentResponse => ({
  id: "abc123",
  slug: "2026-04-26-openai-errors",
  provider_id: "openai",
  severity: "major",
  title: "Elevated error rate",
  status: "ongoing",
  affected_models: ["gpt-4o"],
  affected_regions: ["us-west-2"],
  started_at: new Date(Date.now() - 14 * 60 * 1000).toISOString(), // 14 min ago
  detection_method: "auto",
  human_reviewed: false,
  ...overrides,
});

describe("formatIncidentList", () => {
  it("returns no-incident message for empty list", () => {
    const out = formatIncidentList([]);
    expect(out).toContain("No active incidents");
  });

  it("shows count, provider, title, severity", () => {
    const out = formatIncidentList([makeInc()]);
    expect(out).toContain("1 active incident");
    expect(out).toContain("openai");
    expect(out).toContain("Elevated error rate");
    expect(out).toContain("major");
  });

  it("shows plural for multiple incidents", () => {
    const out = formatIncidentList([makeInc(), makeInc({ id: "def456", provider_id: "anthropic" })]);
    expect(out).toContain("2 active incidents");
  });

  it("filters by provider_id when specified", () => {
    const incidents = [makeInc({ provider_id: "openai" }), makeInc({ id: "z", provider_id: "anthropic" })];
    const out = formatIncidentList(incidents, "openai");
    expect(out).toContain("1 active incident");
    expect(out).not.toContain("anthropic");
  });
});
```

- [ ] **Step 2: Run to verify test fails**

```bash
cd mcp && npm test -- list_active_incidents
```

Expected: FAIL.

- [ ] **Step 3: Implement `mcp/src/tools/list_active_incidents.ts`**

```typescript
import { z } from "zod";
import type { LLMStatusClient } from "../client.js";
import type { IncidentResponse } from "../types.js";

export const TOOL_NAME = "list_active_incidents";
export const TOOL_DESCRIPTION =
  "List all currently active (ongoing) incidents across all monitored AI providers. Optionally filter to one provider.";
export const TOOL_SCHEMA = {
  provider_id: z
    .string()
    .optional()
    .describe("Optional provider ID to filter results, e.g. 'openai'"),
};

function formatAge(isoString: string): string {
  const ms = Date.now() - new Date(isoString).getTime();
  const minutes = Math.floor(ms / 60_000);
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ${minutes % 60}m`;
  return `${Math.floor(hours / 24)}d ${hours % 24}h`;
}

export function formatIncidentList(incidents: IncidentResponse[], providerFilter?: string): string {
  const filtered = providerFilter
    ? incidents.filter((i) => i.provider_id === providerFilter)
    : incidents;

  if (filtered.length === 0) {
    return providerFilter
      ? `No active incidents for provider "${providerFilter}".`
      : "No active incidents. All monitored providers are operating normally.";
  }

  const count = filtered.length;
  const lines = [`${count} active incident${count !== 1 ? "s" : ""}:\n`];
  for (const inc of filtered) {
    const age = formatAge(inc.started_at);
    lines.push(`[${inc.severity}] ${inc.title} (${inc.provider_id}) — started ${age} ago`);
  }
  return lines.join("\n");
}

export async function handleListActiveIncidents(
  providerFilter: string | undefined,
  client: LLMStatusClient
): Promise<string> {
  const incidents = await client.listIncidents({ status: "ongoing", limit: 100 });
  return formatIncidentList(incidents, providerFilter);
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd mcp && npm test -- list_active_incidents
```

Expected: all 4 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add mcp/src/tools/list_active_incidents.ts "mcp/src/tools/__tests__/list_active_incidents.test.ts"
git commit -m "feat(mcp): add list_active_incidents tool"
```

---

## Task 6: `get_incident_detail` tool

**Files:**
- Create: `mcp/src/tools/get_incident_detail.ts`
- Create: `mcp/src/tools/__tests__/get_incident_detail.test.ts`

- [ ] **Step 1: Write the failing formatter test**

Create `mcp/src/tools/__tests__/get_incident_detail.test.ts`:

```typescript
import { describe, expect, it } from "vitest";
import { formatIncident } from "../get_incident_detail.js";
import type { IncidentResponse } from "../../types.js";

const incident: IncidentResponse = {
  id: "uuid-123",
  slug: "2026-04-26-openai-errors",
  provider_id: "openai",
  severity: "major",
  title: "Elevated error rate",
  description: "Error rate rose above 15% across US nodes.",
  status: "ongoing",
  affected_models: ["gpt-4o", "gpt-4o-mini"],
  affected_regions: ["us-west-2", "us-east-1"],
  started_at: "2026-04-26T08:12:00Z",
  detection_method: "auto",
  human_reviewed: true,
};

describe("formatIncident", () => {
  it("includes title, severity, provider, started_at", () => {
    const out = formatIncident(incident);
    expect(out).toContain("Elevated error rate");
    expect(out).toContain("major");
    expect(out).toContain("openai");
    expect(out).toContain("2026-04-26T08:12:00Z");
  });

  it("lists affected models and regions", () => {
    const out = formatIncident(incident);
    expect(out).toContain("gpt-4o");
    expect(out).toContain("us-west-2");
  });

  it("includes description when present", () => {
    const out = formatIncident(incident);
    expect(out).toContain("Error rate rose above 15%");
  });

  it("includes resolved_at when present", () => {
    const resolved = { ...incident, resolved_at: "2026-04-26T09:00:00Z" };
    const out = formatIncident(resolved);
    expect(out).toContain("2026-04-26T09:00:00Z");
  });

  it("omits description line when absent", () => {
    const noDesc = { ...incident, description: undefined };
    const out = formatIncident(noDesc);
    expect(out).not.toContain("undefined");
  });
});
```

- [ ] **Step 2: Run to verify test fails**

```bash
cd mcp && npm test -- get_incident_detail
```

Expected: FAIL.

- [ ] **Step 3: Implement `mcp/src/tools/get_incident_detail.ts`**

```typescript
import { z } from "zod";
import type { LLMStatusClient } from "../client.js";
import type { IncidentResponse } from "../types.js";

export const TOOL_NAME = "get_incident_detail";
export const TOOL_DESCRIPTION =
  "Get full details for a specific incident by its UUID or slug (e.g. '2026-04-26-openai-errors').";
export const TOOL_SCHEMA = {
  id: z
    .string()
    .describe("Incident UUID or slug, e.g. '2026-04-26-openai-elevated-errors' or a UUID"),
};

export function formatIncident(inc: IncidentResponse): string {
  const lines = [
    `Incident: ${inc.title}`,
    `Status: ${inc.status} | Severity: ${inc.severity}`,
    `Provider: ${inc.provider_id}`,
    `Started: ${inc.started_at}`,
  ];
  if (inc.resolved_at) lines.push(`Resolved: ${inc.resolved_at}`);
  if (inc.affected_models.length > 0)
    lines.push(`Affected models: ${inc.affected_models.join(", ")}`);
  if (inc.affected_regions.length > 0)
    lines.push(`Affected regions: ${inc.affected_regions.join(", ")}`);
  if (inc.description) lines.push(`\n${inc.description}`);
  return lines.join("\n");
}

export async function handleGetIncidentDetail(
  id: string,
  client: LLMStatusClient
): Promise<string> {
  const inc = await client.getIncident(id);
  return formatIncident(inc);
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd mcp && npm test -- get_incident_detail
```

Expected: all 5 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add mcp/src/tools/get_incident_detail.ts "mcp/src/tools/__tests__/get_incident_detail.test.ts"
git commit -m "feat(mcp): add get_incident_detail tool"
```

---

## Task 7: `get_provider_history` tool

**Files:**
- Create: `mcp/src/tools/get_provider_history.ts`
- Create: `mcp/src/tools/__tests__/get_provider_history.test.ts`

- [ ] **Step 1: Write the failing formatter test**

Create `mcp/src/tools/__tests__/get_provider_history.test.ts`:

```typescript
import { describe, expect, it } from "vitest";
import { formatHistory } from "../get_provider_history.js";
import type { HistoryBucket } from "../../types.js";

const makeBucket = (uptime: number, p95_ms: number, errors = 0): HistoryBucket => ({
  timestamp: "2026-04-25T00:00:00Z",
  total: 100,
  errors,
  uptime,
  p95_ms,
});

describe("formatHistory", () => {
  it("shows provider name, window, avg uptime, avg p95", () => {
    const buckets = [makeBucket(1.0, 300), makeBucket(0.99, 400), makeBucket(0.98, 500)];
    const out = formatHistory("OpenAI", "7d", buckets);
    expect(out).toContain("OpenAI");
    expect(out).toContain("7d");
    expect(out).toContain("99.00%"); // avg of 1.0+0.99+0.98 = 2.97/3
    expect(out).toContain("400ms"); // avg of 300+400+500 = 400
  });

  it("handles empty bucket list", () => {
    const out = formatHistory("OpenAI", "7d", []);
    expect(out).toContain("No history data");
  });

  it("skips p95 avg when all buckets have zero latency", () => {
    const out = formatHistory("OpenAI", "24h", [makeBucket(1.0, 0)]);
    expect(out).not.toContain("NaN");
  });
});
```

- [ ] **Step 2: Run to verify test fails**

```bash
cd mcp && npm test -- get_provider_history
```

Expected: FAIL.

- [ ] **Step 3: Implement `mcp/src/tools/get_provider_history.ts`**

```typescript
import { z } from "zod";
import type { LLMStatusClient } from "../client.js";
import type { HistoryBucket } from "../types.js";

export const TOOL_NAME = "get_provider_history";
export const TOOL_DESCRIPTION =
  "Get historical uptime and latency data for a provider over a given time window.";
export const TOOL_SCHEMA = {
  id: z.string().describe("Provider ID, e.g. 'openai'"),
  window: z
    .enum(["24h", "7d", "30d"])
    .optional()
    .default("30d")
    .describe("Time window: '24h', '7d', or '30d'. Defaults to '30d'."),
};

export function formatHistory(
  providerName: string,
  window: string,
  buckets: HistoryBucket[]
): string {
  if (buckets.length === 0) {
    return `No history data available for ${providerName} in the ${window} window.`;
  }
  const avgUptime = buckets.reduce((s, b) => s + b.uptime, 0) / buckets.length;
  const withLatency = buckets.filter((b) => b.p95_ms > 0);
  const avgP95 =
    withLatency.length > 0
      ? withLatency.reduce((s, b) => s + b.p95_ms, 0) / withLatency.length
      : 0;
  const totalErrors = buckets.reduce((s, b) => s + b.errors, 0);

  const lines = [
    `${providerName} — past ${window}`,
    `Uptime: ${(avgUptime * 100).toFixed(2)}%`,
  ];
  if (avgP95 > 0) lines.push(`Avg P95 latency: ${Math.round(avgP95)}ms`);
  lines.push(`Total error probes: ${totalErrors}`);
  lines.push(`Buckets: ${buckets.length}`);
  return lines.join("\n");
}

export async function handleGetProviderHistory(
  id: string,
  window: "24h" | "7d" | "30d",
  client: LLMStatusClient
): Promise<string> {
  const [buckets, detail] = await Promise.all([
    client.getProviderHistory(id, window),
    client.getProvider(id),
  ]);
  return formatHistory(detail.name, window, buckets);
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd mcp && npm test -- get_provider_history
```

Expected: all 3 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add mcp/src/tools/get_provider_history.ts "mcp/src/tools/__tests__/get_provider_history.test.ts"
git commit -m "feat(mcp): add get_provider_history tool"
```

---

## Task 8: `compare_providers` tool

**Files:**
- Create: `mcp/src/tools/compare_providers.ts`
- Create: `mcp/src/tools/__tests__/compare_providers.test.ts`

- [ ] **Step 1: Write the failing formatter test**

Create `mcp/src/tools/__tests__/compare_providers.test.ts`:

```typescript
import { describe, expect, it } from "vitest";
import { formatComparison } from "../compare_providers.js";
import type { ProviderDetail } from "../../types.js";

const makeProvider = (
  id: string,
  name: string,
  status: "operational" | "degraded" | "down" = "operational"
): ProviderDetail => ({
  id,
  name,
  category: "official",
  region: "us",
  probe_scope: "global",
  current_status: status,
  uptime_24h: 0.9997,
  p95_ms: 342,
  model_stats: [],
  models: [],
  active_incidents: [],
  region_stats: [],
});

describe("formatComparison", () => {
  it("renders a table with all providers", () => {
    const out = formatComparison([
      { type: "ok", data: makeProvider("openai", "OpenAI") },
      { type: "ok", data: makeProvider("anthropic", "Anthropic") },
    ]);
    expect(out).toContain("OpenAI");
    expect(out).toContain("Anthropic");
    expect(out).toContain("operational");
    expect(out).toContain("99.97%");
  });

  it("shows error row for failed provider fetch", () => {
    const out = formatComparison([
      { type: "ok", data: makeProvider("openai", "OpenAI") },
      { type: "error", id: "bad_id", message: "not found" },
    ]);
    expect(out).toContain("bad_id");
    expect(out).toContain("not found");
  });

  it("handles N/A for missing uptime and latency", () => {
    const noStats = makeProvider("openai", "OpenAI");
    noStats.uptime_24h = undefined;
    noStats.p95_ms = undefined;
    const out = formatComparison([{ type: "ok", data: noStats }]);
    expect(out).toContain("N/A");
    expect(out).not.toContain("undefined");
  });
});
```

- [ ] **Step 2: Run to verify test fails**

```bash
cd mcp && npm test -- compare_providers
```

Expected: FAIL.

- [ ] **Step 3: Implement `mcp/src/tools/compare_providers.ts`**

```typescript
import { z } from "zod";
import type { LLMStatusClient } from "../client.js";
import type { ProviderDetail } from "../types.js";

export const TOOL_NAME = "compare_providers";
export const TOOL_DESCRIPTION =
  "Compare 2 to 5 AI providers side-by-side: current status, 24-hour uptime, and P95 latency.";
export const TOOL_SCHEMA = {
  ids: z
    .array(z.string())
    .min(2)
    .max(5)
    .describe("2 to 5 provider IDs to compare, e.g. ['openai', 'anthropic']"),
};

type CompareRow =
  | { type: "ok"; data: ProviderDetail }
  | { type: "error"; id: string; message: string };

export function formatComparison(rows: CompareRow[]): string {
  const COL = [22, 16, 10, 10];
  const header = [
    "Provider".padEnd(COL[0]),
    "Status".padEnd(COL[1]),
    "Uptime".padEnd(COL[2]),
    "P95 (ms)",
  ].join(" ");
  const sep = "─".repeat(COL[0] + COL[1] + COL[2] + 12);

  const lines = [`Provider comparison (24h):`, sep, header, sep];
  for (const row of rows) {
    if (row.type === "error") {
      lines.push(`${row.id.padEnd(COL[0])} error: ${row.message}`);
    } else {
      const p = row.data;
      const uptime = p.uptime_24h != null ? `${(p.uptime_24h * 100).toFixed(2)}%` : "N/A";
      const p95 = p.p95_ms != null ? `${Math.round(p.p95_ms)}` : "N/A";
      lines.push(
        [
          p.name.padEnd(COL[0]),
          p.current_status.padEnd(COL[1]),
          uptime.padEnd(COL[2]),
          p95,
        ].join(" ")
      );
    }
  }
  return lines.join("\n");
}

export async function handleCompareProviders(
  ids: string[],
  client: LLMStatusClient
): Promise<string> {
  const settled = await Promise.allSettled(ids.map((id) => client.getProvider(id)));
  const rows: CompareRow[] = settled.map((result, i) => {
    if (result.status === "fulfilled") {
      return { type: "ok", data: result.value };
    }
    const msg = result.reason instanceof Error ? result.reason.message : "unknown error";
    return { type: "error", id: ids[i]!, message: msg };
  });
  return formatComparison(rows);
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd mcp && npm test -- compare_providers
```

Expected: all 3 tests PASS.

- [ ] **Step 5: Run all tests to confirm nothing broken**

```bash
cd mcp && npm test
```

Expected: all tests across all tool files PASS.

- [ ] **Step 6: Commit**

```bash
git add mcp/src/tools/compare_providers.ts "mcp/src/tools/__tests__/compare_providers.test.ts"
git commit -m "feat(mcp): add compare_providers tool"
```

---

## Task 9: Entry point — wire the MCP server

**Files:**
- Create: `mcp/src/index.ts`

- [ ] **Step 1: Create `mcp/src/index.ts`**

```typescript
#!/usr/bin/env node
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { LLMStatusClient } from "./client.js";
import {
  TOOL_NAME as LIST_PROVIDERS,
  TOOL_DESCRIPTION as LP_DESC,
  handleListProviders,
} from "./tools/list_providers.js";
import {
  TOOL_NAME as GET_PROVIDER_STATUS,
  TOOL_DESCRIPTION as GPS_DESC,
  TOOL_SCHEMA as GPS_SCHEMA,
  handleGetProviderStatus,
} from "./tools/get_provider_status.js";
import {
  TOOL_NAME as LIST_INCIDENTS,
  TOOL_DESCRIPTION as LI_DESC,
  TOOL_SCHEMA as LI_SCHEMA,
  handleListActiveIncidents,
} from "./tools/list_active_incidents.js";
import {
  TOOL_NAME as GET_INCIDENT,
  TOOL_DESCRIPTION as GI_DESC,
  TOOL_SCHEMA as GI_SCHEMA,
  handleGetIncidentDetail,
} from "./tools/get_incident_detail.js";
import {
  TOOL_NAME as GET_HISTORY,
  TOOL_DESCRIPTION as GH_DESC,
  TOOL_SCHEMA as GH_SCHEMA,
  handleGetProviderHistory,
} from "./tools/get_provider_history.js";
import {
  TOOL_NAME as COMPARE,
  TOOL_DESCRIPTION as CMP_DESC,
  TOOL_SCHEMA as CMP_SCHEMA,
  handleCompareProviders,
} from "./tools/compare_providers.js";
import { ApiError } from "./types.js";

const client = new LLMStatusClient();
const server = new McpServer({ name: "llmstatus", version: "1.0.0" });

function wrapError(err: unknown): string {
  if (err instanceof ApiError) return err.message;
  if (err instanceof Error) return err.message;
  return "An unexpected error occurred.";
}

server.tool(LIST_PROVIDERS, LP_DESC, {}, async () => ({
  content: [{ type: "text" as const, text: await handleListProviders(client).catch(wrapError) }],
}));

server.tool(GET_PROVIDER_STATUS, GPS_DESC, GPS_SCHEMA, async ({ id }) => ({
  content: [
    { type: "text" as const, text: await handleGetProviderStatus(id, client).catch(wrapError) },
  ],
}));

server.tool(LIST_INCIDENTS, LI_DESC, LI_SCHEMA, async ({ provider_id }) => ({
  content: [
    {
      type: "text" as const,
      text: await handleListActiveIncidents(provider_id, client).catch(wrapError),
    },
  ],
}));

server.tool(GET_INCIDENT, GI_DESC, GI_SCHEMA, async ({ id }) => ({
  content: [
    { type: "text" as const, text: await handleGetIncidentDetail(id, client).catch(wrapError) },
  ],
}));

server.tool(GET_HISTORY, GH_DESC, GH_SCHEMA, async ({ id, window }) => ({
  content: [
    {
      type: "text" as const,
      text: await handleGetProviderHistory(id, window ?? "30d", client).catch(wrapError),
    },
  ],
}));

server.tool(COMPARE, CMP_DESC, CMP_SCHEMA, async ({ ids }) => ({
  content: [
    { type: "text" as const, text: await handleCompareProviders(ids, client).catch(wrapError) },
  ],
}));

const transport = new StdioServerTransport();
await server.connect(transport);
```

- [ ] **Step 2: Build**

```bash
cd mcp && npm run build
```

Expected: `dist/` directory created, no TypeScript errors.

- [ ] **Step 3: Run all tests one final time**

```bash
cd mcp && npm test
```

Expected: all tests PASS.

- [ ] **Step 4: Smoke-test the binary locally**

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | node dist/index.js
```

Expected: JSON response listing all 6 tools with their names and descriptions.

- [ ] **Step 5: Commit**

```bash
git add mcp/src/index.ts
git commit -m "feat(mcp): wire MCP server entry point"
```

---

## Task 10: GitHub Actions publish workflow

**Files:**
- Create: `.github/workflows/publish-mcp.yml`

- [ ] **Step 1: Create `.github/workflows/publish-mcp.yml`**

```yaml
name: Publish @llmstatus/mcp

on:
  push:
    tags:
      - "mcp-v*"

jobs:
  publish:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      id-token: write
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: "20"
          registry-url: "https://registry.npmjs.org"

      - name: Install dependencies
        working-directory: mcp
        run: npm ci

      - name: Run tests
        working-directory: mcp
        run: npm test

      - name: Build
        working-directory: mcp
        run: npm run build

      - name: Publish
        working-directory: mcp
        run: npm publish --access public
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```

- [ ] **Step 2: Verify workflow file is valid YAML**

```bash
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/publish-mcp.yml'))" && echo "YAML valid"
```

Expected: `YAML valid`

- [ ] **Step 3: Add `mcp/dist` to `.gitignore` inside `mcp/`**

Create `mcp/.gitignore`:

```
dist/
node_modules/
*.tsbuildinfo
```

- [ ] **Step 4: Commit**

```bash
git add .github/workflows/publish-mcp.yml mcp/.gitignore
git commit -m "ci: add publish-mcp GitHub Actions workflow"
```

---

## Self-Review Checklist

**Spec coverage:**
- [x] §3 Directory structure — Task 1
- [x] §4.1 `list_providers` — Task 3
- [x] §4.2 `get_provider_status` — Task 4 (including 404 → valid IDs fallback)
- [x] §4.3 `list_active_incidents` (with provider_id filter) — Task 5
- [x] §4.4 `get_incident_detail` — Task 6
- [x] §4.5 `get_provider_history` — Task 7
- [x] §4.6 `compare_providers` (capped at 5, `Promise.allSettled`) — Task 8
- [x] §5 Error handling (ApiError, network, 404) — Task 2 client + Task 4 handler
- [x] §6 `LLMSTATUS_API_BASE` env var — Task 2 client constructor
- [x] §7 Installation docs — covered in spec, no code needed
- [x] §8 Publishing workflow — Task 10

**Type consistency:**
- `LLMStatusClient` defined in Task 2, used identically in Tasks 3–9 ✓
- `ApiError` imported from `types.ts` in Task 4 handler ✓
- `CompareRow` discriminated union defined and used in same file (Task 8) ✓
- `TOOL_SCHEMA` exported from each tool and imported in `index.ts` (Task 9) ✓
- `window` parameter typed as `"24h" | "7d" | "30d"` consistently across Task 7 and Task 9 ✓
