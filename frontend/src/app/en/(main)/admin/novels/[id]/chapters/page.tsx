"use client";

import { useEffect, useState, useRef } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import DOMPurify from "isomorphic-dompurify";
import { novels, adminChapters } from "@/lib/api";
import Card from "@/components/ui/Card";
import ChapterContentEditor from "@/components/admin/ChapterContentEditor";
import type { ChapterContentEditorHandle } from "@/components/admin/ChapterContentEditor";
import { txtToHtml } from "@/lib/htmlImport";

interface ChapterItem {
  id: number;
  novel_id: number;
  number: number;
  title: string;
  content: string;
  is_locked: boolean;
  ticket_cost: number;
  created_at: string;
}

function wordCount(text: string): number {
  return text.trim() ? text.trim().split(/\s+/).length : 0;
}

function charCount(text: string): number {
  return text.length;
}

export default function AdminChaptersPage() {
  const { id } = useParams<{ id: string }>();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const addEditorRef = useRef<ChapterContentEditorHandle>(null);
  const editEditorRef = useRef<ChapterContentEditorHandle>(null);
  const [novel, setNovel] = useState<any>(null);
  const [data, setData] = useState<ChapterItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [message, setMessage] = useState("");
  const [messageType, setMessageType] = useState<"success" | "error">("success");
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editForm, setEditForm] = useState<any>({});
  const [showAddForm, setShowAddForm] = useState(false);
  const [previewMode, setPreviewMode] = useState<"add" | number | null>(null);
  const [addForm, setAddForm] = useState({
    number: "", title: "", content: "", is_locked: false, ticket_cost: 0,
  });

  useEffect(() => {
    novels.get(id).then((res: any) => {
      setNovel(res.data || res.novel || res);
    }).catch(() => {});
  }, [id]);

  useEffect(() => {
    fetchChapters(1);
  }, [id]);

  function showMessage(msg: string, type: "success" | "error" = "success") {
    setMessage(msg);
    setMessageType(type);
  }

  async function fetchChapters(p: number) {
    setLoading(true);
    setPage(p);
    try {
      const res = await adminChapters.list(id, { page: p, limit: 50 });
      setData(res.data || []);
      setTotalPages(res.total_pages || 1);
    } catch {
      setData([]);
    } finally {
      setLoading(false);
    }
  }

  function startEdit(ch: ChapterItem) {
    setEditingId(ch.id);
    setPreviewMode(null);
    setEditForm({
      number: ch.number,
      title: ch.title,
      content: ch.content || "",
      is_locked: ch.is_locked,
      ticket_cost: ch.ticket_cost ?? 0,
    });
  }

  async function saveEdit(chapterId: number) {
    showMessage("");
    try {
      const payload: any = {};
      if (editForm.title !== undefined) payload.title = editForm.title;
      if (editForm.content !== undefined) payload.content = editForm.content;
      if (editForm.is_locked !== undefined) payload.is_locked = editForm.is_locked;
      if (editForm.ticket_cost !== undefined) payload.ticket_cost = editForm.ticket_cost;
      if (editForm.number !== undefined) payload.number = parseInt(editForm.number) || undefined;

      await adminChapters.update(id, chapterId, payload);
      showMessage("Chapter updated.");
      setEditingId(null);
      fetchChapters(page);
    } catch (e: any) {
      showMessage(`Update failed: ${e.message}`, "error");
    }
  }

  async function handleMoveChapter(chapterId: number, currentNum: number, direction: "up" | "down") {
    const newNum = direction === "up" ? currentNum - 1 : currentNum + 1;
    if (newNum < 1) return;

    const target = data.find((c) => c.number === newNum);
    if (!target) return;

    showMessage("");
    try {
      await adminChapters.update(id, chapterId, { number: newNum });
      await adminChapters.update(id, target.id, { number: currentNum });
      showMessage(`Chapter moved ${direction}.`);
      fetchChapters(page);
    } catch (e: any) {
      showMessage(`Move failed: ${e.message}`, "error");
    }
  }

  async function handleDelete(chapterId: number) {
    if (!confirm("Delete this chapter? This cannot be undone.")) return;
    showMessage("");
    try {
      await adminChapters.delete(chapterId);
      showMessage("Chapter deleted.");
      fetchChapters(page);
    } catch (e: any) {
      showMessage(`Delete failed: ${e.message}`, "error");
    }
  }

  async function handleAddChapter() {
    if (!addForm.title.trim()) return;
    showMessage("");
    try {
      await adminChapters.create(id, {
        ...addForm,
        number: addForm.number ? parseInt(addForm.number) : undefined,
        ticket_cost: addForm.ticket_cost || 0,
        is_locked: addForm.is_locked,
      });
      showMessage("Chapter created.");
      setAddForm({ number: "", title: "", content: "", is_locked: false, ticket_cost: 0 });
      fetchChapters(page);
    } catch (e: any) {
      showMessage(`Create failed: ${e.message}`, "error");
    }
  }

  async function handleAddAndAnother() {
    if (!addForm.title.trim()) return;
    showMessage("");
    try {
      await adminChapters.create(id, {
        ...addForm,
        number: addForm.number ? parseInt(addForm.number) : undefined,
        ticket_cost: addForm.ticket_cost || 0,
        is_locked: addForm.is_locked,
      });
      showMessage("Chapter created. Add another...");
      setAddForm({ number: "", title: "", content: "", is_locked: false, ticket_cost: 0 });
      fetchChapters(page);
    } catch (e: any) {
      showMessage(`Create failed: ${e.message}`, "error");
    }
  }

  function handleImportFile(field: "add" | "edit") {
    const ref = field === "add" ? addEditorRef.current : editEditorRef.current;
    if (!ref) {
      fileInputRef.current?.click();
      fileInputRef.current!.dataset.target = field;
      return;
    }
    const input = document.createElement("input");
    input.type = "file";
    input.accept = ".txt";
    input.onchange = () => {
      const file = input.files?.[0];
      if (!file) return;
      const reader = new FileReader();
      reader.onload = (ev) => {
        const text = ev.target?.result as string || "";
        ref.importText(txtToHtml(text));
      };
      reader.readAsText(file);
    };
    input.click();
  }

  function formatDate(dateStr: string) {
    if (!dateStr) return "-";
    const d = new Date(dateStr);
    return d.toLocaleDateString("en-US", { year: "numeric", month: "short", day: "numeric" });
  }

  function renderPreview(html: string) {
    if (!html) return null;
    const allowed = ["p", "h2", "h3", "strong", "em", "u", "s", "ul", "ol", "li", "blockquote", "hr", "br"];
    const clean = DOMPurify.sanitize(html, { ALLOWED_TAGS: allowed, ALLOWED_ATTR: ["href", "target", "rel", "class", "id"] });
    return <div className="chapter-content text-gray-300 text-sm" dangerouslySetInnerHTML={{ __html: clean }} />;
  }

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <Link href="/en/admin/novels" className="text-xs text-gray-500 hover:text-accent transition-colors">&larr; Back to Novels</Link>
          <h1 className="text-2xl font-bold text-white mt-1">
            {novel ? novel.Title || novel.title || "Chapters" : "Chapters"}
          </h1>
        </div>
        <button
          onClick={() => { setShowAddForm(!showAddForm); setPreviewMode(null); }}
          className="px-4 py-2 bg-accent hover:bg-accent-dark text-white text-sm rounded-lg transition-colors"
        >
          {showAddForm ? "Cancel" : "+ Add Chapter"}
        </button>
      </div>

      {message && (
        <p className={`text-sm mb-4 ${messageType === "error" ? "text-red-400" : "text-accent-light"}`}>
          {message}
        </p>
      )}

      {showAddForm && (
        <Card className="mb-6 p-4 space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="text-white text-sm font-semibold">Add New Chapter</h2>
            <div className="flex items-center gap-2">
              <button
                onClick={() => handleImportFile("add")}
                className="px-3 py-1.5 bg-card-hover hover:bg-line-light text-gray-300 text-xs rounded-lg transition-colors"
              >
                Import .txt
              </button>
              <button
                onClick={() => setPreviewMode(previewMode === "add" ? null : "add")}
                className={`px-3 py-1.5 text-xs rounded-lg transition-colors ${
                  previewMode === "add"
                    ? "bg-accent text-white"
                    : "bg-card-hover hover:bg-line-light text-gray-300"
                }`}
              >
                {previewMode === "add" ? "Edit" : "Preview"}
              </button>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <input
              type="number"
              min="1"
              value={addForm.number}
              onChange={(e) => setAddForm((p: any) => ({ ...p, number: e.target.value }))}
              className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
              placeholder="Number (auto if empty)"
            />
            <input
              value={addForm.title}
              onChange={(e) => setAddForm((p: any) => ({ ...p, title: e.target.value }))}
              className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
              placeholder="Title *"
            />
          </div>

          {previewMode === "add" ? (
            <div className="w-full bg-black/30 border border-line-light rounded-lg px-3 py-3 text-sm min-h-[300px] max-h-[500px] overflow-y-auto">
              {addForm.content ? renderPreview(addForm.content) : <p className="text-gray-500 italic">No content to preview</p>}
            </div>
          ) : (
            <ChapterContentEditor
              ref={addEditorRef}
              value={addForm.content}
              onChange={(html) => setAddForm((p: any) => ({ ...p, content: html }))}
              onImportError={(msg) => showMessage(msg, "error")}
            />
          )}

          <div className="flex items-center justify-between">
            <div className="flex items-center gap-6">
              <label className="flex items-center gap-2 text-sm text-gray-300 cursor-pointer">
                <input
                  type="checkbox"
                  checked={addForm.is_locked}
                  onChange={(e) => setAddForm((p: any) => ({ ...p, is_locked: e.target.checked }))}
                  className="rounded border-line-light bg-card-hover text-accent focus:ring-accent"
                />
                Locked
              </label>
              <div className="flex items-center gap-2">
                <span className="text-xs text-gray-500">Ticket Cost:</span>
                <input
                  type="number"
                  min="0"
                  value={addForm.ticket_cost}
                  onChange={(e) => setAddForm((p: any) => ({ ...p, ticket_cost: parseInt(e.target.value) || 0 }))}
                  className="w-20 bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
                />
              </div>
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleAddAndAnother}
                disabled={!addForm.title.trim()}
                className="px-4 py-2 bg-card-hover hover:bg-line-light disabled:opacity-50 text-gray-300 text-sm rounded-lg transition-colors"
              >
                Save & Add Another
              </button>
              <button
                onClick={handleAddChapter}
                disabled={!addForm.title.trim()}
                className="px-4 py-2 bg-green-600 hover:bg-green-700 disabled:opacity-50 text-white text-sm rounded-lg transition-colors"
              >
                Create Chapter
              </button>
            </div>
          </div>
        </Card>
      )}

      {loading ? (
        <div className="flex items-center justify-center py-16">
          <div className="w-8 h-8 border-2 border-accent border-t-transparent rounded-full animate-spin" />
        </div>
      ) : data.length === 0 ? (
        <div className="text-center py-16">
          <p className="text-sm text-gray-500 mb-4">No chapters yet.</p>
          <button
            onClick={() => setShowAddForm(true)}
            className="px-4 py-2 bg-accent hover:bg-accent-dark text-white text-sm rounded-lg transition-colors"
          >
            Create First Chapter
          </button>
        </div>
      ) : (<div className="space-y-2">
        {data.map((ch, idx) => (
        <div key={ch.id} className="bg-card border border-line rounded-xl p-4">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 sm:gap-4">
            <div className="flex items-center gap-3 flex-1 min-w-0">
              <div className="flex flex-col items-center gap-0.5 shrink-0">
                <button
                  onClick={() => handleMoveChapter(ch.id, ch.number, "up")}
                  disabled={idx === 0}
                  className="text-[10px] text-gray-600 hover:text-white disabled:opacity-30 transition-colors leading-none"
                  title="Move up"
                >
                  ▲
                </button>
                <span className="text-xs font-mono text-gray-500 w-6 text-center">{ch.number}</span>
                <button
                  onClick={() => handleMoveChapter(ch.id, ch.number, "down")}
                  disabled={idx === data.length - 1}
                  className="text-[10px] text-gray-600 hover:text-white disabled:opacity-30 transition-colors leading-none"
                  title="Move down"
                >
                  ▼
                </button>
              </div>
              <div className="min-w-0 flex-1">
                <h3 className="text-sm font-semibold text-white truncate">{ch.title}</h3>
                <p className="text-xs text-gray-500 mt-0.5">
                  {formatDate(ch.created_at)} &middot; {wordCount(ch.content || "")} words
                </p>
              </div>
            </div>
            <div className="flex items-center gap-3 shrink-0">
              <span className={`text-xs ${ch.is_locked ? "text-amber-400" : "text-green-400"}`}>
                {ch.is_locked ? `Locked (${ch.ticket_cost}t)` : "Free"}
              </span>
              <div className="flex gap-2">
                <button onClick={() => startEdit(ch)} className="px-3 py-1.5 bg-card-hover hover:bg-line-light text-gray-300 text-xs rounded-lg transition-colors">
                  Edit
                </button>
                <button onClick={() => handleDelete(ch.id)} className="px-3 py-1.5 bg-red-900/50 hover:bg-red-800/50 text-red-400 text-xs rounded-lg transition-colors">
                  Delete
                </button>
              </div>
            </div>
          </div>
          {editingId === ch.id && (
            <div className="mt-4 pt-4 border-t border-line-light space-y-3">
              <div className="flex items-center justify-between">
                <h3 className="text-white text-sm font-semibold">Edit Chapter #{ch.number}</h3>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => handleImportFile("edit")}
                    className="px-3 py-1.5 bg-card-hover hover:bg-line-light text-gray-300 text-xs rounded-lg transition-colors"
                  >
                    Import .txt
                  </button>
                  <button
                    onClick={() => setPreviewMode(previewMode === ch.id ? null : ch.id)}
                    className={`px-3 py-1.5 text-xs rounded-lg transition-colors ${
                      previewMode === ch.id
                        ? "bg-accent text-white"
                        : "bg-card-hover hover:bg-line-light text-gray-300"
                    }`}
                  >
                    {previewMode === ch.id ? "Edit" : "Preview"}
                  </button>
                </div>
              </div>

              <div className="grid grid-cols-3 gap-3">
                <input
                  type="number"
                  min="1"
                  value={editForm.number ?? ""}
                  onChange={(e) => setEditForm((p: any) => ({ ...p, number: e.target.value }))}
                  className="w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
                  placeholder="Number"
                />
                <input
                  value={editForm.title}
                  onChange={(e) => setEditForm((p: any) => ({ ...p, title: e.target.value }))}
                  className="col-span-2 w-full bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
                  placeholder="Title"
                />
              </div>

              {previewMode === ch.id ? (
                <div className="w-full bg-black/30 border border-line-light rounded-lg px-3 py-3 text-sm min-h-[300px] max-h-[500px] overflow-y-auto">
                  {editForm.content ? renderPreview(editForm.content) : <p className="text-gray-500 italic">No content</p>}
                </div>
              ) : (
                <ChapterContentEditor
                  ref={editEditorRef}
                  value={editForm.content}
                  onChange={(html) => setEditForm((p: any) => ({ ...p, content: html }))}
                  onImportError={(msg) => showMessage(msg, "error")}
                />
              )}

              <div className="flex items-center gap-6">
                <label className="flex items-center gap-2 text-sm text-gray-300 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={editForm.is_locked}
                    onChange={(e) => setEditForm((p: any) => ({ ...p, is_locked: e.target.checked }))}
                    className="rounded border-line-light bg-card-hover text-accent focus:ring-accent"
                  />
                  Locked
                </label>
                <div className="flex items-center gap-2">
                  <span className="text-xs text-gray-500">Ticket Cost:</span>
                  <input
                    type="number"
                    min="0"
                    value={editForm.ticket_cost ?? 0}
                    onChange={(e) => setEditForm((p: any) => ({ ...p, ticket_cost: parseInt(e.target.value) || 0 }))}
                    className="w-20 bg-card-hover border border-line-light rounded-lg px-3 py-2 text-sm text-gray-200 outline-none"
                  />
                </div>
              </div>
              <div className="flex gap-2">
                <button onClick={() => saveEdit(ch.id)} className="px-4 py-2 bg-accent hover:bg-accent-dark text-white text-sm rounded-lg transition-colors">
                  Save
                </button>
                <button onClick={() => { setEditingId(null); setPreviewMode(null); }} className="px-4 py-2 bg-card-hover text-gray-300 text-sm rounded-lg transition-colors">
                  Cancel
                </button>
              </div>
            </div>
          )}
        </div>
      ))}
      {totalPages > 1 && (
            <div className="flex items-center justify-center gap-2 mt-8">
              <button
                onClick={() => fetchChapters(Math.max(1, page - 1))}
                disabled={page <= 1}
                className="px-3 py-1.5 text-xs rounded-lg bg-card-hover text-gray-300 hover:bg-line-light disabled:opacity-40 transition-colors"
              >
                Previous
              </button>
              <span className="text-xs text-gray-500">Page {page} / {totalPages}</span>
              <button
                onClick={() => fetchChapters(Math.min(totalPages, page + 1))}
                disabled={page >= totalPages}
                className="px-3 py-1.5 text-xs rounded-lg bg-card-hover text-gray-300 hover:bg-line-light disabled:opacity-40 transition-colors"
              >
                Next
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
