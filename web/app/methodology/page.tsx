import type { Metadata } from "next";
import Link from "next/link";

export const dynamic = "force-static";

export const metadata: Metadata = {
  title: "Methodology",
  description:
    "How llmstatus.io measures AI API uptime, latency, and incidents. " +
    "Real probe calls from 7 global locations — every metric is reproducible.",
  openGraph: {
    title: "Methodology — llmstatus.io",
    description:
      "How we measure AI API uptime, latency, and incidents — " +
      "every metric is traceable to specific probe logic.",
  },
};

function Section({ id, title, children }: { id: string; title: string; children: React.ReactNode }) {
  return (
    <section id={id} className="mb-12 scroll-mt-6">
      <h2 className="mb-4 text-sm font-semibold uppercase tracking-wide text-[var(--ink-300)]">
        {title}
      </h2>
      <div className="space-y-3 text-sm text-[var(--ink-200)] leading-6">{children}</div>
    </section>
  );
}

function Rule({
  id,
  severity,
  trigger,
  title,
}: {
  id: string;
  severity: "critical" | "major" | "minor";
  trigger: string;
  title: string;
}) {
  const color =
    severity === "critical"
      ? "text-[var(--signal-down)]"
      : severity === "major"
      ? "text-[var(--signal-warn)]"
      : "text-[var(--ink-300)]";
  return (
    <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-4 py-3">
      <div className="flex items-center gap-3 mb-1">
        <span className="text-xs font-mono text-[var(--ink-500)]">{id}</span>
        <span className={`text-xs font-semibold uppercase tracking-wide ${color}`}>{severity}</span>
      </div>
      <p className="text-sm font-medium text-[var(--ink-100)] mb-1">{title}</p>
      <p className="text-xs text-[var(--ink-400)]">{trigger}</p>
    </div>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-start gap-4">
      <span className="w-32 shrink-0 text-xs text-[var(--ink-400)]">{label}</span>
      <span className="text-sm text-[var(--ink-200)]">{value}</span>
    </div>
  );
}

