"use client";

import { useEffect, useState } from "react";
import { news, adminNews } from "@/lib/api";
import Card from "@/components/ui/Card";
import RequireRole from "@/components/RequireRole";

interface NewsItem {
  ID: number;
  Title: string;
  Content: string;
  Type: string;
  Slug: string;
  CreatedAt: string;
}

export default function AdminNewsPage() {
  return (
    <RequireRole roles={["admin"]}>
      <AdminNews />
    </RequireRole>
  );
}

function AdminNews() {
  const [items, setItems] = useState<NewsItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState("");
  const [editId, setEditId] = useState<number | null>(null);
  const [formTitle, setFormTitle] = useState("");
  const [formContent, setFormContent] = useState("");
  const [formType, setFormType] = useState("news");
  const [showForm, setShowForm] = useState(false);

  useEffect(() => {
    fetchNews();
  }, []);

  async function fetchNews() {
    setLoading(true);
    try {
      const res = await news.list({ limit: 50 });
      setItems((res.data || []) as NewsItem[]);
    } catch {
      setItems([]);
    } finally {
      setLoading(false);
    }
  }

  function resetForm() {
    setFormTitle("");
    setFormContent("");
    setFormType("news");
    setEditId(null);
    setShowForm(false);
  }

  function openEdit(item: NewsItem) {
    setFormTitle(item.Title);
    setFormContent(item.Content);
    setFormType(item.Type);
    setEditId(item.ID);
    setShowForm(true);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setMessage("");
    try {
      if (editId) {
        await adminNews.update(editId, { title: formTitle, content: formContent, type: formType });
        setMessage("News updated.");
      } else {
        await adminNews.create({ title: formTitle, content: formContent, type: formType });
        setMessage("News created.");
      }
      resetForm();
      fetchNews();
    } catch (err: any) {
      setMessage(err.message || "Failed");
    }
  }

  async function handleDelete(id: number) {
    if (!confirm("Delete this news?")) return;
    setMessage("");
    try {
      await adminNews.delete(id);
      setMessage("News deleted.");
      fetchNews();
    } catch (err: any) {
      setMessage(err.message || "Delete failed");
    }
  }

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-white">Manage News</h1>
        <button
          onClick={() => { resetForm(); setShowForm(!showForm); }}
          className="px-4 py-2 bg-violet-600 hover:bg-violet-700 text-white text-sm rounded-lg transition-colors"
        >
          {showForm ? "Cancel" : "+ New News"}
        </button>
      </div>

      {message && <p className="text-sm text-accent-light mb-4">{message}</p>}

      {showForm && (
        <form onSubmit={handleSubmit} className="bg-card border border-line rounded-xl p-6 mb-6 space-y-4">
          <div className="flex gap-4">
            <div className="flex-1">
              <label className="block text-sm text-gray-400 mb-1">Title *</label>
              <input value={formTitle} onChange={(e) => setFormTitle(e.target.value)} required
                className="w-full bg-card-hover border border-line-light rounded-lg px-4 py-2.5 text-sm text-gray-200 outline-none focus:border-violet-600" />
            </div>
            <div className="w-40">
              <label className="block text-sm text-gray-400 mb-1">Type *</label>
              <select value={formType} onChange={(e) => setFormType(e.target.value)}
                className="w-full bg-card-hover border border-line-light rounded-lg px-4 py-2.5 text-sm text-gray-200 outline-none focus:border-violet-600">
                <option value="news">News</option>
                <option value="changelog">Changelog</option>
              </select>
            </div>
          </div>
          <div>
            <label className="block text-sm text-gray-400 mb-1">Content *</label>
            <textarea value={formContent} onChange={(e) => setFormContent(e.target.value)} required rows={8}
              className="w-full bg-card-hover border border-line-light rounded-lg px-4 py-2.5 text-sm text-gray-200 outline-none focus:border-violet-600 resize-y" />
          </div>
          <button type="submit" className="px-6 py-2.5 bg-violet-600 hover:bg-violet-700 text-white text-sm font-medium rounded-lg transition-colors">
            {editId ? "Update News" : "Create News"}
          </button>
        </form>
      )}

      {loading ? (
        <div className="flex items-center justify-center py-16">
          <div className="w-8 h-8 border-2 border-accent border-t-transparent rounded-full animate-spin" />
        </div>
      ) : (
        <>
          <div className="space-y-3">
            {items.map((item) => (
              <Card key={item.ID}>
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-semibold text-white">{item.Title}</span>
                      <span className={`text-[10px] px-1.5 py-0.5 rounded border ${
                        item.Type === "changelog"
                          ? "bg-blue-900/40 text-blue-400 border-blue-800/30"
                          : "bg-green-900/40 text-green-400 border-green-800/30"
                      }`}>
                        {item.Type}
                      </span>
                    </div>
                    <p className="text-xs text-gray-500 mt-1">{new Date(item.CreatedAt).toLocaleDateString()}</p>
                    <p className="text-xs text-gray-400 mt-1 line-clamp-2">{item.Content}</p>
                  </div>
                  <div className="flex gap-2 shrink-0">
                    <button onClick={() => openEdit(item)}
                      className="px-3 py-1.5 bg-violet-900/50 hover:bg-violet-800/50 text-violet-400 text-xs rounded-lg transition-colors">
                      Edit
                    </button>
                    <button onClick={() => handleDelete(item.ID)}
                      className="px-3 py-1.5 bg-red-900/50 hover:bg-red-800/50 text-red-400 text-xs rounded-lg transition-colors">
                      Delete
                    </button>
                  </div>
                </div>
              </Card>
            ))}
          </div>

          {items.length === 0 && (
            <div className="text-center py-16 text-gray-500">
              <p>No news yet.</p>
            </div>
          )}
        </>
      )}
    </div>
  );
}
