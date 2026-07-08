import type { Metadata } from "next";
import Link from "next/link";
import Card from "@/components/ui/Card";

export const metadata: Metadata = {
  title: "About Us - ReadLab",
  description: "Learn more about ReadLab, the machine translation novel platform.",
};

export default function AboutUsPage() {
  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      {/* Breadcrumb */}
      <nav className="text-sm text-gray-500 mb-6">
        <Link href="/en" className="hover:text-violet-400 transition-colors">Home</Link>
        <span className="mx-2">/</span>
        <span className="text-gray-400">About Us</span>
      </nav>

      <h1 className="text-3xl font-bold text-white mb-8">About Us</h1>

      <Card className="p-8 space-y-6">
        <p className="text-sm text-gray-300 leading-relaxed">
          ReadLab is a RAW Novels translator site using automatic Machine Translation (MTL),
          so we can translate the novels faster than human translation.
        </p>

        <p className="text-sm text-gray-300 leading-relaxed">
          The reason for making this site is because there are only a few manual translator
          and it is a bit slow in translating light novels manually. Therefore, if you are
          not patient enough waiting for the manual translation of your favorite novels,
          ReadLab with its machine translation is the solution for you.
        </p>

        <p className="text-sm text-gray-300 leading-relaxed">
          Unlike other platforms, ReadLab doesn&apos;t dictate what novels you read.
          Our community drives the content! Users can request novels from supported raw
          websites, vote on their favorite requests, or use tickets to fast-release popular
          titles. You decide what gets translated next.
        </p>

        <div className="pt-4 border-t border-line">
          <p className="text-sm text-gray-500">
            If you have any questions or suggestions for us, you might contact us or
            email us at{" "}
            <a href="mailto:your-email@example.com" className="text-violet-400 hover:text-violet-300 transition-colors">
                your-email@example.com
            </a>
          </p>
        </div>
      </Card>
    </div>
  );
}
