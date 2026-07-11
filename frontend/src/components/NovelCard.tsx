import Link from "next/link";

interface NovelCardProps {
  title: string;
  genre: string;
  chapters: number;
  rating?: string;
  image?: string;
  href: string;
  compact?: boolean;
  progress?: number; // 0-100 reading progress percentage
}

export default function NovelCard({ title, genre, chapters, rating, image, href, compact, progress }: NovelCardProps) {
  return (
    <Link
      href={href}
      className={`group block ${compact ? "w-36" : "w-44"} flex-shrink-0 focus:outline-none focus-visible:ring-2 focus-visible:ring-accent rounded-lg`}
      data-testid={`novel-card-${title.substring(0, 20)}`}
    >
      <div className="relative aspect-[3/4] rounded-lg overflow-hidden bg-card-hover border border-line-light card-hover transition-all duration-300">
        {image ? (
          <img
            src={image}
            alt={title}
            loading="lazy"
            className="absolute inset-0 w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
            onError={(e) => { (e.currentTarget as HTMLImageElement).style.display = "none"; }}
          />
        ) : (
          <div className="absolute inset-0 flex items-center justify-center text-gray-600 bg-gradient-to-br from-card-hover to-line-light">
            <svg className="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
            </svg>
          </div>
        )}
        {rating && (
          <span className="absolute top-2 right-2 bg-black/70 backdrop-blur-sm text-yellow-400 text-xs px-1.5 py-0.5 rounded flex items-center gap-0.5 z-10">
            <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
              <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
            </svg>
            {rating}
          </span>
        )}
        {typeof progress === "number" && progress > 0 && (
          <div className="absolute bottom-0 left-0 right-0 h-1 bg-black/40 z-10">
            <div className="h-full bg-accent" style={{ width: `${Math.min(100, progress)}%` }} />
          </div>
        )}
        <div className="absolute inset-x-0 bottom-0 h-16 bg-gradient-to-t from-black/70 to-transparent pointer-events-none opacity-0 group-hover:opacity-100 transition-opacity" />
      </div>
      <div className="mt-2 space-y-0.5">
        <h3 className="text-sm font-medium text-gray-200 group-hover:text-accent-light transition-colors line-clamp-2 leading-tight">
          <span>{title}</span>
        </h3>
        <div className="flex items-center gap-2 text-xs text-gray-500">
          <span className="capitalize">{genre}</span>
          <span>•</span>
          <span>{chapters} Ch</span>
        </div>
      </div>
    </Link>
  );
}
