"use client";

import { useEffect, useState } from "react";
import { adminRequests } from "@/lib/api";
import Card from "@/components/ui/Card";
import RequireRole from "@/components/RequireRole";

interface RequestItem {
  id: number;
  user_id: number;
  username: string;
  novel_title: string;
  novel_url: string;
  source: string;
  status: string;
  votes: number;
  created_at: string;
}

const STATUS_TABS = ["all", "pending", "approved", "rejected", "completed"];

export default function AdminRequestsPage() {
  return (
    <RequireRole roles={["admin"]}>
      <AdminRequests />
    </RequireRole>
  );
}

function AdminRequests() {
  const [data, setData] = useState<RequestItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [statusFilter, setStatusFilter] = useState("all");
  const [actionMsg, setActionMsg] = useState("");

  useEffect(() => {
    fetchRequests(1);
  }, [statusFilter]);

  async function fetchRequests(p: number) {
    setLoading(true);
    setPage(p);
    try {
      const res = await adminRequests.list({
        page: p,
        limit: 20,
        status: statusFilter !== "all" ? statusFilter : undefined,
      });
      setData(res.data || []);
      setTotalPages(res.total_pages || 1);
    } catch {
      setData([]);
    } finally {
      setLoading(false);
    }
  }

  async function handleReview(id: number, status: string) {
    setActionMsg("");
    try {
      await adminRequests.review(id, status);
      setActionMsg(`Request #${id} ${status}.`);
      fetchRequests(page);
    } catch (e: any) {
      setActionMsg(`Failed: ${e.message}`);
    }
  }

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold text-white mb-6">Review Requests</h1>

      {actionMsg && <p className="text-sm text-accent-light mb-4">{actionMsg}</p>}

      {/* Status filter tabs */}
      <div className="flex flex-wrap gap-1 mb-6">
        {STATUS_TABS.map((s) => (
          <button
            key={s}
            onClick={() => { setStatusFilter(s); setPage(1); }}
            className={`px-3 py-1.5 text-xs rounded-lg capitalize transition-colors ${
              statusFilter === s
                ? "bg-accent text-white"
                : "bg-card-hover text-gray-400 hover:text-white"
            }`}
          >
            {s}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-16">
          <div className="w-8 h-8 border-2 border-accent border-t-transparent rounded-full animate-spin" />
        </div>
      ) : (
        <>
          <div className="space-y-3">
            {data.map((req) => (
              <Card key={req.id}>
                <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-3 sm:gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 flex-wrap">
                      <h3 className="text-sm font-semibold text-white">{req.novel_title}</h3>
                      <span className="text-xs text-gray-500">by {req.username}</span>
                    </div>
                    {req.novel_url && (
                      <a href={req.novel_url} target="_blank" rel="noopener noreferrer" className="text-xs text-accent hover:underline truncate block mt-0.5">
                        {req.novel_url}
                      </a>
                    )}
                    <div className="flex flex-wrap items-center gap-3 mt-2 text-xs text-gray-500">
                      <span>Source: {req.source || "manual"}</span>
                      <span>Votes: {req.votes}</span>
                      <span className={`px-1.5 py-0.5 rounded text-[10px] font-medium ${
                        req.status === "pending" ? "bg-yellow-900/30 text-yellow-400" :
                        req.status === "approved" ? "bg-green-900/30 text-green-400" :
                        req.status === "rejected" ? "bg-red-900/30 text-red-400" :
                        "bg-blue-900/30 text-blue-400"
                      }`}>
                        {req.status}
                      </span>
                    </div>
                  </div>
                  {req.status === "pending" && (
                    <div className="flex gap-2 shrink-0">
                      <button
                        onClick={() => handleReview(req.id, "approved")}
                        className="flex-1 sm:flex-none px-3 py-1.5 bg-green-700 hover:bg-green-600 text-white text-xs rounded-lg transition-colors"
                      >
                        Approve
                      </button>
                      <button
                        onClick={() => handleReview(req.id, "rejected")}
                        className="flex-1 sm:flex-none px-3 py-1.5 bg-red-700 hover:bg-red-600 text-white text-xs rounded-lg transition-colors"
                      >
                        Reject
                      </button>
                    </div>
                  )}
                  {req.status !== "pending" && (
                    <span className="text-xs text-gray-500 italic shrink-0">
                      {req.status === "approved" ? "Approved" : req.status === "rejected" ? "Rejected" : req.status}
                    </span>
                  )}
                </div>
              </Card>
            ))}
          </div>

          {data.length === 0 && (
            <div className="text-center py-16 text-gray-500">
              <p>No requests found.</p>
            </div>
          )}

          {totalPages > 1 && (
            <div className="flex items-center justify-center gap-2 mt-8">
              <button
                onClick={() => fetchRequests(Math.max(1, page - 1))}
                disabled={page <= 1}
                className="px-3 py-1.5 text-xs rounded-lg bg-card-hover text-gray-300 hover:bg-line-light disabled:opacity-40 transition-colors"
              >
                Previous
              </button>
              <span className="text-xs text-gray-500">Page {page} / {totalPages}</span>
              <button
                onClick={() => fetchRequests(Math.min(totalPages, page + 1))}
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
