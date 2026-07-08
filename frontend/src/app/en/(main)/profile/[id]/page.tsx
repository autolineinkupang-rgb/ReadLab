"use client";

import Card from "@/components/ui/Card";
import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { profile as profileApi } from "@/lib/api";
import { useAuth } from "@/lib/AuthContext";
import { ProfileData } from "@/types";

function xpForLevel(level: number) { return (level - 1) ** 2 * 100; }
function calcLevel(xp: number) { return Math.floor(Math.sqrt(xp / 100)) + 1; }

const tabs = ["overview", "library", "votes", "requests"];

export default function ProfilePage() {
  const params = useParams();
  const profileId = params?.id as string;
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState("overview");
  const [profile, setProfile] = useState<ProfileData>({
    id: 0, username: "reader1", display_name: "Reader One", avatar_url: "", tickets: 150, xp: 0, created_at: "2025-01-01",
  });

  const isOwner = user !== null && profile.id !== 0 && user.id === profile.id;

  useEffect(() => {
    if (!profileId) return;
    profileApi.get(profileId)
      .then((res) => setProfile(res))
      .catch(() => {});
  }, [profileId]);

  const joined = (() => {
    try {
      return new Date(profile.created_at).toLocaleDateString("en-US", { year: "numeric", month: "long" });
    } catch { return "January 2025"; }
  })();

  const level = calcLevel(profile.xp);
  const currentLevelXp = xpForLevel(level);
  const nextLevelXp = xpForLevel(level + 1);
  const progressPct = nextLevelXp > currentLevelXp
    ? Math.min(100, ((profile.xp - currentLevelXp) / (nextLevelXp - currentLevelXp)) * 100)
    : 100;

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <Card className="p-6 mb-6">
        <div className="flex items-center gap-5">
          <div className="w-20 h-20 rounded-full bg-card-hover flex items-center justify-center text-2xl font-bold text-gray-500 shrink-0 border-2 border-violet-800/30 relative">
            {profile.username[0]?.toUpperCase() || "?"}
            <span className="absolute -bottom-1 right-0 text-[10px] px-1.5 py-0.5 rounded-full bg-violet-800/60 text-violet-300 border border-violet-700/50 font-bold">
              Lv.{level}
            </span>
          </div>
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-3">
              <h1 className="text-xl font-bold text-white">{profile.display_name || profile.username}</h1>
              {isOwner && (
                <span className="px-2 py-0.5 rounded-full text-[10px] font-medium bg-violet-800/30 text-violet-400 border border-violet-700/40">
                  You
                </span>
              )}
            </div>
            <p className="text-sm text-gray-500">@{profile.username}</p>
            <div className="flex flex-wrap items-center gap-4 mt-2 text-sm text-gray-400">
              <span>🎫 {profile.tickets.toFixed(2)} Tickets</span>
              <span className="text-violet-400">✦ {profile.xp} XP</span>
            </div>
            <div className="mt-2 max-w-xs">
              <div className="flex items-center justify-between text-xs text-gray-500 mb-0.5">
                <span>Level {level}</span>
                <span>{profile.xp - currentLevelXp} / {nextLevelXp - currentLevelXp} XP</span>
              </div>
              <div className="w-full h-1.5 bg-card-hover rounded-full overflow-hidden">
                <div className="h-full bg-gradient-to-r from-violet-600 to-purple-600 rounded-full transition-all" style={{ width: `${progressPct}%` }} />
              </div>
            </div>
            <p className="text-xs text-gray-600 mt-1">Joined {joined}</p>
          </div>
        </div>
      </Card>

      <div className="flex gap-4 border-b border-line mb-6">
        {tabs.map((tab) => (
          <button key={tab} onClick={() => setActiveTab(tab)}
            className={`pb-3 text-sm font-medium capitalize transition-colors border-b-2 ${
              activeTab === tab ? "text-violet-400 border-violet-500" : "text-gray-500 border-transparent hover:text-gray-300"
            }`}>{tab}</button>
        ))}
      </div>

      {activeTab === "overview" && (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <Link href="/en/profile/vote-serie" className="p-4 bg-card border border-line rounded-xl hover:border-violet-800/40 transition-colors group">
            <p className="text-sm font-medium text-white group-hover:text-violet-400 transition-colors">Vote Novels</p>
            <p className="text-xs text-gray-500 mt-1">Vote for your favorite novels</p>
          </Link>
          <Link href="/en/profile/request-serie" className="p-4 bg-card border border-line rounded-xl hover:border-violet-800/40 transition-colors group">
            <p className="text-sm font-medium text-white group-hover:text-violet-400 transition-colors">Request Novels</p>
            <p className="text-xs text-gray-500 mt-1">Request new novels to be translated</p>
          </Link>
        </div>
      )}
      {activeTab === "library" && <div className="text-center py-8 text-sm text-gray-500"><Link href="/en/library" className="text-violet-400 hover:text-violet-300 underline">Go to Library →</Link></div>}
      {activeTab === "votes" && <div className="text-center py-8 text-sm text-gray-500"><Link href="/en/profile/vote-serie" className="text-violet-400 hover:text-violet-300 underline">View voted novels →</Link></div>}
      {activeTab === "requests" && <div className="text-center py-8 text-sm text-gray-500"><Link href="/en/profile/request-serie" className="text-violet-400 hover:text-violet-300 underline">View requests →</Link></div>}
    </div>
  );
}
