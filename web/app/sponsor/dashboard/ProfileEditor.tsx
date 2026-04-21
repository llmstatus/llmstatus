"use client";

import { useState } from "react";

interface Sponsor {
  id: string;
  name: string;
  website_url: string | null;
  logo_url: string | null;
  tier: string;
  active: boolean;
}

interface Props {
  sponsor: Sponsor;
  apiToken: string;
}

export default function ProfileEditor({ sponsor, apiToken }: Props) {
  const [name, setName] = useState(sponsor.name);
  const [website, setWebsite] = useState(sponsor.website_url ?? "");
  const [logo, setLogo] = useState(sponsor.logo_url ?? "");
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);

  async function save(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setSaved(false);
    setSaving(true);
    const res = await fetch("/api/sponsor/me", {
      method: "PATCH",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${apiToken}` },
      body: JSON.stringify({ name, website_url: website, logo_url: logo }),
    });
    setSaving(false);
    if (!res.ok) {
      const { error: e } = await res.json().catch(() => ({ error: "failed" }));
      setError(e ?? "failed");
      return;
    }
    setSaved(true);
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-sm font-semibold uppercase tracking-wide text-[var(--ink-300)]">Profile</h2>
        <div className="flex items-center gap-2 text-xs text-[var(--ink-500)]">
          <span className="font-mono">/{sponsor.id}</span>
          <span>·</span>
          <span className={sponsor.active ? "text-[var(--signal-up)]" : "text-[var(--signal-down)]"}>
            {sponsor.active ? "active" : "inactive"}
          </span>
          <span>·</span>
          <span>{sponsor.tier}</span>
        </div>
      </div>

      <form onSubmit={save} className="flex flex-col gap-3">
        <label className="block">
          <span className="text-xs font-medium text-[var(--ink-300)]">Display name</span>
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            className="mt-1 block w-full bg-transparent border border-[var(--ink-600)] rounded px-2 py-1 text-sm text-[var(--ink-100)] focus:outline-none focus:border-[var(--ink-400)]"
          />
        </label>
        <label className="block">
          <span className="text-xs font-medium text-[var(--ink-300)]">Website URL</span>
          <input
            type="url"
            value={website}
            onChange={(e) => setWebsite(e.target.value)}
            placeholder="https://example.com"
            className="mt-1 block w-full bg-transparent border border-[var(--ink-600)] rounded px-2 py-1 text-sm text-[var(--ink-100)] focus:outline-none focus:border-[var(--ink-400)]"
          />
        </label>
        <label className="block">
          <span className="text-xs font-medium text-[var(--ink-300)]">Logo URL</span>
          <input
            type="url"
            value={logo}
            onChange={(e) => setLogo(e.target.value)}
            placeholder="https://example.com/logo.png"
            className="mt-1 block w-full bg-transparent border border-[var(--ink-600)] rounded px-2 py-1 text-sm text-[var(--ink-100)] focus:outline-none focus:border-[var(--ink-400)]"
          />
        </label>
        {error && <p className="text-xs text-[var(--signal-down)]">{error}</p>}
        {saved && <p className="text-xs text-[var(--signal-up)]">Saved.</p>}
        <button
          type="submit"
          disabled={saving}
          className="self-start rounded bg-[var(--signal-amber)] px-4 py-1.5 text-xs font-semibold text-[var(--canvas)] hover:opacity-90 disabled:opacity-50"
        >
          {saving ? "Saving…" : "Save changes"}
        </button>
      </form>
    </div>
  );
}
