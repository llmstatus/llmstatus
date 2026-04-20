import { redirect } from "next/navigation";
import Link from "next/link";
import type { Metadata } from "next";
import { getSession } from "@/lib/session";
import SubscriptionsManager from "./SubscriptionsManager";
import DigestSettings from "./DigestSettings";
import type { Subscription } from "./SubscriptionsManager";

export const metadata: Metadata = { title: "Subscriptions — llmstatus.io" };

const API = process.env.API_URL ?? "http://localhost:8081";

async function fetchSubscriptions(token: string): Promise<Subscription[]> {
  try {
    const res = await fetch(`${API}/account/subscriptions`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    if (!res.ok) return [];
    const { data } = await res.json();
    return data ?? [];
  } catch {
    return [];
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

async function fetchMe(token: string): Promise<{ digest_hour: number; timezone: string }> {
  try {
    const res = await fetch(`${API}/auth/me`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    if (!res.ok) return { digest_hour: 8, timezone: "UTC" };
    const { data } = await res.json();
    return { digest_hour: data.digest_hour ?? 8, timezone: data.timezone ?? "UTC" };
  } catch {
    return { digest_hour: 8, timezone: "UTC" };
  }
}

export default async function SubscriptionsPage() {
  const session = await getSession();
  if (!session) redirect("/login");

  const [subscriptions, providers, me] = await Promise.all([
    fetchSubscriptions(session.token),
    fetchProviders(),
    fetchMe(session.token),
  ]);

  return (
    <main className="flex-1 mx-auto w-full max-w-2xl px-6 py-10">
      <div className="flex items-center gap-3 mb-6">
        <Link
          href="/account"
          className="text-xs text-[var(--ink-500)] hover:text-[var(--ink-200)] transition-colors"
        >
          ← Account
        </Link>
        <h1 className="text-xl font-semibold text-[var(--ink-100)]">Subscriptions &amp; alerts</h1>
      </div>

      <div className="flex flex-col gap-8">
        <SubscriptionsManager
          initial={subscriptions}
          providers={providers}
          apiToken={session.token}
        />

        <DigestSettings
          initialHour={me.digest_hour}
          initialTimezone={me.timezone}
          apiToken={session.token}
        />
      </div>
    </main>
  );
}
