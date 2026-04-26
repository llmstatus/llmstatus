import type {
  ApiEnvelope,
  HistoryBucket,
  IncidentResponse,
  ProviderDetail,
  ProviderSummary,
} from "./types.js";
import { ApiError } from "./types.js";

const DEFAULT_BASE_URL = "https://api.llmstatus.io";

export class LLMStatusClient {
  private readonly baseUrl: string;

  constructor(baseUrl?: string) {
    this.baseUrl = (baseUrl ?? process.env["LLMSTATUS_API_BASE"] ?? DEFAULT_BASE_URL).replace(
      /\/$/,
      ""
    );
  }

  private async fetchJSON<T>(path: string): Promise<T> {
    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}${path}`, {
        headers: { "User-Agent": "@llmstatus/mcp/1.0.0" },
        signal: AbortSignal.timeout(10_000),
      });
    } catch {
      throw new ApiError(
        0,
        "llmstatus.io is temporarily unreachable. Please try again shortly."
      );
    }
    if (!response.ok) {
      throw new ApiError(
        response.status,
        `llmstatus.io returned HTTP ${response.status}.${response.status === 404 ? " Resource not found." : " Please try again shortly."}`
      );
    }
    const envelope = (await response.json()) as ApiEnvelope<T>;
    return envelope.data;
  }

  listProviders(): Promise<ProviderSummary[]> {
    return this.fetchJSON<ProviderSummary[]>("/v1/providers");
  }

  getProvider(id: string): Promise<ProviderDetail> {
    return this.fetchJSON<ProviderDetail>(`/v1/providers/${encodeURIComponent(id)}`);
  }

  listIncidents(params: { status?: string; limit?: number } = {}): Promise<IncidentResponse[]> {
    const qs = new URLSearchParams();
    if (params.status) qs.set("status", params.status);
    if (params.limit != null) qs.set("limit", String(params.limit));
    const suffix = qs.toString() ? `?${qs.toString()}` : "";
    return this.fetchJSON<IncidentResponse[]>(`/v1/incidents${suffix}`);
  }

  getIncident(id: string): Promise<IncidentResponse> {
    return this.fetchJSON<IncidentResponse>(`/v1/incidents/${encodeURIComponent(id)}`);
  }

  getProviderHistory(id: string, window = "30d"): Promise<HistoryBucket[]> {
    return this.fetchJSON<HistoryBucket[]>(
      `/v1/providers/${encodeURIComponent(id)}/history?window=${encodeURIComponent(window)}`
    );
  }
}
