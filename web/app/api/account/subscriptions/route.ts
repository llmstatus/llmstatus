import { NextRequest, NextResponse } from "next/server";

const API = process.env.API_URL ?? "http://localhost:8081";

function forward(req: NextRequest, method: string, body?: BodyInit) {
  const auth = req.headers.get("Authorization") ?? "";
  return fetch(`${API}/account/subscriptions`, {
    method,
    headers: { "Content-Type": "application/json", Authorization: auth },
    body,
  });
}

export async function GET(req: NextRequest) {
  const res = await forward(req, "GET");
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}

export async function POST(req: NextRequest) {
  const body = await req.text();
  const res = await forward(req, "POST", body);
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
