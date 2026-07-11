"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import type { LatestNewsItem } from "@/types";

/**
 * Dynamic hero banner shown on the homepage.
 * - Prefers the latest news item (if provided) so admins can drive the message
 *   from the news CMS.
 * - Falls back to a friendly welcome banner when no news is available.
 * - Dismissible per news-item; state persists in localStorage.
 */
export default function HeroBanner({ latestNews }: { latestNews?: LatestNewsItem | null }) {
  const [dismissed, setDismissed] = useState(false);
  const key = latestNews?.ID ? `readlab.hero.dismissed.${latestNews.ID}` : "readlab.hero.dismissed.welcome";

  useEffect(() => {
    try {
      const isDismissed = localStorage.getItem(key) === "1";
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setDismissed(isDismissed);
    } catch {
      /* SSR / disabled storage — ignore */
    }
  }, [key]);

  if (dismissed) return null;

  const dismiss = () => {
    try { localStorage.setItem(key, "1"); } catch {}
    setDismissed(true);
  };

  const title = latestNews?.Title || "Welcome to ReadLab";
  const href = latestNews?.ID ? `/en/news/${latestNews.ID}` : "/en/news";
  const cta = latestNews ? "Read More" : "Explore News";
  const sub = latestNews ? "Latest announcement" : "Read thousands of light novels — free forever";

  return (
    <div
      className="relative overflow-hidden rounded-2xl border border-accent/30 bg-gradient-to-br from-accent/20 via-card to-accent-light/10 p-6 sm:p-8 animate-fade-in"
      data-testid="hero-banner"
    >
      <div
        className="pointer-events-none absolute -top-16 -right-16 w-64 h-64 rounded-full bg-accent/20 blur-3xl"
        aria-hidden="true"
      />
      <div
        className="pointer-events-none absolute -bottom-20 -left-10 w-56 h-56 rounded-full bg-accent-light/10 blur-3xl"
        aria-hidden="true"
      />
      <div className="relative flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <p className="text-[10px] uppercase tracking-[0.2em] text-accent-light font-semibold mb-1">
            {sub}
          </p>
          <h2 className="text-xl sm:text-2xl font-bold text-white line-clamp-2">
            {title}
          </h2>
          <Link
            href={href}
            className="inline-flex items-center gap-1.5 mt-4 px-4 py-2 bg-accent hover:bg-accent-dark text-white text-sm rounded-lg transition-colors shadow-md shadow-accent/30 focus:outline-none focus-visible:ring-2 focus-visible:ring-accent-light"
            data-testid="hero-banner-cta"
          >
            {cta}
            <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
            </svg>
          </Link>
        </div>
        <button
          onClick={dismiss}
          className="shrink-0 p-1.5 text-gray-500 hover:text-white rounded-lg hover:bg-white/5 transition-colors"
          aria-label="Dismiss"
          data-testid="hero-banner-dismiss"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    </div>
  );
}
