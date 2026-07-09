"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { adminUsers } from "@/lib/api";
import Card from "@/components/ui/Card";
import RequireRole from "@/components/RequireRole";

interface Stats {
  total_users: number;
  total_novels: number;
  total_chapters: number;
  total_admins: number;
  max_admins: number;
}

export default function AdminPage() {
  return (
    <RequireRole roles={["admin"]}>
      <AdminDashboard />
    </RequireRole>
  );
}

function AdminDashboard() {
  const [stats, setStats] = useState<Stats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    adminUsers.stats()
      .then(setStats)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const links = [
    { href: "/en/admin/users", label: "Users", desc: "Manage user roles and accounts", icon: "M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0z" },
    { href: "/en/admin/novels", label: "Novels", desc: "Add, edit, or delete novels", icon: "M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" },
    { href: "/en/admin/requests", label: "Requests", desc: "Review novel requests from users", icon: "M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" },
    { href: "/en/admin/import", label: "Import", desc: "Import novels from external sources", icon: "M4 16v2a2 2 0 002 2h12a2 2 0 002-2v-2M7 10l5 5 5-5M12 15V3" },
    { href: "/en/admin/novels", label: "Chapters", desc: "Manage chapter content for novels", icon: "M4 6h16M4 12h16M4 18h7" },
    { href: "/en/admin/reviews", label: "Reviews", desc: "Moderate user reviews", icon: "M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" },
    { href: "/en/admin/ticket-config", label: "Tickets", desc: "Configure ticket costs and rewards", icon: "M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" },
  ];

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold text-white mb-6">Admin Dashboard</h1>

      {loading ? (
        <div className="flex items-center justify-center py-16">
          <div className="w-8 h-8 border-2 border-accent border-t-transparent rounded-full animate-spin" />
        </div>
      ) : stats ? (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-8">
          <StatCard label="Users" value={stats.total_users} color="text-blue-400" />
          <StatCard label="Novels" value={stats.total_novels} color="text-violet-400" />
          <StatCard label="Chapters" value={stats.total_chapters} color="text-green-400" />
          <StatCard label="Admins" value={`${stats.total_admins}/${stats.max_admins}`} color="text-red-400" />
        </div>
      ) : (
        <div className="text-sm text-gray-500 mb-8">Failed to load stats.</div>
      )}

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        {links.map((link) => (
          <Link key={link.href} href={link.href}>
            <Card className="p-5 hover:border-violet-800/40 transition-colors cursor-pointer h-full">
              <div className="flex items-start gap-4">
                <div className="w-10 h-10 rounded-lg bg-accent/10 flex items-center justify-center shrink-0">
                  <svg className="w-5 h-5 text-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={1.5}>
                    <path d={link.icon} />
                  </svg>
                </div>
                <div>
                  <h3 className="text-sm font-semibold text-white">{link.label}</h3>
                  <p className="text-xs text-gray-500 mt-1">{link.desc}</p>
                </div>
              </div>
            </Card>
          </Link>
        ))}
      </div>
    </div>
  );
}

function StatCard({ label, value, color }: { label: string; value: string | number; color: string }) {
  return (
    <Card className="p-4 text-center">
      <div className={`text-2xl font-bold ${color}`}>{value}</div>
      <p className="text-xs text-gray-500 mt-1">{label}</p>
    </Card>
  );
}
