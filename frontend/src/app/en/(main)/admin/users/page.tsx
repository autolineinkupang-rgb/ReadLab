"use client";

import { useEffect, useState } from "react";
import { adminUsers } from "@/lib/api";
import Card from "@/components/ui/Card";
import RequireRole from "@/components/RequireRole";

function xpForLevel(level: number) { return (level - 1) ** 2 * 100; }
function calcLevel(xp: number) { return Math.floor(Math.sqrt(xp / 100)) + 1; }

function maskEmail(email: string): string {
  if (!email || !email.includes("@")) return email;
  const [name, domain] = email.split("@");
  if (name.length <= 2) return `${name[0]}***@${domain}`;
  return `${name[0]}${name[1]}***@${domain}`;
}

interface UserItem {
  id: number;
  username: string;
  email: string;
  password_hash: string;
  display_name: string;
  avatar_url: string;
  role: string;
  tickets: number;
  xp: number;
  created_at: string;
}

export default function AdminUsersPage() {
  return (
    <RequireRole roles={["admin"]}>
      <AdminUsers />
    </RequireRole>
  );
}

function AdminUsers() {
  const [data, setData] = useState<UserItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [roleFilter, setRoleFilter] = useState("");
  const [search, setSearch] = useState("");
  const [error, setError] = useState("");
  const [revealedEmails, setRevealedEmails] = useState<Set<number>>(new Set());
  const [revealedPasswords, setRevealedPasswords] = useState<Set<number>>(new Set());
  const [sendingTickets, setSendingTickets] = useState<number | null>(null);
  const [ticketAmount, setTicketAmount] = useState("");

  const fetchData = (p: number = 1) => {
    setLoading(true);
    adminUsers.list({ page: p, limit: 20, role: roleFilter || undefined, q: search || undefined })
      .then((res) => {
        setData(res.data as unknown as UserItem[]);
        setTotalPages(res.total_pages);
      })
      .catch(() => { setError("Failed to load users"); })
      .finally(() => setLoading(false));
  };

  useEffect(() => { fetchData(1); }, [roleFilter]);

  useEffect(() => { fetchData(page); }, [page]);

  const handleDelete = async (id: number, username: string) => {
    if (!confirm(`Delete user "${username}"? This action cannot be undone.`)) return;
    try {
      await adminUsers.delete(id);
      fetchData(page);
    } catch {
      setError("Failed to delete user");
    }
  };

  const handleRevealEmail = (id: number) => {
    if (!confirm("Reveal this user's email?")) return;
    setRevealedEmails((prev) => new Set(prev).add(id));
  };

  const handleRevealPassword = (id: number) => {
    if (!confirm("Reveal this user's password hash?")) return;
    setRevealedPasswords((prev) => new Set(prev).add(id));
  };

  const handleSendTickets = async (id: number) => {
    const amt = parseFloat(ticketAmount);
    if (isNaN(amt) || amt <= 0) return;
    try {
      await adminUsers.sendTickets(id, amt);
      setSendingTickets(null);
      setTicketAmount("");
      fetchData(page);
    } catch (e: any) {
      setError(e?.message || "Failed to send tickets");
    }
  };

  return (
    <div className="max-w-5xl mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-white">Users</h1>
        <div className="flex gap-3">
          <input
            type="text"
            placeholder="Search users..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            onKeyDown={(e) => { if (e.key === "Enter") fetchData(1); }}
            className="bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none focus:border-accent w-48"
          />
          <select
            value={roleFilter}
            onChange={(e) => setRoleFilter(e.target.value)}
            className="bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
          >
            <option value="">All Roles</option>
            <option value="admin">Admin</option>
            <option value="writer">Writer</option>
            <option value="member">Member</option>
          </select>
        </div>
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-900/30 border border-red-700 rounded-lg text-sm text-red-300">
          {error}
          <button onClick={() => setError("")} className="ml-2 text-red-200 hover:text-white">&times;</button>
        </div>
      )}

      <Card className="divide-y divide-line overflow-x-auto" padding={false}>
        {loading ? (
          <div className="p-6 text-center text-sm text-gray-500">Loading...</div>
        ) : data.length === 0 ? (
          <div className="p-6 text-center text-sm text-gray-500">No users found.</div>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-line text-left">
                <th className="px-3 py-3 text-gray-400 font-medium">ID</th>
                <th className="px-3 py-3 text-gray-400 font-medium">Username</th>
                <th className="px-3 py-3 text-gray-400 font-medium">Email</th>
                <th className="px-3 py-3 text-gray-400 font-medium">Password</th>
                <th className="px-3 py-3 text-gray-400 font-medium">Display Name</th>
                <th className="px-3 py-3 text-gray-400 font-medium">Role</th>
                <th className="px-3 py-3 text-gray-400 font-medium text-right">Level</th>
                <th className="px-3 py-3 text-gray-400 font-medium text-right">XP</th>
                <th className="px-3 py-3 text-gray-400 font-medium text-right">Tickets</th>
                <th className="px-3 py-3 text-gray-400 font-medium text-right">Action</th>
              </tr>
            </thead>
            <tbody>
              {data.map((u) => {
                const level = calcLevel(u.xp || 0);
                const isRevealed = revealedEmails.has(u.id);
                return (
                  <tr key={u.id} className="border-b border-line/50 hover:bg-card-hover/50">
                    <td className="px-3 py-3 text-gray-500 font-mono text-xs">{u.id}</td>
                    <td className="px-3 py-3 text-gray-200">{u.username}</td>
                    <td className="px-3 py-3">
                      {u.email ? (
                        <span className="inline-flex items-center gap-2">
                          <span className="text-gray-400 text-xs font-mono">
                            {isRevealed ? u.email : maskEmail(u.email)}
                          </span>
                          {!isRevealed && (
                            <button
                              onClick={() => handleRevealEmail(u.id)}
                              className="text-[10px] px-1.5 py-0.5 rounded bg-yellow-600/20 hover:bg-yellow-600/40 text-yellow-400 transition-colors"
                            >
                              Reveal
                            </button>
                          )}
                        </span>
                      ) : (
                        <span className="text-gray-600">—</span>
                      )}
                    </td>
                    <td className="px-3 py-3">
                      {u.password_hash ? (
                        <span className="inline-flex items-center gap-2">
                          <span className="text-gray-400 text-xs font-mono max-w-[120px] truncate">
                            {revealedPasswords.has(u.id) ? u.password_hash : "••••••••••••••••"}
                          </span>
                          {!revealedPasswords.has(u.id) && (
                            <button
                              onClick={() => handleRevealPassword(u.id)}
                              className="text-[10px] px-1.5 py-0.5 rounded bg-yellow-600/20 hover:bg-yellow-600/40 text-yellow-400 transition-colors"
                            >
                              Show
                            </button>
                          )}
                        </span>
                      ) : (
                        <span className="text-gray-600">—</span>
                      )}
                    </td>
                    <td className="px-3 py-3 text-gray-400">{u.display_name}</td>
                    <td className="px-3 py-3">
                      <span className={`text-xs px-2 py-0.5 rounded ${
                        u.role === "admin" ? "bg-red-900/40 text-red-300" :
                        u.role === "writer" ? "bg-violet-900/40 text-violet-300" :
                        "bg-gray-700/40 text-gray-400"
                      }`}>
                        {u.role}
                      </span>
                    </td>
                    <td className="px-3 py-3 text-right">
                      <span className="text-accent font-semibold">Lv.{level}</span>
                    </td>
                    <td className="px-3 py-3 text-right text-gray-300">{u.xp?.toLocaleString() || 0}</td>
                    <td className="px-3 py-3 text-right text-gray-300">{u.tickets}</td>
                    <td className="px-3 py-3 text-right">
                      <div className="flex items-center justify-end gap-2">
                        {sendingTickets === u.id ? (
                          <div className="flex items-center gap-1">
                            <input
                              type="number"
                              value={ticketAmount}
                              onChange={(e) => setTicketAmount(e.target.value)}
                              onKeyDown={(e) => { if (e.key === "Enter") handleSendTickets(u.id); }}
                              className="w-20 bg-card-hover border border-line-light rounded px-2 py-1 text-xs text-gray-200 outline-none focus:border-accent"
                              placeholder="Amount"
                              autoFocus
                              step="any"
                            />
                            <button
                              onClick={() => handleSendTickets(u.id)}
                              className="px-2 py-1 bg-green-600/20 hover:bg-green-600/40 text-green-300 text-xs font-medium rounded transition-colors"
                            >
                              Send
                            </button>
                            <button
                              onClick={() => { setSendingTickets(null); setTicketAmount(""); }}
                              className="px-2 py-1 text-gray-500 hover:text-gray-300 text-xs"
                            >
                              Cancel
                            </button>
                          </div>
                        ) : (
                          <button
                            onClick={() => { setSendingTickets(u.id); setTicketAmount(""); }}
                            className="px-2 py-1 bg-blue-600/20 hover:bg-blue-600/40 text-blue-300 text-xs font-medium rounded transition-colors"
                          >
                            Tickets
                          </button>
                        )}
                        <button
                          onClick={() => handleDelete(u.id, u.username)}
                          className="px-2 py-1 bg-red-600/20 hover:bg-red-600/40 text-red-300 text-xs font-medium rounded transition-colors"
                        >
                          Delete
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </Card>

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2 mt-4">
          {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
            <button
              key={p}
              onClick={() => setPage(p)}
              className={`px-3 py-1 text-sm rounded-lg transition-colors ${
                p === page ? "bg-accent text-white" : "bg-card-hover text-gray-400 hover:text-white"
              }`}
            >
              {p}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
