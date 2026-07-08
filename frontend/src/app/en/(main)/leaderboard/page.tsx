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

const MOCK_USERS: User[] = [
  { ID: 1, Username: "Mega_bells", DisplayName: "Mega_bells", Tickets: 3569.76, XP: 5000 },
  { ID: 2, Username: "StandardCrystal", DisplayName: "StandardCrystal", Tickets: 2907.17, XP: 3200 },
  { ID: 3, Username: "Alpha2", DisplayName: "Alpha2", Tickets: 2693.07, XP: 2800 },
  { ID: 4, Username: "WhisperWind", DisplayName: "WhisperWind", Tickets: 1845.50, XP: 1500 },
  { ID: 5, Username: "NightOwl", DisplayName: "NightOwl", Tickets: 1623.80, XP: 1200 },
  { ID: 6, Username: "SilverFox", DisplayName: "SilverFox", Tickets: 1412.25, XP: 900 },
  { ID: 7, Username: "CrimsonTide", DisplayName: "CrimsonTide", Tickets: 1234.00, XP: 700 },
  { ID: 8, Username: "GoldenEagle", DisplayName: "GoldenEagle", Tickets: 1112.90, XP: 500 },
  { ID: 9, Username: "StormChaser", DisplayName: "StormChaser", Tickets: 987.65, XP: 300 },
  { ID: 10, Username: "MoonlitPath", DisplayName: "MoonlitPath", Tickets: 876.54, XP: 100 },
];

const medalColors = ["text-yellow-400", "text-gray-300", "text-amber-600"];

export default function LeaderboardPage() {
  const [users, setUsers] = useState<User[]>(MOCK_USERS);
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
