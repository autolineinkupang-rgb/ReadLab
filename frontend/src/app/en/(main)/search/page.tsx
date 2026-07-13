"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { search, genres as genresApi } from "@/lib/api";
import Pagination from "@/components/ui/Pagination";

interface AutocompleteItem {
  id: number;
  slug: string;
  title: string;
}

interface SearchResult {
  ID: number;
  Slug: string;
  Title: string;
  CoverURL?: string;
  Rating?: number;
  Chapters?: number;
  Genres?: { Slug: string }[];
  Description?: string;
  Views?: number;
}

const GENRE_ICONS: Record<string, string> = {
  action: "⚔️",
  adventure: "🗺️",
  comedy: "😂",
  drama: "🎭",
  fantasy: "🧙",
  horror: "👻",
  mystery: "🔍",
  romance: "❤️",
  scifi: "🚀",
  slice_of_life: "🌸",
  thriller: "🔫",
  "isekai": "🌀",
  martial_arts: "🥋",
  supernatural: "👁️",
  psychological: "🧠",
};

export default function SearchPage() {
  /* ---- state ---- */
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<SearchResult[]>([]);
  const [suggestions, setSuggestions] = useState<AutocompleteItem[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [allGenres, setAllGenres] = useState<{ Slug: string }[]>([]);
  const [selectedGenre, setSelectedGenre] = useState("");
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(0);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [searched, setSearched] = useState(false);

  const inputRef = useRef<HTMLInputElement>(null);
  const suggestionsRef = useRef<HTMLDivElement>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  /* ---- fetch genres ---- */
  useEffect(() => {
    genresApi
      .list()
      .then((res) => {
        if (res.data?.length) setAllGenres(res.data);
      })
      .catch(() => {});
  }, []);

  /* ---- debounce autocomplete ---- */
  const fetchSuggestions = useCallback(async (q: string) => {
    if (q.trim().length < 2) {
      setSuggestions([]);
      return;
    }
    try {
      const res = await search.autocomplete(q);
      setSuggestions(res.data?.slice(0, 8) ?? []);
    } catch {
      setSuggestions([]);
    }
  }, []);

  const handleQueryChange = useCallback(
    (value: string) => {
      setQuery(value);
      setPage(1);

      // Debounce autocomplete
      if (debounceRef.current) clearTimeout(debounceRef.current);
      debounceRef.current = setTimeout(() => {
        fetchSuggestions(value);
        setShowSuggestions(true);
      }, 300);
    },
    [fetchSuggestions],
  );

  /* ---- close suggestions on outside click ---- */
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (
        suggestionsRef.current &&
        !suggestionsRef.current.contains(e.target as Node) &&
        inputRef.current &&
        !inputRef.current.contains(e.target as Node)
      ) {
        setShowSuggestions(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  /* ---- perform search ---- */
  const doSearch = useCallback(
    async (searchQuery?: string, searchPage = 1, searchGenre?: string) => {
      const q = searchQuery ?? query;
      const genre = searchGenre ?? selectedGenre;
      if (!q.trim() && !genre) return;

      setLoading(true);
      setSearched(true);
      setShowSuggestions(false);

      try {
        const res = await search.query(q, {
          page: searchPage,
          limit: 18,
        });
        setResults(res.data ?? []);
        setPage(res.page ?? 1);
        const limit = 18;
        setTotal(Math.ceil((res.total ?? 0) / limit));
        setTotalPages(Math.ceil((res.total ?? 0) / limit));
      } catch {
        setResults([]);
        setTotalPages(0);
        setTotal(0);
      } finally {
        setLoading(false);
      }
    },
    [query, selectedGenre],
  );

  /* ---- keyboard support ---- */
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      doSearch();
    } else if (e.key === "Escape") {
      setShowSuggestions(false);
    }
  };

  /* ---- select suggestion ---- */
  const selectSuggestion = (item: AutocompleteItem) => {
    setQuery(item.title);
    setShowSuggestions(false);
    // Navigate directly to the novel
    window.location.href = `/en/novel/${item.id}/${item.slug}`;
  };

  /* ---- genre filter change triggers re-search ---- */
  useEffect(() => {
    if (searched) {
      doSearch(query, 1, selectedGenre);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedGenre]);

  /* ---- page change ---- */
  const handlePageChange = (p: number) => {
    doSearch(query, p, selectedGenre);
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      {/* ---- Search Header ---- */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-white mb-6">Search Novels</h1>

        {/* Large search input */}
        <div className="relative">
          <div className="relative">
            <svg
              className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-500"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
              />
            </svg>
            <input
              ref={inputRef}
              type="text"
              value={query}
              onChange={(e) => handleQueryChange(e.target.value)}
              onKeyDown={handleKeyDown}
              onFocus={() => {
                if (suggestions.length > 0) setShowSuggestions(true);
              }}
              placeholder="Search by title, author, or keyword…"
              className="w-full pl-12 pr-28 py-4 bg-card border border-line rounded-xl text-gray-200 placeholder-gray-600 text-base focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent transition-colors"
              autoComplete="off"
            />
            <button
              onClick={() => doSearch()}
              className="absolute right-2 top-1/2 -translate-y-1/2 px-5 py-2 bg-accent hover:bg-accent-dark text-white text-sm font-medium rounded-lg transition-colors"
            >
              Search
            </button>
          </div>

          {/* Autocomplete dropdown */}
          {showSuggestions && suggestions.length > 0 && (
            <div
              ref={suggestionsRef}
              className="absolute top-full left-0 right-0 mt-2 bg-card border border-line rounded-xl shadow-xl shadow-black/40 overflow-hidden z-50 animate-fade-in"
            >
              {suggestions.map((item) => (
                <button
                  key={item.id}
                  onClick={() => selectSuggestion(item)}
                  className="w-full text-left px-4 py-3 text-sm text-gray-200 hover:bg-card-hover transition-colors flex items-center gap-3"
                >
                  <svg
                    className="w-4 h-4 text-gray-600 shrink-0"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                    />
                  </svg>
                  <span className="truncate">{item.title}</span>
                  <svg
                    className="w-3.5 h-3.5 text-gray-600 shrink-0 ml-auto"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M9 5l7 7-7 7"
                    />
                  </svg>
                </button>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* ---- Main content: results + sidebar ---- */}
      <div className="flex flex-col lg:flex-row gap-8">
        {/* Genre sidebar */}
        <aside className="lg:w-56 shrink-0">
          <h2 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-3">
            Genres
          </h2>
          <div className="flex flex-row flex-wrap lg:flex-col gap-1.5">
            <button
              onClick={() => setSelectedGenre("")}
              className={`text-left px-3 py-2 rounded-lg text-sm transition-colors ${
                selectedGenre === ""
                  ? "bg-accent/20 text-accent-light font-medium"
                  : "text-gray-400 hover:text-gray-200 hover:bg-card-hover"
              }`}
            >
              All Genres
            </button>
            {allGenres.map((g) => (
              <button
                key={g.Slug}
                onClick={() => setSelectedGenre(g.Slug)}
                className={`text-left px-3 py-2 rounded-lg text-sm transition-colors ${
                  selectedGenre === g.Slug
                    ? "bg-accent/20 text-accent-light font-medium"
                    : "text-gray-400 hover:text-gray-200 hover:bg-card-hover"
                }`}
              >
                <span className="mr-1.5">{GENRE_ICONS[g.Slug] ?? "📖"}</span>
                <span className="capitalize">{g.Slug.replace(/_/g, " ")}</span>
              </button>
            ))}
          </div>
        </aside>

        {/* Results area */}
        <div className="flex-1 min-w-0">
          {/* Status bar */}
          {searched && !loading && (
            <p className="text-sm text-gray-500 mb-4">
              {results.length > 0
                ? `Found ${total} result${total !== 1 ? "s" : ""}`
                : "No results found. Try different keywords."}
            </p>
          )}

          {/* Loading skeleton */}
          {loading && (
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 xl:grid-cols-5 gap-4 animate-pulse">
              {Array.from({ length: 10 }).map((_, i) => (
                <div key={i}>
                  <div className="aspect-[3/4] rounded-lg bg-card-hover border border-line" />
                  <div className="h-3 w-3/4 mt-2 rounded bg-card-hover" />
                  <div className="h-2 w-1/2 mt-1.5 rounded bg-card-hover" />
                </div>
              ))}
            </div>
          )}

          {/* Empty state (no search yet) */}
          {!searched && !loading && (
            <div className="flex flex-col items-center justify-center py-20 text-center">
              <svg
                className="w-16 h-16 text-gray-700 mb-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={1}
                  d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                />
              </svg>
              <p className="text-gray-500 text-sm">
                Type a query above and press Enter or click Search to find
                novels.
              </p>
            </div>
          )}

          {/* No results */}
          {searched && !loading && results.length === 0 && (
            <div className="flex flex-col items-center justify-center py-20 text-center">
              <svg
                className="w-16 h-16 text-gray-700 mb-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={1}
                  d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              <p className="text-gray-500 text-sm">
                No novels matched your search. Try different keywords or
                remove filters.
              </p>
            </div>
          )}

          {/* Results grid */}
          {searched && !loading && results.length > 0 && (
            <>
              <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 xl:grid-cols-5 gap-4 animate-stagger">
                {results.map((novel) => (
                  <Link
                    key={novel.ID}
                    href={`/en/novel/${novel.ID}/${novel.Slug}`}
                    className="group block focus:outline-none focus-visible:ring-2 focus-visible:ring-accent rounded-lg"
                  >
                    <div className="relative aspect-[3/4] rounded-lg overflow-hidden bg-card-hover border border-line card-hover transition-all duration-300">
                      {novel.CoverURL && /^(https?:\/\/|\/api\/)/.test(novel.CoverURL) ? (
                        <img
                          src={novel.CoverURL}
                          alt={novel.Title}
                          loading="lazy"
                          className="absolute inset-0 w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
                          onError={(e) => {
                            (e.currentTarget as HTMLImageElement).style.display =
                              "none";
                          }}
                        />
                      ) : (
                        <div className="absolute inset-0 flex items-center justify-center text-gray-600 bg-gradient-to-br from-card-hover to-line-light">
                          <svg
                            className="w-12 h-12"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                          >
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              strokeWidth={1}
                              d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"
                            />
                          </svg>
                        </div>
                      )}
                      {novel.Rating != null && (
                        <span className="absolute top-2 right-2 bg-black/70 backdrop-blur-sm text-yellow-400 text-xs px-1.5 py-0.5 rounded flex items-center gap-0.5 z-10">
                          <svg
                            className="w-3 h-3"
                            fill="currentColor"
                            viewBox="0 0 20 20"
                          >
                            <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                          </svg>
                          {typeof novel.Rating === "number"
                            ? novel.Rating.toFixed(1)
                            : novel.Rating}
                        </span>
                      )}
                    </div>
                    <div className="mt-2 space-y-0.5">
                      <h3 className="text-sm font-medium text-gray-200 group-hover:text-accent-light transition-colors line-clamp-2 leading-tight">
                        {novel.Title}
                      </h3>
                      <div className="flex items-center gap-2 text-xs text-gray-500">
                        {novel.Genres?.[0]?.Slug && (
                          <span className="capitalize">
                            {novel.Genres[0].Slug.replace(/_/g, " ")}
                          </span>
                        )}
                        {novel.Genres?.[0]?.Slug && novel.Chapters != null && (
                          <span>•</span>
                        )}
                        {novel.Chapters != null && <span>{novel.Chapters} Ch</span>}
                      </div>
                    </div>
                  </Link>
                ))}
              </div>

              {/* Pagination */}
              <Pagination
                page={page}
                totalPages={totalPages}
                onPageChange={handlePageChange}
              />
            </>
          )}
        </div>
      </div>
    </div>
  );
}