import { NextRequest, NextResponse } from "next/server";

const API = process.env.API_URL ?? "http://localhost:8081";

function auth(req: NextRequest) {
  return req.headers.get("Authorization") ?? "";
}

export async function GET(req: NextRequest) {
  const res = await fetch(`${API}/v1/admin/sponsors`, {
    headers: { Authorization: auth(req) },
    cache: "no-store",
  });
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
