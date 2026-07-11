import Link from "next/link";

const footerLinks = [
  { href: "/en/about-us", label: "About Us" },
  { href: "/en/contact-us", label: "Contact Us" },
  { href: "/en/trending", label: "Trending" },
  { href: "/en/recommendation", label: "Recommendations" },
  { href: "/en/news", label: "News" },
  { href: "/en/news?type=changelog", label: "Changelog" },
  { href: "/en/dmca", label: "DMCA" },
  { href: "/en/cookie-policy", label: "Cookie Policy" },
  { href: "/en/privacy-policy", label: "Privacy Policy" },
  { href: "/en/terms-of-use", label: "Terms of Use" },
  { href: "/en/public-stats", label: "Stats" },
  { href: "/en/profile/request-serie", label: "Request Series" },
  { href: "/en/profile/vote-serie", label: "Vote Series" },
];

export default function Footer() {
  return (
    <footer className="border-t border-line mt-16" data-testid="site-footer">
      <div className="max-w-7xl mx-auto px-4 py-8">
        <div className="flex flex-wrap justify-center gap-x-6 gap-y-2 text-sm text-gray-500 mb-6">
          <Link href="/en" className="hover:text-accent-light transition-colors">Home</Link>
          {footerLinks.map((link) => (
            <Link key={link.href} href={link.href} className="hover:text-accent-light transition-colors">
              {link.label}
            </Link>
          ))}
        </div>
        <p className="text-center text-sm text-gray-600">
          Copyright © {new Date().getFullYear()} ReadLab<span className="text-accent ml-2">v1.1.0</span>
        </p>
        <p className="text-center text-xs text-gray-700 mt-2">
          Made with ♥ for readers everywhere
        </p>
      </div>
    </footer>
  );
}