export default function MethodologyPage() {
  return (
    <main className="flex-1 mx-auto w-full max-w-2xl px-6 py-12">
      {/* Header */}
      <div className="mb-12">
        <p className="text-xs font-semibold uppercase tracking-[0.12em] text-[var(--signal-amber)] mb-4">
          Methodology
        </p>
        <h1 className="text-3xl font-semibold text-[var(--ink-100)] leading-tight mb-4">
          How we measure.
        </h1>
        <p className="text-sm text-[var(--ink-400)] leading-relaxed">
          Every number on llmstatus.io is traceable to specific probe logic. This page
          documents exactly what we do, how we classify failures, and how we decide
          when an incident is real.
        </p>
        <p className="mt-3 text-xs text-[var(--ink-500)]">
          The complete specification is in{" "}
          <a
            href="https://github.com/llmstatus/llmstatus/blob/main/METHODOLOGY.md"
            target="_blank"
            rel="noopener noreferrer"
            className="text-[var(--ink-400)] hover:text-[var(--ink-200)] transition-colors underline underline-offset-2"
          >
            METHODOLOGY.md
          </a>{" "}
          on GitHub. This page is the human-readable summary.
        </p>
      </div>

      <Section id="probes" title="How we probe">
        <p>
          From each of 7 server locations, we run two probe types against every
          monitored provider on a fixed schedule:
        </p>
        <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] divide-y divide-[var(--ink-600)] overflow-hidden">
          <div className="px-4 py-3 grid grid-cols-[6rem_1fr] gap-4">
            <span className="text-xs font-semibold text-[var(--ink-400)]">Light inference</span>
            <span className="text-xs text-[var(--ink-200)]">
              Sends a minimal prompt (&quot;Reply with OK&quot;), expects ≤10 token response. Runs every 60 s.
              This is the primary uptime signal.
            </span>
          </div>
          <div className="px-4 py-3 grid grid-cols-[6rem_1fr] gap-4">
            <span className="text-xs font-semibold text-[var(--ink-400)]">Medium inference</span>
            <span className="text-xs text-[var(--ink-200)]">
              ~200-token input, requests ~100-token output. Runs every 5 min.
              This is the latency signal.
            </span>
          </div>
        </div>
        <p>
          We use dedicated paid API accounts for monitoring. We are paying customers
          of every provider we monitor — we do not free-ride on trial accounts.
        </p>
      </Section>

      <Section id="locations" title="Probe locations">
        <div className="space-y-1.5">
          {[
            ["us-west-2", "AWS Oregon — near OpenAI/Anthropic origins"],
            ["us-east-1", "AWS Virginia — near AWS Bedrock, Google"],
            ["eu-west", "Hetzner Germany — EU coverage"],
            ["ap-northeast-1", "AWS Tokyo — APAC"],
            ["ap-southeast-1", "AWS Singapore — Southeast Asia"],
            ["cn-shanghai", "Alibaba Cloud Shanghai — China view"],
            ["cn-guangzhou", "Tencent Cloud Guangzhou — China redundancy"],
          ].map(([region, note]) => (
            <div key={region} className="flex items-start gap-3 text-xs">
              <span className="font-mono text-[var(--ink-400)] w-36 shrink-0">{region}</span>
              <span className="text-[var(--ink-300)]">{note}</span>
            </div>
          ))}
        </div>
      </Section>

      <Section id="what-we-measure" title="What we measure">
        <div className="space-y-3">
          <Metric label="Uptime" value="Fraction of probe attempts that succeed in a given window." />
          <Metric label="P95 latency" value="95th-percentile wall-clock time to receive the complete response, across successful probes only." />
          <Metric label="Error type" value="When a probe fails, we classify the failure: timeout, network, rate_limit, auth, 5xx, 4xx, content_policy, model_overloaded, empty_response, malformed, or unknown." />
        </div>
        <p className="mt-2">
          Rate-limit errors (429) are counted in the error rate. If a provider is
          consistently returning 429, that is a real service-availability issue
          from the customer perspective.
        </p>
      </Section>

      <Section id="detection-rules" title="Detection rules">
        <p>
          The detector runs every 60 seconds and applies four rules to the most
          recent probe data. All rules and thresholds are public in{" "}
          <a
            href="https://github.com/llmstatus/llmstatus/blob/main/internal/detector/rules.go"
            target="_blank"
            rel="noopener noreferrer"
            className="text-[var(--ink-400)] hover:text-[var(--ink-200)] transition-colors underline underline-offset-2"
          >
            rules.go
          </a>
          .
        </p>
        <div className="space-y-3">
          <Rule
            id="Rule 6.1"
            severity="critical"
            title="Provider is DOWN"
            trigger="Error rate > 50% across all nodes in the last 5 minutes, with ≥ 3 probes."
          />
          <Rule
            id="Rule 6.2"
            severity="major"
            title="Elevated error rate"
            trigger="Error rate > 5% in the last 10 minutes. Suppressed if Rule 6.1 is already firing."
          />
          <Rule
            id="Rule 6.3"
            severity="minor"
            title="Latency degradation"
            trigger="P95 latency over the last 5 minutes exceeds 3× the 24-hour baseline, with ≥ 5 successful samples."
          />
          <Rule
            id="Rule 6.4"
            severity="minor"
            title="Regional outage"
            trigger="One probe region exceeds 50% error rate while the provider is not globally down, with ≥ 3 probes from that region."
          />
        </div>
      </Section>

      <Section id="deduplication" title="Deduplication and resolution">
        <p>
          If an incident of the same rule and provider is already ongoing, we do
          not create a duplicate — we update the existing record.
        </p>
        <p>
          An incident auto-resolves when its triggering rule has not fired for
          10 consecutive minutes. Critical incidents are published immediately.
          Minor and major incidents are also published immediately by default;
          a human reviewer can edit the title and description after publication.
        </p>
      </Section>

      <Section id="data-retention" title="Data retention">
        <div className="space-y-3">
          <Metric label="Raw probe data" value="90 days, then rolled up into hourly aggregates." />
          <Metric label="Hourly aggregates" value="Permanent — never deleted." />
          <Metric label="Daily aggregates" value="Permanent — never deleted." />
          <Metric label="Incident records" value="Permanent. We do not retroactively delete or alter incident history." />
        </div>
      </Section>

      <Section id="limitations" title="Known limitations">
        <ul className="space-y-2">
          {[
            "We probe with a short fixed prompt. Real-world workloads — long contexts, system prompts, function calls — may behave differently.",
            "We cover the chat/completion inference API. Embeddings, image generation, audio, and fine-tuning endpoints are not currently monitored.",
            "Our latency baseline uses a 24-hour rolling window (V1). The METHODOLOGY.md specifies a 7-day same-hour median as the correct baseline; we will upgrade in V2.",
            "Regional results depend on our probe servers' network paths. A problem on one transit provider between our server and the API may look like a regional outage.",
            "We cannot distinguish a global rate-limit (429) from a partial outage if the provider uses 429 for both.",
          ].map((item) => (
            <li key={item} className="flex items-start gap-2 text-sm">
              <span className="mt-0.5 text-[var(--signal-warn)] shrink-0">·</span>
              <span>{item}</span>
            </li>
          ))}
        </ul>
      </Section>

      <div className="border-t border-[var(--ink-600)] pt-6">
        <p className="text-xs text-[var(--ink-500)]">
          Questions or corrections?{" "}
          <a
            href="mailto:methodology@llmstatus.io"
            className="text-[var(--ink-400)] hover:text-[var(--ink-200)] transition-colors"
          >
            methodology@llmstatus.io
          </a>
          {" · "}
          <Link
            href="/about"
            className="text-[var(--ink-400)] hover:text-[var(--ink-200)] transition-colors"
          >
            About
          </Link>
        </p>
      </div>
    </main>
  );
}
