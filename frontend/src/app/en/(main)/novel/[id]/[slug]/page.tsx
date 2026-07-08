"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { novels, reviews, reading, author as authorApi } from "@/lib/api";
import { stripHtml, formatViews } from "@/lib/utils";
import { Novel, Chapter } from "@/types";
import { useAuth } from "@/lib/AuthContext";
import type { ReviewResponse, RatingSummary } from "@/lib/api";
import Card from "@/components/ui/Card";

export default function NovelDetailPage() {
  const params = useParams();
  const id = params?.id as string;

  const [novel, setNovel] = useState<Novel | null>(null);
  const [chapters, setChapters] = useState<Chapter[]>([]);
  const [chapterPage, setChapterPage] = useState(1);
  const [totalChapters, setTotalChapters] = useState(0);
  const [activeTab, setActiveTab] = useState<"about" | "toc" | "reviews" | "recommendations">("about");
  const [loading, setLoading] = useState(true);
  const [reviewsData, setReviewsData] = useState<ReviewResponse[]>([]);
  const [ratingSummary, setRatingSummary] = useState<RatingSummary | null>(null);
  const [myReview, setMyReview] = useState<ReviewResponse | null>(null);
  const [chapterCount, setChapterCount] = useState(0);
  const [formRating, setFormRating] = useState(0);
  const [formHoverRating, setFormHoverRating] = useState(0);
  const [formContent, setFormContent] = useState("");
  const [formSubmitting, setFormSubmitting] = useState(false);
  const [formError, setFormError] = useState("");
  const [reviewsLoading, setReviewsLoading] = useState(false);
  const [authorNovels, setAuthorNovels] = useState<Novel[]>([]);
  const { user } = useAuth();

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    const fetchNovel = async () => {
      try {
        const res = await novels.get(id);
        setNovel(res);
        if (res.Author) {
          authorApi.novels(res.Author).then((ar) => {
            setAuthorNovels((ar.data || []).filter((n: Novel) => n.ID !== parseInt(id)));
          }).catch(() => {});
        }
      } catch {
        setNovel(null);
      }
      setLoading(false);
    };
    fetchNovel();
  }, [id, params?.slug]);

  useEffect(() => {
    if (!id) return;
    novels.chapters(id, { page: chapterPage, limit: 50 })
      .then((res) => { setChapters(res.data); setTotalChapters(res.total); })
      .catch(() => { setChapters([]); setTotalChapters(0); });
  }, [id, chapterPage]);

  useEffect(() => {
    if (!id) return;
    setReviewsLoading(true);
    Promise.all([
      reviews.list(parseInt(id)),
      user ? reading.progress(parseInt(id)) : Promise.resolve(null),
    ])
      .then(([reviewsRes, progressRes]) => {
        setReviewsData(reviewsRes.data);
        setRatingSummary(reviewsRes.rating_summary);
        if (progressRes) {
          setChapterCount(progressRes.chapter_count);
          setMyReview(progressRes.my_review);
        }
      })
      .catch(() => {})
      .finally(() => setReviewsLoading(false));
  }, [id, user]);

  if (loading || !novel) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-16 text-center">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-card-hover rounded w-1/2 mx-auto" />
          <div className="h-4 bg-card-hover rounded w-1/3 mx-auto" />
        </div>
      </div>
    );
  }

  const firstChapter = chapters.length > 0 ? chapters[0] : null;
  const pageCount = Math.ceil(totalChapters / 50) || 1;

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <nav className="text-sm text-gray-500 mb-6">
        <Link href="/en" className="hover:text-violet-400 transition-colors">Home</Link>
        <span className="mx-2">/</span>
        <Link href="/en/novel-list" className="hover:text-violet-400 transition-colors">Novels</Link>
        <span className="mx-2">/</span>
        <span className="text-gray-300">{novel.Title.slice(0, 60)}</span>
      </nav>

      {/* Hero */}
      <div className="flex flex-col sm:flex-row gap-6 mb-8">
        <div className="w-44 sm:w-52 aspect-[3/4] rounded-xl bg-card-hover border border-line-light flex-shrink-0 overflow-hidden mx-auto sm:mx-0 shadow-lg">
          {novel.CoverURL ? (
            <img src={novel.CoverURL} alt="" className="w-full h-full object-cover" />
          ) : (
            <div className="w-full h-full flex items-center justify-center text-gray-600">
              <svg className="w-16 h-16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
              </svg>
            </div>
          )}
        </div>

        <div className="flex-1 min-w-0">
          <h1 className="text-2xl sm:text-3xl font-bold text-white leading-tight">{novel.Title}</h1>
          {novel.AltTitle && (
            <p className="text-sm text-gray-500 mt-1">{novel.AltTitle}</p>
          )}

          <div className="flex flex-wrap items-center gap-x-4 gap-y-1.5 mt-3 text-sm">
            <span className={`px-2 py-0.5 rounded text-xs font-medium ${
              novel.Status === "ongoing" ? "bg-green-900/40 text-green-400 border border-green-800/30" :
              novel.Status === "completed" ? "bg-blue-900/40 text-blue-400 border border-blue-800/30" :
              novel.Status === "hiatus" ? "bg-yellow-900/40 text-yellow-400 border border-yellow-800/30" :
              "bg-red-900/40 text-red-400 border border-red-800/30"
            }`}>
              {novel.Status ? novel.Status.charAt(0).toUpperCase() + novel.Status.slice(1) : ""}
            </span>
            <span className="text-gray-400">{formatViews(novel.Views)} Views</span>
            <span className="text-gray-400">{novel.Chapters} Chapters</span>
            {novel.Rating > 0 && (
              <span className="text-yellow-400">★ {novel.Rating.toFixed(1)} ({novel.RatingCount || 0})</span>
            )}
            {novel.Readers !== undefined && <span className="text-gray-400">{novel.Readers} Readers</span>}
            {novel.Chars && <span className="text-gray-400">{novel.Chars}</span>}
          </div>

          {/* AI-Unlock Progress */}
          {novel.AIPercent && (
            <div className="mt-4 max-w-xs">
              <div className="flex items-center justify-between text-sm text-gray-400 mb-1">
                <span>AI-Unlock Progress</span>
                <span className="text-violet-400">{novel.AIPercent}/{novel.Chapters}</span>
              </div>
              <div className="w-full h-2 bg-card-hover rounded-full overflow-hidden">
                <div
                  className="h-full bg-gradient-to-r from-violet-600 to-purple-600 rounded-full transition-all"
                  style={{ width: `${Math.min(100, (parseInt(novel.AIPercent) || 0) / (novel.Chapters || 1) * 100)}%` }}
                />
              </div>
              <p className="text-xs text-gray-600 mt-0.5">All chapters unlocked</p>
            </div>
          )}

          {/* Tags */}
          <div className="flex flex-wrap gap-1.5 mt-4">
            {novel.Genres.map((g) => (
              <Link
                key={g.Slug}
                href={`/en/novel-list?genre=${g.Slug}`}
                className="text-xs px-2.5 py-1 rounded-full bg-violet-900/40 text-violet-300 border border-violet-800/30 hover:bg-violet-800/50 transition-colors"
              >
                {g.Name}
              </Link>
            ))}
          </div>

          {/* Start Reading */}
          <div className="flex flex-wrap gap-3 mt-6">
            {firstChapter && (
              <Link
                href={`/en/novel/${novel.ID}/${novel.Slug}/chapter-${firstChapter.Number}`}
                className="px-8 py-2.5 bg-violet-600 hover:bg-violet-700 text-white text-sm font-semibold rounded-lg transition-colors"
              >
                Start Reading
              </Link>
            )}
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-line mb-6">
        <div className="flex gap-6">
          {(["about", "toc", "reviews", "recommendations"] as const).map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`pb-3 text-sm font-medium transition-colors border-b-2 ${
                activeTab === tab
                  ? "text-violet-400 border-violet-500"
                  : "text-gray-500 border-transparent hover:text-gray-300"
              }`}
            >
              {tab === "about" ? "Novel Summary" : tab === "toc" ? "Table of Contents" : tab === "reviews" ? `Reviews (${novel.RatingCount || 0})` : "Recommendations"}
            </button>
          ))}
        </div>
      </div>

      {/* Tab Content */}
      {activeTab === "about" && (
        <div className="space-y-6">
          <div className="bg-card border border-line rounded-xl p-6">
            <p className="text-sm text-gray-300 leading-relaxed whitespace-pre-line">{stripHtml(novel.Description || "No description available.")}</p>
          </div>

          {authorNovels.length > 0 && (
            <div className="bg-card border border-line rounded-xl p-6">
              <h3 className="text-sm font-semibold text-white mb-4">Author&apos;s Other Novels</h3>
              <div className="space-y-2">
                {authorNovels.slice(0, 10).map((an, i) => (
                  <Link
                    key={an.ID}
                    href={`/en/novel/${an.ID}/${an.Slug}`}
                    className="flex items-center gap-3 text-sm text-gray-300 hover:text-violet-400 transition-colors py-1"
                  >
                    <span className="text-xs text-gray-600 w-5 shrink-0">{i + 1}</span>
                    <span className="truncate">{an.Title}</span>
                  </Link>
                ))}
              </div>
            </div>
          )}

          <div className="bg-card border border-line rounded-xl p-6">
            <h3 className="text-sm font-semibold text-white mb-4">Details</h3>
            <div className="grid grid-cols-2 sm:grid-cols-3 gap-4 text-sm">
              <DetailItem label="Titles" value={
                <div>
                  <p className="text-gray-200">{novel.Title}</p>
                  {novel.AltTitle && <p className="text-gray-500 text-xs mt-0.5">{novel.AltTitle}</p>}
                </div>
              } />
              <DetailItem label="Status" value={<span className="text-gray-200 capitalize">{novel.Status || "-"}</span>} />
              <DetailItem label="Date Added" value={<span className="text-gray-200">July 3, 2026</span>} />
              {novel.Author && (
                <DetailItem label="Author" value={
                  <Link href={`/en/author/${encodeURIComponent(novel.Author)}`} className="text-violet-400 hover:text-violet-300 transition-colors">
                    {novel.Author}
                  </Link>
                } />
              )}
              <DetailItem label="Requested" value={<span className="text-gray-200">{novel.RequestedBy || "-"}</span>} />
              <DetailItem label="Released" value={<span className="text-gray-200">{novel.ReleasedBy || "-"}</span>} />
              <DetailItem label="Total Chapters" value={<span className="text-gray-200">{novel.Chapters}</span>} />
            </div>
          </div>

          {novel.Genres.length > 0 && (
            <div className="bg-card border border-line rounded-xl p-6">
              <h3 className="text-sm font-semibold text-white mb-4">Genre & Tags</h3>
              <div>
                <p className="text-xs text-gray-500 mb-1.5">Genre{novel.Genres.length}</p>
                <div className="flex flex-wrap gap-1.5">
                  {novel.Genres.map((g) => (
                    <Link
                      key={g.Slug}
                      href={`/en/novel-list?genre=${g.Slug}`}
                      className="text-xs px-2.5 py-1 rounded-full bg-violet-900/40 text-violet-300 border border-violet-800/30 hover:bg-violet-800/50 transition-colors"
                    >
                      {g.Name}
                    </Link>
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {activeTab === "toc" && (
        <div className="bg-card border border-line rounded-xl p-4">
          <div className="space-y-1 max-h-[600px] overflow-y-auto">
            {chapters.length === 0 ? (
              <p className="text-sm text-gray-500 text-center py-8">No chapters available.</p>
            ) : (
              chapters.map((ch) => (
                <Link
                  key={ch.ID}
                  href={`/en/novel/${novel.ID}/${novel.Slug}/chapter-${ch.Number}`}
                  className="flex items-center justify-between px-4 py-2.5 rounded-lg hover:bg-card-hover transition-colors group"
                >
                  <div className="flex items-center gap-3 min-w-0">
                    <span className="text-sm text-gray-500 w-8 flex-shrink-0">#{ch.Number}</span>
                    <span className="text-sm text-gray-200 group-hover:text-violet-400 transition-colors line-clamp-1">
                      {ch.Title || `Chapter ${ch.Number}`}
                    </span>
                  </div>
                  <div className="flex items-center gap-2">
                    {ch.IsLocked && (
                      <span className="text-xs text-yellow-500 flex items-center gap-1">
                        <svg className="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 20 20">
                          <path fillRule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clipRule="evenodd" />
                        </svg>
                        {ch.TicketCost}
                      </span>
                    )}
                  </div>
                </Link>
              ))
            )}
          </div>

          {totalChapters > 50 && (
            <div className="flex items-center justify-center gap-2 mt-4 pt-4 border-t border-line">
              <button
                onClick={() => setChapterPage(Math.max(1, chapterPage - 1))}
                disabled={chapterPage <= 1}
                className="px-3 py-1 text-xs rounded bg-card-hover text-gray-400 hover:text-white disabled:opacity-40 transition-colors"
              >
                Previous
              </button>
              <span className="text-xs text-gray-500">
                {chapterPage} / {pageCount}
              </span>
              <button
                onClick={() => setChapterPage(Math.min(pageCount, chapterPage + 1))}
                disabled={chapterPage >= pageCount}
                className="px-3 py-1 text-xs rounded bg-card-hover text-gray-400 hover:text-white disabled:opacity-40 transition-colors"
              >
                Next
              </button>
            </div>
          )}
        </div>
      )}

      {activeTab === "reviews" && (
        <div className="space-y-6">
          {ratingSummary && ratingSummary.count > 0 && (
            <div className="bg-card border border-line rounded-xl p-6">
              <div className="flex flex-col sm:flex-row items-center gap-6">
                <div className="text-center">
                  <div className="text-5xl font-bold text-yellow-400">{ratingSummary.average.toFixed(1)}</div>
                  <div className="text-sm text-gray-500 mt-1">{ratingSummary.count} review{ratingSummary.count !== 1 ? "s" : ""}</div>
                </div>
                <div className="flex-1 w-full space-y-1.5">
                  {[5, 4, 3, 2, 1].map((star) => {
                    const pct = ratingSummary.count > 0
                      ? ((ratingSummary.distribution[star] || 0) / ratingSummary.count) * 100
                      : 0;
                    return (
                      <div key={star} className="flex items-center gap-2 text-sm">
                        <span className="text-yellow-400 w-6 text-right">{star}</span>
                        <svg className="w-3.5 h-3.5 text-yellow-400" fill="currentColor" viewBox="0 0 20 20">
                          <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                        </svg>
                        <div className="flex-1 h-2 bg-card-hover rounded-full overflow-hidden">
                          <div className="h-full bg-yellow-400 rounded-full" style={{ width: `${pct}%` }} />
                        </div>
                        <span className="text-gray-500 w-6 text-right">{ratingSummary.distribution[star] || 0}</span>
                      </div>
                    );
                  })}
                </div>
              </div>
            </div>
          )}

          {user ? (
            myReview ? (
              <div className="bg-card border border-line rounded-xl p-6">
                <p className="text-sm text-green-400 mb-2">You have reviewed this novel.</p>
                <ReviewCard review={myReview} />
              </div>
            ) : chapterCount < 5 ? (
              <div className="bg-card border border-line rounded-xl p-6 text-center">
                <p className="text-sm text-gray-500">
                  Read <strong className="text-yellow-400">{chapterCount}/5</strong> chapters to unlock the review feature.
                  <span className="block mt-1 text-xs text-gray-600">Continue reading to share your thoughts!</span>
                </p>
              </div>
            ) : (
              <div className="bg-card border border-line rounded-xl p-6">
                <h3 className="text-sm font-medium text-gray-200 mb-4">Write Your Review</h3>
                <form onSubmit={async (e) => {
                  e.preventDefault();
                  if (formRating === 0) { setFormError("Please select a rating"); return; }
                  if (formContent.trim().length < 10) { setFormError("Review must be at least 10 characters"); return; }
                  setFormSubmitting(true);
                  setFormError("");
                  try {
                    const res = await reviews.create(parseInt(id), formRating, formContent);
                    setMyReview(res.data);
                    setReviewsData((prev) => [res.data, ...prev]);
                    if (ratingSummary) {
                      const newCount = ratingSummary.count + 1;
                      const newAvg = ((ratingSummary.average * ratingSummary.count) + formRating) / newCount;
                      const newDist = { ...ratingSummary.distribution };
                      newDist[formRating] = (newDist[formRating] || 0) + 1;
                      setRatingSummary({ average: newAvg, count: newCount, distribution: newDist });
                    }
                    setFormRating(0);
                    setFormContent("");
                  } catch (err) {
                    setFormError(err instanceof Error ? err.message : "Failed to submit review");
                  } finally {
                    setFormSubmitting(false);
                  }
                }}>
                  <div className="flex items-center gap-1 mb-4">
                    {[1, 2, 3, 4, 5].map((star) => (
                      <button
                        key={star}
                        type="button"
                        onClick={() => setFormRating(star)}
                        onMouseEnter={() => setFormHoverRating(star)}
                        onMouseLeave={() => setFormHoverRating(0)}
                        className="p-0.5 transition-transform hover:scale-110"
                      >
                        <svg
                          className={`w-7 h-7 ${(formHoverRating || formRating) >= star ? "text-yellow-400" : "text-gray-600"}`}
                          fill="currentColor" viewBox="0 0 20 20">
                          <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                        </svg>
                      </button>
                    ))}
                    {formRating > 0 && (
                      <span className="text-sm text-yellow-400 ml-2">
                        {formRating === 1 ? "Poor" : formRating === 2 ? "Fair" : formRating === 3 ? "Good" : formRating === 4 ? "Very Good" : "Excellent"}
                      </span>
                    )}
                  </div>

                  <textarea
                    value={formContent}
                    onChange={(e) => setFormContent(e.target.value)}
                    placeholder="Share your thoughts about this novel (min. 10 characters)..."
                    rows={4}
                    maxLength={2000}
                    className="w-full bg-card-hover border border-line-light rounded-lg px-4 py-3 text-sm text-gray-200 outline-none focus:border-accent transition-colors resize-none"
                  />
                  <p className="text-xs text-gray-600 mt-1 text-right">{formContent.length}/2000</p>

                  {formError && <p className="text-xs text-red-400 mt-2">{formError}</p>}

                  <button
                    type="submit"
                    disabled={formSubmitting}
                    className="mt-3 px-6 py-2 bg-violet-600 hover:bg-violet-700 disabled:opacity-50 text-white text-sm font-medium rounded-lg transition-colors"
                  >
                    {formSubmitting ? "Submitting..." : "Submit Review"}
                  </button>
                </form>
              </div>
            )
          ) : (
            <GuestReviewBanner />
          )}

          <div className="space-y-3">
            {reviewsLoading ? (
              <div className="text-center text-sm text-gray-500 py-8">Loading reviews...</div>
            ) : reviewsData.length === 0 ? (
              <div className="bg-card border border-line rounded-xl p-6 text-center text-sm text-gray-500">
                No reviews yet. Be the first to review!
              </div>
            ) : (
              reviewsData.map((review) => (
                <ReviewCard key={review.id} review={review} />
              ))
            )}
          </div>
        </div>
      )}

      {activeTab === "recommendations" && (
        <div className="bg-card border border-line rounded-xl p-6 text-center text-sm text-gray-500">
          No recommendations available.
        </div>
      )}
    </div>
  );
}

function DetailItem({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div>
      <span className="text-gray-500 text-xs">{label}</span>
      <div className="mt-0.5">{value}</div>
    </div>
  );
}

function GuestReviewBanner() {
  return (
    <div className="bg-card border border-line rounded-xl p-6 text-center">
      <svg className="w-12 h-12 text-gray-600 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
      </svg>
      <p className="text-sm text-gray-300 mb-1">Want to leave a review?</p>
      <p className="text-sm text-gray-500 mb-4">
        <Link href="/en/login" className="text-violet-400 hover:text-violet-300 transition-colors">Login</Link> or <Link href="/en/register" className="text-violet-400 hover:text-violet-300 transition-colors">Register</Link> to share your thoughts
      </p>
      <div className="flex justify-center gap-3">
        <Link href="/en/login" className="px-5 py-2 bg-violet-600 hover:bg-violet-700 text-white text-sm font-medium rounded-lg transition-colors">
          Login
        </Link>
        <Link href="/en/register" className="px-5 py-2 bg-card-hover hover:bg-line-light text-gray-300 text-sm font-medium rounded-lg transition-colors border border-line-light">
          Register
        </Link>
      </div>
    </div>
  );
}

function ReviewCard({ review }: { review: ReviewResponse }) {
  return (
    <div className="bg-card border border-line rounded-xl p-4">
      <div className="flex items-start gap-3">
        <div className="w-9 h-9 rounded-full bg-accent flex items-center justify-center text-white text-sm font-bold shrink-0">
          {(review.user.display_name || review.user.username)[0].toUpperCase()}
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-sm font-medium text-gray-200">{review.user.display_name || review.user.username}</span>
            <div className="flex items-center gap-0.5">
              {[1, 2, 3, 4, 5].map((star) => (
                <svg
                  key={star}
                  className={`w-3.5 h-3.5 ${star <= review.rating ? "text-yellow-400" : "text-gray-600"}`}
                  fill="currentColor" viewBox="0 0 20 20">
                  <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                </svg>
              ))}
            </div>
          </div>
          <p className="text-sm text-gray-300 mt-2 leading-relaxed">{review.content}</p>
          <p className="text-xs text-gray-600 mt-2">{timeAgo(review.created_at)}</p>
        </div>
      </div>
    </div>
  );
}

function timeAgo(dateStr: string) {
  const now = Date.now();
  const date = new Date(dateStr).getTime();
  const diff = now - date;
  const minutes = Math.floor(diff / 60000);
  if (minutes < 1) return "just now";
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}d ago`;
  const months = Math.floor(days / 30);
  return `${months}mo ago`;
}