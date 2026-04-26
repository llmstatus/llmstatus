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
