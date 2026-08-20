"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import {
  Server as ServerIcon,
  Plus,
  RefreshCw,
  Search,
  Sliders,
  Power,
  PowerOff,
  Trash2,
  Copy,
  Check,
  ExternalLink,
  RotateCcw,
} from "lucide-react";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ServerStatusBadge } from "@/components/server/ServerStatusBadge";
import { CreateServerModal } from "@/components/server/CreateServerModal";
import { ResizeServerModal } from "@/components/server/ResizeServerModal";
import { useServerStore } from "@/stores/useServerStore";
import { Server } from "@/types/server";
import { AppContainers, AppText } from "@/core/theme";

export default function VPSManagementPage() {
  const {
    servers,
    totalServers,
    currentPage,
    totalPages,
    isLoading,
    fetchServers,
    rebootServer,
    shutdownServer,
    startServer,
    deleteServer,
  } = useServerStore();

  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [resizeTarget, setResizeTarget] = useState<Server | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [copiedIP, setCopiedIP] = useState<string | null>(null);
  const [activeActionID, setActiveActionID] = useState<string | null>(null);

  useEffect(() => {
    fetchServers(1);
  }, [fetchServers]);

  const handleCopyIP = (ip?: string) => {
    if (!ip) return;
    navigator.clipboard.writeText(ip);
    setCopiedIP(ip);
    setTimeout(() => setCopiedIP(null), 2000);
  };

  const filteredServers = servers.filter((s) => {
    const matchesSearch =
      s.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      Boolean(s.ip_address?.includes(searchQuery));
    const matchesStatus = statusFilter === "all" || s.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const handleAction = async (id: string, actionFn: (id: string) => Promise<void>) => {
    setActiveActionID(id);
    try {
      await actionFn(id);
    } finally {
      setActiveActionID(null);
    }
  };

  return (
    <div className={AppContainers.pageWrapper}>
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
        <div>
          <h2 className={AppText.h2}>
            Manajemen VPS & Virtual Servers
          </h2>
          <p className={AppText.subtitle}>
            Kelola, monitor, dan orkestrasi seluruh instance server VPS dari satu panel terpadu.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Button
            variant="outline"
            size="sm"
            onClick={() => fetchServers(currentPage)}
            isLoading={isLoading}
            className="cursor-pointer"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            <span>Refresh</span>
          </Button>

          <Button size="sm" onClick={() => setIsCreateOpen(true)} className="cursor-pointer">
            <Plus className="h-3.5 w-3.5" />
            <span>Deploy Server Baru</span>
          </Button>
        </div>
      </div>

      {/* Filter & Search Toolbar */}
      <div className="flex flex-col sm:flex-row items-center gap-3">
        <div className="relative flex-1 w-full">
          <Search className="h-3.5 w-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-[#707070]" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Cari berdasarkan nama server atau alamat IP..."
            className="w-full pl-9 pr-4 py-2 text-xs rounded-lg border border-[#262626] dark:border-[#262626] light:border-[#d1d5db] bg-[#121212] dark:bg-[#121212] light:bg-[#ffffff] text-[#ededed] dark:text-[#ededed] light:text-[#111827] placeholder-[#707070] focus:border-emerald-500 focus:outline-none transition-colors"
          />
        </div>

        <div className="flex items-center gap-2 w-full sm:w-auto">
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="px-3 py-2 text-xs rounded-lg border border-[#262626] dark:border-[#262626] light:border-[#d1d5db] bg-[#121212] dark:bg-[#121212] light:bg-[#ffffff] text-[#ededed] dark:text-[#ededed] light:text-[#111827] focus:border-emerald-500 focus:outline-none"
          >
            <option value="all">Semua Status</option>
            <option value="running">Running</option>
            <option value="stopped">Stopped</option>
            <option value="restarting">Restarting</option>
          </select>
        </div>
      </div>

      {/* Server Table Card */}
      <Card className="overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs text-[#ededed] dark:text-[#ededed] light:text-[#111827]">
            <thead className="border-b border-[#262626] dark:border-[#262626] light:border-[#e5e7eb] bg-[#141414] dark:bg-[#141414] light:bg-[#f4f4f5] text-[11px] font-semibold uppercase tracking-wider text-[#a1a1a1]">
              <tr>
                <th className="px-5 py-3.5">Nama Instance</th>
                <th className="px-5 py-3.5">Status</th>
                <th className="px-5 py-3.5">Provider</th>
                <th className="px-5 py-3.5">Alamat IP</th>
                <th className="px-5 py-3.5">Spesifikasi</th>
                <th className="px-5 py-3.5">Region & OS</th>
                <th className="px-5 py-3.5 text-right">Aksi Kontrol</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#262626] dark:divide-[#262626] light:divide-[#e5e7eb]">
              {filteredServers.length === 0 ? (
                <tr>
                  <td colSpan={7} className="text-center py-12 text-[#707070]">
                    <ServerIcon className="h-8 w-8 mx-auto text-[#404040] mb-2" />
                    {searchQuery
                      ? "Tidak ada server yang cocok dengan kata kunci pencarian"
                      : "Belum ada server yang terdaftar. Klik 'Deploy Server Baru' untuk memulai."}
                  </td>
                </tr>
              ) : (
                filteredServers.map((server) => {
                  const isActionLoading = activeActionID === server.id;

                  return (
                    <tr key={server.id} className="hover:bg-[#1a1a1a] dark:hover:bg-[#1a1a1a] light:hover:bg-[#f3f4f6] transition-colors">
                      {/* Name */}
                      <td className="px-5 py-4 font-medium">
                        <Link
                          href={`/infrastructure/vps/${server.id}`}
                          className="hover:text-emerald-400 transition-colors flex items-center gap-1.5 font-semibold text-xs text-[#ededed] dark:text-[#ededed] light:text-[#111827]"
                        >
                          <span>{server.name}</span>
                          <ExternalLink className="h-3 w-3 text-[#707070]" />
                        </Link>
                        <span className={AppText.caption}>
                          ID: {server.id.slice(0, 8)}...
                        </span>
                      </td>

                      {/* Status */}
                      <td className="px-5 py-4">
                        <ServerStatusBadge status={server.status} />
                      </td>

                      {/* Provider */}
                      <td className="px-5 py-4">
                        <Badge variant="outline" className="text-[10px] py-0 px-2">
                          {server.provider?.name || "Mock Provider"}
                        </Badge>
                      </td>

                      {/* IP Address */}
                      <td className="px-5 py-4 font-mono">
                        {server.ip_address ? (
                          <button
                            type="button"
                            onClick={() => handleCopyIP(server.ip_address)}
                            className="flex items-center gap-1.5 text-[#a1a1a1] hover:text-emerald-400 transition-colors cursor-pointer"
                            title="Salin IP Address"
                          >
                            <span>{server.ip_address}</span>
                            {copiedIP === server.ip_address ? (
                              <Check className="h-3 w-3 text-emerald-400" />
                            ) : (
                              <Copy className="h-3 w-3 text-[#707070]" />
                            )}
                          </button>
                        ) : (
                          <span className="text-[#707070] text-[11px]">-</span>
                        )}
                      </td>

                      {/* Hardware Specs */}
                      <td className="px-5 py-4">
                        <div className="font-medium text-[#ededed] dark:text-[#ededed] light:text-[#111827]">
                          {server.cpu_cores} vCPU • {(server.memory_mb / 1024).toFixed(0)} GB RAM
                        </div>
                        <span className={AppText.caption}>{server.disk_gb} GB SSD</span>
                      </td>

                      {/* Region & OS */}
                      <td className="px-5 py-4">
                        <div className="text-[#ededed] dark:text-[#ededed] light:text-[#111827]">{server.region}</div>
                        <span className={AppText.caption}>{server.os_type}</span>
                      </td>

                      {/* Actions */}
                      <td className="px-5 py-4 text-right">
                        <div className="flex items-center justify-end gap-1.5">
                          {/* Reboot Button */}
                          <button
                            type="button"
                            disabled={isActionLoading || server.status !== "running"}
                            onClick={() => handleAction(server.id, rebootServer)}
                            className="p-1.5 rounded-md text-[#a1a1a1] hover:text-emerald-400 hover:bg-[#202020] disabled:opacity-30 disabled:cursor-not-allowed transition-colors cursor-pointer"
                            title="Reboot Server"
                          >
                            <RotateCcw className="h-3.5 w-3.5" />
                          </button>

                          {/* Shutdown / Start Toggle */}
                          {server.status === "running" ? (
                            <button
                              type="button"
                              disabled={isActionLoading}
                              onClick={() => handleAction(server.id, shutdownServer)}
                              className="p-1.5 rounded-md text-[#a1a1a1] hover:text-amber-400 hover:bg-[#202020] disabled:opacity-30 transition-colors cursor-pointer"
                              title="Shutdown Server"
                            >
                              <PowerOff className="h-3.5 w-3.5" />
                            </button>
                          ) : (
                            <button
                              type="button"
                              disabled={isActionLoading}
                              onClick={() => handleAction(server.id, startServer)}
                              className="p-1.5 rounded-md text-[#a1a1a1] hover:text-emerald-400 hover:bg-[#202020] disabled:opacity-30 transition-colors cursor-pointer"
                              title="Start Server"
                            >
                              <Power className="h-3.5 w-3.5" />
                            </button>
                          )}

                          {/* Resize Button */}
                          <button
                            type="button"
                            disabled={isActionLoading}
                            onClick={() => setResizeTarget(server)}
                            className="p-1.5 rounded-md text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#202020] transition-colors cursor-pointer"
                            title="Resize Spesifikasi"
                          >
                            <Sliders className="h-3.5 w-3.5" />
                          </button>

                          {/* Delete Button */}
                          <button
                            type="button"
                            disabled={isActionLoading}
                            onClick={() => {
                              if (confirm(`Apakah Anda yakin ingin menterminasi server ${server.name}?`)) {
                                handleAction(server.id, deleteServer);
                              }
                            }}
                            className="p-1.5 rounded-md text-[#a1a1a1] hover:text-rose-400 hover:bg-rose-950/40 transition-colors cursor-pointer"
                            title="Hapus / Terminate Server"
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination Footer */}
        {totalPages > 1 && (
          <div className="flex items-center justify-between p-4 border-t border-[#262626] dark:border-[#262626] light:border-[#e5e7eb] bg-[#141414] dark:bg-[#141414] light:bg-[#f4f4f5] text-xs text-[#a1a1a1]">
            <span>
              Menampilkan Halaman {currentPage} dari {totalPages} ({totalServers} Total Server)
            </span>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={currentPage <= 1 || isLoading}
                onClick={() => fetchServers(currentPage - 1)}
              >
                Sebelumnya
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={currentPage >= totalPages || isLoading}
                onClick={() => fetchServers(currentPage + 1)}
              >
                Selanjutnya
              </Button>
            </div>
          </div>
        )}
      </Card>

      {/* Modals */}
      <CreateServerModal isOpen={isCreateOpen} onClose={() => setIsCreateOpen(false)} />
      <ResizeServerModal
        server={resizeTarget}
        isOpen={Boolean(resizeTarget)}
        onClose={() => setResizeTarget(null)}
      />
    </div>
  );
}
