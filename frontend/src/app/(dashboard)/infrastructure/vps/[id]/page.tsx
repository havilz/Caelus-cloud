"use client";

import React, { useState, useEffect, use, useCallback } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  Server as ServerIcon,
  ArrowLeft,
  Power,
  PowerOff,
  RotateCcw,
  Sliders,
  Trash2,
  Copy,
  Check,
  Cpu,
  HardDrive,
  Activity,
  ShieldCheck,
  Calendar,
  Globe,
  Terminal,
  Bell,
  Plus,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { ServerStatusBadge } from "@/components/server/ServerStatusBadge";
import { ResizeServerModal } from "@/components/server/ResizeServerModal";
import { MetricCard } from "@/components/monitoring/MetricCard";
import { MetricTimeSeriesChart } from "@/components/monitoring/MetricTimeSeriesChart";
import { LogViewer } from "@/components/monitoring/LogViewer";
import { CreateAlertRuleModal } from "@/components/monitoring/CreateAlertRuleModal";
import { serverService } from "@/services/server.service";
import { monitoringService } from "@/services/monitoring.service";
import { useRealtimeTelemetry } from "@/hooks/useRealtimeTelemetry";
import { Server } from "@/types/server";
import { ServerMetric, Alert, AlertRule, LogEntry } from "@/types/monitoring";
import { AppContainers, AppText, AppColors } from "@/core/theme";

