"use client";

import React, { useEffect } from "react";
import Link from "next/link";
import {
  Server,
  ShieldCheck,
  Zap,
  Plus,
  ArrowRight,
  HardDrive,
  Cpu,
  RefreshCw,
  ExternalLink,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ServerStatusBadge } from "@/components/server/ServerStatusBadge";
import { useServerStore } from "@/stores/useServerStore";
import { useAuthStore } from "@/stores/useAuthStore";
import { AppContainers, AppText } from "@/core/theme";

export default function OverviewPage() {
  const { servers, totalServers, isLoading, fetchServers } = useServerStore();
  const { user, organization } = useAuthStore();

  useEffect(() => {
    fetchServers(1);
  }, [fetchServers]);

  const runningServers = servers.filter((s) => s.status === "running").length;
  const stoppedServers = servers.filter((s) => s.status === "stopped").length;
  const totalCores = servers.reduce((acc, s) => acc + (s.cpu_cores || 0), 0);
  const totalRAMGB = (servers.reduce((acc, s) => acc + (s.memory_mb || 0), 0) / 1024).toFixed(1);
  const totalDiskGB = servers.reduce((acc, s) => acc + (s.disk_gb || 0), 0);

  return (
    <div className={AppContainers.pageWrapper}>
      {/* Top Greeting & Action Header */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <span className={AppText.categoryTag}>
              {organization?.name || "Workspace"}
            </span>
            <span className="text-[#707070]">•</span>
            <span className={AppText.caption}>Control Panel Overview</span>
          </div>
          <h2 className={AppText.h2}>
            Selamat Datang, {user?.full_name || "Admin"}
          </h2>
          <p className={AppText.subtitle}>
            Ringkasan status infrastruktur server, utilisasi sumber daya, dan skor keamanan Sentinel.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Button
            variant="outline"
            size="sm"
            onClick={() => fetchServers(1)}
            isLoading={isLoading}
            className="cursor-pointer"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            <span>Refresh</span>
          </Button>

          <Link href="/infrastructure/vps">
            <Button size="sm" className="cursor-pointer">
              <Plus className="h-3.5 w-3.5" />
              <span>Deploy Server Baru</span>
            </Button>
          </Link>
        </div>
      </div>

      {/* Aggregate Metric Stats Grid */}
      <div className={AppContainers.metricsGrid}>
        {/* Card 1: Total Servers */}
        <Card className={AppContainers.cardHover}>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <p className={AppText.bodySm}>Total Server VPS</p>
              <div className="p-1.5 rounded-md bg-[#202020] text-emerald-400 border border-[#2e2e2e]">
                <Server className="h-4 w-4" />
              </div>
            </div>
            <div className="mt-2.5 flex items-baseline gap-2">
              <span className="text-2xl font-bold text-[#ededed] dark:text-[#ededed] light:text-[#111827]">{totalServers}</span>
              <span className={AppText.caption}>Instance</span>
            </div>
            <div className="mt-3 flex items-center gap-3 text-xs text-[#a1a1a1] border-t border-[#262626] dark:border-[#262626] light:border-[#e5e7eb] pt-2.5">
              <span className="flex items-center gap-1 text-emerald-400 font-medium">
                <span className="h-1.5 w-1.5 rounded-full bg-emerald-400" />
                {runningServers} Aktif
              </span>
              <span className="text-[#707070]">•</span>
              <span>{stoppedServers} Berhenti</span>
            </div>
          </CardContent>
        </Card>

        {/* Card 2: Total CPU Allocation */}
        <Card className={AppContainers.cardHover}>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <p className={AppText.bodySm}>Alokasi Total vCPU</p>
              <div className="p-1.5 rounded-md bg-[#202020] text-emerald-400 border border-[#2e2e2e]">
                <Cpu className="h-4 w-4" />
              </div>
            </div>
            <div className="mt-2.5 flex items-baseline gap-2">
              <span className="text-2xl font-bold text-[#ededed] dark:text-[#ededed] light:text-[#111827]">{totalCores}</span>
              <span className={AppText.caption}>Cores</span>
            </div>
            <div className="mt-3 text-xs text-[#a1a1a1] border-t border-[#262626] dark:border-[#262626] light:border-[#e5e7eb] pt-2.5 flex items-center justify-between">
              <span>Status CPU:</span>
              <span className="text-emerald-400 font-medium">Optimal</span>
            </div>
          </CardContent>
        </Card>

        {/* Card 3: Memory & Storage Allocation */}
        <Card className={AppContainers.cardHover}>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <p className={AppText.bodySm}>RAM & Disk Terpakai</p>
              <div className="p-1.5 rounded-md bg-[#202020] text-emerald-400 border border-[#2e2e2e]">
                <HardDrive className="h-4 w-4" />
              </div>
            </div>
            <div className="mt-2.5 flex items-baseline gap-2">
              <span className="text-2xl font-bold text-[#ededed] dark:text-[#ededed] light:text-[#111827]">{totalRAMGB} GB</span>
              <span className={AppText.caption}>RAM</span>
            </div>
            <div className="mt-3 text-xs text-[#a1a1a1] border-t border-[#262626] dark:border-[#262626] light:border-[#e5e7eb] pt-2.5 flex items-center justify-between">
              <span>Storage Disk:</span>
              <span className="font-medium text-[#ededed] dark:text-[#ededed] light:text-[#111827]">{totalDiskGB} GB SSD</span>
            </div>
          </CardContent>
        </Card>

        {/* Card 4: Sentinel Security Score */}
        <Card className={AppContainers.cardHover}>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <p className={AppText.bodySm}>Sentinel Security Score</p>
              <div className="p-1.5 rounded-md bg-[#202020] text-emerald-400 border border-[#2e2e2e]">
                <ShieldCheck className="h-4 w-4" />
              </div>
            </div>
            <div className="mt-2.5 flex items-baseline gap-2">
              <span className="text-2xl font-bold text-emerald-400">92</span>
              <span className={AppText.caption}>/ 100 (Aman)</span>
            </div>
            <div className="mt-3 text-xs text-[#a1a1a1] border-t border-[#262626] dark:border-[#262626] light:border-[#e5e7eb] pt-2.5 flex items-center justify-between">
              <span>Port Exposure:</span>
              <span className="text-emerald-400 font-medium">Terkunci Aman</span>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Main Content Split: Server List Summary & System Health */}
      <div className={AppContainers.overviewSplitGrid}>
        {/* Left 2 Cols: Recent Servers */}
        <div className="lg:col-span-2 space-y-3">
          <div className="flex items-center justify-between">
            <div>
              <h3 className={AppText.h3}>Daftar Server Terkini</h3>
              <p className={AppText.subtitle}>Instance server aktif yang terhubung pada workspace</p>
            </div>
            <Link
              href="/infrastructure/vps"
              className="text-xs font-medium text-emerald-400 hover:text-emerald-300 flex items-center gap-1"
            >
              <span>Kelola Semua</span>
              <ArrowRight className="h-3.5 w-3.5" />
            </Link>
          </div>

          <Card className="overflow-hidden">
            {servers.length === 0 ? (
              <div className="py-12 text-center text-[#707070]">
                <Server className="h-8 w-8 mx-auto text-[#404040] mb-2" />
                <p className="text-sm font-medium text-[#ededed] dark:text-[#ededed] light:text-[#111827]">Belum ada server yang terdaftar</p>
                <p className={AppText.subtitle}>
                  Mulai dengan membuat instance server VPS baru menggunakan provider simulasi atau provider cloud Anda.
                </p>
                <Link href="/infrastructure/vps" className="inline-block mt-4">
                  <Button size="sm">
                    <Plus className="h-3.5 w-3.5" />
                    <span>Deploy Server Pertama</span>
                  </Button>
                </Link>
              </div>
            ) : (
              <div className="divide-y divide-[#262626] dark:divide-[#262626] light:divide-[#e5e7eb]">
                {servers.slice(0, 5).map((server) => (
                  <div
                    key={server.id}
                    className="p-4 flex items-center justify-between hover:bg-[#1a1a1a] dark:hover:bg-[#1a1a1a] light:hover:bg-[#f3f4f6] transition-colors"
                  >
                    <div className="flex items-center gap-3">
                      <div className="h-8 w-8 rounded-lg bg-[#202020] flex items-center justify-center text-[#ededed] border border-[#2e2e2e]">
                        <Server className="h-4 w-4 text-emerald-400" />
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="font-semibold text-xs text-[#ededed] dark:text-[#ededed] light:text-[#111827]">{server.name}</span>
                          <Badge variant="outline" className="text-[10px] py-0 px-1.5">
                            {server.provider?.name || "Mock Provider"}
                          </Badge>
                        </div>
                        <p className={AppText.codeMuted}>
                          {server.ip_address || "Alokasi IP..."} • {server.region}
                        </p>
                      </div>
                    </div>

                    <div className="flex items-center gap-4">
                      <div className="text-right hidden sm:block">
                        <p className="text-xs font-medium text-[#ededed] dark:text-[#ededed] light:text-[#111827]">
                          {server.cpu_cores} vCPU • {(server.memory_mb / 1024).toFixed(0)} GB RAM
                        </p>
                        <p className={AppText.caption}>{server.os_type}</p>
                      </div>

                      <ServerStatusBadge status={server.status} />

                      <Link href={`/infrastructure/vps/${server.id}`}>
                        <button
                          type="button"
                          aria-label="Lihat detail server"
                          className="p-1.5 text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#222222] rounded-md transition-colors cursor-pointer"
                        >
                          <ExternalLink className="h-4 w-4" />
                        </button>
                      </Link>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </Card>
        </div>

        {/* Right 1 Col: Sentinel & System Status */}
        <div className="space-y-4">
          <Card>
            <CardHeader className="pb-3">
              <div className="flex items-center gap-2">
                <ShieldCheck className="h-4 w-4 text-emerald-400" />
                <CardTitle className="text-xs">Sentinel Security Scanner</CardTitle>
              </div>
              <CardDescription>Postur asesmen keamanan infrastruktur otomatis</CardDescription>
            </CardHeader>
            <CardContent className="space-y-2.5 pt-2 text-xs">
              <div className="flex items-center justify-between py-1.5 border-b border-[#262626] dark:border-[#262626] light:border-[#e5e7eb]">
                <span className="text-[#a1a1a1]">SSL/TLS Certificates</span>
                <span className="text-emerald-400 font-medium">Valid (A+)</span>
              </div>
              <div className="flex items-center justify-between py-1.5 border-b border-[#262626] dark:border-[#262626] light:border-[#e5e7eb]">
                <span className="text-[#a1a1a1]">SSH Key Access</span>
                <span className="text-emerald-400 font-medium">Enkripsi AES-256</span>
              </div>
              <div className="flex items-center justify-between py-1.5 border-b border-[#262626] dark:border-[#262626] light:border-[#e5e7eb]">
                <span className="text-[#a1a1a1]">Security Headers</span>
                <span className="text-emerald-400 font-medium">Standard Compliance</span>
              </div>
              <div className="flex items-center justify-between py-1.5">
                <span className="text-[#a1a1a1]">Audit Logging</span>
                <span className="text-emerald-400 font-medium">Aktif Merekam</span>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <div className="flex items-center gap-2 text-emerald-400">
                <Zap className="h-4 w-4" />
                <CardTitle className="text-xs">Automation Engine</CardTitle>
              </div>
              <CardDescription>Event-driven automatic actions</CardDescription>
            </CardHeader>
            <CardContent className="space-y-2 pt-2 text-xs">
              <p className="text-[#ededed] dark:text-[#ededed] light:text-[#111827] font-medium">Auto-Restart VPS on High Load</p>
              <p className={AppText.caption}>
                Secara otomatis me-reboot server jika utilisasi RAM &gt; 95% selama 10 menit berturut-turut.
              </p>
              <div className="pt-1">
                <Badge variant="outline" className="text-emerald-400 border-emerald-800/40 bg-emerald-950/30 text-[10px]">
                  Rule Enabled (Active)
                </Badge>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
