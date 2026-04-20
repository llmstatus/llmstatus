import { NextRequest, NextResponse } from "next/server";

const API = process.env.API_URL ?? "http://localhost:8081";

function forward(req: NextRequest, method: string, id: string, body?: BodyInit) {
  const auth = req.headers.get("Authorization") ?? "";
  return fetch(`${API}/account/subscriptions/${id}`, {
    method,
    headers: { "Content-Type": "application/json", Authorization: auth },
    body,
  });
}

export async function PUT(req: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const body = await req.text();
  const res = await forward(req, "PUT", id, body);
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}

export async function DELETE(req: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const res = await forward(req, "DELETE", id);
  return new NextResponse(null, { status: res.status });
}
