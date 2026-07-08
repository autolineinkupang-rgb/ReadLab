"use client";

import { useEffect, useState } from "react";
import { adminReviews } from "@/lib/api";
import Card from "@/components/ui/Card";
import RequireRole from "@/components/RequireRole";

interface ReviewItem {
  id: number;
  user_id: number;
  username: string;
  novel_id: number;
  novel_title: string;
  rating: number;
  content: string;
  created_at: string;
}

export default function AdminReviewsPage() {
  return (
    <RequireRole roles={["admin"]}>
      <AdminReviews />
    </RequireRole>
  );
}

function AdminReviews() {
  const [data, setData] = useState<ReviewItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [message, setMessage] = useState("");

  useEffect(() => {
    fetchReviews(1);
  }, []);

  async function fetchReviews(p: number) {
    setLoading(true);
    setPage(p);
    try {
      const res = await adminReviews.list({ page: p, limit: 20 });
      setData(res.data || []);
      setTotalPages(res.total_pages || 1);
    } catch {
      setData([]);
    } finally {
      setLoading(false);
    }
  }

  async function handleDelete(id: number) {
    if (!confirm("Delete this review?")) return;
    setMessage("");
    try {
      await adminReviews.delete(id);
      setMessage("Review deleted.");
      fetchReviews(page);
    } catch (e: any) {
      setMessage(`Delete failed: ${e.message}`);
    }
  }

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold text-white mb-6">Manage Reviews</h1>

      {message && <p className="text-sm text-accent-light mb-4">{message}</p>}

      {loading ? (
        <div className="flex items-center justify-center py-16">
          <div className="w-8 h-8 border-2 border-accent border-t-transparent rounded-full animate-spin" />
        </div>
      ) : (
        <>
          <div className="space-y-3">
            {data.map((r) => (
              <Card key={r.id}>
                <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-3 sm:gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className="text-sm font-semibold text-white">{r.novel_title}</span>
                      <span className="text-xs text-gray-500">by {r.username}</span>
                    </div>
                    <div className="flex items-center gap-2 mt-1">
                      <div className="flex items-center gap-0.5">
                        {[1, 2, 3, 4, 5].map((s) => (
                          <svg
                            key={s}
                            className={`w-3.5 h-3.5 ${s <= r.rating ? "text-yellow-400" : "text-gray-600"}`}
                            fill="currentColor"
                            viewBox="0 0 20 20"
                          >
                            <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                          </svg>
                        ))}
                      </div>
                      <span className="text-[10px] text-gray-500">{new Date(r.created_at).toLocaleDateString()}</span>
                    </div>
                    <p className="text-xs text-gray-400 mt-2 line-clamp-2">{r.content}</p>
                  </div>
                  <button
                    onClick={() => handleDelete(r.id)}
                    className="px-3 py-1.5 bg-red-900/50 hover:bg-red-800/50 text-red-400 text-xs rounded-lg transition-colors shrink-0"
                  >
                    Delete
                  </button>
                </div>
              </Card>
            ))}
          </div>

          {data.length === 0 && (
            <div className="text-center py-16 text-gray-500">
              <p>No reviews found.</p>
            </div>
          )}

          {totalPages > 1 && (
            <div className="flex items-center justify-center gap-2 mt-8">
              <button
                onClick={() => fetchReviews(Math.max(1, page - 1))}
                disabled={page <= 1}
                className="px-3 py-1.5 text-xs rounded-lg bg-card-hover text-gray-300 hover:bg-line-light disabled:opacity-40 transition-colors"
              >
                Previous
              </button>
              <span className="text-xs text-gray-500">Page {page} / {totalPages}</span>
              <button
                onClick={() => fetchReviews(Math.min(totalPages, page + 1))}
                disabled={page >= totalPages}
                className="px-3 py-1.5 text-xs rounded-lg bg-card-hover text-gray-300 hover:bg-line-light disabled:opacity-40 transition-colors"
              >
                Next
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
