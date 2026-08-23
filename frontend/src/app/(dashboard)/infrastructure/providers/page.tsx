"use client";

import React, { useState, useEffect } from "react";
import {
  Cloud,
  Plus,
  RefreshCw,
  ShieldCheck,
  CheckCircle2,
  AlertCircle,
  Key,
  Trash2,
  Zap,
  Lock,
  Radio,
  Check,
} from "lucide-react";
import { Dialog } from "@/components/ui/dialog";
import { AppTheme } from "@/core/theme";
import { providerService } from "@/services/provider.service";
import { credentialService } from "@/services/credential.service";
import { Provider } from "@/types/server";
import { Credential } from "@/types/credential";

interface ProviderPreset {
  slug: string;
  name: string;
  desc: string;
  badge: string;
  badgeStyle: string;
  iconBoxStyle: string;
}

const PROVIDER_PRESETS: ProviderPreset[] = [
  {
    slug: "aws",
    name: "Amazon Web Services",
    desc: "Deploy dan orkestrasi instance AWS EC2 di berbagai region global.",
    badge: "AWS EC2",
    badgeStyle: "text-amber-400 border-amber-500/30 bg-amber-950/20",
    iconBoxStyle: AppTheme.controls.iconBoxAmber,
  },
  {
    slug: "hetzner",
    name: "Hetzner Cloud",
    desc: "Cloud VPS performa tinggi dengan biaya ekonomis di datacenter Eropa & AS.",
    badge: "Hetzner API",
    badgeStyle: "text-rose-400 border-rose-500/30 bg-rose-950/20",
    iconBoxStyle: "p-2.5 rounded-lg bg-rose-500/10 border border-rose-500/20 text-rose-400",
  },
  {
    slug: "digitalocean",
    name: "DigitalOcean",
    desc: "Provisioning Droplet instan dengan integrasi API v2 dan jaringan VPC.",
    badge: "DO Droplets",
    badgeStyle: "text-cyan-400 border-cyan-500/30 bg-cyan-950/20",
    iconBoxStyle: AppTheme.controls.iconBoxCyan,
  },
  {
    slug: "contabo",
    name: "Contabo Cloud",
    desc: "Cloud VPS dengan kapasitas core CPU dan storage NVMe ekstra lega.",
    badge: "Contabo API",
    badgeStyle: "text-purple-400 border-purple-500/30 bg-purple-950/20",
    iconBoxStyle: AppTheme.controls.iconBoxPurple,
  },
  {
    slug: "custom",
    name: "Custom / BYOS",
    desc: "Hubungkan server fisik, Home Server, atau VPS existing via Caelus Agent.",
    badge: "On-Premises",
    badgeStyle: "text-emerald-400 border-emerald-500/30 bg-emerald-950/20",
    iconBoxStyle: AppTheme.controls.iconBoxEmerald,
  },
  {
    slug: "mock",
    name: "Mock Cloud Sandbox",
    desc: "Simulator provider lokal untuk pengujian siklus hidup VM tanpa biaya API.",
    badge: "Sandbox",
    badgeStyle: "text-purple-400 border-purple-500/30 bg-purple-950/20",
    iconBoxStyle: AppTheme.controls.iconBoxPurple,
  },
];

