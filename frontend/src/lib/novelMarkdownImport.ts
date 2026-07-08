import { cleanImportedHtml } from "./htmlImport";

export interface ParsedChapter {
  number: number;
  title: string;
  content: string;
}

export interface ParsedNovel {
  title: string;
  chapters: ParsedChapter[];
}

function inlineMd(text: string): string {
  let s = text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");
  return s
    .replace(/\*\*\*(.+?)\*\*\*/g, "<strong><em>$1</em></strong>")
    .replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>")
    .replace(/\*(.+?)\*/g, "<em>$1</em>")
    .replace(/~~(.+?)~~/g, "<s>$1</s>")
    .replace(/<u>(.+?)<\/u>/g, "<u>$1</u>");
}

function parseMarkdownContent(md: string): string {
  const lines = md.split("\n");
  const out: string[] = [];

  let i = 0;
  while (i < lines.length) {
    const line = lines[i].trim();

    if (!line) { i++; continue; }

    const h3 = line.match(/^### (.+)/);
    if (h3) { out.push(`<h3>${inlineMd(h3[1])}</h3>`); i++; continue; }

    if (/^---+$/.test(line)) { out.push("<hr>"); i++; continue; }

    const bq = line.match(/^> (.+)/);
    if (bq) { out.push(`<blockquote>${inlineMd(bq[1])}</blockquote>`); i++; continue; }

    if (/^[-*] /.test(line)) {
      out.push("<ul>");
      while (i < lines.length && /^[-*] /.test(lines[i].trim())) {
        out.push(`<li>${inlineMd(lines[i].trim().replace(/^[-*] /, ""))}</li>`);
        i++;
      }
      out.push("</ul>");
      continue;
    }

    if (/^\d+\. /.test(line)) {
      out.push("<ol>");
      while (i < lines.length && /^\d+\. /.test(lines[i].trim())) {
        out.push(`<li>${inlineMd(lines[i].trim().replace(/^\d+\. /, ""))}</li>`);
        i++;
      }
      out.push("</ol>");
      continue;
    }

    const paraLines: string[] = [];
    while (i < lines.length) {
      const l = lines[i].trim();
      if (!l || /^(### |[-*] |\d+\. |> |---+$)/.test(l)) break;
      paraLines.push(lines[i]);
      i++;
    }
    if (paraLines.length > 0) {
      out.push(`<p>${inlineMd(paraLines.join("<br>"))}</p>`);
    }
  }

  return out.join("\n");
}

export function parseMarkdownNovel(markdown: string, filename?: string): ParsedNovel {
  const lines = markdown.split("\n");
  let title = filename?.replace(/\.md$/i, "") || "Untitled";
  const chapters: ParsedChapter[] = [];

  let currentChapterTitle = "";
  let currentChapterContent: string[] = [];
  let inHeader = true;

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmed = line.trim();

    const h1Match = trimmed.match(/^# (.+)/);
    if (h1Match && inHeader) {
      title = h1Match[1].trim();
      continue;
    }

    const h2Match = trimmed.match(/^## (.+)/);
    if (h2Match) {
      inHeader = false;
      flushChapter();
      currentChapterTitle = h2Match[1].trim();
      continue;
    }

    if (!inHeader) {
      currentChapterContent.push(line);
    }
  }

  flushChapter();

  function flushChapter() {
    const content = currentChapterContent.join("\n").trim();
    const chTitle = currentChapterTitle || `Chapter ${chapters.length + 1}`;
    if (content) {
      const html = parseMarkdownContent(content);
      const cleanHtml = cleanImportedHtml(html);
      if (cleanHtml) {
        chapters.push({ number: chapters.length + 1, title: chTitle, content: cleanHtml });
      }
    }
    currentChapterTitle = "";
    currentChapterContent = [];
  }

  if (chapters.length === 0) {
    const bodyContent = markdown.replace(/^# .+$/m, "").trim();
    if (bodyContent) {
      const html = parseMarkdownContent(bodyContent);
      const cleanHtml = cleanImportedHtml(html);
      if (cleanHtml) {
        chapters.push({ number: 1, title: "Chapter 1", content: cleanHtml });
      }
    }
  }

  return { title, chapters };
}
