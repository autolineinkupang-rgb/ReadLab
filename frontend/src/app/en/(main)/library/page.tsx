"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { library as libraryApi, auth } from "@/lib/api";
import Card from "@/components/ui/Card";

type Tab = "updates" | "history" | "folders";

interface FollowItem {
  ID: number;
  NovelID: number;
  Novel: {
    ID: number;
    Title: string;
    Slug: string;
    Chapters: number;
    CoverURL: string;
  };
  CreatedAt: string;
}

interface HistoryItem {
  ID: number;
  NovelID: number;
  ChapterNum: number;
  Novel: {
    ID: number;
    Title: string;
    Slug: string;
  };
  Chapter: {
    Num: number;
    Slug: string;
  };
  CreatedAt: string;
}

export default function LibraryPage() {
  const [activeTab, setActiveTab] = useState<Tab>("updates");
  const [loggedIn, setLoggedIn] = useState(false);
  const [checking, setChecking] = useState(true);
  const [follows, setFollows] = useState<FollowItem[]>([]);
  const [history, setHistory] = useState<HistoryItem[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    auth.me().then(() => {
      setLoggedIn(true);
      fetchData();
    }).catch(() => {}).finally(() => setChecking(false));
  }, []);

  async function fetchData() {
    setLoading(true);
    try {
      const res = await libraryApi.get();
      setFollows(res.follows || []);
      setHistory(res.history || []);
    } catch {
      setFollows([]);
      setHistory([]);
    } finally {
      setLoading(false);
    }
  }

  function timeAgo(dateStr: string): string {
    const diff = Date.now() - new Date(dateStr).getTime();
    const mins = Math.floor(diff / 60000);
    if (mins < 1) return "just now";
    if (mins < 60) return `${mins}m ago`;
    const hrs = Math.floor(mins / 60);
    if (hrs < 24) return `${hrs}h ago`;
    const days = Math.floor(hrs / 24);
    if (days < 30) return `${days}d ago`;
    return `${Math.floor(days / 30)}mo ago`;
  }

  if (checking) return <div className="max-w-4xl mx-auto px-4 py-16 text-center text-sm text-gray-500">Checking...</div>;

  if (!loggedIn) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-16 text-center">
        <Card padding={false} className="p-10 max-w-md mx-auto">
          <svg className="w-16 h-16 text-gray-600 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
          </svg>
          <h2 className="text-lg font-semibold text-white mb-2">Login Required</h2>
          <p className="text-sm text-gray-500 mb-6">You need to login to use Library features.</p>
          <Link href="/en/login" className="inline-block px-6 py-2.5 bg-violet-600 hover:bg-violet-700 text-white text-sm font-medium rounded-lg transition-colors">Go to Login</Link>
        </Card>
      </div>
    );
  }

  return (
    <div className="max-w-5xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold text-white mb-6">Library</h1>
      <div className="flex gap-4 border-b border-line mb-6">
        {(["updates", "history", "folders"] as const).map((tab) => (
          <button key={tab} onClick={() => setActiveTab(tab)}
            className={`pb-3 text-sm font-medium capitalize transition-colors border-b-2 ${
              activeTab === tab ? "text-violet-400 border-violet-500" : "text-gray-500 border-transparent hover:text-gray-300"
            }`}>{tab === "updates" ? "🔥 Updates" : tab === "history" ? "History" : "Followed Folders"}</button>
        ))}
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-16">
          <div className="w-8 h-8 border-2 border-accent border-t-transparent rounded-full animate-spin" />
        </div>
      ) : activeTab === "updates" ? (
        follows.length === 0 ? (
          <div className="text-center py-16 text-sm text-gray-500">No followed novels yet.</div>
        ) : (
          follows.map((item) => (
            <Link key={item.ID} href={`/en/novel/${item.NovelID}/${item.Novel.Slug}`} className="flex items-center gap-4 p-4 mb-3 bg-card border border-line rounded-xl hover:border-violet-800/40 transition-colors group">
              <div className="w-14 h-20 rounded-lg bg-card-hover flex-shrink-0 flex items-center justify-center">
                {item.Novel.CoverURL ? (
                  <img src={item.Novel.CoverURL} alt="" className="w-full h-full object-cover rounded-lg" />
                ) : (
                  <svg className="w-6 h-6 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
                  </svg>
                )}
              </div>
              <div className="min-w-0 flex-1">
                <p className="text-sm text-gray-200 group-hover:text-violet-400 transition-colors font-medium">{item.Novel.Title}</p>
                <p className="text-xs text-gray-500 mt-0.5">{item.Novel.Chapters} chapters</p>
              </div>
            </Link>
          ))
        )
      ) : activeTab === "history" ? (
        history.length === 0 ? (
          <div className="text-center py-16 text-sm text-gray-500">No reading history yet.</div>
        ) : (
          history.map((item) => (
            <Link key={item.ID} href={`/en/novel/${item.NovelID}/${item.Novel.Slug}/chapter-${item.ChapterNum}`} className="flex items-center justify-between p-3 rounded-lg hover:bg-card transition-colors group">
              <div className="min-w-0 flex-1">
                <p className="text-sm text-gray-200 group-hover:text-violet-400 transition-colors">{item.Novel.Title}</p>
                <p className="text-xs text-violet-400 mt-0.5">Chapter {item.ChapterNum}</p>
              </div>
              <span className="text-xs text-gray-600 shrink-0 ml-4">{timeAgo(item.CreatedAt)}</span>
            </Link>
          ))
        )
      ) : (
        <div className="text-center py-8 text-sm text-gray-500">No folders yet. Follow novels to create folders.</div>
      )}
    </div>
  );
}
