export const ALLOWED_TAGS = ["p", "h2", "h3", "strong", "em", "u", "s", "ul", "ol", "li", "blockquote", "hr", "br"];

function mapTag(tagName: string): string | null {
  switch (tagName.toLowerCase()) {
    case "b": return "strong";
    case "i": return "em";
    case "strike": return "s";
    case "strong": case "em": case "u": case "s":
    case "p": case "ul": case "ol": case "li":
    case "blockquote": case "hr": case "br":
      return tagName.toLowerCase();
    case "div": return "p";
    case "h1": case "h2": return "h2";
    case "h3": case "h4": case "h5": case "h6": return "h3";
    default: return null;
  }
}

function styleImpliesTags(el: HTMLElement): string[] {
  const tags: string[] = [];
  const fw = el.style.fontWeight;
  if (fw === "bold" || fw === "700" || parseInt(fw) >= 700) tags.push("strong");
  if (el.style.fontStyle === "italic") tags.push("em");
  const td = el.style.textDecorationLine || el.style.textDecoration || "";
  if (td.includes("underline")) tags.push("u");
  if (td.includes("line-through")) tags.push("s");
  return tags;
}

export function cleanImportedHtml(rawHtml: string): string {
  const doc = new DOMParser().parseFromString(rawHtml, "text/html");

  function walk(node: Node): Node[] {
    if (node.nodeType === Node.TEXT_NODE) return [node.cloneNode()];
    if (node.nodeType !== Node.ELEMENT_NODE) return [];

    const el = node as HTMLElement;
    let children: Node[] = [];
    el.childNodes.forEach((c) => children.push(...walk(c)));

    for (const tag of styleImpliesTags(el)) {
      const wrapper = doc.createElement(tag);
      children.forEach((c) => wrapper.appendChild(c));
      children = [wrapper];
    }

    const mapped = mapTag(el.tagName);
    if (!mapped) return children;

    const newEl = doc.createElement(mapped);
    children.forEach((c) => newEl.appendChild(c));
    return [newEl];
  }

  const container = doc.createElement("div");
  doc.body.childNodes.forEach((n) => walk(n).forEach((x) => container.appendChild(x)));

  const blockTags = new Set(["P", "H2", "H3", "UL", "OL", "BLOCKQUOTE", "HR"]);
  const final = document.createElement("div");
  let currentP: HTMLElement | null = null;
  Array.from(container.childNodes).forEach((node) => {
    const isBlock = node.nodeType === Node.ELEMENT_NODE && blockTags.has((node as HTMLElement).tagName);
    if (isBlock) { currentP = null; final.appendChild(node); }
    else {
      if (!currentP) { currentP = document.createElement("p"); final.appendChild(currentP); }
      currentP.appendChild(node);
    }
  });
  final.querySelectorAll("p").forEach((p) => {
    if (!p.textContent?.trim() && !p.querySelector("br")) p.remove();
  });

  return final.innerHTML;
}

export function txtToHtml(text: string): string {
  return text
    .split(/\n\s*\n/)
    .map((p) => `<p>${p.trim().replace(/\n/g, "<br>")}</p>`)
    .filter((p) => p !== "<p></p>")
    .join("");
}

export async function docxToHtml(file: File): Promise<string> {
  const mammoth = await import("mammoth");
  const arrayBuffer = await file.arrayBuffer();
  const result = await mammoth.convertToHtml({ arrayBuffer });
  return cleanImportedHtml(result.value);
}
