"use client";
import { useState, useMemo } from "react";
import type { IncidentDetail } from "@/lib/api";
import { IncidentCard } from "./IncidentCard";

type StatusFilter = "all" | "active" | "resolved";

interface Props {
  incidents: IncidentDetail[];
}

const STATUS_CHIPS: { value: StatusFilter; label: string }[] = [
  { value: "all",      label: "All" },
  { value: "active",   label: "Ongoing" },
  { value: "resolved", label: "Resolved" },
];

const CHIP_BASE =
  "rounded px-3 py-1 text-xs font-medium transition-colors cursor-pointer";
const CHIP_ON  = `${CHIP_BASE} bg-[var(--canvas-overlay)] text-[var(--ink-100)]`;
const CHIP_OFF = `${CHIP_BASE} text-[var(--ink-400)] hover:text-[var(--ink-200)]`;

export function IncidentsClient({ incidents }: Props) {
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("all");
  const [providerFilter, setProviderFilter] = useState<string>("all");

  // Derive provider list from incidents data.
  const providers = useMemo(() => {
    const seen = new Map<string, string>();
    for (const inc of incidents) {
      if (!seen.has(inc.provider_id)) seen.set(inc.provider_id, inc.provider_id);
    }
    return Array.from(seen.entries()).sort(([a], [b]) => a.localeCompare(b));
  }, [incidents]);

  const filtered = useMemo(() => {
    return incidents.filter((inc) => {
      const statusOk =
        statusFilter === "all" ||
        (statusFilter === "active" && inc.status !== "resolved") ||
        (statusFilter === "resolved" && inc.status === "resolved");
      const providerOk = providerFilter === "all" || inc.provider_id === providerFilter;
      return statusOk && providerOk;
    });
  }, [incidents, statusFilter, providerFilter]);

  return (
    <div>
      {/* Filters */}
      <div className="mb-6 flex flex-wrap items-center gap-4">
        <div className="flex items-center gap-1 rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-sunken)] p-1">
          {STATUS_CHIPS.map(({ value, label }) => (
            <button
              key={value}
              className={statusFilter === value ? CHIP_ON : CHIP_OFF}
              onClick={() => setStatusFilter(value)}
            >
              {label}
            </button>
          ))}
        </div>

        {providers.length > 1 && (
          <select
            className="rounded border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-3 py-1.5 text-xs text-[var(--ink-200)] focus:outline-none focus:ring-1 focus:ring-[var(--ink-400)]"
            value={providerFilter}
            onChange={(e) => setProviderFilter(e.target.value)}
            aria-label="Filter by provider"
          >
            <option value="all">All providers</option>
            {providers.map(([id]) => (
              <option key={id} value={id}>{id}</option>
            ))}
          </select>
        )}

        <span className="ml-auto text-xs text-[var(--ink-500)]">
          {filtered.length} incident{filtered.length !== 1 ? "s" : ""}
        </span>
      </div>

      {/* List */}
      {filtered.length === 0 ? (
        <p className="py-12 text-center text-sm text-[var(--ink-400)]">
          No incidents match the selected filters.
        </p>
      ) : (
        <div className="flex flex-col gap-2">
          {filtered.map((inc) => (
            <IncidentCard key={inc.id} incident={inc} href={`/incidents/${inc.slug}`} />
          ))}
        </div>
      )}
    </div>
  );
}
