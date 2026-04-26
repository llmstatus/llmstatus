import { z } from "zod";
import type { LLMStatusClient } from "../client.js";
import type { ProviderDetail } from "../types.js";

export const TOOL_NAME = "compare_providers";
export const TOOL_DESCRIPTION =
  "Compare 2 to 5 AI providers side-by-side: current status, 24-hour uptime, and P95 latency.";
export const TOOL_SCHEMA = {
  ids: z
    .array(z.string())
    .min(2)
    .max(5)
    .describe("2 to 5 provider IDs to compare, e.g. ['openai', 'anthropic']"),
};

type CompareRow =
  | { type: "ok"; data: ProviderDetail }
  | { type: "error"; id: string; message: string };

export function formatComparison(rows: CompareRow[]): string {
  const COL = [22, 16, 10, 10];
  const header = [
    "Provider".padEnd(COL[0]),
    "Status".padEnd(COL[1]),
    "Uptime".padEnd(COL[2]),
    "P95 (ms)",
  ].join(" ");
  const sep = "─".repeat(COL[0] + COL[1] + COL[2] + 12);

  const lines = [`Provider comparison (24h):`, sep, header, sep];
  for (const row of rows) {
    if (row.type === "error") {
      lines.push(`${row.id.padEnd(COL[0])} error: ${row.message}`);
    } else {
      const p = row.data;
      const uptime = p.uptime_24h != null ? `${(p.uptime_24h * 100).toFixed(2)}%` : "N/A";
      const p95 = p.p95_ms != null ? `${Math.round(p.p95_ms)}` : "N/A";
      lines.push(
        [
          p.name.padEnd(COL[0]),
          p.current_status.padEnd(COL[1]),
          uptime.padEnd(COL[2]),
          p95,
        ].join(" ")
      );
    }
  }
  return lines.join("\n");
}

export async function handleCompareProviders(
  ids: string[],
  client: LLMStatusClient
): Promise<string> {
  const settled = await Promise.allSettled(ids.map((id) => client.getProvider(id)));
  const rows: CompareRow[] = settled.map((result, i) => {
    if (result.status === "fulfilled") {
      return { type: "ok", data: result.value };
    }
    const msg = result.reason instanceof Error ? result.reason.message : "unknown error";
    return { type: "error", id: ids[i]!, message: msg };
  });
  return formatComparison(rows);
}
