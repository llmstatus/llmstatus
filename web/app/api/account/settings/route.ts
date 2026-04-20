import { NextRequest, NextResponse } from "next/server";

const API = process.env.API_URL ?? "http://localhost:8081";

export async function PUT(req: NextRequest) {
  const auth = req.headers.get("Authorization") ?? "";
  const body = await req.text();
  const res = await fetch(`${API}/account/settings`, {
    method: "PUT",
    headers: { "Content-Type": "application/json", Authorization: auth },
    body,
  });
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
