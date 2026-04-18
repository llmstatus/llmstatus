import type { Metadata } from "next";

import { listIncidents } from "@/lib/api";
import { IncidentCard } from "@/components/IncidentCard";

export const revalidate = 30;

export const metadata: Metadata = {
  title: "Incidents",
  description: "All detected incidents across AI API providers, past and present.",
};

export default async function IncidentsPage() {
  const incidents = await listIncidents("all", 50).catch(() => null);

  return (
    <main className="flex-1 mx-auto w-full max-w-4xl px-6 py-10">
      <div className="mb-8">
        <h1 className="text-2xl font-semibold text-[var(--ink-100)] mb-1">Incidents</h1>
        <p className="text-sm text-[var(--ink-400)]">
          All detected provider incidents, most recent first.
        </p>
      </div>

      {incidents === null ? (
        <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-6 py-10 text-center">
          <p className="text-sm text-[var(--ink-400)]">
            Could not reach the API. Check that the backend is running.
          </p>
        </div>
      ) : incidents.length === 0 ? (
        <p className="py-12 text-center text-sm text-[var(--ink-400)]">
          No incidents recorded yet.
        </p>
      ) : (
        <div className="flex flex-col gap-2">
          {incidents.map((inc) => (
            <IncidentCard
              key={inc.id}
              incident={inc}
              href={`/incidents/${inc.slug}`}
            />
          ))}
        </div>
      )}
    </main>
  );
}
