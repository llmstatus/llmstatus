import { notFound } from "next/navigation";
import type { Metadata } from "next";
import Link from "next/link";

import { getIncident, ApiNotFoundError, type IncidentDetail, type Severity } from "@/lib/api";
import { formatDate } from "@/components/IncidentCard";
import { ProbeTimestamp } from "@/components/ProbeTimestamp";

export const revalidate = 60;

type Props = { params: Promise<{ slug: string }> };

function incidentTitle(inc: IncidentDetail): string {
  const date = new Date(inc.started_at).toLocaleDateString("en-US", {
    month: "long",
    day: "numeric",
    year: "numeric",
    timeZone: "UTC",
  });
  return `${inc.provider_id} incident on ${date}: ${inc.title}`;
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;
  try {
    const inc = await getIncident(slug);
    const title = incidentTitle(inc);
    const description = inc.description ?? inc.title;
    return {
      title,
      description,
      openGraph: { title, description },
    };
  } catch {
    return { title: "Incident not found" };
  }
}

const SEVERITY_STYLE: Record<Severity, { label: string; color: string }> = {
  critical: { label: "Critical", color: "text-[var(--signal-down)]" },
  major:    { label: "Major",    color: "text-[var(--signal-warn)]" },
  minor:    { label: "Minor",    color: "text-[var(--ink-300)]" },
};

const STATUS_STYLE: Record<string, string> = {
  ongoing:    "text-[var(--signal-down)]",
  monitoring: "text-[var(--signal-warn)]",
  resolved:   "text-[var(--signal-ok)]",
};

export default async function IncidentPage({ params }: Props) {
  const { slug } = await params;

  let inc: IncidentDetail;
  try {
    inc = await getIncident(slug);
  } catch (e) {
    if (e instanceof ApiNotFoundError) notFound();
    throw e;
  }

  const { label: severityLabel, color: severityColor } =
    SEVERITY_STYLE[inc.severity] ?? SEVERITY_STYLE.minor;

  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "Event",
    name: inc.title,
    startDate: inc.started_at,
    ...(inc.resolved_at ? { endDate: inc.resolved_at } : {}),
    description: inc.description ?? inc.title,
    eventStatus:
      inc.status === "resolved"
        ? "https://schema.org/EventCancelled"
        : "https://schema.org/EventScheduled",
  };

  return (
    <main className="flex-1 mx-auto w-full max-w-4xl px-6 py-10">
      <script type="application/ld+json">
        {JSON.stringify(jsonLd)}
      </script>

      {/* Breadcrumb */}
      <nav className="mb-6 text-xs text-[var(--ink-400)]">
        <Link href="/" className="hover:text-[var(--ink-200)] transition-colors">
          All providers
        </Link>
        <span className="mx-2">/</span>
        <Link href="/incidents" className="hover:text-[var(--ink-200)] transition-colors">
          Incidents
        </Link>
        <span className="mx-2">/</span>
        <span className="text-[var(--ink-300)]">{inc.slug}</span>
      </nav>

      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center gap-3 mb-2">
          <span className={`text-xs font-semibold uppercase tracking-wide ${severityColor}`}>
            {severityLabel}
          </span>
          <span className="text-[var(--ink-600)]">·</span>
          <span className={`text-xs font-semibold uppercase tracking-wide ${STATUS_STYLE[inc.status] ?? ""}`}>
            {inc.status}
          </span>
        </div>
        <h1 className="text-2xl font-semibold text-[var(--ink-100)] mb-1">{inc.title}</h1>
        <div className="flex items-center gap-3 mt-1">
          <p className="text-sm text-[var(--ink-400)]">
            Provider:{" "}
            <Link
              href={`/providers/${inc.provider_id}`}
              className="text-[var(--ink-300)] hover:text-[var(--ink-200)] transition-colors"
            >
              {inc.provider_id}
            </Link>
          </p>
          <span className="text-[var(--ink-600)]">·</span>
          <ProbeTimestamp iso={inc.started_at} prefix="Started" />
        </div>
      </div>

      {/* Timeline */}
      <section className="mb-8 rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] divide-y divide-[var(--ink-600)]">
        <Row label="Started" value={formatDate(inc.started_at)} />
        {inc.resolved_at && (
          <Row label="Resolved" value={formatDate(inc.resolved_at)} />
        )}
        <Row label="Detection" value={inc.detection_method} />
        <Row label="Human reviewed" value={inc.human_reviewed ? "Yes" : "No"} />
      </section>

      {/* Description */}
      {inc.description && (
        <section className="mb-8">
          <h2 className="mb-2 text-sm font-semibold uppercase tracking-wide text-[var(--ink-300)]">
            Description
          </h2>
          <p className="text-sm text-[var(--ink-200)] leading-6">{inc.description}</p>
        </section>
      )}

      {/* Affected models */}
      {inc.affected_models.length > 0 && (
        <section className="mb-6">
          <h2 className="mb-2 text-sm font-semibold uppercase tracking-wide text-[var(--ink-300)]">
            Affected Models
          </h2>
          <ul className="flex flex-wrap gap-2">
            {inc.affected_models.map((m) => (
              <li
                key={m}
                className="rounded px-2 py-0.5 font-mono text-xs bg-[var(--canvas-sunken)] text-[var(--ink-200)]"
              >
                {m}
              </li>
            ))}
          </ul>
        </section>
      )}

      {/* Affected regions */}
      {inc.affected_regions.length > 0 && (
        <section className="mb-8">
          <h2 className="mb-2 text-sm font-semibold uppercase tracking-wide text-[var(--ink-300)]">
            Affected Regions
          </h2>
          <ul className="flex flex-wrap gap-2">
            {inc.affected_regions.map((r) => (
              <li
                key={r}
                className="rounded px-2 py-0.5 font-mono text-xs bg-[var(--canvas-sunken)] text-[var(--ink-200)]"
              >
                {r}
              </li>
            ))}
          </ul>
        </section>
      )}

      <p className="text-xs text-[var(--ink-500)]">
        This URL is permanent. Detection method: {inc.detection_method}.
      </p>
    </main>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center px-4 py-3 gap-6">
      <span className="w-36 shrink-0 text-xs text-[var(--ink-400)]">{label}</span>
      <span className="text-sm text-[var(--ink-200)]">{value}</span>
    </div>
  );
}
