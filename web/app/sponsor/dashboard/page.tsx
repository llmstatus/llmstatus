import { redirect } from "next/navigation";
import type { Metadata } from "next";
import Link from "next/link";
import { getSession } from "@/lib/session";
import KeyManager from "./KeyManager";
import ProfileEditor from "./ProfileEditor";
import RegisterForm from "./RegisterForm";

export const metadata: Metadata = { title: "Sponsor Dashboard — llmstatus.io" };

const API = process.env.API_URL ?? "http://localhost:8081";

interface Sponsor {
  id: string;
  name: string;
  website_url: string | null;
  logo_url: string | null;
  tier: string;
  active: boolean;
}

interface SponsorKey {
  provider_id: string;
  key_hint: string;
  active: boolean;
  last_verified_at: string | null;
  last_error: string | null;
}

async function fetchSponsorMe(token: string): Promise<{ sponsor: Sponsor; keys: SponsorKey[] } | null> {
  try {
    const res = await fetch(`${API}/v1/sponsor/me`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    if (res.status === 404) return null;
    if (!res.ok) return null;
    return await res.json();
  } catch {
    return null;
  }
}

async function fetchProviders(): Promise<{ id: string; name: string }[]> {
  try {
    const res = await fetch(`${API}/v1/providers`, { next: { revalidate: 300 } });
    if (!res.ok) return [];
    const { data } = await res.json();
    return (data ?? []).map((p: { id: string; name: string }) => ({ id: p.id, name: p.name }));
  } catch {
    return [];
  }
}

export default async function SponsorDashboardPage() {
  const session = await getSession();
  if (!session) redirect("/login?next=/sponsor/dashboard");

  const [me, providers] = await Promise.all([
    fetchSponsorMe(session.token),
    fetchProviders(),
  ]);

  return (
    <main className="flex-1 mx-auto w-full max-w-2xl px-6 py-10">
      <div className="flex items-center gap-3 mb-6">
        <Link
          href="/sponsors"
          className="text-xs text-[var(--ink-500)] hover:text-[var(--ink-200)] transition-colors"
        >
          ← Sponsors
        </Link>
        <h1 className="text-xl font-semibold text-[var(--ink-100)]">Sponsor Dashboard</h1>
      </div>

      {me === null ? (
        <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] p-6">
          <h2 className="mb-1 text-base font-semibold text-[var(--ink-100)]">Register as a sponsor</h2>
          <p className="mb-4 text-sm text-[var(--ink-400)]">
            Create your sponsor profile. Once approved, you&#39;ll appear on the sponsors page and
            can add API keys for providers you want to support.
          </p>
          <RegisterForm apiToken={session.token} />
        </div>
      ) : (
        <div className="flex flex-col gap-8">
          <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] p-6">
            <ProfileEditor sponsor={me.sponsor} apiToken={session.token} />
          </div>

          <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] p-6">
            <KeyManager keys={me.keys} providers={providers} apiToken={session.token} />
          </div>
        </div>
      )}
    </main>
  );
}
