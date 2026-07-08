"use client";

export interface ProgressStep {
  label: string;
  status: "pending" | "active" | "done";
}

interface ImportProgressProps {
  steps: ProgressStep[];
  className?: string;
}

export default function ImportProgress({ steps, className = "" }: ImportProgressProps) {
  const doneCount = steps.filter((s) => s.status === "done").length;
  const activeCount = steps.filter((s) => s.status === "active").length;
  const totalCount = steps.length;
  const pct = totalCount > 0 ? (doneCount / totalCount) * 100 : 0;

  return (
    <div className={`space-y-3 ${className}`}>
      <div className="h-1.5 bg-card-hover rounded-full overflow-hidden">
        <div
          className="h-full rounded-full transition-all duration-700 ease-out"
          style={{
            width: `${Math.max(4, pct)}%`,
            background: activeCount > 0
              ? "linear-gradient(90deg, var(--color-accent), var(--color-accent-light))"
              : "var(--color-accent)",
          }}
        />
      </div>
      <div className="space-y-1.5">
        {steps.map((step, i) => (
          <div key={i} className="flex items-center gap-2.5 text-xs">
            {step.status === "done" ? (
              <svg className="w-4 h-4 text-green-400 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M5 13l4 4L19 7" />
              </svg>
            ) : step.status === "active" ? (
              <svg className="w-4 h-4 text-accent shrink-0" viewBox="0 0 24 24" fill="none">
                <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="3" className="opacity-25" />
                <path d="M12 2a10 10 0 019.95 9" stroke="currentColor" strokeWidth="3" strokeLinecap="round" className="animate-spin" style={{ transformOrigin: "12px 12px" }} />
              </svg>
            ) : (
              <div className="w-4 h-4 rounded-full border border-gray-600 shrink-0 flex items-center justify-center">
                <div className="w-1.5 h-1.5 rounded-full bg-gray-700" />
              </div>
            )}
            <span
              className={
                step.status === "active"
                  ? "text-accent-light animate-pulse"
                  : step.status === "done"
                    ? "text-green-400"
                    : "text-gray-500"
              }
            >
              {step.label}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