export default function CloudProvidersPage() {
  const [providers, setProviders] = useState<Provider[]>([]);
  const [credentials, setCredentials] = useState<Credential[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);

  // Modal State
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedProviderSlug, setSelectedProviderSlug] = useState<string>("aws");
  const [credName, setCredName] = useState("");
  const [apiKey, setApiKey] = useState("");
  const [apiSecret, setApiSecret] = useState("");
  const [region, setRegion] = useState("us-east-1");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  // Test Connection State map (credentialId -> { status, count, testing })
  const [testResults, setTestResults] = useState<Record<string, { status: string; count?: number; testing?: boolean }>>({});

  const loadData = async () => {
    try {
      const [provData, credData] = await Promise.all([
        providerService.listProviders().catch(() => []),
        credentialService.listCredentials().catch(() => []),
      ]);
      setProviders(Array.isArray(provData) ? provData : []);
      setCredentials(Array.isArray(credData) ? credData : []);
    } catch (err) {
      console.error("Gagal memuat data provider:", err);
      setProviders([]);
      setCredentials([]);
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleOpenAddModal = (presetSlug?: string) => {
    setFormError(null);
    setCredName("");
    setApiKey("");
    setApiSecret("");
    if (presetSlug) {
      setSelectedProviderSlug(presetSlug);
      if (presetSlug === "aws") setRegion("us-east-1");
      else if (presetSlug === "hetzner") setRegion("fsn1");
      else if (presetSlug === "digitalocean") setRegion("sgp1");
      else if (presetSlug === "contabo") setRegion("EU");
    } else {
      setSelectedProviderSlug("aws");
      setRegion("us-east-1");
    }
    setIsModalOpen(true);
  };

  const handleCreateCredential = async (e: React.SyntheticEvent) => {
    e.preventDefault();
    if (!credName.trim()) {
      setFormError("Nama kredensial wajib diisi.");
      return;
    }

    const targetProvider = providers.find((p) => p.slug === selectedProviderSlug);
    if (!targetProvider) {
      setFormError("Provider tidak valid atau belum terdaftar di sistem.");
      return;
    }

    setIsSubmitting(true);
    setFormError(null);

    try {
      await credentialService.createCredential({
        provider_id: targetProvider.id,
        name: credName.trim(),
        api_key: apiKey.trim(),
        api_secret: apiSecret.trim(),
        metadata: {
          region: region.trim(),
          added_via: "web_dashboard",
        },
      });

      setIsModalOpen(false);
      await loadData();
    } catch (err: any) {
      setFormError(err.response?.data?.message || err.message || "Gagal menyimpan kredensial");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDeleteCredential = async (id: string) => {
    if (!confirm("Apakah Anda yakin ingin menghapus kredensial provider ini?")) return;
    try {
      await credentialService.deleteCredential(id);
      setCredentials((prev) => prev.filter((c) => c.id !== id));
    } catch (err) {
      alert("Gagal menghapus kredensial.");
    }
  };

  const handleTestConnection = async (id: string) => {
    setTestResults((prev) => ({
      ...prev,
      [id]: { status: "testing", testing: true },
    }));

    try {
      const res = await credentialService.testCredential(id);
      setTestResults((prev) => ({
        ...prev,
        [id]: { status: "connected", count: res.server_count, testing: false },
      }));
    } catch (err: any) {
      setTestResults((prev) => ({
        ...prev,
        [id]: { status: "failed", testing: false },
      }));
    }
  };

  const safeCredentials = Array.isArray(credentials) ? credentials : [];
  const safeProviders = Array.isArray(providers) ? providers : [];

  return (
    <div className={AppTheme.containers.pageWrapper}>
      {/* Page Header */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <span className={AppTheme.text.categoryTag}>Cloud Infrastructure</span>
            <span className={AppTheme.controls.badgeActive}>Orchestrator</span>
          </div>
          <h1 className={AppTheme.text.h1}>Multi-Provider Infrastructure</h1>
          <p className={AppTheme.text.subtitle}>
            Hubungkan akun cloud provider (AWS, Hetzner, DigitalOcean, Contabo) dengan enkripsi AES-256-GCM tingkat enterprise.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={() => {
              setIsRefreshing(true);
              loadData();
            }}
            disabled={isRefreshing}
            className={AppTheme.controls.buttonSecondary}
          >
            <RefreshCw className={`h-4 w-4 ${isRefreshing ? "animate-spin" : ""}`} />
            <span>Sinkronisasi</span>
          </button>

          <button
            type="button"
            onClick={() => handleOpenAddModal()}
            className={AppTheme.controls.buttonPrimary}
          >
            <Plus className="h-4 w-4" />
            <span>Tambah Kredensial</span>
          </button>
        </div>
      </div>

      {/* Security Banner */}
      <div className={`p-4 rounded-xl ${AppTheme.colors.brand.primaryLight} flex flex-col sm:flex-row sm:items-center justify-between gap-3`}>
        <div className="flex items-start sm:items-center gap-3">
          <div className={AppTheme.controls.iconBoxEmerald}>
            <ShieldCheck className="h-5 w-5" />
          </div>
          <div>
            <h4 className={AppTheme.text.h4}>Enkripsi Data Kredensial Terproteksi (At Rest & In Transit)</h4>
            <p className={AppTheme.text.subtitle}>
              Kunci API, secret key, dan access token dienkripsi menggunakan algoritma <strong className="text-emerald-400">AES-256-GCM</strong> sebelum disimpan ke basis data.
            </p>
          </div>
        </div>
        <div className="flex items-center gap-1.5 text-xs text-emerald-400 font-mono bg-zinc-900/80 px-3 py-1.5 rounded-lg border border-emerald-500/30 shrink-0">
          <Lock className="h-3.5 w-3.5" />
          <span>AES-256 Verified</span>
        </div>
      </div>

      {/* Metrics Row */}
      <div className={AppTheme.containers.metricsGrid}>
        <div className={`${AppTheme.containers.card} ${AppTheme.containers.cardHover} ${AppTheme.containers.cardContent} flex items-center justify-between`}>
          <div>
            <p className={AppTheme.text.caption}>Provider Terhubung</p>
            <h3 className={`${AppTheme.text.h1} mt-1`}>
              {new Set(safeCredentials.map((c) => c.provider?.slug || c.provider_id)).size}
            </h3>
            <p className={AppTheme.text.caption}>Dari {safeProviders.length} provider terdaftar</p>
          </div>
          <div className={AppTheme.controls.iconBoxCyan}>
            <Cloud className="h-5 w-5" />
          </div>
        </div>

        <div className={`${AppTheme.containers.card} ${AppTheme.containers.cardHover} ${AppTheme.containers.cardContent} flex items-center justify-between`}>
          <div>
            <p className={AppTheme.text.caption}>Total Kredensial</p>
            <h3 className={`${AppTheme.text.h1} mt-1`}>{safeCredentials.length}</h3>
            <p className={AppTheme.text.caption}>Terenkripsi AES-256-GCM</p>
          </div>
          <div className={AppTheme.controls.iconBoxEmerald}>
            <Key className="h-5 w-5" />
          </div>
        </div>

        <div className={`${AppTheme.containers.card} ${AppTheme.containers.cardHover} ${AppTheme.containers.cardContent} flex items-center justify-between`}>
          <div>
            <p className={AppTheme.text.caption}>Status Rekonsiliasi</p>
            <h3 className="text-2xl font-bold text-emerald-400 mt-1">Aktif</h3>
            <p className={AppTheme.text.caption}>Polling interval 60s</p>
          </div>
          <div className={AppTheme.controls.iconBoxPurple}>
            <Radio className="h-5 w-5 animate-pulse" />
          </div>
        </div>

        <div className={`${AppTheme.containers.card} ${AppTheme.containers.cardHover} ${AppTheme.containers.cardContent} flex items-center justify-between`}>
          <div>
            <p className={AppTheme.text.caption}>Driver Didukung</p>
            <h3 className="text-2xl font-bold text-amber-400 mt-1">6 Driver</h3>
            <p className={AppTheme.text.caption}>AWS, Hetzner, DO, Contabo, BYOS, Mock</p>
          </div>
          <div className={AppTheme.controls.iconBoxAmber}>
            <Zap className="h-5 w-5" />
          </div>
        </div>
      </div>

      {/* Provider Catalog Cards */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <h3 className={AppTheme.text.h3}>Katalog Cloud Provider yang Didukung</h3>
          <span className={AppTheme.text.caption}>Pilih provider untuk menghubungkan akun</span>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {PROVIDER_PRESETS.map((preset) => {
            const connectedCount = safeCredentials.filter(
              (c) => c.provider?.slug === preset.slug
            ).length;

            return (
              <div
                key={preset.slug}
                className={`${AppTheme.containers.card} ${AppTheme.containers.cardHover} ${AppTheme.containers.cardContent} flex flex-col justify-between h-full`}
              >
                <div>
                  <div className="flex items-center justify-between mb-3">
                    <div className={preset.iconBoxStyle}>
                      <Cloud className="h-5 w-5" />
                    </div>
                    <span className={`px-2 py-0.5 rounded-md text-[10px] font-mono border ${preset.badgeStyle}`}>
                      {preset.badge}
                    </span>
                  </div>

                  <h4 className={AppTheme.text.h4}>{preset.name}</h4>
                  <p className={`${AppTheme.text.bodySm} mt-1 line-clamp-2`}>{preset.desc}</p>
                </div>

                <div className="mt-5 pt-4 border-t border-[#262626] dark:border-[#262626] light:border-[#e5e7eb] flex items-center justify-between">
                  <div className={AppTheme.text.caption}>
                    {connectedCount > 0 ? (
                      <span className="text-emerald-400 font-medium">{connectedCount} Akun Terhubung</span>
                    ) : (
                      <span>Belum terhubung</span>
                    )}
                  </div>
                  <button
                    type="button"
                    onClick={() => handleOpenAddModal(preset.slug)}
                    className={AppTheme.controls.buttonSecondary}
                  >
                    <Plus className="h-3 w-3" />
                    <span>Hubungkan</span>
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Configured Credentials List Table */}
      <div className={AppTheme.containers.card}>
        <div className={`${AppTheme.containers.cardHeader} flex items-center justify-between`}>
          <div>
            <h3 className={AppTheme.text.h3}>Daftar Kredensial Provider Terpasang</h3>
            <p className={AppTheme.text.subtitle}>
              Kredensial aktif yang dapat digunakan untuk provisioning VM dan auto-sync status.
            </p>
          </div>
          <span className={AppTheme.controls.badgeMono}>
            {safeCredentials.length} Kredensial
          </span>
        </div>

        <div className="p-0">
          {safeCredentials.length === 0 ? (
            <div className="p-8 text-center">
              <Key className="h-8 w-8 text-zinc-600 mx-auto mb-2" />
              <p className={AppTheme.text.h4}>Belum ada kredensial provider yang ditambahkan</p>
              <p className={`${AppTheme.text.subtitle} max-w-md mx-auto mt-1`}>
                Tambahkan kunci API atau token provider cloud Anda untuk memulai provisioning server secara otomatis.
              </p>
              <button
                type="button"
                onClick={() => handleOpenAddModal()}
                className={`${AppTheme.controls.buttonPrimary} mx-auto mt-4`}
              >
                <Plus className="h-4 w-4" />
                <span>Tambah Kredensial Pertama</span>
              </button>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead className="bg-[#141414] dark:bg-[#141414] light:bg-[#f4f4f5] text-[#a1a1a1] font-medium border-b border-[#262626]">
                  <tr>
                    <th className="py-3 px-4">Nama Kredensial</th>
                    <th className="py-3 px-4">Penyedia (Provider)</th>
                    <th className="py-3 px-4">Region / Lokasi</th>
                    <th className="py-3 px-4">Status Enkripsi</th>
                    <th className="py-3 px-4">Koneksi</th>
                    <th className="py-3 px-4 text-right">Aksi</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-[#262626] dark:divide-[#262626] light:divide-[#e5e7eb] text-[#ededed]">
                  {safeCredentials.map((cred) => {
                    const testState = testResults[cred.id];
                    const provSlug = cred.provider?.slug || "custom";

                    return (
                      <tr key={cred.id} className="hover:bg-[#1f1f1f] dark:hover:bg-[#1f1f1f] light:hover:bg-[#f9fafb] transition-colors">
                        <td className="py-3 px-4 font-medium text-[#ededed]">
                          <div className="flex items-center gap-2">
                            <Key className="h-3.5 w-3.5 text-emerald-400" />
                            <span>{cred.name}</span>
                          </div>
                        </td>
                        <td className="py-3 px-4">
                          <span className={AppTheme.controls.badgeMono}>
                            {cred.provider?.name || provSlug}
                          </span>
                        </td>
                        <td className={`py-3 px-4 ${AppTheme.text.codeMuted}`}>
                          {cred.metadata?.region || cred.metadata?.location || "default"}
                        </td>
                        <td className="py-3 px-4">
                          <span className="inline-flex items-center gap-1 text-[11px] text-emerald-400 bg-emerald-950/40 px-2 py-0.5 rounded border border-emerald-800/30 font-mono">
                            <Lock className="h-3 w-3" />
                            AES-256-GCM
                          </span>
                        </td>
                        <td className="py-3 px-4">
                          {testState?.testing ? (
                            <span className="inline-flex items-center gap-1 text-[11px] text-amber-400 animate-pulse">
                              <RefreshCw className="h-3 w-3 animate-spin" />
                              Menguji koneksi...
                            </span>
                          ) : testState?.status === "connected" ? (
                            <span className="inline-flex items-center gap-1 text-[11px] text-emerald-400 font-medium">
                              <CheckCircle2 className="h-3.5 w-3.5" />
                              Terhubung ({testState.count ?? 0} VM)
                            </span>
                          ) : testState?.status === "failed" ? (
                            <span className="inline-flex items-center gap-1 text-[11px] text-rose-400 font-medium">
                              <AlertCircle className="h-3.5 w-3.5" />
                              Gagal terhubung
                            </span>
                          ) : (
                            <span className={AppTheme.text.caption}>Siap diuji</span>
                          )}
                        </td>
                        <td className="py-3 px-4 text-right">
                          <div className="flex items-center justify-end gap-2">
                            <button
                              type="button"
                              onClick={() => handleTestConnection(cred.id)}
                              disabled={testState?.testing}
                              className={AppTheme.controls.buttonAction}
                            >
                              <Zap className="h-3 w-3" />
                              <span>Uji</span>
                            </button>
                            <button
                              type="button"
                              onClick={() => handleDeleteCredential(cred.id)}
                              className={AppTheme.controls.iconButtonDanger}
                              title="Hapus kredensial"
                            >
                              <Trash2 className="h-3.5 w-3.5" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {/* Add Credential Modal */}
      <Dialog
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title="Hubungkan Kredensial Cloud Provider"
        description="Kredensial disimpan dengan enkripsi AES-256-GCM dan hanya didekripsi saat provisioning instance."
        maxWidth="lg"
      >
        <div className="space-y-5">
          {formError && (
            <div className="p-3 rounded-lg bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs flex items-center gap-2">
              <AlertCircle className="h-4 w-4 shrink-0" />
              <span>{formError}</span>
            </div>
          )}

          <form onSubmit={handleCreateCredential} className="space-y-4">
            <div className="space-y-1.5">
              <label className={AppTheme.text.label}>Pilih Provider</label>
              <div className="grid grid-cols-3 gap-2">
                {PROVIDER_PRESETS.filter((p) => p.slug !== "custom").map((p) => {
                  const isSelected = selectedProviderSlug === p.slug;
                  return (
                    <button
                      type="button"
                      key={p.slug}
                      onClick={() => {
                        setSelectedProviderSlug(p.slug);
                        if (p.slug === "aws") setRegion("us-east-1");
                        else if (p.slug === "hetzner") setRegion("fsn1");
                        else if (p.slug === "digitalocean") setRegion("sgp1");
                        else if (p.slug === "contabo") setRegion("EU");
                      }}
                      className={`p-2.5 rounded-lg border text-left text-xs transition-all cursor-pointer ${
                        isSelected
                          ? "bg-emerald-500/10 border-emerald-500 text-emerald-300 font-semibold"
                          : `${AppTheme.colors.bg.surfaceSubtle} ${AppTheme.colors.border.subtle} ${AppTheme.colors.text.secondary} hover:${AppTheme.colors.text.primary}`
                      }`}
                    >
                      <div className="flex items-center justify-between mb-1">
                        <span className="font-medium truncate">{p.badge}</span>
                        {isSelected && <Check className="h-3 w-3 text-emerald-400 shrink-0" />}
                      </div>
                      <p className={AppTheme.text.caption}>{p.name}</p>
                    </button>
                  );
                })}
              </div>
            </div>

            <div className="space-y-1.5">
              <label className={AppTheme.text.label}>Nama Kredensial / Alias</label>
              <input
                type="text"
                placeholder="e.g. Production AWS Account, Hetzner Production Token"
                value={credName}
                onChange={(e) => setCredName(e.target.value)}
                required
                className={AppTheme.controls.input}
              />
            </div>

            {selectedProviderSlug === "aws" && (
              <>
                <div className="space-y-1.5">
                  <label className={AppTheme.text.label}>AWS Access Key ID</label>
                  <input
                    type="text"
                    placeholder="AKIAIOSFODNN7EXAMPLE"
                    value={apiKey}
                    onChange={(e) => setApiKey(e.target.value)}
                    required
                    className={AppTheme.controls.inputMono}
                  />
                </div>
                <div className="space-y-1.5">
                  <label className={AppTheme.text.label}>AWS Secret Access Key</label>
                  <input
                    type="password"
                    placeholder="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
                    value={apiSecret}
                    onChange={(e) => setApiSecret(e.target.value)}
                    required
                    className={AppTheme.controls.inputMono}
                  />
                </div>
                <div className="space-y-1.5">
                  <label className={AppTheme.text.label}>Default Region</label>
                  <input
                    type="text"
                    placeholder="us-east-1, ap-southeast-1, etc."
                    value={region}
                    onChange={(e) => setRegion(e.target.value)}
                    className={AppTheme.controls.inputMono}
                  />
                </div>
              </>
            )}

            {(selectedProviderSlug === "hetzner" || selectedProviderSlug === "digitalocean") && (
              <>
                <div className="space-y-1.5">
                  <label className={AppTheme.text.label}>
                    {selectedProviderSlug === "hetzner" ? "Hetzner Cloud API Token" : "DigitalOcean Personal Access Token"}
                  </label>
                  <input
                    type="password"
                    placeholder="eyJhbGciOi... atau token rahasia"
                    value={apiKey}
                    onChange={(e) => setApiKey(e.target.value)}
                    required
                    className={AppTheme.controls.inputMono}
                  />
                </div>
                <div className="space-y-1.5">
                  <label className={AppTheme.text.label}>Default Location / Datacenter</label>
                  <input
                    type="text"
                    placeholder={selectedProviderSlug === "hetzner" ? "fsn1, nbg1, hel1" : "sgp1, nyc1, ams3"}
                    value={region}
                    onChange={(e) => setRegion(e.target.value)}
                    className={AppTheme.controls.inputMono}
                  />
                </div>
              </>
            )}

            {selectedProviderSlug === "contabo" && (
              <>
                <div className="space-y-1.5">
                  <label className={AppTheme.text.label}>Contabo Client ID</label>
                  <input
                    type="text"
                    placeholder="Contabo Client ID"
                    value={apiKey}
                    onChange={(e) => setApiKey(e.target.value)}
                    required
                    className={AppTheme.controls.inputMono}
                  />
                </div>
                <div className="space-y-1.5">
                  <label className={AppTheme.text.label}>Contabo Client Secret</label>
                  <input
                    type="password"
                    placeholder="Contabo Client Secret"
                    value={apiSecret}
                    onChange={(e) => setApiSecret(e.target.value)}
                    required
                    className={AppTheme.controls.inputMono}
                  />
                </div>
                <div className="space-y-1.5">
                  <label className={AppTheme.text.label}>Default Region</label>
                  <input
                    type="text"
                    placeholder="EU, US-central, SIN"
                    value={region}
                    onChange={(e) => setRegion(e.target.value)}
                    className={AppTheme.controls.inputMono}
                  />
                </div>
              </>
            )}

            {selectedProviderSlug === "mock" && (
              <div className="space-y-1.5">
                <label className={AppTheme.text.label}>Simulasi Region</label>
                <input
                  type="text"
                  placeholder="mock-region-1"
                  value={region}
                  onChange={(e) => setRegion(e.target.value)}
                  className={AppTheme.controls.inputMono}
                />
              </div>
            )}

            <div className="flex items-center justify-end gap-3 pt-4 border-t border-[#262626]">
              <button
                type="button"
                onClick={() => setIsModalOpen(false)}
                className={AppTheme.controls.buttonSecondary}
              >
                Batal
              </button>
              <button
                type="submit"
                disabled={isSubmitting}
                className={AppTheme.controls.buttonPrimary}
              >
                <Lock className="h-3.5 w-3.5" />
                <span>{isSubmitting ? "Menyimpan..." : "Enkripsi & Simpan"}</span>
              </button>
            </div>
          </form>
        </div>
      </Dialog>
    </div>
  );
}
