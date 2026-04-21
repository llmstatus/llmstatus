package adapters

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/llmstatus/llmstatus/internal/probes"
)

// ---- Azure OpenAI -----------------------------------------------------------

func TestAzureOpenAI_Identity(t *testing.T) {
	p := NewAzureOpenAIProvider("k", "myresource", "gpt-4o-mini", "2024-10-21", "r")
	if p.ID() != azureOpenAIProviderID {
		t.Errorf("ID: got %q", p.ID())
	}
	if len(p.Models()) == 0 || p.Models()[0] != "gpt-4o-mini" {
		t.Errorf("Models: got %v", p.Models())
	}
}

func TestAzureOpenAI_ProbeLightInference(t *testing.T) {
	const deployment = "gpt-4o-mini"

	cases := []struct {
		name         string
		httpStatus   int
		fixture      string
		wantSuccess  bool
		wantErrClass probes.ErrorClass
	}{
		{"success", http.StatusOK, "chat_completions_200.json", true, probes.ErrorClassNone},
		{"auth", http.StatusUnauthorized, "chat_completions_401.json", false, probes.ErrorClassAuth},
		{"rate_limit", http.StatusTooManyRequests, "chat_completions_429.json", false, probes.ErrorClassRateLimit},
		{"server_5xx", http.StatusInternalServerError, "chat_completions_500.json", false, probes.ErrorClassServer5xx},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := mustReadFixtureDir(t, "azure_openai", tc.fixture)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				wantPathPrefix := "/openai/deployments/" + deployment
				if !strings.HasPrefix(r.URL.Path, wantPathPrefix) {
					t.Errorf("path: got %q, want prefix %q", r.URL.Path, wantPathPrefix)
				}
				if r.Header.Get("api-key") == "" {
					t.Error("missing api-key header")
				}
				w.WriteHeader(tc.httpStatus)
				_, _ = w.Write(body)
			}))
			t.Cleanup(srv.Close)

			p := NewAzureOpenAIProvider("k", "res", deployment, "2024-10-21", "r",
				WithAzureOpenAIBaseURL(srv.URL))
			result, err := p.ProbeLightInference(context.Background(), deployment)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.ProviderID != azureOpenAIProviderID {
				t.Errorf("ProviderID: got %q", result.ProviderID)
			}
			if result.Success != tc.wantSuccess {
				t.Errorf("Success: got %v, want %v", result.Success, tc.wantSuccess)
			}
			if result.ErrorClass != tc.wantErrClass {
				t.Errorf("ErrorClass: got %q, want %q", result.ErrorClass, tc.wantErrClass)
			}
		})
	}
}

func TestAzureOpenAI_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	p := NewAzureOpenAIProvider("k", "res", "gpt-4o-mini", "2024-10-21", "r",
		WithAzureOpenAIBaseURL(srv.URL),
		WithAzureOpenAIHTTPClient(&http.Client{Timeout: 30 * time.Millisecond}),
	)
	r, _ := p.ProbeLightInference(context.Background(), "gpt-4o-mini")
	if r.ErrorClass != probes.ErrorClassTimeout {
		t.Errorf("want timeout, got %q", r.ErrorClass)
	}
}

func TestAzureOpenAI_Unsupported(t *testing.T) {
	testUnsupported(t, NewAzureOpenAIProvider("k", "res", "gpt-4o-mini", "2024-10-21", "r"))
}

// ---- AI21 Labs --------------------------------------------------------------

func TestAI21_Identity(t *testing.T) {
	testIdentity(t, NewAI21Provider("k", "r"), ai21ProviderID, ai21LightModel)
}

func TestAI21_ProbeLightInference(t *testing.T) {
	testOpenAICompatAdapter(t, ai21ProviderID, "ai21",
		func(base string) probes.Provider {
			return NewAI21Provider("k", "r", WithAI21BaseURL(base))
		},
		ai21LightModel,
	)
}

func TestAI21_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	p := NewAI21Provider("k", "r",
		WithAI21BaseURL(srv.URL),
		WithAI21HTTPClient(&http.Client{Timeout: 30 * time.Millisecond}),
	)
	r, _ := p.ProbeLightInference(context.Background(), ai21LightModel)
	if r.ErrorClass != probes.ErrorClassTimeout {
		t.Errorf("want timeout, got %q", r.ErrorClass)
	}
}

func TestAI21_Unsupported(t *testing.T) {
	testUnsupported(t, NewAI21Provider("k", "r"))
}
