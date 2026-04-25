import { NextRequest, NextResponse } from "next/server";

const API = process.env.API_URL ?? "http://localhost:8081";

function auth(req: NextRequest) {
  return req.headers.get("Authorization") ?? "";
}

export async function POST(
  req: NextRequest,
  { params }: { params: Promise<{ id: string }> },
) {
  const { id } = await params;
  const action = req.nextUrl.searchParams.get("action");
  if (action !== "approve" && action !== "reject") {
    return NextResponse.json({ error: "invalid action" }, { status: 400 });
  }
  const res = await fetch(`${API}/v1/admin/sponsors/${id}/${action}`, {
    method: "POST",
    headers: { Authorization: auth(req) },
  });
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
