"use client";
import { useState } from "react";
import Link from "next/link";
import type { ProviderSummary, ModelStat } from "@/lib/api";
import { StatusPill } from "./StatusPill";
import { LatencySparkline } from "./LatencySparkline";

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

const STATUS_RANK: Record<string, number> = {
  down: 0,
  degraded: 1,
  operational: 2,
};

function formatUptime(u: number | undefined): string {
  if (u === undefined) return "—";
  return `${(u * 100).toFixed(1)}%`;
}

function formatP95(ms: number | undefined): string {
  if (ms === undefined || ms === 0) return "—";
  return ms >= 1000 ? `${(ms / 1000).toFixed(1)}s` : `${Math.round(ms)}ms`;
}

type SortKey = "name" | "uptime" | "p95" | "status";
type SortDir = "asc" | "desc" | "default";

function sortProviders(
  providers: ProviderSummary[],
  key: SortKey,
  dir: SortDir,
): ProviderSummary[] {
  if (dir === "default") {
    // default: worst status first, then alphabetical within group
    return [...providers].sort((a, b) => {
      const rankDiff = (STATUS_RANK[a.current_status] ?? 2) - (STATUS_RANK[b.current_status] ?? 2);
      return rankDiff !== 0 ? rankDiff : a.name.localeCompare(b.name);
    });
  }

  const factor = dir === "asc" ? 1 : -1;
  return [...providers].sort((a, b) => {
    switch (key) {
      case "name":
        return factor * a.name.localeCompare(b.name);
      case "uptime":
        return factor * ((a.uptime_24h ?? -1) - (b.uptime_24h ?? -1));
      case "p95":
        return factor * ((a.p95_ms ?? Infinity) - (b.p95_ms ?? Infinity));
      case "status":
        return factor * ((STATUS_RANK[a.current_status] ?? 2) - (STATUS_RANK[b.current_status] ?? 2));
      default:
        return 0;
    }
  });
}

function SortIcon({ col, active, dir }: { col: SortKey; active: SortKey; dir: SortDir }) {
  if (col !== active || dir === "default") {
    return <span className="ml-1 text-[var(--ink-600)]">↕</span>;
  }
  return <span className="ml-1 text-[var(--ink-300)]">{dir === "asc" ? "▲" : "▼"}</span>;
}

export function ProviderTable({ providers }: { providers: ProviderSummary[] }) {
  const [sortKey, setSortKey] = useState<SortKey>("status");
  const [sortDir, setSortDir] = useState<SortDir>("default");

  function handleSort(key: SortKey) {
    if (key !== sortKey) {
      setSortKey(key);
      setSortDir("asc");
      return;
    }
    setSortDir((prev) => (prev === "default" ? "asc" : prev === "asc" ? "desc" : "default"));
  }

  const sorted = sortProviders(providers, sortKey, sortDir);

  if (providers.length === 0) {
    return (
      <p className="py-12 text-center text-sm text-[var(--ink-400)]">
        No providers configured yet.
      </p>
    );
  }

  const thCls =
    "px-4 py-3 text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--ink-400)] " +
    "cursor-pointer select-none hover:text-[var(--ink-200)] transition-colors";

  return (
    <div className="overflow-hidden rounded-lg border border-[var(--ink-600)]">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-[var(--ink-600)] bg-[var(--canvas-sunken)]">
            <th className={`${thCls} text-left`} onClick={() => handleSort("name")}>
              Provider <SortIcon col="name" active={sortKey} dir={sortDir} />
            </th>
            <th className="px-4 py-3 text-left text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--ink-400)]">
              Category
            </th>
            <th className="px-4 py-3 text-left text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--ink-400)]">
              Region
            </th>
            <th className={`${thCls} text-right`} onClick={() => handleSort("uptime")}>
              <SortIcon col="uptime" active={sortKey} dir={sortDir} /> Uptime 24h
            </th>
            <th className={`${thCls} text-right`} onClick={() => handleSort("p95")}>
              <SortIcon col="p95" active={sortKey} dir={sortDir} /> p95
            </th>
            <th className="px-4 py-3 text-right text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--ink-400)]">
              Trend
            </th>
            <th className={`${thCls} text-right`} onClick={() => handleSort("status")}>
              <SortIcon col="status" active={sortKey} dir={sortDir} /> Status
            </th>
          </tr>
        </thead>
        <tbody>
          {sorted.map((p, idx) => (
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
                <LatencySparkline
                  data={aggregateSparklines(p.model_stats ?? [])}
                  width={80}
                  height={22}
                  area
                />
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
