import { renderHook } from "@testing-library/react";
import { useProviderStatus } from "@/hooks/useProviderStatus";
import { RealtimeProvider } from "@/lib/realtime-context";

// useProviderStatus pulls data through SWR and subscribes to a WebSocket via
// RealtimeProvider. Neither belongs in a unit test: SWR is stubbed to return
// a canned payload synchronously, and RealtimeProvider is rendered unmocked
// (its WebSocket connection fails fast in jsdom, which is the expected path).
jest.mock("swr", () => {
  const actual = jest.requireActual("swr");
  return {
    ...actual,
    __esModule: true,
    default: jest.fn().mockReturnValue({
      data: [
        {
          id: "openai",
          name: "OpenAI",
          status: "operational",
          last_updated: new Date(0).toISOString(),
        },
      ],
      error: undefined,
      isLoading: false,
      mutate: jest.fn(),
    }),
  };
});

const wrapper = ({ children }: { children: React.ReactNode }) => (
  <RealtimeProvider>{children}</RealtimeProvider>
);

describe("useProviderStatus", () => {
  it("should merge SWR provider data with real-time updates", () => {
    const { result } = renderHook(() => useProviderStatus(), { wrapper });

    expect(result.current.isLoading).toBe(false);
    expect(result.current.data?.[0].id).toBe("openai");
    expect(typeof result.current.mutate).toBe("function");
  });
});
