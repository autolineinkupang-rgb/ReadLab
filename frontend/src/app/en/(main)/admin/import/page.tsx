"use client";

import { useState, useEffect, useRef } from "react";
import { importer, adminNovels, scraper, lncrawl } from "@/lib/api";
import Card from "@/components/ui/Card";
import ImportProgress, { ProgressStep } from "@/components/admin/ImportProgress";

export default function AdminImportPage() {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<{ id: string; title: string; url: string; image: string }[]>([]);
  const [searching, setSearching] = useState(false);
  const [importing, setImporting] = useState<string | null>(null);
  const [importResult, setImportResult] = useState<{ id: number; title: string } | null>(null);
  const [error, setError] = useState("");
  const [withChapters, setWithChapters] = useState(false);
  const [showSearch, setShowSearch] = useState(false);
  const [scrapeURL, setScrapeURL] = useState("");
  const [scraping, setScraping] = useState(false);
  const [scrapeResult, setScrapeResult] = useState<any>(null);
  const [scrapeWithContent, setScrapeWithContent] = useState(true);
  const [scrapeChapterRange, setScrapeChapterRange] = useState("");
  const [scrapeImporting, setScrapeImporting] = useState(false);
  const [lncrawlURL, setLncrawlURL] = useState("");
  const [lncrawlMax, setLncrawlMax] = useState(0);
  const [lncrawlChapterRange, setLncrawlChapterRange] = useState("");
  const [lncrawling, setLncrawling] = useState(false);
  const [lncrawlStatus, setLncrawlStatus] = useState("");

  const [scrapeSteps, setScrapeSteps] = useState<ProgressStep[]>([]);
  const [scrapeImportSteps, setScrapeImportSteps] = useState<ProgressStep[]>([]);
  const [lncrawlSteps, setLncrawlSteps] = useState<ProgressStep[]>([]);
  const scrapeTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const importTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const lncrawlTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  function startStepSimulation(
    labels: string[],
    setter: (steps: ProgressStep[]) => void,
    timerRef: React.MutableRefObject<ReturnType<typeof setInterval> | null>,
    intervalMs = 2000,
  ) {
    stopStepSimulation(timerRef);
    setter(labels.map((l, i) => ({ label: l, status: i === 0 ? "active" : "pending" as const })));
    let idx = 0;
    timerRef.current = setInterval(() => {
      idx++;
      if (idx >= labels.length) {
        stopStepSimulation(timerRef);
        return;
      }
      setter(labels.map((l, i) => ({
        label: l,
        status: i < idx ? "done" as const : i === idx ? "active" as const : "pending" as const,
      })));
    }, intervalMs);
  }

  function completeStepSimulation(
    labels: string[],
    setter: (steps: ProgressStep[]) => void,
    timerRef: React.MutableRefObject<ReturnType<typeof setInterval> | null>,
  ) {
    stopStepSimulation(timerRef);
    setter(labels.map((l) => ({ label: l, status: "done" as const })));
  }

  function stopStepSimulation(timerRef: React.MutableRefObject<ReturnType<typeof setInterval> | null>) {
    if (timerRef.current) { clearInterval(timerRef.current); timerRef.current = null; }
  }

  useEffect(() => {
    return () => {
      stopStepSimulation(scrapeTimerRef);
      stopStepSimulation(importTimerRef);
      stopStepSimulation(lncrawlTimerRef);
    };
  }, []);
  const [manualForm, setManualForm] = useState({
    title: "", alt_title: "", author: "", status: "ongoing",
    description: "", cover_url: "", source_url: "",
    chars: "", ai_percent: "", rating: 0,
    chapters: 0,
    chapterContent: "",
  });

  async function handleSearch() {
    if (!query.trim()) return;
    setSearching(true);
    setError("");
    setImportResult(null);
    try {
      const res = await importer.search(query.trim());
      setResults(res.data || []);
    } catch {
      setError("Search failed. Consumet API may be unavailable.");
      setResults([]);
    } finally {
      setSearching(false);
    }
  }

  async function handleLncrawl() {
    if (!lncrawlURL.trim()) return;
    setLncrawling(true);
    setError("");
    setImportResult(null);
    setLncrawlStatus("");
    const rangeLabel = lncrawlChapterRange.trim() ? ` (${lncrawlChapterRange.trim()})` : "";
    const lncrawlLabels = [
      "Connecting to source...",
      "Fetching novel info...",
      `Crawling chapters${rangeLabel}...`,
      "Downloading content...",
      "Importing to database...",
    ];
    startStepSimulation(lncrawlLabels, setLncrawlSteps, lncrawlTimerRef, 3000);
    try {
      const res = await lncrawl.crawl(lncrawlURL.trim(), lncrawlMax || undefined, lncrawlChapterRange.trim() || undefined);
      setImportResult({ id: res.data.ID, title: res.data.Title });
      setLncrawlStatus(`Imported novel with ${res.data.Chapters?.length || 0} chapters`);
      completeStepSimulation(lncrawlLabels, setLncrawlSteps, lncrawlTimerRef);
    } catch (e: any) {
      setError(e.message || "lncrawl failed. See server logs.");
      stopStepSimulation(lncrawlTimerRef);
      setLncrawlSteps([]);
    } finally {
      setLncrawling(false);
    }
  }

  async function handleManualImport() {
    if (!manualForm.title.trim()) return;
    setError("");
    setImportResult(null);
    try {
      const chapters = [];
      const count = Math.min(manualForm.chapters || 0, 1000);
      for (let i = 1; i <= count; i++) {
        chapters.push({
          number: i,
          title: `Chapter ${i}`,
          content: manualForm.chapterContent || `This is the content of chapter ${i}.`,
        });
      }
      const res = await adminNovels.create({
        title: manualForm.title,
        alt_title: manualForm.alt_title,
        author: manualForm.author,
        status: manualForm.status,
        description: manualForm.description,
        cover_url: manualForm.cover_url,
        source_url: manualForm.source_url,
        chars: manualForm.chars,
        ai_percent: manualForm.ai_percent,
        rating: manualForm.rating,
        chapters,
      });
      setImportResult({ id: res.data.ID, title: res.data.Title });
    } catch (e: any) {
      setError(e.message || "Create failed.");
    }
  }

  async function handleScrape() {
    if (!scrapeURL.trim()) return;
    setScraping(true);
    setError("");
    setScrapeResult(null);
    startStepSimulation(
      ["Fetching page metadata...", "Extracting novel info...", "Parsing chapter list..."],
      setScrapeSteps, scrapeTimerRef, 1500,
    );
    try {
      const res = await scraper.scrape(scrapeURL.trim());
      setScrapeResult(res.data);
      completeStepSimulation(
        ["Fetching page metadata...", "Extracting novel info...", "Parsing chapter list..."],
        setScrapeSteps, scrapeTimerRef,
      );
    } catch (e: any) {
      setError(e.message || "Scrape failed.");
      stopStepSimulation(scrapeTimerRef);
      setScrapeSteps([]);
    } finally {
      setScraping(false);
    }
  }

  async function handleScrapeImport() {
    if (!scrapeURL.trim() || !scrapeResult) return;
    setScrapeImporting(true);
    setError("");
    const rangeLabel = scrapeChapterRange.trim() ? ` (${scrapeChapterRange.trim()})` : "";
    const importLabels = scrapeWithContent
      ? [`Fetching chapters${rangeLabel}...`, "Downloading chapter content...", "Saving to database..."]
      : [`Fetching chapters${rangeLabel}...`, "Saving to database..."];
    startStepSimulation(importLabels, setScrapeImportSteps, importTimerRef, 2500);
    try {
      const res = await scraper.import(scrapeURL.trim(), scrapeWithContent, scrapeChapterRange.trim() || undefined);
      setImportResult({ id: res.data.ID, title: res.data.Title });
      setScrapeResult(null);
      completeStepSimulation(importLabels, setScrapeImportSteps, importTimerRef);
    } catch (e: any) {
      setError(e.message || "Scrape import failed.");
      stopStepSimulation(importTimerRef);
      setScrapeImportSteps([]);
    } finally {
      setScrapeImporting(false);
    }
  }

  async function handleImport(sourceID: string) {
    setImporting(sourceID);
    setError("");
    setImportResult(null);
    try {
      const res = await importer.import(sourceID, withChapters);
      setImportResult({ id: res.data.ID, title: res.data.Title });
    } catch (e: any) {
      setError(e.message || "Import failed.");
    } finally {
      setImporting(null);
    }
  }

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold text-white mb-2">Import Novel</h1>

      {/* API Unavailable Notice */}
      <div className="mb-6 p-4 rounded-lg bg-amber-900/20 border border-amber-800/40 text-sm text-amber-300">
        <p className="font-semibold mb-1">Public Import API Unavailable</p>
        <p>
          The free Consumet API (used for NovelUpdates search &amp; import) is no longer available (HTTP 451).
          Even when functional, it does not provide chapter content — only titles.
        </p>
        <p className="mt-2">
          Use the <strong>Manual Entry</strong> form below to add novels with full details and chapter content.
        </p>
      </div>

      {error && (
        <div className="mb-4 p-3 rounded-lg bg-red-900/30 border border-red-900/50 text-sm text-red-400">
          {error}
        </div>
      )}

      {importResult && (
        <div className="mb-4 p-3 rounded-lg bg-green-900/30 border border-green-900/50 text-sm text-green-400">
          Successfully imported: <a href={`/en/novel/${importResult.id}/${importResult.title.toLowerCase().replace(/\s+/g, "-")}`} className="underline">{importResult.title}</a>
        </div>
      )}

      {/* Manual Novel Entry (primary, shown by default) */}
      <Card className="mb-6 p-4 space-y-3">
        <h2 className="text-white text-sm font-semibold">Manual Novel Entry</h2>
          <input
            value={manualForm.title}
            onChange={(e) => setManualForm((p) => ({ ...p, title: e.target.value }))}
            className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
            placeholder="Title *"
          />
          <input
            value={manualForm.source_url}
            onChange={(e) => setManualForm((p) => ({ ...p, source_url: e.target.value }))}
            className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
            placeholder="Source URL (NovelUpdates / RoyalRoad / etc.)"
          />
          <input
            value={manualForm.alt_title}
            onChange={(e) => setManualForm((p) => ({ ...p, alt_title: e.target.value }))}
            className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
            placeholder="Alt Title"
          />
          <input
            value={manualForm.author}
            onChange={(e) => setManualForm((p) => ({ ...p, author: e.target.value }))}
            className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
            placeholder="Author"
          />
          <textarea
            value={manualForm.description}
            onChange={(e) => setManualForm((p) => ({ ...p, description: e.target.value }))}
            className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none resize-y min-h-[80px]"
            placeholder="Description"
          />
          <input
            value={manualForm.cover_url}
            onChange={(e) => setManualForm((p) => ({ ...p, cover_url: e.target.value }))}
            className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
            placeholder="Cover Image URL"
          />
          <div className="grid grid-cols-2 gap-3">
            <input
              value={manualForm.chars}
              onChange={(e) => setManualForm((p) => ({ ...p, chars: e.target.value }))}
              className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
              placeholder="Chars (e.g. 1.2M)"
            />
            <input
              value={manualForm.ai_percent}
              onChange={(e) => setManualForm((p) => ({ ...p, ai_percent: e.target.value }))}
              className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
              placeholder="AI %"
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <input
              type="number"
              step="0.1"
              min="0"
              max="5"
              value={manualForm.rating}
              onChange={(e) => setManualForm((p) => ({ ...p, rating: parseFloat(e.target.value) || 0 }))}
              className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
              placeholder="Rating (0-5)"
            />
            <input
              type="number"
              min="0"
              max="10000"
              value={manualForm.chapters}
              onChange={(e) => setManualForm((p) => ({ ...p, chapters: parseInt(e.target.value) || 0 }))}
              className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
              placeholder="# Chapters"
            />
          </div>
          {(manualForm.chapters || 0) > 0 && (
            <textarea
              value={manualForm.chapterContent}
              onChange={(e) => setManualForm((p) => ({ ...p, chapterContent: e.target.value }))}
              className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none resize-y min-h-[100px]"
              placeholder="Chapter content template (used for all chapters)"
            />
          )}
          <select
            value={manualForm.status}
            onChange={(e) => setManualForm((p) => ({ ...p, status: e.target.value }))}
            className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
          >
            <option value="ongoing">Ongoing</option>
            <option value="completed">Completed</option>
            <option value="hiatus">Hiatus</option>
            <option value="dropped">Dropped</option>
          </select>
          <button
            onClick={handleManualImport}
            disabled={!manualForm.title.trim()}
            className="w-full px-4 py-2 bg-green-600 hover:bg-green-700 disabled:opacity-50 text-white text-sm font-medium rounded-lg transition-colors"
          >
            Create Novel
          </button>
        </Card>

        {/* Web Scraper */}
        <Card className="mb-6 p-4 space-y-3">
          <h2 className="text-white text-sm font-semibold">Web Scraper</h2>
          <p className="text-xs text-gray-500">
            Paste a novel URL from supported sites (NovelBin, FreeWebNovel) to scrape metadata and chapters.
          </p>
          <div className="flex gap-2">
            <input
              value={scrapeURL}
              onChange={(e) => setScrapeURL(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleScrape()}
              placeholder="https://novelbin.com/novel-name"
              className="flex-1 bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
            />
            <button
              onClick={handleScrape}
              disabled={scraping || !scrapeURL.trim()}
              className="shrink-0 px-4 py-2 bg-accent hover:bg-accent-dark disabled:opacity-50 text-white text-sm rounded-lg transition-colors"
            >
              {scraping ? "Scraping..." : "Scrape"}
            </button>
          </div>

          {scraping && scrapeSteps.length > 0 && (
            <ImportProgress steps={scrapeSteps} />
          )}

          <input
            value={scrapeChapterRange}
            onChange={(e) => setScrapeChapterRange(e.target.value)}
            placeholder="Chapter range (e.g. 1-10, 11-30 — leave empty for all)"
            className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
          />

          {scrapeResult && (
            <div className="space-y-2 pt-2 border-t border-line-light">
              <div className="flex gap-3">
                {scrapeResult.CoverURL && (
                  <img src={scrapeResult.CoverURL} alt="" className="w-12 h-16 rounded object-cover shrink-0 bg-card-hover" />
                )}
                <div className="min-w-0">
                  <h3 className="text-sm font-semibold text-white">{scrapeResult.Title}</h3>
                  <p className="text-xs text-gray-500">by {scrapeResult.Author || "?"} &middot; {scrapeResult.Status || "?"}</p>
                  <p className="text-xs text-gray-600">{scrapeResult.Chapters?.length || 0} chapters found</p>
                  {scrapeResult.Genres?.length > 0 && (
                    <div className="flex flex-wrap gap-1 mt-1">
                      {scrapeResult.Genres.map((g: string) => (
                        <span key={g} className="text-[9px] px-1.5 py-0.5 rounded-full bg-accent/10 text-accent-light/80">{g}</span>
                      ))}
                    </div>
                  )}
                </div>
              </div>
              <label className="flex items-center gap-2 text-sm text-gray-400">
                <input
                  type="checkbox"
                  checked={scrapeWithContent}
                  onChange={(e) => setScrapeWithContent(e.target.checked)}
                  className="h-4 w-4 rounded border-line-light accent-accent"
                />
                Also scrape chapter content (slower)
              </label>
              <div className="flex gap-2">
                <button
                  onClick={handleScrapeImport}
                  disabled={scrapeImporting}
                  className="px-4 py-2 bg-green-600 hover:bg-green-700 disabled:opacity-50 text-white text-sm rounded-lg transition-colors"
                >
                  {scrapeImporting ? "Importing..." : "Import to Database"}
                </button>
                <button
                  onClick={() => { setScrapeResult(null); stopStepSimulation(scrapeTimerRef); setScrapeSteps([]); }}
                  className="px-4 py-2 bg-card-hover text-gray-300 text-sm rounded-lg transition-colors"
                >
                  Cancel
                </button>
              </div>
              {scrapeImporting && scrapeImportSteps.length > 0 && (
                <ImportProgress steps={scrapeImportSteps} />
              )}
              {scrapeResult.Chapters?.length > 0 && (
                <div className="max-h-32 overflow-y-auto">
                  <p className="text-xs text-gray-500 mb-1">Chapters preview:</p>
                  {scrapeResult.Chapters.slice(0, 10).map((ch: any) => (
                    <p key={ch.Number} className="text-xs text-gray-600 truncate">{ch.Number}. {ch.Title}</p>
                  ))}
                  {scrapeResult.Chapters.length > 10 && (
                    <p className="text-xs text-gray-600">...and {scrapeResult.Chapters.length - 10} more</p>
                  )}
                </div>
              )}
            </div>
          )}
        </Card>

        {/* Lncrawl (lightnovel-crawler) */}
        <Card className="mb-6 p-4 space-y-3">
          <h2 className="text-white text-sm font-semibold">Light Novel Crawler (lncrawl)</h2>
          <p className="text-xs text-gray-500">
            Paste a novel URL from a supported site (novelfire.net, royalroad.com, etc.) to automatically crawl and import all chapters.
          </p>
          {lncrawlStatus && (
            <p className="text-xs text-green-400">{lncrawlStatus}</p>
          )}
          <input
            value={lncrawlChapterRange}
            onChange={(e) => setLncrawlChapterRange(e.target.value)}
            placeholder="Chapter range (e.g. 1-10, 11-30 — leave empty for all)"
            className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
          />
          <div className="flex gap-2 flex-col sm:flex-row">
            <input
              value={lncrawlURL}
              onChange={(e) => setLncrawlURL(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleLncrawl()}
              placeholder="https://novelfire.net/book/sand-mage-of-the-burnt-desert"
              className="flex-1 bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
            />
            <div className="flex gap-2 shrink-0">
              <input
                type="number"
                min="0"
                max="9999"
                value={lncrawlMax}
                onChange={(e) => setLncrawlMax(parseInt(e.target.value) || 0)}
                className="w-20 bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
                placeholder="Max"
                title="Max chapters to crawl (0 = all)"
              />
              <button
                onClick={handleLncrawl}
                disabled={lncrawling || !lncrawlURL.trim()}
                className="shrink-0 px-4 py-2 bg-accent hover:bg-accent-dark disabled:opacity-50 text-white text-sm rounded-lg transition-colors"
              >
                {lncrawling ? "Crawling..." : "Crawl & Import"}
              </button>
            </div>
          </div>
          {lncrawling && lncrawlSteps.length > 0 && (
            <ImportProgress steps={lncrawlSteps} />
          )}
        </Card>

        {/* Public API Search (collapsed by default - API is dead) */}
        <button
          onClick={() => setShowSearch(!showSearch)}
          className="text-sm text-accent hover:text-accent-light transition-colors mb-4 block"
        >
          {showSearch ? "− Hide NovelUpdates search" : "+ NovelUpdates search (API unavailable)"}
        </button>

        {showSearch && (
          <>
            <div className="flex gap-3 mb-6">
              <div className="flex-1 relative">
                <input
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleSearch()}
                  placeholder="Search novel name..."
                  className="w-full bg-card-hover border border-line-light rounded-xl pl-4 pr-12 py-3 text-sm text-gray-200 outline-none focus:border-accent"
                />
                <button
                  onClick={handleSearch}
                  disabled={searching}
                  className="absolute right-2 top-1/2 -translate-y-1/2 px-4 py-1.5 bg-accent hover:bg-accent-dark disabled:opacity-50 text-white text-sm rounded-lg transition-colors"
                >
                  {searching ? "Searching..." : "Search"}
                </button>
              </div>
            </div>
            <label className="flex items-center gap-2 text-sm text-gray-400 mb-6">
              <input
                type="checkbox"
                checked={withChapters}
                onChange={(e) => setWithChapters(e.target.checked)}
                className="h-4 w-4 rounded border-line-light accent-accent"
              />
              Also import chapter list
            </label>
            {searching && <p className="text-sm text-accent mb-4">Searching...</p>}
            {results.length > 0 && (
              <div className="space-y-3">
                {results.map((r) => (
                  <Card key={r.id} className="flex flex-col sm:flex-row items-start sm:items-center gap-3 sm:gap-4">
                    <div className="flex items-center gap-3 w-full sm:w-auto">
                      {r.image ? (
                        <img src={r.image} alt="" className="w-10 h-14 sm:w-12 sm:h-16 rounded-lg object-cover bg-card-hover shrink-0" />
                      ) : (
                        <div className="w-10 h-14 sm:w-12 sm:h-16 rounded-lg bg-card-hover flex items-center justify-center shrink-0">
                          <svg className="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
                          </svg>
                        </div>
                      )}
                      <div className="flex-1 min-w-0 sm:hidden">
                        <h3 className="text-sm font-semibold text-white line-clamp-2">{r.title}</h3>
                        <p className="text-xs text-gray-500 truncate mt-0.5">{r.id}</p>
                      </div>
                    </div>
                    <div className="hidden sm:block flex-1 min-w-0">
                      <h3 className="text-sm font-semibold text-white truncate">{r.title}</h3>
                      <p className="text-xs text-gray-500 truncate mt-0.5">{r.id}</p>
                    </div>
                    <button
                      onClick={() => handleImport(r.id)}
                      disabled={importing === r.id}
                      className="w-full sm:w-auto shrink-0 px-4 py-2 bg-accent hover:bg-accent-dark disabled:opacity-50 text-white text-sm rounded-lg transition-colors"
                    >
                      {importing === r.id ? "Importing..." : "Import"}
                    </button>
                  </Card>
                ))}
              </div>
            )}
            {!searching && results.length === 0 && (
              <p className="text-center text-gray-500 text-sm py-8">Search for novels from NovelUpdates to import.</p>
            )}
          </>
        )}
    </div>
  );
}
