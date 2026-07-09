"use client";

import { useState } from "react";
import { useAuth } from "@/lib/AuthContext";
import { tickets } from "@/lib/api";

const PACKAGES = [
  { amount: 10, label: "10 Tickets", price: "$1.00" },
  { amount: 50, label: "50 Tickets", price: "$4.00", popular: true },
  { amount: 100, label: "100 Tickets", price: "$7.00" },
  { amount: 500, label: "500 Tickets", price: "$30.00" },
];

export default function TicketsPage() {
  const { user, refresh } = useAuth();
  const [selected, setSelected] = useState(50);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState("");

  const handlePurchase = async () => {
    setLoading(true);
    setMessage("");
    try {
      const res = await tickets.purchase(selected);
      setMessage(`Successfully purchased ${res.amount} tickets!`);
      refresh();
    } catch (err) {
      setMessage(err instanceof Error ? err.message : "Purchase failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold text-white mb-2">Get Tickets</h1>
      <p className="text-sm text-gray-400 mb-6">
        Current balance: <span className="text-accent-light font-semibold">{user?.tickets ?? 0} Tickets</span>
      </p>

      <div className="grid grid-cols-2 gap-4 mb-6">
        {PACKAGES.map((pkg) => (
          <button
            key={pkg.amount}
            onClick={() => setSelected(pkg.amount)}
            className={`relative rounded-xl border-2 p-4 text-left transition-all ${
              selected === pkg.amount
                ? "border-violet-500 bg-violet-900/20"
                : "border-line-light bg-card-hover hover:border-gray-600"
            }`}
          >
            {pkg.popular && (
              <span className="absolute -top-2 right-3 bg-violet-600 text-white text-[10px] font-bold px-2 py-0.5 rounded-full">
                POPULAR
              </span>
            )}
            <p className="text-lg font-bold text-white">{pkg.label}</p>
            <p className="text-sm text-gray-400">{pkg.price}</p>
          </button>
        ))}
      </div>

      {message && (
        <div className={`text-sm rounded-lg px-4 py-3 mb-4 ${
          message.includes("Successfully")
            ? "bg-emerald-900/20 border border-emerald-700/30 text-emerald-400"
            : "bg-red-900/20 border border-red-700/30 text-red-400"
        }`}>
          {message}
        </div>
      )}

      <button
        onClick={handlePurchase}
        disabled={loading}
        className="w-full py-3 bg-violet-600 hover:bg-violet-700 disabled:opacity-50 text-white font-medium rounded-xl transition-colors"
      >
        {loading ? "Processing..." : `Purchase ${selected} Tickets`}
      </button>

      <p className="text-xs text-gray-600 text-center mt-4">
        This is a mock purchase. No real payment will be processed.
      </p>
    </div>
  );
}
