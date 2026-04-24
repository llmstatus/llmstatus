"use server";

import { getProviderHistory, type HistoryBucket, type HistoryWindow } from "@/lib/api";

const API_URL = process.env.API_URL ?? "http://localhost:8081";

export async function postReport(providerId: string): Promise<void> {
  await fetch(`${API_URL}/v1/providers/${encodeURIComponent(providerId)}/report`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
  });
}

export async function fetchProviderHistory(
  providerId: string,
  window: HistoryWindow,
): Promise<HistoryBucket[] | null> {
  return getProviderHistory(providerId, window).catch(() => null);
}
