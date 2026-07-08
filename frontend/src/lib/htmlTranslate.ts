export async function translateHtmlPreservingStructure(
  html: string,
  translateFn: (text: string) => Promise<string>
): Promise<string> {
  const doc = new DOMParser().parseFromString(html, "text/html");
  const blocks = Array.from(doc.body.children);

  for (const block of blocks) {
    const tag = block.tagName.toLowerCase();

    if (tag === "hr") continue;

    if (tag === "ul" || tag === "ol") {
      const items = Array.from(block.querySelectorAll("li"));
      for (const li of items) {
        const original = li.textContent || "";
        if (!original.trim()) continue;
        li.textContent = await translateFn(original);
      }
      continue;
    }

    const original = block.textContent || "";
    if (!original.trim()) continue;
    const translated = await translateFn(original);
    block.textContent = translated;
  }

  return doc.body.innerHTML;
}
