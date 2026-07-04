import Link from "next/link";

interface NovelCardSmallProps {
  rank: number;
  title: string;
  views: string;
  rating: string;
  href: string;
}

export default function NovelCardSmall({ rank, title, views, rating, href }: NovelCardSmallProps) {
  return (
    <Link href={href} className="flex items-start gap-3 p-2 rounded-lg hover:bg-[#1e1e3a] transition-colors group">
      <span className="text-lg font-bold text-gray-600 w-6 text-right shrink-0">#{rank}</span>
      <div className="min-w-0 flex-1">
        <p className="text-sm text-gray-200 group-hover:text-[#6dd5ed] transition-colors line-clamp-2 leading-snug">
          {title}
        </p>
        <div className="flex items-center gap-2 mt-1 text-xs text-gray-500">
          <span>{views} Views</span>
          <span>★{rating}</span>
        </div>
      </div>
    </Link>
  );
}
