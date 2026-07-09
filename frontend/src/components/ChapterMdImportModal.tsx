"use client";

import { useState, useRef } from "react";
import { chapters } from "@/lib/api";

interface ParsedChapter {
  number: number;
  title: string;
  content_md: string;
  content_html: string;
  exists: boolean;
}

interface Props {
  novelId: number | string;
  onClose: () => void;
  onImported: () => void;
}

export default function ChapterMdImportModal({ novelId, onClose, onImported }: Props) {
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<ParsedChapter[] | null>(null);
  const [selected, setSelected] = useState<Set<number>>(new Set());
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [result, setResult] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0];
    if (!f) return;
    setFile(f);
    setPreview(null);
    setResult("");
    setError("");
  };

  const handleUpload = async (mode: "preview" | "save") => {
    if (!file) return;
    setLoading(true);
    setError("");
    try {
      const res = await chapters.importMd(novelId, file, mode);
      if (mode === "preview") {
        const data = res as { chapters: ParsedChapter[]; warnings?: string[] };
        setPreview(data.chapters);
        setSelected(new Set(data.chapters.map((c: ParsedChapter) => c.number)));
        if (data.warnings?.length) setError(data.warnings.join("\n"));
      } else {
        const data = res as { message: string; imported: number; skipped: number };
        setResult(data.message);
        setPreview(null);
        setFile(null);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Upload failed");
    } finally {
      setLoading(false);
    }
  };

  const toggleSelect = (num: number) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(num)) next.delete(num);
      else next.add(num);
      return next;
    });
  };

  const handleSave = async () => {
    await handleUpload("save");
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60">
      <div className="bg-card border border-line-light rounded-2xl w-full max-w-3xl max-h-[90vh] overflow-y-auto mx-4">
        <div className="px-6 py-4 border-b border-line-light flex items-center justify-between">
          <h2 className="text-lg font-bold text-white">Import Markdown</h2>
          <button onClick={onClose} className="text-gray-500 hover:text-white transition-colors text-xl">&times;</button>
        </div>

        <div className="px-6 py-4 space-y-4">
          {result ? (
            <div className="space-y-4">
              <div className="bg-emerald-900/20 border border-emerald-700/30 text-emerald-400 text-sm rounded-lg px-4 py-3">
                {result}
              </div>
              <button onClick={onImported} className="px-4 py-2 bg-violet-600 hover:bg-violet-700 text-white text-sm rounded-lg transition-colors">
                Done
              </button>
            </div>
          ) : (
            <>
              <div className="border-2 border-dashed border-line-light rounded-xl p-8 text-center">
                <input
                  ref={inputRef}
                  type="file"
                  accept=".md,.markdown,.zip"
                  onChange={handleFileChange}
                  className="hidden"
                />
                {file ? (
                  <div className="space-y-2">
                    <p className="text-sm text-gray-200">{file.name} ({(file.size / 1024).toFixed(1)} KB)</p>
                    <button
                      onClick={() => inputRef.current?.click()}
                      className="text-xs text-violet-400 hover:text-violet-300 transition-colors"
                    >
                      Choose different file
                    </button>
                  </div>
                ) : (
                  <button
                    onClick={() => inputRef.current?.click()}
                    className="text-sm text-gray-400 hover:text-white transition-colors"
                  >
                    Click to select .md or .zip file
                  </button>
                )}
              </div>

              {error && (
                <div className="bg-red-900/20 border border-red-700/30 text-red-400 text-sm rounded-lg px-4 py-3 whitespace-pre-wrap">
                  {error}
                </div>
              )}

              {file && !preview && (
                <button
                  onClick={() => handleUpload("preview")}
                  disabled={loading}
                  className="w-full py-2.5 bg-violet-600 hover:bg-violet-700 disabled:opacity-50 text-white text-sm font-medium rounded-lg transition-colors"
                >
                  {loading ? "Parsing..." : "Preview"}
                </button>
              )}

              {preview && (
                <div className="space-y-3">
                  <p className="text-sm text-gray-400">{preview.length} chapter(s) found</p>

                  <div className="max-h-64 overflow-y-auto border border-line-light rounded-lg divide-y divide-line-light">
                    {preview.map((ch) => (
                      <label
                        key={ch.number}
                        className={`flex items-center gap-3 px-4 py-2.5 text-sm cursor-pointer transition-colors ${
                          ch.exists ? "bg-gray-800/50 opacity-60" : "hover:bg-card-hover"
                        }`}
                      >
                        <input
                          type="checkbox"
                          checked={selected.has(ch.number)}
                          onChange={() => toggleSelect(ch.number)}
                          disabled={ch.exists}
                          className="accent-violet-500"
                        />
                        <span className="text-gray-500 w-8 shrink-0">#{ch.number}</span>
                        <span className="text-gray-200 truncate">{ch.title || "(untitled)"}</span>
                        {ch.exists && <span className="text-[10px] text-gray-600 ml-auto">exists</span>}
                      </label>
                    ))}
                  </div>

                  <div className="flex gap-3">
                    <button
                      onClick={() => setPreview(null)}
                      disabled={loading}
                      className="flex-1 py-2.5 border border-line-light text-gray-300 hover:text-white text-sm rounded-lg transition-colors"
                    >
                      Cancel
                    </button>
                    <button
                      onClick={handleSave}
                      disabled={loading || selected.size === 0}
                      className="flex-1 py-2.5 bg-violet-600 hover:bg-violet-700 disabled:opacity-50 text-white text-sm font-medium rounded-lg transition-colors"
                    >
                      {loading ? "Importing..." : `Import ${selected.size} Chapter(s)`}
                    </button>
                  </div>
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}