export default function ServerDetailPage({
  params,
}: Readonly<{
  params: Promise<{ id: string }>;
}>) {
  const resolvedParams = use(params);
  const router = useRouter();
  const serverId = resolvedParams.id;

  const [server, setServer] = useState<Server | null>(null);
  const [activeTab, setActiveTab] = useState<"overview" | "metrics" | "logs" | "alerts">("overview");
  const [timeRange, setTimeRange] = useState("1h");
  const [metricsHistory, setMetricsHistory] = useState<ServerMetric[]>([]);
  const [serverAlerts, setServerAlerts] = useState<Alert[]>([]);
  const [serverRules, setServerRules] = useState<AlertRule[]>([]);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isActionLoading, setIsActionLoading] = useState(false);
  const [isResizeOpen, setIsResizeOpen] = useState(false);
  const [isCreateRuleOpen, setIsCreateRuleOpen] = useState(false);
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const onMetricUpdate = useCallback((metric: ServerMetric) => {
    setMetricsHistory((prev) => {
      const next = [...prev, metric];
      if (next.length > 100) next.shift();
      return next;
    });

    setLogs((prev) => [
      ...prev.slice(-199),
      {
        id: Math.random().toString(36).substring(7),
        timestamp: new Date().toISOString(),
        line: `[agent-telemetry] Reported CPU: ${metric.cpu_usage_pct.toFixed(1)}%, RAM: ${metric.memory_usage_pct.toFixed(1)}%, Net: ${metric.network_in_rate_kbps.toFixed(1)} KB/s`,
        level: metric.cpu_usage_pct > 85 ? "WARN" : "INFO",
        service: "caelus-agent",
      },
    ]);
  }, []);

  const onAlertReceived = useCallback((alert: Alert) => {
    setServerAlerts((prev) => [alert, ...prev]);
    setLogs((prev) => [
      ...prev.slice(-199),
      {
        id: Math.random().toString(36).substring(7),
        timestamp: new Date().toISOString(),
        line: `[ALERT TRIGGERED] ${alert.title}: ${alert.message}`,
        level: "ERROR",
        service: "alert-evaluator",
      },
    ]);
  }, []);

  const { isConnected, latestMetric } = useRealtimeTelemetry({
    serverId,
    onMetricUpdate,
    onAlert: onAlertReceived,
  });

  const fetchServer = async () => {
    setIsLoading(true);
    try {
      const data = await serverService.getServer(serverId);
      setServer(data);
    } catch (err: any) {
      setError(err.response?.data?.message || "Gagal memuat detail server.");
    } finally {
      setIsLoading(false);
    }
  };

  const fetchMetricsHistory = async () => {
    try {
      const history = await monitoringService.getMetricHistory(serverId, timeRange);
      setMetricsHistory(history);
    } catch {
      
    }
  };

  const fetchAlertsAndRules = async () => {
    try {
      const alertsRes = await monitoringService.listAlerts(undefined, 1, 30);
      const serverSpecificAlerts = (alertsRes.data || []).filter(
        (a) => a.server_id === serverId
      );
      setServerAlerts(serverSpecificAlerts);

      const rulesRes = await monitoringService.listAlertRules();
      const serverSpecificRules = rulesRes.filter(
        (r) => !r.server_id || r.server_id === serverId
      );
      setServerRules(serverSpecificRules);
    } catch {
      
    }
  };

  useEffect(() => {
    fetchServer();
    fetchMetricsHistory();
    fetchAlertsAndRules();

    setLogs([
      {
        id: "log-1",
        timestamp: new Date(Date.now() - 60000).toISOString(),
        line: "Host daemon caelus-agent v1.0.0 started and registered",
        level: "INFO",
        service: "systemd",
      },
      {
        id: "log-2",
        timestamp: new Date(Date.now() - 45000).toISOString(),
        line: "Connected to Unix socket /var/run/docker.sock successfully",
        level: "INFO",
        service: "docker-inspector",
      },
      {
        id: "log-3",
        timestamp: new Date(Date.now() - 30000).toISOString(),
        line: "Ingestion pipeline TLS handshake established with Caelus Backend",
        level: "INFO",
        service: "transport",
      },
    ]);
  }, [serverId]);

  useEffect(() => {
    fetchMetricsHistory();
  }, [timeRange]);

  const handleCopy = (text?: string) => {
    if (!text) return;
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleAction = async (actionFn: (id: string) => Promise<void>) => {
    setIsActionLoading(true);
    try {
      await actionFn(serverId);
      await fetchServer();
    } finally {
      setIsActionLoading(false);
    }
  };

  const handleDelete = async () => {
    if (!server) return;
    if (confirm(`Apakah Anda yakin ingin menterminasi dan menghapus server ${server.name}?`)) {
      setIsActionLoading(true);
      try {
        await serverService.deleteServer(serverId);
        router.push("/infrastructure/vps");
      } finally {
        setIsActionLoading(false);
      }
    }
  };

  const currentMetric = latestMetric || (metricsHistory.length > 0 ? metricsHistory[metricsHistory.length - 1] : null);

  if (isLoading) {
    return (
      <div className={`flex h-96 items-center justify-center ${AppColors.text.secondary}`}>
        <div className="flex flex-col items-center gap-3">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-emerald-500 border-t-transparent" />
          <p className={AppText.bodySm}>Memuat detail server VPS...</p>
        </div>
      </div>
    );
  }

  if (error || !server) {
    return (
      <div className="flex h-96 flex-col items-center justify-center gap-4 text-center">
        <div className={`rounded-full ${AppColors.status.danger.bg} p-3 ${AppColors.status.danger.text}`}>
          <ServerIcon className="h-6 w-6" />
        </div>
        <div>
          <h2 className={AppText.h3}>Server Tidak Ditemukan</h2>
          <p className={AppText.subtitle}>{error || "Instance mungkin telah dihapus."}</p>
        </div>
        <Link href="/infrastructure/vps">
          <Button variant="outline" size="sm" className={`${AppColors.text.secondary} ${AppColors.border.subtle}`}>
            <ArrowLeft className="h-3.5 w-3.5 mr-1.5" />
            Kembali ke Daftar Server
          </Button>
        </Link>
      </div>
    );
  }

  return (
    <div className={AppContainers.pageWrapper}>
      {}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <Link href="/infrastructure/vps">
            <Button variant="ghost" size="sm" className={`h-8 w-8 p-0 ${AppColors.text.secondary} hover:${AppColors.text.primary}`}>
              <ArrowLeft className="h-4 w-4" />
            </Button>
          </Link>
          <div>
            <div className="flex items-center gap-2.5">
              <h1 className={AppText.h2}>{server.name}</h1>
              <ServerStatusBadge status={server.status} />
              {isConnected && (
                <span className={`flex items-center gap-1 text-[10px] ${AppColors.brand.accent} font-mono ${AppColors.brand.primaryLight} px-2 py-0.5 rounded-full`}>
                  <span className="relative flex h-1.5 w-1.5">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
                    <span className="relative inline-flex rounded-full h-1.5 w-1.5 bg-emerald-500" />
                  </span>
                  LIVE TELEMETRY
                </span>
              )}
            </div>
            <p className={AppText.codeMuted}>{server.id}</p>
          </div>
        </div>

        {}
        <div className="flex items-center gap-1.5 self-start sm:self-auto flex-wrap">
          {server.status === "stopped" ? (
            <Button
              variant="outline"
              size="sm"
              disabled={isActionLoading}
              onClick={() => handleAction(serverService.startServer)}
              className={`${AppColors.brand.accent} ${AppColors.brand.border} hover:bg-emerald-950/30`}
            >
              <Power className="h-3.5 w-3.5 mr-1" />
              Start
            </Button>
          ) : (
            <Button
              variant="outline"
              size="sm"
              disabled={isActionLoading || server.status !== "running"}
              onClick={() => handleAction(serverService.shutdownServer)}
              className={`${AppColors.status.restarting.text} ${AppColors.status.restarting.border} hover:bg-amber-950/30`}
            >
              <PowerOff className="h-3.5 w-3.5 mr-1" />
              Shutdown
            </Button>
          )}

          <Button
            variant="outline"
            size="sm"
            disabled={isActionLoading || server.status !== "running"}
            onClick={() => handleAction(serverService.rebootServer)}
            className={`${AppColors.text.secondary} ${AppColors.border.subtle} hover:${AppColors.text.primary}`}
          >
            <RotateCcw className="h-3.5 w-3.5 mr-1" />
            Reboot
          </Button>

          <Button
            variant="outline"
            size="sm"
            disabled={isActionLoading}
            onClick={() => setIsResizeOpen(true)}
            className={`${AppColors.text.secondary} ${AppColors.border.subtle} hover:${AppColors.text.primary}`}
          >
            <Sliders className="h-3.5 w-3.5 mr-1" />
            Resize
          </Button>

          <Button
            variant="outline"
            size="sm"
            disabled={isActionLoading}
            onClick={handleDelete}
            className={`${AppColors.status.danger.text} ${AppColors.status.danger.border} hover:bg-rose-950/30`}
          >
            <Trash2 className="h-3.5 w-3.5 mr-1" />
            Terminate
          </Button>
        </div>
      </div>

      {}
      <div className={`flex items-center gap-1 border-b ${AppColors.border.subtle} pb-0`}>
        {[
          { id: "overview", label: "Overview & Hardware", icon: ServerIcon },
          { id: "metrics", label: "Live Telemetry & Metrics", icon: Activity, badge: isConnected ? "Live" : undefined },
          { id: "logs", label: "Console Logs", icon: Terminal },
          { id: "alerts", label: "Alerts & Thresholds", icon: Bell, badge: serverAlerts.length > 0 ? `${serverAlerts.length}` : undefined },
        ].map((tab) => {
          const Icon = tab.icon;
          const isActive = activeTab === tab.id;
          return (
            <button
              key={tab.id}
              type="button"
              onClick={() => setActiveTab(tab.id as any)}
              className={`flex items-center gap-2 px-4 py-2.5 text-xs font-medium border-b-2 -mb-px transition-all cursor-pointer ${
                isActive
                  ? `border-emerald-500 ${AppColors.brand.accent} font-semibold`
                  : `border-transparent ${AppColors.text.secondary} hover:${AppColors.text.primary}`
              }`}
            >
              <Icon className="h-3.5 w-3.5" />
              <span>{tab.label}</span>
              {tab.badge && (
                <span
                  className={`px-1.5 py-0.2 rounded-full text-[10px] font-bold ${
                    tab.badge === "Live"
                      ? `${AppColors.brand.primaryLight}`
                      : "bg-rose-500 text-white"
                  }`}
                >
                  {tab.badge}
                </span>
              )}
            </button>
          );
        })}
      </div>

      {}
      {activeTab === "overview" && (
        <div className={AppContainers.overviewSplitGrid}>
          <div className="lg:col-span-2 space-y-4">
            <Card className={AppContainers.card}>
              <CardHeader className={AppContainers.cardHeader}>
                <CardTitle className={AppText.h4}>Spesifikasi Komputasi</CardTitle>
                <CardDescription className={AppText.caption}>Rincian alokasi hardware yang diprovisioning</CardDescription>
              </CardHeader>
              <CardContent className={AppContainers.cardContent}>
                <div className={AppContainers.specsGrid}>
                  <div className={`p-3.5 rounded-lg ${AppColors.bg.surfaceSubtle} border ${AppColors.border.subtle}`}>
                    <div className={`flex items-center gap-2 ${AppColors.text.secondary} mb-1`}>
                      <Cpu className={`h-4 w-4 ${AppColors.brand.accent}`} />
                      <span>vCPU Cores</span>
                    </div>
                    <p className={`text-base font-bold ${AppColors.text.primary}`}>{server.cpu_cores} Cores</p>
                  </div>

                  <div className={`p-3.5 rounded-lg ${AppColors.bg.surfaceSubtle} border ${AppColors.border.subtle}`}>
                    <div className={`flex items-center gap-2 ${AppColors.text.secondary} mb-1`}>
                      <HardDrive className={`h-4 w-4 ${AppColors.brand.accent}`} />
                      <span>Memory RAM</span>
                    </div>
                    <p className={`text-base font-bold ${AppColors.text.primary}`}>
                      {(server.memory_mb / 1024).toFixed(0)} GB RAM
                    </p>
                    <p className={AppText.codeMuted}>{server.memory_mb} MB</p>
                  </div>

                  <div className={`p-3.5 rounded-lg ${AppColors.bg.surfaceSubtle} border ${AppColors.border.subtle}`}>
                    <div className={`flex items-center gap-2 ${AppColors.text.secondary} mb-1`}>
                      <HardDrive className={`h-4 w-4 ${AppColors.brand.accent}`} />
                      <span>SSD Storage</span>
                    </div>
                    <p className={`text-base font-bold ${AppColors.text.primary}`}>{server.disk_gb} GB</p>
                  </div>

                  <div className={`p-3.5 rounded-lg ${AppColors.bg.surfaceSubtle} border ${AppColors.border.subtle}`}>
                    <div className={`flex items-center gap-2 ${AppColors.text.secondary} mb-1`}>
                      <Globe className={`h-4 w-4 ${AppColors.brand.accent}`} />
                      <span>Data Center Region</span>
                    </div>
                    <p className={`text-xs font-semibold ${AppColors.text.primary}`}>{server.region}</p>
                  </div>

                  <div className={`p-3.5 rounded-lg ${AppColors.bg.surfaceSubtle} border ${AppColors.border.subtle}`}>
                    <div className={`flex items-center gap-2 ${AppColors.text.secondary} mb-1`}>
                      <Activity className={`h-4 w-4 ${AppColors.brand.accent}`} />
                      <span>Sistem Operasi</span>
                    </div>
                    <p className={`text-xs font-semibold ${AppColors.text.primary}`}>{server.os_type}</p>
                  </div>

                  <div className={`p-3.5 rounded-lg ${AppColors.bg.surfaceSubtle} border ${AppColors.border.subtle}`}>
                    <div className={`flex items-center gap-2 ${AppColors.text.secondary} mb-1`}>
                      <Calendar className={`h-4 w-4 ${AppColors.brand.accent}`} />
                      <span>Tanggal Dibuat</span>
                    </div>
                    <p className={`text-xs font-medium ${AppColors.text.primary}`}>
                      {new Date(server.created_at).toLocaleDateString("id-ID", {
                        day: "numeric",
                        month: "short",
                        year: "numeric",
                      })}
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className={AppContainers.card}>
              <CardHeader className={AppContainers.cardHeader}>
                <CardTitle className={AppText.h4}>Jaringan & Host Endpoint</CardTitle>
                <CardDescription className={AppText.caption}>Alamat IP publik dan identifier host instance</CardDescription>
              </CardHeader>
              <CardContent className="p-5 pt-4 space-y-3 text-xs">
                <div className={`flex items-center justify-between p-3 rounded-lg ${AppColors.bg.surfaceSubtle} border ${AppColors.border.subtle}`}>
                  <div>
                    <p className={AppText.caption}>Public IPv4 Address</p>
                    <p className={`text-sm font-mono font-semibold ${AppColors.text.primary} mt-0.5`}>
                      {server.ip_address || "Sedang dialokasikan..."}
                    </p>
                  </div>
                  {server.ip_address && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleCopy(server.ip_address)}
                      className={`${AppColors.text.secondary} ${AppColors.border.subtle}`}
                    >
                      {copied ? <Check className={`h-3.5 w-3.5 ${AppColors.brand.accent}`} /> : <Copy className="h-3.5 w-3.5" />}
                      <span>{copied ? "Tersalin" : "Salin IP"}</span>
                    </Button>
                  )}
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div className={`p-3 rounded-lg ${AppColors.bg.surfaceSubtle} border ${AppColors.border.subtle}`}>
                    <p className={AppText.caption}>External Provider ID</p>
                    <p className={AppText.codeMuted}>{server.external_server_id || "-"}</p>
                  </div>
                  <div className={`p-3 rounded-lg ${AppColors.bg.surfaceSubtle} border ${AppColors.border.subtle}`}>
                    <p className={AppText.caption}>Hostname</p>
                    <p className={AppText.codeMuted}>{server.hostname || server.name}</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          <div className="space-y-4">
            <Card className={AppContainers.card}>
              <CardHeader className={AppContainers.cardHeader}>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Activity className={`h-4 w-4 ${AppColors.brand.accent}`} />
                    <CardTitle className={AppText.h4}>Live Snapshot</CardTitle>
                  </div>
                  {isConnected && (
                    <span className="flex h-2 w-2 relative">
                      <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
                      <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
                    </span>
                  )}
                </div>
                <CardDescription className={AppText.caption}>Beban sumber daya real-time instance</CardDescription>
              </CardHeader>
              <CardContent className="p-5 pt-4 space-y-4 text-xs">
                <div>
                  <div className={`flex justify-between mb-1 ${AppColors.text.secondary}`}>
                    <span>vCPU Utilization</span>
                    <span className={`font-mono font-semibold ${AppColors.brand.accent}`}>
                      {currentMetric ? `${currentMetric.cpu_usage_pct.toFixed(1)}%` : "0.0%"}
                    </span>
                  </div>
                  <div className="w-full bg-[#202020] rounded-full h-1.5 overflow-hidden">
                    <div
                      className="bg-emerald-500 h-1.5 rounded-full transition-all duration-500"
                      style={{ width: `${Math.min(currentMetric?.cpu_usage_pct || 0, 100)}%` }}
                    />
                  </div>
                </div>

                <div>
                  <div className={`flex justify-between mb-1 ${AppColors.text.secondary}`}>
                    <span>RAM Memory Load</span>
                    <span className={`font-mono font-semibold ${AppColors.brand.accent}`}>
                      {currentMetric
                        ? `${currentMetric.memory_usage_pct.toFixed(1)}% (${(currentMetric.memory_used_mb / 1024).toFixed(1)} / ${(currentMetric.memory_total_mb / 1024).toFixed(1)} GB)`
                        : "0.0%"}
                    </span>
                  </div>
                  <div className="w-full bg-[#202020] rounded-full h-1.5 overflow-hidden">
                    <div
                      className="bg-emerald-500 h-1.5 rounded-full transition-all duration-500"
                      style={{ width: `${Math.min(currentMetric?.memory_usage_pct || 0, 100)}%` }}
                    />
                  </div>
                </div>

                <div>
                  <div className={`flex justify-between mb-1 ${AppColors.text.secondary}`}>
                    <span>SSD Storage Usage</span>
                    <span className={`font-mono font-semibold ${AppColors.brand.accent}`}>
                      {currentMetric
                        ? `${currentMetric.disk_usage_pct.toFixed(1)}% (${currentMetric.disk_used_gb.toFixed(1)} / ${currentMetric.disk_total_gb.toFixed(1)} GB)`
                        : "0.0%"}
                    </span>
                  </div>
                  <div className="w-full bg-[#202020] rounded-full h-1.5 overflow-hidden">
                    <div
                      className="bg-emerald-500 h-1.5 rounded-full transition-all duration-500"
                      style={{ width: `${Math.min(currentMetric?.disk_usage_pct || 0, 100)}%` }}
                    />
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className={AppContainers.card}>
              <CardHeader className="p-4 pb-2">
                <div className={`flex items-center gap-2 ${AppColors.brand.accent}`}>
                  <ShieldCheck className="h-4 w-4" />
                  <CardTitle className={AppText.h4}>Sentinel Security</CardTitle>
                </div>
              </CardHeader>
              <CardContent className="p-4 pt-2 space-y-2 text-xs">
                <div className={`flex items-center justify-between py-1 border-b ${AppColors.border.subtle}`}>
                  <span className={AppColors.text.secondary}>Port Firewall</span>
                  <span className={`${AppColors.brand.accent} font-medium`}>Port 22, 80, 443</span>
                </div>
                <div className="flex items-center justify-between py-1">
                  <span className={AppColors.text.secondary}>Agent Daemon</span>
                  <span className={isConnected ? `${AppColors.brand.accent} font-medium` : `${AppColors.status.restarting.text} font-medium`}>
                    {isConnected ? "Connected" : "Idle / Polling"}
                  </span>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      )}

      {}
      {activeTab === "metrics" && (
        <div className="space-y-4">
          {}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            <MetricCard
              title="vCPU Load"
              value={currentMetric ? currentMetric.cpu_usage_pct.toFixed(1) : "0"}
              unit="%"
              percentage={currentMetric?.cpu_usage_pct}
              icon={Cpu}
              color="emerald"
            />
            <MetricCard
              title="RAM Used"
              value={currentMetric ? (currentMetric.memory_used_mb / 1024).toFixed(1) : "0"}
              unit="GB"
              subtitle={currentMetric ? `Total: ${(currentMetric.memory_total_mb / 1024).toFixed(1)} GB` : undefined}
              percentage={currentMetric?.memory_usage_pct}
              icon={HardDrive}
              color="blue"
            />
            <MetricCard
              title="Disk Used"
              value={currentMetric ? currentMetric.disk_used_gb.toFixed(1) : "0"}
              unit="GB"
              subtitle={currentMetric ? `Total: ${currentMetric.disk_total_gb.toFixed(1)} GB` : undefined}
              percentage={currentMetric?.disk_usage_pct}
              icon={HardDrive}
              color="purple"
            />
            <MetricCard
              title="Net Throughput"
              value={currentMetric ? currentMetric.network_in_rate_kbps.toFixed(1) : "0"}
              unit="KB/s"
              subtitle={currentMetric ? `Out: ${currentMetric.network_out_rate_kbps.toFixed(1)} KB/s` : undefined}
              icon={Activity}
              color="amber"
            />
          </div>

          {}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <MetricTimeSeriesChart
              title="vCPU Utilization Percentage"
              description="Beban komputasi prosesor server secara deret waktu"
              data={metricsHistory}
              dataKey="cpu_usage_pct"
              unit="%"
              color="#10b981"
              selectedRange={timeRange}
              onRangeChange={setTimeRange}
              isLive={isConnected}
            />

            <MetricTimeSeriesChart
              title="RAM Memory Utilization"
              description="Persentase alokasi pemakaian memori sistem"
              data={metricsHistory}
              dataKey="memory_usage_pct"
              unit="%"
              color="#3b82f6"
              selectedRange={timeRange}
              onRangeChange={setTimeRange}
              isLive={isConnected}
            />

            <MetricTimeSeriesChart
              title="Disk Capacity Usage"
              description="Persentase kapasitas penyimpanan hard disk"
              data={metricsHistory}
              dataKey="disk_usage_pct"
              unit="%"
              color="#a855f7"
              selectedRange={timeRange}
              onRangeChange={setTimeRange}
              isLive={isConnected}
            />

            <MetricTimeSeriesChart
              title="Network Inbound Throughput"
              description="Laju lalu lintas jaringan masuk (download/inbound)"
              data={metricsHistory}
              dataKey="network_in_rate_kbps"
              unit="KB/s"
              color="#f59e0b"
              selectedRange={timeRange}
              onRangeChange={setTimeRange}
              isLive={isConnected}
            />
          </div>
        </div>
      )}

      {}
      {activeTab === "logs" && (
        <div className="space-y-4">
          <LogViewer
            logs={logs}
            onClearLogs={() => setLogs([])}
            title={`Console Stream: ${server.name} (${server.hostname || "host"})`}
            isStreaming={isConnected}
          />
        </div>
      )}

      {}
      {activeTab === "alerts" && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h3 className={AppText.h4}>Aturan & Insiden Alert Server</h3>
              <p className={AppText.caption}>Peringatan ambang batas yang dikonfigurasi untuk instance ini</p>
            </div>
            <Button
              size="sm"
              onClick={() => setIsCreateRuleOpen(true)}
              className={`${AppColors.brand.primary} text-xs`}
            >
              <Plus className="h-3.5 w-3.5 mr-1" />
              Tambah Aturan Alert
            </Button>
          </div>

          {}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <Card className={AppContainers.card}>
              <CardHeader className={AppContainers.cardHeader}>
                <CardTitle className={AppText.h4}>Insiden Alert Terpicu ({serverAlerts.length})</CardTitle>
                <CardDescription className={AppText.caption}>Riwayat peringatan performa instan</CardDescription>
              </CardHeader>
              <CardContent className="p-4 space-y-3">
                {serverAlerts.length === 0 ? (
                  <p className={`text-center py-6 ${AppText.caption}`}>
                    Tidak ada insiden alert aktif untuk server ini.
                  </p>
                ) : (
                  serverAlerts.map((alert) => (
                    <div
                      key={alert.id}
                      className={`p-3 rounded-lg border ${AppColors.border.subtle} ${AppColors.bg.surfaceSubtle} space-y-2 text-xs`}
                    >
                      <div className="flex items-center justify-between">
                        <span className={`font-semibold ${AppColors.text.primary}`}>{alert.title}</span>
                        <span className={`text-[10px] uppercase font-bold ${AppColors.status.danger.text} ${AppColors.status.danger.bg} px-2 py-0.5 rounded border ${AppColors.status.danger.border}`}>
                          {alert.severity}
                        </span>
                      </div>
                      <p className={AppText.bodySm}>{alert.message}</p>
                      <p className={AppText.codeMuted}>
                        Waktu: {new Date(alert.triggered_at).toLocaleString("id-ID")}
                      </p>
                    </div>
                  ))
                )}
              </CardContent>
            </Card>

            {}
            <Card className={AppContainers.card}>
              <CardHeader className={AppContainers.cardHeader}>
                <CardTitle className={AppText.h4}>Aturan Evaluasi Aktif ({serverRules.length})</CardTitle>
                <CardDescription className={AppText.caption}>Kriteria ambang batas pemantauan otomatis</CardDescription>
              </CardHeader>
              <CardContent className="p-4 space-y-3">
                {serverRules.length === 0 ? (
                  <p className={`text-center py-6 ${AppText.caption}`}>
                    Belum ada aturan threshold khusus yang terdaftar.
                  </p>
                ) : (
                  serverRules.map((rule) => (
                    <div
                      key={rule.id}
                      className={`p-3 rounded-lg border ${AppColors.border.subtle} ${AppColors.bg.surfaceSubtle} flex items-center justify-between text-xs`}
                    >
                      <div>
                        <p className={`font-semibold ${AppColors.text.primary}`}>{rule.name}</p>
                        <p className={AppText.codeMuted}>
                          Kondisi: {rule.metric_name} {rule.operator} {rule.threshold}% ({rule.duration_seconds}s)
                        </p>
                      </div>
                      <span className={`text-[10px] uppercase font-bold ${AppColors.brand.accent} ${AppColors.brand.primaryLight} px-2 py-0.5 rounded`}>
                        {rule.severity}
                      </span>
                    </div>
                  ))
                )}
              </CardContent>
            </Card>
          </div>
        </div>
      )}

      {}
      <ResizeServerModal
        server={server}
        isOpen={isResizeOpen}
        onClose={() => {
          setIsResizeOpen(false);
          fetchServer();
        }}
      />

      {}
      <CreateAlertRuleModal
        serverId={server.id}
        isOpen={isCreateRuleOpen}
        onClose={() => setIsCreateRuleOpen(false)}
        onRuleCreated={fetchAlertsAndRules}
      />
    </div>
  );
}
