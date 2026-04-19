import { notFound } from "next/navigation";
import type { Metadata } from "next";
import Link from "next/link";

import { getProvider, getProviderHistory, ApiNotFoundError } from "@/lib/api";
import { StatusPill } from "@/components/StatusPill";
import { IncidentCard } from "@/components/IncidentCard";
import { ModelStatsTable } from "@/components/ModelStatsTable";
import { HistorySection } from "./HistorySection";

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
      alternates: { types: { "application/rss+xml": `/api/feed/${id}` } },
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

  // History is best-effort — the client component handles empty state.
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
          <h1 className="text-2xl font-semibold text-[var(--ink-100)]">{provider.name}</h1>
          <p className="mt-1 text-sm text-[var(--ink-400)]">
            {provider.category}
            {provider.region && (
              <span className="text-[var(--ink-600)]"> · Coverage: {provider.region}</span>
            )}
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
            <Link href={`/api/feed/${id}`} className="hover:text-[var(--ink-200)] transition-colors">
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

      {/* History charts with window selector (client component) */}
      <HistorySection providerId={id} initialHistory={history} />

      {/* Regional breakdown */}
      {(provider.region_stats ?? []).length > 0 && (
        <section className="mb-8">
          <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-[var(--ink-300)]">
            Regional Status
          </h2>
          <div className="overflow-hidden rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)]">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-[var(--ink-600)]">
                  <th className="px-4 py-2.5 text-left text-[10px] font-semibold uppercase tracking-[0.1em] text-[var(--ink-500)]">
                    Probe region
                  </th>
                  <th className="px-4 py-2.5 text-right text-[10px] font-semibold uppercase tracking-[0.1em] text-[var(--ink-500)]">
                    Uptime 24h
                  </th>
                  <th className="px-4 py-2.5 text-right text-[10px] font-semibold uppercase tracking-[0.1em] text-[var(--ink-500)]">
                    p95 latency
                  </th>
                </tr>
              </thead>
              <tbody>
                {[...(provider.region_stats ?? [])]
                  .sort((a, b) => a.uptime_24h - b.uptime_24h)
                  .map((reg) => {
                    const dotColor =
                      reg.uptime_24h >= 0.995 ? "bg-[var(--signal-ok)]"
                      : reg.uptime_24h >= 0.95 ? "bg-[var(--signal-warn)]"
                      : "bg-[var(--signal-down)]";
                    const uptimeColor =
                      reg.uptime_24h >= 0.995 ? "text-[var(--signal-ok)]"
                      : reg.uptime_24h >= 0.95 ? "text-[var(--signal-warn)]"
                      : "text-[var(--signal-down)]";
                    const barColor =
                      reg.p95_ms <= 500  ? "bg-[var(--signal-ok)]"
                      : reg.p95_ms <= 2000 ? "bg-[var(--signal-warn)]"
                      : "bg-[var(--signal-down)]";
                    const barWidth = reg.p95_ms > 0
                      ? Math.min(100, Math.round((reg.p95_ms / 3000) * 100))
                      : 0;
                    const p95 = reg.p95_ms > 0
                      ? reg.p95_ms < 1000
                        ? `${Math.round(reg.p95_ms)}ms`
                        : `${(reg.p95_ms / 1000).toFixed(1)}s`
                      : "—";
                    return (
                      <tr key={reg.region_id} className="border-t border-[var(--ink-600)] first:border-t-0">
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-[3px]">
                            <span className={`h-2 w-2 flex-shrink-0 rounded-full ${dotColor}`} />
                            <span className="font-mono text-[12px] text-[var(--ink-300)]">
                              {reg.region_id}
                            </span>
                          </div>
                        </td>
                        <td className={`px-4 py-3 text-right font-mono tabular-nums text-[12px] ${uptimeColor}`}>
                          {(reg.uptime_24h * 100).toFixed(1)}%
                        </td>
                        <td className="px-4 py-3">
                          <div className="flex items-center justify-end gap-2.5">
                            <div className="h-1.5 w-24 overflow-hidden rounded-full bg-[var(--ink-600)]">
                              <div
                                className={`h-full rounded-full ${barColor}`}
                                style={{ width: `${barWidth}%` }}
                              />
                            </div>
                            <span className="w-12 text-right font-mono tabular-nums text-[12px] text-[var(--ink-400)]">
                              {p95}
                            </span>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
              </tbody>
            </table>
          </div>
        </section>
      )}

      {/* Per-model stats */}
      <section>
        <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-[var(--ink-300)]">
          Models
        </h2>
        <ModelStatsTable models={provider.model_stats ?? []} />
      </section>
    </main>
  );
}
