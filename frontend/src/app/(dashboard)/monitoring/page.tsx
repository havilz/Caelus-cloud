"use client";

import React, { useState, useEffect, useCallback } from "react";
import {
  Activity,
  Bell,
  CheckCircle2,
  AlertTriangle,
  Clock,
  Plus,
  Server,
  Filter,
  Check,
  Trash2,
  RefreshCw,
  Sliders,
  ShieldCheck,
} from "lucide-react";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { CreateAlertRuleModal } from "@/components/monitoring/CreateAlertRuleModal";
import { monitoringService } from "@/services/monitoring.service";
import { serverService } from "@/services/server.service";
import { useAuthStore } from "@/stores/useAuthStore";
import { useRealtimeTelemetry } from "@/hooks/useRealtimeTelemetry";
import { Alert, AlertRule, AlertSeverity } from "@/types/monitoring";
import { Server as ServerType } from "@/types/server";
import { AppContainers, AppText, AppColors } from "@/core/theme";

export default function MonitoringPage() {
  const { organization } = useAuthStore();
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [rules, setRules] = useState<AlertRule[]>([]);
  const [servers, setServers] = useState<ServerType[]>([]);
  const [statusFilter, setStatusFilter] = useState<string>("active");
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isCreateRuleOpen, setIsCreateRuleOpen] = useState(false);
  const [actionLoadingId, setActionLoadingId] = useState<string | null>(null);

  const onAlertReceived = useCallback((newAlert: Alert) => {
    setAlerts((prev) => [newAlert, ...prev]);
  }, []);

  const { isConnected } = useRealtimeTelemetry({
    orgId: organization?.id,
    onAlert: onAlertReceived,
  });

  const fetchData = async () => {
    setIsLoading(true);
    try {
      const [alertsRes, rulesRes, serversRes] = await Promise.all([
        monitoringService.listAlerts(
          statusFilter === "all" ? undefined : statusFilter,
          1,
          50
        ),
        monitoringService.listAlertRules(),
        serverService.listServers(1, 100),
      ]);

      setAlerts(alertsRes.data || []);
      setRules(rulesRes || []);
      setServers(serversRes.data || []);
    } catch {
      // Abaikan error saat fetch
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [statusFilter]);

  const handleRefresh = () => {
    setIsRefreshing(true);
    fetchData();
  };

  const handleAcknowledge = async (id: string) => {
    setActionLoadingId(id);
    try {
      await monitoringService.acknowledgeAlert(id);
      await fetchData();
    } finally {
      setActionLoadingId(null);
    }
  };

  const handleResolve = async (id: string) => {
    setActionLoadingId(id);
    try {
      await monitoringService.resolveAlert(id);
      await fetchData();
    } finally {
      setActionLoadingId(null);
    }
  };

  const handleDeleteRule = async (ruleId: string) => {
    if (confirm("Apakah Anda yakin ingin menghapus aturan alert ini?")) {
      try {
        await monitoringService.deleteAlertRule(ruleId);
        await fetchData();
      } catch {
        // Handle error
      }
    }
  };

  const activeAlertsCount = alerts.filter((a) => a.status === "active").length;
  const criticalCount = alerts.filter((a) => a.status === "active" && a.severity === "critical").length;
  const runningServersCount = servers.filter((s) => s.status === "running").length;

  const getSeverityBadge = (sev: AlertSeverity) => {
    switch (sev) {
      case "critical":
        return (
          <span className={`px-2 py-0.5 rounded text-[10px] font-bold ${AppColors.status.danger.bg} ${AppColors.status.danger.text} border ${AppColors.status.danger.border}`}>
            CRITICAL
          </span>
        );
      case "warning":
        return (
          <span className={`px-2 py-0.5 rounded text-[10px] font-bold ${AppColors.status.restarting.bg} ${AppColors.status.restarting.text} border ${AppColors.status.restarting.border}`}>
            WARNING
          </span>
        );
      default:
        return (
          <span className={`px-2 py-0.5 rounded text-[10px] font-bold ${AppColors.status.info.bg} ${AppColors.status.info.text} border ${AppColors.status.info.border}`}>
            INFO
          </span>
        );
    }
  };

  return (
    <div className={AppContainers.pageWrapper}>
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <span className={AppText.categoryTag}>
              {organization?.name || "Workspace"}
            </span>
            <span className={AppColors.text.muted}>•</span>
            <span className={AppText.caption}>Telemetry & Alert Operations</span>
          </div>
          <div className="flex items-center gap-2.5">
            <h1 className={AppText.h2}>Monitoring & Observability Hub</h1>
            {isConnected && (
              <span className={`flex items-center gap-1 text-[10px] ${AppColors.brand.accent} font-mono ${AppColors.brand.primaryLight} px-2 py-0.5 rounded-full`}>
                <span className="relative flex h-1.5 w-1.5">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
                  <span className="relative inline-flex rounded-full h-1.5 w-1.5 bg-emerald-500" />
                </span>
                LIVE EVENT STREAM
              </span>
            )}
          </div>
          <p className={AppText.subtitle}>
            Pusat pemantauan metrik telemetri, deteksi anomali, dan manajemen insiden alert organisasi.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleRefresh}
            className={`${AppColors.text.secondary} ${AppColors.border.subtle} hover:${AppColors.text.primary}`}
          >
            <RefreshCw className={`h-3.5 w-3.5 mr-1.5 ${isRefreshing ? "animate-spin" : ""}`} />
            Segarkan
          </Button>
          <Button
            size="sm"
            onClick={() => setIsCreateRuleOpen(true)}
            className={`${AppColors.brand.primary} text-xs`}
          >
            <Plus className="h-3.5 w-3.5 mr-1.5" />
            Tambah Aturan Alert
          </Button>
        </div>
      </div>

      {/* Aggregate Overview Cards */}
      <div className={AppContainers.metricsGrid}>
        <Card className={`${AppContainers.card} ${AppContainers.cardHover}`}>
          <CardContent className="p-4 flex items-center justify-between">
            <div>
              <p className={AppText.bodySm}>Total Server Terpantau</p>
              <p className={`text-2xl font-bold ${AppColors.text.primary} font-mono mt-1`}>{servers.length}</p>
              <p className={`text-[11px] ${AppColors.brand.accent} font-mono mt-0.5`}>{runningServersCount} Running Host</p>
            </div>
            <div className={`flex h-10 w-10 items-center justify-center rounded-xl ${AppColors.brand.primaryLight}`}>
              <Server className="h-5 w-5" />
            </div>
          </CardContent>
        </Card>

        <Card className={`${AppContainers.card} ${AppContainers.cardHover}`}>
          <CardContent className="p-4 flex items-center justify-between">
            <div>
              <p className={AppText.bodySm}>Alert Aktif Saat Ini</p>
              <p className={`text-2xl font-bold ${AppColors.status.danger.text} font-mono mt-1`}>{activeAlertsCount}</p>
              <p className={`text-[11px] ${AppColors.text.muted} font-mono mt-0.5`}>{criticalCount} Tingkat Kritis</p>
            </div>
            <div className={`flex h-10 w-10 items-center justify-center rounded-xl ${AppColors.status.danger.bg} ${AppColors.status.danger.text} border ${AppColors.status.danger.border}`}>
              <Bell className="h-5 w-5" />
            </div>
          </CardContent>
        </Card>

        <Card className={`${AppContainers.card} ${AppContainers.cardHover}`}>
          <CardContent className="p-4 flex items-center justify-between">
            <div>
              <p className={AppText.bodySm}>Aturan Threshold Aktif</p>
              <p className={`text-2xl font-bold ${AppColors.text.primary} font-mono mt-1`}>{rules.length}</p>
              <p className={AppText.caption}>Auto-evaluator 60s</p>
            </div>
            <div className={`flex h-10 w-10 items-center justify-center rounded-xl ${AppColors.status.info.bg} ${AppColors.status.info.text} border ${AppColors.status.info.border}`}>
              <Sliders className="h-5 w-5" />
            </div>
          </CardContent>
        </Card>

        <Card className={`${AppContainers.card} ${AppContainers.cardHover}`}>
          <CardContent className="p-4 flex items-center justify-between">
            <div>
              <p className={AppText.bodySm}>Sentinel Status</p>
              <p className={`text-2xl font-bold ${AppColors.brand.accent} font-mono mt-1`}>100%</p>
              <p className={AppText.caption}>Seluruh node terlindungi</p>
            </div>
            <div className={`flex h-10 w-10 items-center justify-center rounded-xl bg-purple-500/10 text-purple-400 border border-purple-500/20`}>
              <ShieldCheck className="h-5 w-5" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Main Content Grid: Alerts Incident Table & Threshold Rules */}
      <div className={AppContainers.overviewSplitGrid}>
        {/* Left 2 Cols: Alerts Incident Center */}
        <div className="lg:col-span-2 space-y-4">
          <Card className={AppContainers.card}>
            <CardHeader className={`${AppContainers.cardHeader} flex flex-col sm:flex-row sm:items-center justify-between gap-3`}>
              <div>
                <CardTitle className={AppText.h4}>Pusat Insiden Alert</CardTitle>
                <CardDescription className={AppText.caption}>Daftar peringatan yang terpicu dari anomali metrik</CardDescription>
              </div>

              {/* Status Filter Tabs */}
              <div className={`flex items-center gap-1 ${AppColors.bg.surface} p-1 rounded-lg border ${AppColors.border.subtle}`}>
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
                    className={`px-3 py-1 text-xs font-medium rounded-md transition-all cursor-pointer ${
                      statusFilter === tab.id
                        ? `${AppColors.bg.surfaceElevated} ${AppColors.brand.accent} font-semibold shadow-xs`
                        : `${AppColors.text.muted} hover:${AppColors.text.primary}`
                    }`}
                  >
                    {tab.label}
                  </button>
                ))}
              </div>
            </CardHeader>

            <CardContent className="p-0">
              {isLoading ? (
                <div className={`flex h-64 items-center justify-center ${AppText.bodySm}`}>
                  <div className="flex flex-col items-center gap-2">
                    <div className="h-5 w-5 animate-spin rounded-full border-2 border-emerald-500 border-t-transparent" />
                    <p>Memuat daftar insiden alert...</p>
                  </div>
                </div>
              ) : alerts.length === 0 ? (
                <div className="flex h-64 flex-col items-center justify-center text-center p-6">
                  <CheckCircle2 className="h-10 w-10 mb-2 text-emerald-500/50" />
                  <p className={AppText.h4}>Tidak Ada Alert</p>
                  <p className={AppText.subtitle}>
                    Seluruh performa server pada status ini beroperasi optimal.
                  </p>
                </div>
              ) : (
                <div className={`divide-y ${AppColors.border.subtle}`}>
                  {alerts.map((alert) => (
                    <div
                      key={alert.id}
                      className={`p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3 ${AppColors.bg.surfaceHover} transition-colors`}
                    >
                      <div className="space-y-1.5">
                        <div className="flex items-center gap-2">
                          {getSeverityBadge(alert.severity)}
                          <span className={AppText.h4}>{alert.title}</span>
                        </div>
                        <p className={AppText.bodySm}>{alert.message}</p>
                        <div className={`flex items-center gap-3 ${AppText.codeMuted}`}>
                          <span>Waktu: {new Date(alert.triggered_at).toLocaleString("id-ID")}</span>
                          {alert.current_value !== undefined && (
                            <span className={`${AppColors.status.danger.text} font-bold`}>
                              Terukur: {alert.current_value.toFixed(1)}%
                            </span>
                          )}
                        </div>
                      </div>

                      <div className="flex items-center gap-2 self-start sm:self-auto shrink-0">
                        {alert.status === "active" && (
                          <Button
                            variant="outline"
                            size="sm"
                            disabled={actionLoadingId === alert.id}
                            onClick={() => handleAcknowledge(alert.id)}
                            className={`${AppColors.status.restarting.text} ${AppColors.status.restarting.border} hover:${AppColors.status.restarting.bg} text-xs h-7.5`}
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
                            className={`${AppColors.brand.accent} ${AppColors.brand.border} hover:bg-emerald-950/20 text-xs h-7.5`}
                          >
                            <CheckCircle2 className="h-3 w-3 mr-1" />
                            Selesaikan
                          </Button>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Right 1 Col: Alert Rules Management */}
        <div className="space-y-4">
          <Card className={AppContainers.card}>
            <CardHeader className={`${AppContainers.cardHeader} flex items-center justify-between`}>
              <div>
                <CardTitle className={AppText.h4}>Aturan Evaluasi ({rules.length})</CardTitle>
                <CardDescription className={AppText.caption}>Kriteria threshold alert otomatis</CardDescription>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setIsCreateRuleOpen(true)}
                className={`h-7 text-[11px] ${AppColors.border.subtle}`}
              >
                <Plus className="h-3 w-3 mr-1" />
                Tambah
              </Button>
            </CardHeader>
            <CardContent className="p-3 space-y-2.5">
              {rules.length === 0 ? (
                <p className={`text-center py-6 ${AppText.caption}`}>
                  Belum ada aturan threshold yang dikonfigurasi.
                </p>
              ) : (
                rules.map((rule) => (
                  <div
                    key={rule.id}
                    className={`p-3 rounded-lg border ${AppColors.border.subtle} ${AppColors.bg.surfaceSubtle} flex items-center justify-between gap-2`}
                  >
                    <div className="space-y-0.5">
                      <p className={`text-xs font-semibold ${AppColors.text.primary}`}>{rule.name}</p>
                      <p className={AppText.codeMuted}>
                        {rule.metric_name} {rule.operator} {rule.threshold}% ({rule.duration_seconds}s)
                      </p>
                      <span className={`inline-block text-[9px] uppercase font-bold ${AppColors.brand.accent} ${AppColors.brand.primaryLight} px-1.5 py-0.2 rounded mt-1`}>
                        {rule.severity}
                      </span>
                    </div>

                    <button
                      type="button"
                      onClick={() => handleDeleteRule(rule.id)}
                      className={`p-1.5 ${AppColors.text.muted} hover:${AppColors.status.danger.text} transition-colors cursor-pointer`}
                      aria-label="Hapus aturan"
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </button>
                  </div>
                ))
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Create Alert Rule Modal */}
      <CreateAlertRuleModal
        isOpen={isCreateRuleOpen}
        onClose={() => setIsCreateRuleOpen(false)}
        onRuleCreated={fetchData}
      />
    </div>
  );
}
