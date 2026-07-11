import Link from "next/link";

interface UpdateItemProps {
  title: string;
  chapter: string;
  chapterHref: string;
  novelHref: string;
  image?: string;
  hasImage?: boolean;
  timeAgo?: string;
}

export default function UpdateItem({ title, chapter, chapterHref, novelHref, image, hasImage, timeAgo }: UpdateItemProps) {
  const hasValidImage = !!image && /^https?:\/\//.test(image);
  const showThumb = hasImage || hasValidImage;
  return (
    <div className="flex gap-3 py-2.5 border-b border-line last:border-0">
      {showThumb && (
        <Link href={novelHref} className="w-10 h-14 rounded bg-card-hover flex-shrink-0 flex items-center justify-center overflow-hidden border border-line-light hover:border-accent transition-colors">
          {hasValidImage ? (
            <img
              src={image}
              alt={title}
              loading="lazy"
              className="w-full h-full object-cover"
              onError={(e) => { (e.currentTarget as HTMLImageElement).style.display = "none"; }}
            />
          ) : (
            <svg className="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
            </svg>
          )}
        </Link>
      )}
      <div className="min-w-0 flex-1">
        <Link href={novelHref} className="text-sm text-gray-200 hover:text-accent-light transition-colors line-clamp-1 font-medium">
          {title}
        </Link>
        <div className="flex items-center gap-2 mt-0.5">
          <Link href={chapterHref} className="text-xs text-accent hover:text-accent-light transition-colors line-clamp-1">
            {chapter}
          </Link>
          {timeAgo && <span className="text-[10px] text-gray-600 shrink-0">• {timeAgo}</span>}
        </div>
      </div>
    </div>
  );
}
