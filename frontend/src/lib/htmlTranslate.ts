const MAX_BATCH_LENGTH = 1800;

export async function translateHtmlPreservingStructure(
  html: string,
  translateFn: (text: string) => Promise<string>
): Promise<string> {
  const doc = new DOMParser().parseFromString(html, "text/html");
  const blocks = Array.from(doc.body.children);

  if (blocks.length === 0 && doc.body.childNodes.length > 0) {
    for (const node of Array.from(doc.body.childNodes)) {
      if (node.nodeType === Node.TEXT_NODE) {
        const original = node.textContent || "";
        if (!original.trim()) continue;
        node.textContent = await translateFn(original);
      }
    }
    return doc.body.innerHTML;
  }

  function getBlockKey(block: Element): string {
    const tag = block.tagName.toLowerCase();
    if (tag === "ul" || tag === "ol") return tag;
    return "text";
  }

  type Batch = { blocks: Element[]; text: string; isList: boolean };
  const batches: Batch[] = [];
  let current: Batch | null = null;

  for (const block of blocks) {
    const tag = block.tagName.toLowerCase();
    if (tag === "hr") continue;

    if (tag === "ul" || tag === "ol") {
      const items = Array.from(block.querySelectorAll("li"));
      const listText = items.map((li) => li.textContent || "").filter(Boolean);
      if (listText.length === 0) continue;
      const combined = listText.join("\n• ");
      if (current && !current.isList && current.text.length + combined.length + 2 < MAX_BATCH_LENGTH) {
        current.text += "\n\n" + combined;
        current.blocks.push(block);
      } else {
        if (current) batches.push(current);
        current = { blocks: [block], text: combined, isList: true };
      }
      continue;
    }

    const original = (block.textContent || "").trim();
    if (!original) continue;

    if (current && !current.isList && current.text.length + original.length + 2 < MAX_BATCH_LENGTH) {
      current.text += "\n\n" + original;
      current.blocks.push(block);
    } else {
      if (current) batches.push(current);
      current = { blocks: [block], text: original, isList: false };
    }
  }
  if (current) batches.push(current);

  const translatedTexts = await Promise.all(
    batches.map((batch) => translateFn(batch.text).then((t) => ({ batch, result: t })))
  );

  for (const { batch, result } of translatedTexts) {
    if (batch.isList) {
      const lines = result.split("\n").map((s) => s.trim()).filter(Boolean);
      const items = Array.from(batch.blocks[0].querySelectorAll("li"));
      items.forEach((li, i) => {
        if (i < lines.length) li.textContent = lines[i].replace(/^•\s*/, "");
      });
    } else if (batch.blocks.length === 1) {
      batch.blocks[0].textContent = result;
    } else {
      const lines = result.split("\n").map((s) => s.trim()).filter(Boolean);
      batch.blocks.forEach((block, i) => {
        block.textContent = lines[i] || "";
      });
    }
  }

  return doc.body.innerHTML;
}
