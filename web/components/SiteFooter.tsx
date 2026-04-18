interface Props {
  message?: string;
}

const DEFAULT_MESSAGE =
  "Data sourced from real API calls, not status pages. Updated every 30 s.";

export function SiteFooter({ message = DEFAULT_MESSAGE }: Props) {
  return (
    <footer className="border-t border-[var(--ink-600)] px-6 py-4">
      <div className="mx-auto max-w-4xl text-xs text-[var(--ink-400)]">
        {message}
      </div>
    </footer>
  );
}
