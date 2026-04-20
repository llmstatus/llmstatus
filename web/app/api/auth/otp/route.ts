import { NextRequest, NextResponse } from "next/server";
import { setSessionCookie } from "@/lib/session";

const API = process.env.API_URL ?? "http://localhost:8081";

// POST /api/auth/otp?action=send|verify
export async function POST(req: NextRequest) {
  const action = req.nextUrl.searchParams.get("action");
  const body = await req.json();

  if (action === "send") {
    const res = await fetch(`${API}/auth/otp/send`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: body.email }),
    });
    return res.status === 204
      ? new NextResponse(null, { status: 204 })
      : NextResponse.json({ error: "failed to send code" }, { status: 502 });
  }

  if (action === "verify") {
    const res = await fetch(`${API}/auth/otp/verify`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: body.email, code: body.code }),
    });
    if (!res.ok) {
      return NextResponse.json({ error: "invalid code" }, { status: 401 });
    }
    const { token } = await res.json();
    await setSessionCookie(token);
    return NextResponse.json({ ok: true });
  }

  return NextResponse.json({ error: "unknown action" }, { status: 400 });
}
