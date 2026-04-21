import type { Metadata } from "next";
import Link from "next/link";

export const metadata: Metadata = { title: "Sponsors — llmstatus.io" };

const API = process.env.API_URL ?? "http://localhost:8081";

interface Sponsor {
  id: string;
  name: string;
  website_url?: string;
  logo_url?: string;
  tier: string;
}

async function fetchSponsors(): Promise<Sponsor[]> {
  try {
    const res = await fetch(`${API}/v1/sponsors`, { next: { revalidate: 300 } });
    if (!res.ok) return [];
    return await res.json();
  } catch {
    return [];
  }
}

export default async function SponsorsPage() {
  const sponsors = await fetchSponsors();

  return (
    <main className="flex-1 mx-auto w-full max-w-3xl px-6 py-10">
      <h1 className="mb-2 text-2xl font-semibold text-[var(--ink-100)]">Sponsors</h1>
      <p className="mb-8 text-sm text-[var(--ink-400)]">
        These organizations sponsor llmstatus.io, helping us run probes and keep the
        data freely accessible.
      </p>

      {sponsors.length > 0 ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-12">
          {sponsors.map((sp) => (
            <SponsorCard key={sp.id} sponsor={sp} />
          ))}
        </div>
      ) : (
        <div className="mb-12 rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] p-6 text-center">
          <p className="text-sm text-[var(--ink-500)]">No sponsors yet — be the first!</p>
        </div>
      )}

      <div className="rounded-lg border border-[var(--signal-amber)] bg-[var(--canvas-raised)] p-6">
        <h2 className="mb-2 text-base font-semibold text-[var(--ink-100)]">Become a Sponsor</h2>
        <p className="mb-4 text-sm text-[var(--ink-400)] leading-relaxed">
          Sponsoring gives your organization a listing here and lets you supply your own
          API keys so we can probe your services with higher quota limits. All probe
          methodology stays independent and public.
        </p>
        <Link
          href="/sponsor/dashboard"
          className="inline-block rounded bg-[var(--signal-amber)] px-4 py-2 text-sm font-semibold text-[var(--canvas)] hover:opacity-90 transition-opacity"
        >
          Get started →
        </Link>
      </div>
    </main>
  );
}

function SponsorCard({ sponsor }: { sponsor: Sponsor }) {
  const inner = (
    <div className="flex items-center gap-4 rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] p-4 hover:border-[var(--ink-400)] transition-colors">
      {sponsor.logo_url ? (
        // eslint-disable-next-line @next/next/no-img-element
        <img
          src={sponsor.logo_url}
          alt={`${sponsor.name} logo`}
          className="h-10 w-10 rounded object-contain bg-[var(--canvas)]"
        />
      ) : (
        <div className="h-10 w-10 rounded bg-[var(--ink-600)] flex items-center justify-center text-xs font-bold text-[var(--ink-300)]">
          {sponsor.name.slice(0, 2).toUpperCase()}
        </div>
      )}
      <div>
        <p className="text-sm font-medium text-[var(--ink-100)]">{sponsor.name}</p>
        {sponsor.website_url && (
          <p className="text-xs text-[var(--ink-500)] mt-0.5 truncate max-w-48">
            {new URL(sponsor.website_url).hostname}
          </p>
        )}
      </div>
    </div>
  );

  if (sponsor.website_url) {
    return (
      <a href={sponsor.website_url} target="_blank" rel="noopener noreferrer">
        {inner}
      </a>
    );
  }
  return inner;
}
