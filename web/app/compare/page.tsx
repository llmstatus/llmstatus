import type { Metadata } from "next";
import Link from "next/link";

import {
  listProviders,
  getProvider,
  getProviderHistory,
  type ProviderDetail,
  type HistoryBucket,
} from "@/lib/api";
import { StatusPill } from "@/components/StatusPill";
import { UptimeSparkline } from "@/components/UptimeSparkline";
import { IncidentCard } from "@/components/IncidentCard";
import { CompareSelector } from "@/components/CompareSelector";

type SearchParams = Promise<{ a?: string; b?: string }>;
type Props = { searchParams: SearchParams };

export async function generateMetadata({ searchParams }: Props): Promise<Metadata> {
  const { a, b } = await searchParams;
  if (!a || !b) return { title: "Compare Providers" };

  const [left, right] = await Promise.all([
    getProvider(a).catch(() => null),
    getProvider(b).catch(() => null),
  ]);
  const nameA = left?.name ?? a;
  const nameB = right?.name ?? b;

  const title = `Compare ${nameA} vs ${nameB}`;
  const description = `Side-by-side uptime, latency, and incident comparison for ${nameA} and ${nameB} — llmstatus.io`;
  return {
    title,
    description,
    openGraph: { title, description },
  };
}

function fmt(n: number | undefined, unit: string, decimals = 0): string {
  if (n === undefined) return "—";
  return `${n.toFixed(decimals)} ${unit}`;
}

function uptimePct(n: number | undefined): string {
  if (n === undefined) return "—";
  return `${(n * 100).toFixed(2)} %`;
}

interface ComparisonRowProps {
  label: string;
  left: React.ReactNode;
  right: React.ReactNode;
}

function Row({ label, left, right }: ComparisonRowProps) {
  return (
    <div className="grid grid-cols-[1fr_1fr_1fr] divide-x divide-[var(--ink-600)] border-b border-[var(--ink-600)] last:border-b-0">
      <div className="px-4 py-3 text-xs text-[var(--ink-400)]">{label}</div>
      <div className="px-4 py-3 text-sm text-[var(--ink-200)]">{left}</div>
      <div className="px-4 py-3 text-sm text-[var(--ink-200)]">{right}</div>
    </div>
  );
}

interface CompareDataProps {
  left: ProviderDetail;
  right: ProviderDetail;
  leftHistory: HistoryBucket[] | null;
  rightHistory: HistoryBucket[] | null;
  a: string;
  b: string;
  providers: Awaited<ReturnType<typeof listProviders>>;
}

function CompareData({ left, right, leftHistory, rightHistory, a, b, providers }: CompareDataProps) {
  return (
    <div>
      {/* Selector row */}
      <div className="mb-8">
        <CompareSelector providers={providers} a={a} b={b} />
      </div>

      {/* Header names */}
      <div className="grid grid-cols-[1fr_1fr_1fr] mb-4 gap-4">
        <div />
        <Link href={`/providers/${left.id}`} className="group">
          <h2 className="text-lg font-semibold text-[var(--ink-100)] group-hover:text-[var(--ink-200)] transition-colors">
            {left.name}
          </h2>
        </Link>
        <Link href={`/providers/${right.id}`} className="group">
          <h2 className="text-lg font-semibold text-[var(--ink-100)] group-hover:text-[var(--ink-200)] transition-colors">
            {right.name}
          </h2>
        </Link>
      </div>

      {/* Metrics table */}
      <div className="mb-8 rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] divide-y divide-[var(--ink-600)] overflow-hidden">
        <Row
          label="Status"
          left={<StatusPill status={left.current_status} />}
          right={<StatusPill status={right.current_status} />}
        />
        <Row
          label="Uptime (24h)"
          left={uptimePct(left.uptime_24h)}
          right={uptimePct(right.uptime_24h)}
        />
        <Row
          label="P95 latency"
          left={fmt(left.p95_ms, "ms")}
          right={fmt(right.p95_ms, "ms")}
        />
        <Row
          label="Active incidents"
          left={`${left.active_incidents.length}`}
          right={`${right.active_incidents.length}`}
        />
        <Row
          label="Category"
          left={left.category}
          right={right.category}
        />
        <Row
          label="Region"
          left={left.region}
          right={right.region}
        />
      </div>

      {/* Uptime sparklines */}
      {(leftHistory !== null || rightHistory !== null) && (
        <section className="mb-8">
          <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-[var(--ink-300)]">
            Uptime History (30d)
          </h2>
          <div className="flex flex-col gap-4">
            {leftHistory !== null && (
              <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-4 py-4">
                <p className="mb-2 text-xs font-medium text-[var(--ink-300)]">{left.name}</p>
                <UptimeSparkline buckets={leftHistory} days={30} />
              </div>
            )}
            {rightHistory !== null && (
              <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-4 py-4">
                <p className="mb-2 text-xs font-medium text-[var(--ink-300)]">{right.name}</p>
                <UptimeSparkline buckets={rightHistory} days={30} />
              </div>
            )}
          </div>
        </section>
      )}

      {/* Active incidents for both */}
      {(left.active_incidents.length > 0 || right.active_incidents.length > 0) && (
        <section className="mb-8">
          <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-[var(--ink-300)]">
            Active Incidents
          </h2>
          <div className="flex flex-col gap-2">
            {[...left.active_incidents, ...right.active_incidents].map((inc) => (
              <IncidentCard key={inc.id} incident={inc} href={`/incidents/${inc.slug}`} />
            ))}
          </div>
        </section>
      )}
    </div>
  );
}

