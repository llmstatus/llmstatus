"use client";

import { useState } from "react";

export interface SponsorKey {
  provider_id: string;
  key_hint: string;
  active: boolean;
  last_verified_at: string | null;
  last_error: string | null;
}

interface Props {
  keys: SponsorKey[];
  providers: { id: string; name: string }[];
  apiToken: string;
}

async function apiFetch(path: string, init: RequestInit, token: string) {
  return fetch(path, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
      ...(init.headers as Record<string, string>),
    },
  });
}

export default function KeyManager({ keys: initial, providers, apiToken }: Props) {
  const [keys, setKeys] = useState<SponsorKey[]>(initial);
  const [editing, setEditing] = useState<string | null>(null);
  const [draft, setDraft] = useState("");
  const [adding, setAdding] = useState(false);
  const [newProvider, setNewProvider] = useState("");
  const [newKey, setNewKey] = useState("");
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  const keyedProviders = new Set(keys.map((k) => k.provider_id));
  const unkeyedProviders = providers.filter((p) => !keyedProviders.has(p.id));

  async function upsert(providerId: string, keyValue: string) {
    setSaving(true);
    setError("");
    const res = await apiFetch(
      `/api/sponsor/me/keys/${providerId}`,
      { method: "PUT", body: JSON.stringify({ key: keyValue }) },
      apiToken
    );
    setSaving(false);
    if (!res.ok) {
      const { error: e } = await res.json().catch(() => ({ error: "failed" }));
      setError(e ?? "failed");
      return false;
    }
    const row: SponsorKey = await res.json();
    setKeys((prev) =>
      prev.some((k) => k.provider_id === providerId)
        ? prev.map((k) => (k.provider_id === providerId ? row : k))
        : [...prev, row]
    );
    return true;
  }

  async function remove(providerId: string) {
    setSaving(true);
    setError("");
    await apiFetch(`/api/sponsor/me/keys/${providerId}`, { method: "DELETE" }, apiToken);
    setSaving(false);
    setKeys((prev) => prev.filter((k) => k.provider_id !== providerId));
  }

  async function saveEdit(providerId: string) {
    const ok = await upsert(providerId, draft);
    if (ok) { setEditing(null); setDraft(""); }
  }

  async function saveNew() {
    if (!newProvider || !newKey) { setError("Provider and key required"); return; }
    const ok = await upsert(newProvider, newKey);
    if (ok) { setAdding(false); setNewProvider(""); setNewKey(""); }
  }

  const providerName = (id: string) => providers.find((p) => p.id === id)?.name ?? id;

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-sm font-semibold uppercase tracking-wide text-[var(--ink-300)]">
          API Keys
        </h2>
        {!adding && unkeyedProviders.length > 0 && (
          <button
            onClick={() => { setAdding(true); setError(""); }}
            className="text-xs text-[var(--signal-amber)] hover:opacity-80 transition-opacity"
          >
            + Add key
          </button>
        )}
      </div>

      {error && <p className="mb-3 text-xs text-[var(--signal-down)]">{error}</p>}

      {adding && (
        <div className="mb-4 rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] p-4 flex flex-col gap-3">
          <p className="text-xs text-[var(--ink-400)]">
            Keys are encrypted at rest. We only store the last 4 characters for display.
          </p>
          <select
            value={newProvider}
            onChange={(e) => setNewProvider(e.target.value)}
            className="bg-transparent border border-[var(--ink-600)] rounded px-2 py-1 text-sm text-[var(--ink-100)] focus:outline-none"
          >
            <option value="">Select provider…</option>
            {unkeyedProviders.map((p) => (
              <option key={p.id} value={p.id}>{p.name}</option>
            ))}
          </select>
          <input
            type="password"
            value={newKey}
            onChange={(e) => setNewKey(e.target.value)}
            placeholder="sk-… or API key"
            className="bg-transparent border border-[var(--ink-600)] rounded px-2 py-1 text-sm text-[var(--ink-100)] focus:outline-none font-mono"
          />
          <div className="flex gap-2">
            <button
              onClick={saveNew}
              disabled={saving}
              className="text-xs bg-[var(--signal-amber)] text-[var(--canvas)] rounded px-3 py-1 font-semibold hover:opacity-90 disabled:opacity-50"
            >
              {saving ? "Saving…" : "Save"}
            </button>
            <button
              onClick={() => { setAdding(false); setError(""); setNewKey(""); }}
              className="text-xs text-[var(--ink-500)] hover:text-[var(--ink-200)]"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      {keys.length === 0 && !adding && (
        <p className="text-sm text-[var(--ink-500)]">No keys yet.</p>
      )}

      <div className="flex flex-col gap-2">
        {keys.map((k) => (
          <div
            key={k.provider_id}
            className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] p-4"
          >
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-[var(--ink-100)]">
                  {providerName(k.provider_id)}
                  <span className={`ml-2 text-xs ${k.active ? "text-[var(--signal-up)]" : "text-[var(--signal-down)]"}`}>
                    {k.active ? "active" : "error"}
                  </span>
                </p>
                <p className="text-xs text-[var(--ink-500)] font-mono mt-0.5">{k.key_hint}</p>
                {k.last_error && (
                  <p className="text-xs text-[var(--signal-down)] mt-1">{k.last_error}</p>
                )}
              </div>
              <div className="flex gap-3 ml-4 shrink-0">
                <button
                  onClick={() => { setEditing(k.provider_id); setDraft(""); setError(""); }}
                  className="text-xs text-[var(--ink-500)] hover:text-[var(--ink-200)]"
                >
                  Replace
                </button>
                <button
                  onClick={() => remove(k.provider_id)}
                  className="text-xs text-[var(--ink-500)] hover:text-[var(--signal-down)]"
                >
                  Remove
                </button>
              </div>
            </div>

            {editing === k.provider_id && (
              <div className="mt-3 flex gap-2">
                <input
                  type="password"
                  value={draft}
                  onChange={(e) => setDraft(e.target.value)}
                  placeholder="New key value"
                  className="flex-1 bg-transparent border-b border-[var(--ink-500)] text-sm text-[var(--ink-100)] font-mono focus:outline-none py-0.5"
                />
                <button
                  onClick={() => saveEdit(k.provider_id)}
                  disabled={saving}
                  className="text-xs text-[var(--signal-amber)] hover:opacity-80 disabled:opacity-50"
                >
                  {saving ? "…" : "Save"}
                </button>
                <button
                  onClick={() => { setEditing(null); setDraft(""); }}
                  className="text-xs text-[var(--ink-500)] hover:text-[var(--ink-200)]"
                >
                  Cancel
                </button>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
