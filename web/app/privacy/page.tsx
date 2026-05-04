import type { Metadata } from "next";
import Link from "next/link";

export const dynamic = "force-static";

export const metadata: Metadata = {
  title: "Privacy Policy",
  description:
    "Privacy Policy for llmstatus.io — what data we collect, how we use it, and your rights.",
  openGraph: {
    title: "Privacy Policy — llmstatus.io",
    description: "What data llmstatus.io collects, how it is used, and how to request deletion.",
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

function Item({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex gap-3">
      <span className="shrink-0 text-[var(--ink-400)] w-40">{label}</span>
      <span>{children}</span>
    </div>
  );
}

export default function PrivacyPage() {
  return (
    <main className="flex-1 mx-auto w-full max-w-2xl px-6 py-12">
      <div className="mb-12">
        <p className="text-xs font-semibold uppercase tracking-[0.12em] text-[var(--signal-amber)] mb-4">
          llmstatus.io
        </p>
        <h1 className="text-3xl font-semibold text-[var(--ink-100)] leading-tight mb-4">
          Privacy Policy
        </h1>
        <p className="text-sm text-[var(--ink-400)]">
          Effective: {EFFECTIVE_DATE}
        </p>
      </div>

      <Section title="Overview">
        <p>
          llmstatus.io is a public monitoring service. Most of the site requires no account and
          collects no personal data. An account is only needed to manage email alert subscriptions
          or to participate in the sponsor programme.
        </p>
        <p>
          This policy describes what personal data we collect, why we collect it, and how you can
          request its deletion.
        </p>
      </Section>

      <Section title="Data we collect">
        <p>
          We collect the minimum data necessary to operate the service.
        </p>

        <div className="border border-[var(--ink-600)] divide-y divide-[var(--ink-600)] mt-4">
          <div className="px-4 py-2 grid grid-cols-3 gap-4 text-xs font-semibold uppercase tracking-wide text-[var(--ink-400)]">
            <span>Data</span>
            <span>Source</span>
            <span>Why</span>
          </div>
          {[
            ["Email address", "You / Google sign-in", "Deliver alert emails; identify your account"],
            ["Google account ID (sub)", "Google sign-in", "Link your Google identity to your account; never exposed publicly"],
            ["GitHub account ID (node_id)", "GitHub sign-in", "Link your GitHub identity to your account; never exposed publicly"],
            ["Session token", "Our servers", "Keep you signed in across requests; HttpOnly cookie"],
            ["Alert subscription settings", "You", "Determine which alerts to send you"],
            ["Sponsor profile (optional)", "You", "Display your logo / link on the Sponsors page"],
          ].map(([data, source, why]) => (
            <div key={data} className="px-4 py-3 grid grid-cols-3 gap-4 text-sm">
              <span className="text-[var(--ink-100)]">{data}</span>
              <span className="text-[var(--ink-300)]">{source}</span>
              <span className="text-[var(--ink-400)]">{why}</span>
            </div>
          ))}
        </div>

        <p className="mt-4">
          We do <strong className="text-[var(--ink-100)]">not</strong> collect your name, profile
          picture, phone number, payment details, location, or any Google data beyond email address
          and account ID. Our Google OAuth request uses only the{" "}
          <code className="text-[var(--ink-300)] bg-[var(--canvas-raised)] px-1 py-0.5 text-xs">openid</code>{" "}
          and{" "}
          <code className="text-[var(--ink-300)] bg-[var(--canvas-raised)] px-1 py-0.5 text-xs">email</code>{" "}
          scopes.
        </p>
      </Section>

      <Section title="How we use your data">
        <div className="space-y-2">
          <Item label="Email address">
            Send incident alert digests you have subscribed to. Transactional emails only —
            no marketing.
          </Item>
          <Item label="Google / GitHub ID">
            Authenticate you on return visits. Not shared with any third party.
          </Item>
          <Item label="Session token">
            Maintain your authenticated session. Expires when you sign out or after 30 days of
            inactivity.
          </Item>
          <Item label="Sponsor profile">
            Displayed publicly on <Link href="/sponsors" className="text-[var(--ink-300)] hover:text-[var(--ink-100)] transition-colors underline underline-offset-2">/sponsors</Link> if
            you have registered. You control all fields and can remove them at any time.
          </Item>
        </div>

        <p className="mt-4">
          We do not sell, rent, or share personal data with third parties for advertising or
          analytics. We do not use Google user data to serve advertisements or to train machine
          learning models.
        </p>
      </Section>

      <Section title="Third-party services">
        <p>
          We use the following third-party services that may process your data:
        </p>
        <div className="space-y-3 mt-2">
          <div>
            <span className="text-[var(--ink-100)]">Google OAuth 2.0</span>
            <span className="text-[var(--ink-400)]"> — authenticates your identity. Governed by{" "}
              <a
                href="https://policies.google.com/privacy"
                target="_blank"
                rel="noopener noreferrer"
                className="text-[var(--ink-300)] hover:text-[var(--ink-100)] transition-colors underline underline-offset-2"
              >
                Google&apos;s Privacy Policy
              </a>.
            </span>
          </div>
          <div>
            <span className="text-[var(--ink-100)]">GitHub OAuth</span>
            <span className="text-[var(--ink-400)]"> — alternative authentication. Governed by{" "}
              <a
                href="https://docs.github.com/en/site-policy/privacy-policies/github-privacy-statement"
                target="_blank"
                rel="noopener noreferrer"
                className="text-[var(--ink-300)] hover:text-[var(--ink-100)] transition-colors underline underline-offset-2"
              >
                GitHub&apos;s Privacy Statement
              </a>.
            </span>
          </div>
          <div>
            <span className="text-[var(--ink-100)]">Resend</span>
            <span className="text-[var(--ink-400)]"> — transactional email delivery. Your email
              address is transmitted to Resend solely to deliver messages you have requested.
            </span>
          </div>
        </div>
        <p className="mt-3">
          We do not use Google Analytics, Facebook Pixel, or any third-party behavioural tracking
          scripts.
        </p>
      </Section>

      <Section title="Data retention">
        <p>
          Account data (email, provider IDs, subscription settings) is retained while your account
          is active. If you request deletion, we remove your personal data within 30 days, except
          where retention is required by law.
        </p>
        <p>
          Alert subscription logs (which alerts were sent, when) are retained for 90 days for
          operational debugging, then purged.
        </p>
        <p>
          Monitoring data (probe results, incidents) is anonymised time-series data with no
          connection to user accounts and is retained indefinitely as public record.
        </p>
      </Section>

      <Section title="Your rights">
        <p>
          You may at any time:
        </p>
        <ul className="space-y-2 list-none mt-2">
          {[
            "Sign in and delete your account from the Account page — this deletes your email, provider IDs, and all subscriptions.",
            "Unsubscribe from all alerts without deleting your account.",
            "Request a copy of the personal data we hold about you.",
            "Request correction of inaccurate data.",
          ].map((item) => (
            <li key={item} className="flex items-start gap-2">
              <span className="mt-0.5 text-[var(--signal-ok)] shrink-0">→</span>
              <span>{item}</span>
            </li>
          ))}
        </ul>
        <p className="mt-3">
          To exercise any right, email{" "}
          <a
            href="mailto:privacy@llmstatus.io"
            className="text-[var(--ink-300)] hover:text-[var(--ink-100)] transition-colors"
          >
            privacy@llmstatus.io
          </a>{" "}
          from the address associated with your account.
        </p>
      </Section>

      <Section title="Cookies">
        <p>
          We set two first-party cookies:
        </p>
        <div className="space-y-2 mt-2">
          <Item label="session">
            HttpOnly, Secure session token. Required to stay signed in. Expires after 30 days of
            inactivity or on sign-out.
          </Item>
          <Item label="oauth_state / oauth_cv">
            Temporary CSRF-protection cookies set during the OAuth flow. Deleted immediately after
            authentication completes. Max-age: 10 minutes.
          </Item>
        </div>
        <p className="mt-3">
          We do not set advertising, analytics, or tracking cookies.
        </p>
      </Section>

      <Section title="Security">
        <p>
          All data in transit is encrypted via HTTPS. Session tokens are stored in HttpOnly,
          Secure, SameSite=Lax cookies and are not accessible to JavaScript. Provider OAuth IDs
          are stored in our database alongside your email; the database is not publicly accessible.
        </p>
        <p>
          To report a security vulnerability:{" "}
          <a
            href="mailto:security@llmstatus.io"
            className="text-[var(--ink-300)] hover:text-[var(--ink-100)] transition-colors"
          >
            security@llmstatus.io
          </a>
        </p>
      </Section>

      <Section title="Changes to this policy">
        <p>
          We will update the effective date and post a notice on this page when changes are made.
          Material changes affecting existing users will be communicated by email.
        </p>
      </Section>

      <Section title="Contact">
        <p>
          Privacy questions:{" "}
          <a
            href="mailto:privacy@llmstatus.io"
            className="text-[var(--ink-300)] hover:text-[var(--ink-100)] transition-colors"
          >
            privacy@llmstatus.io
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
