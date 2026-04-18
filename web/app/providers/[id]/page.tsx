import { notFound } from "next/navigation";
import type { Metadata } from "next";
import Link from "next/link";

import { getProvider, ApiNotFoundError } from "@/lib/api";
import { StatusPill } from "@/components/StatusPill";
import { IncidentCard } from "@/components/IncidentCard";
import { ModelList } from "@/components/ModelList";

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
    return {
      title: `Is ${provider.name} API down?`,
      description: `${provider.name} API is currently ${statusLabel}. Real-time uptime monitoring from llmstatus.io.`,
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

  return (
    <div className="flex flex-col min-h-screen">
      <header className="border-b border-[var(--ink-600)] px-6 py-4">
        <div className="mx-auto max-w-4xl flex items-center justify-between">
          <Link
            href="/"
            className="font-mono text-sm font-semibold tracking-widest text-[var(--signal-amber)] uppercase hover:opacity-80 transition-opacity"
          >
            llmstatus.io
          </Link>
          <span className="text-xs text-[var(--ink-400)]">
            Real-time AI API monitoring
          </span>
        </div>
      </header>

      <main className="flex-1 mx-auto w-full max-w-4xl px-6 py-10">
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

        {/* Models */}
        <section>
          <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-[var(--ink-300)]">
            Monitored Models
          </h2>
          <ModelList models={provider.models} />
        </section>
      </main>

      <footer className="border-t border-[var(--ink-600)] px-6 py-4">
        <div className="mx-auto max-w-4xl text-xs text-[var(--ink-400)]">
          Data sourced from real API calls, not status pages. Updated every 60 s.
        </div>
      </footer>
    </div>
  );
}
