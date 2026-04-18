# Methodology

**Version**: 0.1
**Last updated**: 2026-04-17
**Published at**: https://llmstatus.io/methodology

---

This document describes exactly how llmstatus.io measures what it measures.
It exists so that every metric on our site can be independently verified,
questioned, and — if needed — corrected.

If anything below is unclear, incomplete, or appears contradicted by our
data, please contact us at `methodology@llmstatus.io`.

---

## 1. What We Measure

We measure **observed behavior of AI API endpoints** from a fixed set of
geographic locations. Specifically:

1. **Reachability**: Does the API endpoint respond at all?
2. **Success rate**: What percentage of authenticated requests succeed?
3. **Latency**: How long does it take to get a response?
4. **Streaming performance**: Time-to-first-token (TTFT) and tokens-per-second
   for streaming endpoints.
5. **Error taxonomy**: When requests fail, what kind of failure is it?

We do **not** measure:

- Output quality or correctness (except in future versions, see §9)
- Cost or billing accuracy
- Internal metrics of the provider (we have no access)
- Anything about user-side code, network, or environment

---

## 2. What We Do NOT Do

To be explicit about what our data is not:

- **We do not scrape official status pages.** Sites like status.openai.com
  and status.anthropic.com are publisher-controlled. Our data is
  independently collected.

- **We do not use public traffic or user-submitted reports** like
  Downdetector's crowdsourced model. All our data comes from probes we run.

- **We do not measure end-user experience.** We measure what our probes
  observe. Your experience may differ based on your network, region,
  account tier, or usage pattern.

- **We do not represent the provider's position.** Our observations may
  differ from what providers publish on their own status pages. This is
  expected and by design.

---

## 3. Measurement Architecture

### 3.1 Probe Nodes

Probes originate from **7 geographic locations**:

| Region | Provider | Rationale |
|---|---|---|
| US West (Oregon) | AWS us-west-2 | Near most North American AI provider origins |
| US East (Virginia) | AWS us-east-1 | Secondary North American coverage |
| Europe (Germany) | Hetzner | Independent of major cloud providers |
| Japan (Tokyo) | AWS ap-northeast-1 | East Asian coverage |
| Singapore | AWS ap-southeast-1 | Southeast Asian coverage |
| China (Shanghai) | Alibaba Cloud | Mainland China access reality |
| China (Guangzhou) | Tencent Cloud | Mainland China redundancy |

All nodes run the same probe code. Code version is recorded with each probe
result.

### 3.2 Probe Frequency

| Probe type | Frequency | Purpose |
|---|---|---|
| HTTP reachability | every 30 seconds | Catch hard outages quickly |
| Light inference | every 60 seconds | End-to-end check with minimal payload |
| Medium inference | every 5 minutes | Realistic latency measurement |
| Streaming probe | every 5 minutes | TTFT and throughput |

All probe types run from all 7 nodes independently. Times are coordinated
via NTP; each probe result includes a timestamp in UTC.

### 3.3 What Each Probe Does

**Light inference** sends the prompt `"Reply with OK"` to a selected model
and expects a short response. Token budget: 10 output tokens.

**Medium inference** sends a ~200-token prompt and requests up to 100 output
tokens. The exact prompt is the same across all providers and is published
at https://llmstatus.io/methodology/probe-prompts.

**Streaming probe** uses the same medium-inference prompt but requests a
streaming response. We measure:

- **TTFT**: time from request start to first non-empty SSE event
- **Total time**: request start to stream close
- **Tokens/sec**: output tokens divided by (total time - TTFT)

---

## 4. Authentication and Accounts

We use **paid API accounts** with every provider we monitor. We do not use:

- Free trial credits
- Shared keys
- Keys obtained from aggregators
- Third-party reseller accounts

This means our probe traffic is treated as normal paying-customer traffic
by each provider, and our measurements reflect the experience of such
customers.

Each provider account is **dedicated solely to monitoring**. We do not mix
monitoring calls with any other production traffic operated by the
maintainers.

---

## 5. Data Definitions

### 5.1 Success

A probe is classified as **successful** if **all** of the following are true:

