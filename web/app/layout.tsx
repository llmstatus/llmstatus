import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";

const geistSans = Geist({ variable: "--font-geist-sans", subsets: ["latin"] });
const geistMono = Geist_Mono({ variable: "--font-geist-mono", subsets: ["latin"] });

export const metadata: Metadata = {
  title: {
    default: "LLM Status — Real-time AI API monitoring",
    template: "%s — LLM Status",
  },
  description:
    "Independent real-time uptime and latency monitoring for AI API providers. Live data, not scraped status pages.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}>
      <body className="min-h-full flex flex-col bg-[var(--canvas-base)] text-[var(--ink-100)]">
        {children}
      </body>
    </html>
  );
}
