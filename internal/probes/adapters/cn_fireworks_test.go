package adapters

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/llmstatus/llmstatus/internal/probes"
)

// testOpenAICompatTimeout is a shared timeout sub-test for adapters that use
// probeOpenAICompat. The caller supplies a provider already configured with
// a short-timeout HTTP client.
func runTimeoutTest(t *testing.T, p probes.Provider, model string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	_ = srv
}

func testIdentity(t *testing.T, p probes.Provider, wantID, wantModel string) {
	t.Helper()
	if p.ID() != wantID {
		t.Errorf("ID: got %q, want %q", p.ID(), wantID)
	}
	if len(p.Models()) == 0 || p.Models()[0] != wantModel {
		t.Errorf("Models: got %v", p.Models())
	}
}

func testUnsupported(t *testing.T, p probes.Provider) {
	t.Helper()
	var ns *probes.ErrNotSupported
	for _, fn := range []func() error{
		func() error { _, e := p.ProbeQuality(context.Background(), "m"); return e },
		func() error { _, e := p.ProbeEmbedding(context.Background(), "m"); return e },
		func() error { _, e := p.ProbeStreaming(context.Background(), "m"); return e },
	} {
		if !errors.As(fn(), &ns) {
			t.Error("want ErrNotSupported")
		}
	}
}

// ---- Moonshot ---------------------------------------------------------------

func TestMoonshot_Identity(t *testing.T) {
	testIdentity(t, NewMoonshotProvider("k", "r"), "moonshot", moonshotLightModel)
}

func TestMoonshot_ProbeLightInference(t *testing.T) {
	testOpenAICompatAdapter(t, "moonshot", "moonshot",
		func(base string) probes.Provider { return NewMoonshotProvider("k", "r", WithNewMoonshotProviderBaseURL(base)) },
		moonshotLightModel,
	)
}

func TestMoonshot_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond); w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	p := NewMoonshotProvider("k", "r", WithNewMoonshotProviderBaseURL(srv.URL), WithNewMoonshotProviderHTTPClient(&http.Client{Timeout: 30 * time.Millisecond}))
	r, _ := p.ProbeLightInference(context.Background(), moonshotLightModel)
	if r.ErrorClass != probes.ErrorClassTimeout {
		t.Errorf("want timeout, got %q", r.ErrorClass)
	}
}

func TestMoonshot_Unsupported(t *testing.T) { testUnsupported(t, NewMoonshotProvider("k", "r")) }

// ---- Zhipu ------------------------------------------------------------------

func TestZhipu_Identity(t *testing.T) {
	testIdentity(t, NewZhipuProvider("k", "r"), "zhipu", zhipuLightModel)
}

func TestZhipu_ProbeLightInference(t *testing.T) {
	testOpenAICompatAdapter(t, "zhipu", "zhipu",
		func(base string) probes.Provider { return NewZhipuProvider("k", "r", WithNewZhipuProviderBaseURL(base)) },
		zhipuLightModel,
	)
}

func TestZhipu_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond); w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	p := NewZhipuProvider("k", "r", WithNewZhipuProviderBaseURL(srv.URL), WithNewZhipuProviderHTTPClient(&http.Client{Timeout: 30 * time.Millisecond}))
	r, _ := p.ProbeLightInference(context.Background(), zhipuLightModel)
	if r.ErrorClass != probes.ErrorClassTimeout {
		t.Errorf("want timeout, got %q", r.ErrorClass)
	}
}

func TestZhipu_Unsupported(t *testing.T) { testUnsupported(t, NewZhipuProvider("k", "r")) }

// ---- 01.AI ------------------------------------------------------------------

func TestZeroOne_Identity(t *testing.T) {
	testIdentity(t, NewZeroOneProvider("k", "r"), "zeroone_ai", zerooneLightModel)
}

func TestZeroOne_ProbeLightInference(t *testing.T) {
	testOpenAICompatAdapter(t, "zeroone_ai", "zeroone",
		func(base string) probes.Provider { return NewZeroOneProvider("k", "r", WithNewZeroOneProviderBaseURL(base)) },
		zerooneLightModel,
	)
}

func TestZeroOne_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond); w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	p := NewZeroOneProvider("k", "r", WithNewZeroOneProviderBaseURL(srv.URL), WithNewZeroOneProviderHTTPClient(&http.Client{Timeout: 30 * time.Millisecond}))
	r, _ := p.ProbeLightInference(context.Background(), zerooneLightModel)
	if r.ErrorClass != probes.ErrorClassTimeout {
		t.Errorf("want timeout, got %q", r.ErrorClass)
	}
}

func TestZeroOne_Unsupported(t *testing.T) { testUnsupported(t, NewZeroOneProvider("k", "r")) }

// ---- Qwen -------------------------------------------------------------------

func TestQwen_Identity(t *testing.T) {
	testIdentity(t, NewQwenProvider("k", "r"), "qwen", qwenLightModel)
}

func TestQwen_ProbeLightInference(t *testing.T) {
	testOpenAICompatAdapter(t, "qwen", "qwen",
		func(base string) probes.Provider { return NewQwenProvider("k", "r", WithNewQwenProviderBaseURL(base)) },
		qwenLightModel,
	)
}

func TestQwen_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond); w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	p := NewQwenProvider("k", "r", WithNewQwenProviderBaseURL(srv.URL), WithNewQwenProviderHTTPClient(&http.Client{Timeout: 30 * time.Millisecond}))
	r, _ := p.ProbeLightInference(context.Background(), qwenLightModel)
	if r.ErrorClass != probes.ErrorClassTimeout {
		t.Errorf("want timeout, got %q", r.ErrorClass)
	}
}

func TestQwen_Unsupported(t *testing.T) { testUnsupported(t, NewQwenProvider("k", "r")) }

// ---- Fireworks --------------------------------------------------------------

func TestFireworks_Identity(t *testing.T) {
	testIdentity(t, NewFireworksProvider("k", "r"), "fireworks", fireworksLightModel)
}

func TestFireworks_ProbeLightInference(t *testing.T) {
	testOpenAICompatAdapter(t, "fireworks", "fireworks",
		func(base string) probes.Provider {
			return NewFireworksProvider("k", "r", WithNewFireworksProviderBaseURL(base))
		},
		fireworksLightModel,
	)
}

func TestFireworks_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond); w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	p := NewFireworksProvider("k", "r", WithNewFireworksProviderBaseURL(srv.URL), WithNewFireworksProviderHTTPClient(&http.Client{Timeout: 30 * time.Millisecond}))
	r, _ := p.ProbeLightInference(context.Background(), fireworksLightModel)
	if r.ErrorClass != probes.ErrorClassTimeout {
		t.Errorf("want timeout, got %q", r.ErrorClass)
	}
}

func TestFireworks_Unsupported(t *testing.T) { testUnsupported(t, NewFireworksProvider("k", "r")) }
