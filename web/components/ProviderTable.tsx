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
            <th className="px-4 py-3 text-left font-medium text-[var(--ink-300)]">Provider</th>
            <th className="px-4 py-3 text-left font-medium text-[var(--ink-300)]">Category</th>
            <th className="px-4 py-3 text-left font-medium text-[var(--ink-300)]">Region</th>
            <th className="px-4 py-3 text-right font-medium text-[var(--ink-300)]">Status</th>
          </tr>
        </thead>
        <tbody>
          {providers.map((p, idx) => (
            <tr
              key={p.id}
              className={`border-b border-[var(--ink-600)] last:border-0 ${
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
