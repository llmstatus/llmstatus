const API_URL = process.env.API_URL ?? "http://localhost:8081";

interface ApiEnvelope<T> {
  data: T;
  meta: { generated_at: string; cache_ttl_s: number };
}

export type ProviderStatus = "operational" | "degraded" | "down";

export interface ProviderSummary {
  id: string;
  name: string;
  category: string;
  region: string;
  current_status: ProviderStatus;
  active_incident_id?: string;
}

export interface IncidentSummary {
  id: string;
  slug: string;
  provider_id: string;
  severity: "minor" | "major" | "critical";
  title: string;
  status: "ongoing" | "monitoring" | "resolved";
  started_at: string;
  resolved_at?: string;
}

async function apiFetch<T>(path: string, revalidate: number): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    next: { revalidate },
    headers: { Accept: "application/json" },
  });
  if (!res.ok) {
    throw new Error(`API ${path} returned ${res.status}`);
  }
  const envelope: ApiEnvelope<T> = await res.json();
  return envelope.data;
}

export function listProviders(): Promise<ProviderSummary[]> {
  return apiFetch<ProviderSummary[]>("/v1/providers", 30);
}

export function listIncidents(status = "ongoing"): Promise<IncidentSummary[]> {
  return apiFetch<IncidentSummary[]>(`/v1/incidents?status=${status}`, 30);
}
