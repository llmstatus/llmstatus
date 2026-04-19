"use client";

import { useState } from "react";
import type { ProviderSummary, ProviderStatus } from "@/lib/api";
import { ProviderTable } from "./ProviderTable";

type CategoryFilter = "all" | "official" | "aggregator" | "chinese_official";
type StatusFilter = "all" | ProviderStatus;

interface FilterChipProps {
  label: string;
  active: boolean;
  onClick: () => void;
}

function FilterChip({ label, active, onClick }: FilterChipProps) {
  return (
    <button
      onClick={onClick}
      className={`px-3 py-1 rounded text-[11px] font-semibold uppercase tracking-[0.08em] border transition-colors ${
        active
          ? "bg-[var(--canvas-overlay)] border-[var(--ink-400)] text-[var(--ink-100)]"
          : "border-[var(--ink-600)] text-[var(--ink-400)] hover:text-[var(--ink-200)] hover:border-[var(--ink-500)]"
      }`}
    >
      {label}
    </button>
  );
}

interface FilterRowProps {
  label: string;
  children: React.ReactNode;
}

function FilterRow({ label, children }: FilterRowProps) {
  return (
    <div className="flex items-center gap-3 flex-wrap">
      <span className="text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--ink-400)] w-20 shrink-0">
        {label}
      </span>
      {children}
    </div>
  );
}

interface Props {
  providers: ProviderSummary[];
}

export function ProvidersClient({ providers }: Props) {
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("all");
  const [categoryFilter, setCategoryFilter] = useState<CategoryFilter>("all");

  const filtered = providers.filter((p) => {
    const statusOk = statusFilter === "all" || p.current_status === statusFilter;
    const categoryOk = categoryFilter === "all" || p.category === categoryFilter;
    return statusOk && categoryOk;
  });

  const counts = {
    operational: providers.filter((p) => p.current_status === "operational").length,
    degraded: providers.filter((p) => p.current_status === "degraded").length,
    down: providers.filter((p) => p.current_status === "down").length,
    official: providers.filter((p) => p.category === "official").length,
    aggregator: providers.filter((p) => p.category === "aggregator").length,
    chinese_official: providers.filter((p) => p.category === "chinese_official").length,
  };

  const statusChips: { key: StatusFilter; label: string }[] = [
    { key: "all", label: "All" },
    { key: "operational", label: `Operational (${counts.operational})` },
    { key: "degraded", label: `Degraded (${counts.degraded})` },
    { key: "down", label: `Down (${counts.down})` },
  ];

  const categoryChips: { key: CategoryFilter; label: string }[] = [
    { key: "all", label: "All" },
    { key: "official", label: `Official (${counts.official})` },
    { key: "aggregator", label: `Aggregator (${counts.aggregator})` },
    { key: "chinese_official", label: `CN Official (${counts.chinese_official})` },
  ];

  return (
    <div>
      {/* Filters */}
      <div className="mb-6 rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-4 py-4 flex flex-col gap-3">
        <FilterRow label="Status">
          {statusChips.map(({ key, label }) => (
            <FilterChip
              key={key}
              label={label}
              active={statusFilter === key}
              onClick={() => setStatusFilter(key)}
            />
          ))}
        </FilterRow>
        <FilterRow label="Category">
          {categoryChips.map(({ key, label }) => (
            <FilterChip
              key={key}
              label={label}
              active={categoryFilter === key}
              onClick={() => setCategoryFilter(key)}
            />
          ))}
        </FilterRow>
      </div>

      {/* Result count */}
      <p className="mb-3 text-xs text-[var(--ink-400)]">
        {filtered.length === providers.length
          ? `${providers.length} providers`
          : `${filtered.length} of ${providers.length} providers`}
      </p>

      <ProviderTable providers={filtered} />
    </div>
  );
}
