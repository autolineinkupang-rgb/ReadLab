"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { leaderboard } from "@/lib/api";
import Card from "@/components/ui/Card";

function xpForLevel(level: number) { return (level - 1) ** 2 * 100; }
function calcLevel(xp: number) { return Math.floor(Math.sqrt(xp / 100)) + 1; }

interface User {
  ID: number;
  Username: string;
  DisplayName: string;
  Tickets: number;
  XP: number;
}

const medalColors = ["text-yellow-400", "text-gray-300", "text-amber-600"];

export default function LeaderboardPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [sortBy, setSortBy] = useState("xp");

  useEffect(() => {
    leaderboard.get(sortBy)
      .then((res) => {
        if (res.data?.length) setUsers(res.data);
      })
      .catch(() => {});
  }, [sortBy]);

  return (
    <div className="max-w-3xl mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-white">Leaderboard</h1>
        <div className="flex gap-1 rounded-lg overflow-hidden border border-line-light">
          <button
            onClick={() => setSortBy("xp")}
            className={`px-3 py-1.5 text-xs font-medium transition-colors ${sortBy === "xp" ? "bg-violet-600 text-white" : "text-gray-400 hover:text-white"}`}
          >
            XP
          </button>
          <button
            onClick={() => setSortBy("ticket_count")}
            className={`px-3 py-1.5 text-xs font-medium transition-colors ${sortBy === "ticket_count" ? "bg-violet-600 text-white" : "text-gray-400 hover:text-white"}`}
          >
            Tickets
          </button>
        </div>
      </div>
      <Card className="divide-y divide-line">
        {users.map((u, i) => {
          const level = calcLevel(u.XP || 0);
          return (
            <Link
              key={u.ID}
              href={`/en/profile/${u.ID}`}
              className="flex items-center gap-4 p-4 hover:bg-[#1a1a3a] transition-colors group"
            >
              <span className={`text-lg font-bold w-8 text-center shrink-0 ${
                i < 3 ? medalColors[i] : "text-gray-600"
              }`}>#{i + 1}</span>
              <div className="w-10 h-10 rounded-full bg-card-hover flex items-center justify-center text-sm text-gray-500 shrink-0">
                {u.Username[0].toUpperCase()}
              </div>
              <div className="min-w-0 flex-1">
                <p className="text-sm text-gray-200 group-hover:text-violet-400 transition-colors font-medium">{u.Username}</p>
              </div>
              {sortBy === "xp" ? (
                <div className="text-right">
                  <p className="text-sm text-violet-400 font-medium">✦ {u.XP.toLocaleString()}</p>
                  <p className="text-[10px] text-gray-600">Lv.{level}</p>
                </div>
              ) : (
                <span className="text-sm text-violet-400 font-medium">{u.Tickets.toLocaleString()} Tickets</span>
              )}
            </Link>
          );
        })}
      </Card>
    </div>
  );
}
