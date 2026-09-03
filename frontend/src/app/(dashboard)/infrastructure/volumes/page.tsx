"use client";

import React, { useState, useEffect } from "react";
import {
  HardDrive,
  Plus,
  Trash2,
  CheckCircle2,
  AlertCircle,
  Search,
  Layers,
  Database,
  ExternalLink,
  Lock,
  ArrowRight,
  RefreshCw,
  Sliders,
  Server,
  Activity,
} from "lucide-react";
import { Dialog } from "@/components/ui/dialog";
import { AppTheme } from "@/core/theme";
import { volumeService, StorageVolume, StoragePoolStats } from "@/services/volume.service";

export default function VolumesManagementPage() {
  const [volumes, setVolumes] = useState<StorageVolume[]>([]);
  const [stats, setStats] = useState<StoragePoolStats | null>(null);
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [servers, setServers] = useState<any[]>([]);
  const [selectedServerFilter, setSelectedServerFilter] = useState<string>("all");
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

  const [isCreateModalOpen, setIsCreateModalOpen] = useState<boolean>(false);
  const [isResizeModalOpen, setIsResizeModalOpen] = useState<boolean>(false);
  const [selectedVolume, setSelectedVolume] = useState<StorageVolume | null>(null);

  const [volServerId, setVolServerId] = useState<string>("");
  const [volName, setVolName] = useState<string>("");
  const [volType, setVolType] = useState<"nvme" | "ssd" | "docker-volume">("nvme");
  const [volSize, setVolSize] = useState<number>(1);
  const [volFs, setVolFs] = useState<"ext4" | "xfs" | "btrfs">("ext4");
  const [volMount, setVolMount] = useState<string>("/mnt/data");

  const loadData = async () => {
    try {
      setIsLoading(true);
      setErrorMsg(null);
      const [vols, poolStats, serversRes] = await Promise.all([
        volumeService.listVolumes(),
        volumeService.getStoragePoolStats().catch(() => null),
        import("@/services/server.service").then((m) => m.serverService.listServers().catch(() => ({ data: [] }))),
      ]);
      setVolumes(vols || []);
      if (poolStats) setStats(poolStats);
      if (serversRes && serversRes.data) {
        setServers(serversRes.data);
      }
    } catch (err: any) {
      console.error("Gagal memuat volume:", err);
      setErrorMsg(err.response?.data?.message || "Gagal memuat daftar volume dari server");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleCreateVolume = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!volName.trim()) return;

    const selectedTargetServer = servers.find((s: any) => s.id === volServerId);
    
    let maxAllowedGB = 75;
    if (selectedTargetServer) {
      const totalGB = selectedTargetServer.disk_gb || selectedTargetServer.diskGb || 25;
      const usedGB = volumes
        .filter((v: any) => v.server_id === selectedTargetServer.id || v.serverId === selectedTargetServer.id)
        .reduce((sum: number, v: any) => sum + (Number(v.size_gb || v.sizeGb) || 0), 0);
      maxAllowedGB = Math.max(0, totalGB - usedGB);
    } else if (stats) {
      maxAllowedGB = Math.floor(stats.free_gb);
    }

    if (!selectedTargetServer && stats && volSize > stats.free_gb) {
      setErrorMsg(`Kapasitas (${volSize} GB) melebihi ruang bebas disk fisik host lokal (${stats.free_gb.toFixed(1)} GB tersedia).`);
      return;
    }
    if (selectedTargetServer && volSize > maxAllowedGB) {
      setErrorMsg(`Kapasitas (${volSize} GB) melebihi sisa ruang bebas server target ${selectedTargetServer.name} (${maxAllowedGB} GB sisa tersedia).`);
      return;
    }

    try {
      setIsSubmitting(true);
      setErrorMsg(null);
      await volumeService.createVolume({
        server_id: volServerId ? volServerId : null,
        name: volName.trim(),
        size_gb: Number(volSize),
        type: volType,
        fs_type: volFs,
        mount_path: volMount.trim() || "/mnt/data",
      });

      setSuccessMsg(`Volume "${volName}" berhasil dibuat secara fisik.`);
      setVolName("");
      setVolSize(1);
      setIsCreateModalOpen(false);
      await loadData();
    } catch (err: any) {
      setErrorMsg(err.response?.data?.message || "Gagal membuat volume fisik");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDeleteVolume = async (id: string, name: string) => {
    if (!confirm(`Apakah Anda yakin ingin menghapus volume "${name}" secara permanen dari server? Data di dalamnya akan dihapus.`)) {
      return;
    }

    try {
      setErrorMsg(null);
      await volumeService.deleteVolume(id);
      setSuccessMsg(`Volume "${name}" berhasil dihapus.`);
      await loadData();
    } catch (err: any) {
      setErrorMsg(err.response?.data?.message || "Gagal menghapus volume");
    }
  };

  const [selectedVolumeDetail, setSelectedVolumeDetail] = useState<StorageVolume | null>(null);

  const serverScopedVolumes = volumes.filter((v) => {
    if (selectedServerFilter === "all") return true;
    if (selectedServerFilter === "local") return !v.server_id;
    return v.server_id === selectedServerFilter;
  });

  const filteredVolumes = serverScopedVolumes.filter((v) =>
    v.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    v.mount_path?.toLowerCase().includes(searchQuery.toLowerCase()) ||
    (v.attached_container_name && v.attached_container_name.toLowerCase().includes(searchQuery.toLowerCase()))
  );

  const activeServer = servers.find((s: any) => s.id === selectedServerFilter);
  const currentAllocatedGB = serverScopedVolumes.reduce((acc, v) => acc + (v.size_gb || v.sizeGB || 0), 0);
  
  const activePhysicalDiskGB = activeServer
    ? (activeServer.disk_gb || activeServer.diskGb || 25)
    : (stats ? stats.total_gb : 85);

  const activeFreeDiskGB = activeServer
    ? Math.max(0, activePhysicalDiskGB - currentAllocatedGB)
    : (stats ? stats.free_gb : 75);

  const activeUsedDiskGB = activeServer
    ? currentAllocatedGB
    : (stats ? stats.used_gb : 10);

  const activeUsagePercent = activePhysicalDiskGB > 0
    ? (activeUsedDiskGB / activePhysicalDiskGB) * 100
    : 0;

  return (
    <div className="space-y-6">
      {}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-100 flex items-center gap-2.5">
            <HardDrive className="w-6 h-6 text-emerald-400" />
            Persistent Storage & Block Volumes
          </h1>
          <p className="text-sm text-slate-400 mt-1">
            Kelola Block Storage Volumes, Docker Persistent Volumes, dan partisi disk server fisik secara persisten.
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2.5">
          {}
          <div className="flex items-center gap-2 bg-[#11141a] border border-[#22272e] rounded-lg px-3 py-1.5">
            <Server className="w-3.5 h-3.5 text-emerald-400" />
            <select
              value={selectedServerFilter}
              onChange={(e) => setSelectedServerFilter(e.target.value)}
              className="bg-transparent text-xs text-slate-200 focus:outline-none font-mono cursor-pointer"
            >
              <option value="all" className="bg-[#161b22] text-slate-200">All Server Nodes ({volumes.length} Volumes)</option>
              <option value="local" className="bg-[#161b22] text-slate-200">Local Host ({volumes.filter(v => !v.server_id).length} Volumes)</option>
              {servers.map((s: any) => (
                <option key={s.id} value={s.id} className="bg-[#161b22] text-slate-200">
                  {s.name} ({volumes.filter(v => v.server_id === s.id).length} Volumes)
                </option>
              ))}
            </select>
          </div>

          <button
            onClick={loadData}
            disabled={isLoading}
            className="p-2 text-slate-400 hover:text-slate-200 bg-[#161b22] border border-slate-800 rounded-lg hover:border-slate-700 transition-colors cursor-pointer"
            title="Refresh Volume Data"
          >
            <RefreshCw className={`w-4 h-4 ${isLoading ? "animate-spin text-emerald-400" : ""}`} />
          </button>
          <button
            onClick={() => setIsCreateModalOpen(true)}
            className="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-slate-950 bg-emerald-400 hover:bg-emerald-300 rounded-lg transition-colors shadow-sm cursor-pointer"
          >
            <Plus className="w-4 h-4" />
            Create Block Volume
          </button>
        </div>
      </div>

      {}
      {errorMsg && (
        <div className="p-3.5 bg-rose-500/10 border border-rose-500/30 rounded-xl flex items-center justify-between text-xs text-rose-300">
          <div className="flex items-center gap-2">
            <AlertCircle className="w-4 h-4 text-rose-400 shrink-0" />
            <span>{errorMsg}</span>
          </div>
          <button onClick={() => setErrorMsg(null)} className="text-rose-400 hover:text-rose-200 font-bold cursor-pointer">×</button>
        </div>
      )}

      {successMsg && (
        <div className="p-3.5 bg-emerald-500/10 border border-emerald-500/30 rounded-xl flex items-center justify-between text-xs text-emerald-300">
          <div className="flex items-center gap-2">
            <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
            <span>{successMsg}</span>
          </div>
          <button onClick={() => setSuccessMsg(null)} className="text-emerald-400 hover:text-emerald-200 font-bold cursor-pointer">×</button>
        </div>
      )}

      {}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-[#141414] border border-[#262626] rounded-xl p-4 flex flex-col justify-between shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs text-[#a1a1a1] font-medium truncate">
              {activeServer ? `Disk ${activeServer.name}` : (selectedServerFilter === "local" ? "Host Lokal SSD" : "Physical Storage Pool")}
            </span>
            <div className="p-1.5 rounded-lg bg-emerald-500/10 text-emerald-400">
              <HardDrive className="w-4 h-4" />
            </div>
          </div>
          <p className="text-xl font-bold text-[#ededed] font-mono mt-2">
            {activePhysicalDiskGB.toFixed(1)} GB
          </p>
          <div className="mt-2 text-[11px] text-[#a1a1a1] flex items-center justify-between">
            <span>Free: <strong className="text-emerald-400">{activeFreeDiskGB.toFixed(1)} GB</strong></span>
            <span>Used: {activeUsedDiskGB.toFixed(1)} GB</span>
          </div>
          <div className="w-full bg-[#222222] rounded-full h-1.5 mt-2.5 overflow-hidden">
            <div
              className="bg-emerald-400 h-1.5 rounded-full transition-all"
              style={{ width: `${Math.min(100, Math.max(5, activeUsagePercent))}%` }}
            ></div>
          </div>
        </div>

        <div className="bg-[#141414] border border-[#262626] rounded-xl p-4 flex flex-col justify-between shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs text-[#a1a1a1] font-medium">Total Volumes</span>
            <div className="p-1.5 rounded-lg bg-purple-500/10 text-purple-400">
              <Database className="w-4 h-4" />
            </div>
          </div>
          <p className="text-xl font-bold text-[#ededed] font-mono mt-2">
            {serverScopedVolumes.length}
          </p>
          <p className="text-[11px] text-[#a1a1a1] mt-2">
            {serverScopedVolumes.filter((v) => v.status === "in-use").length} in-use / attached
          </p>
        </div>

        <div className="bg-[#141414] border border-[#262626] rounded-xl p-4 flex flex-col justify-between shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs text-[#a1a1a1] font-medium">Allocated Storage</span>
            <div className="p-1.5 rounded-lg bg-cyan-500/10 text-cyan-400">
              <Layers className="w-4 h-4" />
            </div>
          </div>
          <p className="text-xl font-bold text-[#ededed] font-mono mt-2">
            {currentAllocatedGB} GB
          </p>
          <p className="text-[11px] text-[#a1a1a1] mt-2">
            Across {serverScopedVolumes.length} persistent block drives
          </p>
        </div>

        <div className="bg-[#141414] border border-[#262626] rounded-xl p-4 flex flex-col justify-between shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs text-[#a1a1a1] font-medium">I/O Performance</span>
            <div className="p-1.5 rounded-lg bg-amber-500/10 text-amber-400">
              <Activity className="w-4 h-4" />
            </div>
          </div>
          <p className="text-xl font-bold text-[#ededed] font-mono mt-2">
            3,000 - 5,000
          </p>
          <p className="text-[11px] text-[#a1a1a1] mt-2">
            IOPS NVMe Direct Attached
          </p>
        </div>
      </div>

      {}
      <div className="bg-[#11141a] border border-[#22272e] rounded-xl overflow-hidden shadow-sm">
        <div className="p-4 border-b border-[#22272e] flex flex-col sm:flex-row items-center justify-between gap-3">
          <div className="relative w-full sm:w-72">
            <Search className="w-4 h-4 text-slate-400 absolute left-3 top-2.5" />
            <input
              type="text"
              placeholder="Search volume name, mount path..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full bg-[#161b22] border border-slate-800 rounded-lg pl-9 pr-3 py-1.5 text-xs text-slate-200 placeholder-slate-400 focus:outline-none focus:border-emerald-500 transition-colors"
            />
          </div>
          <div className="text-xs text-slate-400 font-mono">
            Showing {filteredVolumes.length} of {serverScopedVolumes.length} volumes
            {selectedServerFilter !== "all" && (
              <span className="text-emerald-400 ml-1">
                (Scoped: {activeServer ? activeServer.name : "Local Host"})
              </span>
            )}
          </div>
        </div>

        {}
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse text-xs">
            <thead>
              <tr className="bg-[#161b22] border-b border-[#22272e] text-slate-400 font-semibold">
                <th className="py-3 px-4">Volume Name</th>
                <th className="py-3 px-4">Server Node</th>
                <th className="py-3 px-4">Type & Filesystem</th>
                <th className="py-3 px-4">Allocated Size</th>
                <th className="py-3 px-4">Mount Target</th>
                <th className="py-3 px-4">Attached Target</th>
                <th className="py-3 px-4">Status</th>
                <th className="py-3 px-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#22272e] text-slate-300">
              {filteredVolumes.map((vol) => {
                const size = vol.size_gb || vol.sizeGB || 0;
                const fs = vol.fs_type || vol.fsType || "ext4";
                const mount = vol.mount_path || vol.mountPath || "/mnt/data";
                const attached = vol.attached_container_name || vol.attachedServerName || null;
                const serverNode = servers.find((s: any) => s.id === vol.server_id);

                return (
                  <tr key={vol.id} className="hover:bg-slate-850/50 transition-colors cursor-pointer" onClick={() => setSelectedVolumeDetail(vol)}>
                    <td className="py-3.5 px-4 font-mono font-medium text-slate-200">
                      <div className="flex items-center gap-2">
                        <HardDrive className="w-4 h-4 text-emerald-400 shrink-0" />
                        <div>
                          <p className="font-semibold text-slate-100 hover:text-emerald-400 transition-colors">{vol.name}</p>
                          <p className="text-[10px] text-slate-400">caelus-{vol.name}</p>
                        </div>
                      </div>
                    </td>
                    <td className="py-3.5 px-4">
                      {serverNode ? (
                        <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-blue-500/10 text-blue-300 border border-blue-500/30 text-[11px] font-mono">
                          <Server className="w-3 h-3 text-blue-400" />
                          {serverNode.name}
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-slate-800 text-slate-300 border border-slate-700 text-[11px] font-mono">
                          <Server className="w-3 h-3 text-slate-400" />
                          Local Host
                        </span>
                      )}
                    </td>
                    <td className="py-3.5 px-4">
                      <span className="px-2 py-0.5 rounded text-[10px] font-mono bg-slate-800 border border-slate-700 text-slate-300 uppercase mr-1.5">
                        {vol.type}
                      </span>
                      <span className="text-[11px] font-mono text-slate-400">{fs}</span>
                    </td>
                    <td className="py-3.5 px-4 font-mono font-semibold text-slate-100">
                      {size} GB
                      <span className="block text-[10px] font-normal text-slate-400 font-sans">
                        {vol.iops || 3000} IOPS
                      </span>
                    </td>
                    <td className="py-3.5 px-4 font-mono text-slate-300">
                      {mount}
                    </td>
                    <td className="py-3.5 px-4">
                      {attached ? (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-purple-500/10 text-purple-300 border border-purple-500/30 text-[11px] font-mono font-semibold">
                          <Layers className="w-3 h-3" />
                          {attached}
                        </span>
                      ) : (
                        <span className="text-slate-400 text-[11px] italic">Unattached (Available)</span>
                      )}
                    </td>
                    <td className="py-3.5 px-4">
                      {vol.status === "in-use" ? (
                        <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 text-[10px] font-semibold">
                          <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
                          IN-USE
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-slate-800 text-slate-300 border border-slate-700 text-[10px] font-semibold">
                          AVAILABLE
                        </span>
                      )}
                    </td>
                    <td className="py-3.5 px-4 text-right" onClick={(e) => e.stopPropagation()}>
                      <button
                        onClick={() => handleDeleteVolume(vol.id, vol.name)}
                        className="p-1.5 text-slate-400 hover:text-rose-400 hover:bg-rose-500/10 border border-slate-800 rounded transition-colors cursor-pointer"
                        title="Delete volume"
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </button>
                    </td>
                  </tr>
                );
              })}

              {filteredVolumes.length === 0 && !isLoading && (
                <tr>
                  <td colSpan={8} className="py-12 text-center text-slate-400">
                    <HardDrive className="w-8 h-8 mx-auto mb-2 text-slate-600 opacity-50" />
                    <p className="text-sm font-medium">Belum ada persistent block volume pada node ini.</p>
                    <p className="text-xs text-slate-400 mt-1">
                      Klik &quot;Create Block Volume&quot; untuk membuat partisi penyimpanan fisik baru.
                    </p>
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {}
      {selectedVolumeDetail && (
        <Dialog
          isOpen={!!selectedVolumeDetail}
          onClose={() => setSelectedVolumeDetail(null)}
          title={`Detail Volume: ${selectedVolumeDetail.name.length > 24 ? selectedVolumeDetail.name.slice(0, 12) + "..." + selectedVolumeDetail.name.slice(-8) : selectedVolumeDetail.name}`}
          description="Informasi spesifikasi partisi fisik, server kepemilikan, dan status keterikatan kontainer."
          maxWidth="lg"
        >
          <div className="space-y-4 text-xs py-1">
            <div className="p-4 bg-[#141414] border border-[#262626] rounded-xl space-y-2.5">
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between py-1.5 border-b border-[#262626] gap-1">
                <span className="text-[#a1a1a1] shrink-0 font-medium">Volume ID</span>
                <span className="text-[#ededed] font-mono text-xs select-all break-all">{selectedVolumeDetail.id}</span>
              </div>
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between py-1.5 border-b border-[#262626] gap-1">
                <span className="text-[#a1a1a1] shrink-0 font-medium">Server Node</span>
                <span className="text-emerald-400 font-medium">
                  {servers.find((s: any) => s.id === selectedVolumeDetail.server_id)?.name || "Local Host (Mesin Utama)"}
                </span>
              </div>
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between py-1.5 border-b border-[#262626] gap-1">
                <span className="text-[#a1a1a1] shrink-0 font-medium">Volume Name</span>
                <span className="text-cyan-400 font-mono text-xs select-all break-all">{selectedVolumeDetail.name}</span>
              </div>
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between py-1.5 border-b border-[#262626] gap-1">
                <span className="text-[#a1a1a1] shrink-0 font-medium">Kapasitas Alokasi</span>
                <span className="text-[#ededed] font-semibold font-mono text-xs">{selectedVolumeDetail.size_gb || selectedVolumeDetail.sizeGB || 1} GB</span>
              </div>
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between py-1.5 border-b border-[#262626] gap-1">
                <span className="text-[#a1a1a1] shrink-0 font-medium">Storage Type / FS</span>
                <span className="text-[#ededed] font-mono uppercase text-xs">
                  {selectedVolumeDetail.type || "docker-volume"} ({selectedVolumeDetail.fs_type || selectedVolumeDetail.fsType || "ext4"})
                </span>
              </div>
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between py-1.5 border-b border-[#262626] gap-1">
                <span className="text-[#a1a1a1] shrink-0 font-medium">Mount Target</span>
                <span className="text-[#ededed] font-mono text-xs break-all select-all">{selectedVolumeDetail.mount_path || selectedVolumeDetail.mountPath || "/mnt/data"}</span>
              </div>
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between py-1.5 border-b border-[#262626] gap-1">
                <span className="text-[#a1a1a1] shrink-0 font-medium">Performance Rating</span>
                <span className="text-amber-400 font-mono text-xs font-semibold">{selectedVolumeDetail.iops || 3000} IOPS</span>
              </div>
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between py-1.5 gap-1">
                <span className="text-[#a1a1a1] shrink-0 font-medium">Attached Container</span>
                <span className="font-mono text-xs">
                  {selectedVolumeDetail.attached_container_name || selectedVolumeDetail.attachedServerName ? (
                    <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-purple-500/10 text-purple-400 border border-purple-500/20">
                      {selectedVolumeDetail.attached_container_name || selectedVolumeDetail.attachedServerName}
                    </span>
                  ) : (
                    <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-zinc-900 text-zinc-400 border border-zinc-800">
                      Unattached (Available)
                    </span>
                  )}
                </span>
              </div>
            </div>

            <div className="flex justify-end pt-2">
              <button
                type="button"
                onClick={() => setSelectedVolumeDetail(null)}
                className="px-4 py-2 bg-[#1a1a1a] hover:bg-[#222222] border border-[#2e2e2e] text-[#ededed] rounded-lg text-xs font-medium transition-colors cursor-pointer"
              >
                Tutup
              </button>
            </div>
          </div>
        </Dialog>
      )}

      {}
      <Dialog
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        title="Create Persistent Block Volume"
        description="Alokasikan ruang penyimpanan fisik persisten dari Storage Pool host Anda."
        maxWidth="md"
      >
        <div className="space-y-4 text-xs">
          {(() => {
            const currentTarget = servers.find((s: any) => s.id === volServerId);
            
            let maxCap = 75;
            if (currentTarget) {
              const totalCap = Number(currentTarget.disk_gb || currentTarget.diskGb) || 25;
              const usedCap = volumes
                .filter((v: any) => {
                  const sId = v.server_id || v.serverId || v.attachedServerId;
                  return sId && currentTarget.id && String(sId).toLowerCase() === String(currentTarget.id).toLowerCase();
                })
                .reduce((sum: number, v: any) => sum + (Number(v.size_gb || v.sizeGB || v.sizeGb) || 0), 0);
              maxCap = Math.max(1, totalCap - usedCap);
            } else if (stats) {
              maxCap = Math.max(1, Math.floor(stats.free_gb));
            }

            return (
              <form onSubmit={handleCreateVolume} className="space-y-4 text-xs">
                <div>
                  <label className="block text-slate-300 font-medium mb-1">Target Server / Host Node</label>
                  <select
                    value={volServerId}
                    onChange={(e) => {
                      setVolServerId(e.target.value);
                      setVolSize(1);
                    }}
                    className="w-full bg-[#161b22] border border-slate-800 rounded-lg px-3 py-2 text-slate-100 focus:outline-none focus:border-emerald-500 font-mono"
                  >
                    <option value="">Local Host (Current Machine - {stats ? `${stats.free_gb.toFixed(1)} GB Free` : "75.0 GB Free"})</option>
                    {servers.map((s: any) => (
                      <option key={s.id} value={s.id}>
                        {s.name} ({s.status})
                      </option>
                    ))}
                  </select>
                </div>

                  <div>
                    <label className="block text-slate-300 font-medium mb-1">Volume Name</label>
                    <input
                      type="text"
                      required
                      placeholder="misal: pg-database-data"
                      value={volName}
                      onChange={(e) => setVolName(e.target.value)}
                      className="w-full bg-[#161b22] border border-slate-800 rounded-lg px-3 py-2 text-slate-100 placeholder-slate-400 focus:outline-none focus:border-emerald-500"
                    />
                  </div>

                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-slate-300 font-medium mb-1">Storage Type</label>
                      <select
                        value={volType}
                        onChange={(e: any) => setVolType(e.target.value)}
                        className="w-full bg-[#161b22] border border-slate-800 rounded-lg px-3 py-2 text-slate-100 focus:outline-none focus:border-emerald-500"
                      >
                        <option value="nvme">NVMe SSD (High IOPS)</option>
                        <option value="ssd">Standard SSD</option>
                        <option value="docker-volume">Docker Named Volume</option>
                      </select>
                    </div>

                    <div>
                      <label className="block text-slate-300 font-medium mb-1">Filesystem Format</label>
                      <select
                        value={volFs}
                        onChange={(e: any) => setVolFs(e.target.value)}
                        className="w-full bg-[#161b22] border border-slate-800 rounded-lg px-3 py-2 text-slate-100 focus:outline-none focus:border-emerald-500"
                      >
                        <option value="ext4">ext4 (Recommended)</option>
                        <option value="xfs">xfs (Enterprise)</option>
                        <option value="btrfs">btrfs</option>
                      </select>
                    </div>
                  </div>

                  <div>
                    <div className="flex items-center justify-between mb-1">
                      <label className="text-slate-300 font-medium">Capacity (GB)</label>
                      <span className="text-slate-400 text-[10px]">
                        Maksimum: {maxCap} GB
                      </span>
                    </div>
                    <input
                      type="number"
                      required
                      min={1}
                      max={maxCap}
                      value={volSize}
                      onChange={(e) => setVolSize(Number(e.target.value))}
                      className="w-full bg-[#161b22] border border-slate-800 rounded-lg px-3 py-2 text-slate-100 font-mono focus:outline-none focus:border-emerald-500"
                    />
                    {volSize > maxCap && (
                      <p className="text-[11px] text-rose-400 mt-1">
                        Kapasitas melebihi batas {currentTarget ? `server ${currentTarget.name}` : "host lokal"} ({maxCap} GB).
                      </p>
                    )}
                  </div>

            <div>
              <label className="block text-slate-300 font-medium mb-1">Default Mount Path</label>
              <input
                type="text"
                required
                placeholder="/mnt/data"
                value={volMount}
                onChange={(e) => setVolMount(e.target.value)}
                className="w-full bg-[#161b22] border border-slate-800 rounded-lg px-3 py-2 text-slate-100 font-mono placeholder-slate-400 focus:outline-none focus:border-emerald-500"
              />
            </div>

                  <div className="flex items-center justify-end gap-2.5 pt-3 border-t border-slate-800">
                    <button
                      type="button"
                      onClick={() => setIsCreateModalOpen(false)}
                      className="px-3.5 py-1.5 text-xs text-slate-400 hover:text-slate-200 rounded-lg transition-colors cursor-pointer"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      disabled={isSubmitting || volSize > maxCap}
                      className="px-4 py-1.5 text-xs font-semibold text-slate-950 bg-emerald-400 hover:bg-emerald-300 disabled:opacity-50 disabled:cursor-not-allowed rounded-lg transition-colors cursor-pointer"
                    >
                      {isSubmitting ? "Allocating Volume..." : "Create Volume"}
                    </button>
                  </div>
                </form>
            );
          })()}
        </div>
      </Dialog>
    </div>
  );
}
