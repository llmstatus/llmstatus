interface Props {
  data: number[];
  width?: number;
  height?: number;
  className?: string;
  area?: boolean;
  color?: string;
  gradientId?: string;
}

function buildLinePath(data: number[], w: number, h: number): string {
  const valid = data.filter((v) => v > 0);
  if (valid.length === 0) return "";

  const max = Math.max(...valid);
  const pad = 2;
  const divisor = data.length > 1 ? data.length - 1 : 1;
  let d = "";
  let prevHadData = false;

  for (let i = 0; i < data.length; i++) {
    const x = ((i / divisor) * w).toFixed(1);
    if (data[i] <= 0) { prevHadData = false; continue; }
    const y = (h - pad - ((data[i] / max) * (h - pad * 2))).toFixed(1);
    d += prevHadData ? `L${x} ${y}` : `M${x} ${y}`;
    prevHadData = true;
  }
  return d;
}

function buildAreaPaths(data: number[], w: number, h: number): string {
  const valid = data.filter((v) => v > 0);
  if (valid.length === 0) return "";

  const max = Math.max(...valid);
  const pad = 2;
  const divisor = data.length > 1 ? data.length - 1 : 1;

  const segments: { x: number; y: number }[][] = [];
  let cur: { x: number; y: number }[] = [];

  for (let i = 0; i < data.length; i++) {
    const x = (i / divisor) * w;
    if (data[i] <= 0) {
      if (cur.length > 0) { segments.push(cur); cur = []; }
      continue;
    }
    const y = h - pad - ((data[i] / max) * (h - pad * 2));
    cur.push({ x, y });
  }
  if (cur.length > 0) segments.push(cur);

  return segments
    .filter((s) => s.length >= 2)  // single-point segments produce no visible line or area
    .map((seg) => {
      const first = seg[0];
      const last = seg[seg.length - 1];
      let p = `M${first.x.toFixed(1)} ${h}L${first.x.toFixed(1)} ${first.y.toFixed(1)}`;
      for (let i = 1; i < seg.length; i++) {
        p += `L${seg[i].x.toFixed(1)} ${seg[i].y.toFixed(1)}`;
      }
      p += `L${last.x.toFixed(1)} ${h}Z`;
      return p;
    })
    .join(" ");
}

export function LatencySparkline({
  data,
  width = 88,
  height = 28,
  className,
  area = false,
  color = "var(--viz-1)",
  gradientId,
}: Props) {
  const lineD = buildLinePath(data, width, height);
  if (!lineD) {
    return <span className="text-xs text-[var(--ink-500)]">no data</span>;
  }

  const areaD = area ? buildAreaPaths(data, width, height) : "";
  // Derive a stable, unique gradient ID from the color string so multiple SVGs
  // on the same page don't collide (url(#id) is document-scoped, not SVG-scoped).
  const gid = gradientId ?? `sp-grad-${color.replace(/[^a-z0-9]/gi, "")}`;

  return (
    <svg
      width={className ? undefined : width}
      height={height}
      viewBox={`0 0 ${width} ${height}`}
      preserveAspectRatio="none"
      className={className}
      aria-hidden="true"
    >
      {area && (
        <defs>
          <linearGradient id={gid} x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor={color} stopOpacity="0.28" />
            <stop offset="100%" stopColor={color} stopOpacity="0" />
          </linearGradient>
        </defs>
      )}
      {area && areaD && (
        <path d={areaD} fill={`url(#${gid})`} />
      )}
      <path
        d={lineD}
        fill="none"
        stroke={color}
        strokeWidth="1.5"
        strokeLinejoin="round"
        strokeLinecap="round"
        opacity="0.9"
      />
    </svg>
  );
}
