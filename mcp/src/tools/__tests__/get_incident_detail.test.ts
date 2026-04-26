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
