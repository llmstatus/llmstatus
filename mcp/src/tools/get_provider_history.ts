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
