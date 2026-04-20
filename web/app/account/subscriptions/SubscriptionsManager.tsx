"use client";

import { useState } from "react";

export interface Subscription {
  id: number;
  provider_id: string;
  provider_name: string;
  min_severity: string;
  email_alerts: boolean;
  email_digest: boolean;
  webhook_url: string | null;
  rss_url: string;
}

interface Props {
  initial: Subscription[];
  providers: { id: string; name: string }[];
  apiToken: string;
}

const SEVERITIES = ["minor", "major", "critical"];
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

export default function SubscriptionsManager({ initial, providers, apiToken }: Props) {
  const [subs, setSubs] = useState<Subscription[]>(initial);
  const [adding, setAdding] = useState(false);
  const [newProvider, setNewProvider] = useState("");
  const [newSeverity, setNewSeverity] = useState("major");
  const [error, setError] = useState("");

  const unsubscribed = providers.filter((p) => !subs.some((s) => s.provider_id === p.id));

  async function patch(id: number, body: Partial<Subscription>) {
    const res = await apiFetch(`/api/account/subscriptions/${id}`, {
      method: "PUT",
      body: JSON.stringify(body),
    }, apiToken);
    if (!res.ok) return;
    const json = await res.json();
    const updated = json.data ?? json;
    setSubs((prev) => prev.map((s) => (s.id === id ? { ...s, ...updated } : s)));
  }

  async function remove(id: number) {
    await apiFetch(`/api/account/subscriptions/${id}`, { method: "DELETE" }, apiToken);
    setSubs((prev) => prev.filter((s) => s.id !== id));
  }

  async function add() {
    if (!newProvider) return;
    setError("");
    const res = await apiFetch("/api/account/subscriptions", {
      method: "POST",
      body: JSON.stringify({ provider_id: newProvider, min_severity: newSeverity }),
    }, apiToken);
    if (!res.ok) {
      const { error: e } = await res.json().catch(() => ({ error: "failed" }));
      setError(e ?? "failed");
      return;
    }
    const { data } = await res.json();
    setSubs((prev) => [...prev, data]);
    setAdding(false);
    setNewProvider("");
    setNewSeverity("major");
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-sm font-semibold uppercase tracking-wide text-[var(--ink-300)]">
          Subscriptions
        </h2>
        {!adding && unsubscribed.length > 0 && (
          <button
            onClick={() => setAdding(true)}
            className="text-xs text-[var(--signal-amber)] hover:opacity-80 transition-opacity"
          >
            + Add
          </button>
        )}
      </div>

      {adding && (
        <div className="mb-4 rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] p-4 flex flex-col gap-3">
          <div className="flex gap-3">
            <select
              value={newProvider}
              onChange={(e) => setNewProvider(e.target.value)}
              className="flex-1 bg-transparent border border-[var(--ink-600)] rounded px-2 py-1 text-sm text-[var(--ink-100)] focus:outline-none"
            >
              <option value="">Select provider…</option>
              {unsubscribed.map((p) => (
                <option key={p.id} value={p.id}>{p.name}</option>
              ))}
            </select>
            <select
              value={newSeverity}
              onChange={(e) => setNewSeverity(e.target.value)}
              className="bg-transparent border border-[var(--ink-600)] rounded px-2 py-1 text-sm text-[var(--ink-100)] focus:outline-none"
            >
              {SEVERITIES.map((s) => <option key={s} value={s}>{s}</option>)}
            </select>
          </div>
          {error && <p className="text-xs text-[var(--signal-down)]">{error}</p>}
          <div className="flex gap-2">
            <button
              onClick={add}
              className="text-xs bg-[var(--signal-amber)] text-[var(--canvas)] rounded px-3 py-1 font-semibold hover:opacity-90"
            >
              Subscribe
            </button>
            <button
              onClick={() => { setAdding(false); setError(""); }}
              className="text-xs text-[var(--ink-500)] hover:text-[var(--ink-200)]"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      {subs.length === 0 && !adding && (
        <p className="text-sm text-[var(--ink-500)]">No subscriptions yet.</p>
      )}

      <div className="flex flex-col gap-3">
        {subs.map((sub) => (
          <div
            key={sub.id}
            className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] p-4"
          >
            <div className="flex items-start justify-between mb-3">
              <div>
                <p className="text-sm font-medium text-[var(--ink-100)]">{sub.provider_name}</p>
                <p className="text-xs text-[var(--ink-500)] mt-0.5">{sub.provider_id}</p>
              </div>
              <button
                onClick={() => remove(sub.id)}
                className="text-xs text-[var(--ink-500)] hover:text-[var(--signal-down)] transition-colors ml-4"
              >
                Remove
              </button>
            </div>

            <div className="grid grid-cols-2 gap-x-6 gap-y-2 text-xs">
              <label className="flex items-center gap-2 text-[var(--ink-300)]">
                <input
                  type="checkbox"
                  checked={sub.email_alerts}
                  onChange={(e) => patch(sub.id, { email_alerts: e.target.checked })}
                  className="accent-[var(--signal-amber)]"
                />
                Email alerts
              </label>
              <label className="flex items-center gap-2 text-[var(--ink-300)]">
                <input
                  type="checkbox"
                  checked={sub.email_digest}
                  onChange={(e) => patch(sub.id, { email_digest: e.target.checked })}
                  className="accent-[var(--signal-amber)]"
                />
                Daily digest
              </label>

              <div className="flex items-center gap-2 text-[var(--ink-300)]">
                <span>Min severity:</span>
                <select
                  value={sub.min_severity}
                  onChange={(e) => patch(sub.id, { min_severity: e.target.value })}
                  className="bg-transparent border border-[var(--ink-600)] rounded px-1 text-[var(--ink-100)] focus:outline-none"
                >
                  {SEVERITIES.map((s) => <option key={s} value={s}>{s}</option>)}
                </select>
              </div>

              <div className="flex items-center gap-2 text-[var(--ink-300)]">
                <span>RSS:</span>
                <a
                  href={sub.rss_url}
                  className="text-[var(--signal-amber)] hover:underline truncate"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  feed.xml
                </a>
              </div>
            </div>

            <div className="mt-3">
              <WebhookInput
                value={sub.webhook_url ?? ""}
                onSave={(url) => patch(sub.id, { webhook_url: url || null })}
              />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function WebhookInput({ value, onSave }: { value: string; onSave: (v: string) => void }) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(value);

  if (!editing) {
    return (
      <div className="flex items-center gap-2 text-xs text-[var(--ink-400)]">
        <span>Webhook:</span>
        <span className="text-[var(--ink-300)] truncate max-w-48">
          {value || <span className="italic text-[var(--ink-600)]">none</span>}
        </span>
        <button onClick={() => { setDraft(value); setEditing(true); }} className="hover:text-[var(--ink-100)]">
          edit
        </button>
      </div>
    );
  }
  return (
    <div className="flex items-center gap-2 text-xs">
      <input
        type="url"
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
        placeholder="https://…"
        className="flex-1 bg-transparent border-b border-[var(--ink-500)] text-[var(--ink-100)] focus:outline-none py-0.5"
      />
      <button
        onClick={() => { onSave(draft); setEditing(false); }}
        className="text-[var(--signal-amber)] hover:opacity-80"
      >
        Save
      </button>
      <button onClick={() => setEditing(false)} className="text-[var(--ink-500)] hover:text-[var(--ink-200)]">
        Cancel
      </button>
    </div>
  );
}
