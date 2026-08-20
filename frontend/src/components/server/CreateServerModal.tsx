"use client";

import React, { useState, useEffect } from "react";
import { Dialog } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Provider } from "@/types/server";
import { providerService } from "@/services/provider.service";
import { useServerStore } from "@/stores/useServerStore";
import { Server, Cloud } from "lucide-react";
import { AppText } from "@/core/theme";

interface CreateServerModalProps {
  readonly isOpen: boolean;
  readonly onClose: () => void;
}

export const CreateServerModal: React.FC<CreateServerModalProps> = ({ isOpen, onClose }) => {
  const { createServer } = useServerStore();
  const [providers, setProviders] = useState<Provider[]>([]);
  const [selectedProviderID, setSelectedProviderID] = useState<string>("");
  const [name, setName] = useState("");
  const [region, setRegion] = useState("ap-southeast-1");
  const [osType, setOsType] = useState("ubuntu-22.04");
  const [cpuCores, setCpuCores] = useState(2);
  const [memoryMB, setMemoryMB] = useState(4096);
  const [diskGB, setDiskGB] = useState(50);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen) {
      providerService.listProviders().then((data) => {
        setProviders(data);
        if (data.length > 0) {
          const mock = data.find((p) => p.slug === "mock") || data[0];
          setSelectedProviderID(mock.id);
        }
      }).catch(() => {});
    }
  }, [isOpen]);

  const handleSubmit = async (e: React.SyntheticEvent) => {
    e.preventDefault();
    setError(null);
    setIsLoading(true);

    try {
      await createServer({
        provider_id: selectedProviderID,
        name,
        region,
        os_type: osType,
        plan_id: `std-${cpuCores}vcpu-${(memoryMB / 1024).toFixed(0)}gb`,
        cpu_cores: Number(cpuCores),
        memory_mb: Number(memoryMB),
        disk_gb: Number(diskGB),
      });
      onClose();
      setName("");
    } catch (err: any) {
      setError(err.response?.data?.message || err.response?.data?.errors || "Gagal membuat server.");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog
      isOpen={isOpen}
      onClose={onClose}
      title="Deploy Server VPS Baru"
      description="Pilih provider cloud dan spesifikasi komputasi instance baru Anda"
      maxWidth="lg"
    >
      {error && (
        <div className="mb-4 rounded-lg border border-rose-800/60 bg-rose-950/40 p-3 text-xs text-rose-300">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4 pb-2">
        {/* Provider Selection */}
        <div className="space-y-1.5">
          <span className={AppText.label}>Penyedia Cloud (Provider)</span>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
            {providers.map((p) => (
              <button
                type="button"
                key={p.id}
                onClick={() => setSelectedProviderID(p.id)}
                className={`p-3 rounded-lg border text-left flex items-center justify-between transition-all cursor-pointer ${
                  selectedProviderID === p.id
                    ? "border-emerald-500 bg-emerald-950/30 text-[#ededed] dark:text-[#ededed] light:text-[#111827] shadow-sm"
                    : "border-[#262626] dark:border-[#262626] light:border-[#e5e7eb] bg-[#121212] dark:bg-[#121212] light:bg-[#f9fafb] text-[#a1a1a1] hover:border-[#383838] hover:text-[#ededed]"
                }`}
              >
                <div className="flex items-center gap-2.5">
                  <Cloud className="h-4 w-4 text-emerald-400" />
                  <div>
                    <p className="text-xs font-semibold text-[#ededed] dark:text-[#ededed] light:text-[#111827]">{p.name}</p>
                    <p className="text-[10px] text-[#707070] font-mono">slug: {p.slug}</p>
                  </div>
                </div>
                {p.slug === "mock" && (
                  <span className="text-[10px] bg-emerald-500/15 text-emerald-400 px-1.5 py-0.5 rounded border border-emerald-500/30">
                    Simulasi
                  </span>
                )}
              </button>
            ))}
          </div>
        </div>

        {/* Server Name & Region */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <Input
            id="create-server-name"
            label="Nama Instance Server"
            required
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="web-production-01"
          />

          <div className="space-y-1.5">
            <label htmlFor="create-server-region" className={AppText.label}>
              Lokasi Data Center (Region)
            </label>
            <select
              id="create-server-region"
              value={region}
              onChange={(e) => setRegion(e.target.value)}
              className="w-full rounded-lg border border-[#262626] dark:border-[#262626] light:border-[#d1d5db] bg-[#121212] dark:bg-[#121212] light:bg-[#ffffff] px-3.5 py-2 text-xs text-[#ededed] dark:text-[#ededed] light:text-[#111827] focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500 transition-colors"
            >
              <option value="ap-southeast-1">Singapore (ap-southeast-1)</option>
              <option value="us-east-1">N. Virginia (us-east-1)</option>
              <option value="eu-central-1">Frankfurt (eu-central-1)</option>
              <option value="ap-northeast-1">Tokyo (ap-northeast-1)</option>
            </select>
          </div>
        </div>

        {/* OS Selection */}
        <div className="space-y-1.5">
          <span className={AppText.label}>Sistem Operasi (OS Image)</span>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
            {[
              { id: "ubuntu-22.04", label: "Ubuntu 22.04 LTS" },
              { id: "ubuntu-24.04", label: "Ubuntu 24.04 LTS" },
              { id: "debian-12", label: "Debian 12 Bookworm" },
              { id: "almalinux-9", label: "AlmaLinux 9" },
            ].map((os) => (
              <button
                type="button"
                key={os.id}
                onClick={() => setOsType(os.id)}
                className={`py-2 px-3 rounded-lg border text-xs text-center font-medium transition-colors cursor-pointer ${
                  osType === os.id
                    ? "border-emerald-500 bg-emerald-500/15 text-emerald-400"
                    : "border-[#262626] dark:border-[#262626] light:border-[#e5e7eb] bg-[#121212] dark:bg-[#121212] light:bg-[#f9fafb] text-[#a1a1a1] hover:border-[#383838]"
                }`}
              >
                {os.label}
              </button>
            ))}
          </div>
        </div>

        {/* Hardware Specifications */}
        <div className="space-y-2 pt-2 border-t border-[#262626] dark:border-[#262626] light:border-[#e5e7eb]">
          <span className={AppText.label}>Spesifikasi Komputasi Hardware</span>
          <div className="grid grid-cols-3 gap-3">
            <div>
              <label htmlFor="create-server-cpu" className="block text-[11px] text-[#a1a1a1] mb-1">
                vCPU Cores
              </label>
              <select
                id="create-server-cpu"
                value={cpuCores}
                onChange={(e) => setCpuCores(Number(e.target.value))}
                className="w-full rounded-lg border border-[#262626] dark:border-[#262626] light:border-[#d1d5db] bg-[#121212] dark:bg-[#121212] light:bg-[#ffffff] px-3 py-1.5 text-xs text-[#ededed] dark:text-[#ededed] light:text-[#111827] focus:border-emerald-500 focus:outline-none"
              >
                <option value={1}>1 Core</option>
                <option value={2}>2 Cores</option>
                <option value={4}>4 Cores</option>
                <option value={8}>8 Cores</option>
              </select>
            </div>

            <div>
              <label htmlFor="create-server-memory" className="block text-[11px] text-[#a1a1a1] mb-1">
                Memory RAM
              </label>
              <select
                id="create-server-memory"
                value={memoryMB}
                onChange={(e) => setMemoryMB(Number(e.target.value))}
                className="w-full rounded-lg border border-[#262626] dark:border-[#262626] light:border-[#d1d5db] bg-[#121212] dark:bg-[#121212] light:bg-[#ffffff] px-3 py-1.5 text-xs text-[#ededed] dark:text-[#ededed] light:text-[#111827] focus:border-emerald-500 focus:outline-none"
              >
                <option value={1024}>1 GB RAM</option>
                <option value={2048}>2 GB RAM</option>
                <option value={4096}>4 GB RAM</option>
                <option value={8192}>8 GB RAM</option>
                <option value={16384}>16 GB RAM</option>
              </select>
            </div>

            <div>
              <label htmlFor="create-server-disk" className="block text-[11px] text-[#a1a1a1] mb-1">
                SSD Storage
              </label>
              <select
                id="create-server-disk"
                value={diskGB}
                onChange={(e) => setDiskGB(Number(e.target.value))}
                className="w-full rounded-lg border border-[#262626] dark:border-[#262626] light:border-[#d1d5db] bg-[#121212] dark:bg-[#121212] light:bg-[#ffffff] px-3 py-1.5 text-xs text-[#ededed] dark:text-[#ededed] light:text-[#111827] focus:border-emerald-500 focus:outline-none"
              >
                <option value={25}>25 GB SSD</option>
                <option value={50}>50 GB SSD</option>
                <option value={100}>100 GB SSD</option>
                <option value={200}>200 GB SSD</option>
              </select>
            </div>
          </div>
        </div>

        {/* Modal Actions */}
        <div className="flex items-center justify-end gap-3 pt-4 border-t border-[#262626] dark:border-[#262626] light:border-[#e5e7eb]">
          <Button variant="outline" type="button" onClick={onClose} disabled={isLoading}>
            Batal
          </Button>
          <Button type="submit" isLoading={isLoading}>
            <Server className="h-4 w-4" />
            <span>Deploy Instance</span>
          </Button>
        </div>
      </form>
    </Dialog>
  );
};
