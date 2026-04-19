"use client";
import { useRouter } from "next/navigation";
import type { ProviderSummary } from "@/lib/api";

interface Props {
  providers: ProviderSummary[];
  a: string;
  b: string;
}

const SELECT_CLS =
  "rounded border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-3 py-1.5 " +
  "text-sm text-[var(--ink-200)] focus:outline-none focus:ring-1 focus:ring-[var(--ink-400)]";

export function CompareSelector({ providers, a, b }: Props) {
  const router = useRouter();

  function navigate(nextA: string, nextB: string) {
    if (nextA && nextB) {
      router.push(`/compare?a=${encodeURIComponent(nextA)}&b=${encodeURIComponent(nextB)}`);
    }
  }

  return (
    <div className="flex flex-wrap items-center gap-3">
      <select
        className={SELECT_CLS}
        value={a}
        onChange={(e) => navigate(e.target.value, b)}
        aria-label="First provider"
      >
        <option value="">Select provider…</option>
        {providers.map((p) => (
          <option key={p.id} value={p.id}>{p.name}</option>
        ))}
      </select>
      <span className="text-sm text-[var(--ink-400)]">vs</span>
      <select
        className={SELECT_CLS}
        value={b}
        onChange={(e) => navigate(a, e.target.value)}
        aria-label="Second provider"
      >
        <option value="">Select provider…</option>
        {providers.map((p) => (
          <option key={p.id} value={p.id}>{p.name}</option>
        ))}
      </select>
    </div>
  );
}
