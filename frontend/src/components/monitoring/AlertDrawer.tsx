"use client";

import React, { useState, useEffect } from "react";
import {
  X,
  AlertTriangle,
  CheckCircle2,
  Bell,
  Clock,
  Check,
  Server,
  RefreshCw,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Alert, AlertStatus, AlertSeverity } from "@/types/monitoring";
import { monitoringService } from "@/services/monitoring.service";

interface AlertDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  onAlertUpdated?: () => void;
}

export const AlertDrawer: React.FC<AlertDrawerProps> = ({
  isOpen,
  onClose,
  onAlertUpdated,
}) => {
  const [statusFilter, setStatusFilter] = useState<string>("active");
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [actionLoadingId, setActionLoadingId] = useState<string | null>(null);

  const fetchAlerts = async () => {
    setIsLoading(true);
    try {
      const res = await monitoringService.listAlerts(
        statusFilter === "all" ? undefined : statusFilter,
        1,
        50
      );
      setAlerts(res.data || []);
    } catch {
      
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (isOpen) {
      fetchAlerts();
    }
  }, [isOpen, statusFilter]);

  const handleAcknowledge = async (id: string) => {
    setActionLoadingId(id);
    try {
      await monitoringService.acknowledgeAlert(id);
      await fetchAlerts();
      onAlertUpdated?.();
    } finally {
      setActionLoadingId(null);
    }
  };

  const handleResolve = async (id: string) => {
    setActionLoadingId(id);
    try {
      await monitoringService.resolveAlert(id);
      await fetchAlerts();
      onAlertUpdated?.();
    } finally {
      setActionLoadingId(null);
    }
  };

  if (!isOpen) return null;

  const getSeverityBadge = (sev: AlertSeverity) => {
    switch (sev) {
      case "critical":
        return (
          <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-rose-500/10 text-rose-400 border border-rose-500/20">
            CRITICAL
          </span>
        );
      case "warning":
        return (
          <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-amber-500/10 text-amber-400 border border-amber-500/20">
            WARNING
          </span>
        );
      default:
        return (
          <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-blue-500/10 text-blue-400 border border-blue-500/20">
            INFO
          </span>
        );
    }
  };

  return (
    <div className="fixed inset-0 z-50 overflow-hidden">
      {}
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-xs transition-opacity animate-in fade-in"
        onClick={onClose}
      />

      <div className="fixed inset-y-0 right-0 max-w-full flex pl-10">
        <div className="w-screen max-w-md bg-[#121212] border-l border-[#262626] shadow-2xl flex flex-col animate-in slide-in-from-right duration-200">
          {}
          <div className="p-4 border-b border-[#262626] flex items-center justify-between bg-[#171717]">
            <div className="flex items-center gap-2">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-rose-500/10 text-rose-400 border border-rose-500/20">
                <Bell className="h-4 w-4" />
              </div>
              <div>
                <h2 className="text-sm font-semibold text-[#ededed]">Pusat Notifikasi Alert</h2>
                <p className="text-[11px] text-[#a1a1a1]">Insiden telemetri dan ambang batas sistem</p>
              </div>
            </div>
            <div className="flex items-center gap-1">
              <Button
                variant="ghost"
                size="sm"
                onClick={fetchAlerts}
                className="h-8 w-8 p-0 text-[#a1a1a1] hover:text-[#ededed]"
              >
                <RefreshCw className={`h-4 w-4 ${isLoading ? "animate-spin" : ""}`} />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={onClose}
                className="h-8 w-8 p-0 text-[#a1a1a1] hover:text-[#ededed]"
              >
                <X className="h-4 w-4" />
              </Button>
            </div>
          </div>

          {}
          <div className="p-3 border-b border-[#262626] bg-[#141414] flex gap-1">
            {[
              { id: "active", label: "Aktif" },
              { id: "acknowledged", label: "Ditandai" },
              { id: "resolved", label: "Selesai" },
              { id: "all", label: "Semua" },
            ].map((tab) => (
              <button
                key={tab.id}
                type="button"
                onClick={() => setStatusFilter(tab.id)}
                className={`flex-1 py-1.5 text-xs font-medium rounded-md transition-all cursor-pointer ${
                  statusFilter === tab.id
                    ? "bg-[#262626] text-emerald-400 font-semibold shadow-xs"
                    : "text-[#888888] hover:text-[#ededed]"
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>

          {}
          <div className="flex-1 overflow-y-auto p-4 space-y-3">
            {isLoading && alerts.length === 0 ? (
              <div className="flex h-64 items-center justify-center text-xs text-[#a1a1a1]">
                <div className="flex flex-col items-center gap-2">
                  <div className="h-5 w-5 animate-spin rounded-full border-2 border-emerald-500 border-t-transparent" />
                  <p>Memuat daftar insiden alert...</p>
                </div>
              </div>
            ) : alerts.length === 0 ? (
              <div className="flex h-64 flex-col items-center justify-center text-center text-[#666666]">
                <CheckCircle2 className="h-8 w-8 mb-2 text-emerald-500/50" />
                <p className="text-xs font-medium text-[#a1a1a1]">Tidak ada alert pada status ini</p>
                <p className="text-[11px] text-[#707070] mt-1">Seluruh metrik server beroperasi normal</p>
              </div>
            ) : (
              alerts.map((alert) => (
                <div
                  key={alert.id}
                  className="rounded-xl border border-[#262626] bg-[#171717] p-3.5 space-y-2.5 transition-all hover:border-[#383838]"
                >
                  <div className="flex items-start justify-between gap-2">
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        {getSeverityBadge(alert.severity)}
                        <span className="text-[11px] text-[#707070] flex items-center gap-1 font-mono">
                          <Clock className="h-3 w-3" />
                          {new Date(alert.triggered_at).toLocaleTimeString("id-ID", {
                            hour: "2-digit",
                            minute: "2-digit",
                          })}
                        </span>
                      </div>
                      <h4 className="text-xs font-semibold text-[#ededed] leading-snug">{alert.title}</h4>
                    </div>
                  </div>

                  <p className="text-[11px] text-[#a1a1a1] leading-relaxed">{alert.message}</p>

                  {}
                  {alert.current_value !== undefined && alert.threshold_value !== undefined && (
                    <div className="flex items-center gap-2 text-[11px] font-mono p-2 rounded-lg bg-[#111111] border border-[#222222]">
                      <span className="text-[#888888]">Nilai Terukur:</span>
                      <span className="text-rose-400 font-bold">{alert.current_value.toFixed(1)}%</span>
                      <span className="text-[#555555]">/</span>
                      <span className="text-[#888888]">Threshold:</span>
                      <span className="text-emerald-400 font-bold">{alert.threshold_value.toFixed(1)}%</span>
                    </div>
                  )}

                  {}
                  <div className="flex items-center justify-between pt-1 border-t border-[#222222] text-xs">
                    <span className="text-[10px] uppercase font-bold tracking-wider text-[#666666]">
                      Status: <span className="text-[#a1a1a1]">{alert.status}</span>
                    </span>

                    <div className="flex items-center gap-1.5">
                      {alert.status === "active" && (
                        <Button
                          variant="outline"
                          size="sm"
                          disabled={actionLoadingId === alert.id}
                          onClick={() => handleAcknowledge(alert.id)}
                          className="h-6 px-2 text-[10px] text-amber-400 hover:bg-amber-950/20 border-amber-500/20"
                        >
                          <Check className="h-3 w-3 mr-1" />
                          Tandai
                        </Button>
                      )}
                      {alert.status !== "resolved" && (
                        <Button
                          variant="outline"
                          size="sm"
                          disabled={actionLoadingId === alert.id}
                          onClick={() => handleResolve(alert.id)}
                          className="h-6 px-2 text-[10px] text-emerald-400 hover:bg-emerald-950/20 border-emerald-500/20"
                        >
                          <CheckCircle2 className="h-3 w-3 mr-1" />
                          Selesaikan
                        </Button>
                      )}
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
