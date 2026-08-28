// Horizontal stacked bar showing the operational / degraded / down mix,
// with a color legend and counts. Server-component friendly (pure CSS).
interface Props {
  operational: number;
  degraded: number;
  down: number;
}

export function StatusDistributionBar({ operational, degraded, down }: Props) {
  const total = operational + degraded + down;
  if (total === 0) return null;

  const segments = [
    {
      key: "operational",
      count: operational,
      label: "Operational",
      cls: "bg-[var(--signal-ok)]",
    },
    {
      key: "degraded",
      count: degraded,
      label: "Degraded",
      cls: "bg-[var(--signal-warn)]",
    },
    { key: "down", count: down, label: "Down", cls: "bg-[var(--signal-down)]" },
  ].filter((s) => s.count > 0);

  return (
    <div>
      <div
        className="flex h-2 w-full overflow-hidden rounded-sm bg-[var(--canvas-overlay)]"
        role="img"
        aria-label={`${operational} operational, ${degraded} degraded, ${down} down`}
      >
        {segments.map((s, i) => (
          <div
            key={s.key}
            className={`bar-segment h-full ${s.cls}`}
            style={{
              width: `${(s.count / total) * 100}%`,
              animationDelay: `${i * 120}ms`,
            }}
          />
        ))}
      </div>
      <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1">
        {segments.map((s) => (
          <span
            key={s.key}
            className="inline-flex items-center gap-1.5 text-xs text-[var(--ink-400)]"
          >
            <span
              className={`h-2 w-2 rounded-full ${s.cls}`}
              aria-hidden="true"
            />
            {s.label} {s.count}
          </span>
        ))}
      </div>
    </div>
  );
}
