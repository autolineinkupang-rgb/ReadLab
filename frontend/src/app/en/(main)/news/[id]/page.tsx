"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { news } from "@/lib/api";
import { formatDate, stripHtml } from "@/lib/utils";

interface NewsDetail {
  ID: number;
  Title: string;
  Content: string;
  Type: string;
  CreatedAt: string;
}

export default function NewsDetailPage() {
  const params = useParams();
  const id = parseInt((params?.id as string) || "0");
  const [detail, setDetail] = useState<NewsDetail | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!id) return;
    news.get(id)
      .then((res) => setDetail(res))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [id]);

  if (loading && !detail) {
    return <div className="max-w-4xl mx-auto px-4 py-16 text-center text-sm text-gray-500">Loading...</div>;
  }

  if (!detail) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-16 text-center">
        <p className="text-sm text-gray-500">News not found.</p>
        <Link href="/en/news" className="text-violet-400 hover:text-violet-300 text-sm mt-2 inline-block">← Back to News</Link>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <nav className="text-sm text-gray-500 mb-6">
        <Link href="/en" className="hover:text-violet-400 transition-colors">Home</Link>
        <span className="mx-2">/</span>
        <Link href="/en/news" className="hover:text-violet-400 transition-colors">News</Link>
        <span className="mx-2">/</span>
        <span className="text-gray-400">{detail.Title.slice(0, 50)}</span>
      </nav>
      <article className="bg-card border border-line rounded-xl p-8">
        <div className="flex items-center gap-3 mb-4">
          <span className={`text-xs px-2 py-0.5 rounded ${
            detail.Type === "changelog"
              ? "bg-blue-900/40 text-blue-400 border border-blue-800/30"
              : "bg-violet-900/40 text-violet-300 border border-violet-800/30"
          }`}>{detail.Type}</span>
          <span className="text-xs text-gray-600">{formatDate(detail.CreatedAt)}</span>
        </div>
        <h1 className="text-2xl font-bold text-white mb-6">{detail.Title}</h1>
        <div className="text-sm text-gray-300 leading-relaxed whitespace-pre-line">{stripHtml(detail.Content)}</div>
        <div className="mt-8 pt-6 border-t border-line">
          <Link href="/en/news" className="text-sm text-violet-400 hover:text-violet-300 transition-colors">← Back to News</Link>
        </div>
      </article>
    </div>
  );
}
