import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "ReadLab - Read Light Novels in English Machine Translation",
  description: "Read English MTL (Machine Translation) Novels on ReadLab. All light novels here are translated from raw. Sign up to save your reading progress.",
  openGraph: {
    title: "ReadLab - Read Light Novels in English Machine Translation",
    description: "Read English MTL (Machine Translation) Novels on ReadLab. All light novels here are translated from raw.",
    type: "website",
    url: "https://readlab.my.id",
    siteName: "ReadLab",
    images: [
      {
        url: "/assets/favicon/favicon.svg",
        width: 96,
        height: 96,
        alt: "ReadLab",
      },
    ],
  },
  twitter: {
    card: "summary",
    title: "ReadLab - Read Light Novels in English Machine Translation",
    description: "Read English MTL (Machine Translation) Novels on ReadLab. All light novels here are translated from raw.",
    images: ["/assets/favicon/favicon.svg"],
  },
  robots: {
    index: true,
    follow: true,
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:ital,wght@0,100..800;1,100..800&family=Nunito+Sans:opsz,wght@6..12,200..1000&display=swap" rel="stylesheet" />
        <link rel="icon" type="image/svg+xml" href="/assets/favicon/favicon.svg" />
        <meta name="apple-mobile-web-app-title" content="ReadLab" />
        <link rel="manifest" href="/assets/favicon/site.webmanifest" />
      </head>
      <body className="min-h-screen gradient-bg">
        {children}
      </body>
    </html>
  );
}
