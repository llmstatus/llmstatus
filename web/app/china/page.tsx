import type { Metadata } from "next";
import Link from "next/link";
import { listProviders, listIncidents } from "@/lib/api";
import { ProviderCard } from "@/components/ProviderCard";
import { IncidentCard } from "@/components/IncidentCard";

export const revalidate = 30;

export const metadata: Metadata = {
  title: "China AI API Status — llmstatus.io",
  description:
    "Real-time status for Chinese AI API providers: Moonshot, Zhipu AI, 01.AI, Qwen, Minimax, ByteDance (Doubao). " +
    "Monitored from nodes in Shanghai and Guangzhou.",
  openGraph: {
    title: "China AI API Status — llmstatus.io",
    description:
      "Real-time uptime and latency for Chinese AI API providers, measured from mainland China nodes.",
  },
  alternates: {
    types: { "application/rss+xml": "/api/feed" },
  },
};

export default async function ChinaPage() {
  const [providerResult, incidentResult] = await Promise.allSettled([
    listProviders(),
    listIncidents("all", 20),
  ]);

  const allProviders = providerResult.status === "fulfilled" ? providerResult.value : null;
  const cnProviders = allProviders?.filter((p) => p.region === "cn") ?? null;

  const cnProviderIds = new Set(cnProviders?.map((p) => p.id) ?? []);
  const allIncidents = incidentResult.status === "fulfilled" ? incidentResult.value : [];
  const cnIncidents = allIncidents
    .filter((inc) => cnProviderIds.has(inc.provider_id))
    .sort((a, b) => (a.status === "resolved" ? 1 : 0) - (b.status === "resolved" ? 1 : 0))
    .slice(0, 5);

  const allOk =
    cnProviders !== null &&
    cnProviders.length > 0 &&
    cnProviders.every((p) => p.current_status === "operational");

  const hasOutage =
    cnProviders !== null && cnProviders.some((p) => p.current_status === "down");

  const summaryText =
    cnProviders === null
      ? "Unable to load provider data."
      : cnProviders.length === 0
      ? "No Chinese providers found."
      : allOk
      ? "All Chinese providers operational."
      : hasOutage
      ? "One or more Chinese providers are experiencing issues."
      : "Some Chinese providers are degraded.";

  const summaryColor =
    cnProviders === null || hasOutage
      ? "text-[var(--signal-down)]"
      : allOk
      ? "text-[var(--signal-ok)]"
      : "text-[var(--signal-warn)]";

  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "WebPage",
    name: "China AI API Status",
    description: "Real-time status monitoring for Chinese AI API providers.",
    url: "https://llmstatus.io/china",
  };

  return (
    <main className="flex-1 mx-auto w-full max-w-4xl px-6">
      <script type="application/ld+json">{JSON.stringify(jsonLd)}</script>

      <div
        className="py-14 mb-2"
        style={{
          backgroundImage:
            "linear-gradient(to right, rgba(255,255,255,0.02) 1px, transparent 1px)," +
            "linear-gradient(to bottom, rgba(255,255,255,0.02) 1px, transparent 1px)",
          backgroundSize: "8px 8px",
        }}
      >
        <p className="text-xs font-semibold uppercase tracking-[0.12em] text-[var(--signal-amber)] mb-4">
          China view
        </p>
        <h1 className="text-3xl font-semibold text-[var(--ink-100)] leading-tight mb-3">
          Chinese AI API status.
        </h1>
        <p className="text-sm text-[var(--ink-400)] leading-relaxed">
          Monitored from nodes in Shanghai and Guangzhou.
          <br />
          Real API calls — not scraped from status pages.
        </p>
      </div>

      <div className="mb-6">
        <p className={`text-sm font-medium ${summaryColor}`}>{summaryText}</p>
      </div>

      {/* Active incidents for Chinese providers */}
      {cnIncidents.length > 0 && (
        <section className="mb-8">
          <div className="mb-3 flex items-center justify-between">
            <h2 className="text-sm font-semibold uppercase tracking-wide text-[var(--ink-300)]">
              Recent Incidents
            </h2>
            <Link
              href="/incidents"
              className="text-xs text-[var(--ink-400)] hover:text-[var(--ink-200)] transition-colors"
            >
              All incidents →
            </Link>
          </div>
          <div className="flex flex-col gap-2">
            {cnIncidents.map((inc) => (
              <IncidentCard key={inc.id} incident={inc} href={`/incidents/${inc.slug}`} />
            ))}
          </div>
        </section>
      )}

      {cnProviders !== null && cnProviders.length > 0 ? (
        <section>
          <div className="mb-4 flex items-center justify-between">
            <h2 className="text-sm font-semibold uppercase tracking-wide text-[var(--ink-300)]">
              Providers
            </h2>
            <Link
              href="/providers"
              className="text-xs text-[var(--ink-400)] hover:text-[var(--ink-200)] transition-colors"
            >
              All providers →
            </Link>
          </div>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
            {cnProviders.map((p) => (
              <ProviderCard key={p.id} provider={p} />
            ))}
          </div>
        </section>
      ) : (
        <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-6 py-10 text-center">
          <p className="text-sm text-[var(--ink-400)]">
            {cnProviders === null
              ? "Could not reach the API. Check that the backend is running."
              : "No Chinese providers found."}
          </p>
        </div>
      )}

      <div className="mt-10 rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-5 py-4">
        <p className="text-xs font-semibold uppercase tracking-[0.1em] text-[var(--ink-400)] mb-1">
          About this view
        </p>
        <p className="text-xs text-[var(--ink-400)] leading-relaxed">
          These providers are monitored from nodes located inside mainland China. Results reflect
          latency and availability as experienced by users in China. Global providers are available
          on the{" "}
          <Link href="/providers" className="underline underline-offset-2 hover:text-[var(--ink-200)] transition-colors">
            providers page
          </Link>
          .
        </p>
      </div>
    </main>
  );
}
