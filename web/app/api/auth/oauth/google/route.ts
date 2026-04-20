import { NextRequest, NextResponse } from "next/server";
import { Google, decodeIdToken, generateCodeVerifier } from "arctic";
import { setSessionCookie } from "@/lib/session";

const API = process.env.API_URL ?? "http://localhost:8081";
const SITE = process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000";

function client() {
  const id = process.env.GOOGLE_CLIENT_ID;
  const secret = process.env.GOOGLE_CLIENT_SECRET;
  if (!id || !secret) return null;
  return new Google(id, secret, `${SITE}/api/auth/oauth/google?action=callback`);
}

// GET /api/auth/oauth/google?action=init  — redirect to Google
// GET /api/auth/oauth/google?action=callback&code=...&state=...
export async function GET(req: NextRequest) {
  const action = req.nextUrl.searchParams.get("action");
  const google = client();
  if (!google) return NextResponse.json({ error: "Google OAuth not configured" }, { status: 501 });

  if (action === "init") {
    const state = crypto.randomUUID();
    const codeVerifier = generateCodeVerifier();
    const url = google.createAuthorizationURL(state, codeVerifier, ["openid", "email"]);
    const res = NextResponse.redirect(url);
    res.cookies.set("oauth_state", state, { httpOnly: true, sameSite: "lax", maxAge: 600 });
    res.cookies.set("oauth_cv", codeVerifier, { httpOnly: true, sameSite: "lax", maxAge: 600 });
    return res;
  }

  if (action === "callback") {
    const storedState = req.cookies.get("oauth_state")?.value;
    const codeVerifier = req.cookies.get("oauth_cv")?.value;
    const state = req.nextUrl.searchParams.get("state");
    const code = req.nextUrl.searchParams.get("code");
    if (!storedState || storedState !== state || !code || !codeVerifier) {
      return NextResponse.json({ error: "invalid state" }, { status: 400 });
    }

    let tokens;
    try {
      tokens = await google.validateAuthorizationCode(code, codeVerifier);
    } catch {
      return NextResponse.redirect(`${SITE}/login?error=1`);
    }

    const idToken = tokens.idToken();
    const claims = decodeIdToken(idToken) as { sub?: unknown; email?: unknown };
    if (typeof claims.sub !== "string" || !claims.sub || typeof claims.email !== "string" || !claims.email) {
      return NextResponse.redirect(`${SITE}/login?error=1`);
    }

    const internalSecret = process.env.INTERNAL_SECRET ?? "";
    const upsert = await fetch(`${API}/auth/oauth/upsert`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Internal-Token": internalSecret,
      },
      body: JSON.stringify({ provider: "google", sub: claims.sub, email: claims.email }),
    });
    if (!upsert.ok) return NextResponse.redirect(`${SITE}/login?error=1`);

    const { token } = await upsert.json();
    await setSessionCookie(token);
    const res = NextResponse.redirect(`${SITE}/account`);
    res.cookies.delete("oauth_state");
    res.cookies.delete("oauth_cv");
    return res;
  }

  return NextResponse.json({ error: "unknown action" }, { status: 400 });
}
