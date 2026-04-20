import { redirect } from "next/navigation";
import type { Metadata } from "next";
import Link from "next/link";
import { getSession } from "@/lib/session";

export const metadata: Metadata = { title: "Account — llmstatus.io" };

export default async function AccountPage() {
  const session = await getSession();
  if (!session) redirect("/login");

  return (
    <main className="flex-1 mx-auto w-full max-w-2xl px-6 py-10">
      <h1 className="mb-6 text-xl font-semibold text-[var(--ink-100)]">Account</h1>

      <div className="mb-8 rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] p-4">
        <p className="text-xs text-[var(--ink-500)] mb-1">Signed in as</p>
        <p className="text-sm text-[var(--ink-100)]">{session.email}</p>
      </div>

      <div className="mb-8">
        <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-[var(--ink-300)]">
          Manage
        </h2>
        <div className="flex flex-col gap-2">
          <Link
            href="/account/subscriptions"
            className="flex items-center justify-between rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-4 py-3 text-sm text-[var(--ink-200)] hover:border-[var(--ink-400)] transition-colors"
          >
            <span>Subscriptions &amp; alerts</span>
            <span className="text-[var(--ink-500)]">→</span>
          </Link>
        </div>
      </div>

      <form action="/api/auth/logout" method="POST">
        <button
          type="submit"
          className="text-xs text-[var(--ink-500)] hover:text-[var(--signal-down)] transition-colors"
        >
          Sign out
        </button>
      </form>
    </main>
  );
}
