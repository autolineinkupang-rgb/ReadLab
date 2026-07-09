"use client";

import { Suspense, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { auth } from "@/lib/api";

export default function ResetPasswordPage() {
  return (
    <Suspense fallback={<div className="min-h-[70vh] flex items-center justify-center"><div className="w-8 h-8 border-2 border-accent border-t-transparent rounded-full animate-spin" /></div>}>
      <ResetForm />
    </Suspense>
  );
}

function ResetForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get("token") || "";
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (password !== confirmPassword) {
      setError("Passwords do not match");
      return;
    }

    setLoading(true);
    try {
      await auth.resetPassword(token, password);
      setSuccess(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "An error occurred");
    } finally {
      setLoading(false);
    }
  };

  if (!token) {
    return (
      <div className="min-h-[70vh] flex items-center justify-center px-4">
        <div className="bg-card border border-line-light rounded-2xl p-8 text-center max-w-md">
          <p className="text-red-400 mb-4">Invalid reset link. No token provided.</p>
          <Link href="/en/forgot-password" className="text-violet-400 hover:text-violet-300 transition-colors text-sm">Request a new reset link</Link>
        </div>
      </div>
    );
  }

  if (success) {
    return (
      <div className="min-h-[70vh] flex items-center justify-center px-4">
        <div className="bg-card border border-line-light rounded-2xl p-8 text-center max-w-md">
          <div className="bg-emerald-900/20 border border-emerald-700/30 text-emerald-400 text-sm rounded-lg px-4 py-3 mb-4">
            Password has been reset successfully!
          </div>
          <Link href="/en/login" className="text-violet-400 hover:text-violet-300 transition-colors text-sm">Go to Login</Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-[70vh] flex items-center justify-center px-4">
      <div className="w-full max-w-md">
        <div className="bg-card border border-line-light rounded-2xl p-8">
          <h1 className="text-2xl font-bold text-white mb-2">Set New Password</h1>
          <p className="text-sm text-gray-400 mb-6">Enter your new password below</p>

          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <div className="bg-red-900/20 border border-red-700/30 text-red-400 text-sm rounded-lg px-4 py-3">
                {error}
              </div>
            )}
            <div>
              <label className="block text-sm text-gray-400 mb-1.5">New Password</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                minLength={8}
                className="w-full bg-card-hover border border-line-light rounded-lg px-4 py-2.5 text-sm text-gray-200 outline-none focus:border-accent transition-colors"
                placeholder="Min. 8 characters"
              />
            </div>
            <div>
              <label className="block text-sm text-gray-400 mb-1.5">Confirm Password</label>
              <input
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
                minLength={8}
                className="w-full bg-card-hover border border-line-light rounded-lg px-4 py-2.5 text-sm text-gray-200 outline-none focus:border-accent transition-colors"
                placeholder="Repeat password"
              />
            </div>
            <button
              type="submit"
              disabled={loading}
              className="w-full py-2.5 bg-violet-600 hover:bg-violet-700 disabled:opacity-50 text-white text-sm font-medium rounded-lg transition-colors"
            >
              {loading ? "Please wait..." : "Reset Password"}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
