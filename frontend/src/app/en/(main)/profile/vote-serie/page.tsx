"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { novels, votes } from "@/lib/api";
import { Novel } from "@/types";
import { useAuth } from "@/lib/AuthContext";
import Card from "@/components/ui/Card";

export default function VoteSeriePage() {
  const [novelList, setNovelList] = useState<Novel[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [sortBy, setSortBy] = useState<"votes" | "views" | "title">("votes");
  const [voting, setVoting] = useState<Record<number, boolean>>({});
  const [xpToast, setXpToast] = useState<{ show: boolean; message: string }>({ show: false, message: "" });
  const { user } = useAuth();

  const showXpToast = (msg: string) => {
    setXpToast({ show: true, message: msg });
    setTimeout(() => setXpToast({ show: false, message: "" }), 3000);
  };

  useEffect(() => {
    novels.list({ page: 1, limit: 100, sort: "views", order: "desc" })
      .then((res) => {
        setNovelList((res.data || []).filter((n: Novel) => n.ID));
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const filtered = useMemo(() => {
    let list = novelList;
    if (search.trim()) {
      const q = search.toLowerCase();
      list = list.filter((n) => n.Title.toLowerCase().includes(q));
    }
    return [...list].sort((a, b) => {
      if (sortBy === "votes") return (b.Votes || 0) - (a.Votes || 0);
      if (sortBy === "views") return (b.Views || 0) - (a.Views || 0);
      return a.Title.localeCompare(b.Title);
    });
  }, [novelList, search, sortBy]);

  const handleVote = async (novelId: number) => {
    if (voting[novelId]) return;
    setVoting((prev) => ({ ...prev, [novelId]: true }));
    try {
      const res = await votes.create(novelId);
      setNovelList((prev) =>
        prev.map((n) => (n.ID === novelId ? { ...n, Votes: (n.Votes || 0) + 1 } : n))
      );
      if (res.xp_earned > 0) {
        showXpToast(`+${res.xp_earned} XP from voting!`);
      }
    } catch {
      showXpToast("Already voted this novel");
    } finally {
      setVoting((prev) => ({ ...prev, [novelId]: false }));
    }
  };

  return (
    <div className="max-w-3xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold text-white mb-2">Vote Novels</h1>
      <p className="text-sm text-gray-500 mb-6">
        Vote for the novels you want to see prioritized. The most voted novels get translated faster.
      </p>

      {/* Search & Sort */}
      <div className="flex items-center gap-3 mb-6">
        <div className="relative flex-1">
          <svg className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search novels..."
            className="w-full bg-card-hover border border-line-light rounded-lg pl-10 pr-4 py-2.5 text-sm text-gray-200 outline-none focus:border-accent transition-colors"
          />
        </div>
        <div className="flex gap-1">
          {(["votes", "views", "title"] as const).map((s) => (
            <button
              key={s}
              onClick={() => setSortBy(s)}
              className={`px-3 py-2 text-xs font-medium rounded-lg transition-colors ${
                sortBy === s
                  ? "bg-violet-600 text-white"
                  : "bg-card-hover text-gray-400 hover:text-white hover:bg-line-light"
              }`}
            >
              {s === "votes" ? "Votes" : s === "views" ? "Views" : "Title"}
            </button>
          ))}
        </div>
      </div>

      {loading ? (
        <div className="text-center text-sm text-gray-500 py-8">Loading novels...</div>
      ) : filtered.length === 0 ? (
        <div className="bg-card border border-line rounded-xl p-6 text-center text-sm text-gray-500">
          {search ? "No novels match your search." : "No novels available to vote."}
        </div>
      ) : (
        <div className="space-y-3">
          {filtered.map((novel) => (
            <Card key={novel.ID} className="flex items-center gap-4">
              <div className="w-12 h-16 rounded-lg bg-card-hover flex-shrink-0 flex items-center justify-center overflow-hidden">
                {novel.CoverURL ? (
                  <img src={novel.CoverURL} alt="" className="w-full h-full object-cover" />
                ) : (
                  <svg className="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
                  </svg>
                )}
              </div>
              <div className="min-w-0 flex-1">
                <Link
                  href={`/en/novel/${novel.ID}/${novel.Slug}`}
                  className="text-sm text-gray-200 hover:text-violet-400 transition-colors font-medium"
                >
                  {novel.Title}
                </Link>
                <p className="text-xs text-gray-500 mt-0.5">{novel.Votes || 0} votes · {novel.Views?.toLocaleString() || 0} views</p>
              </div>
              <button
                onClick={() => handleVote(novel.ID)}
                disabled={voting[novel.ID]}
                className="px-4 py-2 text-xs font-medium rounded-lg transition-colors bg-card-hover text-gray-400 hover:text-white hover:bg-line-light disabled:opacity-50"
              >
                {voting[novel.ID] ? "..." : "Vote"}
              </button>
            </Card>
          ))}
        </div>
      )}

      {xpToast.show && (
        <div className="fixed bottom-6 right-6 z-50 animate-slide-up">
          <div className="bg-emerald-600 text-white px-5 py-3 rounded-xl shadow-lg text-sm font-medium flex items-center gap-2">
            <span>✦</span>
            {xpToast.message}
          </div>
        </div>
      )}
    </div>
  );
}
