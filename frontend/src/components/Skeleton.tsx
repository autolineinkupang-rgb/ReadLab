interface SkeletonProps {
  className?: string;
}

export function Skeleton({ className = "" }: SkeletonProps) {
  return (
    <div
      className={`animate-pulse bg-card-hover rounded ${className}`}
      aria-hidden="true"
      data-testid="skeleton"
    />
  );
}

export function NovelCardSkeleton({ compact }: { compact?: boolean }) {
  return (
    <div className={`${compact ? "w-36" : "w-44"} flex-shrink-0`} data-testid="novel-card-skeleton">
      <Skeleton className="aspect-[3/4] rounded-lg" />
      <div className="mt-2 space-y-2">
        <Skeleton className="h-3 w-3/4" />
        <Skeleton className="h-2 w-1/2" />
      </div>
    </div>
  );
}

export function NovelCardSkeletonRow({ count = 6, compact }: { count?: number; compact?: boolean }) {
  return (
    <div className="flex gap-4 overflow-hidden pb-2">
      {Array.from({ length: count }).map((_, i) => (
        <NovelCardSkeleton key={i} compact={compact} />
      ))}
    </div>
  );
}

export function UpdateItemSkeleton() {
  return (
    <div className="flex gap-3 py-2.5 border-b border-line last:border-0">
      <Skeleton className="w-10 h-14 rounded" />
      <div className="flex-1 space-y-2 py-1">
        <Skeleton className="h-3 w-3/4" />
        <Skeleton className="h-2 w-1/2" />
      </div>
    </div>
  );
}
