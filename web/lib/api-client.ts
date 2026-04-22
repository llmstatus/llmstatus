const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081'

export interface Provider {
  id: string
  name: string
  status: 'operational' | 'degraded' | 'down'
  last_updated: string
}

export interface Incident {
  id: string
  title: string
  status: 'investigating' | 'identified' | 'monitoring' | 'resolved'
  created_at: string
  updated_at: string
}

export class APIError extends Error {
  constructor(public status: number, message: string) {
    super(message)
    this.name = 'APIError'
  }
}

async function fetcher<T>(url: string): Promise<T> {
  const response = await fetch(`${API_BASE}${url}`)

  if (!response.ok) {
    throw new APIError(response.status, `API Error: ${response.statusText}`)
  }

  return response.json()
}

export const api = {
  providers: () => fetcher<Provider[]>('/api/providers'),
  provider: (id: string) => fetcher<Provider>(`/api/providers/${id}`),
  incidents: () => fetcher<Incident[]>('/api/incidents'),
  incident: (id: string) => fetcher<Incident>(`/api/incidents/${id}`),
}

export { fetcher }