"use client";

import Link from "next/link";
import { useRouter, usePathname } from "next/navigation";
import { useState, useEffect } from "react";
import { navSections as sections } from "@/lib/navigation";
import { SearchIcon } from "@/components/ui/Icons";
import { useAuth } from "@/lib/AuthContext";

export default function Sidebar() {
  const [search, setSearch] = useState("");
  const [hydrated, setHydrated] = useState(false);
  const { user, loading, logout } = useAuth();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => { setHydrated(true); }, []);

  const isActive = (href: string) => hydrated ? (pathname?.startsWith(href) ?? false) : false;

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (search.trim()) {
      router.push(`/en/novel-finder?q=${encodeURIComponent(search.trim())}`);
      setSearch("");
    }
  };

  return (
    <aside className="hidden lg:flex fixed left-0 top-0 h-full w-64 flex-col bg-card border-r border-line z-50 shadow-2xl shadow-black/50">
      <div className="px-6 pt-6 pb-4">
        <Link
          href="/en"
          className="text-2xl font-bold bg-gradient-to-r from-accent to-accent-light bg-clip-text text-transparent"
        >
          ReadLab
        </Link>
      </div>

      <form onSubmit={handleSearch} className="px-4 pb-4">
        <div className="relative">
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search novels..."
            className="w-full bg-card-hover border border-line-light rounded-lg pl-4 pr-10 py-2 text-sm text-gray-200 outline-none focus:border-accent transition-colors"
          />
          <button
            type="submit"
            className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-500 hover:text-accent-light transition-colors"
          >
            <SearchIcon className="w-4 h-4" />
          </button>
        </div>
      </form>

      <nav className="flex-1 overflow-y-auto px-3 pb-4 space-y-6">
        {sections
          .filter((s) => {
            if (s.title === "Admin" && user?.role !== "admin") return false;
            if (s.title === "Writer" && user?.role !== "writer" && user?.role !== "admin") return false;
            return true;
          })
          .map((section) => (
          <div key={section.title}>
            <p className="text-[10px] uppercase tracking-wider text-gray-600 px-3 pb-1 font-semibold">
              {section.title}
            </p>
            {section.links.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                className={`flex items-center text-sm px-3 py-2 rounded-lg transition-colors ${
                  isActive(link.href)
                    ? "text-accent-light bg-accent/10"
                    : "text-gray-300 hover:text-white hover:bg-card-hover"
                }`}
              >
                {link.label}
              </Link>
            ))}
          </div>
        ))}
      </nav>

      <div className="px-4 pb-6 pt-4 border-t border-line">
        {loading ? (
          <div className="w-full h-10 rounded-lg bg-card-hover animate-pulse" />
        ) : user ? (
          <div className="space-y-2">
            <Link
              href={`/en/profile/${user.id}`}
              className="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-card-hover transition-colors"
            >
              <div className="w-9 h-9 rounded-full bg-accent flex items-center justify-center text-white text-sm font-bold shrink-0">
                {(user.display_name || user.username)[0].toUpperCase()}
              </div>
              <div className="min-w-0">
                <p className="text-sm font-medium text-gray-200 truncate">{user.display_name || user.username}</p>
                <p className="text-[10px] text-gray-500">{user.tickets.toLocaleString()} Tickets</p>
              </div>
            </Link>
            <button
              onClick={logout}
              className="w-full text-center text-sm px-4 py-2 rounded-lg border border-red-900/30 text-red-400 hover:bg-red-900/20 transition-colors"
            >
              Logout
            </button>
          </div>
        ) : (
          <Link
            href="/en/login"
            className="block text-center text-sm px-4 py-2.5 rounded-lg bg-accent hover:bg-accent-dark text-white transition-colors"
          >
            Login
          </Link>
        )}
      </div>
    </aside>
  );
}
