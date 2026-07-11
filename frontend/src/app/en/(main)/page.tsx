"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import NovelCard from "@/components/NovelCard";
import SectionHeader from "@/components/SectionHeader";
import NovelCardSmall from "@/components/NovelCardSmall";
import UpdateItem from "@/components/UpdateItem";
import HeroBanner from "@/components/HeroBanner";
import { NovelCardSkeletonRow, UpdateItemSkeleton, Skeleton } from "@/components/Skeleton";
import { novels, updates as updatesApi, news as newsApi, leaderboard, library as libraryApi } from "@/lib/api";
import { stripHtml } from "@/lib/utils";
import Card from "@/components/ui/Card";
import GenreTag from "@/components/ui/GenreTag";
import { Novel, LatestNewsItem } from "@/types";
import { useAuth } from "@/lib/AuthContext";

export default function Home() {
  const { user } = useAuth();
  const [recentUpdates, setRecentUpdates] = useState<any[]>([]);
  const [newNovels, setNewNovels] = useState<any[]>([]);
  const [ranking, setRanking] = useState<any[]>([]);
  const [recommendations, setRecommendations] = useState<Novel[]>([]);
  const [randomNovels, setRandomNovels] = useState<Novel[]>([]);
  const [newsItems, setNewsItems] = useState<LatestNewsItem[]>([]);
  const [topSpenders, setTopSpenders] = useState<any[]>([]);
  const [readingHistory, setReadingHistory] = useState<any[]>([]);

  const [loadingUpdates, setLoadingUpdates] = useState(true);
  const [loadingNew, setLoadingNew] = useState(true);
  const [loadingRanking, setLoadingRanking] = useState(true);
  const [loadingRecs, setLoadingRecs] = useState(true);
  const [loadingRandom, setLoadingRandom] = useState(true);

  useEffect(() => {
    updatesApi.recent(6)
      .then((res) => {
        if (res.data?.length) {
          setRecentUpdates(res.data.map((u: any) => ({
            title: u.Novel?.Title || "Unknown",
            chapter: `#${u.Number} ${u.Title || ""}`,
            chapterHref: `/en/novel/${u.NovelID}/${u.Novel?.Slug || ""}/chapter-${u.Number}`,
            novelHref: `/en/novel/${u.NovelID}/${u.Novel?.Slug || ""}`,
            image: u.Novel?.CoverURL,
            hasImage: true,
          })));
        }
      })
      .catch(() => {})
      .finally(() => setLoadingUpdates(false));

    novels.list({ sort: "created_at", order: "desc", limit: 10 })
      .then((res) => { if (res.data?.length) setNewNovels(res.data); })
      .catch(() => {})
      .finally(() => setLoadingNew(false));

    novels.list({ sort: "views", order: "desc", limit: 5 })
      .then((res) => { if (res.data?.length) setRanking(res.data); })
      .catch(() => {})
      .finally(() => setLoadingRanking(false));

    novels.recommendations()
      .then((res) => { if (res.data?.length) setRecommendations(res.data); })
      .catch(() => {})
      .finally(() => setLoadingRecs(false));

    novels.random(6)
      .then((res) => { if (res.data?.length) setRandomNovels(res.data); })
      .catch(() => {})
      .finally(() => setLoadingRandom(false));

    newsApi.list({ limit: 3 })
      .then((res) => { if (res.data?.length) setNewsItems(res.data); })
      .catch(() => {});

    leaderboard.get("tickets")
      .then((res) => { if (res.data?.length) setTopSpenders(res.data.slice(0, 3)); })
      .catch(() => {});

    if (user) {
      libraryApi.get().then((res) => {
        if (res.history?.length) setReadingHistory(res.history);
      }).catch(() => {});
    }
  }, [user]);

  return (
    <div className="max-w-7xl mx-auto px-4 py-6 space-y-10">
      {/* Hero Banner (dynamic — pulls latest news) */}
      <HeroBanner latestNews={newsItems[0] || null} />

      {/* Login Prompt */}
      {!user && (
        <div className="text-center text-sm text-gray-500" data-testid="login-prompt">
          Login to keep track of where you left off in the novel.
        </div>
      )}

      {/* Continue Reading */}
      {user && readingHistory.length > 0 && (
        <section data-testid="continue-reading-section">
          <SectionHeader title="Continue Reading" />
          <div className="flex gap-4 overflow-x-auto pb-2 scrollbar-hide animate-stagger">
            {readingHistory.slice(0, 6).map((h: any) => (
              <Link
                key={h.ID}
                href={`/en/novel/${h.Novel.ID}/${h.Novel.Slug}/chapter-${h.Chapter.Number}`}
                className="flex-shrink-0 w-36 sm:w-40 group focus:outline-none focus-visible:ring-2 focus-visible:ring-accent rounded-xl"
              >
                <div className="aspect-[3/4] rounded-xl bg-card-hover border border-line-light overflow-hidden mb-2">
                  {h.Novel.CoverURL ? (
                    <img src={h.Novel.CoverURL} alt="" loading="lazy" className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500" onError={(e) => { (e.currentTarget as HTMLImageElement).style.display = "none"; }} />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center text-gray-600">
                      <svg className="w-10 h-10" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
                      </svg>
                    </div>
                  )}
                </div>
                <p className="text-sm text-gray-200 font-medium line-clamp-1 group-hover:text-accent-light transition-colors">{h.Novel.Title}</p>
                <p className="text-xs text-gray-500 mt-0.5">Chapter {h.Chapter.Number}</p>
              </Link>
            ))}
          </div>
        </section>
      )}

      {/* New Novels */}
      <section data-testid="new-novels-section">
        <SectionHeader title="New Novels" href="/en/novel-list" />
        {loadingNew ? (
          <NovelCardSkeletonRow count={6} />
        ) : newNovels.length === 0 ? (
          <EmptyState message="No new novels yet — check back soon." />
        ) : (
          <div className="flex gap-4 overflow-x-auto pb-2 scrollbar-hide animate-stagger">
            {newNovels.map((novel) => (
              <NovelCard
                key={novel.ID}
                title={novel.Title}
                genre={novel.Genres?.[0]?.Slug || "action"}
                chapters={novel.Chapters}
                rating={novel.Rating?.toString()}
                image={novel.CoverURL}
                href={`/en/novel/${novel.ID}/${novel.Slug}`}
              />
            ))}
          </div>
        )}
      </section>

      {/* Novel Ranking */}
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
        <div className="lg:col-span-2">
          <SectionHeader title="Novel Ranking" href="/en/ranking/daily" tabs={[{ label: "Daily", active: true }, { label: "Weekly" }, { label: "Monthly" }]} />
          <Card className="space-y-1">
            {loadingRanking ? (
              Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="flex items-center gap-3 p-2">
                  <Skeleton className="w-6 h-5" />
                  <div className="flex-1 space-y-2">
                    <Skeleton className="h-3 w-3/4" />
                    <Skeleton className="h-2 w-1/2" />
                  </div>
                </div>
              ))
            ) : ranking.length === 0 ? (
              <EmptyState message="No ranking data available." />
            ) : (
              ranking.map((novel, i) => (
                <NovelCardSmall
                  key={novel.ID}
                  rank={i + 1}
                  title={novel.Title}
                  views={novel.Views?.toLocaleString() ?? "0"}
                  rating={novel.Rating?.toFixed(1) ?? "0.0"}
                  image={novel.CoverURL}
                  href={`/en/novel/${novel.ID}/${novel.Slug}`}
                />
              ))
            )}
          </Card>
          <div className="flex gap-3 mt-4">
            <Link href="/en/profile/vote-serie" className="flex-1 text-center text-sm py-2.5 rounded-lg bg-card-hover hover:bg-line-light text-gray-300 transition-colors">
              Vote Novels
            </Link>
            <Link href="/en/profile/request-serie" className="flex-1 text-center text-sm py-2.5 rounded-lg bg-card-hover hover:bg-line-light text-gray-300 transition-colors">
              Request Novels
            </Link>
          </div>
        </div>

        <div className="lg:col-span-2">
          <SectionHeader title="Community" tabs={[{ label: "Community Folders" }, { label: "Trending", active: true }]} />
          <Card className="p-5">
            {loadingRecs ? (
              <div className="flex gap-4">
                <Skeleton className="w-20 sm:w-28 aspect-[3/4] rounded-lg" />
                <div className="flex-1 space-y-3">
                  <Skeleton className="h-5 w-3/4" />
                  <Skeleton className="h-3 w-1/2" />
                  <Skeleton className="h-3 w-full" />
                  <Skeleton className="h-3 w-5/6" />
                </div>
              </div>
            ) : (
              <div className="flex gap-4">
                <div className="w-20 sm:w-28 aspect-[3/4] rounded-lg bg-card-hover border border-line-light flex-shrink-0 flex items-center justify-center overflow-hidden">
                  {recommendations[0]?.CoverURL ? (
                    <img src={recommendations[0].CoverURL} alt="" loading="lazy" className="w-full h-full object-cover" onError={(e) => { (e.currentTarget as HTMLImageElement).style.display = "none"; }} />
                  ) : (
                    <svg className="w-8 h-8 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
                    </svg>
                  )}
                </div>
                <div className="min-w-0 flex-1">
                  <Link href={recommendations[0] ? `/en/novel/${recommendations[0].ID}/${recommendations[0].Slug}` : "#"} className="text-base font-semibold text-white hover:text-accent-light transition-colors line-clamp-2">
                    {recommendations[0]?.Title || "No recommendations yet"}
                  </Link>
                  {recommendations[0] && (
                    <>
                      <div className="flex items-center gap-3 mt-1 text-sm text-gray-400">
                        <span className="text-yellow-500">★ {recommendations[0].Rating?.toFixed(1)}</span>
                        <span>📚 {recommendations[0].Views?.toLocaleString()}</span>
                      </div>
                      <div className="flex flex-wrap gap-1.5 mt-2">
                        {recommendations[0].Genres?.slice(0, 4).map((g: any) => (
                          <GenreTag key={g.Slug} label={g.Slug} />
                        ))}
                      </div>
                      <p className="text-sm text-gray-400 mt-2 line-clamp-3 leading-relaxed">
                        {stripHtml(recommendations[0].Description || "")}
                      </p>
                      <Link
                        href={`/en/novel/${recommendations[0].ID}/${recommendations[0].Slug}`}
                        className="inline-block mt-3 px-4 py-1.5 bg-accent hover:bg-accent-dark text-white text-sm rounded-lg transition-colors"
                      >
                        START READING
                      </Link>
                    </>
                  )}
                </div>
              </div>
            )}
          </Card>
        </div>
      </div>

      {/* Recommendations */}
      <section data-testid="recommendations-section">
        <SectionHeader title="Recommendations" href="/en/recommendation" />
        {loadingRecs ? (
          <NovelCardSkeletonRow count={6} compact />
        ) : recommendations.length === 0 ? (
          <EmptyState message="No recommendations available yet." />
        ) : (
          <div className="flex gap-4 overflow-x-auto pb-2 scrollbar-hide animate-stagger">
            {recommendations.slice(0, 6).map((novel) => (
              <NovelCard
                key={novel.ID}
                title={novel.Title}
                genre={novel.Genres?.[0]?.Slug || "action"}
                chapters={novel.Chapters}
                rating={novel.Rating?.toString()}
                image={novel.CoverURL}
                href={`/en/novel/${novel.ID}/${novel.Slug}`}
                compact
              />
            ))}
          </div>
        )}
      </section>

      {/* Bug Reports / Patreon */}
      <div className="flex flex-wrap gap-4 justify-center">
        <Link href="https://discord.gg/YOUR_DISCORD_INVITE" className="flex items-center gap-2 px-5 py-2.5 bg-card-hover hover:bg-line-light rounded-lg text-sm text-gray-300 transition-colors">
          <svg className="w-5 h-5 text-accent-light" viewBox="0 0 24 24" fill="currentColor">
            <path d="M20.317 4.3698a19.7913 19.7913 0 00-4.8851-1.5152.0741.0741 0 00-.0785.0371c-.211.3753-.4447.8648-.6083 1.2495-1.8447-.2762-3.68-.2762-5.4868 0-.1636-.3933-.4058-.8742-.6177-1.2495a.077.077 0 00-.0785-.037 19.7363 19.7363 0 00-4.8852 1.515.0699.0699 0 00-.0321.0277C.5334 9.0458-.319 13.5799.0992 18.0578a.0824.0824 0 00.0312.0561c2.0528 1.5076 4.0413 2.4228 5.9929 3.0294a.0777.0777 0 00.0842-.0276c.4616-.6304.8731-1.2952 1.226-1.9942a.076.076 0 00-.0416-.1057c-.6528-.2476-1.2743-.5495-1.8722-.8923a.077.077 0 01-.0076-.1277c.1258-.0943.2517-.1923.3718-.2914a.0743.0743 0 01.0776-.0105c3.9278 1.7933 8.18 1.7933 12.0614 0a.0739.0739 0 01.0785.0095c.1202.099.246.1981.3728.2924a.077.077 0 01-.0066.1276 12.2986 12.2986 0 01-1.873.8914.0766.0766 0 00-.0407.1067c.3604.698.7719 1.3628 1.225 1.9932a.076.076 0 00.0842.0286c1.961-.6067 3.9495-1.5219 6.0023-3.0294a.077.077 0 00.0313-.0552c.5004-5.177-.8382-9.6739-3.5485-13.6604a.061.061 0 00-.0312-.0286zM8.02 15.3312c-1.1825 0-2.1569-1.0857-2.1569-2.419 0-1.3332.9555-2.4189 2.157-2.4189 1.2108 0 2.1757 1.0952 2.1568 2.419 0 1.3332-.9555 2.4189-2.1569 2.4189zm7.9748 0c-1.1825 0-2.1569-1.0857-2.1569-2.419 0-1.3332.9554-2.4189 2.1569-2.4189 1.2108 0 2.1757 1.0952 2.1568 2.419 0 1.3332-.946 2.4189-2.1568 2.4189z" />
          </svg>
          For bug reports please use our discord.
        </Link>
        <Link href="https://patreon.com/YOUR_PATREON" className="flex items-center gap-2 px-5 py-2.5 bg-card-hover hover:bg-line-light rounded-lg text-sm text-gray-300 transition-colors">
          <svg className="w-5 h-5 text-accent-light" viewBox="0 0 24 24" fill="currentColor">
            <path d="M14.82 2.41C18.78 2.41 22 5.65 22 9.62C22 13.58 18.78 16.8 14.82 16.8C10.85 16.8 7.61 13.58 7.61 9.62C7.61 5.65 10.85 2.41 14.82 2.41M2 21.59H5.81V2.41H2V21.59Z" />
          </svg>
          Do you like this site? Support us.
        </Link>
      </div>

      {/* Recent Updates + Latest News + Top Spenders */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <SectionHeader title="Recent Updates" />
          <Card className="max-h-[600px] overflow-y-auto">
            {loadingUpdates ? (
              Array.from({ length: 5 }).map((_, i) => <UpdateItemSkeleton key={i} />)
            ) : recentUpdates.length === 0 ? (
              <EmptyState message="No recent updates yet." />
            ) : (
              recentUpdates.map((item, i) => <UpdateItem key={i} {...item} />)
            )}
            {recentUpdates.length > 0 && (
              <Link
                href="/en/novel-list?sort=updated_at&order=desc"
                className="w-full block text-center text-sm text-accent hover:text-accent-light py-3 transition-colors"
                data-testid="load-more-updates"
              >
                View All Updates
              </Link>
            )}
          </Card>
        </div>

        <div className="space-y-6">
          <div>
            <SectionHeader title="Latest News" href="/en/news" />
            <Card className="space-y-3">
              {newsItems.length === 0 ? (
                <EmptyState message="No news yet." small />
              ) : (
                newsItems.map((item) => (
                  <Link key={item.ID} href={`/en/news/${item.ID}`} className="block text-sm text-gray-200 hover:text-accent-light transition-colors line-clamp-2">
                    {item.Title}
                  </Link>
                ))
              )}
            </Card>
          </div>

          <div>
            <SectionHeader title="Top Spenders" href="/en/leaderboard" />
            <Card className="space-y-3">
              {topSpenders.length === 0 ? (
                <EmptyState message="Leaderboard is empty." small />
              ) : (
                topSpenders.map((spender, i) => (
                  <Link key={i} href={`/en/profile/${spender.ID}`} className="flex items-center justify-between group">
                    <span className="text-sm text-gray-200 group-hover:text-accent-light transition-colors truncate">
                      <span className="text-accent-light mr-1">#{i + 1}</span>
                      {spender.Username || spender.Name}
                    </span>
                    <span className="text-xs text-gray-500 shrink-0 ml-2">{spender.Tickets?.toFixed(2)} 🎫</span>
                  </Link>
                ))
              )}
            </Card>
          </div>
        </div>
      </div>

      {/* Random Novels */}
      <section data-testid="random-novels-section">
        <SectionHeader title="Random Novels" href="/en/random-novels" />
        {loadingRandom ? (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i}>
                <Skeleton className="aspect-[3/4] rounded-lg" />
                <Skeleton className="h-3 w-3/4 mt-2" />
                <Skeleton className="h-2 w-1/2 mt-2" />
              </div>
            ))}
          </div>
        ) : randomNovels.length === 0 ? (
          <EmptyState message="No novels available." />
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4 animate-stagger">
            {randomNovels.map((novel) => (
              <NovelCard
                key={novel.ID}
                title={novel.Title}
                genre={novel.Genres?.[0]?.Slug || "action"}
                chapters={novel.Chapters}
                rating={novel.Rating?.toString()}
                image={novel.CoverURL}
                href={`/en/novel/${novel.ID}/${novel.Slug}`}
                compact
              />
            ))}
          </div>
        )}
      </section>
    </div>
  );
}

function EmptyState({ message, small }: { message: string; small?: boolean }) {
  return (
    <div className={`text-center ${small ? "py-4" : "py-10"} text-gray-500`} data-testid="empty-state">
      <svg className={`mx-auto ${small ? "w-6 h-6" : "w-10 h-10"} text-gray-700 mb-2`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
      </svg>
      <p className={`${small ? "text-xs" : "text-sm"}`}>{message}</p>
    </div>
  );
}
