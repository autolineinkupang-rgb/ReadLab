"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { novels, genres as genresApi } from "@/lib/api";

interface Novel {
  ID: number;
  Title: string;
  AltTitle: string;
  Slug: string;
  Author: string;
  Status: string;
  Views: number;
  Rating: number;
  Chapters: number;
  Readers: number;
  Chars: string;
  AIPercent: string;
  Description: string;
  CoverURL: string;
  Genres: { ID: number; Slug: string; Name: string }[];
}

const genreList = [
  "action","adult","adventure","comedy","drama","ecchi","erciyuan","fan-fiction","fantasy",
  "game","gender-bender","harem","historical","horror","josei","martial-arts","mature",
  "mecha","military","mystery","psychological","romance","school-life","sci-fi","seinen",
  "shoujo","shoujo-ai","shounen","shounen-ai","slice-of-life","smut","sports","supernatural",
  "tragedy","urban-life","wuxia","xianxia","xuanhuan","yaoi","yuri",
];

const mockNovels: Novel[] = [
  { ID: 1, Title: "Having Dinner with His Brother, the Cold and Aloof Tycoon Becomes Addicted to His Doting Affections", AltTitle: "陪哥哥吃饭，冷欲大佬强宠上瘾", Slug: "having-dinner-with-his-brother", Author: "半条活鱼", Status: "completed", Views: 3142, Rating: 3.5, Chapters: 135, Readers: 17, Chars: "250K", AIPercent: "37%", Description: "Cold and aloof tycoon × Bright and delicate princess.", CoverURL: "", Genres: [{ ID: 22, Slug: "romance", Name: "Romance" }, { ID: 30, Slug: "slice-of-life", Name: "Slice of Life" }, { ID: 35, Slug: "urban-life", Name: "Urban Life" }] },
  { ID: 2, Title: "Corpse Puppet Phoenix Girl", AltTitle: "尸傀凰女", Slug: "corpse-puppet-phoenix-girl", Author: "佚名", Status: "ongoing", Views: 2105, Rating: 3.8, Chapters: 242, Readers: 22, Chars: "653K", AIPercent: "20.7%", Description: "Meeting you at the most beautiful street corner.", CoverURL: "", Genres: [{ ID: 2, Slug: "adult", Name: "Adult" }, { ID: 4, Slug: "adventure", Name: "Adventure" }, { ID: 12, Slug: "fantasy", Name: "Fantasy" }, { ID: 22, Slug: "romance", Name: "Romance" }] },
  { ID: 3, Title: "Reborn As the Little Delicate Wife of the Domineering Ceo", AltTitle: "豪门重生，夫人超超超厉害", Slug: "reborn-as-the-little-delicate-wife", Author: "佚名", Status: "completed", Views: 4521, Rating: 4.0, Chapters: 378, Readers: 9, Chars: "638K", AIPercent: "13.2%", Description: "Sweet and fluffy, incredibly romantic.", CoverURL: "", Genres: [{ ID: 22, Slug: "romance", Name: "Romance" }, { ID: 35, Slug: "urban-life", Name: "Urban Life" }] },
  { ID: 4, Title: "The Corpse Family is Heavy", AltTitle: "尸家重地", Slug: "the-corpse-family-is-heavy", Author: "佚名", Status: "completed", Views: 890, Rating: 3.2, Chapters: 252, Readers: 11, Chars: "469K", AIPercent: "19.8%", Description: "You can't be greedy for cheap deals.", CoverURL: "", Genres: [{ ID: 1, Slug: "action", Name: "Action" }, { ID: 2, Slug: "adult", Name: "Adult" }, { ID: 12, Slug: "fantasy", Name: "Fantasy" }, { ID: 20, Slug: "mystery", Name: "Mystery" }, { ID: 33, Slug: "supernatural", Name: "Supernatural" }] },
  { ID: 5, Title: "Can You Please Comfort Me?", AltTitle: "可不可以哄哄我", Slug: "can-you-please-comfort-me", Author: "佚名", Status: "completed", Views: 1567, Rating: 3.0, Chapters: 149, Readers: 9, Chars: "235K", AIPercent: "33.6%", Description: "As a child, Shen Shengsheng played house.", CoverURL: "", Genres: [{ ID: 5, Slug: "drama", Name: "Drama" }, { ID: 14, Slug: "josei", Name: "Josei" }, { ID: 22, Slug: "romance", Name: "Romance" }, { ID: 34, Slug: "tragedy", Name: "Tragedy" }, { ID: 35, Slug: "urban-life", Name: "Urban Life" }] },
  { ID: 6, Title: "First-rank Di Consort", AltTitle: "一品嫡妃", Slug: "first-rank-di-consort", Author: "佚名", Status: "completed", Views: 3456, Rating: 3.6, Chapters: 387, Readers: 6, Chars: "2.99M", AIPercent: "15.5%", Description: "Song Anran, a wealthy and beautiful woman.", CoverURL: "", Genres: [{ ID: 11, Slug: "historical", Name: "Historical" }, { ID: 22, Slug: "romance", Name: "Romance" }] },
  { ID: 7, Title: "I am the Crown Prince of the Ming Dynasty", AltTitle: "我在大明当太子", Slug: "i-am-the-crown-prince-of-the-ming-dynasty", Author: "佚名", Status: "completed", Views: 78901, Rating: 4.2, Chapters: 1592, Readers: 505, Chars: "2.96M", AIPercent: "3.14%", Description: "College student Zhu Yu transmigrates.", CoverURL: "", Genres: [{ ID: 1, Slug: "action", Name: "Action" }, { ID: 2, Slug: "adult", Name: "Adult" }, { ID: 4, Slug: "adventure", Name: "Adventure" }, { ID: 5, Slug: "drama", Name: "Drama" }, { ID: 12, Slug: "fantasy", Name: "Fantasy" }, { ID: 11, Slug: "historical", Name: "Historical" }, { ID: 19, Slug: "military", Name: "Military" }, { ID: 38, Slug: "xuanhuan", Name: "Xuanhuan" }] },
  { ID: 8, Title: "Could I Really End Up 'collapsing My Image' Even in the World of Rule Horror", AltTitle: "我还能在规则怪谈里塌房不成？", Slug: "could-i-really-end-up-collapsing-my-image", Author: "佚名", Status: "ongoing", Views: 8120, Rating: 4.1, Chapters: 925, Readers: 15, Chars: "1.75M", AIPercent: "5.4%", Description: "Infinite Flow + Rule-Based Ghost Stories.", CoverURL: "", Genres: [{ ID: 20, Slug: "mystery", Name: "Mystery" }, { ID: 21, Slug: "psychological", Name: "Psychological" }] },
  { ID: 9, Title: "The Legend of the Mountain and Sea Demon Subduing", AltTitle: "大丰小道士", Slug: "legend-mountain-sea-demon", Author: "佚名", Status: "completed", Views: 4580, Rating: 3.9, Chapters: 1522, Readers: 3, Chars: "2.82M", AIPercent: "3.29%", Description: "In the realm of mountains and seas.", CoverURL: "", Genres: [{ ID: 1, Slug: "action", Name: "Action" }, { ID: 4, Slug: "adventure", Name: "Adventure" }, { ID: 16, Slug: "martial-arts", Name: "Martial Arts" }] },
  { ID: 10, Title: "Don't Be Too Wild", AltTitle: "别太野", Slug: "dont-be-too-wild", Author: "佚名", Status: "ongoing", Views: 7020, Rating: 3.4, Chapters: 160, Readers: 13, Chars: "309K", AIPercent: "31.3%", Description: "A seemingly innocent but actually rebellious heiress.", CoverURL: "", Genres: [{ ID: 22, Slug: "romance", Name: "Romance" }, { ID: 23, Slug: "school-life", Name: "School Life" }, { ID: 30, Slug: "slice-of-life", Name: "Slice of Life" }] },
  { ID: 11, Title: "Naruto: In Konoha Village, I Awakened Wood Release at the Start", AltTitle: "火影：木叶村，开局觉醒木遁", Slug: "naruto-konoha-wood-release", Author: "佚名", Status: "ongoing", Views: 123456, Rating: 3.8, Chapters: 1002, Readers: 342, Chars: "1.8M", AIPercent: "8.5%", Description: "Konoha 52nd year. Chiba awakened her memories.", CoverURL: "", Genres: [{ ID: 1, Slug: "action", Name: "Action" }, { ID: 8, Slug: "fan-fiction", Name: "Fan-Fiction" }, { ID: 12, Slug: "fantasy", Name: "Fantasy" }, { ID: 16, Slug: "martial-arts", Name: "Martial Arts" }, { ID: 25, Slug: "seinen", Name: "Seinen" }, { ID: 33, Slug: "supernatural", Name: "Supernatural" }] },
  { ID: 12, Title: "I Just Started High School, But the System Insists I'm an Emperor in My Twilight Years", AltTitle: "刚上高一，系统非说我是晚年大帝", Slug: "high-school-emperor-system", Author: "佚名", Status: "ongoing", Views: 51120, Rating: 1.9, Chapters: 264, Readers: 88, Chars: "450K", AIPercent: "12%", Description: "Jiang Feng is an ordinary high school student.", CoverURL: "", Genres: [{ ID: 1, Slug: "action", Name: "Action" }, { ID: 3, Slug: "comedy", Name: "Comedy" }, { ID: 12, Slug: "fantasy", Name: "Fantasy" }, { ID: 16, Slug: "martial-arts", Name: "Martial Arts" }, { ID: 23, Slug: "school-life", Name: "School Life" }, { ID: 33, Slug: "supernatural", Name: "Supernatural" }, { ID: 35, Slug: "urban-life", Name: "Urban Life" }] },
];

