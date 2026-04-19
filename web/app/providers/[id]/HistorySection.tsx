"use client";

import { useState, useTransition } from "react";
import type { HistoryBucket, HistoryWindow } from "@/lib/api";
import { fetchProviderHistory } from "./actions";
import { UptimeSparkline } from "@/components/UptimeSparkline";
import { LatencyBar } from "@/components/LatencyBar";

const WINDOWS: { value: HistoryWindow; label: string }[] = [
  { value: "24h", label: "24h" },
  { value: "7d",  label: "7d"  },
  { value: "30d", label: "30d" },
];

interface Props {
  providerId: string;
  initialHistory: HistoryBucket[] | null;
}

export function HistorySection({ providerId, initialHistory }: Props) {
  const [selected, setSelected] = useState<HistoryWindow>("30d");
  const [history, setHistory] = useState<HistoryBucket[] | null>(initialHistory);
  const [isPending, startTransition] = useTransition();

  function selectWindow(w: HistoryWindow) {
    if (w === selected) return;
    setSelected(w);
    startTransition(async () => {
      const data = await fetchProviderHistory(providerId, w);
      setHistory(data);
    });
  }

  const hasData = history !== null && history.length > 0;
  const hasLatency = hasData && history!.some((b) => b.p95_ms > 0);

  return (
    <section className="mb-8">
      {/* Header + window tabs */}
      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-sm font-semibold uppercase tracking-wide text-[var(--ink-300)]">
          History
        </h2>
        <div className="flex gap-px rounded bg-[var(--canvas-sunken)] p-0.5">
          {WINDOWS.map((w) => (
            <button
              key={w.value}
              onClick={() => selectWindow(w.value)}
              className={`rounded px-3 py-1 text-xs font-medium transition-colors ${
                selected === w.value
                  ? "bg-[var(--canvas-raised)] text-[var(--ink-100)] shadow-sm"
                  : "text-[var(--ink-400)] hover:text-[var(--ink-200)]"
              }`}
            >
              {w.label}
            </button>
          ))}
        </div>
      </div>

      {/* Charts */}
      <div className={`space-y-4 transition-opacity duration-150 ${isPending ? "opacity-40" : ""}`}>
        {hasData ? (
          <>
            <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-4 py-4">
              <UptimeSparkline buckets={history!} window={selected} />
            </div>
            {hasLatency && (
              <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-4 py-4">
                <LatencyBar buckets={history!} window={selected} />
              </div>
            )}
          </>
        ) : (
          <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-6 py-8 text-center">
            <p className="text-sm text-[var(--ink-400)]">No history data for this window.</p>
          </div>
        )}
      </div>
    </section>
  );
}
