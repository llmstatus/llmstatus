import type { Metadata } from "next";
import { listProviders } from "@/lib/api";
import { CopyButton } from "@/components/CopyButton";

export const revalidate = 3600;

export const metadata: Metadata = {
  title: "Status Badges",
  description:
    "Embed llmstatus.io status badges on your site, README, or docs. " +
    "Live SVG badges for every monitored AI provider.",
  openGraph: {
    title: "Status Badges — llmstatus.io",
    description:
      "Embed llmstatus.io status badges on your site, README, or docs.",
  },
};

const SITE_URL =
  process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000";

// Public badge URL — in production nginx proxies /badge/* to the Go API.
// In dev the Next.js proxy route at /api/badge/{id} is used for preview.
function publicBadgeUrl(providerId: string, style?: "simple" | "detailed") {
  const base = `${SITE_URL}/badge/${providerId}.svg`;
  return style === "detailed" ? `${base}?style=detailed` : base;
}

function previewBadgeUrl(providerId: string, style?: "simple" | "detailed") {
  const base = `/api/badge/${providerId}`;
  return style === "detailed" ? `${base}?style=detailed` : base;
}

type BadgeStyle = "simple" | "detailed";

function BadgeRow({ name, id }: { name: string; id: string }) {
  const styles: { key: BadgeStyle; label: string }[] = [
    { key: "simple", label: "Simple" },
    { key: "detailed", label: "Detailed" },
  ];

  return (
    <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] overflow-hidden">
      {/* Provider header + badge previews */}
      <div className="px-4 py-3 border-b border-[var(--ink-600)]">
        <span className="text-sm font-semibold text-[var(--ink-100)] block mb-3">{name}</span>
        <div className="flex items-center gap-4 flex-wrap">
          {styles.map(({ key, label }) => (
            <div key={key} className="flex flex-col gap-1.5 items-start">
              <span className="text-[10px] font-semibold uppercase tracking-[0.08em] text-[var(--ink-400)]">
                {label}
              </span>
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={previewBadgeUrl(id, key)}
                alt={`${name} ${label} status badge`}
                height={20}
                className="h-5"
              />
            </div>
          ))}
        </div>
      </div>

      {/* Embed snippets — one section per style */}
      {styles.map(({ key, label }) => {
        const url = publicBadgeUrl(id, key);
        const markdown = `[![${name} status](${url})](${SITE_URL}/providers/${id})`;
        const html = `<a href="${SITE_URL}/providers/${id}"><img src="${url}" alt="${name} status" /></a>`;
        return (
          <div key={key} className="border-t border-[var(--ink-600)]">
            <div className="px-4 py-2 bg-[var(--canvas-sunken)]">
              <span className="text-[10px] font-semibold uppercase tracking-[0.08em] text-[var(--ink-300)]">
                {label} style
              </span>
            </div>
            <div className="divide-y divide-[var(--ink-600)]">
              <CodeSnippet label="Markdown" code={markdown} />
              <CodeSnippet label="HTML" code={html} />
              <CodeSnippet label="URL" code={url} />
            </div>
          </div>
        );
      })}
    </div>
  );
}

function CodeSnippet({ label, code }: { label: string; code: string }) {
  return (
    <div className="px-4 py-3">
      <div className="flex items-center justify-between mb-1.5">
        <span className="text-[10px] font-semibold uppercase tracking-[0.08em] text-[var(--ink-400)]">
          {label}
        </span>
        <CopyButton text={code} />
      </div>
      <pre className="text-[11px] font-mono text-[var(--ink-300)] break-all whitespace-pre-wrap leading-relaxed">
        {code}
      </pre>
    </div>
  );
}

export default async function BadgesPage() {
  const providers = await listProviders().catch(() => null);

  return (
    <main className="flex-1 mx-auto w-full max-w-4xl px-6">
      <div className="py-10 mb-6">
        <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-[var(--signal-amber)] mb-4">
          Badges
        </p>
        <h1 className="text-2xl font-semibold text-[var(--ink-100)] mb-3">
          Embed a live status badge
        </h1>
        <p className="text-sm text-[var(--ink-400)] leading-relaxed max-w-xl">
          Badges update every 30 seconds. Paste the snippet below into your
          README, documentation, or website.
        </p>
      </div>

      {providers === null ? (
        <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-6 py-10 text-center">
          <p className="text-sm text-[var(--ink-400)]">
            Could not reach the API. Check that the backend is running.
          </p>
        </div>
      ) : (
        <div className="flex flex-col gap-4">
          {providers.map((p) => (
            <BadgeRow key={p.id} name={p.name} id={p.id} />
          ))}
        </div>
      )}

      {/* RSS feeds section */}
      <div className="mt-16 mb-16">
        <div className="mb-6">
          <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-[var(--signal-amber)] mb-4">
            RSS Feeds
          </p>
          <h2 className="text-xl font-semibold text-[var(--ink-100)] mb-2">
            Subscribe to incident feeds
          </h2>
          <p className="text-sm text-[var(--ink-400)] leading-relaxed max-w-xl">
            Subscribe to the global feed or to any provider&#39;s feed in your RSS reader.
            Feeds update within 60 seconds of a new incident.
          </p>
        </div>

        <div className="flex flex-col gap-3">
          {/* Global feed */}
          <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] overflow-hidden">
            <div className="px-4 py-3 border-b border-[var(--ink-600)]">
              <span className="text-sm font-semibold text-[var(--ink-100)]">All providers</span>
            </div>
            <div className="px-4 py-3">
              <div className="flex items-center justify-between mb-1.5">
                <span className="text-[10px] font-semibold uppercase tracking-[0.08em] text-[var(--ink-400)]">
                  Feed URL
                </span>
                <CopyButton text={`${SITE_URL}/api/feed`} />
              </div>
              <pre className="text-[11px] font-mono text-[var(--ink-300)] break-all whitespace-pre-wrap leading-relaxed">
                {SITE_URL}/api/feed
              </pre>
            </div>
          </div>

          {/* Per-provider feeds */}
          {providers !== null &&
            providers.map((p) => (
              <div
                key={p.id}
                className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] overflow-hidden"
              >
                <div className="px-4 py-3 border-b border-[var(--ink-600)]">
                  <span className="text-sm font-semibold text-[var(--ink-100)]">{p.name}</span>
                </div>
                <div className="px-4 py-3">
                  <div className="flex items-center justify-between mb-1.5">
                    <span className="text-[10px] font-semibold uppercase tracking-[0.08em] text-[var(--ink-400)]">
                      Feed URL
                    </span>
                    <CopyButton text={`${SITE_URL}/api/feed/${p.id}`} />
                  </div>
                  <pre className="text-[11px] font-mono text-[var(--ink-300)] break-all whitespace-pre-wrap leading-relaxed">
                    {SITE_URL}/api/feed/{p.id}
                  </pre>
                </div>
              </div>
            ))}
        </div>
      </div>
    </main>
  );
}
