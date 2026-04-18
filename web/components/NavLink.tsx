"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

interface Props {
  href: string;
  children: React.ReactNode;
}

export function NavLink({ href, children }: Props) {
  const pathname = usePathname();
  const active = pathname === href || (href !== "/" && pathname.startsWith(href));
  return (
    <Link
      href={href}
      className={`text-xs transition-colors ${
        active
          ? "text-[var(--ink-100)] font-medium"
          : "text-[var(--ink-400)] hover:text-[var(--ink-200)]"
      }`}
    >
      {children}
    </Link>
  );
}
