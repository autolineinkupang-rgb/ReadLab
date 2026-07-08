"use client";

import { useEditor, EditorContent } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Underline from "@tiptap/extension-underline";
import Placeholder from "@tiptap/extension-placeholder";
import CharacterCount from "@tiptap/extension-character-count";
import { useEffect, useRef, useState, forwardRef, useImperativeHandle } from "react";
import DOMPurify from "isomorphic-dompurify";
import { cleanImportedHtml, txtToHtml, docxToHtml, ALLOWED_TAGS } from "@/lib/htmlImport";

export interface ChapterContentEditorHandle {
  importText: (text: string) => void;
}

interface ChapterContentEditorProps {
  value: string;
  onChange: (html: string) => void;
  placeholder?: string;
  onImportError?: (msg: string) => void;
}

const ChapterContentEditor = forwardRef<ChapterContentEditorHandle, ChapterContentEditorProps>(
  function ChapterContentEditor({ value, onChange, placeholder = "Tulis isi chapter...", onImportError }, ref) {
    const [isDraggingFile, setIsDraggingFile] = useState(false);
    const editorRef = useRef<HTMLDivElement>(null);

    const editor = useEditor({
      immediatelyRender: false,
      extensions: [
        StarterKit.configure({
          heading: { levels: [2, 3] },
          codeBlock: false,
          code: false,
        }),
        Underline,
        Placeholder.configure({ placeholder }),
        CharacterCount.configure({ limit: 50000 }),
      ],
      content: value || "",
      editorProps: {
        transformPastedHTML(html) {
          const cleaned = cleanImportedHtml(html);
          return DOMPurify.sanitize(cleaned, { ALLOWED_TAGS });
        },
        handleDrop(view, event, _slice, moved) {
          if (moved) return false;
          const file = event.dataTransfer?.files?.[0];
          if (!file) return false;
          event.preventDefault();

          const coords = view.posAtCoords({ left: event.clientX, top: event.clientY });
          const pos = coords ? coords.pos : view.state.selection.from;

          const insertHtml = (html: string) => {
            const clean = DOMPurify.sanitize(cleanImportedHtml(html), { ALLOWED_TAGS });
            const { tr } = view.state;
            const node = view.state.schema.text(clean);
            view.dispatch(tr.insert(pos, node));
          };

          if (file.name.toLowerCase().endsWith(".docx")) {
            docxToHtml(file).then(insertHtml).catch(() => onImportError?.("Gagal membaca file .docx"));
          } else if (file.name.toLowerCase().endsWith(".txt")) {
            file.text().then((text) => insertHtml(txtToHtml(text)));
          } else {
            onImportError?.("Format tidak didukung — gunakan .txt atau .docx");
          }
          return true;
        },
      },
      onUpdate: ({ editor }) => {
        onChange(editor.getHTML());
      },
    });

    useImperativeHandle(ref, () => ({
      importText(text: string) {
        if (!editor) return;
        editor.commands.setContent(text);
      },
    }));

    useEffect(() => {
      if (editor && value !== editor.getHTML()) {
        editor.commands.setContent(value || "");
      }
    }, [value, editor]);

    if (!editor) return null;

    return (
      <div
        className="relative"
        onDragEnter={(e) => { e.preventDefault(); setIsDraggingFile(true); }}
        onDragOver={(e) => e.preventDefault()}
        onDragLeave={(e) => { if (editorRef.current && !editorRef.current.contains(e.relatedTarget as Node)) setIsDraggingFile(false); }}
        onDrop={(e) => { setIsDraggingFile(false); }}
      >
        {isDraggingFile && (
          <div className="absolute inset-0 border-2 border-dashed border-accent rounded-lg bg-accent/10 flex items-center justify-center text-sm text-accent-light pointer-events-none z-10">
            Lepas file .txt atau .docx di sini
          </div>
        )}
        <div className="flex flex-wrap items-center gap-0.5 mb-2">
          <ToolbarButton
            active={editor.isActive("bold")}
            onClick={() => editor.chain().focus().toggleBold().run()}
            label="Bold"
          >
            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
              <path d="M15.6 10.79c.97-.67 1.65-1.77 1.65-2.79 0-2.26-1.75-4-4-4H7v14h7.04c2.09 0 3.71-1.7 3.71-3.79 0-1.52-.86-2.82-2.15-3.42zM10 6.5h3c.83 0 1.5.67 1.5 1.5s-.67 1.5-1.5 1.5h-3v-3zm3.5 9H10v-3h3.5c.83 0 1.5.67 1.5 1.5s-.67 1.5-1.5 1.5z" />
            </svg>
          </ToolbarButton>
          <ToolbarButton
            active={editor.isActive("italic")}
            onClick={() => editor.chain().focus().toggleItalic().run()}
            label="Italic"
          >
            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
              <path d="M10 4v3h2.21l-3.42 8H6v3h8v-3h-2.21l3.42-8H18V4z" />
            </svg>
          </ToolbarButton>
          <ToolbarButton
            active={editor.isActive("underline")}
            onClick={() => editor.chain().focus().toggleUnderline().run()}
            label="Underline"
          >
            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
              <path d="M6 3v7a6 6 0 0012 0V3M4 21h16" />
            </svg>
          </ToolbarButton>
          <ToolbarButton
            active={editor.isActive("strike")}
            onClick={() => editor.chain().focus().toggleStrike().run()}
            label="Strikethrough"
          >
            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
              <path d="M6 12h12M8 6c0-1.1.9-2 2-2h4a2 2 0 012 2M16 18a2 2 0 01-2 2h-4a2 2 0 01-2-2" />
            </svg>
          </ToolbarButton>
          <span className="w-px h-5 bg-line-light mx-1" />
          <select
            value={editor.isActive("heading", { level: 2 }) ? "h2" : editor.isActive("heading", { level: 3 }) ? "h3" : "p"}
            onChange={(e) => {
              const v = e.target.value;
              editor.chain().focus();
              if (v === "p") editor.chain().focus().setParagraph().run();
              else editor.chain().focus().toggleHeading({ level: parseInt(v) as 2 | 3 }).run();
            }}
            className="h-8 px-2 text-xs rounded bg-card-hover border border-line-light text-gray-300 outline-none cursor-pointer"
          >
            <option value="p">Normal</option>
            <option value="h2">Heading 2</option>
            <option value="h3">Heading 3</option>
          </select>
          <span className="w-px h-5 bg-line-light mx-1" />
          <ToolbarButton
            active={editor.isActive("bulletList")}
            onClick={() => editor.chain().focus().toggleBulletList().run()}
            label="Bullet List"
          >
            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
              <path d="M8 6h13M8 12h13M8 18h13M3 6h.01M3 12h.01M3 18h.01" />
            </svg>
          </ToolbarButton>
          <ToolbarButton
            active={editor.isActive("orderedList")}
            onClick={() => editor.chain().focus().toggleOrderedList().run()}
            label="Ordered List"
          >
            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
              <path d="M10 6h11M10 12h11M10 18h11M4 6h1v4M4 10h2M6 18H4m0 0l2-2.5a1.5 1.5 0 10-2 2.5" />
            </svg>
          </ToolbarButton>
          <ToolbarButton
            active={editor.isActive("blockquote")}
            onClick={() => editor.chain().focus().toggleBlockquote().run()}
            label="Blockquote"
          >
            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
              <path d="M6 17h3l2-4V7H5v6h3zm8 0h3l2-4V7h-6v6h3z" />
            </svg>
          </ToolbarButton>
          <ToolbarButton
            active={false}
            onClick={() => editor.chain().focus().setHorizontalRule().run()}
            label="Pemisah scene"
          >
            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
              <path d="M5 12h14" />
            </svg>
          </ToolbarButton>
          <span className="w-px h-5 bg-line-light mx-1" />
          <ToolbarButton
            active={false}
            onClick={() => editor.chain().focus().undo().run()}
            label="Undo"
          >
            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
              <path d="M3 10h10a5 5 0 015 5v2" /><path d="M7 6l-4 4 4 4" />
            </svg>
          </ToolbarButton>
          <ToolbarButton
            active={false}
            onClick={() => editor.chain().focus().redo().run()}
            label="Redo"
          >
            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
              <path d="M21 10H11a5 5 0 00-5 5v2" /><path d="M17 6l4 4-4 4" />
            </svg>
          </ToolbarButton>
        </div>
        <div
          ref={editorRef}
          className="bg-card-hover border border-line-light rounded-lg px-3 py-2 min-h-[300px] max-h-[600px] overflow-y-auto text-sm text-gray-200 cursor-text"
          onClick={() => editor.commands.focus()}
        >
          <EditorContent editor={editor} />
        </div>
        <div className="absolute bottom-3 right-3 text-[10px] text-gray-600 bg-card-hover px-2 py-0.5 rounded pointer-events-none">
          {editor.state.doc.textContent.trim()
            ? `${editor.state.doc.textContent.trim().split(/\s+/).length}w · ${editor.state.doc.textContent.length}c`
            : "0w · 0c"}
        </div>
      </div>
    );
  }
);

function ToolbarButton({
  active,
  onClick,
  label,
  children,
}: {
  active: boolean;
  onClick: () => void;
  label: string;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      title={label}
      className={`p-1.5 rounded transition-colors ${
        active
          ? "bg-accent/20 text-accent"
          : "text-gray-400 hover:text-white hover:bg-card-hover"
      }`}
    >
      {children}
    </button>
  );
}

export default ChapterContentEditor;
