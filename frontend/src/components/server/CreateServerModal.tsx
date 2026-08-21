"use client";

import React, { useState, useEffect } from "react";
import { Dialog } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Provider, Server } from "@/types/server";
import { providerService } from "@/services/provider.service";
import { useServerStore } from "@/stores/useServerStore";
import {
  Server as ServerIcon,
  Cloud,
  Terminal,
  Copy,
  Check,
  ShieldCheck,
  CheckCircle2,
  Cpu,
  HardDrive,
  Activity,
  Sparkles,
} from "lucide-react";
import { AppText } from "@/core/theme";

interface CreateServerModalProps {
  readonly isOpen: boolean;
  readonly onClose: () => void;
}

type OnboardingTab = "byos" | "cloud";

export const CreateServerModal: React.FC<CreateServerModalProps> = ({ isOpen, onClose }) => {
  const { createServer, fetchServers } = useServerStore();
  const [tab, setTab] = useState<OnboardingTab>("byos");
  const [providers, setProviders] = useState<Provider[]>([]);
  const [selectedProviderID, setSelectedProviderID] = useState<string>("");
  const [name, setName] = useState("");
  const [region, setRegion] = useState("id-cgk");
  const [osType, setOsType] = useState("ubuntu-24.04");
  const [cpuCores, setCpuCores] = useState(2);
  const [memoryMB, setMemoryMB] = useState(4096);
  const [diskGB, setDiskGB] = useState(50);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Success Connection Step State
  const [createdServer, setCreatedServer] = useState<Server | null>(null);
  const [copiedType, setCopiedType] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen) {
      setCreatedServer(null);
      setError(null);
      setName("");
      providerService.listProviders().then((data) => {
        setProviders(data);
        if (data.length > 0) {
          const customProv = data.find((p) => p.slug === "custom") || data.find((p) => p.slug === "mock") || data[0];
          setSelectedProviderID(customProv.id);
        }
      }).catch(() => {});
    }
  }, [isOpen]);

  const handleTabChange = (selectedTab: OnboardingTab) => {
    setTab(selectedTab);
    if (selectedTab === "byos") {
      const customProv = providers.find((p) => p.slug === "custom") || providers.find((p) => p.slug === "mock");
      if (customProv) setSelectedProviderID(customProv.id);
      setRegion("id-cgk");
    } else {
      const mockProv = providers.find((p) => p.slug === "mock") || providers[0];
      if (mockProv) setSelectedProviderID(mockProv.id);
      setRegion("ap-southeast-1");
    }
  };

  const handleSubmit = async (e: React.SyntheticEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      setError("Nama server wajib diisi.");
      return;
    }

    setError(null);
    setIsLoading(true);

    try {
      let targetProviderID = selectedProviderID;
      if (tab === "byos") {
        const customProv = providers.find((p) => p.slug === "custom") || providers.find((p) => p.slug === "mock");
        if (customProv) targetProviderID = customProv.id;
      }

      const server = await createServer({
        provider_id: targetProviderID,
        name: name.trim(),
        region: tab === "byos" ? "custom" : region,
        os_type: tab === "byos" ? "auto-detect" : osType,
        plan_id: tab === "byos" ? "byos-custom" : `std-${cpuCores}vcpu-${(memoryMB / 1024).toFixed(0)}gb`,
        cpu_cores: tab === "byos" ? 1 : Number(cpuCores),
        memory_mb: tab === "byos" ? 1024 : Number(memoryMB),
        disk_gb: tab === "byos" ? 25 : Number(diskGB),
      });

      if (tab === "byos" && server) {
        setCreatedServer(server);
      } else {
        onClose();
        setName("");
      }
      fetchServers(1);
    } catch (err: any) {
      setError(err.response?.data?.message || err.response?.data?.errors || "Gagal mendaftarkan server.");
    } finally {
      setIsLoading(false);
    }
  };

  const apiEndpoint = typeof window !== "undefined" ? `${window.location.protocol}//${window.location.hostname}:8080` : "http://localhost:8080";
  const agentSecret = createdServer ? "caelus_agent_sec_" + createdServer.id.replace(/-/g, "").substring(0, 16) : "";
  const oneLineCommand = createdServer
    ? `curl -sSL ${apiEndpoint}/install.sh | sudo bash -s -- --server-id="${createdServer.id}" --secret="${agentSecret}" --api="${apiEndpoint}"`
    : "";

  const handleCopy = (text: string, type: string) => {
    navigator.clipboard.writeText(text);
    setCopiedType(type);
    setTimeout(() => setCopiedType(null), 2000);
  };

  return (
    <Dialog
      isOpen={isOpen}
      onClose={onClose}
      title={createdServer ? "Langkah Terakhir: Hubungkan Caelus Agent" : "Tambah atau Hubungkan Server"}
      description={
        createdServer
          ? `Server "${createdServer.name}" berhasil didaftarkan. Jalankan perintah instalasi di VPS Anda.`
          : "Hubungkan VPS existing (IDCloudHost, Home Server) atau deploy instance baru via Cloud IaaS"
      }
      maxWidth="lg"
    >
      {error && (
        <div className="mb-4 rounded-lg border border-rose-800/60 bg-rose-950/40 p-3 text-xs text-rose-300">
          {error}
        </div>
      )}

      {/* Step 2: Post-Creation Installation Command Screen */}
      {createdServer ? (
        <div className="space-y-4 py-1">
          <div className="p-3.5 rounded-xl bg-[#141414] border border-emerald-500/30 bg-emerald-950/20 flex items-center justify-between">
            <div>
              <span className="text-[10px] text-emerald-400 font-semibold uppercase tracking-wider">Server Terdaftar</span>
              <h4 className="text-sm font-bold text-[#ededed]">{createdServer.name}</h4>
              <p className="text-xs font-mono text-[#a1a1a1]">Status: Menunggu sinyal koneksi agent...</p>
            </div>
            <div className="p-2 rounded-lg bg-emerald-500/20 text-emerald-400">
              <CheckCircle2 className="w-5 h-5" />
            </div>
          </div>

          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Terminal className="w-4 h-4 text-emerald-400" />
                <span className="text-xs font-semibold text-[#ededed]">Perintah Instalasi 1-Line (Jalankan di Terminal VPS)</span>
              </div>
              <span className="text-[10px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 px-2 py-0.5 rounded">
                Linux / Ubuntu / Debian
              </span>
            </div>

            <div className="relative group">
              <pre className="p-3.5 rounded-xl bg-[#0d0d0d] border border-[#222222] text-xs font-mono text-emerald-400 overflow-x-auto whitespace-pre-wrap break-all select-all">
                {oneLineCommand}
              </pre>
              <button
                onClick={() => handleCopy(oneLineCommand, "oneline")}
                className="absolute top-2.5 right-2.5 px-2.5 py-1 rounded-lg bg-[#1f1f1f] hover:bg-[#2a2a2a] text-[#ededed] border border-[#333333] text-xs flex items-center gap-1.5 transition-colors cursor-pointer shadow-md"
              >
                {copiedType === "oneline" ? (
                  <>
                    <Check className="w-3.5 h-3.5 text-emerald-400" />
                    <span className="text-emerald-400 text-[11px]">Tersalin!</span>
                  </>
                ) : (
                  <>
                    <Copy className="w-3.5 h-3.5" />
                    <span className="text-[11px]">Salin Perintah</span>
                  </>
                )}
              </button>
            </div>
            <p className="text-[11px] text-[#707070]">
              Begitu script ini dijalankan, Caelus Agent akan otomatis mendeteksi CPU, RAM, Disk, OS, dan IP VPS Anda lalu memperbarui panel secara realtime.
            </p>
          </div>

          <div className="grid grid-cols-2 gap-2.5 p-3 rounded-xl bg-[#121212] border border-[#262626] text-[11px]">
            <div>
              <span className="text-[#707070] uppercase tracking-wider block mb-0.5">Server UUID</span>
              <span className="font-mono text-emerald-400 select-all">{createdServer.id}</span>
            </div>
            <div>
              <span className="text-[#707070] uppercase tracking-wider block mb-0.5">Agent Secret</span>
              <span className="font-mono text-[#a1a1a1] select-all">{agentSecret}</span>
            </div>
          </div>

          <div className="flex items-center justify-end gap-3 pt-3 border-t border-[#262626]">
            <Button
              type="button"
              onClick={() => {
                onClose();
                setName("");
                setCreatedServer(null);
              }}
            >
              Selesai & Buka Daftar Server
            </Button>
          </div>
        </div>
      ) : (
        /* Step 1: Creation Form */
        <div>
          {/* Tab Navigation */}
          <div className="flex items-center gap-2 p-1 mb-4 rounded-lg bg-[#141414] border border-[#262626]">
            <button
              type="button"
              onClick={() => handleTabChange("byos")}
              className={`flex-1 py-2 px-3 rounded-md text-xs font-semibold flex items-center justify-center gap-2 transition-all cursor-pointer ${
                tab === "byos"
                  ? "bg-emerald-500 text-zinc-950 shadow-sm"
                  : "text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#1f1f1f]"
              }`}
            >
              <Terminal className="w-4 h-4" />
              <span>Hubungkan VPS / Home Server (BYOS)</span>
            </button>

            <button
              type="button"
              onClick={() => handleTabChange("cloud")}
              className={`flex-1 py-2 px-3 rounded-md text-xs font-semibold flex items-center justify-center gap-2 transition-all cursor-pointer ${
                tab === "cloud"
                  ? "bg-emerald-500 text-zinc-950 shadow-sm"
                  : "text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#1f1f1f]"
              }`}
            >
              <Cloud className="w-4 h-4" />
              <span>Deploy Cloud Baru (IaaS API)</span>
            </button>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4 pb-2">
            {/* BYOS Mode: Super Simple Name-Only Flow */}
            {tab === "byos" ? (
              <div className="space-y-4">
                <div className="p-3.5 rounded-xl bg-emerald-950/20 border border-emerald-500/20 text-xs text-[#ededed] flex items-start gap-3">
                  <ShieldCheck className="w-5 h-5 text-emerald-400 shrink-0 mt-0.5" />
                  <div>
                    <h5 className="font-semibold text-emerald-400 mb-0.5">Integrasi Instan Bring-Your-Own-Server</h5>
                    <p className="text-[#a1a1a1] leading-relaxed">
                      Cukup beri nama server Anda. Anda <strong>tidak perlu memasukkan spesifikasi atau OS</strong> secara manual karena Caelus Agent akan menginspeksi dan menyinkronkan seluruh hardware VPS Anda secara otomatis.
                    </p>
                  </div>
                </div>

                <Input
                  id="create-server-name"
                  label="Nama Server / Label"
                  required
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g. idcloud-ubuntu-prod atau homelab-proxmox-01"
                />

                {/* Auto-Detection Features Preview */}
                <div className="p-4 rounded-xl bg-[#141414] border border-[#262626] space-y-3">
                  <div className="flex items-center gap-2 text-xs font-semibold text-[#ededed]">
                    <Sparkles className="w-4 h-4 text-emerald-400" />
                    <span>Data yang Otomatis Diambil oleh Caelus Agent:</span>
                  </div>

                  <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 text-[11px]">
                    <div className="p-2.5 rounded-lg bg-[#1a1a1a] border border-[#262626] flex items-center gap-2">
                      <Cpu className="w-3.5 h-3.5 text-cyan-400 shrink-0" />
                      <div>
                        <span className="text-[#707070] block text-[10px]">CPU Cores</span>
                        <span className="font-medium text-[#ededed]">Auto-Sync</span>
                      </div>
                    </div>

                    <div className="p-2.5 rounded-lg bg-[#1a1a1a] border border-[#262626] flex items-center gap-2">
                      <Activity className="w-3.5 h-3.5 text-purple-400 shrink-0" />
                      <div>
                        <span className="text-[#707070] block text-[10px]">RAM Capacity</span>
                        <span className="font-medium text-[#ededed]">Auto-Sync</span>
                      </div>
                    </div>

                    <div className="p-2.5 rounded-lg bg-[#1a1a1a] border border-[#262626] flex items-center gap-2">
                      <HardDrive className="w-3.5 h-3.5 text-amber-400 shrink-0" />
                      <div>
                        <span className="text-[#707070] block text-[10px]">SSD / Disk</span>
                        <span className="font-medium text-[#ededed]">Auto-Sync</span>
                      </div>
                    </div>

                    <div className="p-2.5 rounded-lg bg-[#1a1a1a] border border-[#262626] flex items-center gap-2">
                      <Terminal className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                      <div>
                        <span className="text-[#707070] block text-[10px]">OS & Hostname</span>
                        <span className="font-medium text-[#ededed]">Auto-Sync</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            ) : (
              /* Cloud Provider Mode: Needs explicit Specs */
              <div className="space-y-4">
                <div className="space-y-1.5">
                  <span className={AppText.label}>Penyedia Cloud (Provider Driver)</span>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                    {providers
                      .filter((p) => p.slug !== "custom")
                      .map((p) => (
                        <button
                          type="button"
                          key={p.id}
                          onClick={() => setSelectedProviderID(p.id)}
                          className={`p-3 rounded-lg border text-left flex items-center justify-between transition-all cursor-pointer ${
                            selectedProviderID === p.id
                              ? "border-emerald-500 bg-emerald-950/30 text-[#ededed] shadow-sm"
                              : "border-[#262626] bg-[#121212] text-[#a1a1a1] hover:border-[#383838] hover:text-[#ededed]"
                          }`}
                        >
                          <div className="flex items-center gap-2.5">
                            <Cloud className="h-4 w-4 text-emerald-400" />
                            <div>
                              <p className="text-xs font-semibold text-[#ededed]">{p.name}</p>
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
                      className="w-full rounded-lg border border-[#262626] bg-[#121212] px-3.5 py-2 text-xs text-[#ededed] focus:border-emerald-500 focus:outline-none transition-colors"
                    >
                      <option value="ap-southeast-1">Singapore (ap-southeast-1)</option>
                      <option value="us-east-1">N. Virginia (us-east-1)</option>
                      <option value="eu-central-1">Frankfurt (eu-central-1)</option>
                      <option value="ap-northeast-1">Tokyo (ap-northeast-1)</option>
                    </select>
                  </div>
                </div>

                {/* OS Selection for Cloud Deploy */}
                <div className="space-y-1.5">
                  <span className={AppText.label}>Sistem Operasi (OS)</span>
                  <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
                    {[
                      { id: "ubuntu-24.04", label: "Ubuntu 24.04 LTS" },
                      { id: "ubuntu-22.04", label: "Ubuntu 22.04 LTS" },
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
                            : "border-[#262626] bg-[#121212] text-[#a1a1a1] hover:border-[#383838]"
                        }`}
                      >
                        {os.label}
                      </button>
                    ))}
                  </div>
                </div>

                {/* Hardware Specifications for Cloud Deploy */}
                <div className="space-y-2 pt-2 border-t border-[#262626]">
                  <span className={AppText.label}>Spesifikasi Komputasi Instance Baru</span>
                  <div className="grid grid-cols-3 gap-3">
                    <div>
                      <label htmlFor="create-server-cpu" className="block text-[11px] text-[#a1a1a1] mb-1">
                        vCPU Cores
                      </label>
                      <select
                        id="create-server-cpu"
                        value={cpuCores}
                        onChange={(e) => setCpuCores(Number(e.target.value))}
                        className="w-full rounded-lg border border-[#262626] bg-[#121212] px-3 py-1.5 text-xs text-[#ededed] focus:border-emerald-500 focus:outline-none"
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
                        className="w-full rounded-lg border border-[#262626] bg-[#121212] px-3 py-1.5 text-xs text-[#ededed] focus:border-emerald-500 focus:outline-none"
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
                        className="w-full rounded-lg border border-[#262626] bg-[#121212] px-3 py-1.5 text-xs text-[#ededed] focus:border-emerald-500 focus:outline-none"
                      >
                        <option value={25}>25 GB SSD</option>
                        <option value={50}>50 GB SSD</option>
                        <option value={100}>100 GB SSD</option>
                        <option value={200}>200 GB SSD</option>
                      </select>
                    </div>
                  </div>
                </div>
              </div>
            )}

            {/* Modal Actions */}
            <div className="flex items-center justify-end gap-3 pt-4 border-t border-[#262626]">
              <Button variant="outline" type="button" onClick={onClose} disabled={isLoading}>
                Batal
              </Button>
              <Button type="submit" isLoading={isLoading}>
                {tab === "byos" ? (
                  <>
                    <Terminal className="h-4 w-4" />
                    <span>Lanjut & Dapatkan Perintah Install</span>
                  </>
                ) : (
                  <>
                    <ServerIcon className="h-4 w-4" />
                    <span>Deploy Instance</span>
                  </>
                )}
              </Button>
            </div>
          </form>
        </div>
      )}
    </Dialog>
  );
};
