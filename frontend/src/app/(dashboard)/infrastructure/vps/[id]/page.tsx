"use client";

import React, { useState, useEffect, use } from "react";
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
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { ServerStatusBadge } from "@/components/server/ServerStatusBadge";
import { ResizeServerModal } from "@/components/server/ResizeServerModal";
import { serverService } from "@/services/server.service";
import { Server } from "@/types/server";
import { AppContainers, AppText } from "@/core/theme";

export default function ServerDetailPage({
  params,
}: Readonly<{
  params: Promise<{ id: string }>;
}>) {
  const resolvedParams = use(params);
  const router = useRouter();
  const serverId = resolvedParams.id;

  const [server, setServer] = useState<Server | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isActionLoading, setIsActionLoading] = useState(false);
  const [isResizeOpen, setIsResizeOpen] = useState(false);
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState<string | null>(null);

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

  useEffect(() => {
    fetchServer();
  }, [serverId]);

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

  if (isLoading) {
    return (
      <div className="flex h-96 items-center justify-center text-[#a1a1a1]">
        <div className="flex flex-col items-center gap-3">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-emerald-500 border-t-transparent" />
          <p className="text-xs">Memuat detail server VPS...</p>
        </div>
      </div>
    );
  }

  if (error || !server) {
    return (
      <div className="max-w-4xl mx-auto py-12 text-center">
        <p className="text-rose-400 text-sm font-medium">{error || "Server tidak ditemukan"}</p>
        <Link href="/infrastructure/vps" className="inline-block mt-4">
          <Button variant="outline" size="sm">
            <ArrowLeft className="h-4 w-4 mr-1.5" />
            Kembali ke Daftar Server
          </Button>
        </Link>
      </div>
    );
  }

  return (
    <div className={AppContainers.pageWrapper}>
      {/* Back Link & Header */}
      <div>
        <Link
          href="/infrastructure/vps"
          className="inline-flex items-center gap-1.5 text-xs text-[#a1a1a1] hover:text-[#ededed] transition-colors mb-3"
        >
          <ArrowLeft className="h-3.5 w-3.5" />
          <span>Kembali ke Daftar VPS</span>
        </Link>

        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-lg bg-[#202020] border border-[#2e2e2e] flex items-center justify-center text-emerald-400">
              <ServerIcon className="h-5 w-5" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h2 className={AppText.h2}>{server.name}</h2>
                <ServerStatusBadge status={server.status} />
              </div>
              <p className={AppText.codeMuted}>
                ID: {server.id} • {server.provider?.name || "Mock Provider"}
              </p>
            </div>
          </div>

          {/* Action Toolbar */}
          <div className="flex items-center flex-wrap gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={isActionLoading || server.status !== "running"}
              onClick={() => handleAction(serverService.rebootServer)}
            >
              <RotateCcw className="h-3.5 w-3.5 text-emerald-400" />
              <span>Reboot</span>
            </Button>

            {server.status === "running" ? (
              <Button
                variant="outline"
                size="sm"
                disabled={isActionLoading}
                onClick={() => handleAction(serverService.shutdownServer)}
              >
                <PowerOff className="h-3.5 w-3.5 text-amber-400" />
                <span>Shutdown</span>
              </Button>
            ) : (
              <Button
                variant="outline"
                size="sm"
                disabled={isActionLoading}
                onClick={() => handleAction(serverService.startServer)}
              >
                <Power className="h-3.5 w-3.5 text-emerald-400" />
                <span>Start</span>
              </Button>
            )}

            <Button
              variant="outline"
              size="sm"
              disabled={isActionLoading}
              onClick={() => setIsResizeOpen(true)}
            >
              <Sliders className="h-3.5 w-3.5 text-[#a1a1a1]" />
              <span>Resize</span>
            </Button>

            <Button
              variant="danger"
              size="sm"
              disabled={isActionLoading}
              onClick={handleDelete}
            >
              <Trash2 className="h-3.5 w-3.5" />
              <span>Terminate</span>
            </Button>
          </div>
        </div>
      </div>

      {/* Main Specs & Details Grid */}
      <div className={AppContainers.overviewSplitGrid}>
        {/* Hardware & Network Specs */}
        <div className="lg:col-span-2 space-y-4">
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-xs font-semibold">Spesifikasi Komputasi</CardTitle>
              <CardDescription>Rincian alokasi hardware yang diprovisioning</CardDescription>
            </CardHeader>
            <CardContent>
              <div className={AppContainers.specsGrid}>
                <div className="p-3.5 rounded-lg bg-[#121212] dark:bg-[#121212] light:bg-[#f9fafb] border border-[#262626] dark:border-[#262626] light:border-[#e5e7eb]">
                  <div className="flex items-center gap-2 text-[#a1a1a1] mb-1">
                    <Cpu className="h-4 w-4 text-emerald-400" />
                    <span>vCPU Cores</span>
                  </div>
                  <p className="text-base font-bold text-[#ededed] dark:text-[#ededed] light:text-[#111827]">{server.cpu_cores} Cores</p>
                </div>

                <div className="p-3.5 rounded-lg bg-[#121212] dark:bg-[#121212] light:bg-[#f9fafb] border border-[#262626] dark:border-[#262626] light:border-[#e5e7eb]">
                  <div className="flex items-center gap-2 text-[#a1a1a1] mb-1">
                    <HardDrive className="h-4 w-4 text-emerald-400" />
                    <span>Memory RAM</span>
                  </div>
                  <p className="text-base font-bold text-[#ededed] dark:text-[#ededed] light:text-[#111827]">
                    {(server.memory_mb / 1024).toFixed(0)} GB RAM
                  </p>
                  <p className="text-[10px] text-[#707070] font-mono mt-0.5">{server.memory_mb} MB</p>
                </div>

                <div className="p-3.5 rounded-lg bg-[#121212] dark:bg-[#121212] light:bg-[#f9fafb] border border-[#262626] dark:border-[#262626] light:border-[#e5e7eb]">
                  <div className="flex items-center gap-2 text-[#a1a1a1] mb-1">
                    <HardDrive className="h-4 w-4 text-emerald-400" />
                    <span>SSD Storage</span>
                  </div>
                  <p className="text-base font-bold text-[#ededed] dark:text-[#ededed] light:text-[#111827]">{server.disk_gb} GB</p>
                </div>

                <div className="p-3.5 rounded-lg bg-[#121212] dark:bg-[#121212] light:bg-[#f9fafb] border border-[#262626] dark:border-[#262626] light:border-[#e5e7eb]">
                  <div className="flex items-center gap-2 text-[#a1a1a1] mb-1">
                    <Globe className="h-4 w-4 text-emerald-400" />
                    <span>Data Center Region</span>
                  </div>
                  <p className="text-xs font-semibold text-[#ededed] dark:text-[#ededed] light:text-[#111827]">{server.region}</p>
                </div>

                <div className="p-3.5 rounded-lg bg-[#121212] dark:bg-[#121212] light:bg-[#f9fafb] border border-[#262626] dark:border-[#262626] light:border-[#e5e7eb]">
                  <div className="flex items-center gap-2 text-[#a1a1a1] mb-1">
                    <Activity className="h-4 w-4 text-emerald-400" />
                    <span>Sistem Operasi</span>
                  </div>
                  <p className="text-xs font-semibold text-[#ededed] dark:text-[#ededed] light:text-[#111827]">{server.os_type}</p>
                </div>

                <div className="p-3.5 rounded-lg bg-[#121212] dark:bg-[#121212] light:bg-[#f9fafb] border border-[#262626] dark:border-[#262626] light:border-[#e5e7eb]">
                  <div className="flex items-center gap-2 text-[#a1a1a1] mb-1">
                    <Calendar className="h-4 w-4 text-emerald-400" />
                    <span>Tanggal Dibuat</span>
                  </div>
                  <p className="text-xs font-medium text-[#ededed] dark:text-[#ededed] light:text-[#111827]">
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

          {/* Network & Host Information */}
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-xs font-semibold">Jaringan & Host Endpoint</CardTitle>
              <CardDescription>Alamat IP publik dan identifier host instance</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3 text-xs">
              <div className="flex items-center justify-between p-3 rounded-lg bg-[#121212] dark:bg-[#121212] light:bg-[#f9fafb] border border-[#262626] dark:border-[#262626] light:border-[#e5e7eb]">
                <div>
                  <p className="text-[11px] text-[#a1a1a1]">Public IPv4 Address</p>
                  <p className="text-sm font-mono font-semibold text-[#ededed] dark:text-[#ededed] light:text-[#111827] mt-0.5">
                    {server.ip_address || "Sedang dialokasikan..."}
                  </p>
                </div>
                {server.ip_address && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handleCopy(server.ip_address)}
                  >
                    {copied ? <Check className="h-3.5 w-3.5 text-emerald-400" /> : <Copy className="h-3.5 w-3.5" />}
                    <span>{copied ? "Tersalin" : "Salin IP"}</span>
                  </Button>
                )}
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div className="p-3 rounded-lg bg-[#121212] dark:bg-[#121212] light:bg-[#f9fafb] border border-[#262626] dark:border-[#262626] light:border-[#e5e7eb]">
                  <p className="text-[11px] text-[#a1a1a1]">External Provider ID</p>
                  <p className="text-xs font-mono text-[#ededed] dark:text-[#ededed] light:text-[#111827] mt-0.5">{server.external_server_id || "-"}</p>
                </div>
                <div className="p-3 rounded-lg bg-[#121212] dark:bg-[#121212] light:bg-[#f9fafb] border border-[#262626] dark:border-[#262626] light:border-[#e5e7eb]">
                  <p className="text-[11px] text-[#a1a1a1]">Hostname</p>
                  <p className="text-xs font-mono text-[#ededed] dark:text-[#ededed] light:text-[#111827] mt-0.5">{server.hostname || server.name}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Right 1 Col: Monitoring Preview & Sentinel */}
        <div className="space-y-4">
          <Card>
            <CardHeader className="pb-3">
              <div className="flex items-center gap-2">
                <Activity className="h-4 w-4 text-emerald-400" />
                <CardTitle className="text-xs font-semibold">Telemetry Monitoring</CardTitle>
              </div>
              <CardDescription>Estimasi beban sumber daya instance</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4 text-xs">
              <div>
                <div className="flex justify-between mb-1 text-[#a1a1a1]">
                  <span>vCPU Utilization</span>
                  <span className="font-mono font-semibold text-emerald-400">18%</span>
                </div>
                <div className="w-full bg-[#202020] rounded-full h-1.5">
                  <div className="bg-emerald-500 h-1.5 rounded-full w-[18%]" />
                </div>
              </div>

              <div>
                <div className="flex justify-between mb-1 text-[#a1a1a1]">
                  <span>RAM Memory Load</span>
                  <span className="font-mono font-semibold text-emerald-400">32% (1.3 / 4.0 GB)</span>
                </div>
                <div className="w-full bg-[#202020] rounded-full h-1.5">
                  <div className="bg-emerald-500 h-1.5 rounded-full w-[32%]" />
                </div>
              </div>

              <div>
                <div className="flex justify-between mb-1 text-[#a1a1a1]">
                  <span>SSD Disk Usage</span>
                  <span className="font-mono font-semibold text-emerald-400">24% (12 / 50 GB)</span>
                </div>
                <div className="w-full bg-[#202020] rounded-full h-1.5">
                  <div className="bg-emerald-500 h-1.5 rounded-full w-[24%]" />
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <div className="flex items-center gap-2 text-emerald-400">
                <ShieldCheck className="h-4 w-4" />
                <CardTitle className="text-xs font-semibold">Sentinel Status</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-2 text-xs">
              <div className="flex items-center justify-between py-1 border-b border-[#262626] dark:border-[#262626] light:border-[#e5e7eb]">
                <span className="text-[#a1a1a1]">Port Firewall</span>
                <span className="text-emerald-400 font-medium">Port 22, 80, 443</span>
              </div>
              <div className="flex items-center justify-between py-1">
                <span className="text-[#a1a1a1]">Security Health</span>
                <span className="text-emerald-400 font-medium">Passed (100%)</span>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      <ResizeServerModal
        server={server}
        isOpen={isResizeOpen}
        onClose={() => {
          setIsResizeOpen(false);
          fetchServer();
        }}
      />
    </div>
  );
}
