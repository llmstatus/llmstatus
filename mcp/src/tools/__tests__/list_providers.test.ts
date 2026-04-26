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
