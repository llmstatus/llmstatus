import Link from "next/link";
import type { ProviderSummary, ModelStat } from "@/lib/api";
import { LatencySparkline } from "./LatencySparkline";

// ── Status ─────────────────────────────────────────────────────────────────

const STATUS_PILL: Record<string, string> = {
  operational: "bg-[var(--signal-ok-bg)] text-[var(--signal-ok)]",
  degraded:    "bg-[var(--signal-warn-bg)] text-[var(--signal-warn)]",
  down:        "bg-[var(--signal-down-bg)] text-[var(--signal-down)]",
};

const STATUS_DOT: Record<string, string> = {
  operational: "bg-[var(--signal-ok)]",
  degraded:    "bg-[var(--signal-warn)]",
  down:        "bg-[var(--signal-down)]",
};

const STATUS_ACCENT: Record<string, string> = {
  operational: "shadow-[inset_3px_0_0_var(--signal-ok)]",
  degraded:    "shadow-[inset_3px_0_0_var(--signal-warn)]",
  down:        "shadow-[inset_3px_0_0_var(--signal-down)]",
};

const STATUS_COLOR: Record<string, string> = {
  operational: "var(--signal-ok)",
  degraded:    "var(--signal-warn)",
  down:        "var(--signal-down)",
};

// ── Uptime color ────────────────────────────────────────────────────────────

function uptimeColor(u: number): string {
  if (u >= 0.995) return "text-[var(--signal-ok)]";
  if (u >= 0.95)  return "text-[var(--signal-warn)]";
  return "text-[var(--signal-down)]";
}

// ── Aggregate sparklines across all models ──────────────────────────────────

function aggregateSparklines(modelStats: ModelStat[]): number[] {
  const BUCKETS = 60;
  const sums = new Array<number>(BUCKETS).fill(0);
  const counts = new Array<number>(BUCKETS).fill(0);
  for (const m of modelStats) {
    for (let i = 0; i < BUCKETS; i++) {
      if (m.sparkline[i] > 0) {
        sums[i] += m.sparkline[i];
        counts[i]++;
      }
    }
  }
  return sums.map((s, i) => (counts[i] > 0 ? s / counts[i] : 0));
}

// ── Provider card ───────────────────────────────────────────────────────────

export function ProviderCard({ provider: p }: { provider: ProviderSummary }) {
  const uptime = p.uptime_24h !== undefined
    ? (p.uptime_24h * 100).toFixed(1) + "%"
    : "—";

  const p95 = p.p95_ms !== undefined
    ? p.p95_ms < 1000
      ? `${Math.round(p.p95_ms)}ms`
      : `${(p.p95_ms / 1000).toFixed(1)}s`
    : "—";

  const modelNames = (p.model_stats ?? []).map((m) => m.display_name);
  const sparkline = aggregateSparklines(p.model_stats ?? []);

  const pillCls = STATUS_PILL[p.current_status] ?? "bg-[var(--canvas-overlay)] text-[var(--ink-400)]";
  const dotCls  = STATUS_DOT[p.current_status]  ?? "bg-[var(--ink-500)]";
  const accentCls = STATUS_ACCENT[p.current_status] ?? "";
  const sparkColor = STATUS_COLOR[p.current_status] ?? "var(--viz-1)";

  return (
    <Link
      href={`/providers/${p.id}`}
      className={`flex flex-col rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)]
                  hover:border-[var(--ink-500)] hover:bg-[var(--canvas-overlay)]
                  transition-colors ${accentCls}`}
    >
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3.5">
        <div className="flex items-center gap-2.5 min-w-0">
          <span className={`w-2 h-2 rounded-full shrink-0 ${dotCls}`} />
          <span className="text-base font-semibold text-[var(--ink-100)] truncate">{p.name}</span>
        </div>
        <span className={`shrink-0 rounded px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider ${pillCls}`}>
          {p.current_status}
        </span>
      </div>

      {/* Sparkline — full width, taller, area fill */}
      <div className="px-4 pb-2">
        <LatencySparkline
          data={sparkline}
          width={300}
          height={48}
          className="w-full"
          area
          color={sparkColor}
        />
      </div>

      {/* Stats row */}
      <div className="flex items-center gap-6 border-t border-[var(--ink-600)] px-4 py-3">
        <div className="flex flex-col gap-0.5">
          <span className="text-[10px] font-semibold uppercase tracking-[0.1em] text-[var(--ink-500)]">
            Uptime 24h
          </span>
          <span className={`text-base font-mono tabular-nums font-semibold ${
            p.uptime_24h !== undefined ? uptimeColor(p.uptime_24h) : "text-[var(--ink-500)]"
          }`}>
            {uptime}
          </span>
        </div>
        <div className="flex flex-col gap-0.5">
          <span className="text-[10px] font-semibold uppercase tracking-[0.1em] text-[var(--ink-500)]">
            p95 Latency
          </span>
          <span className="text-base font-mono tabular-nums font-semibold text-[var(--ink-200)]">
            {p95}
          </span>
        </div>
      </div>

      {/* Model tags */}
      {modelNames.length > 0 && (
        <div className="flex flex-wrap gap-1 border-t border-[var(--ink-600)] px-4 pb-3 pt-2.5">
          {modelNames.slice(0, 4).map((name) => (
            <span
              key={name}
              className="rounded bg-[var(--canvas-overlay)] px-2 py-0.5 text-[11px] text-[var(--ink-400)]"
            >
              {name}
            </span>
          ))}
          {modelNames.length > 4 && (
            <span className="px-1 py-0.5 text-[11px] text-[var(--ink-600)]">
              +{modelNames.length - 4}
            </span>
          )}
        </div>
      )}
    </Link>
  );
}