const SORT_OPTIONS = [
  { value: "created_at", label: "Addition Date" },
  { value: "title", label: "Name" },
  { value: "views", label: "View" },
  { value: "readers", label: "Reader" },
  { value: "chapters", label: "Chapter" },
];

const STATUS_OPTIONS = ["All", "Ongoing", "Completed"];

export default function NovelListPage() {
  const [data, setData] = useState<Novel[]>(mockNovels);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(7687);
  const [sort, setSort] = useState("created_at");
  const [order, setOrder] = useState("desc");
  const [status, setStatus] = useState("");
  const [genre, setGenre] = useState("");
  const [showGenreDropdown, setShowGenreDropdown] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        const res = await novels.list({ page, limit: 20, status: status || undefined, genre: genre || undefined, sort, order });
        setData(res.data);
        setTotalPages(res.total_pages);
      } catch {
        const filtered = mockNovels.filter((n) => {
          if (status && n.Status !== status) return false;
          if (genre && !n.Genres.some((g) => g.Slug === genre)) return false;
          return true;
        });
        const sorted = [...filtered].sort((a, b) => {
          let cmp = 0;
          if (sort === "title") cmp = a.Title.localeCompare(b.Title);
          else if (sort === "views") cmp = a.Views - b.Views;
          else if (sort === "readers") cmp = a.Readers - b.Readers;
          else if (sort === "chapters") cmp = a.Chapters - b.Chapters;
          else cmp = a.ID - b.ID;
          return order === "desc" ? -cmp : cmp;
        });
        setData(sorted);
        setTotalPages(Math.ceil(filtered.length / 20) || 1);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, [page, sort, order, status, genre]);

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold text-white mb-6">Novel List</h1>

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-4 mb-6">
        {/* Sort */}
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-400">Order by</span>
          <select
            value={sort}
            onChange={(e) => { setSort(e.target.value); setPage(1); }}
            className="bg-[#1e1e3a] text-sm text-gray-200 px-3 py-2 rounded-lg border border-[#2a2a4a] outline-none"
          >
            {SORT_OPTIONS.map((o) => (
              <option key={o.value} value={o.value}>{o.label}</option>
            ))}
          </select>
        </div>

        {/* Order direction */}
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-400">Order</span>
          <select
            value={order}
            onChange={(e) => { setOrder(e.target.value); setPage(1); }}
            className="bg-[#1e1e3a] text-sm text-gray-200 px-3 py-2 rounded-lg border border-[#2a2a4a] outline-none"
          >
            <option value="desc">Descending</option>
            <option value="asc">Ascending</option>
          </select>
        </div>

        {/* Status filter */}
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-400">Status</span>
          <div className="flex bg-[#1e1e3a] rounded-lg p-0.5 border border-[#2a2a4a]">
            {STATUS_OPTIONS.map((s) => (
              <button
                key={s}
                onClick={() => { setStatus(s === "All" ? "" : s.toLowerCase()); setPage(1); }}
                className={`px-3 py-1.5 text-xs rounded-md transition-colors ${
                  (s === "All" && !status) || s.toLowerCase() === status
                    ? "bg-violet-600 text-white"
                    : "text-gray-400 hover:text-white"
                }`}
              >
                {s}
              </button>
            ))}
          </div>
        </div>

        {/* Genre filter */}
        <div className="relative">
          <span className="text-sm text-gray-400 mr-2">Genre</span>
          <button
            onClick={() => setShowGenreDropdown(!showGenreDropdown)}
            className="bg-[#1e1e3a] text-sm text-gray-200 px-3 py-2 rounded-lg border border-[#2a2a4a] outline-none min-w-[120px] text-left"
          >
            {genre || "All"}
          </button>
          {showGenreDropdown && (
            <div className="absolute top-full mt-1 left-0 z-50 bg-[#1e1e3a] border border-[#2a2a4a] rounded-xl p-2 max-h-60 overflow-y-auto w-48 shadow-xl">
              <button
                onClick={() => { setGenre(""); setShowGenreDropdown(false); setPage(1); }}
                className={`block w-full text-left text-sm px-3 py-1.5 rounded ${
                  !genre ? "text-violet-400" : "text-gray-300 hover:text-white"
                }`}
              >
                All
              </button>
              {genreList.map((g) => (
                <button
                  key={g}
                  onClick={() => { setGenre(g); setShowGenreDropdown(false); setPage(1); }}
                  className={`block w-full text-left text-sm px-3 py-1.5 rounded capitalize ${
                    genre === g ? "text-violet-400" : "text-gray-300 hover:text-white"
                  }`}
                >
                  {g.replace(/-/g, " ")}
                </button>
              ))}
            </div>
          )}
        </div>

        {loading && <span className="text-sm text-violet-400 ml-2">Loading...</span>}
      </div>

      {/* Novel Grid */}
      <div className="space-y-4">
        {data.map((novel) => (
          <Link
            key={novel.ID}
            href={`/en/novel/${novel.ID}/${novel.Slug}`}
            className="flex gap-4 p-4 bg-[#12122a] border border-[#1e1e3a] rounded-xl hover:border-violet-800/40 transition-colors group"
          >
            {/* Cover */}
            <div className="w-20 sm:w-24 aspect-[3/4] rounded-lg bg-[#1e1e3a] border border-[#2a2a4a] flex-shrink-0 flex items-center justify-center overflow-hidden">
              {novel.CoverURL ? (
                <img src={novel.CoverURL} alt="" className="w-full h-full object-cover" />
              ) : (
                <svg className="w-8 h-8 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
                </svg>
              )}
            </div>

            {/* Info */}
            <div className="min-w-0 flex-1">
              <h2 className="text-base font-semibold text-white group-hover:text-violet-400 transition-colors line-clamp-2">
                {novel.Title}
              </h2>
              {novel.AltTitle && (
                <p className="text-xs text-gray-500 mt-0.5">{novel.AltTitle}</p>
              )}
              <div className="flex flex-wrap items-center gap-x-3 gap-y-1 mt-2 text-xs text-gray-400">
                <span className={`px-1.5 py-0.5 rounded ${
                  novel.Status === "ongoing" ? "bg-green-900/40 text-green-400" : "bg-blue-900/40 text-blue-400"
                }`}>
                  {novel.Status}
                </span>
                <span>{novel.Views.toLocaleString()} Views</span>
                <span>{novel.Chapters} Chapters</span>
                <span>{novel.Readers} Readers</span>
                <span>{novel.Chars}</span>
                {novel.Rating > 0 && <span>★ {novel.Rating.toFixed(1)}</span>}
                <span>AI {novel.AIPercent}</span>
              </div>
              <div className="flex flex-wrap gap-1.5 mt-2">
                {novel.Genres.map((g) => (
                  <button
                    key={g.ID}
                    onClick={(e) => { e.preventDefault(); setGenre(g.Slug); setPage(1); }}
                    className="text-xs px-2 py-0.5 rounded-full bg-violet-900/40 text-violet-300 border border-violet-800/30 hover:bg-violet-800/50 transition-colors"
                  >
                    {g.Name}
                  </button>
                ))}
              </div>
              <p className="text-sm text-gray-500 mt-2 line-clamp-2 leading-relaxed">
                {novel.Description}
              </p>
              <div className="mt-2">
                <span className="text-xs text-violet-400 hover:text-violet-300 transition-colors">
                  Novel Details →
                </span>
              </div>
            </div>
          </Link>
        ))}
      </div>

      {/* Pagination */}
      <div className="flex items-center justify-center gap-2 mt-8">
        <button
          onClick={() => setPage(Math.max(1, page - 1))}
          disabled={page <= 1}
          className="px-3 py-1.5 text-sm rounded-lg bg-[#1e1e3a] text-gray-300 hover:bg-[#2a2a4a] disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
        >
          Previous
        </button>

        {page > 2 && (
          <button onClick={() => setPage(1)} className="px-3 py-1.5 text-sm rounded-lg bg-[#1e1e3a] text-gray-400 hover:text-white transition-colors">
            1
          </button>
        )}
        {page > 3 && <span className="text-gray-600 text-sm">...</span>}

        {page > 1 && (
          <button onClick={() => setPage(page - 1)} className="px-3 py-1.5 text-sm rounded-lg bg-[#1e1e3a] text-gray-400 hover:text-white transition-colors">
            {page - 1}
          </button>
        )}

        <span className="px-3 py-1.5 text-sm rounded-lg bg-violet-600 text-white font-medium">
          {page}
        </span>

        {page < totalPages && (
          <button onClick={() => setPage(page + 1)} className="px-3 py-1.5 text-sm rounded-lg bg-[#1e1e3a] text-gray-400 hover:text-white transition-colors">
            {page + 1}
          </button>
        )}

        {page < totalPages - 2 && <span className="text-gray-600 text-sm">...</span>}
        {page < totalPages - 1 && (
          <button onClick={() => setPage(totalPages)} className="px-3 py-1.5 text-sm rounded-lg bg-[#1e1e3a] text-gray-400 hover:text-white transition-colors">
            {totalPages}
          </button>
        )}

        <button
          onClick={() => setPage(Math.min(totalPages, page + 1))}
          disabled={page >= totalPages}
          className="px-3 py-1.5 text-sm rounded-lg bg-[#1e1e3a] text-gray-300 hover:bg-[#2a2a4a] disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
        >
          Next
        </button>

        <span className="text-xs text-gray-600 ml-2">{page} / {totalPages}</span>
      </div>
    </div>
  );
}
