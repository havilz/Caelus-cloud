"use client";

import React, { useState, useEffect } from "react";
import { Dialog } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Server } from "@/types/server";
import { useServerStore } from "@/stores/useServerStore";
import { Sliders } from "lucide-react";
import { AppText } from "@/core/theme";

interface ResizeServerModalProps {
  readonly server: Server | null;
  readonly isOpen: boolean;
  readonly onClose: () => void;
}

export const ResizeServerModal: React.FC<ResizeServerModalProps> = ({
  server,
  isOpen,
  onClose,
}) => {
  const { resizeServer } = useServerStore();
  const [cpuCores, setCpuCores] = useState(2);
  const [memoryMB, setMemoryMB] = useState(4096);
  const [diskGB, setDiskGB] = useState(50);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (server) {
      setCpuCores(server.cpu_cores);
      setMemoryMB(server.memory_mb);
      setDiskGB(server.disk_gb);
    }
  }, [server]);

  if (!server) return null;

  const handleSubmit = async (e: React.SyntheticEvent) => {
    e.preventDefault();
    setError(null);
    setIsLoading(true);

    try {
      await resizeServer(server.id, {
        cpu_cores: Number(cpuCores),
        memory_mb: Number(memoryMB),
        disk_gb: Number(diskGB),
        plan_id: `std-${cpuCores}vcpu-${(memoryMB / 1024).toFixed(0)}gb`,
      });
      onClose();
    } catch (err: any) {
      setError(err.response?.data?.message || "Gagal mengubah spesifikasi server.");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog
      isOpen={isOpen}
      onClose={onClose}
      title={`Resize Spesifikasi: ${server.name}`}
      description="Ubah kapasitas vCPU, RAM, dan Disk instance server ini"
      maxWidth="md"
    >
      {error && (
        <div className="mb-4 rounded-lg border border-rose-800/60 bg-rose-950/40 p-3 text-xs text-rose-300">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4 pb-2">
        <div className="space-y-3">
          <div>
            <label htmlFor="resize-cpu" className={AppText.label}>
              vCPU Cores
            </label>
            <select
              id="resize-cpu"
              value={cpuCores}
              onChange={(e) => setCpuCores(Number(e.target.value))}
              className="w-full mt-1 rounded-lg border border-[#262626] dark:border-[#262626] light:border-[#d1d5db] bg-[#121212] dark:bg-[#121212] light:bg-[#ffffff] px-3 py-2 text-xs text-[#ededed] dark:text-[#ededed] light:text-[#111827] focus:border-emerald-500 focus:outline-none"
            >
              <option value={1}>1 Core</option>
              <option value={2}>2 Cores</option>
              <option value={4}>4 Cores</option>
              <option value={8}>8 Cores</option>
              <option value={16}>16 Cores</option>
            </select>
          </div>

          <div>
            <label htmlFor="resize-memory" className={AppText.label}>
              Memory RAM
            </label>
            <select
              id="resize-memory"
              value={memoryMB}
              onChange={(e) => setMemoryMB(Number(e.target.value))}
              className="w-full mt-1 rounded-lg border border-[#262626] dark:border-[#262626] light:border-[#d1d5db] bg-[#121212] dark:bg-[#121212] light:bg-[#ffffff] px-3 py-2 text-xs text-[#ededed] dark:text-[#ededed] light:text-[#111827] focus:border-emerald-500 focus:outline-none"
            >
              <option value={1024}>1 GB RAM (1024 MB)</option>
              <option value={2048}>2 GB RAM (2048 MB)</option>
              <option value={4096}>4 GB RAM (4096 MB)</option>
              <option value={8192}>8 GB RAM (8192 MB)</option>
              <option value={16384}>16 GB RAM (16384 MB)</option>
              <option value={32768}>32 GB RAM (32768 MB)</option>
            </select>
          </div>

          <div>
            <label htmlFor="resize-disk" className={AppText.label}>
              SSD Storage Disk
            </label>
            <select
              id="resize-disk"
              value={diskGB}
              onChange={(e) => setDiskGB(Number(e.target.value))}
              className="w-full mt-1 rounded-lg border border-[#262626] dark:border-[#262626] light:border-[#d1d5db] bg-[#121212] dark:bg-[#121212] light:bg-[#ffffff] px-3 py-2 text-xs text-[#ededed] dark:text-[#ededed] light:text-[#111827] focus:border-emerald-500 focus:outline-none"
            >
              <option value={25}>25 GB SSD</option>
              <option value={50}>50 GB SSD</option>
              <option value={100}>100 GB SSD</option>
              <option value={200}>200 GB SSD</option>
              <option value={500}>500 GB SSD</option>
            </select>
          </div>
        </div>

        <div className="flex items-center justify-end gap-3 pt-4 border-t border-[#262626] dark:border-[#262626] light:border-[#e5e7eb]">
          <Button variant="outline" type="button" onClick={onClose} disabled={isLoading}>
            Batal
          </Button>
          <Button type="submit" isLoading={isLoading}>
            <Sliders className="h-4 w-4" />
            <span>Terapkan Resize</span>
          </Button>
        </div>
      </form>
    </Dialog>
  );
};
