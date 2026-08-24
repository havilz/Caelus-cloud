"use client";

import React, { useState } from "react";
import {
  HardDrive,
  Plus,
  Server,
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
} from "lucide-react";
import { Dialog } from "@/components/ui/dialog";
import { AppTheme } from "@/core/theme";

interface StorageVolume {
  id: string;
  name: string;
  type: "nvme" | "ssd" | "docker-volume";
  sizeGB: number;
  fsType: "ext4" | "xfs" | "btrfs";
  mountPath: string;
  attachedServerName: string | null;
  attachedServerId: string | null;
  status: "in-use" | "available" | "mounting";
  iops: number;
  createdAt: string;
}

export default function VolumesManagementPage() {
  const [volumes, setVolumes] = useState<StorageVolume[]>([]);
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [isCreateModalOpen, setIsCreateModalOpen] = useState<boolean>(false);
  const [isResizeModalOpen, setIsResizeModalOpen] = useState<boolean>(false);
  const [selectedVolume, setSelectedVolume] = useState<StorageVolume | null>(null);

  // New Volume Form State
  const [volName, setVolName] = useState<string>("");
  const [volType, setVolType] = useState<"nvme" | "ssd" | "docker-volume">("nvme");
  const [volSize, setVolSize] = useState<number>(50);
  const [volFs, setVolFs] = useState<"ext4" | "xfs" | "btrfs">("ext4");
  const [volMount, setVolMount] = useState<string>("/mnt/data");
  const [newSize, setNewSize] = useState<number>(100);

  const handleCreateVolume = (e: React.FormEvent) => {
    e.preventDefault();
    if (!volName.trim()) return;

    const newVol: StorageVolume = {
      id: `vol-${Date.now()}`,
      name: volName.trim().toLowerCase().replace(/\s+/g, "-"),
      type: volType,
      sizeGB: Number(volSize),
      fsType: volFs,
      mountPath: volMount.trim() || "/mnt/data",
      attachedServerName: null,
      attachedServerId: null,
      status: "available",
      iops: volType === "nvme" ? 3000 : 1200,
      createdAt: new Date().toISOString().split("T")[0],
    };

    setVolumes([newVol, ...volumes]);
    setVolName("");
    setIsCreateModalOpen(false);
  };

  const handleResizeVolume = (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedVolume) return;

    setVolumes(
      volumes.map((v) =>
        v.id === selectedVolume.id ? { ...v, sizeGB: Number(newSize) } : v
      )
    );
    setIsResizeModalOpen(false);
    setSelectedVolume(null);
  };

  const handleDeleteVolume = (id: string) => {
    setVolumes(volumes.filter((v) => v.id !== id));
  };

  const filteredVolumes = volumes.filter((v) =>
    v.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    v.mountPath.toLowerCase().includes(searchQuery.toLowerCase()) ||
    (v.attachedServerName && v.attachedServerName.toLowerCase().includes(searchQuery.toLowerCase()))
  );

  const totalCapacityGB = volumes.reduce((acc, v) => acc + v.sizeGB, 0);
  const attachedCapacityGB = volumes.filter((v) => v.status === "in-use").reduce((acc, v) => acc + v.sizeGB, 0);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-100 flex items-center gap-2.5">
            <HardDrive className="w-6 h-6 text-emerald-400" />
            Persistent Storage & Block Volumes
          </h1>
          <p className="text-sm text-slate-400 mt-1">
            Kelola Block Storage Volumes, Docker Persistent Volumes, dan Mount Point server terpusat.
          </p>
        </div>
        <div>
          <button
            onClick={() => setIsCreateModalOpen(true)}
            className="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-slate-950 bg-emerald-400 hover:bg-emerald-300 rounded-lg transition-colors shadow-sm"
          >
            <Plus className="w-4 h-4" />
            Create Block Volume
          </button>
        </div>
      </div>

      {/* Top Stats Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="p-4 rounded-xl bg-slate-900/60 border border-slate-800/80">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-slate-400">Total Allocated Storage</span>
            <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-400">
              <Database className="w-4 h-4" />
            </div>
          </div>
          <p className="text-2xl font-bold text-slate-100 mt-2">{totalCapacityGB} GB</p>
          <span className="text-xs text-slate-400 mt-1 block">Across {volumes.length} volumes</span>
        </div>

        <div className="p-4 rounded-xl bg-slate-900/60 border border-slate-800/80">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-slate-400">Attached Storage</span>
            <div className="p-2 rounded-lg bg-cyan-500/10 text-cyan-400">
              <Server className="w-4 h-4" />
            </div>
          </div>
          <p className="text-2xl font-bold text-slate-100 mt-2">{attachedCapacityGB} GB</p>
          <span className="text-xs text-emerald-400 flex items-center gap-1 mt-1">
            <CheckCircle2 className="w-3 h-3" /> Mounted to active compute
          </span>
        </div>

        <div className="p-4 rounded-xl bg-slate-900/60 border border-slate-800/80">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-slate-400">Available Unattached</span>
            <div className="p-2 rounded-lg bg-amber-500/10 text-amber-400">
              <HardDrive className="w-4 h-4" />
            </div>
          </div>
          <p className="text-2xl font-bold text-slate-100 mt-2">{totalCapacityGB - attachedCapacityGB} GB</p>
          <span className="text-xs text-amber-400 mt-1 block">Ready to mount</span>
        </div>

        <div className="p-4 rounded-xl bg-slate-900/60 border border-slate-800/80">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-slate-400">Peak Storage IOPS</span>
            <div className="p-2 rounded-lg bg-purple-500/10 text-purple-400">
              <Layers className="w-4 h-4" />
            </div>
          </div>
          <p className="text-2xl font-bold text-slate-100 mt-2">3,000 IOPS</p>
          <span className="text-xs text-purple-400 mt-1 block">Ultra NVMe SSD Tier</span>
        </div>
      </div>

      {/* Search & Filter Bar */}
      <div className="flex items-center justify-between gap-4 border-b border-slate-800 pb-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
          <input
            type="text"
            placeholder="Search volumes, mount points, or servers..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-9 pr-4 py-2 text-xs rounded-lg bg-slate-900/80 border border-slate-800 text-slate-200 placeholder-slate-500 focus:outline-none focus:border-emerald-500/50"
          />
        </div>
      </div>

      {/* Volumes Table */}
      <div className="rounded-xl bg-slate-900/60 border border-slate-800/80 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-slate-950/60 text-slate-400 border-b border-slate-800 font-medium">
              <tr>
                <th className="px-4 py-3">Volume Name</th>
                <th className="px-4 py-3">Type</th>
                <th className="px-4 py-3">Capacity</th>
                <th className="px-4 py-3">Filesystem</th>
                <th className="px-4 py-3">Mount Point</th>
                <th className="px-4 py-3">Attached Server</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60 text-slate-300">
              {filteredVolumes.length === 0 ? (
                <tr>
                  <td colSpan={8} className="px-6 py-12 text-center">
                    <div className="flex flex-col items-center justify-center space-y-3">
                      <div className="p-3 rounded-full bg-slate-800/60 border border-slate-700 text-slate-400">
                        <HardDrive className="w-6 h-6" />
                      </div>
                      <div className="space-y-1">
                        <p className="text-sm font-semibold text-slate-200">No Storage Volumes Found</p>
                        <p className="text-xs text-slate-400 max-w-sm">
                          {searchQuery
                            ? `No volumes matching "${searchQuery}".`
                            : "Create high-performance NVMe/SSD block storage volumes to attach to your compute nodes."}
                        </p>
                      </div>
                      {!searchQuery && (
                        <button
                          onClick={() => setIsCreateModalOpen(true)}
                          className="mt-2 px-3.5 py-2 rounded-lg text-xs font-semibold bg-emerald-500 hover:bg-emerald-400 text-slate-950 transition-colors flex items-center gap-1.5"
                        >
                          <Plus className="w-3.5 h-3.5" />
                          <span>Create Volume</span>
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ) : (
                filteredVolumes.map((vol) => (
                  <tr key={vol.id} className="hover:bg-slate-800/30 transition-colors">
                    <td className="px-4 py-3.5 font-medium text-slate-200">
                      <div className="flex items-center gap-2.5">
                        <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                          <HardDrive className="w-4 h-4" />
                        </div>
                        <div>
                          <div className="font-semibold text-slate-100">{vol.name}</div>
                          <span className="text-[10px] font-mono text-slate-500">{vol.id}</span>
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3.5 font-mono uppercase text-slate-300">
                      <span className="px-2 py-0.5 rounded text-[10px] font-medium bg-slate-800 text-slate-300">
                        {vol.type}
                      </span>
                    </td>
                    <td className="px-4 py-3.5 font-semibold text-emerald-400">{vol.sizeGB} GB</td>
                    <td className="px-4 py-3.5 font-mono text-slate-400">{vol.fsType}</td>
                    <td className="px-4 py-3.5 font-mono text-slate-300">{vol.mountPath}</td>
                    <td className="px-4 py-3.5">
                      {vol.attachedServerName ? (
                        <span className="inline-flex items-center gap-1.5 text-cyan-400 font-medium">
                          <Server className="w-3.5 h-3.5" />
                          {vol.attachedServerName}
                        </span>
                      ) : (
                        <span className="text-slate-500 italic">None (Unattached)</span>
                      )}
                    </td>
                    <td className="px-4 py-3.5">
                      <span
                        className={`px-2 py-0.5 rounded text-[10px] font-medium uppercase ${
                          vol.status === "in-use"
                            ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                            : "bg-amber-500/10 text-amber-400 border border-amber-500/20"
                        }`}
                      >
                        {vol.status}
                      </span>
                    </td>
                    <td className="px-4 py-3.5 text-right">
                      <div className="flex items-center justify-end gap-1.5">
                        <button
                          onClick={() => {
                            setSelectedVolume(vol);
                            setNewSize(vol.sizeGB + 50);
                            setIsResizeModalOpen(true);
                          }}
                          className="p-1.5 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-slate-800 transition-colors"
                          title="Resize Volume"
                        >
                          <Sliders className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => handleDeleteVolume(vol.id)}
                          className="p-1.5 rounded-lg text-slate-500 hover:text-rose-400 hover:bg-rose-500/10 transition-colors"
                          title="Delete Volume"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Modal: Create Volume */}
      <Dialog
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        title="Create Block Storage Volume"
        description="Buat volume penyimpanan persisten berkinerja tinggi untuk server atau kontainer."
      >
        <form onSubmit={handleCreateVolume} className="space-y-4 mt-2">
          <div>
            <label className="block text-xs font-medium text-slate-300 mb-1">Volume Name</label>
            <input
              type="text"
              required
              placeholder="misal: postgres-data-vol"
              value={volName}
              onChange={(e) => setVolName(e.target.value)}
              className="w-full px-3 py-2 text-xs rounded-lg bg-slate-950 border border-slate-800 text-slate-200 focus:outline-none focus:border-emerald-500/50"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-slate-300 mb-1">Storage Type</label>
              <select
                value={volType}
                onChange={(e) => setVolType(e.target.value as any)}
                className="w-full px-3 py-2 text-xs rounded-lg bg-slate-950 border border-slate-800 text-slate-200"
              >
                <option value="nvme">NVMe SSD (High IOPS)</option>
                <option value="ssd">Standard SSD</option>
                <option value="docker-volume">Docker Named Volume</option>
              </select>
            </div>
            <div>
              <label className="block text-xs font-medium text-slate-300 mb-1">Capacity (GB)</label>
              <input
                type="number"
                min="10"
                max="2000"
                value={volSize}
                onChange={(e) => setVolSize(Number(e.target.value))}
                className="w-full px-3 py-2 text-xs rounded-lg bg-slate-950 border border-slate-800 text-slate-200 font-mono"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-slate-300 mb-1">Filesystem</label>
              <select
                value={volFs}
                onChange={(e) => setVolFs(e.target.value as any)}
                className="w-full px-3 py-2 text-xs rounded-lg bg-slate-950 border border-slate-800 text-slate-200"
              >
                <option value="ext4">ext4 (Linux Default)</option>
                <option value="xfs">XFS (High Throughput)</option>
                <option value="btrfs">Btrfs (Snapshots)</option>
              </select>
            </div>
            <div>
              <label className="block text-xs font-medium text-slate-300 mb-1">Default Mount Path</label>
              <input
                type="text"
                placeholder="/mnt/data"
                value={volMount}
                onChange={(e) => setVolMount(e.target.value)}
                className="w-full px-3 py-2 text-xs rounded-lg bg-slate-950 border border-slate-800 text-slate-200 font-mono"
              />
            </div>
          </div>

          <div className="flex justify-end gap-2 pt-3 border-t border-slate-800">
            <button
              type="button"
              onClick={() => setIsCreateModalOpen(false)}
              className="px-4 py-2 text-xs text-slate-400 hover:text-slate-200 bg-slate-800 rounded-lg"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 text-xs font-medium text-slate-950 bg-emerald-400 hover:bg-emerald-300 rounded-lg transition-colors"
            >
              Create Volume
            </button>
          </div>
        </form>
      </Dialog>

      {/* Modal: Resize Volume */}
      <Dialog
        isOpen={isResizeModalOpen}
        onClose={() => setIsResizeModalOpen(false)}
        title="Resize Storage Volume"
        description={`Perluas kapasitas volume ${selectedVolume?.name || ''} secara online tanpa downtime.`}
      >
        <form onSubmit={handleResizeVolume} className="space-y-4 mt-2">
          <div>
            <label className="block text-xs font-medium text-slate-300 mb-1">Current Size</label>
            <div className="px-3 py-2 text-xs rounded-lg bg-slate-950 border border-slate-800 text-slate-400 font-mono">
              {selectedVolume?.sizeGB} GB
            </div>
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-300 mb-1">New Capacity (GB)</label>
            <input
              type="number"
              min={selectedVolume ? selectedVolume.sizeGB + 1 : 10}
              max="5000"
              value={newSize}
              onChange={(e) => setNewSize(Number(e.target.value))}
              className="w-full px-3 py-2 text-xs rounded-lg bg-slate-950 border border-slate-800 text-slate-200 font-mono focus:outline-none focus:border-emerald-500/50"
            />
          </div>

          <div className="flex justify-end gap-2 pt-3 border-t border-slate-800">
            <button
              type="button"
              onClick={() => setIsResizeModalOpen(false)}
              className="px-4 py-2 text-xs text-slate-400 hover:text-slate-200 bg-slate-800 rounded-lg"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 text-xs font-medium text-slate-950 bg-emerald-400 hover:bg-emerald-300 rounded-lg transition-colors"
            >
              Apply Resize
            </button>
          </div>
        </form>
      </Dialog>
    </div>
  );
}
