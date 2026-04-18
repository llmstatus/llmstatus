export default function LoadingPage() {
  return (
    <main className="flex-1 mx-auto w-full max-w-4xl px-6 py-10 animate-pulse">
      {/* Hero skeleton */}
      <div className="py-14 mb-2">
        <div className="h-2.5 w-24 rounded bg-[var(--ink-600)] mb-4" />
        <div className="h-8 w-80 rounded bg-[var(--ink-600)] mb-2" />
        <div className="h-8 w-56 rounded bg-[var(--ink-600)] mb-4" />
        <div className="h-3 w-64 rounded bg-[var(--ink-700)] mb-1" />
        <div className="h-3 w-52 rounded bg-[var(--ink-700)]" />
      </div>

      {/* Status summary skeleton */}
      <div className="mb-6">
        <div className="h-3 w-40 rounded bg-[var(--ink-600)]" />
      </div>

      {/* Table skeleton */}
      <div className="overflow-hidden rounded-lg border border-[var(--ink-600)]">
        {/* Header row */}
        <div className="flex gap-4 px-4 py-3 border-b border-[var(--ink-600)] bg-[var(--canvas-sunken)]">
          {[40, 24, 16, 20, 12, 16].map((w, i) => (
            <div key={i} className={`h-2 w-${w} rounded bg-[var(--ink-700)]`} />
          ))}
        </div>
        {/* Data rows */}
        {Array.from({ length: 6 }).map((_, i) => (
          <div
            key={i}
            className="flex items-center gap-4 px-4 py-3 border-b border-[var(--ink-600)] last:border-0"
          >
            <div className="h-3 w-32 rounded bg-[var(--ink-600)]" />
            <div className="h-3 w-20 rounded bg-[var(--ink-700)]" />
            <div className="h-3 w-12 rounded bg-[var(--ink-700)]" />
            <div className="h-3 w-16 rounded bg-[var(--ink-700)]" />
            <div className="h-3 w-10 rounded bg-[var(--ink-700)]" />
            <div className="h-2 w-2 rounded-full bg-[var(--ink-600)] ml-auto" />
          </div>
        ))}
      </div>
    </main>
  );
}
