import type { ModelStat } from "@/lib/api";
import { LatencySparkline } from "./LatencySparkline";

function uptimeColor(u: number): string {
  if (u >= 0.995) return "text-[var(--signal-ok)]";
  if (u >= 0.95)  return "text-[var(--signal-warn)]";
  return "text-[var(--signal-down)]";
}

function formatP95(ms: number): string {
  if (ms <= 0) return "—";
  return ms < 1000 ? `${Math.round(ms)}ms` : `${(ms / 1000).toFixed(1)}s`;
}

export function ModelStatsTable({ models }: { models: ModelStat[] }) {
  if (models.length === 0) {
    return <p className="text-sm text-[var(--ink-400)]">No probe data yet.</p>;
  }

  return (
    <div className="overflow-hidden rounded-lg border border-[var(--ink-600)]">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-[var(--ink-600)] bg-[var(--canvas-sunken)]">
            <th className="px-4 py-2.5 text-left text-xs font-semibold uppercase tracking-[0.1em] text-[var(--ink-400)]">
              Model
            </th>
            <th className="px-4 py-2.5 text-right text-xs font-semibold uppercase tracking-[0.1em] text-[var(--ink-400)]">
              Uptime 24h
            </th>
            <th className="px-4 py-2.5 text-right text-xs font-semibold uppercase tracking-[0.1em] text-[var(--ink-400)]">
              p95
            </th>
            <th className="px-4 py-2.5 text-right text-xs font-semibold uppercase tracking-[0.1em] text-[var(--ink-400)]">
              Latency 24h
            </th>
          </tr>
        </thead>
        <tbody>
          {models.map((m) => (
            <tr
              key={m.model_id}
              className="border-t border-[var(--ink-600)] first:border-t-0 bg-[var(--canvas-raised)]"
            >
              <td className="px-4 py-2.5">
                <span
                  className="block font-mono text-sm text-[var(--ink-200)]"
                  title={m.model_id}
                >
                  {m.display_name}
                </span>
                <span className="block font-mono text-xs text-[var(--ink-500)]">
                  {m.model_id}
                </span>
              </td>
              <td className={`px-4 py-2.5 text-right font-mono tabular-nums text-sm ${uptimeColor(m.uptime_24h)}`}>
                {(m.uptime_24h * 100).toFixed(1)}%
              </td>
              <td className="px-4 py-2.5 text-right font-mono tabular-nums text-sm text-[var(--ink-400)]">
                {formatP95(m.p95_ms)}
              </td>
              <td className="px-4 py-2.5 text-right">
                <LatencySparkline data={m.sparkline} width={96} height={24} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
