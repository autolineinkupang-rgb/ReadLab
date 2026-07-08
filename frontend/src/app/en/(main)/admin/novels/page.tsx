"use client";

import { useEffect, useState, useRef } from "react";
import Link from "next/link";
import { novels, genres as genresApi, adminNovels } from "@/lib/api";
import { parseMarkdownNovel, ParsedNovel } from "@/lib/novelMarkdownImport";
import Card from "@/components/ui/Card";

interface NovelItem {
  ID: number;
  Title: string;
  AltTitle: string;
  Slug: string;
  Author: string;
  Status: string;
  Chapters: number;
  Views: number;
  Rating: number;
  Genres: { ID: number; Slug: string; Name: string }[];
}

export default function AdminNovelsPage() {
  const [data, setData] = useState<NovelItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [message, setMessage] = useState("");
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editForm, setEditForm] = useState<any>({});
  const [genreOptions, setGenreOptions] = useState<{ ID: number; Name: string }[]>([]);
  const [showAddForm, setShowAddForm] = useState(false);
  const [parsedNovel, setParsedNovel] = useState<ParsedNovel | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [addForm, setAddForm] = useState({
    title: "", alt_title: "", author: "", status: "ongoing",
    description: "", cover_url: "", chars: "", ai_percent: "", rating: 0,
    genre_ids: [] as number[],
  });

  useEffect(() => {
    genresApi.list().then((res: any) => {
      if (res.data) setGenreOptions(res.data);
    }).catch(() => {});
  }, []);

  useEffect(() => {
    fetchNovels(1);
  }, []);

  async function fetchNovels(p: number) {
    setLoading(true);
    setPage(p);
    try {
      const res = await novels.list({ page: p, limit: 20, sort: "created_at", order: "desc" });
      setData(res.data || []);
      setTotalPages(res.total_pages || 1);
    } catch {
      setData([]);
    } finally {
      setLoading(false);
    }
  }

  function startEdit(n: NovelItem & { chars?: string; ai_percent?: string }) {
    setEditingId(n.ID);
    setEditForm({
      title: n.Title,
      alt_title: n.AltTitle || "",
      author: n.Author || "",
      status: n.Status,
      description: "",
      cover_url: "",
      chars: (n as any).chars || "",
      ai_percent: (n as any).ai_percent || "",
      rating: n.Rating,
      genre_ids: n.Genres.map((g) => g.ID),
    });
  }

  async function saveEdit(id: number) {
    setMessage("");
    try {
      await adminNovels.update(id, editForm);
      setMessage("Novel updated.");
      setEditingId(null);
      fetchNovels(page);
    } catch (e: any) {
      setMessage(`Update failed: ${e.message}`);
    }
  }

  async function handleDelete(id: number) {
    if (!confirm("Delete this novel? This cannot be undone.")) return;
    setMessage("");
    try {
      await adminNovels.delete(id);
      setMessage("Novel deleted.");
      fetchNovels(page);
    } catch (e: any) {
      setMessage(`Delete failed: ${e.message}`);
    }
  }

  async function handleAddNovel() {
    if (!addForm.title.trim()) return;
    setMessage("");
    try {
      const payload = parsedNovel
        ? { ...addForm, chapters: parsedNovel.chapters.map((ch) => ({ number: ch.number, title: ch.title, content: ch.content })) }
        : addForm;
      await adminNovels.create(payload);
      setMessage("Novel created.");
      setShowAddForm(false);
      setParsedNovel(null);
      setAddForm({ title: "", alt_title: "", author: "", status: "ongoing", description: "", cover_url: "", chars: "", ai_percent: "", rating: 0, genre_ids: [] });
      fetchNovels(page);
    } catch (e: any) {
      setMessage(`Create failed: ${e.message}`);
    }
  }

  async function handleImportMd(file: File) {
    const text = await file.text();
    const parsed = parseMarkdownNovel(text, file.name);
    setParsedNovel(parsed);
    setAddForm((prev: any) => ({ ...prev, title: parsed.title }));
  }

  function toggleAddGenre(id: number) {
    setAddForm((prev: any) => ({
      ...prev,
      genre_ids: prev.genre_ids.includes(id)
        ? prev.genre_ids.filter((g: number) => g !== id)
        : [...prev.genre_ids, id],
    }));
  }

  function toggleGenre(id: number) {
    setEditForm((prev: any) => ({
      ...prev,
      genre_ids: prev.genre_ids.includes(id)
        ? prev.genre_ids.filter((g: number) => g !== id)
        : [...prev.genre_ids, id],
    }));
  }

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-white">Manage Novels</h1>
        <button
          onClick={() => setShowAddForm(!showAddForm)}
          className="px-4 py-2 bg-accent hover:bg-accent-dark text-white text-sm rounded-lg transition-colors"
        >
          {showAddForm ? "Cancel" : "+ Add Novel"}
        </button>
      </div>

      {message && <p className="text-sm text-accent-light mb-4">{message}</p>}

      {showAddForm && (
        <Card className="mb-6 p-4 space-y-3">
          <h2 className="text-white text-sm font-semibold">Add New Novel</h2>
          <div className="flex gap-2">
            <input
              ref={fileInputRef}
              type="file"
              accept=".md,.markdown"
              className="hidden"
              onChange={(e) => {
                const f = e.target.files?.[0];
                if (f) handleImportMd(f);
                e.target.value = "";
              }}
            />
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              className="px-3 py-1.5 bg-violet-900/50 hover:bg-violet-800/50 text-violet-400 text-xs rounded-lg transition-colors"
            >
              + Import from Markdown
            </button>
            {parsedNovel && (
              <button
                type="button"
                onClick={() => { setParsedNovel(null); }}
                className="px-3 py-1.5 bg-red-900/50 hover:bg-red-800/50 text-red-400 text-xs rounded-lg transition-colors"
              >
                Clear Import
              </button>
            )}
          </div>
          <input
            value={addForm.title}
            onChange={(e) => setAddForm((p: any) => ({ ...p, title: e.target.value }))}
            className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
            placeholder="Title *"
          />
          <input
            value={addForm.alt_title}
            onChange={(e) => setAddForm((p: any) => ({ ...p, alt_title: e.target.value }))}
            className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
            placeholder="Alt Title"
          />
          <input
            value={addForm.author}
            onChange={(e) => setAddForm((p: any) => ({ ...p, author: e.target.value }))}
            className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
            placeholder="Author"
          />
          <textarea
            value={addForm.description}
            onChange={(e) => setAddForm((p: any) => ({ ...p, description: e.target.value }))}
            className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none resize-y min-h-[80px]"
            placeholder="Description"
          />
          <input
            value={addForm.cover_url}
            onChange={(e) => setAddForm((p: any) => ({ ...p, cover_url: e.target.value }))}
            className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
            placeholder="Cover Image URL"
          />
          <div className="grid grid-cols-3 gap-3">
            <input
              value={addForm.chars}
              onChange={(e) => setAddForm((p: any) => ({ ...p, chars: e.target.value }))}
              className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
              placeholder="Chars (e.g. 1.2M)"
            />
            <input
              value={addForm.ai_percent}
              onChange={(e) => setAddForm((p: any) => ({ ...p, ai_percent: e.target.value }))}
              className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
              placeholder="AI %"
            />
            <input
              type="number"
              step="0.1"
              min="0"
              max="5"
              value={addForm.rating}
              onChange={(e) => setAddForm((p: any) => ({ ...p, rating: parseFloat(e.target.value) || 0 }))}
              className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
              placeholder="Rating (0-5)"
            />
          </div>
          <select
            value={addForm.status}
            onChange={(e) => setAddForm((p: any) => ({ ...p, status: e.target.value }))}
            className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
          >
            <option value="ongoing">Ongoing</option>
            <option value="completed">Completed</option>
            <option value="hiatus">Hiatus</option>
            <option value="dropped">Dropped</option>
          </select>
          <div>
            <p className="text-xs text-gray-500 mb-1">Genres</p>
            <div className="flex flex-wrap gap-1.5 max-h-32 overflow-y-auto">
              {genreOptions.map((g) => (
                <button
                  key={g.ID}
                  type="button"
                  onClick={() => toggleAddGenre(g.ID)}
                  className={`text-xs px-2 py-1 rounded-full border transition-colors ${
                    addForm.genre_ids.includes(g.ID)
                      ? "bg-accent text-white border-accent"
                      : "bg-card-hover text-gray-400 border-line-light hover:text-white"
                  }`}
                >
                  {g.Name}
                </button>
              ))}
            </div>
          </div>
          {parsedNovel && (
            <div className="border border-line-light rounded-lg p-3 space-y-2">
              <p className="text-xs text-gray-400 font-semibold">
                Imported Chapters ({parsedNovel.chapters.length})
              </p>
              <div className="max-h-48 overflow-y-auto space-y-1">
                {parsedNovel.chapters.map((ch) => (
                  <div key={ch.number} className="flex items-start gap-2 text-xs">
                    <span className="text-gray-500 shrink-0 w-6 text-right">#{ch.number}</span>
                    <div className="min-w-0 flex-1">
                      <p className="text-gray-200 truncate">{ch.title}</p>
                      <p className="text-gray-500 truncate">{ch.content.replace(/<[^>]+>/g, "").slice(0, 80)}...</p>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
          <button
            onClick={handleAddNovel}
            disabled={!addForm.title.trim()}
            className="px-4 py-2 bg-green-600 hover:bg-green-700 disabled:opacity-50 text-white text-sm rounded-lg transition-colors"
          >
            {parsedNovel ? "Create Novel with Chapters" : "Create Novel"}
          </button>
        </Card>
      )}

      {loading ? (
        <p className="text-accent">Loading...</p>
      ) : (
        <>
          <div className="space-y-3">
            {data.map((n) => (
              <Card key={n.ID}>
                {editingId === n.ID ? (
                  <div className="space-y-3">
                    <input
                      value={editForm.title}
                      onChange={(e) => setEditForm((p: any) => ({ ...p, title: e.target.value }))}
                      className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
                      placeholder="Title"
                    />
                    <input
                      value={editForm.alt_title}
                      onChange={(e) => setEditForm((p: any) => ({ ...p, alt_title: e.target.value }))}
                      className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
                      placeholder="Alt Title"
                    />
                    <input
                      value={editForm.cover_url}
                      onChange={(e) => setEditForm((p: any) => ({ ...p, cover_url: e.target.value }))}
                      className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
                      placeholder="Cover URL"
                    />
                    <div className="grid grid-cols-3 gap-3">
                      <input
                        value={editForm.chars || ""}
                        onChange={(e) => setEditForm((p: any) => ({ ...p, chars: e.target.value }))}
                        className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
                        placeholder="Chars (e.g. 1.2M)"
                      />
                      <input
                        value={editForm.ai_percent || ""}
                        onChange={(e) => setEditForm((p: any) => ({ ...p, ai_percent: e.target.value }))}
                        className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
                        placeholder="AI %"
                      />
                      <input
                        type="number"
                        step="0.1"
                        min="0"
                        max="5"
                        value={editForm.rating ?? ""}
                        onChange={(e) => setEditForm((p: any) => ({ ...p, rating: parseFloat(e.target.value) || 0 }))}
                        className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
                        placeholder="Rating (0-5)"
                      />
                    </div>
                    <select
                      value={editForm.status}
                      onChange={(e) => setEditForm((p: any) => ({ ...p, status: e.target.value }))}
                      className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
                    >
                      <option value="ongoing">Ongoing</option>
                      <option value="completed">Completed</option>
                      <option value="hiatus">Hiatus</option>
                      <option value="dropped">Dropped</option>
                    </select>
                    <div>
                      <p className="text-xs text-gray-500 mb-1">Genres</p>
                      <div className="flex flex-wrap gap-1.5 max-h-32 overflow-y-auto">
                        {genreOptions.map((g) => (
                          <button
                            key={g.ID}
                            type="button"
                            onClick={() => toggleGenre(g.ID)}
                            className={`text-xs px-2 py-1 rounded-full border transition-colors ${
                              editForm.genre_ids.includes(g.ID)
                                ? "bg-accent text-white border-accent"
                                : "bg-card-hover text-gray-400 border-line-light hover:text-white"
                            }`}
                          >
                            {g.Name}
                          </button>
                        ))}
                      </div>
                    </div>
                    <div className="flex gap-2">
                      <button onClick={() => saveEdit(n.ID)} className="px-4 py-2 bg-accent hover:bg-accent-dark text-white text-sm rounded-lg transition-colors">
                        Save
                      </button>
                      <button onClick={() => setEditingId(null)} className="px-4 py-2 bg-card-hover text-gray-300 text-sm rounded-lg transition-colors">
                        Cancel
                      </button>
                    </div>
                  </div>
                ) : (
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 sm:gap-4">
                    <div className="flex-1 min-w-0">
                      <h3 className="text-sm font-semibold text-white truncate">{n.Title}</h3>
                      <p className="text-xs text-gray-500 truncate mt-0.5">
                        {n.Author} &middot; {n.Status} &middot; {n.Chapters}ch &middot; {n.Views.toLocaleString()}v &middot; {n.Rating.toFixed(1)}
                      </p>
                      <div className="flex flex-wrap gap-1 mt-1">
                        {n.Genres?.slice(0, 4).map((g) => (
                          <span key={g.ID} className="text-[9px] px-1.5 py-0.5 rounded-full bg-accent/10 text-accent-light/80 border border-accent/20">
                            {g.Name}
                          </span>
                        ))}
                      </div>
                    </div>
                    <div className="flex gap-2 shrink-0">
                      <Link href={`/en/admin/novels/${n.ID}/chapters`} className="px-3 py-1.5 bg-violet-900/50 hover:bg-violet-800/50 text-violet-400 text-xs rounded-lg transition-colors">
                        Ch
                      </Link>
                      <button onClick={() => startEdit(n)} className="flex-1 sm:flex-none px-3 py-1.5 bg-card-hover hover:bg-line-light text-gray-300 text-xs rounded-lg transition-colors">
                        Edit
                      </button>
                      <button onClick={() => handleDelete(n.ID)} className="flex-1 sm:flex-none px-3 py-1.5 bg-red-900/50 hover:bg-red-800/50 text-red-400 text-xs rounded-lg transition-colors">
                        Delete
                      </button>
                    </div>
                  </div>
                )}
              </Card>
            ))}
          </div>

          {totalPages > 1 && (
            <div className="flex items-center justify-center gap-2 mt-8">
              <button
                onClick={() => fetchNovels(Math.max(1, page - 1))}
                disabled={page <= 1}
                className="px-3 py-1.5 text-xs rounded-lg bg-card-hover text-gray-300 hover:bg-line-light disabled:opacity-40 transition-colors"
              >
                Previous
              </button>
              <span className="text-xs text-gray-500">Page {page} / {totalPages}</span>
              <button
                onClick={() => fetchNovels(Math.min(totalPages, page + 1))}
                disabled={page >= totalPages}
                className="px-3 py-1.5 text-xs rounded-lg bg-card-hover text-gray-300 hover:bg-line-light disabled:opacity-40 transition-colors"
              >
                Next
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
