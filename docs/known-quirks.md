# Known Quirks

Provider-specific weirdness we have observed or confirmed.

Every LLM API has quirks. This file exists so we don't re-discover the
same issues every time we add or modify an adapter, and so humans can
audit our assumptions.

**When to add an entry**: whenever you encounter behavior that is not
documented, contradicts the documentation, or is surprising enough that
a future developer would also be surprised.

**When to remove an entry**: only when the provider confirms the quirk
is fixed AND you have verified it against their live API.

---

## Format

```
## {provider_id}

### {short description of the quirk}
**First observed**: YYYY-MM-DD
**Confirmed**: YYYY-MM-DD (by whom and how)
**What happens**: [concrete description]
**How we handle it**: [our code's response]
**References**: [links to docs, issues, or commits]
```

---

## openai

### HTTP 200 responses can carry an `error` object
**First observed**: LLMS-002 scaffolding, 2026-04-18
**What happens**: The OpenAI API occasionally returns a well-formed
`application/json` body with HTTP 200 whose top-level object is an
`{"error": {...}}` envelope rather than a `chat.completion` response.
A naive `status < 300` check will wrongly mark such responses as
successful.
**How we handle it**: `classifyOpenAIResponse` (see
`internal/probes/adapters/openai.go`) requires the decoded body to
match the `chat.completion` shape (`choices` + `usage`). Bodies that
decode but lack that shape fall through to `ErrorClassMalformedBody`.
**References**: OpenAI Discuss threads 2024-Q4 report sporadic 200+error
behaviour during transient routing issues.

### Authentication surface: `401` can carry varied `code` strings
**First observed**: 2026-04-18 (fixtures)
**What happens**: An invalid key returns HTTP 401 with
`error.code` ∈ `{"invalid_api_key", "missing_api_key",
"mismatched_organization", ...}`. We should classify all 401/403 as
`ErrorClassAuth` without relying on the code string, because the code
taxonomy has changed at least twice.
**How we handle it**: `classifyOpenAIStatus` treats 401 and 403 as
`auth` regardless of body content.
**References**: `internal/probes/adapters/openai.go:classifyOpenAIStatus`

### Suspects not yet confirmed (fill in as real probe data arrives)
- Rate limit headers (`x-ratelimit-remaining-requests` vs
  `x-ratelimit-remaining-tokens`) may be absent on some endpoints.
- o1/o3 models use `max_completion_tokens`, not `max_tokens`.
- Organization and project headers are sometimes required.
- Streaming (`data: [DONE]`) formatting differs between legacy and new
  models.

---

## anthropic

### Auth: `x-api-key` header, not `Authorization: Bearer`
**First observed**: LLMS-008, 2026-04-18
**What happens**: Anthropic authentication uses the header `x-api-key: <key>`
instead of the OAuth2-style `Authorization: Bearer <key>` used by OpenAI and
most other providers.
**How we handle it**: `buildLightRequest` in `anthropic.go` sets `x-api-key`
explicitly. An incorrect `Authorization` header is silently ignored by the API.
**References**: `internal/probes/adapters/anthropic.go:buildLightRequest`

### `anthropic-version` header is required on every request
**First observed**: LLMS-008, 2026-04-18
**What happens**: Requests without the `anthropic-version` header are rejected
with a 400. The minimum accepted value is `2023-06-01`; we pin to this.
**How we handle it**: Set unconditionally in `buildLightRequest`.
**References**: https://docs.anthropic.com/en/api/versioning

### HTTP 529 = overloaded (non-standard)
**First observed**: Reported from Claude 3.5 launch weekend 2024-10; added
to fixtures LLMS-008, 2026-04-18
**What happens**: Anthropic returns HTTP 529 with
`{"type":"error","error":{"type":"overloaded_error","message":"Overloaded"}}`
when capacity is constrained. HTTP 529 is not an IANA-registered status code.
**How we handle it**: `classifyAnthropicStatus` treats 529 the same as 5xx
(`ErrorClassServer5xx`). We do not use a separate `overloaded` class to
keep the taxonomy stable (see METHODOLOGY.md §5.4).
**References**: `internal/probes/adapters/anthropic.go:classifyAnthropicStatus`

