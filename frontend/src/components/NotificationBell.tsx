"use client";

import { useState, useEffect, useRef } from "react";
import Link from "next/link";
import { notifications } from "@/lib/api";
import { useAuth } from "@/lib/AuthContext";

export default function NotificationBell() {
  const { user } = useAuth();
  const [unread, setUnread] = useState(0);
  const [list, setList] = useState<any[]>([]);
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!user) return;
    const fetchUnread = () => {
      notifications.unreadCount().then((res) => setUnread(res.unread_count)).catch(() => {});
    };
    fetchUnread();
    const interval = setInterval(fetchUnread, 30000);
    return () => clearInterval(interval);
  }, [user]);

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  const toggle = async () => {
    if (!open) {
      try {
        const res = await notifications.list();
        setList(res.data);
        setUnread(res.unread_count);
      } catch { /* ignore */ }
    }
    setOpen(!open);
  };

  const markRead = async (id: number) => {
    try {
      await notifications.markRead(id);
      setList((prev) => prev.map((n) => (n.ID === id ? { ...n, Read: true } : n)));
      setUnread((prev) => Math.max(0, prev - 1));
    } catch { /* ignore */ }
  };

  if (!user) return null;

  return (
    <div ref={ref} className="relative">
      <button onClick={toggle} className="relative p-2 text-gray-400 hover:text-white transition-colors" aria-label="Notifications">
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
        </svg>
        {unread > 0 && (
          <span className="absolute top-0.5 right-0.5 w-4 h-4 bg-red-500 text-white text-[10px] font-bold rounded-full flex items-center justify-center">
            {unread > 9 ? "9+" : unread}
          </span>
        )}
      </button>

      {open && (
        <div className="absolute top-full right-0 mt-1 w-80 bg-card border border-line-light rounded-lg shadow-xl z-50 max-h-96 overflow-y-auto">
          <div className="px-3 py-2 border-b border-line-light flex items-center justify-between">
            <span className="text-xs font-semibold text-gray-300">Notifications</span>
            {unread > 0 && (
              <button
                onClick={async () => { await notifications.markRead("all"); setList((prev) => prev.map((n) => ({ ...n, Read: true }))); setUnread(0); }}
                className="text-[10px] text-violet-400 hover:text-violet-300 transition-colors"
              >
                Mark all read
              </button>
            )}
          </div>
          {list.length === 0 ? (
            <p className="text-xs text-gray-500 text-center py-6">No notifications</p>
          ) : (
            list.map((n: any) => (
              <div key={n.ID} className={`px-3 py-2.5 border-b border-line-light last:border-0 ${n.Read ? "" : "bg-accent/5"}`}>
                <div className="flex items-start justify-between gap-2">
                  <div className="min-w-0 flex-1">
                    <Link
                      href={n.Link || "#"}
                      onClick={() => { setOpen(false); if (!n.Read) markRead(n.ID); }}
                      className="text-xs text-gray-200 hover:text-white transition-colors line-clamp-2"
                    >
                      {n.Message}
                    </Link>
                  </div>
                  {!n.Read && (
                    <button
                      onClick={() => markRead(n.ID)}
                      className="shrink-0 w-1.5 h-1.5 mt-1.5 rounded-full bg-violet-500 hover:bg-violet-400"
                      aria-label="Mark read"
                    />
                  )}
                </div>
                <p className="text-[10px] text-gray-600 mt-1">{new Date(n.CreatedAt).toLocaleDateString()}</p>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
}
