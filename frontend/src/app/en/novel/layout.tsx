import { AuthProvider } from "@/lib/AuthContext";

export default function NovelLayout({ children }: { children: React.ReactNode }) {
  return <AuthProvider>{children}</AuthProvider>;
}
