import { NextRequest, NextResponse } from "next/server";

const API = process.env.API_URL ?? "http://localhost:8081";

function auth(req: NextRequest) {
  return req.headers.get("Authorization") ?? "";
}

export async function PUT(
  req: NextRequest,
  { params }: { params: Promise<{ provider_id: string }> }
) {
  const { provider_id } = await params;
  const body = await req.text();
  const res = await fetch(`${API}/v1/sponsor/me/keys/${provider_id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json", Authorization: auth(req) },
    body,
  });
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}

export async function DELETE(
  req: NextRequest,
  { params }: { params: Promise<{ provider_id: string }> }
) {
  const { provider_id } = await params;
  const res = await fetch(`${API}/v1/sponsor/me/keys/${provider_id}`, {
    method: "DELETE",
    headers: { Authorization: auth(req) },
  });
  if (res.status === 204) return new NextResponse(null, { status: 204 });
  const data = await res.json().catch(() => ({}));
  return NextResponse.json(data, { status: res.status });
}
