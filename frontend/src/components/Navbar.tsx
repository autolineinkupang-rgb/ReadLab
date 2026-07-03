"use client";

import Link from "next/link";
import { useState } from "react";

const navLinks = [
  { href: "/en/library", label: "Library" },
  { href: "/en/novel-list", label: "Novels" },
  { href: "/en/ranking/daily", label: "Ranking" },
  { href: "/en/leaderboard", label: "Leaderboard" },
];

export default function Navbar() {
  const [open, setOpen] = useState(false);

  return (
    <header className="sticky top-0 z-50 nav-gradient pb-4">
      <div className="max-w-7xl mx-auto px-4 flex items-center justify-between h-16">
        <Link href="/en" className="text-2xl font-bold bg-gradient-to-r from-violet-400 to-purple-600 bg-clip-text text-transparent">
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
          <Link
            href="/en/login"
            className="text-sm px-4 py-2 rounded-lg bg-violet-600 hover:bg-violet-700 text-white transition-colors"
          >
            Login
          </Link>
        </nav>

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
          <Link
            href="/en/login"
            className="block text-sm px-4 py-2 rounded-lg bg-violet-600 hover:bg-violet-700 text-white text-center"
            onClick={() => setOpen(false)}
          >
            Login
          </Link>
        </div>
      )}
    </header>
  );
}
