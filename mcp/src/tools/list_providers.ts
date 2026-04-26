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
