import { type NextRequest, NextResponse } from "next/server";

const API_URL = process.env.API_URL ?? "http://localhost:8081";

// Proxies /api/badge/{id} → Go API /badge/{id}.svg.
// Allows the Next.js app to serve badge previews without exposing
// the internal API_URL to the browser.
export async function GET(
  _req: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const upstream = `${API_URL}/badge/${encodeURIComponent(id)}.svg`;

  let res: Response;
  try {
    res = await fetch(upstream, { next: { revalidate: 60 } });
  } catch {
    return new NextResponse("upstream unavailable", { status: 502 });
  }

  const svg = await res.text();
  return new NextResponse(svg, {
    status: res.status,
    headers: {
      "Content-Type": "image/svg+xml; charset=utf-8",
      "Cache-Control": "public, max-age=30, s-maxage=30",
      "X-Content-Type-Options": "nosniff",
    },
  });
}
