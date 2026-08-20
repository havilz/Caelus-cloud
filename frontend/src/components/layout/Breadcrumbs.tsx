"use client";

import React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { ChevronRight, Home } from "lucide-react";

export const Breadcrumbs: React.FC = () => {
  const pathname = usePathname();
  const segments = pathname.split("/").filter(Boolean);

  const formatSegment = (seg: string) => {
    return seg
      .replaceAll("-", " ")
      .replace(/\b\w/g, (c) => c.toUpperCase());
  };

  return (
    <nav className="flex items-center gap-1.5 text-xs text-slate-400">
      <Link
        href="/overview"
        className="flex items-center gap-1 hover:text-slate-200 transition-colors"
      >
        <Home className="h-3.5 w-3.5" />
      </Link>

      {segments.map((segment, index) => {
        const url = `/${segments.slice(0, index + 1).join("/")}`;
        const isLast = index === segments.length - 1;

        return (
          <React.Fragment key={url}>
            <ChevronRight className="h-3 w-3 text-slate-600" />
            {isLast ? (
              <span className="font-medium text-slate-200">{formatSegment(segment)}</span>
            ) : (
              <Link href={url} className="hover:text-slate-200 transition-colors">
                {formatSegment(segment)}
              </Link>
            )}
          </React.Fragment>
        );
      })}
    </nav>
  );
};
