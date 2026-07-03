"use client";

import { useState } from "react";
import Link from "next/link";

const genreList = [
  "action","adult","adventure","comedy","drama","ecchi","erciyuan","fan-fiction","fantasy",
  "game","gender-bender","harem","historical","horror","josei","martial-arts","mature",
  "mecha","military","mystery","psychological","romance","school-life","sci-fi","seinen",
  "shoujo","shoujo-ai","shounen","shounen-ai","slice-of-life","smut","sports","supernatural",
  "tragedy","urban-life","wuxia","xianxia","xuanhuan","yaoi","yuri",
];

const statusOptions = ["ongoing", "completed"];
const sortOptions = [
  { value: "created_at", label: "Addition Date" },
  { value: "title", label: "Name" },
  { value: "views", label: "Views" },
  { value: "chapters", label: "Chapters" },
  { value: "rating", label: "Rating" },
];

export default function NovelFinderPage() {
  const [title, setTitle] = useState("");
  const [author, setAuthor] = useState("");
  const [selectedGenres, setSelectedGenres] = useState<string[]>([]);
  const [status, setStatus] = useState("");
  const [sort, setSort] = useState("created_at");
  const [order, setOrder] = useState("desc");
  const [searched, setSearched] = useState(false);

  const handleSearch = () => setSearched(true);

  const toggleGenre = (g: string) => {
    setSelectedGenres((prev) =>
      prev.includes(g) ? prev.filter((x) => x !== g) : [...prev, g]
    );
  };

  const buildUrl = () => {
    const params = new URLSearchParams();
    if (title) params.set("q", title);
    if (author) params.set("author", author);
    if (status) params.set("status", status);
    if (selectedGenres.length > 0) params.set("genre", selectedGenres[0]);
    if (sort) params.set("sort", sort);
    if (order) params.set("order", order);
    return `/en/novel-list?${params.toString()}`;
  };

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold text-white mb-6">Advanced Search</h1>

      <div className="bg-[#12122a] border border-[#1e1e3a] rounded-xl p-6 space-y-6">
        {/* Title */}
        <div>
          <label className="block text-sm text-gray-400 mb-1">Title</label>
          <input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Novel title..."
            className="w-full bg-[#1e1e3a] border border-[#2a2a4a] rounded-lg px-4 py-2.5 text-sm text-gray-200 outline-none focus:border-violet-600 transition-colors"
          />
        </div>

        {/* Author */}
        <div>
          <label className="block text-sm text-gray-400 mb-1">Author</label>
          <input
            value={author}
            onChange={(e) => setAuthor(e.target.value)}
            placeholder="Author name..."
            className="w-full bg-[#1e1e3a] border border-[#2a2a4a] rounded-lg px-4 py-2.5 text-sm text-gray-200 outline-none focus:border-violet-600 transition-colors"
          />
        </div>

        {/* Status */}
        <div>
          <label className="block text-sm text-gray-400 mb-1">Status</label>
          <div className="flex gap-2">
            <button
              onClick={() => setStatus("")}
              className={`px-4 py-1.5 text-xs rounded-lg border transition-colors ${
                !status ? "bg-violet-600 text-white border-violet-500" : "bg-[#1e1e3a] text-gray-400 border-[#2a2a4a] hover:text-white"
              }`}
            >
              All
            </button>
            {statusOptions.map((s) => (
              <button
                key={s}
                onClick={() => setStatus(s)}
                className={`px-4 py-1.5 text-xs rounded-lg border capitalize transition-colors ${
                  status === s ? "bg-violet-600 text-white border-violet-500" : "bg-[#1e1e3a] text-gray-400 border-[#2a2a4a] hover:text-white"
                }`}
              >
                {s}
              </button>
            ))}
          </div>
        </div>

        {/* Genres */}
        <div>
          <label className="block text-sm text-gray-400 mb-2">Genres</label>
          <div className="flex flex-wrap gap-1.5 max-h-48 overflow-y-auto">
            {genreList.map((g) => (
              <button
                key={g}
                onClick={() => toggleGenre(g)}
                className={`text-xs px-2.5 py-1 rounded-full border transition-colors capitalize ${
                  selectedGenres.includes(g)
                    ? "bg-violet-600 text-white border-violet-500"
                    : "bg-[#1e1e3a] text-gray-400 border-[#2a2a4a] hover:text-white"
                }`}
              >
                {g.replace(/-/g, " ")}
              </button>
            ))}
          </div>
        </div>

        {/* Sort */}
        <div className="flex flex-wrap gap-4">
          <div>
            <label className="block text-sm text-gray-400 mb-1">Sort by</label>
            <select
              value={sort}
              onChange={(e) => setSort(e.target.value)}
              className="bg-[#1e1e3a] text-sm text-gray-200 px-3 py-2 rounded-lg border border-[#2a2a4a] outline-none"
            >
              {sortOptions.map((o) => (
                <option key={o.value} value={o.value}>{o.label}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm text-gray-400 mb-1">Order</label>
            <select
              value={order}
              onChange={(e) => setOrder(e.target.value)}
              className="bg-[#1e1e3a] text-sm text-gray-200 px-3 py-2 rounded-lg border border-[#2a2a4a] outline-none"
            >
              <option value="desc">Descending</option>
              <option value="asc">Ascending</option>
            </select>
          </div>
        </div>

        {/* Search button */}
        <Link
          href={buildUrl()}
          onClick={handleSearch}
          className="inline-block px-8 py-3 bg-violet-600 hover:bg-violet-700 text-white text-sm font-medium rounded-lg transition-colors"
        >
          Search
        </Link>
      </div>

      {searched && (
        <div className="mt-8 text-center text-sm text-gray-500">
          <p>Search results will appear on the Novel List page.</p>
          <Link href={buildUrl()} className="text-violet-400 hover:text-violet-300 underline mt-1 inline-block">
            Go to results →
          </Link>
        </div>
      )}
    </div>
  );
}
