"use client";

import Card from "@/components/ui/Card";
import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { profile as profileApi, aiSettings as aiSettingsApi, auth, AITranslateSettings } from "@/lib/api";
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
  const [profile, setProfile] = useState<ProfileData | null>(null);

  const isOwner = user !== null && profile !== null && profile.id !== 0 && user.id === profile.id;

  const [aiSettings, setAiSettings] = useState<AITranslateSettings>({
    provider: "openrouter", model: "google/gemini-2.0-flash-exp:free",
    endpoint: "https://openrouter.ai/api/v1/chat/completions", key: "", has_key: false,
    target_language: "id-ID", instruction: "",
  });
  const [aiKeyInput, setAiKeyInput] = useState("");
  const [aiSaving, setAiSaving] = useState(false);
  const [aiMessage, setAiMessage] = useState("");
  const [aiShowForm, setAiShowForm] = useState(false);

  const [pwCurrent, setPwCurrent] = useState("");
  const [pwNew, setPwNew] = useState("");
  const [pwConfirm, setPwConfirm] = useState("");
  const [pwSaving, setPwSaving] = useState(false);
  const [pwMessage, setPwMessage] = useState("");
  const [pwShowForm, setPwShowForm] = useState(false);

  useEffect(() => {
    if (!profileId) return;
    profileApi.get(profileId)
      .then((res) => setProfile(res))
      .catch(() => setProfile({ id: 0, username: "", display_name: "", avatar_url: "", tickets: 0, xp: 0, created_at: "" }));
  }, [profileId]);

  useEffect(() => {
    if (!isOwner) return;
    aiSettingsApi.get()
      .then((res) => {
        setAiSettings(res);
        if (res.has_key) setAiKeyInput("");
      })
      .catch(() => {});
  }, [isOwner]);

  async function handleSaveAiSettings() {
    setAiSaving(true);
    setAiMessage("");
    try {
      const data: any = {
        provider: aiSettings.provider, model: aiSettings.model, endpoint: aiSettings.endpoint,
        target_language: aiSettings.target_language, instruction: aiSettings.instruction,
      };
      if (aiKeyInput) data.key = aiKeyInput;
      await aiSettingsApi.update(data);
      setAiMessage("Settings saved successfully");
      setAiSettings((prev) => ({ ...prev, has_key: !!aiKeyInput || prev.has_key }));
      setAiKeyInput("");
      setTimeout(() => setAiMessage(""), 3000);
    } catch (e: any) {
      setAiMessage(e.message || "Failed to save settings");
    } finally {
      setAiSaving(false);
    }
  }

  if (!profile) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-16 text-center">
        <div className="animate-pulse space-y-4">
          <div className="h-20 w-20 bg-card-hover rounded-full mx-auto" />
          <div className="h-6 bg-card-hover rounded w-48 mx-auto" />
          <div className="h-4 bg-card-hover rounded w-32 mx-auto" />
        </div>
      </div>
    );
  }

  if (profile.id === 0) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-16 text-center">
        <p className="text-gray-500">User not found.</p>
      </div>
    );
  }

  const joined = new Date(profile.created_at).toLocaleDateString("en-US", { year: "numeric", month: "long" });
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
        <div className="space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <Link href="/en/profile/vote-serie" className="p-4 bg-card border border-line rounded-xl hover:border-violet-800/40 transition-colors group">
              <p className="text-sm font-medium text-white group-hover:text-violet-400 transition-colors">Vote Novels</p>
              <p className="text-xs text-gray-500 mt-1">Vote for your favorite novels</p>
            </Link>
            <Link href="/en/profile/request-serie" className="p-4 bg-card border border-line rounded-xl hover:border-violet-800/40 transition-colors group">
              <p className="text-sm font-medium text-white group-hover:text-violet-400 transition-colors">Request Novels</p>
              <p className="text-xs text-gray-500 mt-1">Request new novels to be translated</p>
            </Link>
            {isOwner && (
              <button onClick={() => setAiShowForm(!aiShowForm)} className="p-4 bg-card border border-line rounded-xl hover:border-violet-800/40 transition-colors group text-left">
                <p className="text-sm font-medium text-white group-hover:text-violet-400 transition-colors">AI Translation</p>
                <p className="text-xs text-gray-500 mt-1">Configure your AI translation API key</p>
              </button>
            )}
            {isOwner && (
              <button onClick={() => { setPwShowForm(!pwShowForm); setPwMessage(""); }} className="p-4 bg-card border border-line rounded-xl hover:border-violet-800/40 transition-colors group text-left">
                <p className="text-sm font-medium text-white group-hover:text-violet-400 transition-colors">Change Password</p>
                <p className="text-xs text-gray-500 mt-1">Update your account password</p>
              </button>
            )}
          </div>

          {isOwner && aiShowForm && (
            <Card className="p-5 space-y-4">
              <h3 className="text-sm font-semibold text-white">AI Translation Settings</h3>
              <p className="text-xs text-gray-500">
                Get a free API key from{" "}
                <a href="https://openrouter.ai/keys" target="_blank" rel="noopener noreferrer" className="text-violet-400 hover:text-violet-300 underline">OpenRouter</a>
                {" "}(no credit card, throwaway email works). Free models like{" "}
                <code className="text-[11px] px-1 py-0.5 bg-gray-800 rounded">google/gemini-2.0-flash-exp:free</code> work well for translation.
              </p>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div>
                  <label className="text-xs text-gray-500 block mb-1">Provider</label>
                  <input value={aiSettings.provider} onChange={(e) => setAiSettings(p => ({ ...p, provider: e.target.value }))}
                    className="w-full px-3 py-2 text-sm rounded-lg border bg-transparent" style={{ borderColor: "#2a2a4a", color: "#e5e7eb" }} />
                </div>
                <div>
                  <label className="text-xs text-gray-500 block mb-1">Endpoint</label>
                  <input value={aiSettings.endpoint} onChange={(e) => setAiSettings(p => ({ ...p, endpoint: e.target.value }))}
                    className="w-full px-3 py-2 text-sm rounded-lg border bg-transparent" style={{ borderColor: "#2a2a4a", color: "#e5e7eb" }} />
                </div>
                <div>
                  <label className="text-xs text-gray-500 block mb-1">Model</label>
                  <input value={aiSettings.model} onChange={(e) => setAiSettings(p => ({ ...p, model: e.target.value }))}
                    className="w-full px-3 py-2 text-sm rounded-lg border bg-transparent" style={{ borderColor: "#2a2a4a", color: "#e5e7eb" }} />
                </div>
                <div>
                  <label className="text-xs text-gray-500 block mb-1">API Key</label>
                  <input type="password" value={aiKeyInput} onChange={(e) => setAiKeyInput(e.target.value)}
                    placeholder={aiSettings.has_key ? "•••••••• (leave blank to keep current)" : "Enter your API key"}
                    className="w-full px-3 py-2 text-sm rounded-lg border bg-transparent" style={{ borderColor: "#2a2a4a", color: "#e5e7eb" }} />
                </div>
                <div>
                  <label className="text-xs text-gray-500 block mb-1">Target Language</label>
                  <select value={aiSettings.target_language} onChange={(e) => setAiSettings(p => ({ ...p, target_language: e.target.value }))}
                    className="w-full px-3 py-2 text-sm rounded-lg border bg-transparent" style={{ borderColor: "#2a2a4a", color: "#e5e7eb" }}>
                    <option value="en-US">English</option>
                    <option value="id-ID">Indonesian</option>
                    <option value="ja-JP">Japanese</option>
                    <option value="ko-KR">Korean</option>
                    <option value="zh-CN">Chinese</option>
                    <option value="fr-FR">French</option>
                    <option value="de-DE">German</option>
                    <option value="es-ES">Spanish</option>
                    <option value="pt-PT">Portuguese</option>
                    <option value="ru-RU">Russian</option>
                    <option value="ar-SA">Arabic</option>
                    <option value="hi-IN">Hindi</option>
                    <option value="th-TH">Thai</option>
                    <option value="vi-VN">Vietnamese</option>
                  </select>
                </div>
              </div>
              <div>
                <label className="text-xs text-gray-500 block mb-1">AI Style Instruction <span className="text-gray-600">(optional)</span></label>
                <textarea value={aiSettings.instruction} onChange={(e) => setAiSettings(p => ({ ...p, instruction: e.target.value }))}
                  placeholder="e.g. Use casual Indonesian, add natural nuances, avoid literal translation."
                  rows={3}
                  className="w-full px-3 py-2 text-sm rounded-lg border bg-transparent resize-none"
                  style={{ borderColor: "#2a2a4a", color: "#e5e7eb" }} />
              </div>
              <div className="flex items-center gap-3">
                <button onClick={handleSaveAiSettings} disabled={aiSaving}
                  className="px-4 py-2 text-sm font-medium rounded-lg transition-colors disabled:opacity-50"
                  style={{ backgroundColor: "#7c3aed", color: "#ffffff" }}>
                  {aiSaving ? "Saving..." : "Save Settings"}
                </button>
                {aiMessage && <span className="text-xs" style={{ color: aiMessage.includes("Error") ? "#ef4444" : "#22c55e" }}>{aiMessage}</span>}
              </div>
            </Card>
          )}

          {isOwner && pwShowForm && (
            <Card className="p-5 space-y-4">
              <h3 className="text-sm font-semibold text-white">Change Password</h3>
              <div>
                <label className="text-xs text-gray-500 block mb-1">Current Password</label>
                <input type="password" value={pwCurrent} onChange={(e) => setPwCurrent(e.target.value)}
                  className="w-full px-3 py-2 text-sm rounded-lg border bg-transparent"
                  style={{ borderColor: "#2a2a4a", color: "#e5e7eb" }} />
              </div>
              <div>
                <label className="text-xs text-gray-500 block mb-1">New Password</label>
                <input type="password" value={pwNew} onChange={(e) => setPwNew(e.target.value)}
                  className="w-full px-3 py-2 text-sm rounded-lg border bg-transparent"
                  style={{ borderColor: "#2a2a4a", color: "#e5e7eb" }} />
              </div>
              <div>
                <label className="text-xs text-gray-500 block mb-1">Confirm New Password</label>
                <input type="password" value={pwConfirm} onChange={(e) => setPwConfirm(e.target.value)}
                  className="w-full px-3 py-2 text-sm rounded-lg border bg-transparent"
                  style={{ borderColor: "#2a2a4a", color: "#e5e7eb" }} />
              </div>
              <div className="flex items-center gap-3">
                <button onClick={async () => {
                  if (!pwCurrent || !pwNew || !pwConfirm) { setPwMessage("Please fill in all fields"); return; }
                  if (pwNew !== pwConfirm) { setPwMessage("New passwords do not match"); return; }
                  if (pwNew.length < 8) { setPwMessage("Password must be at least 8 characters"); return; }
                  setPwSaving(true);
                  setPwMessage("");
                  try {
                    await auth.changePassword(pwCurrent, pwNew);
                    setPwMessage("Password updated successfully");
                    setPwCurrent("");
                    setPwNew("");
                    setPwConfirm("");
                  } catch (err: any) {
                    setPwMessage(err.message || "Failed to update password");
                  } finally {
                    setPwSaving(false);
                  }
                }} disabled={pwSaving}
                  className="px-4 py-2 text-sm font-medium rounded-lg transition-colors disabled:opacity-50"
                  style={{ backgroundColor: "#7c3aed", color: "#ffffff" }}>
                  {pwSaving ? "Saving..." : "Update Password"}
                </button>
                {pwMessage && <span className="text-xs" style={{ color: pwMessage.includes("Error") || pwMessage.includes("incorrect") || pwMessage.includes("match") || pwMessage.includes("fill") ? "#ef4444" : "#22c55e" }}>{pwMessage}</span>}
              </div>
            </Card>
          )}
        </div>
      )}
      {activeTab === "library" && <div className="text-center py-8 text-sm text-gray-500"><Link href="/en/library" className="text-violet-400 hover:text-violet-300 underline">Go to Library →</Link></div>}
      {activeTab === "votes" && <div className="text-center py-8 text-sm text-gray-500"><Link href="/en/profile/vote-serie" className="text-violet-400 hover:text-violet-300 underline">View voted novels →</Link></div>}
      {activeTab === "requests" && <div className="text-center py-8 text-sm text-gray-500"><Link href="/en/profile/request-serie" className="text-violet-400 hover:text-violet-300 underline">View requests →</Link></div>}
    </div>
  );
}
