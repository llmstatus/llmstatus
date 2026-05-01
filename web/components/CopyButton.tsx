"use client";

import { useState } from "react";

interface CopyButtonProps {
  text: string;
}

export function CopyButton({ text }: CopyButtonProps) {
  const [copied, setCopied] = useState(false);

  async function handleCopy() {
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 1800);
  }

  return (
    <button
      onClick={handleCopy}
      className="text-xs font-semibold uppercase tracking-[0.08em] px-2 py-0.5 rounded border border-[var(--ink-600)] text-[var(--ink-300)] hover:text-[var(--ink-100)] hover:border-[var(--ink-500)] transition-colors"
    >
      {copied ? "Copied" : "Copy"}
    </button>
  );
}
