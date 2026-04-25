import Link from "next/link";
import { NavLink } from "./NavLink";
import { getSession } from "@/lib/session";

export async function SiteHeader() {
  const session = await getSession().catch(() => null);

  return (
    <header className="sticky top-0 z-50 border-b border-[var(--ink-600)] bg-[var(--canvas-raised)] px-6">
      <div className="mx-auto max-w-4xl flex items-center justify-between h-14">

        {/* Logo — pulsing heartbeat dot + wordmark */}
        <Link
          href="/"
          className="flex items-center gap-2.5 group shrink-0"
          aria-label="llmstatus home"
        >
          <span className="relative flex h-2.5 w-2.5 shrink-0">
            <span className="beat-ring absolute inset-0 rounded-full bg-[var(--signal-ok)]" />
            <span className="beat relative inline-flex h-2.5 w-2.5 rounded-full bg-[var(--signal-ok)]" />
          </span>
          <span className="font-mono font-bold text-base text-[var(--ink-100)] tracking-tight leading-none">
            llmstatus
            <span className="text-[var(--ink-500)] font-normal text-sm">.io</span>
          </span>
        </Link>

        {/* Navigation — grouped */}
        <nav className="flex items-center gap-1" aria-label="Site navigation">
          {/* Primary */}
          <NavLink href="/providers">Providers</NavLink>
          <NavLink href="/incidents">Incidents</NavLink>
          <NavLink href="/china">China</NavLink>
          <NavLink href="/compare">Compare</NavLink>

          {/* Separator */}
          <span className="mx-1 h-4 w-px bg-[var(--ink-600)]" aria-hidden="true" />

          {/* Secondary */}
          <NavLink href="/api">API</NavLink>
          <NavLink href="/badges">Badges</NavLink>
          <NavLink href="/sponsors">Sponsors</NavLink>

          {/* Separator */}
          <span className="mx-1 h-4 w-px bg-[var(--ink-600)]" aria-hidden="true" />

          {/* Auth */}
          {session ? (
            <NavLink href="/account" variant="pill">Account</NavLink>
          ) : (
            <NavLink href="/login" variant="pill">Sign in</NavLink>
          )}
        </nav>
      </div>
    </header>
  );
}
