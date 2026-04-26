import type { Metadata } from "next";
import Link from "next/link";

export const metadata: Metadata = { title: "Sponsors — llmstatus.io" };

const API = process.env.API_URL ?? "http://localhost:8081";

interface Sponsor {
  id: string;
  name: string;
  website_url?: string;
  logo_url?: string;
  tagline?: string;
  tier: string;
  is_system?: boolean;
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

function safeHostname(url: string): string | null {
  try {
    return new URL(url).hostname;
  } catch {
    return null;
  }
}

const TIER_ORDER = ["platinum", "gold", "silver"] as const;
type Tier = (typeof TIER_ORDER)[number];

const TIER_LABEL: Record<Tier, string> = {
  platinum: "Platinum",
  gold: "Gold",
  silver: "Silver",
};

const TIER_ACCENT: Record<Tier, string> = {
  platinum: "var(--ink-100)",
  gold: "var(--signal-amber)",
  silver: "var(--ink-300)",
};

const KNOWN_TIERS = new Set<string>(TIER_ORDER);

export default async function SponsorsPage() {
  const all = await fetchSponsors();
  const byTier = Object.fromEntries(
    TIER_ORDER.map((t) => [t, all.filter((s) => s.tier === t)])
  ) as Record<Tier, Sponsor[]>;
  const legacy = all.filter((s) => !KNOWN_TIERS.has(s.tier));

  return (
    <main className="flex-1 mx-auto w-full max-w-3xl px-6 py-10">
      <h1 className="mb-2 text-2xl font-semibold text-[var(--ink-100)]">Sponsors</h1>
      <p className="mb-10 text-sm text-[var(--ink-400)]">
        These organizations sponsor llmstatus.io, helping us run probes and keep the
        data freely accessible.
      </p>

      {TIER_ORDER.map((tier) => (
        <TierSection
          key={tier}
          tier={tier}
          sponsors={byTier[tier]}
        />
      ))}

      {legacy.length > 0 && (
        <section className="mb-10">
          <h2 className="mb-4 text-xs font-semibold uppercase tracking-widest text-[var(--ink-400)]">
            Partners
          </h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {legacy.map((sp) => (
              <SponsorCard key={sp.id} sponsor={sp} />
            ))}
          </div>
        </section>
      )}

      <div className="rounded-lg border border-[var(--signal-amber)] bg-[var(--canvas-raised)] p-6 mt-4">
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

function TierSection({ tier, sponsors }: { tier: Tier; sponsors: Sponsor[] }) {
  return (
    <section className="mb-10">
      <h2
        className="mb-4 text-xs font-semibold uppercase tracking-widest"
        style={{ color: TIER_ACCENT[tier] }}
      >
        {TIER_LABEL[tier]}
      </h2>

      {sponsors.length > 0 ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {sponsors.map((sp) => (
            <SponsorCard key={sp.id} sponsor={sp} />
          ))}
        </div>
      ) : (
        <PlaceholderSlot tier={tier} />
      )}
    </section>
  );
}

function PlaceholderSlot({ tier }: { tier: Tier }) {
  return (
    <div className="rounded-lg border border-dashed border-[var(--ink-600)] bg-[var(--canvas-raised)] p-5 text-center">
      <p className="text-sm text-[var(--ink-500)]">
        {TIER_LABEL[tier]} sponsor slot available —{" "}
        <Link href="/sponsor/dashboard" className="underline underline-offset-2 hover:text-[var(--ink-300)] transition-colors">
          get in touch
        </Link>
      </p>
    </div>
  );
}

function SponsorCard({ sponsor }: { sponsor: Sponsor }) {
  const hostname = sponsor.website_url ? safeHostname(sponsor.website_url) : null;

  const inner = (
    <div className="flex gap-4 rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] p-4 hover:border-[var(--ink-400)] transition-colors h-full">
      {sponsor.logo_url ? (
        // eslint-disable-next-line @next/next/no-img-element
        <img
          src={sponsor.logo_url}
          alt={`${sponsor.name} logo`}
          className="h-12 w-12 shrink-0 rounded object-contain bg-[var(--canvas)]"
        />
      ) : (
        <div className="h-12 w-12 shrink-0 rounded bg-[var(--ink-600)] flex items-center justify-center text-xs font-bold text-[var(--ink-300)]">
          {sponsor.name.slice(0, 2).toUpperCase()}
        </div>
      )}
      <div className="min-w-0">
        <p className="text-sm font-semibold text-[var(--ink-100)]">{sponsor.name}</p>
        {hostname && (
          <p className="text-xs text-[var(--ink-500)] mt-0.5 truncate">
            {hostname}
          </p>
        )}
        {sponsor.tagline && (
          <p className="text-xs text-[var(--ink-400)] mt-1.5 leading-relaxed line-clamp-3">
            {sponsor.tagline}
          </p>
        )}
      </div>
    </div>
  );

  if (sponsor.website_url) {
    return (
      <a href={sponsor.website_url} target="_blank" rel="noopener noreferrer" className="flex">
        {inner}
      </a>
    );
  }
  return inner;
}