- The HTTP response status is 2xx
- The response body is well-formed JSON (or a valid SSE stream, for streaming)
- The response contains a non-empty completion
- No explicit error field is present in the body

Partial successes (e.g., HTTP 200 with empty body) count as **failures**
under error type `empty_response`.

### 5.2 Uptime

Uptime is calculated as:

```
uptime = successful_probes / total_probes
```

...over the stated window (24h, 7d, 30d).

**Important**: we weight all probe types equally in this calculation. A
detailed breakdown by probe type is available on each provider's detail page.

### 5.3 Latency

Latency metrics (p50, p95, p99) are computed from **successful probe
responses only**. Failed probes have no latency value.

Latency is **wall-clock time** measured at the probe node: from `connect_start`
to `response_received`. DNS lookup is included. TLS handshake is included.

For streaming probes, we report both TTFT (time to first token) and total
completion time as separate metrics.

### 5.4 Error Types

Failed probes are classified into exactly one of these categories:

| Type | Meaning |
|---|---|
| `timeout` | No response received within 30 seconds |
| `network` | DNS, TCP, or TLS failure before HTTP response |
| `rate_limit` | HTTP 429 or provider-specific rate limit signal |
| `auth` | HTTP 401 or 403 |
| `5xx` | HTTP 500-599 |
| `4xx` | HTTP 400-499 other than 401/403/429 |
| `content_policy` | Response blocked by provider's safety filter |
| `model_overloaded` | Provider-specific overload signals (e.g., Anthropic 529) |
| `empty_response` | HTTP 200 with empty or malformed body |
| `malformed` | Response does not parse as expected schema |
| `unknown` | Anything not covered above |

When a single response could match multiple categories, we use the
**most specific** applicable type in that order.

### 5.5 "Blocked" Status for China View

Providers shown as "blocked" in the China View are those that our China-based
probes cannot reach due to network-level interference (DNS poisoning, TCP
reset, or connection timeout). This is a **purely empirical observation** and
does not represent a legal or policy judgment.

---

## 6. Incident Detection

Incidents are detected by automated rules applied to the last 10 minutes
of probe data. The current rules are:

### Rule: Major Disruption
Error rate > 50% across all nodes and probe types, sustained for 5+ minutes.
Severity: `critical`

### Rule: Elevated Errors
Error rate > 5% in the last 10 minutes AND baseline error rate (same hour,
past 7 days) < 1%.
Severity: `major`

### Rule: Degraded Latency
p95 latency > 3x the baseline (same hour, past 7 days), sustained for 5+
minutes.
Severity: `minor`

### Rule: Regional Outage
Error rate > 50% from one specific node for 5+ minutes, while other nodes
show normal behavior.
Severity: `minor`

### Deduplication and Resolution

If an incident of the same rule + provider is already open, new matches
update the existing incident rather than creating new ones. Incidents
automatically resolve when the triggering condition has not been met for
10 consecutive minutes.

### Human Review

Incidents of severity `major` or `minor` are held in a review queue for up
to 2 hours, during which a human can edit, downgrade, or reject them before
they appear on the public site.

Incidents of severity `critical` are published immediately, because delay
would defeat their purpose. Humans can still edit them retroactively.

Every published incident includes a `detection_method` field indicating
whether it was auto-detected, manually created, or human-edited.

---

## 7. Baselines and Comparisons

When we report "elevated" or "degraded" metrics, the comparison baseline is:

- **Same hour-of-day** (to account for daily usage patterns)
- **Past 7 days** at that hour (to account for weekly patterns)
- **Median** of those 7 samples (to reduce outlier influence)

We do not use rolling 24-hour averages for baseline calculation, because
that would mask real regressions against the historical norm.

---

## 8. Known Limitations

We are honest about the limits of our methodology:

### 8.1 Seven probes is not "everywhere"
Users accessing providers from locations not in our probe set may observe
different behavior. Our China View specifically captures the mainland
Chinese perspective but does not speak for users on other networks within
China (e.g., behind corporate VPNs).

### 8.2 Paid-tier behavior may differ
We use standard paid tier accounts. Users on Enterprise tiers, with
committed-use discounts, or on beta program access may experience
different reliability.

