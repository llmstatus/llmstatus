import type { Metadata } from "next";
import Link from "next/link";

export const dynamic = "force-static";

export const metadata: Metadata = {
  title: "About",
  description:
    "llmstatus.io is an independent real-time monitoring service for AI API providers. " +
    "We make real API calls from 7 global locations — not scraped from official status pages.",
  openGraph: {
    title: "About — llmstatus.io",
    description:
      "Independent real-time monitoring for the AI infrastructure. " +
      "Measured from 7 global locations. Not scraped from official status pages.",
  },
};

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="mb-10">
      <h2 className="mb-4 text-sm font-semibold uppercase tracking-wide text-[var(--ink-300)]">
        {title}
      </h2>
      <div className="space-y-3 text-sm text-[var(--ink-200)] leading-6">{children}</div>
    </section>
  );
}

export default function AboutPage() {
  return (
    <main className="flex-1 mx-auto w-full max-w-2xl px-6 py-12">
      {/* Header */}
      <div className="mb-12">
        <p className="text-xs font-semibold uppercase tracking-[0.12em] text-[var(--signal-amber)] mb-4">
          llmstatus.io
        </p>
        <h1 className="text-3xl font-semibold text-[var(--ink-100)] leading-tight mb-4">
          Independent monitoring
          <br />
          for the AI infrastructure.
        </h1>
        <p className="text-sm text-[var(--ink-400)] leading-relaxed">
          We make real API calls from 7 global locations every 30–60 seconds
          and publish what we find — uptime, latency, and incidents — as a public good.
        </p>
      </div>

      <Section title="What this is">
        <p>
          llmstatus.io is a third-party monitoring service for AI API providers.
          We measure uptime, latency, and reliability by sending real inference
          requests from multiple geographic regions — not by reading official status
          pages or aggregating user reports.
        </p>
        <p>
          When you see a provider listed as operational on their own status page
          while your application is failing, llmstatus.io is the independent data
          source you can point to.
        </p>
      </Section>

      <Section title="Why we built it">
        <p>
          Official status pages are operated by the same teams experiencing the
          incidents. Historically, updates lag reality by 15–60 minutes on average.
          Crowdsourced reports like Downdetector reflect consumer sentiment, not
          API behavior.
        </p>
        <p>
          Developers and infrastructure teams need accurate, timely, independent
          data to set SLAs, triage incidents, and make provider decisions. That
          data did not exist for AI APIs. We built it.
        </p>
      </Section>

      <Section title="How we work">
        <p>
          Every 30–60 seconds, probes run from 7 independent server locations
          across North America, Europe, and Asia-Pacific. Each probe makes a real
          authenticated API call — a short inference request — and records whether
          it succeeded, how long it took, and what error type occurred if it failed.
        </p>
        <p>
          An incident is declared when our detection rules fire based on aggregate
          probe data. Rules are public and documented in full on the{" "}
          <Link
            href="/methodology"
            className="text-[var(--ink-300)] hover:text-[var(--ink-100)] transition-colors underline underline-offset-2"
          >
            Methodology
          </Link>{" "}
          page.
        </p>
      </Section>

      <Section title="Independence">
        <p>
          We pay for our own API keys. We have no commercial relationship with
          any provider we monitor. No provider can influence what we publish.
        </p>
        <p>
          If our own infrastructure has an issue, we publish it with the same
          rules as any other provider. Our source code — including all probe
          logic and detection rules — is public on{" "}
          <a
            href="https://github.com/llmstatus/llmstatus"
            target="_blank"
            rel="noopener noreferrer"
            className="text-[var(--ink-300)] hover:text-[var(--ink-100)] transition-colors underline underline-offset-2"
          >
            GitHub
          </a>
          .
        </p>
      </Section>

      <Section title="What we don't do">
        <ul className="space-y-2 list-none">
          {[
            "Scrape official status pages.",
            "Use crowdsourced incident reports.",
            "Make up or infer data we didn't measure.",
            "Silently drop anomalies we don't understand.",
            "Delete historical data.",
          ].map((item) => (
            <li key={item} className="flex items-start gap-2">
              <span className="mt-0.5 text-[var(--signal-down)] shrink-0">✕</span>
              <span>{item}</span>
            </li>
          ))}
        </ul>
      </Section>

      <Section title="Contact">
        <p>
          For questions about our methodology or data:{" "}
          <a
            href="mailto:methodology@llmstatus.io"
            className="text-[var(--ink-300)] hover:text-[var(--ink-100)] transition-colors"
          >
            methodology@llmstatus.io
          </a>
        </p>
        <p>
          To report a security issue:{" "}
          <a
            href="mailto:security@llmstatus.io"
            className="text-[var(--ink-300)] hover:text-[var(--ink-100)] transition-colors"
          >
            security@llmstatus.io
          </a>
        </p>
        <p>
          General:{" "}
          <a
            href="mailto:contact@llmstatus.io"
            className="text-[var(--ink-300)] hover:text-[var(--ink-100)] transition-colors"
          >
            contact@llmstatus.io
          </a>
        </p>
      </Section>
    </main>
  );
}
