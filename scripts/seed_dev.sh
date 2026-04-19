#!/usr/bin/env bash
# scripts/seed_dev.sh — seed the local dev database with realistic mock data.
#
# Usage:
#   bash scripts/seed_dev.sh
#
# Requires: docker (running llmstatus-db), python3
# Idempotent: safe to run multiple times (ON CONFLICT DO NOTHING).
set -euo pipefail

INFLUX_HOST="${INFLUX_HOST:-http://localhost:18086}"
INFLUX_DB="${INFLUX_DATABASE:-llmstatus}"

# ── PostgreSQL ────────────────────────────────────────────────────────────────
echo "==> Seeding PostgreSQL …"

SQL_FILE=$(mktemp /tmp/seed_dev_XXXXXX.sql)
trap 'rm -f "$SQL_FILE"' EXIT

cat > "$SQL_FILE" <<'SQL'
-- providers
INSERT INTO providers (id, name, category, base_url, auth_type, status_page_url, documentation_url, region, active)
VALUES
  ('openai',    'OpenAI',        'official',   'https://api.openai.com',
   'bearer',         'https://status.openai.com',       'https://platform.openai.com/docs',   'global', true),
  ('anthropic', 'Anthropic',     'official',   'https://api.anthropic.com',
   'api_key_header', 'https://status.anthropic.com',    'https://docs.anthropic.com',         'global', true),
  ('google',    'Google AI',     'official',   'https://generativelanguage.googleapis.com',
   'api_key_header', 'https://status.cloud.google.com', 'https://ai.google.dev/docs',         'global', true),
  ('mistral',   'Mistral AI',    'official',   'https://api.mistral.ai',
   'bearer',         'https://status.mistral.ai',       'https://docs.mistral.ai',            'global', true),
  ('cohere',    'Cohere',        'official',   'https://api.cohere.com',
   'bearer',         'https://status.cohere.com',       'https://docs.cohere.com',            'global', true),
  ('together',  'Together AI',   'aggregator', 'https://api.together.xyz',
   'bearer',         NULL,                               'https://docs.together.ai',           'global', true),
  ('groq',      'Groq',          'official',   'https://api.groq.com',
   'bearer',         NULL,                               'https://console.groq.com/docs',      'global', true)
ON CONFLICT (id) DO NOTHING;

-- models
INSERT INTO models (provider_id, model_id, display_name, model_type, active) VALUES
  ('openai',    'gpt-4o',                                      'GPT-4o',               'chat', true),
  ('openai',    'gpt-4o-mini',                                 'GPT-4o mini',          'chat', true),
  ('openai',    'gpt-4-turbo',                                 'GPT-4 Turbo',          'chat', true),
  ('anthropic', 'claude-3-5-sonnet-20241022',                  'Claude 3.5 Sonnet',    'chat', true),
  ('anthropic', 'claude-3-5-haiku-20241022',                   'Claude 3.5 Haiku',     'chat', true),
  ('anthropic', 'claude-haiku-4-5-20251001',                   'Claude Haiku 4.5',     'chat', true),
  ('anthropic', 'claude-3-opus-20240229',                      'Claude 3 Opus',        'chat', true),
  ('google',    'gemini-2.0-flash',                            'Gemini 2.0 Flash',     'chat', true),
  ('google',    'gemini-1.5-pro',                              'Gemini 1.5 Pro',       'chat', true),
  ('google',    'gemini-1.5-flash',                            'Gemini 1.5 Flash',     'chat', true),
  ('mistral',   'mistral-large-latest',                        'Mistral Large',        'chat', true),
  ('mistral',   'mistral-small-latest',                        'Mistral Small',        'chat', true),
  ('cohere',    'command-r-plus',                              'Command R+',           'chat', true),
  ('cohere',    'command-r',                                   'Command R',            'chat', true),
  ('together',  'meta-llama/Meta-Llama-3.1-8B-Instruct-Turbo','Llama 3.1 8B Turbo',  'chat', true),
  ('together',  'meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo','Llama 3.1 70B Turbo','chat', true),
  ('groq',      'llama-3.1-8b-instant',                        'Llama 3.1 8B Instant','chat', true),
  ('groq',      'mixtral-8x7b-32768',                          'Mixtral 8x7B',        'chat', true)
