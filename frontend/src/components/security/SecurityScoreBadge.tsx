"use client";

import React from "react";
import { AppTheme } from "@/core/theme";
import { ShieldCheck, ShieldAlert, ShieldX } from "lucide-react";

interface SecurityScoreBadgeProps {
  score: number;
  grade: "A" | "B" | "C" | "D" | "F";
  criticalCount: number;
  highCount: number;
  openFindings: number;
}

export const SecurityScoreBadge: React.FC<SecurityScoreBadgeProps> = ({
  score,
  grade,
  criticalCount,
  highCount,
  openFindings,
}) => {
  const getGradeColor = () => {
    switch (grade) {
      case "A":
        return "text-emerald-400 border-emerald-500/30 bg-emerald-950/30";
      case "B":
        return "text-cyan-400 border-cyan-500/30 bg-cyan-950/30";
      case "C":
        return "text-amber-400 border-amber-500/30 bg-amber-950/30";
      case "D":
        return "text-orange-400 border-orange-500/30 bg-orange-950/30";
      default:
        return "text-rose-400 border-rose-500/30 bg-rose-950/30";
    }
  };

  const getStatusText = () => {
    if (criticalCount > 0) return "Tindakan Kritis Diperlukan";
    if (highCount > 0) return "Risiko Tinggi Terdeteksi";
    if (score >= 90) return "Infrastruktur Terlindungi";
    if (score >= 70) return "Kondisi Cukup Aman";
    return "Postur Rentan";
  };

  return (
    <div className={AppTheme.containers.card}>
      <div className="flex flex-col sm:flex-row items-center justify-between gap-6">
        <div className="flex items-center gap-5">
          <div
            className={`w-20 h-20 rounded-2xl flex flex-col items-center justify-center border ${getGradeColor()} shadow-inner`}
          >
            <span className="text-3xl font-extrabold tracking-tight">{score}</span>
            <span className="text-[10px] uppercase font-mono tracking-widest opacity-80">
              Grade {grade}
            </span>
          </div>

          <div className="space-y-1 text-center sm:text-left">
            <div className="flex items-center gap-2 justify-center sm:justify-start">
              {grade === "A" || grade === "B" ? (
                <ShieldCheck className="w-5 h-5 text-emerald-400" />
              ) : grade === "C" ? (
                <ShieldAlert className="w-5 h-5 text-amber-400" />
              ) : (
                <ShieldX className="w-5 h-5 text-rose-400" />
              )}
              <h2 className="text-base font-semibold text-zinc-100">{getStatusText()}</h2>
            </div>
            <p className="text-xs text-zinc-400">
              Sentinel Security Engine menganalisis {openFindings} temuan aktif pada infrastruktur
              cloud Anda.
            </p>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <div className="text-center px-4 py-2 rounded-lg bg-[#141414] border border-[#262626]">
            <p className="text-[10px] uppercase tracking-wider text-rose-400 font-medium">Critical</p>
            <p className="text-lg font-bold text-rose-400">{criticalCount}</p>
          </div>
          <div className="text-center px-4 py-2 rounded-lg bg-[#141414] border border-[#262626]">
            <p className="text-[10px] uppercase tracking-wider text-orange-400 font-medium">High</p>
            <p className="text-lg font-bold text-orange-400">{highCount}</p>
          </div>
          <div className="text-center px-4 py-2 rounded-lg bg-[#141414] border border-[#262626]">
            <p className="text-[10px] uppercase tracking-wider text-zinc-400 font-medium">Total Open</p>
            <p className="text-lg font-bold text-zinc-200">{openFindings}</p>
          </div>
        </div>
      </div>
    </div>
  );
};
