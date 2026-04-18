const API_URL = process.env.API_URL ?? "http://localhost:8081";

interface ApiEnvelope<T> {
  data: T;
  meta: { generated_at: string; cache_ttl_s: number };
}

export class ApiNotFoundError extends Error {
  constructor(path: string) {
    super(`Not found: ${path}`);
    this.name = "ApiNotFoundError";
  }
}

export type ProviderStatus = "operational" | "degraded" | "down";
export type Severity = "minor" | "major" | "critical";

export interface ProviderSummary {
  id: string;
  name: string;
  category: string;
  region: string;
  current_status: ProviderStatus;
  active_incident_id?: string;
}

export interface IncidentRef {
  id: string;
  slug: string;
  severity: Severity;
  title: string;
  status: string;
  started_at: string;
}

export interface ModelSummary {
  model_id: string;
  display_name: string;
  model_type: string;
  active: boolean;
}

export interface ProviderDetail extends ProviderSummary {
  status_page_url?: string;
  documentation_url?: string;
  models: ModelSummary[];
  active_incidents: IncidentRef[];
}

export type IncidentStatus = "ongoing" | "monitoring" | "resolved";

export interface IncidentDetail {
  id: string;
  slug: string;
  provider_id: string;
  severity: Severity;
  title: string;
  description?: string;
  status: IncidentStatus;
  affected_models: string[];
  affected_regions: string[];
  started_at: string;
  resolved_at?: string;
  detection_method: string;
  human_reviewed: boolean;
}

async function apiFetch<T>(path: string, revalidate: number): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    next: { revalidate },
    headers: { Accept: "application/json" },
  });
  if (res.status === 404) throw new ApiNotFoundError(path);
  if (!res.ok) throw new Error(`API ${path} returned ${res.status}`);
  const envelope: ApiEnvelope<T> = await res.json();
  return envelope.data;
}

export function listProviders(): Promise<ProviderSummary[]> {
  return apiFetch<ProviderSummary[]>("/v1/providers", 30);
}

export function getProvider(id: string): Promise<ProviderDetail> {
  return apiFetch<ProviderDetail>(`/v1/providers/${encodeURIComponent(id)}`, 60);
}

export function listIncidents(status = "all", limit = 50): Promise<IncidentDetail[]> {
  return apiFetch<IncidentDetail[]>(`/v1/incidents?status=${status}&limit=${limit}`, 30);
}

export function getIncident(slug: string): Promise<IncidentDetail> {
  return apiFetch<IncidentDetail>(`/v1/incidents/${encodeURIComponent(slug)}`, 60);
}
