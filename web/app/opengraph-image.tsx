import { ImageResponse } from "next/og";

export const runtime = "edge";
export const alt = "llmstatus.io — AI API Status Monitor";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default function Image() {
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
        {/* Dot grid background */}
        <div
          style={{
            position: "absolute",
            inset: 0,
            backgroundImage:
              "radial-gradient(circle, rgba(255,255,255,0.04) 1px, transparent 1px)",
            backgroundSize: "32px 32px",
          }}
        />

        {/* Site name */}
        <div
          style={{
            fontSize: 20,
            fontWeight: 700,
            letterSpacing: "0.15em",
            textTransform: "uppercase",
            color: "#D97706",
            marginBottom: 32,
          }}
        >
          llmstatus.io
        </div>

        {/* Headline */}
        <div
          style={{
            fontSize: 56,
            fontWeight: 700,
            color: "#E8ECF1",
            lineHeight: 1.15,
            marginBottom: 24,
          }}
        >
          Independent real-time
          <br />
          AI API monitoring.
        </div>

        {/* Sub-line */}
        <div
          style={{
            fontSize: 22,
            color: "#6B7280",
            lineHeight: 1.5,
          }}
        >
          Measured from 7 global locations.
          <br />
          Not scraped from official status pages.
        </div>

        {/* Provider count badge */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 10,
            marginTop: 40,
            background: "#111827",
            border: "1px solid #1F2937",
            borderRadius: 8,
            padding: "12px 24px",
          }}
        >
          <div
            style={{
              width: 10,
              height: 10,
              borderRadius: "50%",
              background: "#5EEAD4",
            }}
          />
          <span style={{ fontSize: 18, fontWeight: 600, color: "#5EEAD4" }}>
            20+ AI providers tracked
          </span>
        </div>
      </div>
    ),
    { ...size },
  );
}
