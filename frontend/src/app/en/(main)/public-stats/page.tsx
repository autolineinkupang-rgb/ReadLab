"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { stats } from "@/lib/api";
import Card from "@/components/ui/Card";

interface StatsData {
  total_novels: number;
  total_chapters: number;
  total_users: number;
  total_views: number;
  total_votes: number;
  total_requests: number;
}

const statCards = [
  { key: "total_novels" as const, label: "Total Novels", icon: "📚" },
  { key: "total_chapters" as const, label: "Total Chapters", icon: "📝" },
  { key: "total_users" as const, label: "Total Users", icon: "👥" },
  { key: "total_views" as const, label: "Total Views", icon: "👁️" },
  { key: "total_votes" as const, label: "Total Votes Cast", icon: "🗳️" },
  { key: "total_requests" as const, label: "Novels Requested", icon: "📋" },
];

export default function PublicStatsPage() {
  const [data, setData] = useState<StatsData | null>(null);

  useEffect(() => {
    stats.get()
      .then((res) => setData(res))
      .catch(() => {});
  }, []);

  return (
    <div className="max-w-5xl mx-auto px-4 py-8">
      <nav className="text-sm text-gray-500 mb-6">
        <Link href="/en" className="hover:text-violet-400 transition-colors">Home</Link>
        <span className="mx-2">/</span>
        <span className="text-gray-400">Public Stats</span>
      </nav>
      <h1 className="text-3xl font-bold text-white mb-2">Public Statistics</h1>
      <p className="text-sm text-gray-500 mb-8">Real-time platform statistics for ReadLab</p>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {!data ? (
          statCards.map((s) => (
            <Card key={s.key} className="p-5">
              <div className="flex items-center justify-between mb-3">
                <span className="text-2xl">{s.icon}</span>
              </div>
              <div className="h-8 bg-card-hover rounded w-24 animate-pulse" />
              <div className="h-4 bg-card-hover rounded w-20 mt-2 animate-pulse" />
            </Card>
          ))
        ) : (
          statCards.map((s) => (
            <Card key={s.key} className="p-5">
              <div className="flex items-center justify-between mb-3">
                <span className="text-2xl">{s.icon}</span>
              </div>
              <p className="text-2xl font-bold text-white">{data[s.key].toLocaleString()}</p>
              <p className="text-sm text-gray-500 mt-1">{s.label}</p>
            </Card>
          ))
        )}
      </div>
    </div>
  );
}
