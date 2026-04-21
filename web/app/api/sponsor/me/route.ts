import { NextRequest, NextResponse } from "next/server";

const API = process.env.API_URL ?? "http://localhost:8081";

function auth(req: NextRequest) {
  return req.headers.get("Authorization") ?? "";
}

export async function GET(req: NextRequest) {
  const res = await fetch(`${API}/v1/sponsor/me`, {
    headers: { Authorization: auth(req) },
    cache: "no-store",
  });
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}

export async function PATCH(req: NextRequest) {
  const body = await req.text();
  const res = await fetch(`${API}/v1/sponsor/me`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json", Authorization: auth(req) },
    body,
  });
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
