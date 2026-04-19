interface Props {
  data: number[];   // 60 avg_ms values; 0 means no data
  width?: number;
  height?: number;
  className?: string; // applied to the <svg> element; use "w-full" to fill container
}

// buildPath converts 60 avg_ms values into an SVG path string.
// Zero values are treated as gaps (M move instead of L line).
function buildPath(data: number[], w: number, h: number): string {
  const valid = data.filter((v) => v > 0);
  if (valid.length === 0) return "";

  const max = Math.max(...valid);
  const pad = 2; // vertical padding in px

  let d = "";
  let prevHadData = false;

  for (let i = 0; i < data.length; i++) {
    const x = ((i / (data.length - 1)) * w).toFixed(1);
    if (data[i] <= 0) {
      prevHadData = false;
      continue;
    }
    // Invert Y: higher latency = lower on the chart (nearer the bottom).
    const y = (h - pad - ((data[i] / max) * (h - pad * 2))).toFixed(1);
    d += prevHadData ? `L${x} ${y}` : `M${x} ${y}`;
    prevHadData = true;
  }
  return d;
}

export function LatencySparkline({ data, width = 88, height = 28, className }: Props) {
  const d = buildPath(data, width, height);
  if (!d) {
    return <span className="text-[10px] text-[var(--ink-500)]">no data</span>;
  }

  return (
    <svg
      width={className ? undefined : width}
      height={height}
      viewBox={`0 0 ${width} ${height}`}
      preserveAspectRatio="none"
      className={className}
      aria-hidden="true"
    >
      <path
        d={d}
        fill="none"
        stroke="var(--viz-1)"
        strokeWidth="1.5"
        strokeLinejoin="round"
        strokeLinecap="round"
        opacity="0.85"
      />
    </svg>
  );
}
