import type { Metadata } from "next";
import Link from "next/link";

export const dynamic = "force-static";

export const metadata: Metadata = {
  title: "Terms of Service",
  description: "Terms of Service for llmstatus.io — the rules for using this monitoring service.",
  openGraph: {
    title: "Terms of Service — llmstatus.io",
    description: "Terms of Service for llmstatus.io.",
  },
};

const EFFECTIVE_DATE = "2026-04-25";

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

export default function TosPage() {
  return (
    <main className="flex-1 mx-auto w-full max-w-2xl px-6 py-12">
      <div className="mb-12">
        <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-[var(--signal-amber)] mb-4">
          llmstatus.io
        </p>
        <h1 className="text-3xl font-semibold text-[var(--ink-100)] leading-tight mb-4">
          Terms of Service
        </h1>
        <p className="text-sm text-[var(--ink-400)]">
          Effective: {EFFECTIVE_DATE}
        </p>
      </div>

      <Section title="Acceptance">
        <p>
          By accessing llmstatus.io or using any of its APIs, you agree to these Terms of Service.
          If you do not agree, do not use the service.
        </p>
        <p>
          These terms apply to all visitors, registered users, and API consumers.
        </p>
      </Section>

      <Section title="The service">
        <p>
          llmstatus.io is an independent monitoring service that measures the uptime, latency, and
          reliability of AI API providers by making real inference calls from multiple geographic
          locations. The data is published as a public good.
        </p>
        <p>
          The service is provided free of charge for public access. A sponsor programme exists for
          organisations that wish to support the project and receive additional visibility.
        </p>
        <p>
          llmstatus.io is not affiliated with, endorsed by, or in any commercial relationship with
          any of the AI providers it monitors.
        </p>
      </Section>

      <Section title="Accounts">
        <p>
          Creating an account is optional. You need an account only to manage email alert
          subscriptions or to participate in the sponsor programme.
        </p>
        <p>
          You are responsible for maintaining the security of your session. You must not share your
          account credentials or allow others to act under your account.
        </p>
        <p>
          We reserve the right to suspend or delete accounts that violate these terms or that we
          reasonably believe are being used for abuse.
        </p>
      </Section>

      <Section title="Acceptable use">
        <p>You may use llmstatus.io for any lawful purpose. You must not:</p>
        <ul className="space-y-2 list-none mt-2">
          {[
            "Attempt to scrape, mirror, or republish the monitoring data in a way that misrepresents its origin or methodology.",
            "Use the service to make claims about a provider that go beyond what the data shows.",
            "Attempt to interfere with, overload, or reverse-engineer the service infrastructure.",
            "Use automated access to the public API at rates that materially harm other users — the public API is rate-limited; stay within the published limits.",
            "Create accounts for the purpose of spam, harassment, or abuse.",
          ].map((item) => (
            <li key={item} className="flex items-start gap-2">
              <span className="mt-0.5 text-[var(--signal-down)] shrink-0">✕</span>
              <span>{item}</span>
            </li>
          ))}
        </ul>
      </Section>

      <Section title="Data and attribution">
        <p>
          All monitoring data published by llmstatus.io is licensed under{" "}
          <a
            href="https://creativecommons.org/licenses/by/4.0/"
            target="_blank"
            rel="noopener noreferrer"
            className="text-[var(--ink-300)] hover:text-[var(--ink-100)] transition-colors underline underline-offset-2"
          >
            CC BY 4.0
          </a>
          . You may use, share, and adapt the data for any purpose, including commercial use,
          provided you give appropriate credit to llmstatus.io and link to the source.
        </p>
        <p>
          Attribution format:{" "}
          <span className="font-mono text-xs text-[var(--ink-300)] bg-[var(--canvas-raised)] px-1 py-0.5">
            Source: llmstatus.io (CC BY 4.0)
          </span>
        </p>
        <p>
          The source code is licensed under Apache 2.0 and available on{" "}
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

      <Section title="Privacy">
        <p>
          Our{" "}
          <Link
            href="/privacy"
            className="text-[var(--ink-300)] hover:text-[var(--ink-100)] transition-colors underline underline-offset-2"
          >
            Privacy Policy
          </Link>{" "}
          describes what personal data we collect, how we use it, and how to request deletion.
          By using the service, you also agree to that policy.
        </p>
      </Section>

      <Section title="Disclaimers">
        <p>
          <strong className="text-[var(--ink-100)]">No warranty.</strong> llmstatus.io is provided
          "as is" and "as available." We make no warranties, express or implied, regarding accuracy,
          completeness, availability, or fitness for a particular purpose.
        </p>
        <p>
          <strong className="text-[var(--ink-100)]">Data is observational.</strong> Probe results
          reflect what our infrastructure observed from specific locations at specific times. They
          may not reflect all users' experience and should not be used as the sole basis for legal
          or contractual claims against a provider.
        </p>
        <p>
          <strong className="text-[var(--ink-100)]">Service availability.</strong> We aim for high
          availability but do not guarantee uninterrupted access. Maintenance windows and
          infrastructure incidents may cause gaps in data.
        </p>
      </Section>

      <Section title="Limitation of liability">
        <p>
          To the maximum extent permitted by applicable law, llmstatus.io and its maintainers
          shall not be liable for any indirect, incidental, special, or consequential damages
          arising from your use of or reliance on the service or its data, even if we have been
          advised of the possibility of such damages.
        </p>
        <p>
          Our total liability to you for any claim arising out of these terms shall not exceed
          USD 100.
        </p>
      </Section>

      <Section title="Changes to the service and these terms">
        <p>
          We may change the service, its features, or these terms at any time. We will update the
          effective date above when terms change. Continued use of the service after a change
          constitutes acceptance of the new terms.
        </p>
        <p>
          We will provide reasonable notice of material changes by posting a notice on the site.
          For changes that materially affect registered users, we will also send an email.
        </p>
      </Section>

      <Section title="Governing law">
        <p>
          These terms are governed by the laws of the jurisdiction in which the service operator
          is registered. Disputes will be resolved in the courts of that jurisdiction, unless
          applicable law requires otherwise.
        </p>
      </Section>

      <Section title="Contact">
        <p>
          Questions about these terms:{" "}
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