ON CONFLICT (provider_id, model_id) DO NOTHING;

-- incidents
INSERT INTO incidents (id, slug, provider_id, severity, title, description, status,
                       affected_models, affected_regions, started_at, resolved_at,
                       detection_method, detection_rule, human_reviewed)
VALUES
  (gen_random_uuid(),
   'openai-2026-04-16-elevated-errors', 'openai', 'major',
   'Elevated error rates on GPT-4o',
   'Probe error rate exceeded 15% threshold for gpt-4o. HTTP 500 responses from api.openai.com.',
   'resolved',
   ARRAY['gpt-4o'], ARRAY['us','eu'],
   NOW() - INTERVAL '3 days 4 hours', NOW() - INTERVAL '3 days 1 hour',
   'auto', 'error_rate_15pct_5min', true),

  (gen_random_uuid(),
   'anthropic-2026-04-18-high-latency', 'anthropic', 'minor',
   'Elevated latency on Claude 3.5 Sonnet',
   'P95 latency exceeded 8000ms for claude-3-5-sonnet-20241022 across all regions.',
   'resolved',
   ARRAY['claude-3-5-sonnet-20241022'], ARRAY['global'],
   NOW() - INTERVAL '22 hours', NOW() - INTERVAL '19 hours',
   'auto', 'p95_latency_8000ms_10min', false),

  (gen_random_uuid(),
   'mistral-2026-04-19-degraded', 'mistral', 'minor',
   'Elevated error rate — Mistral Large',
   'Error rate on mistral-large-latest has been above 25% for the past hour. Requests returning HTTP 503.',
   'ongoing',
   ARRAY['mistral-large-latest'], ARRAY['global'],
   NOW() - INTERVAL '90 minutes', NULL,
   'auto', 'error_rate_15pct_5min', false)
ON CONFLICT (slug) DO NOTHING;
SQL

docker cp "$SQL_FILE" llmstatus-db:/tmp/seed_dev.sql
docker exec llmstatus-db psql -U llmstatus llmstatus -f /tmp/seed_dev.sql
echo "==> PostgreSQL seeded."

# ── InfluxDB ──────────────────────────────────────────────────────────────────
echo "==> Generating InfluxDB probe data (48 h × 11 provider/model pairs) …"

INFLUX_HOST="$INFLUX_HOST" INFLUX_DB="$INFLUX_DB" python3 - <<'PYEOF'
import urllib.request
import urllib.error
import random
import time
import sys
import os

INFLUX_HOST = os.environ["INFLUX_HOST"]
INFLUX_DB   = os.environ["INFLUX_DB"]

# (provider_id, model, success_rate, p50_ms, stddev_ms)
PROVIDERS = [
    ("openai",    "gpt-4o",                       0.993, 380, 100),
    ("openai",    "gpt-4o-mini",                   0.994, 240,  70),
    ("anthropic", "claude-3-5-sonnet-20241022",    0.991, 650, 180),
    ("anthropic", "claude-3-5-haiku-20241022",     0.993, 300,  90),
    ("anthropic", "claude-haiku-4-5-20251001",     0.993, 280,  75),
    ("google",    "gemini-2.0-flash",              0.987, 420, 110),
    ("google",    "gemini-1.5-pro",                0.985, 720, 200),
    ("mistral",   "mistral-large-latest",          0.68,  1400, 600),
    ("mistral",   "mistral-small-latest",          0.71,   900, 350),
    ("cohere",    "command-r-plus",                0.988,  510, 140),
    ("together",  "meta-llama/Meta-Llama-3.1-8B-Instruct-Turbo", 0.980, 310, 90),
    ("groq",      "llama-3.1-8b-instant",          0.992,  120,  35),
]

