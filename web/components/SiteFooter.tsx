import Link from "next/link";

interface Props {
  message?: string;
}

const DEFAULT_MESSAGE =
  "Data sourced from real API calls, not status pages. Updated every 30 s.";

export function SiteFooter({ message = DEFAULT_MESSAGE }: Props) {
  return (
    <footer className="border-t border-[var(--ink-600)] px-6 py-4">
      <div className="mx-auto max-w-4xl flex items-center justify-between text-xs text-[var(--ink-400)]">
        <span>{message}</span>
        <Link
          href="/api/feed"
          className="hover:text-[var(--ink-200)] transition-colors"
          aria-label="RSS feed"
        >
          RSS
        </Link>
      </div>
    </footer>
  );
}
