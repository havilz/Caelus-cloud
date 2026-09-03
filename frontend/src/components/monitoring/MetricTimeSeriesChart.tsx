"use client";

import React, { useState, useMemo, useRef } from "react";
import { Activity, Clock } from "lucide-react";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { ServerMetric } from "@/types/monitoring";

interface MetricTimeSeriesChartProps {
  title: string;
  description?: string;
  data: ServerMetric[];
  dataKey: "cpu_usage_pct" | "memory_usage_pct" | "disk_usage_pct" | "network_in_rate_kbps";
  unit?: string;
  color?: string;
  selectedRange: string;
  onRangeChange: (range: string) => void;
  isLive?: boolean;
}

export const MetricTimeSeriesChart: React.FC<MetricTimeSeriesChartProps> = ({
  title,
  description,
  data,
  dataKey,
  unit = "%",
  color = "#10b981", 
  selectedRange,
  onRangeChange,
  isLive = false,
}) => {
  const [hoverIndex, setHoverIndex] = useState<number | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  const points = useMemo(() => {
    if (!data || data.length === 0) return [];
    return data.map((d, index) => ({
      index,
      value: Number(d[dataKey]) || 0,
      timestamp: new Date(d.recorded_at),
    }));
  }, [data, dataKey]);

  const stats = useMemo(() => {
    if (points.length === 0) {
      return { current: 0, min: 0, max: 0, avg: 0 };
    }
    const values = points.map((p) => p.value);
    const current = values[values.length - 1];
    const min = Math.min(...values);
    const max = Math.max(...values);
    const avg = values.reduce((acc, curr) => acc + curr, 0) / values.length;
    return { current, min, max, avg };
  }, [points]);

  const width = 600;
  const height = 200;
  const padding = { top: 20, right: 15, bottom: 25, left: 35 };

  const chartCoordinates = useMemo(() => {
    if (points.length < 2) return [];

    const effectiveMax = Math.max(stats.max * 1.15, unit === "%" ? 100 : 10);
    const effectiveMin = 0;

    const innerWidth = width - padding.left - padding.right;
    const innerHeight = height - padding.top - padding.bottom;

    return points.map((p, i) => {
      const x = padding.left + (i / (points.length - 1)) * innerWidth;
      const y =
        padding.top +
        innerHeight -
        ((p.value - effectiveMin) / (effectiveMax - effectiveMin)) * innerHeight;
      return { x, y, value: p.value, timestamp: p.timestamp };
    });
  }, [points, stats.max, unit]);

  const pathD = useMemo(() => {
    if (chartCoordinates.length < 2) return "";
    let d = `M ${chartCoordinates[0].x} ${chartCoordinates[0].y}`;
    for (let i = 0; i < chartCoordinates.length - 1; i++) {
      const current = chartCoordinates[i];
      const next = chartCoordinates[i + 1];
      const controlX = (current.x + next.x) / 2;
      d += ` C ${controlX} ${current.y}, ${controlX} ${next.y}, ${next.x} ${next.y}`;
    }
    return d;
  }, [chartCoordinates]);

  const areaD = useMemo(() => {
    if (chartCoordinates.length < 2) return "";
    const bottomY = height - padding.bottom;
    const firstX = chartCoordinates[0].x;
    const lastX = chartCoordinates[chartCoordinates.length - 1].x;
    return `${pathD} L ${lastX} ${bottomY} L ${firstX} ${bottomY} Z`;
  }, [pathD, chartCoordinates]);

  const handleMouseMove = (e: React.MouseEvent<SVGSVGElement>) => {
    if (!containerRef.current || chartCoordinates.length === 0) return;
    const rect = e.currentTarget.getBoundingClientRect();
    const mouseX = e.clientX - rect.left;
    const innerWidth = width - padding.left - padding.right;

    const relativeX = (mouseX / rect.width) * width;
    const clampedX = Math.max(padding.left, Math.min(width - padding.right, relativeX));
    const ratio = (clampedX - padding.left) / innerWidth;
    const closestIdx = Math.round(ratio * (chartCoordinates.length - 1));

    setHoverIndex(closestIdx);
  };

  const handleMouseLeave = () => {
    setHoverIndex(null);
  };

  const hoveredPoint = hoverIndex !== null ? chartCoordinates[hoverIndex] : null;

  return (
    <Card className="border-[#262626] bg-[#121212]">
      <CardHeader className="pb-3 border-b border-[#1f1f1f]">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
          <div>
            <div className="flex items-center gap-2">
              <Activity className="h-4 w-4" style={{ color }} />
              <CardTitle className="text-xs font-semibold text-[#ededed]">{title}</CardTitle>
              {isLive && (
                <span className="flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[10px] font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  <span className="relative flex h-1.5 w-1.5">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
                    <span className="relative inline-flex rounded-full h-1.5 w-1.5 bg-emerald-500" />
                  </span>
                  LIVE
                </span>
              )}
            </div>
            {description && <CardDescription className="text-[11px] mt-0.5">{description}</CardDescription>}
          </div>

          {}
          <div className="flex items-center gap-1 rounded-lg bg-[#171717] p-1 border border-[#262626]">
            {["1h", "6h", "24h", "7d"].map((range) => (
              <button
                key={range}
                type="button"
                onClick={() => onRangeChange(range)}
                className={`px-2.5 py-1 text-[11px] font-medium rounded-md transition-all cursor-pointer ${
                  selectedRange === range
                    ? "bg-[#262626] text-emerald-400 shadow-sm"
                    : "text-[#a1a1a1] hover:text-[#ededed]"
                }`}
              >
                {range}
              </button>
            ))}
          </div>
        </div>

        {}
        <div className="grid grid-cols-4 gap-2 pt-3 border-t border-[#1a1a1a] mt-2">
          <div>
            <p className="text-[10px] text-[#707070] uppercase tracking-wider font-mono">Current</p>
            <p className="text-xs font-bold text-[#ededed] font-mono mt-0.5">
              {stats.current.toFixed(1)} {unit}
            </p>
          </div>
          <div>
            <p className="text-[10px] text-[#707070] uppercase tracking-wider font-mono">Average</p>
            <p className="text-xs font-semibold text-[#a1a1a1] font-mono mt-0.5">
              {stats.avg.toFixed(1)} {unit}
            </p>
          </div>
          <div>
            <p className="text-[10px] text-[#707070] uppercase tracking-wider font-mono">Min</p>
            <p className="text-xs font-semibold text-[#a1a1a1] font-mono mt-0.5">
              {stats.min.toFixed(1)} {unit}
            </p>
          </div>
          <div>
            <p className="text-[10px] text-[#707070] uppercase tracking-wider font-mono">Max Peak</p>
            <p className="text-xs font-semibold text-rose-400 font-mono mt-0.5">
              {stats.max.toFixed(1)} {unit}
            </p>
          </div>
        </div>
      </CardHeader>

      <CardContent className="p-4">
        {chartCoordinates.length < 2 ? (
          <div className="flex h-48 items-center justify-center text-xs text-[#707070] flex-col gap-2">
            <Clock className="h-5 w-5 opacity-40" />
            <span>Mengumpulkan data deret waktu telemetri...</span>
          </div>
        ) : (
          <div className="relative w-full overflow-hidden" ref={containerRef}>
            <svg
              viewBox={`0 0 ${width} ${height}`}
              className="w-full h-48 overflow-visible select-none cursor-crosshair"
              onMouseMove={handleMouseMove}
              onMouseLeave={handleMouseLeave}
            >
              <defs>
                <linearGradient id={`gradient-${dataKey}`} x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor={color} stopOpacity="0.35" />
                  <stop offset="100%" stopColor={color} stopOpacity="0.0" />
                </linearGradient>
              </defs>

              {}
              {[0, 0.25, 0.5, 0.75, 1].map((ratio) => {
                const y = padding.top + (height - padding.top - padding.bottom) * ratio;
                return (
                  <line
                    key={ratio}
                    x1={padding.left}
                    y1={y}
                    x2={width - padding.right}
                    y2={y}
                    stroke="#222222"
                    strokeDasharray="3 3"
                    strokeWidth="1"
                  />
                );
              })}

              {}
              <path d={areaD} fill={`url(#gradient-${dataKey})`} />

              {}
              <path d={pathD} fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" />

              {}
              {hoveredPoint && (
                <g>
                  {}
                  <line
                    x1={hoveredPoint.x}
                    y1={padding.top}
                    x2={hoveredPoint.x}
                    y2={height - padding.bottom}
                    stroke="#666666"
                    strokeDasharray="2 2"
                    strokeWidth="1"
                  />
                  {}
                  <circle cx={hoveredPoint.x} cy={hoveredPoint.y} r="6" fill={color} opacity="0.3" />
                  <circle cx={hoveredPoint.x} cy={hoveredPoint.y} r="3.5" fill="#ffffff" stroke={color} strokeWidth="2" />
                </g>
              )}
            </svg>

            {}
            {hoveredPoint && (
              <div
                className="absolute z-20 pointer-events-none rounded-lg border border-[#333333] bg-[#1a1a1a]/95 px-2.5 py-1.5 shadow-xl backdrop-blur-sm text-[11px] text-[#ededed] font-mono animate-in fade-in duration-75"
                style={{
                  left: `${(hoveredPoint.x / width) * 100}%`,
                  top: `10px`,
                  transform: "translateX(-50%)",
                }}
              >
                <div className="flex items-center gap-1.5">
                  <span className="h-2 w-2 rounded-full" style={{ backgroundColor: color }} />
                  <span className="font-bold">
                    {hoveredPoint.value.toFixed(1)} {unit}
                  </span>
                </div>
                <div className="text-[10px] text-[#888888] mt-0.5">
                  {hoveredPoint.timestamp.toLocaleTimeString("id-ID", {
                    hour: "2-digit",
                    minute: "2-digit",
                    second: "2-digit",
                  })}
                </div>
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
};
