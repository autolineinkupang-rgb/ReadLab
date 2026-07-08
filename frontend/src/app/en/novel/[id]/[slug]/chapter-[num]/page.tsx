"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useParams, usePathname } from "next/navigation";
import { novels, chapters as chaptersApi } from "@/lib/api";
import ChapterReader from "@/components/ChapterReader";

export default function ChapterReaderPage() {
  const params = useParams();
  const pathname = usePathname();
  const slug = params?.slug as string;
  const rawId = params?.id as string;

  const [chapter, setChapter] = useState<{ number: number; title: string; content: string; isLocked: boolean } | null>(null);
  const [novel, setNovel] = useState<{ id: number; slug: string; title: string; totalChapters: number; coverUrl?: string; description?: string; author?: string; sourceUrl?: string } | null>(null);
  const [chapters, setChapters] = useState<{ number: number; title: string; createdAt?: string }[]>([]);
  const [loading, setLoading] = useState(true);
  const [chapterLoading, setChapterLoading] = useState(false);
  const [inLibrary, setInLibrary] = useState(false);
  const [error, setError] = useState("");

  const novelCache = useRef<{ id: number } | null>(null);

  const chapterNum = (() => {
    const match = pathname.match(/chapter-(\d+)$/);
    return match ? parseInt(match[1]) : 1;
  })();

  useEffect(() => {
    if (!rawId) return;

    setError("");

    const novelId = parseInt(rawId);
    const isNewNovel = !novelCache.current || novelCache.current.id !== novelId;

    if (isNewNovel) {
      setLoading(true);
      setChapter(null);

      Promise.all([
        novels.get(novelId),
        novels.chapters(novelId, { page: 1, limit: 9999 }),
        chaptersApi.getByNovel(novelId, chapterNum),
      ])
        .then(([novelRes, chListRes, chRes]) => {
          const n = novelRes;
          const chList = (chListRes.data || []).map((ch: any) => ({
            number: ch.Number,
            title: ch.Title || "",
            createdAt: ch.CreatedAt,
          }));

          setNovel({
            id: n.ID,
            slug: n.Slug || slug,
            title: n.Title,
            totalChapters: n.Chapters || chList.length,
            coverUrl: n.CoverURL || "",
            description: n.Description || "",
            author: n.Author || "",
            sourceUrl: n.SourceURL || "",
          });
          setChapters(chList);
          setChapter({
            number: chRes.Number,
            title: chRes.Title || "",
            content: chRes.Content || "",
            isLocked: chRes.IsLocked || false,
          });
          novelCache.current = { id: novelId };
        })
        .catch((e: any) => {
          setError(e.message || "Failed to load chapter");
        })
        .finally(() => {
          setLoading(false);
        });
    } else {
      setChapterLoading(true);

      chaptersApi.getByNovel(novelId, chapterNum)
        .then((chRes) => {
          setChapter({
            number: chRes.Number,
            title: chRes.Title || "",
            content: chRes.Content || "",
            isLocked: chRes.IsLocked || false,
          });
        })
        .catch((e: any) => {
          setError(e.message || "Failed to load chapter");
        })
        .finally(() => {
          setChapterLoading(false);
        });
    }
  }, [pathname]);

  const totalChapters = novel?.totalChapters || 0;

  const handleAddToLibrary = useCallback(() => {
    setInLibrary((prev) => !prev);
  }, []);

  const novelHref = `/en/novel/${rawId}/${slug}`;
  const prevHref = chapterNum > 1 ? `/en/novel/${rawId}/${slug}/chapter-${chapterNum - 1}` : undefined;
  const nextHref = chapterNum < totalChapters ? `/en/novel/${rawId}/${slug}/chapter-${chapterNum + 1}` : undefined;

  return (
    <ChapterReader
      chapter={chapter}
      novel={novel}
      chapters={chapters}
      loading={loading}
      chapterLoading={chapterLoading}
      error={error}
      prevHref={prevHref}
      nextHref={nextHref}
      novelHref={novelHref}
      onAddToLibrary={handleAddToLibrary}
      inLibrary={inLibrary}
    />
  );
}
