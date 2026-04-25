import type { Metadata } from "next";
import Link from "next/link";
import { listProviders, listIncidents } from "@/lib/api";
import { ProviderCard } from "@/components/ProviderCard";
import { IncidentCard } from "@/components/IncidentCard";

export const revalidate = 30;

export const metadata: Metadata = {
  title: "AI API Status Monitor",
  description:
    "Independent real-time monitoring for the AI infrastructure. " +
    "Measured from 7 global locations. Not scraped from official status pages.",
  openGraph: {
    title: "llmstatus.io — AI API Status Monitor",
    description:
      "Independent real-time monitoring for the AI infrastructure. " +
      "Measured from 7 global locations. Not scraped from official status pages.",
  },
};

export default async function HomePage() {
  const [providerResult, incidentResult] = await Promise.allSettled([
    listProviders(),
    listIncidents("all", 10),
  ]);

  const providers = providerResult.status === "fulfilled" ? providerResult.value : null;
  const allIncidents = incidentResult.status === "fulfilled" ? incidentResult.value : [];

  // Ongoing/monitoring incidents first, then resolved; cap at 5.
  const recentIncidents = [...allIncidents]
    .sort((a, b) => (a.status === "resolved" ? 1 : 0) - (b.status === "resolved" ? 1 : 0))
    .slice(0, 5);

  const allOk =
    providers !== null &&
    providers.length > 0 &&
    providers.every((p) => p.current_status === "operational");

  const hasOutage =
    providers !== null && providers.some((p) => p.current_status === "down");

  const summaryText =
    providers === null
      ? "Unable to load provider data."
      : allOk
      ? "All systems operational."
      : hasOutage
      ? "One or more providers are experiencing issues."
      : "Some providers are degraded.";

  const summaryColor =
    providers === null || hasOutage
      ? "text-[var(--signal-down)]"
      : allOk
      ? "text-[var(--signal-ok)]"
      : "text-[var(--signal-warn)]";

  return (
    <main className="flex-1 mx-auto w-full max-w-4xl px-6">
      {/* Hero — brand spec §7.1 + optional grid background §6.5 */}
      <div
        className="py-14 mb-2"
        style={{
          backgroundImage:
            "linear-gradient(to right, rgba(255,255,255,0.02) 1px, transparent 1px)," +
            "linear-gradient(to bottom, rgba(255,255,255,0.02) 1px, transparent 1px)",
          backgroundSize: "8px 8px",
        }}
      >
        <p className="text-xs font-semibold uppercase tracking-[0.14em] text-[var(--signal-amber)] mb-4">
          llmstatus.io
        </p>
        <h1 className="text-4xl font-semibold text-[var(--ink-100)] leading-tight mb-4">
          Independent real-time monitoring
          <br />
          for the AI infrastructure.
        </h1>
        <p className="text-base text-[var(--ink-400)] leading-relaxed">
          Measured from 7 global locations.
          <br />
          Not scraped from official status pages.
        </p>
      </div>

      <div className="mb-6">
        <p className={`text-base font-medium ${summaryColor}`}>{summaryText}</p>
      </div>

      {/* Recent incidents — shown only when data exists */}
      {recentIncidents.length > 0 && (
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
            {recentIncidents.map((inc) => (
              <IncidentCard key={inc.id} incident={inc} href={`/incidents/${inc.slug}`} />
            ))}
          </div>
        </section>
      )}

      {providers !== null ? (
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
            {providers.map((p) => (
              <ProviderCard key={p.id} provider={p} />
            ))}
          </div>
        </section>
      ) : (
        <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-6 py-10 text-center">
          <p className="text-sm text-[var(--ink-400)]">
            Could not reach the API. Check that the backend is running.
          </p>
        </div>
      )}
    </main>
  );
}
