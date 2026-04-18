import type { Metadata } from "next";
import { listProviders } from "@/lib/api";
import { ProviderTable } from "@/components/ProviderTable";

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
  const providers = await listProviders().catch(() => null);

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
        <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-[var(--signal-amber)] mb-4">
          llmstatus.io
        </p>
        <h1 className="text-3xl font-semibold text-[var(--ink-100)] leading-tight mb-3">
          Independent real-time monitoring
          <br />
          for the AI infrastructure.
        </h1>
        <p className="text-sm text-[var(--ink-400)] leading-relaxed">
          Measured from 7 global locations.
          <br />
          Not scraped from official status pages.
        </p>
      </div>

      <div className="mb-6">
        <p className={`text-sm font-medium ${summaryColor}`}>{summaryText}</p>
      </div>

      {providers !== null ? (
        <ProviderTable providers={providers} />
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
