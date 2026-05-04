"use client";

import { useState, useTransition } from "react";
import { postReport } from "@/app/providers/[id]/actions";

const COOLDOWN_MS = 5 * 60 * 1000;

function storageKey(id: string) {
  return `llms_report_${id}`;
}

function isCoolingDown(id: string): boolean {
  if (typeof window === "undefined") return false;
  const raw = localStorage.getItem(storageKey(id));
  if (!raw) return false;
  return Date.now() - Number(raw) < COOLDOWN_MS;
}

interface Props {
  providerId: string;
  providerName: string;
}

export function ReportButton({ providerId, providerName }: Props) {
  const [reported, setReported] = useState(() => isCoolingDown(providerId));
  const [pending, startTransition] = useTransition();

  function handleClick() {
    if (reported || pending) return;
    startTransition(async () => {
      await postReport(providerId);
      localStorage.setItem(storageKey(providerId), String(Date.now()));
      setReported(true);
    });
  }

  const xText = encodeURIComponent(
    `Having issues with ${providerName} API? Report and discuss at llmstatus.io`,
  );
  const xUrl = encodeURIComponent(`https://llmstatus.io/providers/${providerId}`);
  const xHref = `https://twitter.com/intent/tweet?text=${xText}&url=${xUrl}`;

  return (
    <div className="flex items-center gap-3 flex-wrap">
      <button
        onClick={handleClick}
        disabled={reported || pending}
        className={[
          "flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium transition-colors",
          reported || pending
            ? "cursor-default bg-[var(--ink-700)] text-[var(--ink-500)]"
            : "bg-[var(--signal-down)]/15 text-[var(--signal-down)] hover:bg-[var(--signal-down)]/25",
        ].join(" ")}
      >
        <span className="inline-block h-1.5 w-1.5 rounded-full bg-current" />
        {reported ? "Reported — thanks" : pending ? "Reporting…" : "Report an issue"}
      </button>

      <a
        href={xHref}
        target="_blank"
        rel="noopener noreferrer"
        className="text-xs text-[var(--ink-400)] hover:text-[var(--ink-200)] transition-colors"
      >
        Discuss on X →
      </a>
    </div>
  );
}
