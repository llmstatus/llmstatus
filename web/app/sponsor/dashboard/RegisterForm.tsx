"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

interface Props {
  apiToken: string;
}

export default function RegisterForm({ apiToken }: Props) {
  const router = useRouter();
  const [id, setId] = useState("");
  const [name, setName] = useState("");
  const [website, setWebsite] = useState("");
  const [logo, setLogo] = useState("");
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setSaving(true);
    const res = await fetch("/api/sponsor/register", {
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${apiToken}` },
      body: JSON.stringify({ id, name, website_url: website, logo_url: logo }),
    });
    setSaving(false);
    if (!res.ok) {
      const { error: e } = await res.json().catch(() => ({ error: "failed" }));
      setError(e ?? "failed");
      return;
    }
    router.refresh();
  }

  return (
    <form onSubmit={submit} className="flex flex-col gap-4">
      <Field label="Sponsor ID" hint="2–40 lowercase letters, digits, or hyphens (permanent)">
        <input
          value={id}
          onChange={(e) => setId(e.target.value.toLowerCase())}
          placeholder="my-company"
          required
          className="input-field"
        />
      </Field>
      <Field label="Display name">
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="My Company"
          required
          className="input-field"
        />
      </Field>
      <Field label="Website URL" hint="Optional">
        <input
          type="url"
          value={website}
          onChange={(e) => setWebsite(e.target.value)}
          placeholder="https://example.com"
          className="input-field"
        />
      </Field>
      <Field label="Logo URL" hint="Optional — publicly accessible PNG or SVG">
        <input
          type="url"
          value={logo}
          onChange={(e) => setLogo(e.target.value)}
          placeholder="https://example.com/logo.png"
          className="input-field"
        />
      </Field>
      {error && <p className="text-xs text-[var(--signal-down)]">{error}</p>}
      <button
        type="submit"
        disabled={saving}
        className="self-start rounded bg-[var(--signal-amber)] px-4 py-2 text-sm font-semibold text-[var(--canvas)] hover:opacity-90 disabled:opacity-50"
      >
        {saving ? "Registering…" : "Register"}
      </button>
    </form>
  );
}

function Field({ label, hint, children }: { label: string; hint?: string; children: React.ReactNode }) {
  return (
    <div>
      <label className="block text-xs font-medium text-[var(--ink-300)] mb-1">
        {label} {hint && <span className="text-[var(--ink-500)] font-normal">— {hint}</span>}
      </label>
      {children}
    </div>
  );
}
