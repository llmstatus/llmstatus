import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { LLMStatusClient } from "../../client.js";
import { ApiError } from "../../types.js";

const mockFetch = vi.fn();
beforeEach(() => {
  vi.stubGlobal("fetch", mockFetch);
});
afterEach(() => {
  vi.restoreAllMocks();
});

function makeEnvelope<T>(data: T) {
  return { data, meta: { generated_at: "2026-04-26T00:00:00Z", cache_ttl_s: 30 } };
}

describe("LLMStatusClient.listProviders", () => {
  it("returns parsed provider array on 200", async () => {
    const providers = [{ id: "openai", name: "OpenAI" }];
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => makeEnvelope(providers),
    });
    const client = new LLMStatusClient("http://localhost");
    const result = await client.listProviders();
    expect(result).toEqual(providers);
    expect(mockFetch).toHaveBeenCalledWith(
      "http://localhost/v1/providers",
      expect.objectContaining({ headers: expect.any(Object) })
    );
  });

  it("throws ApiError on 404", async () => {
    mockFetch.mockResolvedValueOnce({ ok: false, status: 404 });
    const client = new LLMStatusClient("http://localhost");
    try {
      await client.listProviders();
      expect.fail("should have thrown");
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError);
      expect(err).toHaveProperty("status", 404);
    }
  });

  it("throws ApiError on network failure", async () => {
    mockFetch.mockRejectedValueOnce(new TypeError("fetch failed"));
    const client = new LLMStatusClient("http://localhost");
    try {
      await client.listProviders();
      expect.fail("should have thrown");
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError);
      expect(err).toHaveProperty("status", 0);
    }
  });

  it("throws ApiError when response body is not valid JSON", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: async () => {
        throw new SyntaxError("Unexpected token");
      },
    });
    const client = new LLMStatusClient("http://localhost");
    try {
      await client.listProviders();
      expect.fail("should have thrown");
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError);
    }
  });
});

describe("LLMStatusClient.listIncidents", () => {
  it("appends status query param when provided", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => makeEnvelope([]),
    });
    const client = new LLMStatusClient("http://localhost");
    await client.listIncidents({ status: "ongoing" });
    expect(mockFetch).toHaveBeenCalledWith(
      "http://localhost/v1/incidents?status=ongoing",
      expect.anything()
    );
  });
});
