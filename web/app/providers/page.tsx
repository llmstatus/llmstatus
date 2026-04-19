import type { Metadata } from "next";
import { listProviders } from "@/lib/api";
import { ProvidersClient } from "@/components/ProvidersClient";

export const revalidate = 30;

export const metadata: Metadata = {
  title: "All Providers",
  description:
    "Real-time status for 20+ AI API providers. Filter by status and category. " +
    "Independent monitoring from 7 global locations.",
  openGraph: {
    title: "AI API Providers — llmstatus.io",
    description:
      "Real-time status for 20+ AI API providers monitored from 7 global locations.",
  },
};

export default async function ProvidersPage() {
  const providers = await listProviders().catch(() => null);

  return (
    <main className="flex-1 mx-auto w-full max-w-4xl px-6 py-10">
      <div className="mb-8">
        <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-[var(--signal-amber)] mb-4">
          Providers
        </p>
        <h1 className="text-2xl font-semibold text-[var(--ink-100)] mb-2">
          All monitored providers
        </h1>
        <p className="text-sm text-[var(--ink-400)]">
          Real API calls from 7 global locations. Updated every 30 s.
        </p>
      </div>

      {providers === null ? (
        <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-6 py-10 text-center">
          <p className="text-sm text-[var(--ink-400)]">
            Could not reach the API. Check that the backend is running.
          </p>
        </div>
      ) : (
        <ProvidersClient providers={providers} />
      )}
    </main>
  );
}
