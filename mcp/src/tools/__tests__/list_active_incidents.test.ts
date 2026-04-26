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
