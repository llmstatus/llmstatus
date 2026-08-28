// Abstract "live signal" hero visual — a radar-style sweep with pulsing
// blips. Pure inline SVG + CSS (BRAND_SYSTEM.md §6.7); no image assets,
// no animation libraries. Decorative only: `aria-hidden` in callers.
export function HeroSignal() {
  return (
    <div className="relative aspect-square w-full max-w-[340px]">
      {/* Ambient glow */}
      <div
        className="hero-glow absolute inset-0 rounded-full"
        aria-hidden="true"
      />

      <svg
        viewBox="0 0 400 400"
        className="relative h-full w-full"
        aria-hidden="true"
        role="presentation"
      >
        <defs>
          <linearGradient id="hero-sweep-gradient" x1="0" y1="0" x2="1" y2="1">
            <stop
              offset="0%"
              style={{ stopColor: "var(--signal-ok)", stopOpacity: 0.35 }}
            />
            <stop
              offset="100%"
              style={{ stopColor: "var(--signal-ok)", stopOpacity: 0 }}
            />
          </linearGradient>
        </defs>

        {/* Radar rings */}
        {[70, 130, 190].map((r) => (
          <circle
            key={r}
            cx="200"
            cy="200"
            r={r}
            fill="none"
            style={{ stroke: "var(--ink-600)" }}
            strokeWidth="1"
          />
        ))}

        {/* Crosshair */}
        <line
          x1="10"
          y1="200"
          x2="390"
          y2="200"
          style={{ stroke: "var(--ink-600)" }}
          strokeWidth="1"
        />
        <line
          x1="200"
          y1="10"
          x2="200"
          y2="390"
          style={{ stroke: "var(--ink-600)" }}
          strokeWidth="1"
        />

        {/* Rotating sweep */}
        <g className="hero-sweep" style={{ transformOrigin: "200px 200px" }}>
          <path
            d="M200 200 L200 10 A190 190 0 0 1 390 200 Z"
            fill="url(#hero-sweep-gradient)"
          />
          <line
            x1="200"
            y1="200"
            x2="200"
            y2="10"
            style={{ stroke: "var(--signal-ok)" }}
            strokeWidth="1.5"
          />
        </g>

        {/* Probe blips */}
        <circle
          className="hero-blip"
          cx="200"
          cy="70"
          r="3"
          style={{ fill: "var(--signal-ok)" }}
        />
        <circle
          className="hero-blip hero-blip--2"
          cx="318"
          cy="262"
          r="3"
          style={{ fill: "var(--signal-warn)" }}
        />
        <circle
          className="hero-blip hero-blip--3"
          cx="92"
          cy="304"
          r="3"
          style={{ fill: "var(--viz-2)" }}
        />
      </svg>
    </div>
  );
}
