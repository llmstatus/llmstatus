import type { ReportHistogramBucket } from "@/lib/api";

interface Props {
  buckets: ReportHistogramBucket[];
}

export function UserReportHistogram({ buckets }: Props) {
  const max = Math.max(...buckets.map((b) => b.count), 1);
  const total = buckets.reduce((s, b) => s + b.count, 0);

  return (
    <div>
      <div className="mb-1.5 flex items-center justify-between text-[10px] text-[var(--ink-500)]">
        <span>Community reports — last 24 hours</span>
        <span>{total} total</span>
      </div>
      <div className="flex h-10 items-end gap-px">
        {buckets.map((b) => {
          const pct = Math.round((b.count / max) * 100);
          const hour = new Date(b.hour).toLocaleTimeString("en-US", {
            hour: "2-digit",
            minute: "2-digit",
            hour12: false,
            timeZone: "UTC",
          });
          return (
            <div
              key={b.hour}
              title={`${hour} UTC — ${b.count} report${b.count !== 1 ? "s" : ""}`}
              className="group relative flex flex-1 flex-col justify-end"
            >
              <div
                className={[
                  "w-full rounded-sm transition-colors",
                  b.count === 0
                    ? "bg-[var(--ink-700)]"
                    : b.count >= max * 0.6
                    ? "bg-[var(--signal-down)]"
                    : "bg-[var(--signal-warn)]",
                ].join(" ")}
                style={{ height: `${Math.max(pct, 4)}%` }}
              />
            </div>
          );
        })}
      </div>
      <div className="mt-1 flex justify-between text-[9px] text-[var(--ink-600)]">
        <span>24h ago</span>
        <span>now</span>
      </div>
    </div>
  );
}
