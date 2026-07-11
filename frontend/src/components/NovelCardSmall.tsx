import Link from "next/link";

interface NovelCardSmallProps {
  rank: number;
  title: string;
  views: string;
  rating: string;
  href: string;
  image?: string;
}

export default function NovelCardSmall({ rank, title, views, rating, href, image }: NovelCardSmallProps) {
  const rankColor = rank === 1 ? "text-yellow-400" : rank === 2 ? "text-gray-300" : rank === 3 ? "text-orange-400" : "text-gray-600";
  const hasValidImage = !!image && /^(https?:\/\/|\/api\/)/.test(image);
  return (
    <Link
      href={href}
      className="flex items-start gap-3 p-2 rounded-lg hover:bg-card-hover transition-colors group focus:outline-none focus-visible:ring-2 focus-visible:ring-accent"
      data-testid={`novel-small-rank-${rank}`}
    >
      <span className={`text-lg font-bold w-6 text-right shrink-0 ${rankColor}`}>#{rank}</span>
      {hasValidImage ? (
        <img
          src={image}
          alt={title}
          loading="lazy"
          className="w-10 h-14 rounded object-cover flex-shrink-0 border border-line-light"
          onError={(e) => { (e.currentTarget as HTMLImageElement).style.display = "none"; }}
        />
      ) : null}
      <div className="min-w-0 flex-1">
        <p className="text-sm text-gray-200 group-hover:text-accent-light transition-colors line-clamp-2 leading-snug">
          {title}
        </p>
        <div className="flex items-center gap-2 mt-1 text-xs text-gray-500">
          <span>{views} Views</span>
          <span className="text-yellow-500">★{rating}</span>
        </div>
      </div>
    </Link>
  );
}
