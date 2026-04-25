"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

interface Props {
  href: string;
  children: React.ReactNode;
  variant?: "default" | "pill";
}

export function NavLink({ href, children, variant = "default" }: Props) {
  const pathname = usePathname();
  const active = pathname === href || (href !== "/" && pathname.startsWith(href));

  if (variant === "pill") {
    return (
      <Link
        href={href}
        className={`px-3 py-1 rounded border text-xs font-semibold tracking-wide transition-colors ${
          active
            ? "border-[var(--signal-ok)] text-[var(--signal-ok)] bg-[var(--signal-ok-bg)]"
            : "border-[var(--ink-500)] text-[var(--ink-300)] hover:border-[var(--signal-ok)] hover:text-[var(--signal-ok)]"
        }`}
      >
        {children}
      </Link>
    );
  }

  return (
    <Link
      href={href}
      className={`relative px-2.5 py-1.5 rounded text-sm transition-colors ${
        active
          ? "text-[var(--ink-100)] font-semibold"
          : "text-[var(--ink-400)] hover:text-[var(--ink-200)] hover:bg-[var(--canvas-overlay)]"
      }`}
    >
      {children}
      {active && (
        <span
          className="absolute bottom-0 left-2.5 right-2.5 h-[2px] rounded-full bg-[var(--signal-ok)]"
          aria-hidden="true"
        />
      )}
    </Link>
  );
}
