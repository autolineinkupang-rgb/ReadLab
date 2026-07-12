"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { adminBank } from "@/lib/api";
import Card from "@/components/ui/Card";
import RequireRole from "@/components/RequireRole";

export default function AdminBankPage() {
  return (
    <RequireRole roles={["admin"]}>
      <AdminBank />
    </RequireRole>
  );
}

function AdminBank() {
  const [balance, setBalance] = useState(0);
  const [units, setUnits] = useState(0);
  const [loading, setLoading] = useState(true);
  const [amount, setAmount] = useState("");
  const [claiming, setClaiming] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const fetchBalance = () => {
    setLoading(true);
    adminBank.balance()
      .then((res) => {
        setBalance(res.balance);
        setUnits(res.units);
      })
      .catch(() => setError("Failed to load bank balance"))
      .finally(() => setLoading(false));
  };

  useEffect(() => { fetchBalance(); }, []);

  const handleClaim = async () => {
    const amt = parseFloat(amount);
    if (isNaN(amt) || amt <= 0) return;
    setClaiming(true);
    setError("");
    setSuccess("");
    try {
      await adminBank.claim(amt);
      setSuccess(`Claimed ${amt} tickets from bank`);
      setAmount("");
      fetchBalance();
    } catch (e: any) {
      setError(e?.message || "Failed to claim tickets");
    } finally {
      setClaiming(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto px-4 py-8">
      <Link href="/en/admin" className="inline-flex items-center gap-1 text-sm text-gray-500 hover:text-gray-300 mb-4 transition-colors">
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={2}>
          <path d="M15 19l-7-7 7-7" />
        </svg>
        Back to Admin
      </Link>
      <h1 className="text-2xl font-bold text-white mb-6">Ticket Bank</h1>

      {error && (
        <div className="mb-4 p-3 bg-red-900/30 border border-red-700 rounded-lg text-sm text-red-300">
          {error}
          <button onClick={() => setError("")} className="ml-2 text-red-200 hover:text-white">&times;</button>
        </div>
      )}

      {success && (
        <div className="mb-4 p-3 bg-green-900/30 border border-green-700 rounded-lg text-sm text-green-300">
          {success}
          <button onClick={() => setSuccess("")} className="ml-2 text-green-200 hover:text-white">&times;</button>
        </div>
      )}

      <Card className="mb-6">
        <div className="p-6">
          {loading ? (
            <div className="text-sm text-gray-500">Loading...</div>
          ) : (
            <div className="grid grid-cols-2 gap-6">
              <div>
                <div className="text-xs text-gray-500 uppercase tracking-wide mb-1">Bank Balance</div>
                <div className="text-3xl font-bold text-accent">{balance.toLocaleString()}</div>
              </div>
              <div>
                <div className="text-xs text-gray-500 uppercase tracking-wide mb-1">Ticket Units</div>
                <div className="text-3xl font-bold text-gray-200">{units.toLocaleString()}</div>
              </div>
            </div>
          )}
        </div>
      </Card>

      <Card>
        <div className="p-6">
          <h2 className="text-lg font-semibold text-white mb-4">Claim Tickets from Bank</h2>
          <div className="flex items-center gap-3">
            <input
              type="number"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              onKeyDown={(e) => { if (e.key === "Enter") handleClaim(); }}
              className="flex-1 bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none focus:border-accent"
              placeholder="Amount to claim"
              step="any"
              min="0.01"
            />
            <button
              onClick={handleClaim}
              disabled={claiming || !amount || parseFloat(amount) <= 0}
              className="px-4 py-2 bg-accent hover:bg-accent/80 disabled:opacity-50 text-white text-sm font-medium rounded-lg transition-colors"
            >
              {claiming ? "Claiming..." : "Claim"}
            </button>
          </div>
        </div>
      </Card>
    </div>
  );
}
