"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/AuthContext";

interface RequireRoleProps {
  roles: string[];
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

export default function RequireRole({ roles, children, fallback }: RequireRoleProps) {
  const { user, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading && !user) {
      router.replace("/en/login");
    }
  }, [loading, user, router]);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="w-8 h-8 border-2 border-accent border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  if (!user) return null;

  if (!roles.includes(user.role)) {
    if (fallback) return <>{fallback}</>;
    return (
      <div className="max-w-md mx-auto px-4 py-16 text-center">
        <div className="p-10 bg-card border border-line rounded-xl">
          <h2 className="text-lg font-semibold text-white mb-2">Access Denied</h2>
          <p className="text-sm text-gray-500 mb-6">You do not have permission to access this page.</p>
        </div>
      </div>
    );
  }

  return <>{children}</>;
}
