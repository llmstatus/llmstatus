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
