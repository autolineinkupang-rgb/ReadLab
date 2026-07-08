import type { Metadata } from "next";
import Link from "next/link";
import Card from "@/components/ui/Card";

export const metadata: Metadata = {
  title: "Privacy Policy - ReadLab",
};

export default function PrivacyPolicyPage() {
  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <nav className="text-sm text-gray-500 mb-6">
        <Link href="/en" className="hover:text-violet-400 transition-colors">Home</Link>
        <span className="mx-2">/</span>
        <span className="text-gray-400">Privacy Policy</span>
      </nav>
      <h1 className="text-3xl font-bold text-white mb-8">Privacy Policy</h1>
      <Card className="p-8 space-y-4 text-sm text-gray-300 leading-relaxed">
        <p>Your privacy is important to us. This Privacy Policy explains how ReadLab collects, uses, and protects your personal information.</p>
        <h2 className="text-lg font-semibold text-white mt-6">Information We Collect</h2>
        <ul className="list-disc pl-5 space-y-1">
          <li><strong>Account Information:</strong> Username, email address, and password (encrypted) when you register</li>
          <li><strong>Reading Data:</strong> Your reading history, bookmarks, and preferences</li>
          <li><strong>Usage Data:</strong> Pages visited, time spent, and interactions with the site</li>
        </ul>
        <h2 className="text-lg font-semibold text-white mt-6">How We Use Your Information</h2>
        <ul className="list-disc pl-5 space-y-1">
          <li>To provide and maintain our service</li>
          <li>To personalize your reading experience</li>
          <li>To improve our website and features</li>
          <li>To communicate with you about updates and changes</li>
        </ul>
        <h2 className="text-lg font-semibold text-white mt-6">Data Protection</h2>
        <p>We implement appropriate security measures to protect your personal information. Passwords are encrypted and we never share your data with third parties without your consent.</p>
        <h2 className="text-lg font-semibold text-white mt-6">Contact</h2>
        <p>For privacy-related inquiries, please contact us at <a href="mailto:your-email@example.com" className="text-violet-400 hover:text-violet-300">your-email@example.com</a>.</p>
      </Card>
    </div>
  );
}
