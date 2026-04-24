"use client";

import { useEffect, useState } from "react";

function relativeTime(iso: string): string {
  const diffMs = Date.now() - new Date(iso).getTime();
  const diffS = Math.floor(diffMs / 1000);

  if (diffS < 60) return `${diffS}s ago`;
  const diffM = Math.floor(diffS / 60);
  if (diffM < 60) return `${diffM}m ago`;
  const diffH = Math.floor(diffM / 60);
  const remM = diffM % 60;
  if (diffH < 24) return remM > 0 ? `${diffH}h ${remM}m ago` : `${diffH}h ago`;
  const diffD = Math.floor(diffH / 24);
  return `${diffD}d ago`;
}

interface Props {
  /** ISO 8601 timestamp string */
  iso: string;
  /** Prefix shown before the relative time, e.g. "Started" */
  prefix?: string;
}

// Brand spec §6.4: relative "X ago" display, auto-refreshes every 10 s.
export function ProbeTimestamp({ iso, prefix }: Props) {
  const [label, setLabel] = useState(() => relativeTime(iso));

  useEffect(() => {
    // Sync label to new `iso` prop on change, then poll every 10 s.
    // The synchronous setState is intentional: it avoids a one-tick stale display.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setLabel(relativeTime(iso));
    const id = setInterval(() => setLabel(relativeTime(iso)), 10_000);
    return () => clearInterval(id);
  }, [iso]);

  return (
    <time
      dateTime={iso}
      title={new Date(iso).toUTCString()}
      className="text-[11px] text-[var(--ink-400)]"
    >
      {prefix ? `${prefix} ${label}` : label}
    </time>
  );
}
