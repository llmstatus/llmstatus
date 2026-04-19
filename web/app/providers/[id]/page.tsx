import { notFound } from "next/navigation";
import type { Metadata } from "next";
import Link from "next/link";

import { getProvider, getProviderHistory, ApiNotFoundError } from "@/lib/api";
import { StatusPill } from "@/components/StatusPill";
import { IncidentCard } from "@/components/IncidentCard";
import { ModelList } from "@/components/ModelList";
import { UptimeSparkline } from "@/components/UptimeSparkline";
import { LatencyBar } from "@/components/LatencyBar";

export const revalidate = 60;

type Props = { params: Promise<{ id: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { id } = await params;
  try {
    const provider = await getProvider(id);
    const statusLabel =
      provider.current_status === "operational"
        ? "operational"
        : provider.current_status === "down"
        ? "experiencing an outage"
        : "degraded";
    const description = `${provider.name} API is currently ${statusLabel}. Real-time uptime monitoring from llmstatus.io.`;
    return {
      title: `Is ${provider.name} API down?`,
      description,
      openGraph: { title: `${provider.name} API Status`, description },
      alternates: {
        types: {
          "application/rss+xml": `/api/feed/${id}`,
        },
      },
    };
  } catch {
    return { title: "Provider not found" };
  }
}

export default async function ProviderPage({ params }: Props) {
  const { id } = await params;

  let provider;
  try {
    provider = await getProvider(id);
  } catch (e) {
    if (e instanceof ApiNotFoundError) notFound();
    throw e;
  }

  // History fetch is best-effort — charts are hidden if unavailable.
  const history = await getProviderHistory(id, "30d").catch(() => null);

  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "Service",
    name: `${provider.name} API`,
    serviceType: "AI API",
    ...(provider.status_page_url ? { url: provider.status_page_url } : {}),
    provider: { "@type": "Organization", name: provider.name },
    serviceOutput: "Text / AI inference",
  };

  return (
    <main className="flex-1 mx-auto w-full max-w-4xl px-6 py-10">
      {/* React 19 renders <script> children without HTML-escaping and auto-escapes
          </script> sequences, so plain JSON.stringify is safe here. */}
      <script type="application/ld+json">{JSON.stringify(jsonLd)}</script>
      {/* Breadcrumb */}
      <nav className="mb-6 text-xs text-[var(--ink-400)]">
        <Link href="/" className="hover:text-[var(--ink-200)] transition-colors">
          All providers
        </Link>
        <span className="mx-2">/</span>
        <span className="text-[var(--ink-300)]">{provider.name}</span>
      </nav>

      {/* Header */}
      <div className="mb-8 flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold text-[var(--ink-100)]">
            {provider.name}
          </h1>
          <p className="mt-1 text-sm text-[var(--ink-400)]">
            {provider.category} · {provider.region}
            {provider.status_page_url && (
              <>
                {" · "}
                <a
                  href={provider.status_page_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="hover:text-[var(--ink-200)] transition-colors"
                >
                  Status page ↗
                </a>
              </>
            )}
            {" · "}
            <Link
              href={`/api/feed/${id}`}
              className="hover:text-[var(--ink-200)] transition-colors"
            >
              RSS
            </Link>
          </p>
        </div>
        <StatusPill status={provider.current_status} />
      </div>

      {/* Active incidents */}
      {provider.active_incidents.length > 0 && (
        <section className="mb-8">
          <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-[var(--ink-300)]">
            Active Incidents
          </h2>
          <div className="flex flex-col gap-2">
            {provider.active_incidents.map((inc) => (
              <IncidentCard key={inc.id} incident={inc} href={`/incidents/${inc.slug}`} />
            ))}
          </div>
        </section>
      )}

      {/* Uptime sparkline */}
      {history !== null && (
        <section className="mb-8">
          <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-[var(--ink-300)]">
            Uptime History
          </h2>
          <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-4 py-4">
            <UptimeSparkline buckets={history} days={30} />
          </div>
        </section>
      )}

      {/* Latency bar */}
      {history !== null && history.some((b) => b.p95_ms > 0) && (
        <section className="mb-8">
          <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-[var(--ink-300)]">
            Latency (p95)
          </h2>
          <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-4 py-4">
            <LatencyBar buckets={history} days={30} />
          </div>
        </section>
      )}

      {/* Models */}
      <section>
        <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-[var(--ink-300)]">
          Monitored Models
        </h2>
        <ModelList models={provider.models} />
      </section>
    </main>
  );
}
