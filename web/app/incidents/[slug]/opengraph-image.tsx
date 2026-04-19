import { ImageResponse } from "next/og";
import { getIncident } from "@/lib/api";

export const runtime = "edge";
export const alt = "Incident details";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

const SEVERITY_COLOR: Record<string, string> = {
  critical: "#F87171",
  major:    "#F59E0B",
  minor:    "#94A3B8",
};

const STATUS_COLOR: Record<string, string> = {
  ongoing:    "#F87171",
  monitoring: "#F59E0B",
  resolved:   "#5EEAD4",
};

const STATUS_BG: Record<string, string> = {
  ongoing:    "#3A1515",
  monitoring: "#3D2909",
  resolved:   "#0F3530",
};

export default async function Image({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;

  let title = slug;
  let provider = "";
  let severity = "minor";
  let status = "resolved";

  try {
    const inc = await getIncident(slug);
    title = inc.title;
    provider = inc.provider_id;
    severity = inc.severity;
    status = inc.status;
  } catch {
    // Fall through with defaults if API unavailable.
  }

  const sevColor = SEVERITY_COLOR[severity] ?? SEVERITY_COLOR.minor;
  const statusColor = STATUS_COLOR[status] ?? STATUS_COLOR.resolved;
  const statusBg = STATUS_BG[status] ?? STATUS_BG.resolved;
  const statusLabel = status.charAt(0).toUpperCase() + status.slice(1);
  const sevLabel = severity.charAt(0).toUpperCase() + severity.slice(1);

  // Truncate long titles so they don't overflow the card.
  const displayTitle = title.length > 72 ? title.slice(0, 70) + "…" : title;

  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          alignItems: "flex-start",
          padding: "80px 96px",
          background: "#0A0E14",
          fontFamily: "sans-serif",
        }}
      >
        {/* Site label */}
        <div
          style={{
            fontSize: 18,
            fontWeight: 600,
            letterSpacing: "0.12em",
            textTransform: "uppercase",
            color: "#D97706",
            marginBottom: 24,
          }}
        >
          llmstatus.io
        </div>

        {/* Severity + provider row */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 16,
            marginBottom: 20,
          }}
        >
          <span
            style={{
              fontSize: 13,
              fontWeight: 700,
              letterSpacing: "0.1em",
              textTransform: "uppercase",
              color: sevColor,
            }}
          >
            {sevLabel}
          </span>
          {provider && (
            <>
              <span style={{ color: "#374151", fontSize: 13 }}>·</span>
              <span style={{ fontSize: 14, color: "#6B7280" }}>{provider}</span>
            </>
          )}
        </div>

        {/* Incident title */}
        <div
          style={{
            fontSize: 44,
            fontWeight: 700,
            color: "#E8ECF1",
            lineHeight: 1.2,
            marginBottom: 36,
          }}
        >
          {displayTitle}
        </div>

        {/* Status pill */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 10,
            background: statusBg,
            border: `1px solid ${statusColor}40`,
            borderRadius: 8,
            padding: "10px 20px",
          }}
        >
          <div
            style={{
              width: 10,
              height: 10,
              borderRadius: "50%",
              background: statusColor,
            }}
          />
          <span style={{ fontSize: 20, fontWeight: 600, color: statusColor }}>
            {statusLabel}
          </span>
        </div>

        {/* Tagline */}
        <div
          style={{
            position: "absolute",
            bottom: 80,
            left: 96,
            fontSize: 16,
            color: "#5A6472",
          }}
        >
          Independent real-time monitoring from 7 global locations
        </div>
      </div>
    ),
    { ...size },
  );
}
