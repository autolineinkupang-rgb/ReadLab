"use client";

import { useState } from "react";
import Link from "next/link";
import { auth } from "@/lib/api";

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [sent, setSent] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      await auth.forgotPassword(email);
      setSent(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "An error occurred");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-[70vh] flex items-center justify-center px-4">
      <div className="w-full max-w-md">
        <div className="bg-card border border-line-light rounded-2xl p-8">
          <h1 className="text-2xl font-bold text-white mb-2">Reset Password</h1>
          <p className="text-sm text-gray-400 mb-6">Enter your email and we&apos;ll send you a reset link</p>

          {sent ? (
            <div className="space-y-4">
              <div className="bg-emerald-900/20 border border-emerald-700/30 text-emerald-400 text-sm rounded-lg px-4 py-3">
                If the email exists, a reset link has been sent. Check your inbox.
              </div>
              <Link href="/en" className="block text-center text-sm text-violet-400 hover:text-violet-300 transition-colors">
                Back to Home
              </Link>
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="space-y-4">
              {error && (
                <div className="bg-red-900/20 border border-red-700/30 text-red-400 text-sm rounded-lg px-4 py-3">
                  {error}
                </div>
              )}
              <div>
                <label className="block text-sm text-gray-400 mb-1.5">Email</label>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  className="w-full bg-card-hover border border-line-light rounded-lg px-4 py-2.5 text-sm text-gray-200 outline-none focus:border-accent transition-colors"
                  placeholder="you@example.com"
                />
              </div>
              <button
                type="submit"
                disabled={loading}
                className="w-full py-2.5 bg-violet-600 hover:bg-violet-700 disabled:opacity-50 text-white text-sm font-medium rounded-lg transition-colors"
              >
                {loading ? "Please wait..." : "Send Reset Link"}
              </button>
              <p className="text-center text-sm text-gray-500">
                Remember your password?{" "}
                <Link href="/en/login" className="text-violet-400 hover:text-violet-300 transition-colors">
                  Login
                </Link>
              </p>
            </form>
          )}
        </div>
      </div>
    </div>
  );
}
