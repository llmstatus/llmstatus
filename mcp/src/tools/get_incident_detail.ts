import { z } from "zod";
import type { LLMStatusClient } from "../client.js";
import { ApiError } from "../types.js";
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
  try {
    const inc = await client.getIncident(id);
    return formatIncident(inc);
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) {
      return `Incident "${id}" not found. Use list_active_incidents to see current incidents.`;
    }
    throw err;
  }
}
