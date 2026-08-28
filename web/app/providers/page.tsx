import type { Metadata } from "next";
import { listProviders } from "@/lib/api";
import { ProvidersClient } from "@/components/ProvidersClient";
import { StatusDistributionBar } from "@/components/StatusDistributionBar";

export const revalidate = 30;

export const metadata: Metadata = {
  title: "All Providers",
  description:
    "Real-time status for 20+ AI API providers. Filter by status and category. " +
    "Independent monitoring from global locations.",
  openGraph: {
    title: "AI API Providers — llmstatus.io",
    description:
      "Real-time status for 20+ AI API providers monitored from global locations.",
  },
};

export default async function ProvidersPage() {
  const providers = await listProviders().catch(() => null);

  const opCount =
    providers === null
      ? 0
      : providers.filter((p) => p.current_status === "operational").length;
  const degCount =
    providers === null
      ? 0
      : providers.filter((p) => p.current_status === "degraded").length;
  const downCount =
    providers === null
      ? 0
      : providers.filter((p) => p.current_status === "down").length;

  return (
    <main className="flex-1 mx-auto w-full max-w-4xl px-6 py-10">
      <section className="relative overflow-hidden rounded-lg mb-8 px-6 py-8 md:px-8">
        <div
          className="hero-grid absolute inset-0 opacity-60"
          aria-hidden="true"
        />
        <div className="hero-glow absolute inset-0" aria-hidden="true" />
        <div className="relative animate-fade-up">
          <div className="flex items-center gap-3 mb-4">
            <span
              className="status-pulse text-[var(--signal-amber)]"
              aria-hidden="true"
            />
            <p className="text-xs font-semibold uppercase tracking-[0.12em] text-[var(--signal-amber)]">
              Providers
            </p>
          </div>
          <h1 className="text-2xl font-semibold text-[var(--ink-100)] mb-2">
            All monitored providers
          </h1>
          <p className="text-sm text-[var(--ink-400)]">
            Real API calls from global locations. Updated every 30 s.
          </p>
        </div>
      </section>

      {providers === null ? (
        <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-6 py-10 text-center">
          <p className="text-sm text-[var(--ink-400)]">
            Could not reach the API. Check that the backend is running.
          </p>
        </div>
      ) : (
        <>
          <div
            className="mb-6 animate-fade-up"
            style={{ animationDelay: "0.1s" }}
          >
            <StatusDistributionBar
              operational={opCount}
              degraded={degCount}
              down={downCount}
            />
          </div>
          <ProvidersClient providers={providers} />
        </>
      )}
    </main>
  );
}
