import type { ModelStat } from "./api";

export function aggregateSparklines(modelStats: ModelStat[]): number[] {
  const BUCKETS = 60;
  const sums = new Array<number>(BUCKETS).fill(0);
  const counts = new Array<number>(BUCKETS).fill(0);
  for (const m of modelStats) {
    for (let i = 0; i < BUCKETS; i++) {
      if (m.sparkline[i] > 0) {
        sums[i] += m.sparkline[i];
        counts[i]++;
      }
    }
  }
  return sums.map((s, i) => (counts[i] > 0 ? s / counts[i] : 0));
}