export default async function ComparePage({ searchParams }: Props) {
  const { a, b } = await searchParams;

  const providers = await listProviders().catch(() => []);

  if (!a || !b) {
    return (
      <main className="flex-1 mx-auto w-full max-w-4xl px-6 py-10">
        <div className="mb-8">
          <h1 className="text-2xl font-semibold text-[var(--ink-100)] mb-1">Compare Providers</h1>
          <p className="text-sm text-[var(--ink-400)]">
            Select two providers to compare uptime, latency, and incidents side by side.
          </p>
        </div>
        <CompareSelector providers={providers} a="" b="" />
      </main>
    );
  }

  const [leftResult, rightResult, leftHistResult, rightHistResult] = await Promise.allSettled([
    getProvider(a),
    getProvider(b),
    getProviderHistory(a, "30d"),
    getProviderHistory(b, "30d"),
  ]);

  const left = leftResult.status === "fulfilled" ? leftResult.value : null;
  const right = rightResult.status === "fulfilled" ? rightResult.value : null;

  // If either provider lookup failed, fall back to the selector.
  if (!left || !right) {
    return (
      <main className="flex-1 mx-auto w-full max-w-4xl px-6 py-10">
        <div className="mb-6">
          <h1 className="text-2xl font-semibold text-[var(--ink-100)] mb-1">Compare Providers</h1>
          <p className="text-sm text-[var(--signal-down)]">
            One or both provider IDs could not be found. Please choose again.
          </p>
        </div>
        <CompareSelector providers={providers} a={a} b={b} />
      </main>
    );
  }

  const leftHistory = leftHistResult.status === "fulfilled" ? leftHistResult.value : null;
  const rightHistory = rightHistResult.status === "fulfilled" ? rightHistResult.value : null;

  return (
    <main className="flex-1 mx-auto w-full max-w-4xl px-6 py-10">
      {/* Breadcrumb */}
      <nav className="mb-6 text-xs text-[var(--ink-400)]">
        <Link href="/" className="hover:text-[var(--ink-200)] transition-colors">
          All providers
        </Link>
        <span className="mx-2">/</span>
        <span className="text-[var(--ink-300)]">Compare</span>
      </nav>

      <div className="mb-8">
        <h1 className="text-2xl font-semibold text-[var(--ink-100)]">
          {left.name} <span className="text-[var(--ink-500)]">vs</span> {right.name}
        </h1>
      </div>

      <CompareData
        left={left}
        right={right}
        leftHistory={leftHistory}
        rightHistory={rightHistory}
        a={a}
        b={b}
        providers={providers}
      />
    </main>
  );
}
