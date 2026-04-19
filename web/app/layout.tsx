import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { SiteHeader } from "@/components/SiteHeader";
import { SiteFooter } from "@/components/SiteFooter";

const geistSans = Geist({ variable: "--font-geist-sans", subsets: ["latin"] });
const geistMono = Geist_Mono({ variable: "--font-geist-mono", subsets: ["latin"] });

const DESCRIPTION =
  "Independent real-time uptime and latency monitoring for AI API providers. " +
  "Measured from 7 global locations. Not scraped from official status pages.";

export const metadata: Metadata = {
  title: {
    default: "llmstatus.io — AI API Status Monitor",
    template: "%s | llmstatus.io",
  },
  description: DESCRIPTION,
  openGraph: {
    siteName: "llmstatus.io",
    type: "website",
    description: DESCRIPTION,
  },
  twitter: {
    card: "summary",
  },
  alternates: {
    types: {
      "application/rss+xml": "/api/feed",
    },
  },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}>
      <body className="min-h-full flex flex-col bg-[var(--canvas-base)] text-[var(--ink-100)]">
        <SiteHeader />
        <div className="flex-1 flex flex-col">
          {children}
        </div>
        <SiteFooter />
      </body>
    </html>
  );
}
