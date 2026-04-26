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
      const all = await client.listProviders().catch((e: unknown) => {
        console.error("[llmstatus-mcp] failed to fetch provider list for 404 hint:", e);
        return [];
      });
      const ids = all.map((p) => p.id).join(", ");
      return `Provider "${id}" not found.${ids ? ` Valid IDs: ${ids}` : ""}`;
    }
    throw err;
  }
}
