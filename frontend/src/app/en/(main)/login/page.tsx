"use client";

import { Suspense, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useAuth } from "@/lib/AuthContext";

function passwordStrength(pw: string): { score: number; label: string; color: string } {
  let score = 0;
  if (pw.length >= 8) score++;
  if (pw.length >= 12) score++;
  if (/[a-z]/.test(pw)) score++;
  if (/[A-Z]/.test(pw)) score++;
  if (/[0-9]/.test(pw)) score++;
  if (/[^a-zA-Z0-9]/.test(pw)) score++;

  if (score <= 2) return { score, label: "Weak", color: "bg-red-500" };
  if (score <= 4) return { score, label: "Fair", color: "bg-yellow-500" };
  if (score <= 5) return { score, label: "Good", color: "bg-blue-500" };
  return { score: 6, label: "Strong", color: "bg-green-500" };
}

export default function LoginPage() {
  return (
    <Suspense fallback={<div className="min-h-[70vh] flex items-center justify-center"><div className="w-8 h-8 border-2 border-accent border-t-transparent rounded-full animate-spin" /></div>}>
      <LoginForm />
    </Suspense>
  );
}

function LoginForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { login, register } = useAuth();
  const [mode, setMode] = useState<"login" | "register">((searchParams.get("mode") as "login" | "register") || "login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [username, setUsername] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const strength = useMemo(() => passwordStrength(password), [password]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      if (mode === "login") {
        await login(email, password);
      } else {
        await register(username, email, password);
      }
      router.push("/en");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "An error occurred");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-[70vh] flex items-center justify-center px-4">
      <div className="w-full max-w-md">
        <h1 className="text-2xl font-bold text-white text-center mb-8">
          {mode === "login" ? "Welcome Back" : "Create Account"}
        </h1>

        {/* Tab toggle */}
        <div className="flex bg-card-hover rounded-lg p-0.5 mb-6">
          <button
            onClick={() => { setMode("login"); setError(""); }}
            className={`flex-1 py-2 text-sm rounded-md transition-colors ${
              mode === "login" ? "bg-violet-600 text-white" : "text-gray-400 hover:text-white"
            }`}
          >
            Login
          </button>
          <button
            onClick={() => { setMode("register"); setError(""); }}
            className={`flex-1 py-2 text-sm rounded-md transition-colors ${
              mode === "register" ? "bg-violet-600 text-white" : "text-gray-400 hover:text-white"
            }`}
          >
            Register
          </button>
        </div>

        <form onSubmit={handleSubmit} className="bg-card border border-line rounded-xl p-6 space-y-4">
          {error && (
            <div className="p-3 rounded-lg bg-red-900/30 border border-red-800/30 text-sm text-red-400">
              {error}
            </div>
          )}

          {mode === "register" && (
            <div>
              <label className="block text-sm text-gray-400 mb-1">Username</label>
              <input
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
                minLength={3}
                autoComplete="username"
                placeholder="Your username"
                className="w-full bg-card-hover border border-line-light rounded-lg px-4 py-2.5 text-sm text-gray-200 outline-none focus:border-violet-600 transition-colors"
              />
            </div>
          )}

          <div>
            <label className="block text-sm text-gray-400 mb-1">Email</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              autoComplete="email"
              placeholder="your@email.com"
              className="w-full bg-card-hover border border-line-light rounded-lg px-4 py-2.5 text-sm text-gray-200 outline-none focus:border-violet-600 transition-colors"
            />
          </div>

          <div>
            <label className="block text-sm text-gray-400 mb-1">Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={8}
              autoComplete={mode === "register" ? "new-password" : "current-password"}
              placeholder="••••••••"
              className="w-full bg-card-hover border border-line-light rounded-lg px-4 py-2.5 text-sm text-gray-200 outline-none focus:border-violet-600 transition-colors"
            />
            {mode === "register" && password.length > 0 && (
              <div className="mt-2">
                <div className="flex gap-1">
                  {[1, 2, 3, 4, 5, 6].map((i) => (
                    <div
                      key={i}
                      className={`h-1 flex-1 rounded-full transition-colors ${
                        i <= strength.score ? strength.color : "bg-gray-700"
                      }`}
                    />
                  ))}
                </div>
                <p className={`text-xs mt-1 ${strength.score <= 2 ? "text-red-400" : strength.score <= 4 ? "text-yellow-400" : "text-green-400"}`}>
                  {strength.label}
                </p>
              </div>
            )}
            {mode === "register" && (
              <p className="text-[10px] text-gray-600 mt-1">
                Min 8 characters with uppercase, lowercase, number &amp; special character
              </p>
            )}
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full py-2.5 bg-violet-600 hover:bg-violet-700 disabled:opacity-50 text-white text-sm font-medium rounded-lg transition-colors"
          >
            {loading ? "Please wait..." : mode === "login" ? "Login" : "Create Account"}
          </button>
        </form>

        {mode === "login" && (
          <p className="text-center text-xs mt-3">
            <Link href="/en/forgot-password" className="text-gray-500 hover:text-violet-400 transition-colors">
              Forgot Password?
            </Link>
          </p>
        )}

        <p className="text-center text-xs text-gray-600 mt-4">
          {mode === "login"
            ? "Don't have an account? Switch to Register above."
            : "Already have an account? Switch to Login above."}
        </p>
      </div>
    </div>
  );
}
