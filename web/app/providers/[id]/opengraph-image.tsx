import { ImageResponse } from "next/og";
import { getProvider } from "@/lib/api";

export const runtime = "edge";
export const alt = "Provider status";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

const STATUS_COLOR = {
  operational: "#5EEAD4",
  degraded:    "#F59E0B",
  down:        "#F87171",
};

const STATUS_BG = {
  operational: "#0F3530",
  degraded:    "#3D2909",
  down:        "#3A1515",
};

export default async function Image({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  let name = id;
  let status: "operational" | "degraded" | "down" = "operational";

  try {
    const provider = await getProvider(id);
    name = provider.name;
    status = provider.current_status;
  } catch {
    // Fall through with defaults if API unavailable.
  }

  const color = STATUS_COLOR[status];
  const bg = STATUS_BG[status];
  const label = status.charAt(0).toUpperCase() + status.slice(1);

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
            marginBottom: 32,
          }}
        >
          llmstatus.io
        </div>

        {/* Provider name */}
        <div
          style={{
            fontSize: 64,
            fontWeight: 700,
            color: "#E8ECF1",
            lineHeight: 1.1,
            marginBottom: 32,
          }}
        >
          {name}
        </div>

        {/* Status pill */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 12,
            background: bg,
            border: `1px solid ${color}40`,
            borderRadius: 8,
            padding: "12px 24px",
          }}
        >
          <div
            style={{
              width: 12,
              height: 12,
              borderRadius: "50%",
              background: color,
            }}
          />
          <span style={{ fontSize: 24, fontWeight: 600, color }}>
            {label}
          </span>
        </div>

        {/* Tagline */}
        <div
          style={{
            position: "absolute",
            bottom: 80,
            left: 96,
            fontSize: 18,
            color: "#5A6472",
          }}
        >
          Independent real-time monitoring from 7 global locations
        </div>
      </div>
    ),
    {
      ...size,
    }
  );
}
