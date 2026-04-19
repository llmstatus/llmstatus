"use server";

import { getProviderHistory, type HistoryBucket, type HistoryWindow } from "@/lib/api";

export async function fetchProviderHistory(
  providerId: string,
  window: HistoryWindow,
): Promise<HistoryBucket[] | null> {
  return getProviderHistory(providerId, window).catch(() => null);
}
