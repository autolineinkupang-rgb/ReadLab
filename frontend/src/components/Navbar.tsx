"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

const navLinks = [
  { href: "/en/library", label: "Library" },
  { href: "/en/novel-list", label: "Novels" },
  { href: "/en/ranking/daily", label: "Ranking" },
  { href: "/en/leaderboard", label: "Leaderboard" },
];

const adminLinks = [
  { href: "/en/admin/import", label: "Import" },
  { href: "/en/admin/requests", label: "Requests" },
  { href: "/en/admin/novels", label: "Manage" },
];

export default function Navbar() {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");
  const router = useRouter();

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (search.trim()) {
      router.push(`/en/novel-finder?q=${encodeURIComponent(search.trim())}`);
      setSearch("");
    }
  };

  return (
    <header className="sticky top-0 z-50 nav-gradient pb-4">
      <div className="max-w-7xl mx-auto px-4 flex items-center justify-between h-16 gap-4">
        <Link href="/en" className="text-2xl font-bold bg-gradient-to-r from-[#2193b0] to-[#6dd5ed] bg-clip-text text-transparent shrink-0">
          WTR-LAB
        </Link>

        <nav className="hidden md:flex items-center gap-6">
          {navLinks.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              className="text-sm text-gray-300 hover:text-white transition-colors"
            >
              {link.label}
            </Link>
          ))}
          <div className="relative group">
            <button className="text-sm text-gray-500 hover:text-[#6dd5ed] transition-colors flex items-center gap-1">
              Admin
              <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
              </svg>
            </button>
            <div className="absolute top-full right-0 mt-1 w-40 bg-[#12122a] border border-[#1e1e3a] rounded-xl p-2 shadow-xl opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 z-50">
              {adminLinks.map((link) => (
                <Link
                  key={link.href}
                  href={link.href}
                  className="block text-sm text-gray-300 hover:text-white px-3 py-2 rounded-lg hover:bg-[#1e1e3a] transition-colors"
                >
                  {link.label}
                </Link>
              ))}
            </div>
          </div>
        </nav>

        <form onSubmit={handleSearch} className="hidden md:flex flex-1 max-w-md">
          <div className="relative w-full">
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search novels..."
              className="w-full bg-[#1e1e3a] border border-[#2a2a4a] rounded-lg pl-4 pr-10 py-2 text-sm text-gray-200 outline-none focus:border-[#2193b0] transition-colors"
            />
            <button type="submit" className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-500 hover:text-[#6dd5ed] transition-colors">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </button>
          </div>
        </form>

        <Link
          href="/en/login"
          className="hidden md:inline-block text-sm px-4 py-2 rounded-lg bg-[#2193b0] hover:bg-[#1a7a94] text-white transition-colors shrink-0"
        >
          Login
        </Link>

        <button className="md:hidden text-white p-2" onClick={() => setOpen(!open)}>
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            {open ? (
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            ) : (
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
            )}
          </svg>
        </button>
      </div>

      {open && (
        <div className="md:hidden bg-[#0a0a1a]/95 backdrop-blur border-t border-[#1e1e3a] px-4 py-4 space-y-3">
          <form onSubmit={handleSearch}>
            <div className="relative">
              <input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Search novels..."
                className="w-full bg-[#1e1e3a] border border-[#2a2a4a] rounded-lg pl-4 pr-10 py-2 text-sm text-gray-200 outline-none focus:border-[#2193b0] transition-colors"
              />
              <button type="submit" className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-500 hover:text-[#6dd5ed] transition-colors">
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
              </button>
            </div>
          </form>
          {navLinks.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              className="block text-sm text-gray-300 hover:text-white py-1"
              onClick={() => setOpen(false)}
            >
              {link.label}
            </Link>
          ))}
          <div className="pt-2 border-t border-[#1e1e3a]">
            <p className="text-xs text-gray-500 px-1 pb-1">Admin</p>
            {adminLinks.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                className="block text-sm text-gray-400 hover:text-white py-1"
                onClick={() => setOpen(false)}
              >
                {link.label}
              </Link>
            ))}
          </div>
          <Link
            href="/en/login"
            className="block text-sm px-4 py-2 rounded-lg bg-[#2193b0] hover:bg-[#1a7a94] text-white text-center"
            onClick={() => setOpen(false)}
          >
            Login
          </Link>
        </div>
      )}
    </header>
  );
}