### Usage fields differ from OpenAI convention
**First observed**: LLMS-008, 2026-04-18
**What happens**: Anthropic response bodies use `input_tokens`/`output_tokens`
(inside a `usage` object) instead of OpenAI's `prompt_tokens`/`completion_tokens`.
**How we handle it**: `anthropicMessagesResponse.Usage` maps to the correct
fields. `ProbeResult.TokensIn`/`TokensOut` are populated accordingly.
**References**: `internal/probes/adapters/anthropic.go:anthropicMessagesResponse`

---

## google_gemini

### (placeholder)

Known suspects:
- Safety filter frequently blocks neutral test prompts; probe prompts must
  be carefully chosen
- Different base URLs for Vertex AI vs generative language API
- Authentication via OAuth2 for Vertex vs API key for generative language
- Streaming format differs from OpenAI SSE convention

---

## aws_bedrock

### (placeholder)

Known suspects:
- IAM-based auth, not API keys; requires AWS SDK
- Different model IDs (`anthropic.claude-opus-4-v1` format)
- Throughput provisioning affects latency measurements
- Regional model availability varies

---

## deepseek

### (placeholder)

Known suspects:
- Occasional CDN hiccups from outside China
- Pricing changes announced on short notice
- Model versions sometimes change without new model ID
- Context caching behavior unique to DeepSeek

---

## moonshot

### (placeholder)

Known suspects:
- Rate limiting by both RPM and TPM with separate windows
- Context window extensions sometimes change billing structure

---

## zhipu

### (placeholder)

Known suspects:
- Chinese-only support channels
- JWT authentication, not static API keys
- Different endpoint URL for different model families

---

## qwen

### (placeholder)

Known suspects:
- DashScope vs Bailian endpoints behave differently
- International and mainland versions are separate accounts with
  separate pricing

---

## openrouter

### (placeholder)

Known suspects:
- Requests can route to different underlying providers within the same
  model name — latency will vary
- BYOK (bring your own key) mode has different billing and may have
  different latency characteristics
- Fallback routing can produce surprising latency spikes when a
  preferred provider is down

---

## Self-monitored services (template)

_Any service operated by the maintainers (see METHODOLOGY.md §11) appears
in this section. No such service is listed publicly until the operator
authorizes disclosure — but the invariants below apply from day one._

### Self-monitoring invariants
Data from a self-monitored service must be treated identically to any
other provider. Any divergent handling is a bug.

**Tests to maintain**:
- Self-monitored probes run from the same nodes as other providers
- Self-monitored incidents surface through the same detection rules
- Self-monitored detail pages use the same template
- Self-monitored services do not receive preferential latency reporting

---

## Cross-cutting quirks (affect multiple providers)

### Timestamp inconsistency
Some providers return timestamps in local time, some in UTC, some in
Unix epoch seconds, some in milliseconds. Always normalize to UTC ISO 8601
on ingestion. Never trust client-side timestamps from provider responses.

### Token count discrepancies
Provider-reported token counts may disagree with `tiktoken` or other
local tokenizer libraries by 1-5% for long inputs. Use provider counts
as the source of truth for our metrics.

### Rate limit header naming
The `Retry-After` header is standard, but many providers add their own:
- OpenAI: `x-ratelimit-remaining-requests`, `x-ratelimit-reset-requests`
- Anthropic: `anthropic-ratelimit-requests-remaining`, `anthropic-ratelimit-requests-reset`
- Others: varied, or absent

Parse conservatively. When in doubt, back off with exponential jitter.

### Gradual rollouts
Providers deploy changes regionally over hours. Two probes sent within
seconds of each other from different nodes may hit different versions.
This is real data, not an error. Record node origin with every probe.

---

End of known-quirks.md — populate as real behavior is observed.
