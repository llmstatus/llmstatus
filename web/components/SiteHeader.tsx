import Link from "next/link";
import { NavLink } from "./NavLink";

export function SiteHeader() {
  return (
    <header className="border-b border-[var(--ink-600)] px-6 py-4">
      <div className="mx-auto max-w-4xl flex items-center justify-between">
        <Link
          href="/"
          className="font-mono text-sm font-semibold tracking-widest text-[var(--signal-amber)] uppercase hover:opacity-80 transition-opacity"
        >
          llmstatus.io
        </Link>
        <nav className="flex items-center gap-6" aria-label="Site navigation">
          <NavLink href="/">Providers</NavLink>
          <NavLink href="/incidents">Incidents</NavLink>
          <NavLink href="/badges">Badges</NavLink>
        </nav>
      </div>
    </header>
  );
}
