export interface ApiEnvelope<T> {
  data: T;
  meta: {
    generated_at: string;
    cache_ttl_s: number;
  };
}

export interface ModelStat {
  model_id: string;
  display_name: string;
  uptime_24h: number;
  p95_ms: number;
  sparkline: number[];
}

export interface ProviderSummary {
  id: string;
  name: string;
  category: string;
  region: string;
  probe_scope: string;
  current_status: "operational" | "degraded" | "down";
  active_incident_id?: string;
  uptime_24h?: number;
  p95_ms?: number;
  model_stats: ModelStat[];
}

export interface ModelSummary {
  model_id: string;
  display_name: string;
  model_type: string;
  active: boolean;
}

export interface IncidentRef {
  id: string;
  slug: string;
  severity: string;
  title: string;
  status: string;
  started_at: string;
}

export interface RegionStat {
  region_id: string;
  uptime_24h: number;
  p95_ms: number;
}

export interface ProviderDetail extends ProviderSummary {
  status_page_url?: string;
  documentation_url?: string;
  models: ModelSummary[];
  active_incidents: IncidentRef[];
  region_stats: RegionStat[];
}

export interface IncidentResponse {
  id: string;
  slug: string;
  provider_id: string;
  severity: "critical" | "major" | "minor";
  title: string;
  description?: string;
  status: "ongoing" | "monitoring" | "resolved";
  affected_models: string[];
  affected_regions: string[];
  started_at: string;
  resolved_at?: string;
  detection_method: string;
  human_reviewed: boolean;
}

export interface HistoryBucket {
  timestamp: string;
  total: number;
  errors: number;
  uptime: number;
  p95_ms: number;
}

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    message: string
  ) {
    super(message);
    this.name = "ApiError";
  }
}