REGION      = "local-dev"
PROBE_TYPE  = "chat_basic"
INTERVAL_S  = 300         # 5-min samples
WINDOW_S    = 48 * 3600   # 48 h of history
BATCH_LINES = 5000

now_ns = int(time.time() * 1_000_000_000)
steps  = WINDOW_S // INTERVAL_S
url    = f"{INFLUX_HOST}/api/v2/write?precision=ns&bucket={INFLUX_DB}"
hdrs   = {"Content-Type": "text/plain; charset=utf-8"}

# Incident windows (ns)
OPENAI_OUTAGE_S  = now_ns - int(3.17 * 24 * 3600 * 1e9)
OPENAI_OUTAGE_E  = now_ns - int(3.04 * 24 * 3600 * 1e9)
ANTHRO_LATENCY_S = now_ns - int(22 * 3600 * 1e9)
ANTHRO_LATENCY_E = now_ns - int(19 * 3600 * 1e9)

def esc(s):
    return s.replace(",", r"\,").replace(" ", r"\ ").replace("=", r"\=")

def make_line(pid, model, ts, sr, p50, sd):
    _sr, _p50, _sd = sr, p50, sd
    if pid == "openai" and model == "gpt-4o" and OPENAI_OUTAGE_S <= ts <= OPENAI_OUTAGE_E:
        _sr, _p50, _sd = 0.15, 800, 300
    if pid == "anthropic" and "sonnet" in model and ANTHRO_LATENCY_S <= ts <= ANTHRO_LATENCY_E:
        _sr, _p50, _sd = 0.90, 9000, 2000

    ok = random.random() < _sr
    dur    = max(50,  int(random.gauss(_p50, _sd)))  if ok else max(100, int(random.gauss(_p50 * 1.5, _sd * 2)))
    status = 200                                       if ok else random.choices([429, 500, 503], weights=[3, 4, 3])[0]
    ecls   = ""                                        if ok else ("rate_limit" if status == 429 else "server_error")

    tags = f"provider_id={esc(pid)},model={esc(model)},probe_type={PROBE_TYPE},region_id={REGION}"
    if ecls:
        tags += f",error_class={ecls}"

    ok_s   = "true" if ok else "false"
    fields = f"success={ok_s},duration_ms={dur}i,http_status={status}i"
    if ok:
        fields += f",tokens_in=50i,tokens_out={random.randint(80, 350)}i"

    return f"probes,{tags} {fields} {ts}"

buf, total = [], 0

def flush():
    global total
    payload = "\n".join(buf).encode()
    req = urllib.request.Request(url, data=payload, headers=hdrs, method="POST")
    try:
        urllib.request.urlopen(req)
        total += len(buf)
    except urllib.error.HTTPError as e:
        print(f"  WARN: {e.code} {e.read(256).decode(errors='replace')}", file=sys.stderr)

for step in range(steps):
    ts = now_ns - (steps - step) * INTERVAL_S * 1_000_000_000
    ts += random.randint(-30_000_000_000, 30_000_000_000)
    for (pid, model, sr, p50, sd) in PROVIDERS:
        buf.append(make_line(pid, model, ts, sr, p50, sd))
        if len(buf) >= BATCH_LINES:
            flush(); buf.clear()
            print(f"  flushed {total} points …")

if buf:
    flush()

print(f"  done — {total} points written.")
PYEOF

echo ""
echo "==> Dev seed complete."
echo "    Providers : openai, anthropic, google, mistral (degraded), cohere, together, groq"
echo "    Incidents : 1 ongoing (mistral), 2 resolved (openai, anthropic)"
echo "    Probe data: 48 h × 11 provider/model pairs, 5-min intervals"
