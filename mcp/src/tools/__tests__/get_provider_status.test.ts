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
