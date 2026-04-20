import { NextRequest, NextResponse } from "next/server";
import { GitHub } from "arctic";
import { setSessionCookie } from "@/lib/session";

const API = process.env.API_URL ?? "http://localhost:8081";
const SITE = process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000";

function client() {
  const id = process.env.GITHUB_CLIENT_ID;
  const secret = process.env.GITHUB_CLIENT_SECRET;
  if (!id || !secret) return null;
  return new GitHub(id, secret, `${SITE}/api/auth/oauth/github?action=callback`);
}

// GET /api/auth/oauth/github?action=init
// GET /api/auth/oauth/github?action=callback&code=...&state=...
export async function GET(req: NextRequest) {
  const action = req.nextUrl.searchParams.get("action");
  const github = client();
  if (!github) return NextResponse.json({ error: "GitHub OAuth not configured" }, { status: 501 });

  if (action === "init") {
    const state = crypto.randomUUID();
    const url = github.createAuthorizationURL(state, ["user:email"]);
    const res = NextResponse.redirect(url);
    res.cookies.set("oauth_state", state, { httpOnly: true, sameSite: "lax", maxAge: 600 });
    return res;
  }

  if (action === "callback") {
    const storedState = req.cookies.get("oauth_state")?.value;
    const state = req.nextUrl.searchParams.get("state");
    const code = req.nextUrl.searchParams.get("code");
    if (!storedState || storedState !== state || !code) {
      return NextResponse.json({ error: "invalid state" }, { status: 400 });
    }

    const tokens = await github.validateAuthorizationCode(code);
    const accessToken = tokens.accessToken();

    // Fetch GitHub user
    const userRes = await fetch("https://api.github.com/user", {
      headers: { Authorization: `Bearer ${accessToken}`, Accept: "application/vnd.github+json" },
    });
    const ghUser = await userRes.json() as { id: number; email?: string };
    let emailAddr = ghUser.email ?? "";

    if (!emailAddr) {
      const emailRes = await fetch("https://api.github.com/user/emails", {
        headers: { Authorization: `Bearer ${accessToken}`, Accept: "application/vnd.github+json" },
      });
      const emails = await emailRes.json() as Array<{ email: string; primary: boolean }>;
      emailAddr = emails.find((e) => e.primary)?.email ?? "";
    }
    if (!emailAddr) return NextResponse.redirect(`${SITE}/login?error=1`);

    const internalSecret = process.env.INTERNAL_SECRET ?? "";
    const upsert = await fetch(`${API}/auth/oauth/upsert`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Internal-Token": internalSecret,
      },
      body: JSON.stringify({ provider: "github", sub: String(ghUser.id), email: emailAddr }),
    });
    if (!upsert.ok) return NextResponse.redirect(`${SITE}/login?error=1`);

    const { token } = await upsert.json();
    await setSessionCookie(token);
    const res = NextResponse.redirect(`${SITE}/account`);
    res.cookies.delete("oauth_state");
    return res;
  }

  return NextResponse.json({ error: "unknown action" }, { status: 400 });
}
