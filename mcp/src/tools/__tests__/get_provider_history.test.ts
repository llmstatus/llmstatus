import { describe, expect, it } from "vitest";
import { formatHistory } from "../get_provider_history.js";
import type { HistoryBucket } from "../../types.js";

const makeBucket = (uptime: number, p95_ms: number, errors = 0): HistoryBucket => ({
  timestamp: "2026-04-25T00:00:00Z",
  total: 100,
  errors,
  uptime,
  p95_ms,
});

describe("formatHistory", () => {
  it("shows provider name, window, avg uptime, avg p95", () => {
    const buckets = [makeBucket(1.0, 300), makeBucket(0.99, 400), makeBucket(0.98, 500)];
    const out = formatHistory("OpenAI", "7d", buckets);
    expect(out).toContain("OpenAI");
    expect(out).toContain("7d");
    expect(out).toContain("99.00%"); // avg of 1.0+0.99+0.98 = 2.97/3
    expect(out).toContain("400ms"); // avg of 300+400+500 = 400
  });

  it("handles empty bucket list", () => {
    const out = formatHistory("OpenAI", "7d", []);
    expect(out).toContain("No history data");
  });

  it("skips p95 avg when all buckets have zero latency", () => {
    const out = formatHistory("OpenAI", "24h", [makeBucket(1.0, 0)]);
    expect(out).not.toContain("NaN");
  });
});
