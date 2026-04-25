"use client";

import { useState } from "react";

interface PendingSponsor {
  id: string;
  name: string;
  website_url: string | null;
  logo_url: string | null;
  tier: string;
  user_id: number;
}

async function doAction(id: string, action: "approve" | "reject", token: string) {
  const res = await fetch(`/api/admin/sponsors/${id}?action=${action}`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) throw new Error(await res.text());
}

export default function SponsorReviewList({
  sponsors,
  apiToken,
}: {
  sponsors: PendingSponsor[];
  apiToken: string;
}) {
  const [list, setList] = useState(sponsors);
  const [busy, setBusy] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function handle(id: string, action: "approve" | "reject") {
    setBusy(id + action);
    setError(null);
    try {
      await doAction(id, action, apiToken);
      setList((prev) => prev.filter((s) => s.id !== id));
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(null);
    }
  }

  if (list.length === 0) {
    return (
      <p className="text-sm text-[var(--ink-500)]">No pending applications.</p>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      {error && (
        <p className="text-sm text-[var(--signal-down)]">{error}</p>
      )}
      {list.map((sp) => (
        <div
          key={sp.id}
          className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] p-4"
        >
          <div className="flex items-start justify-between gap-4">
            <div>
              <p className="text-sm font-semibold text-[var(--ink-100)]">{sp.name}</p>
              <p className="text-xs text-[var(--ink-500)] mt-0.5">id: {sp.id}</p>
              {sp.website_url && (
                <p className="text-xs text-[var(--ink-400)] mt-0.5">{sp.website_url}</p>
              )}
              {sp.logo_url && (
                <p className="text-xs text-[var(--ink-500)] mt-0.5">logo: {sp.logo_url}</p>
              )}
            </div>
            <div className="flex gap-2 shrink-0">
              <button
                disabled={busy !== null}
                onClick={() => handle(sp.id, "approve")}
                className="px-3 py-1 rounded text-xs font-medium bg-[var(--signal-up)] text-black hover:opacity-80 disabled:opacity-40 transition-opacity"
              >
                {busy === sp.id + "approve" ? "…" : "Approve"}
              </button>
              <button
                disabled={busy !== null}
                onClick={() => handle(sp.id, "reject")}
                className="px-3 py-1 rounded text-xs font-medium bg-[var(--signal-down)] text-white hover:opacity-80 disabled:opacity-40 transition-opacity"
              >
                {busy === sp.id + "reject" ? "…" : "Reject"}
              </button>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
