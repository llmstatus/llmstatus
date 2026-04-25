import type { MetadataRoute } from "next";
import { listProviders, listIncidents } from "@/lib/api";

const SITE_URL = process.env.NEXT_PUBLIC_SITE_URL ?? "https://llmstatus.io";

export const revalidate = 3600; // rebuild hourly

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const now = new Date();

  const staticRoutes: MetadataRoute.Sitemap = [
    { url: SITE_URL, lastModified: now, changeFrequency: "always", priority: 1 },
    { url: `${SITE_URL}/providers`, lastModified: now, changeFrequency: "always", priority: 0.95 },
    { url: `${SITE_URL}/incidents`, lastModified: now, changeFrequency: "always", priority: 0.9 },
    { url: `${SITE_URL}/china`, lastModified: now, changeFrequency: "always", priority: 0.85 },
    { url: `${SITE_URL}/api`, lastModified: now, changeFrequency: "monthly", priority: 0.6 },
    { url: `${SITE_URL}/badges`, lastModified: now, changeFrequency: "monthly", priority: 0.5 },
    { url: `${SITE_URL}/compare`, lastModified: now, changeFrequency: "monthly", priority: 0.5 },
    { url: `${SITE_URL}/about`, lastModified: now, changeFrequency: "monthly", priority: 0.6 },
    { url: `${SITE_URL}/methodology`, lastModified: now, changeFrequency: "monthly", priority: 0.6 },
    { url: `${SITE_URL}/privacy`, lastModified: now, changeFrequency: "yearly", priority: 0.3 },
    { url: `${SITE_URL}/tos`, lastModified: now, changeFrequency: "yearly", priority: 0.3 },
  ];

  const providerRoutes: MetadataRoute.Sitemap = await listProviders()
    .then((providers) =>
      providers.map((p) => ({
        url: `${SITE_URL}/providers/${p.id}`,
        lastModified: now,
        changeFrequency: "always" as const,
        priority: 0.8,
      }))
    )
    .catch(() => []);

  const incidentRoutes: MetadataRoute.Sitemap = await listIncidents("all", 200)
    .then((incidents) =>
      incidents.map((inc) => ({
        url: `${SITE_URL}/incidents/${inc.slug}`,
        lastModified: new Date(inc.resolved_at ?? inc.started_at),
        changeFrequency: (inc.status === "ongoing" ? "always" : "never") as
          | "always"
          | "never",
        priority: inc.status === "ongoing" ? 0.7 : 0.5,
      }))
    )
    .catch(() => []);

  return [...staticRoutes, ...providerRoutes, ...incidentRoutes];
}
