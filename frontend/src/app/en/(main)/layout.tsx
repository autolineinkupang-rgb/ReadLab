"use client";

import { AuthProvider } from "@/lib/AuthContext";
import { SidebarProvider, useSidebar } from "@/lib/SidebarContext";
import Navbar from "@/components/Navbar";
import Sidebar from "@/components/Sidebar";
import Footer from "@/components/Footer";
import BackToTop from "@/components/BackToTop";
import ErrorBoundary from "@/components/ErrorBoundary";
import { ToastProvider } from "@/components/ui/Toast";

function SidebarLayout({ children }: { children: React.ReactNode }) {
  const { collapsed } = useSidebar();
  return (
    <div className={`${collapsed ? "lg:ml-16" : "lg:ml-64"} min-h-screen flex flex-col transition-all duration-300`}>
      <main className="flex-1">{children}</main>
      <Footer />
    </div>
  );
}

export default function MainLayout({ children }: { children: React.ReactNode }) {
  return (
    <AuthProvider>
      <ToastProvider>
        <ErrorBoundary>
          <SidebarProvider>
            <Navbar />
            <Sidebar />
            <SidebarLayout>{children}</SidebarLayout>
          </SidebarProvider>
          <BackToTop />
        </ErrorBoundary>
      </ToastProvider>
    </AuthProvider>
  );
}