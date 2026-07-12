"use client";

import { useEffect, useState } from "react";
import { adminTicketConfig } from "@/lib/api";
import { useAuth } from "@/lib/AuthContext";
import Card from "@/components/ui/Card";
import RequireRole from "@/components/RequireRole";

interface TicketConfigItem {
  ID: number;
  Key: string;
  Value: number;
  Label: string;
}

export default function TicketConfigPage() {
  return (
    <RequireRole roles={["admin"]}>
      <TicketConfigDashboard />
    </RequireRole>
  );
}

function TicketConfigDashboard() {
  const { refresh: refreshUser } = useAuth();
  const [configs, setConfigs] = useState<TicketConfigItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState<Record<string, boolean>>({});
  const [editingKey, setEditingKey] = useState<string | null>(null);
  const [editValue, setEditValue] = useState("");
  const [error, setError] = useState("");

  const load = () => {
    setLoading(true);
    adminTicketConfig.list()
      .then((res) => setConfigs(res.data as unknown as TicketConfigItem[]))
      .catch(() => { setConfigs([]); setError("Failed to load config"); })
      .finally(() => setLoading(false));
  };

  useEffect(() => { load(); }, []);

  const handleSave = async (key: string) => {
    const val = parseFloat(editValue);
    if (isNaN(val) || val < 0) {
      setError("Value must be a positive number");
      return;
    }
    setSaving((p) => ({ ...p, [key]: true }));
    setError("");
    try {
      await adminTicketConfig.update(key, val);
      setEditingKey(null);
      load();
      refreshUser();
    } catch {
      setError("Failed to update config");
    } finally {
      setSaving((p) => ({ ...p, [key]: false }));
    }
  };

  const labelMap: Record<string, string> = {
    daily_reward: "Daily Reward",
    novel_contribution: "Novel Contribution",
    monthly_leaderboard: "Monthly Leaderboard",
    edit_reset_cost: "Edit Reset",
    gate_bypass_cost: "Gate Bypass",
    replace_review_cost: "Replace Review",
  };

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-white">Ticket Configuration</h1>
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-900/30 border border-red-700 rounded-lg text-sm text-red-300">
          {error}
          <button onClick={() => setError("")} className="ml-2 text-red-200 hover:text-white">&times;</button>
        </div>
      )}

      <Card className="divide-y divide-line overflow-hidden" padding={false}>
        {loading ? (
          <div className="p-6 text-center text-sm text-gray-500">Loading...</div>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-line text-left">
                <th className="px-4 py-3 text-gray-400 font-medium">Setting</th>
                <th className="px-4 py-3 text-gray-400 font-medium">Key</th>
                <th className="px-4 py-3 text-gray-400 font-medium text-right">Current Value</th>
                <th className="px-4 py-3 text-gray-400 font-medium text-right">Action</th>
              </tr>
            </thead>
            <tbody>
              {configs.filter((cfg) => !cfg.Key.startsWith("xp_")).map((cfg) => (
                <tr key={cfg.ID} className="border-b border-line/50 hover:bg-card-hover/50">
                  <td className="px-4 py-3 text-gray-200">{labelMap[cfg.Key] || cfg.Label}</td>
                  <td className="px-4 py-3 text-gray-500 font-mono text-xs">{cfg.Key}</td>
                  <td className="px-4 py-3 text-right">
                    {editingKey === cfg.Key ? (
                      <input
                        type="number"
                        value={editValue}
                        onChange={(e) => setEditValue(e.target.value)}
                        className="w-24 bg-card-hover border border-line-light rounded px-2 py-1 text-sm text-gray-200 text-right outline-none focus:border-accent"
                        autoFocus
                        min={0}
                        step={1}
                      />
                    ) : (
                      <span className="text-accent font-semibold">{cfg.Value}</span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-right">
                    {editingKey === cfg.Key ? (
                      <div className="flex gap-2 justify-end">
                        <button
                          onClick={() => handleSave(cfg.Key)}
                          disabled={saving[cfg.Key]}
                          className="px-3 py-1 bg-green-600 hover:bg-green-700 disabled:opacity-50 text-white text-xs font-medium rounded transition-colors"
                        >
                          {saving[cfg.Key] ? "Saving..." : "Save"}
                        </button>
                        <button
                          onClick={() => setEditingKey(null)}
                          className="px-3 py-1 bg-card-hover hover:bg-line-light text-gray-300 text-xs font-medium rounded border border-line-light transition-colors"
                        >
                          Cancel
                        </button>
                      </div>
                    ) : (
                      <button
                        onClick={() => {
                          setEditingKey(cfg.Key);
                          setEditValue(String(cfg.Value));
                          setError("");
                        }}
                        className="px-3 py-1 bg-violet-600/20 hover:bg-violet-600/40 text-violet-300 text-xs font-medium rounded transition-colors"
                      >
                        Edit
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Card>

      <p className="text-xs text-gray-600 mt-4 leading-relaxed">
        These settings control ticket costs and rewards across the platform.
        Changes take effect immediately after saving.
      </p>
    </div>
  );
}
