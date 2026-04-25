import { redirect } from "next/navigation";
import type { Metadata } from "next";
import { getSession } from "@/lib/session";
import SponsorReviewList from "./SponsorReviewList";

export const metadata: Metadata = { title: "Admin — Sponsor Review" };

const API = process.env.API_URL ?? "http://localhost:8081";

interface PendingSponsor {
  id: string;
  name: string;
  website_url: string | null;
  logo_url: string | null;
  tier: string;
  user_id: number;
}

async function fetchIsAdmin(token: string): Promise<boolean> {
  try {
    const res = await fetch(`${API}/auth/me`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    if (!res.ok) return false;
    const { data } = await res.json();
    return data?.is_admin === true;
  } catch {
    return false;
  }
}

async function fetchPendingSponsors(token: string): Promise<PendingSponsor[]> {
  try {
    const res = await fetch(`${API}/v1/admin/sponsors`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    if (!res.ok) return [];
    return await res.json();
  } catch {
    return [];
  }
}

export default async function AdminSponsorsPage() {
  const session = await getSession();
  if (!session) redirect("/login");

  const isAdmin = await fetchIsAdmin(session.token);
  if (!isAdmin) redirect("/");

  const sponsors = await fetchPendingSponsors(session.token);

  return (
    <main className="flex-1 mx-auto w-full max-w-2xl px-6 py-10">
      <h1 className="mb-2 text-xl font-semibold text-[var(--ink-100)]">Sponsor Review</h1>
      <p className="mb-6 text-sm text-[var(--ink-500)]">
        {sponsors.length} pending application{sponsors.length !== 1 ? "s" : ""}
      </p>
      <SponsorReviewList sponsors={sponsors} apiToken={session.token} />
    </main>
  );
}