### 8.3 Model-specific behavior
We probe a limited set of models per provider (typically 1-3). Other models
on the same provider may behave differently. Detailed per-model data is
available on provider detail pages, but we cannot probe every available
model.

### 8.4 Our probes are small
Our medium-inference prompts are ~200 tokens in and ~100 tokens out. Very
long contexts (100K+ tokens) may show different failure modes that our
monitoring does not capture.

### 8.5 Non-English probes
Our standard probe prompts are in English. We have not yet characterized
whether non-English requests have systematically different reliability.
This is planned for a future version.

### 8.6 Time zones and rollouts
Providers sometimes deploy changes region-by-region. A probe from us-west
may hit an old version while a probe from eu-central hits a new version
within the same minute. We record the node origin with every probe, which
lets users filter, but this is a genuine source of variance.

### 8.7 Network conditions between probe node and provider
Latency measurements include network time between our probe nodes and the
provider's endpoints. A slowdown in intermediate ISP infrastructure would
be reflected in our latency data, even if the provider itself is healthy.
We cannot distinguish these cases perfectly.

---

## 9. What's Coming (Planned)

These are **not** currently measured, but planned for future versions:

- **Output quality monitoring**: running golden-set prompts and comparing
  responses over time to detect silent model downgrades
- **Cost accuracy**: comparing billed token counts to observed token counts
- **Longer context probes**: testing behavior at 10K, 100K+ token contexts
- **Multilingual probes**: Chinese, Japanese, Spanish, Arabic
- **Per-feature monitoring**: function calling, vision inputs, tool use

When any of these are added, this document will be updated, and a
changelog entry will be published at https://llmstatus.io/methodology/changelog.

---

## 10. Data Retention and Access

- **Raw probe data**: retained for 90 days
- **Hourly aggregates**: retained permanently
- **Daily aggregates**: retained permanently
- **Incident records**: retained permanently

All historical aggregates are available via our public API at
https://llmstatus.io/api/v1. No authentication is required for read
access.

Monthly CSV snapshots of aggregate data are published at
https://llmstatus.io/data for researchers and journalists.

---

## 11. Conflicts of Interest

The maintainers of llmstatus.io may also operate AI services that are
themselves monitored on this site. Any such service is disclosed in the
provider listing with a small "(self-monitored)" note, so users are aware
of the relationship.

We handle this conflict as follows:

- Self-monitored services are probed using **the exact same methodology**
  as every other provider.
- Their incidents are published under **the exact same rules**. If a
  self-monitored service is degraded, it shows as degraded.
- The "(self-monitored)" label is the only distinction.
- No provider — including any self-monitored service — has the ability
  to delay, edit, or suppress their own incident pages.

If you observe llmstatus.io displaying data that appears to favor a
self-monitored service over other providers, please contact
`methodology@llmstatus.io` immediately. This is a failure of our core
mission and we will investigate and correct.

---

## 12. Provider Communication

If you are a representative of a provider we monitor and you have a
concern — factual errors, methodology disputes, or a request — please
contact `methodology@llmstatus.io`.

We commit to:

- Responding to provider correspondence within 48 hours
- Publishing corrections if we are factually wrong, with a clear note of
  what changed and when
- Considering methodology improvements based on feedback
- Describing our reasoning openly if we disagree with a requested change

We do **not**:

- Remove data based on reputational concerns alone
- Delay publication of incidents on request
- Offer "embargo" agreements or pre-publication review
- Modify our measurement frequency, targets, or calculation at provider request

---

## 13. Source Code

The core probe logic — adapters, schedulers, error classifiers — is
published at https://github.com/llmstatus/llmstatus.

You can read exactly what our probes send and how we classify responses.
You are welcome to fork, study, and improve.

---

## 14. Contact

- **Methodology questions**: methodology@llmstatus.io
- **Provider concerns**: methodology@llmstatus.io (same)
- **Security issues**: security@llmstatus.io
- **Press and research**: press@llmstatus.io

---

## 15. Changelog

| Version | Date | Change |
|---|---|---|
| 0.1 | 2026-04-17 | Initial publication |

---

This methodology document is version-controlled. Every substantive change
is recorded here. You can view the full history at
https://github.com/llmstatus/llmstatus/commits/main/METHODOLOGY.md.
