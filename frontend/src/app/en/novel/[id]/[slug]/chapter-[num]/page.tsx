"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";

interface ChapterData {
  ID: number;
  Number: number;
  Title: string;
  Content: string;
  NovelID: number;
  IsLocked: boolean;
}

interface NovelInfo {
  ID: number;
  Title: string;
  Slug: string;
  Chapters: number;
}

function generateMockContent(chapterNum: number): string {
  const paragraphs = [
    "The morning sun filtered through the thin curtains, casting golden stripes across the wooden floor. Dust motes danced lazily in the warm light, and the distant sound of birds could be heard from the garden outside.",
    "She stood at the window, her reflection barely visible in the glass. The cup of tea in her hands had long gone cold, but she didn't notice. Her mind was elsewhere, wandering through memories that felt both distant and painfully close.",
    "He was running. His lungs burned with each breath, and his legs felt like lead. But he couldn't stop. Not now. Not when he was so close. The roaring of the wind in his ears drowned out everything else, leaving only the primal instinct to survive.",
    "The room was silent except for the ticking of the old grandfather clock. Each second stretched into an eternity as they sat across from each other, words hanging unspoken in the air between them. Someone had to break first.",
    "The city lights flickered to life as dusk settled over the skyline. From this height, everything below looked like a miniature world—cars threading through streets like beads on a string, people reduced to tiny specks hurrying home.",
  ];

  const content = [];
  for (let i = 0; i < 8 + (chapterNum % 5); i++) {
    content.push(paragraphs[i % paragraphs.length]);
  }
  content.push("", `--- End of Chapter ${chapterNum} ---`);
  return content.join("\n\n");
}

export default function ChapterReaderPage() {
  const params = useParams();
  const id = params?.id as string;
  const slug = params?.slug as string;
  const numStr = params?.num as string;

  const [chapter, setChapter] = useState<ChapterData | null>(null);
  const [novel, setNovel] = useState<NovelInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [fontSize, setFontSize] = useState(18);

  const chapterNum = parseInt(numStr?.replace("chapter-", "") || "1");

  useEffect(() => {
    if (!id) return;
    setLoading(true);

    const ch: ChapterData = {
      ID: chapterNum,
      Number: chapterNum,
      Title: `Chapter ${chapterNum}`,
      Content: generateMockContent(chapterNum),
      NovelID: parseInt(id),
      IsLocked: chapterNum > 70,
    };

    const nv: NovelInfo = {
      ID: parseInt(id),
      Title: slug ? slug.replace(/-/g, " ") : "Novel",
      Slug: slug || "",
      Chapters: 135,
    };

    setTimeout(() => {
      setChapter(ch);
      setNovel(nv);
      setLoading(false);
    }, 200);
  }, [id, chapterNum, slug]);

  if (loading || !chapter || !novel) {
    return (
      <div className="max-w-3xl mx-auto px-4 py-16">
        <div className="animate-pulse space-y-4">
          <div className="h-6 bg-[#1e1e3a] rounded w-1/3 mx-auto" />
          <div className="h-4 bg-[#1e1e3a] rounded w-1/4 mx-auto" />
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="h-3 bg-[#1e1e3a] rounded w-full" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto px-4 py-6">
      {/* Breadcrumb */}
      <nav className="text-xs text-gray-500 mb-4">
        <Link href="/en" className="hover:text-violet-400 transition-colors">Home</Link>
        <span className="mx-1">/</span>
        <Link href="/en/novel-list" className="hover:text-violet-400 transition-colors">Novels</Link>
        <span className="mx-1">/</span>
        <Link href={`/en/novel/${id}/${slug}`} className="hover:text-violet-400 transition-colors">
          {novel.Title.length > 40 ? novel.Title.slice(0, 40) + "..." : novel.Title}
        </Link>
        <span className="mx-1">/</span>
        <span className="text-gray-400">Chapter {chapter.Number}</span>
      </nav>

      {/* Reader Settings */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-2">
          <button
            onClick={() => setFontSize(Math.max(14, fontSize - 2))}
            className="w-8 h-8 rounded-lg bg-[#1e1e3a] text-gray-400 hover:text-white flex items-center justify-center text-sm transition-colors"
          >
            A-
          </button>
          <span className="text-xs text-gray-500 w-8 text-center">{fontSize}</span>
          <button
            onClick={() => setFontSize(Math.min(28, fontSize + 2))}
            className="w-8 h-8 rounded-lg bg-[#1e1e3a] text-gray-400 hover:text-white flex items-center justify-center text-sm transition-colors"
          >
            A+
          </button>
        </div>
      </div>

      {/* Chapter title */}
      <h1 className="text-xl font-bold text-white text-center mb-8">
        {chapter.Title}
      </h1>

      {/* Content */}
      <div
        className="text-gray-300 leading-relaxed space-y-6"
        style={{ fontSize: `${fontSize}px` }}
      >
        {chapter.Content.split("\n\n").map((para, i) => (
          <p key={i} className="text-justify">{para}</p>
        ))}
      </div>

      {/* Navigation */}
      <div className="flex items-center justify-between mt-12 pt-6 border-t border-[#1e1e3a]">
        {chapter.Number > 1 ? (
          <Link
            href={`/en/novel/${id}/${slug}/chapter-${chapter.Number - 1}`}
            className="flex items-center gap-2 px-4 py-2 bg-[#1e1e3a] hover:bg-[#2a2a4a] rounded-lg text-sm text-gray-300 transition-colors"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
            Previous Chapter
          </Link>
        ) : (
          <div />
        )}

        <Link
          href={`/en/novel/${id}/${slug}`}
          className="text-sm text-violet-400 hover:text-violet-300 transition-colors"
        >
          Novel Page
        </Link>

        {chapter.Number < novel.Chapters ? (
          <Link
            href={`/en/novel/${id}/${slug}/chapter-${chapter.Number + 1}`}
            className="flex items-center gap-2 px-4 py-2 bg-[#1e1e3a] hover:bg-[#2a2a4a] rounded-lg text-sm text-gray-300 transition-colors"
          >
            Next Chapter
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </Link>
        ) : (
          <div />
        )}
      </div>

      {/* Chapter select */}
      <div className="mt-6 text-center">
        <Link
          href={`/en/novel/${id}/${slug}`}
          className="text-sm text-gray-500 hover:text-violet-400 transition-colors"
        >
          ← Back to Table of Contents
        </Link>
      </div>
    </div>
  );
}
