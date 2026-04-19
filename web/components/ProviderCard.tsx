import Link from "next/link";
import type { ProviderSummary, ModelStat } from "@/lib/api";
import { LatencySparkline } from "./LatencySparkline";

// ── Status pill ────────────────────────────────────────────────────────────

const STATUS_STYLES = {
  operational: "bg-[var(--signal-ok-bg)] text-[var(--signal-ok)]",
  degraded:    "bg-[var(--signal-warn-bg)] text-[var(--signal-warn)]",
  down:        "bg-[var(--signal-down-bg)] text-[var(--signal-down)]",
} as const;

function StatusPill({ status }: { status: string }) {
  const cls = STATUS_STYLES[status as keyof typeof STATUS_STYLES]
    ?? "bg-[var(--canvas-overlay)] text-[var(--ink-400)]";
  return (
    <span className={`rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide ${cls}`}>
      {status}
    </span>
  );
}

// ── Uptime badge (per-model) ────────────────────────────────────────────────

function uptimeColor(u: number): string {
  if (u >= 0.995) return "text-[var(--signal-ok)]";
  if (u >= 0.95)  return "text-[var(--signal-warn)]";
  return "text-[var(--signal-down)]";
}

function UptimeBadge({ uptime }: { uptime: number }) {
  const pct = (uptime * 100).toFixed(1);
  return (
    <span className={`text-[11px] font-mono tabular-nums ${uptimeColor(uptime)}`}>
      {pct}%
    </span>
  );
}

// ── Model row ──────────────────────────────────────────────────────────────

function ModelRow({ m }: { m: ModelStat }) {
  return (
    <div className="flex items-center gap-2 py-1.5 border-t border-[var(--ink-600)] first:border-t-0">
      <span
        className="flex-1 truncate text-[11px] text-[var(--ink-300)] font-mono"
        title={m.model_id}
      >
        {m.display_name}
      </span>
      <UptimeBadge uptime={m.uptime_24h} />
      <LatencySparkline data={m.sparkline} width={72} height={22} />
      {m.p95_ms > 0 && (
        <span className="w-[42px] text-right text-[11px] font-mono tabular-nums text-[var(--ink-400)]">
          {m.p95_ms < 1000
            ? `${Math.round(m.p95_ms)}ms`
            : `${(m.p95_ms / 1000).toFixed(1)}s`}
        </span>
      )}
    </div>
  );
}

// ── Provider card ──────────────────────────────────────────────────────────

interface Props {
  provider: ProviderSummary;
}

export function ProviderCard({ provider: p }: Props) {
  return (
    <Link
      href={`/providers/${p.id}`}
      className="block rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)]
                 hover:border-[var(--ink-500)] hover:bg-[var(--canvas-overlay)]
                 transition-colors"
    >
      {/* Card header */}
      <div className="flex items-center justify-between px-4 py-3">
        <span className="text-sm font-semibold text-[var(--ink-100)]">{p.name}</span>
        <StatusPill status={p.current_status} />
      </div>

      {/* Model rows */}
      {p.model_stats.length > 0 ? (
        <div className="px-4 pb-3">
          {p.model_stats.map((m) => (
            <ModelRow key={m.model_id} m={m} />
          ))}
        </div>
      ) : (
        <div className="px-4 pb-3">
          <p className="text-[11px] text-[var(--ink-500)]">No probe data yet.</p>
        </div>
      )}
    </Link>
  );
}
