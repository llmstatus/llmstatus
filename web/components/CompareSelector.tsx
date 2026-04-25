"use client";
import { useState } from "react";
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
  const [selA, setSelA] = useState(a);
  const [selB, setSelB] = useState(b);

  function handleA(value: string) {
    setSelA(value);
    if (value && selB) {
      router.push(`/compare?a=${encodeURIComponent(value)}&b=${encodeURIComponent(selB)}`);
    }
  }

  function handleB(value: string) {
    setSelB(value);
    if (selA && value) {
      router.push(`/compare?a=${encodeURIComponent(selA)}&b=${encodeURIComponent(value)}`);
    }
  }

  return (
    <div className="flex flex-wrap items-center gap-3">
      <select
        className={SELECT_CLS}
        value={selA}
        onChange={(e) => handleA(e.target.value)}
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
        value={selB}
        onChange={(e) => handleB(e.target.value)}
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
