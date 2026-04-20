package adapters

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/llmstatus/llmstatus/internal/probes"
)

// ---- Hugging Face -----------------------------------------------------------

func TestHuggingFace_Identity(t *testing.T) {
	testIdentity(t, NewHuggingFaceProvider("k", "r"), "huggingface", huggingfaceLightModel)
}

func TestHuggingFace_ProbeLightInference(t *testing.T) {
	testOpenAICompatAdapter(t, "huggingface", "huggingface",
		func(base string) probes.Provider {
			return NewHuggingFaceProvider("k", "r", WithNewHuggingFaceProviderBaseURL(base))
		},
		huggingfaceLightModel,
	)
}

func TestHuggingFace_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	p := NewHuggingFaceProvider("k", "r",
		WithNewHuggingFaceProviderBaseURL(srv.URL),
		WithNewHuggingFaceProviderHTTPClient(&http.Client{Timeout: 30 * time.Millisecond}),
	)
	r, _ := p.ProbeLightInference(context.Background(), huggingfaceLightModel)
	if r.ErrorClass != probes.ErrorClassTimeout {
		t.Errorf("want timeout, got %q", r.ErrorClass)
	}
}

func TestHuggingFace_Unsupported(t *testing.T) {
	testUnsupported(t, NewHuggingFaceProvider("k", "r"))
}

// ---- Minimax ----------------------------------------------------------------

func TestMinimax_Identity(t *testing.T) {
	testIdentity(t, NewMinimaxProvider("k", "r"), "minimax", minimaxLightModel)
}

func TestMinimax_ProbeLightInference(t *testing.T) {
	testOpenAICompatAdapter(t, "minimax", "minimax",
		func(base string) probes.Provider {
			return NewMinimaxProvider("k", "r", WithNewMinimaxProviderBaseURL(base))
		},
		minimaxLightModel,
	)
}

func TestMinimax_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	p := NewMinimaxProvider("k", "r",
		WithNewMinimaxProviderBaseURL(srv.URL),
		WithNewMinimaxProviderHTTPClient(&http.Client{Timeout: 30 * time.Millisecond}),
	)
	r, _ := p.ProbeLightInference(context.Background(), minimaxLightModel)
	if r.ErrorClass != probes.ErrorClassTimeout {
		t.Errorf("want timeout, got %q", r.ErrorClass)
	}
}

func TestMinimax_Unsupported(t *testing.T) {
	testUnsupported(t, NewMinimaxProvider("k", "r"))
}

// ---- ByteDance (Doubao) -----------------------------------------------------

func TestDoubao_Identity(t *testing.T) {
	testIdentity(t, NewDoubaoProvider("k", "r"), "doubao", doubaoLightModel)
}

func TestDoubao_ProbeLightInference(t *testing.T) {
	testOpenAICompatAdapter(t, "doubao", "doubao",
		func(base string) probes.Provider {
			return NewDoubaoProvider("k", "r", WithNewDoubaoProviderBaseURL(base))
		},
		doubaoLightModel,
	)
}

func TestDoubao_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	p := NewDoubaoProvider("k", "r",
		WithNewDoubaoProviderBaseURL(srv.URL),
		WithNewDoubaoProviderHTTPClient(&http.Client{Timeout: 30 * time.Millisecond}),
	)
	r, _ := p.ProbeLightInference(context.Background(), doubaoLightModel)
	if r.ErrorClass != probes.ErrorClassTimeout {
		t.Errorf("want timeout, got %q", r.ErrorClass)
	}
}

func TestDoubao_Unsupported(t *testing.T) {
	testUnsupported(t, NewDoubaoProvider("k", "r"))
}
