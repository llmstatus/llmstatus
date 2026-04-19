import { type NextRequest, NextResponse } from "next/server";

const API_URL = process.env.API_URL ?? "http://localhost:8081";

// Proxies /api/feed → Go API /feed.xml (global incident RSS feed).
export async function GET(_req: NextRequest) {
  let res: Response;
  try {
    res = await fetch(`${API_URL}/feed.xml`, { next: { revalidate: 60 } });
  } catch {
    return new NextResponse("upstream unavailable", { status: 502 });
  }

  const xml = await res.text();
  return new NextResponse(xml, {
    status: res.status,
    headers: {
      "Content-Type": "application/rss+xml; charset=utf-8",
      "Cache-Control": "public, max-age=60, s-maxage=60",
    },
  });
}
