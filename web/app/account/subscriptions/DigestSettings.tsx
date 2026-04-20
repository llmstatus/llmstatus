"use client";

import { useState } from "react";

// Common IANA timezones — abbreviated list for V1.
const TIMEZONES = [
  "UTC",
  "America/New_York",
  "America/Chicago",
  "America/Denver",
  "America/Los_Angeles",
  "America/Sao_Paulo",
  "Europe/London",
  "Europe/Paris",
  "Europe/Berlin",
  "Europe/Moscow",
  "Asia/Dubai",
  "Asia/Kolkata",
  "Asia/Shanghai",
  "Asia/Tokyo",
  "Asia/Seoul",
  "Australia/Sydney",
  "Pacific/Auckland",
];

interface Props {
  initialHour: number;
  initialTimezone: string;
  apiToken: string;
}

export default function DigestSettings({ initialHour, initialTimezone, apiToken }: Props) {
  const [hour, setHour] = useState(initialHour);
  const [tz, setTz] = useState(initialTimezone);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState("");

  async function save() {
    setSaving(true);
    setError("");
    setSaved(false);
    try {
      const res = await fetch("/api/account/settings", {
        method: "PUT",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${apiToken}` },
        body: JSON.stringify({ digest_hour: hour, timezone: tz }),
      });
      if (!res.ok) {
        const { error: e } = await res.json().catch(() => ({ error: "failed" }));
        setError(e ?? "failed");
      } else {
        setSaved(true);
        setTimeout(() => setSaved(false), 2000);
      }
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] p-4">
      <h2 className="text-sm font-semibold uppercase tracking-wide text-[var(--ink-300)] mb-4">
        Digest settings
      </h2>
      <div className="flex flex-wrap gap-4 items-end">
        <div className="flex flex-col gap-1">
          <label className="text-xs text-[var(--ink-400)]">Send at hour (local)</label>
          <select
            value={hour}
            onChange={(e) => setHour(Number(e.target.value))}
            className="bg-transparent border border-[var(--ink-600)] rounded px-2 py-1 text-sm text-[var(--ink-100)] focus:outline-none"
          >
            {Array.from({ length: 24 }, (_, i) => (
              <option key={i} value={i}>
                {String(i).padStart(2, "0")}:00
              </option>
            ))}
          </select>
        </div>

        <div className="flex flex-col gap-1">
          <label className="text-xs text-[var(--ink-400)]">Timezone</label>
          <select
            value={tz}
            onChange={(e) => setTz(e.target.value)}
            className="bg-transparent border border-[var(--ink-600)] rounded px-2 py-1 text-sm text-[var(--ink-100)] focus:outline-none"
          >
            {TIMEZONES.map((t) => (
              <option key={t} value={t}>{t}</option>
            ))}
          </select>
        </div>

        <button
          onClick={save}
          disabled={saving}
          className="text-xs bg-[var(--signal-amber)] text-[var(--canvas)] rounded px-3 py-1.5 font-semibold hover:opacity-90 disabled:opacity-50"
        >
          {saving ? "Saving…" : saved ? "Saved ✓" : "Save"}
        </button>
      </div>
      {error && <p className="text-xs text-[var(--signal-down)] mt-2">{error}</p>}
    </div>
  );
}
