# llmstatus.io

> Independent real-time monitoring for the AI infrastructure.

[![Anthropic status](https://llmstatus.io/badge/anthropic.svg)](https://llmstatus.io/providers/anthropic)
[![OpenAI status](https://llmstatus.io/badge/openai.svg)](https://llmstatus.io/providers/openai)
[![Google Gemini status](https://llmstatus.io/badge/google_gemini.svg)](https://llmstatus.io/providers/google_gemini)
[![DeepSeek status](https://llmstatus.io/badge/deepseek.svg)](https://llmstatus.io/providers/deepseek)
[![Mistral AI status](https://llmstatus.io/badge/mistral.svg)](https://llmstatus.io/providers/mistral)
[![Groq status](https://llmstatus.io/badge/groq.svg)](https://llmstatus.io/providers/groq)

`[` **[llmstatus.io](https://llmstatus.io)** `]` observes 40+ AI API providers
in real time. We measure what actually happens when you call these APIs,
from seven geographic locations — we do not scrape official status pages.

This is a public project. Data is free to access, embed, and cite.

---

## What it is

- **Real-time monitoring** of 20+ AI API providers
- **7 geographic probe nodes** including two in mainland China
- **Independent measurement** — no scraping, no crowdsourcing, only direct probes
- **Permanent historical data** — uptime, latency, incidents all stored forever
- **Public API and RSS feeds** for anyone to consume
- **Open source probe logic** for auditability

## What it is NOT

- Not a SaaS — there is no paid tier, no user accounts in V1
- Not a wrapper around official status pages
- Not a general-purpose monitoring tool (use Datadog for that)
- Not a replacement for your own observability

---

## Quick Links

- 🌐 Website: [llmstatus.io](https://llmstatus.io)
- 📖 Methodology: [llmstatus.io/methodology](https://llmstatus.io/methodology)
- 🔌 Public API: [llmstatus.io/api](https://llmstatus.io/api)
- 🇨🇳 China View: [llmstatus.io/china](https://llmstatus.io/china)
- 📰 RSS feed: [llmstatus.io/feed.xml](https://llmstatus.io/feed.xml)

---

## For Developers

### Embed a status badge

Show any provider's current status in your README or docs:

```markdown
![OpenAI Status](https://llmstatus.io/badge/openai.svg)
```

### Use our public API

```bash
# Current status of all providers
curl https://llmstatus.io/api/v1/providers

# Historical uptime for one provider
curl 'https://llmstatus.io/api/v1/providers/openai/history?window=30d&metric=uptime'

# Recent incidents
curl https://llmstatus.io/api/v1/incidents
```

No authentication required for read access. Rate limited to 60 requests
per minute per IP.

### Subscribe to incidents via RSS

```
https://llmstatus.io/feed.xml                      # all providers
https://llmstatus.io/providers/openai/feed.xml     # single provider
```

### Webhooks

Sign up to receive webhook notifications when specific providers are down:

```
https://llmstatus.io/subscribe
```

---

## Running Locally

Requirements: Docker + Docker Compose.

```bash
git clone https://github.com/llmstatus/llmstatus.git
cd llmstatus
cp .env.example .env
# Edit .env — at minimum, add one provider API key
docker compose up -d
open http://localhost:3000
```

See `CONTRIBUTING.md` and `docs/ROADMAP.md` for development documentation.

---

## Contributing

We welcome contributions, especially new provider adapters.

Please read `CONTRIBUTING.md` before starting.
For security issues, see `SECURITY.md`.

---

## Methodology

Our measurement methodology is fully documented at
[llmstatus.io/methodology](https://llmstatus.io/methodology) and in
`METHODOLOGY.md` in this repo.

Key principles:

1. **We use paid API accounts** — no free trials, no shared keys
2. **We probe every 30-60 seconds** from 7 geographic nodes
3. **We classify errors via a fixed taxonomy** — not ad-hoc labels
4. **We publish everything** — probe logic, methodology, raw data
5. **We monitor ourselves** — any service operated by the maintainers is probed like any other provider

---

## Who builds this

llmstatus.io is built and operated by the `llmstatus.io` maintainers.

The maintainers may also operate AI services that are themselves
monitored on this site. We address this conflict of interest transparently
in `METHODOLOGY.md` §11 — any self-monitored service is probed with the
same methodology and its incidents are published under the same rules as
any other provider.

---

## License

Source code: [Apache License 2.0](LICENSE).
Data and methodology documents (`METHODOLOGY.md`, `docs/known-quirks.md`,
and any aggregated data published at llmstatus.io): [CC BY 4.0](LICENSE-DATA)
— you may use, share, and adapt our data with attribution.

---

## Contact

- General: `contact@llmstatus.io`
- Methodology and provider concerns: `methodology@llmstatus.io`
- Security: `security@llmstatus.io`
- Press: `press@llmstatus.io`
