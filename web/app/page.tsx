import { listProviders } from "@/lib/api";
import { ProviderTable } from "@/components/ProviderTable";

export const revalidate = 30;

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
    <main className="flex-1 mx-auto w-full max-w-4xl px-6 py-10">
      <div className="mb-8">
        <h1 className="text-2xl font-semibold text-[var(--ink-100)] mb-1">
          AI API Status
        </h1>
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
