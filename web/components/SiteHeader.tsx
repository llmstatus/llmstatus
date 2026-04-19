import Link from "next/link";
import { NavLink } from "./NavLink";

export function SiteHeader() {
  return (
    <header className="border-b border-[var(--ink-600)] px-6 py-4">
      <div className="mx-auto max-w-4xl flex items-center justify-between">
        <Link
          href="/"
          className="font-mono text-sm font-medium hover:opacity-80 transition-opacity"
          aria-label="llmstatus home"
        >
          <span className="text-[var(--signal-amber)]">[</span>
          <span className="text-[var(--ink-100)] mx-1.5">llmstatus</span>
          <span className="text-[var(--signal-amber)]">]</span>
        </Link>
        <nav className="flex items-center gap-6" aria-label="Site navigation">
          <NavLink href="/providers">Providers</NavLink>
          <NavLink href="/incidents">Incidents</NavLink>
          <NavLink href="/compare">Compare</NavLink>
          <NavLink href="/badges">Badges</NavLink>
          <NavLink href="/api">API</NavLink>
        </nav>
      </div>
    </header>
  );
}
