"use client";

import { useEffect, useState } from "react";
import { adminUsers } from "@/lib/api";
import Card from "@/components/ui/Card";
import RequireRole from "@/components/RequireRole";

interface UserItem {
  id: number;
  username: string;
  email: string;
  display_name: string;
  avatar_url: string;
  role: string;
  tickets: number;
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
  const [message, setMessage] = useState("");
  const [search, setSearch] = useState("");
  const [roleFilter, setRoleFilter] = useState("");
  const [stats, setStats] = useState<{ total_users: number; total_admins: number; max_admins: number } | null>(null);
  const [showCreateAdmin, setShowCreateAdmin] = useState(false);
  const [createForm, setCreateForm] = useState({ username: "", email: "", password: "" });

  useEffect(() => {
    adminUsers.stats().then(setStats).catch(() => {});
  }, []);

  useEffect(() => {
    fetchUsers(1);
  }, [roleFilter]);

  async function fetchUsers(p: number) {
    setLoading(true);
    setPage(p);
    try {
      const res = await adminUsers.list({ page: p, limit: 20, role: roleFilter || undefined, q: search || undefined });
      setData(res.data || []);
      setTotalPages(res.total_pages || 1);
    } catch {
      setData([]);
    } finally {
      setLoading(false);
    }
  }

  async function handleRoleChange(id: number, role: string) {
    setMessage("");
    try {
      await adminUsers.update(id, { role });
      setMessage("Role updated.");
      fetchUsers(page);
    } catch (e: any) {
      setMessage(e.message);
    }
  }

  async function handleDelete(id: number) {
    if (!confirm("Delete this user?")) return;
    setMessage("");
    try {
      await adminUsers.delete(id);
      setMessage("User deleted.");
      fetchUsers(page);
    } catch (e: any) {
      setMessage(e.message);
    }
  }

  async function handleCreateAdmin() {
    if (!createForm.username.trim() || !createForm.email.trim() || !createForm.password.trim()) return;
    setMessage("");
    try {
      await adminUsers.createAdmin(createForm);
      setMessage("Admin created.");
      setShowCreateAdmin(false);
      setCreateForm({ username: "", email: "", password: "" });
      adminUsers.stats().then(setStats).catch(() => {});
      fetchUsers(page);
    } catch (e: any) {
      setMessage(e.message);
    }
  }

  async function handleSearch() {
    fetchUsers(1);
  }

  const adminCount = data.filter((u) => u.role === "admin").length;

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-white">Manage Users</h1>
          {stats && (
            <p className="text-xs text-gray-500 mt-1">
              {stats.total_users} users &middot; {stats.total_admins}/{stats.max_admins} admins
            </p>
          )}
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => setShowCreateAdmin(!showCreateAdmin)}
            disabled={stats ? stats.total_admins >= stats.max_admins : false}
            className="px-4 py-2 bg-accent hover:bg-accent-dark disabled:opacity-50 text-white text-sm rounded-lg transition-colors"
          >
            {showCreateAdmin ? "Cancel" : "+ New Admin"}
          </button>
        </div>
      </div>

      {message && <p className="text-sm text-accent-light mb-4">{message}</p>}

      {showCreateAdmin && (
        <Card className="mb-6 p-4 space-y-3">
          <h2 className="text-white text-sm font-semibold">Create New Admin</h2>
          <input
            value={createForm.username}
            onChange={(e) => setCreateForm((p) => ({ ...p, username: e.target.value }))}
            className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
            placeholder="Username *"
          />
          <input
            value={createForm.email}
            onChange={(e) => setCreateForm((p) => ({ ...p, email: e.target.value }))}
            className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
            placeholder="Email *"
          />
          <input
            type="password"
            value={createForm.password}
            onChange={(e) => setCreateForm((p) => ({ ...p, password: e.target.value }))}
            className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
            placeholder="Password *"
          />
          <button
            onClick={handleCreateAdmin}
            disabled={!createForm.username.trim() || !createForm.email.trim() || !createForm.password.trim()}
            className="px-4 py-2 bg-green-600 hover:bg-green-700 disabled:opacity-50 text-white text-sm rounded-lg transition-colors"
          >
            Create Admin
          </button>
        </Card>
      )}

      {/* Search & Filter */}
      <div className="flex gap-3 mb-4">
        <div className="flex-1">
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleSearch()}
            className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
            placeholder="Search users..."
          />
        </div>
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
        <button onClick={handleSearch} className="px-4 py-2 bg-card-hover hover:bg-line-light text-gray-300 text-sm rounded-lg transition-colors border border-line-light">
          Search
        </button>
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-16">
          <div className="w-8 h-8 border-2 border-accent border-t-transparent rounded-full animate-spin" />
        </div>
      ) : (
        <>
          <div className="space-y-2">
            {data.map((u) => (
              <Card key={u.id} className="!p-4">
                <div className="flex items-center gap-4">
                  <div className="w-10 h-10 rounded-full bg-accent flex items-center justify-center text-white font-bold shrink-0">
                    {(u.display_name || u.username)[0].toUpperCase()}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium text-white">{u.display_name || u.username}</span>
                      <RoleBadge role={u.role} />
                    </div>
                    <p className="text-xs text-gray-500">{u.email}</p>
                  </div>
                  <div className="flex items-center gap-2 shrink-0">
                    <select
                      value={u.role}
                      onChange={(e) => handleRoleChange(u.id, e.target.value)}
                      className="text-xs bg-card-hover border border-line-light rounded px-2 py-1.5 text-gray-200 outline-none"
                    >
                      <option value="member">Member</option>
                      <option value="writer">Writer</option>
                      <option value="admin">Admin</option>
                    </select>
                    <button
                      onClick={() => handleDelete(u.id)}
                      className="px-2 py-1.5 text-xs rounded bg-red-900/50 hover:bg-red-800/50 text-red-400 transition-colors"
                    >
                      Delete
                    </button>
                  </div>
                </div>
              </Card>
            ))}
          </div>

          {totalPages > 1 && (
            <div className="flex items-center justify-center gap-2 mt-8">
              <button
                onClick={() => fetchUsers(Math.max(1, page - 1))}
                disabled={page <= 1}
                className="px-3 py-1.5 text-xs rounded-lg bg-card-hover text-gray-300 hover:bg-line-light disabled:opacity-40 transition-colors"
              >
                Previous
              </button>
              <span className="text-xs text-gray-500">Page {page} / {totalPages}</span>
              <button
                onClick={() => fetchUsers(Math.min(totalPages, page + 1))}
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

function RoleBadge({ role }: { role: string }) {
  const styles: Record<string, string> = {
    admin: "bg-red-900/40 text-red-400 border-red-800/30",
    writer: "bg-violet-900/40 text-violet-400 border-violet-800/30",
    member: "bg-blue-900/40 text-blue-400 border-blue-800/30",
  };
  return (
    <span className={`text-[10px] px-1.5 py-0.5 rounded border ${styles[role] || styles.member}`}>
      {role}
    </span>
  );
}
