"use client";

import React from "react";
import { LucideIcon } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";

interface MetricCardProps {
  title: string;
  value: string | number;
  unit?: string;
  subtitle?: string;
  icon: LucideIcon;
  percentage?: number;
  color?: "emerald" | "blue" | "amber" | "rose" | "purple";
}

export const MetricCard: React.FC<MetricCardProps> = ({
  title,
  value,
  unit,
  subtitle,
  icon: Icon,
  percentage,
  color = "emerald",
}) => {
  const getColorClasses = () => {
    switch (color) {
      case "blue":
        return {
          icon: "text-blue-400 bg-blue-500/10 border-blue-500/20",
          bar: "bg-blue-500",
          text: "text-blue-400",
        };
      case "amber":
        return {
          icon: "text-amber-400 bg-amber-500/10 border-amber-500/20",
          bar: "bg-amber-500",
          text: "text-amber-400",
        };
      case "rose":
        return {
          icon: "text-rose-400 bg-rose-500/10 border-rose-500/20",
          bar: "bg-rose-500",
          text: "text-rose-400",
        };
      case "purple":
        return {
          icon: "text-purple-400 bg-purple-500/10 border-purple-500/20",
          bar: "bg-purple-500",
          text: "text-purple-400",
        };
      default:
        return {
          icon: "text-emerald-400 bg-emerald-500/10 border-emerald-500/20",
          bar: "bg-emerald-500",
          text: "text-emerald-400",
        };
    }
  };

  const colorStyles = getColorClasses();

  return (
    <Card className="hover:border-[#383838] transition-all">
      <CardContent className="p-4">
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium text-[#a1a1a1]">{title}</span>
          <div className={`flex h-7 w-7 items-center justify-center rounded-lg border ${colorStyles.icon}`}>
            <Icon className="h-4 w-4" />
          </div>
        </div>

        <div className="mt-2 flex items-baseline gap-1">
          <span className="text-2xl font-bold tracking-tight text-[#ededed] font-mono">
            {value}
          </span>
          {unit && <span className="text-xs text-[#a1a1a1]">{unit}</span>}
        </div>

        {percentage !== undefined && (
          <div className="mt-3">
            <div className="flex justify-between text-[11px] text-[#707070] mb-1 font-mono">
              <span>Utilisasi</span>
              <span className={colorStyles.text}>{percentage.toFixed(1)}%</span>
            </div>
            <div className="h-1.5 w-full rounded-full bg-[#202020] overflow-hidden">
              <div
                className={`h-full rounded-full ${colorStyles.bar} transition-all duration-500`}
                style={{ width: `${Math.min(Math.max(percentage, 0), 100)}%` }}
              />
            </div>
          </div>
        )}

        {subtitle && (
          <p className="mt-2 text-[11px] text-[#707070] truncate">{subtitle}</p>
        )}
      </CardContent>
    </Card>
  );
};
