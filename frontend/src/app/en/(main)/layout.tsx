import { AuthProvider } from "@/lib/AuthContext";
import Navbar from "@/components/Navbar";
import Sidebar from "@/components/Sidebar";
import Footer from "@/components/Footer";
import BackToTop from "@/components/BackToTop";

export default function MainLayout({ children }: { children: React.ReactNode }) {
  return (
    <AuthProvider>
      <Navbar />
      <Sidebar />
      <div className="lg:ml-64 min-h-screen flex flex-col">
        <main className="flex-1">{children}</main>
        <Footer />
      </div>
      <BackToTop />
    </AuthProvider>
  );
}
