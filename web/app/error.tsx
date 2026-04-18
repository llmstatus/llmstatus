"use client";

export default function ErrorPage({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <main className="flex-1 flex flex-col items-center justify-center px-6 py-20 text-center">
      <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-[var(--signal-amber)] mb-4">
        Error
      </p>
      <h1 className="text-2xl font-semibold text-[var(--ink-100)] mb-3">
        Something went wrong.
      </h1>
      <p className="text-sm text-[var(--ink-400)] mb-8 max-w-xs">
        {error.digest ? (
          <>Could not load page data. Error ID: {error.digest}</>
        ) : (
          <>Could not load page data. Check that the backend is reachable.</>
        )}
      </p>
      <button
        onClick={reset}
        className="text-xs border border-[var(--ink-500)] text-[var(--ink-100)] px-4 py-2 hover:border-[var(--ink-300)] transition-colors"
      >
        Try again
      </button>
    </main>
  );
}
