import Link from "next/link";
import type { ProviderSummary } from "@/lib/api";
import { StatusPill } from "./StatusPill";

const CATEGORY_LABEL: Record<string, string> = {
  official: "Official",
  aggregator: "Aggregator",
  chinese_official: "CN Official",
};

const REGION_LABEL: Record<string, string> = {
  global: "Global",
  us: "US",
  cn: "CN",
  eu: "EU",
};

function formatUptime(u: number | undefined): string {
  if (u === undefined) return "—";
  return `${(u * 100).toFixed(1)}%`;
}

function formatP95(ms: number | undefined): string {
  if (ms === undefined || ms === 0) return "—";
  return ms >= 1000 ? `${(ms / 1000).toFixed(1)}s` : `${Math.round(ms)}ms`;
}

export function ProviderTable({ providers }: { providers: ProviderSummary[] }) {
  if (providers.length === 0) {
    return (
      <p className="py-12 text-center text-sm text-[var(--ink-400)]">
        No providers configured yet.
      </p>
    );
  }

  return (
    <div className="overflow-hidden rounded-lg border border-[var(--ink-600)]">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-[var(--ink-600)] bg-[var(--canvas-sunken)]">
            <th className="px-4 py-3 text-left text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--ink-400)]">Provider</th>
            <th className="px-4 py-3 text-left text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--ink-400)]">Category</th>
            <th className="px-4 py-3 text-left text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--ink-400)]">Region</th>
            <th className="px-4 py-3 text-right text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--ink-400)]">Uptime 24h</th>
            <th className="px-4 py-3 text-right text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--ink-400)]">p95</th>
            <th className="px-4 py-3 text-right text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--ink-400)]">Status</th>
          </tr>
        </thead>
        <tbody>
          {providers.map((p, idx) => (
            <tr
              key={p.id}
              className={`border-b border-[var(--ink-600)] last:border-0 transition-colors hover:bg-[var(--canvas-overlay)] ${
                idx % 2 === 0 ? "bg-[var(--canvas-raised)]" : "bg-[var(--canvas-base)]"
              }`}
            >
              <td className="px-4 py-3 font-medium text-[var(--ink-100)]">
                <Link
                  href={`/providers/${p.id}`}
                  className="hover:text-[var(--signal-ok)] transition-colors"
                >
                  {p.name}
                </Link>
              </td>
              <td className="px-4 py-3 text-[var(--ink-300)]">
                {CATEGORY_LABEL[p.category] ?? p.category}
              </td>
              <td className="px-4 py-3 text-[var(--ink-300)]">
                {REGION_LABEL[p.region] ?? p.region}
              </td>
              <td className="px-4 py-3 text-right font-mono text-[var(--ink-200)]">
                {formatUptime(p.uptime_24h)}
              </td>
              <td className="px-4 py-3 text-right font-mono text-[var(--ink-200)]">
                {formatP95(p.p95_ms)}
              </td>
              <td className="px-4 py-3 text-right">
                <StatusPill status={p.current_status} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
