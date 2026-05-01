import type { Metadata } from "next";
import Link from "next/link";

export const metadata: Metadata = {
  title: "Page not found",
};

export default function NotFound() {
  return (
    <main className="flex-1 flex flex-col items-center justify-center px-6 py-20 text-center">
      <p className="text-xs font-semibold uppercase tracking-[0.12em] text-[var(--signal-amber)] mb-4">
        404
      </p>
      <h1 className="text-2xl font-semibold text-[var(--ink-100)] mb-3">
        Page not found.
      </h1>
      <p className="text-sm text-[var(--ink-400)] mb-8 max-w-xs">
        This URL does not exist. If you followed a link, it may have moved.
      </p>
      <Link
        href="/"
        className="text-xs text-[var(--ink-300)] hover:text-[var(--ink-100)] transition-colors underline underline-offset-4"
      >
        ← Back to status
      </Link>
    </main>
  );
}
