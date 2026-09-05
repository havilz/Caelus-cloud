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
import { useRoleGuard } from "@/hooks/useRoleGuard";

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
    slug: "cloudflare",
    name: "Cloudflare",
    desc: "R2 Object Storage (Zero Egress), Cloudflare Tunnel & Dynamic DNS Management.",
    badge: "R2 & DNS",
    badgeStyle: "text-amber-400 border-amber-500/30 bg-amber-950/20",
    iconBoxStyle: AppTheme.controls.iconBoxAmber,
  },
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
    desc: "High-performance, cost-effective cloud VPS in European & US datacenters.",
    badge: "Hetzner API",
    badgeStyle: "text-rose-400 border-rose-500/30 bg-rose-950/20",
    iconBoxStyle: "p-2.5 rounded-lg bg-rose-500/10 border border-rose-500/20 text-rose-400",
  },
  {
    slug: "digitalocean",
    name: "DigitalOcean",
    desc: "Instant Droplet provisioning with API v2 integration and VPC networking.",
    badge: "DO Droplets",
    badgeStyle: "text-cyan-400 border-cyan-500/30 bg-cyan-950/20",
    iconBoxStyle: AppTheme.controls.iconBoxCyan,
  },
  {
    slug: "contabo",
    name: "Contabo Cloud",
    desc: "Cloud VPS with high core counts and generous NVMe storage capacity.",
    badge: "Contabo API",
    badgeStyle: "text-purple-400 border-purple-500/30 bg-purple-950/20",
    iconBoxStyle: AppTheme.controls.iconBoxPurple,
  },
  {
    slug: "custom",
    name: "Custom / BYOS",
    desc: "Connect bare-metal servers, home servers, or existing VPS via Caelus Agent.",
    badge: "On-Premises",
    badgeStyle: "text-emerald-400 border-emerald-500/30 bg-emerald-950/20",
    iconBoxStyle: AppTheme.controls.iconBoxEmerald,
  },
  {
    slug: "mock",
    name: "Mock Cloud Sandbox",
    desc: "Local provider simulator for testing VM lifecycles without API costs.",
    badge: "Sandbox",
    badgeStyle: "text-purple-400 border-purple-500/30 bg-purple-950/20",
    iconBoxStyle: AppTheme.controls.iconBoxPurple,
  },
];

export default function CloudProvidersPage() {
  const { canManageCredentials } = useRoleGuard();
  const [providers, setProviders] = useState<Provider[]>([]);
  const [credentials, setCredentials] = useState<Credential[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedProviderSlug, setSelectedProviderSlug] = useState<string>("aws");
  const [credName, setCredName] = useState("");
  const [accountId, setAccountId] = useState("");
  const [apiKey, setApiKey] = useState("");
  const [apiSecret, setApiSecret] = useState("");
  const [region, setRegion] = useState("us-east-1");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const [testResults, setTestResults] = useState<Record<string, { status?: string; count?: number; testing?: boolean; message?: string }>>({});

  const loadData = async () => {
    try {
      const [provData, credData] = await Promise.all([
        providerService.listProviders().catch(() => []),
        credentialService.listCredentials().catch(() => []),
      ]);
      setProviders(Array.isArray(provData) ? provData : []);
      setCredentials(Array.isArray(credData) ? credData : []);
    } catch (err) {
      console.error("Failed to load provider data:", err);
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
    setAccountId("");
    setApiKey("");
    setApiSecret("");
    if (presetSlug) {
      setSelectedProviderSlug(presetSlug);
      if (presetSlug === "cloudflare") setRegion("auto");
      else if (presetSlug === "aws") setRegion("us-east-1");
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
      setFormError("Credential name is required.");
      return;
    }

    const targetProvider = providers.find((p) => p.slug === selectedProviderSlug);
    if (!targetProvider) {
      setFormError("Invalid or unregistered provider.");
      return;
    }

    setIsSubmitting(true);
    setFormError(null);

    const cleanInput = (val: string) => {
      let res = val.trim();
      if (res.includes(":")) {
        res = res.split(":").pop()?.trim() || res;
      }
      res = res.replace(/^https?:\/\//, "");
      if (res.includes(".")) {
        res = res.split(".")[0].trim();
      }
      return res.trim();
    };

    const cleanAccId = selectedProviderSlug === "cloudflare" ? cleanInput(accountId) : accountId.trim();
    const cleanKey = apiKey.trim();
    const cleanSecret = apiSecret.trim();

    try {
      await credentialService.createCredential({
        provider_id: targetProvider.id,
        name: credName.trim(),
        api_key: cleanKey,
        api_secret: cleanSecret,
        metadata: {
          region: region.trim(),
          account_id: cleanAccId || undefined,
          endpoint: cleanAccId ? `https://${cleanAccId}.r2.cloudflarestorage.com` : undefined,
          added_via: "web_dashboard",
        },
      });

      setIsModalOpen(false);
      await loadData();
    } catch (err: any) {
      const errMsg = err.response?.data?.errors || err.response?.data?.message || err.message || "Failed to save credentials";
      setFormError(typeof errMsg === 'string' ? errMsg : JSON.stringify(errMsg));
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDeleteCredential = async (id: string) => {
    if (!confirm("Are you sure you want to delete this provider credential?")) return;
    try {
      await credentialService.deleteCredential(id);
      setCredentials((prev) => prev.filter((c) => c.id !== id));
    } catch (err) {
      alert("Failed to delete credential.");
    }
  };

  const handleTestConnection = async (id: string) => {
    setTestResults((prev) => ({
      ...prev,
      [id]: { testing: true },
    }));

    try {
      const res = await credentialService.testCredential(id);
      setTestResults((prev) => ({
        ...prev,
        [id]: { testing: false, status: res.status, count: res.server_count ?? 0 },
      }));
    } catch (err: any) {
      setTestResults((prev) => ({
        ...prev,
        [id]: { testing: false, status: "failed", message: err.message },
      }));
    }
  };

  const safeProviders = providers || [];
  const safeCredentials = credentials || [];

  return (
    <div className={AppTheme.containers.pageWrapper}>
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className={AppTheme.text.h1}>Cloud Providers & API Credentials</h1>
          <p className={AppTheme.text.subtitle}>
            Multi-cloud infrastructure integrations, automated provisioning keys, and hybrid VPS orchestrators.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={loadData}
            disabled={isRefreshing}
            className={AppTheme.controls.buttonSecondary}
          >
            <RefreshCw className={`h-3.5 w-3.5 ${isRefreshing ? "animate-spin" : ""}`} />
            <span>Sync Providers</span>
          </button>

          {canManageCredentials && (
            <button
              type="button"
              onClick={() => handleOpenAddModal()}
              className={AppTheme.controls.buttonPrimary}
            >
              <Plus className="h-4 w-4" />
              <span>Add Credential</span>
            </button>
          )}
        </div>
      </div>

      {!canManageCredentials && (
        <div className="p-4 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-300 flex items-center gap-3 text-sm">
          <Lock className="h-5 w-5 shrink-0 text-amber-400" />
          <span>
            <strong>Restricted Access:</strong> Only organization members with <strong>Admin</strong> or <strong>Owner</strong> roles are authorized to manage cloud provider credentials.
          </span>
        </div>
      )}

      <div className={`p-4 rounded-xl ${AppTheme.colors.brand.primaryLight} flex flex-col sm:flex-row sm:items-center justify-between gap-3`}>
        <div className="flex items-start sm:items-center gap-3">
          <div className={AppTheme.controls.iconBoxEmerald}>
            <ShieldCheck className="h-5 w-5" />
          </div>
          <div>
            <h4 className={AppTheme.text.h4}>Protected Credential Encryption (At Rest & In Transit)</h4>
            <p className={AppTheme.text.subtitle}>
              API keys, secret keys, and access tokens are encrypted using <strong className="text-emerald-400">AES-256-GCM</strong> before persisting to the database.
            </p>
          </div>
        </div>
        <div className="flex items-center gap-1.5 text-xs text-emerald-400 font-mono bg-zinc-900/80 px-3 py-1.5 rounded-lg border border-emerald-500/30 shrink-0">
          <Lock className="h-3.5 w-3.5" />
          <span>AES-256 Verified</span>
        </div>
      </div>

      <div className={AppTheme.containers.metricsGrid}>
        <div className={`${AppTheme.containers.card} ${AppTheme.containers.cardHover} ${AppTheme.containers.cardContent} flex items-center justify-between`}>
          <div>
            <p className={AppTheme.text.caption}>Connected Providers</p>
            <h3 className={`${AppTheme.text.h1} mt-1`}>
              {new Set(safeCredentials.map((c) => c.provider?.slug || c.provider_id)).size}
            </h3>
            <p className={AppTheme.text.caption}>From {safeProviders.length} registered providers</p>
          </div>
          <div className={AppTheme.controls.iconBoxCyan}>
            <Cloud className="h-5 w-5" />
          </div>
        </div>

        <div className={`${AppTheme.containers.card} ${AppTheme.containers.cardHover} ${AppTheme.containers.cardContent} flex items-center justify-between`}>
          <div>
            <p className={AppTheme.text.caption}>Total Credentials</p>
            <h3 className={`${AppTheme.text.h1} mt-1`}>{safeCredentials.length}</h3>
            <p className={AppTheme.text.caption}>Encrypted AES-256-GCM</p>
          </div>
          <div className={AppTheme.controls.iconBoxEmerald}>
            <Key className="h-5 w-5" />
          </div>
        </div>

        <div className={`${AppTheme.containers.card} ${AppTheme.containers.cardHover} ${AppTheme.containers.cardContent} flex items-center justify-between`}>
          <div>
            <p className={AppTheme.text.caption}>Reconciliation Status</p>
            <h3 className="text-2xl font-bold text-emerald-400 mt-1">Active</h3>
            <p className={AppTheme.text.caption}>Polling interval 60s</p>
          </div>
          <div className={AppTheme.controls.iconBoxPurple}>
            <Radio className="h-5 w-5 animate-pulse" />
          </div>
        </div>

        <div className={`${AppTheme.containers.card} ${AppTheme.containers.cardHover} ${AppTheme.containers.cardContent} flex items-center justify-between`}>
          <div>
            <p className={AppTheme.text.caption}>Supported Drivers</p>
            <h3 className="text-2xl font-bold text-amber-400 mt-1">6 Drivers</h3>
            <p className={AppTheme.text.caption}>AWS, Hetzner, DO, Contabo, BYOS, Mock</p>
          </div>
          <div className={AppTheme.controls.iconBoxAmber}>
            <Zap className="h-5 w-5" />
          </div>
        </div>
      </div>

      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <h3 className={AppTheme.text.h3}>Supported Cloud Providers Catalog</h3>
          <span className={AppTheme.text.caption}>Select a provider to connect your account</span>
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
                      <span className="text-emerald-400 font-medium">{connectedCount} Connected</span>
                    ) : (
                      <span>Not connected</span>
                    )}
                  </div>
                  {canManageCredentials ? (
                    <button
                      type="button"
                      onClick={() => handleOpenAddModal(preset.slug)}
                      className={AppTheme.controls.buttonSecondary}
                    >
                      <Plus className="h-3 w-3" />
                      <span>Connect</span>
                    </button>
                  ) : (
                    <span className="text-xs text-zinc-500">Locked</span>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </div>

      <div className={AppTheme.containers.card}>
        <div className={`${AppTheme.containers.cardHeader} flex items-center justify-between`}>
          <div>
            <h3 className={AppTheme.text.h3}>Configured Provider Credentials</h3>
            <p className={AppTheme.text.subtitle}>
              Active credentials available for VM provisioning and auto-sync reconciliation.
            </p>
          </div>
          <span className={AppTheme.controls.badgeMono}>
            {safeCredentials.length} Credentials
          </span>
        </div>

        <div className="p-0">
          {safeCredentials.length === 0 ? (
            <div className="p-8 text-center">
              <Key className="h-8 w-8 text-zinc-600 mx-auto mb-2" />
              <p className={AppTheme.text.h4}>No provider credentials added yet</p>
              <p className={`${AppTheme.text.subtitle} max-w-md mx-auto mt-1`}>
                Add your cloud provider API keys or tokens to initiate automated VM provisioning.
              </p>
              <button
                type="button"
                onClick={() => handleOpenAddModal()}
                className={`${AppTheme.controls.buttonPrimary} mx-auto mt-4`}
              >
                <Plus className="h-4 w-4" />
                <span>Add First Credential</span>
              </button>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead className="bg-[#141414] dark:bg-[#141414] light:bg-[#f4f4f5] text-[#a1a1a1] font-medium border-b border-[#262626]">
                  <tr>
                    <th className="py-3 px-4">Credential Name</th>
                    <th className="py-3 px-4">Provider</th>
                    <th className="py-3 px-4">Region / Location</th>
                    <th className="py-3 px-4">Encryption Status</th>
                    <th className="py-3 px-4">Connection</th>
                    <th className="py-3 px-4 text-right">Actions</th>
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
                              Testing connection...
                            </span>
                          ) : testState?.status === "connected" ? (
                            <span className="inline-flex items-center gap-1 text-[11px] text-emerald-400 font-medium">
                              <CheckCircle2 className="h-3.5 w-3.5" />
                              Connected ({testState.count ?? 0} VMs)
                            </span>
                          ) : testState?.status === "failed" ? (
                            <span className="inline-flex items-center gap-1 text-[11px] text-rose-400 font-medium">
                              <AlertCircle className="h-3.5 w-3.5" />
                              Connection failed
                            </span>
                          ) : (
                            <span className={AppTheme.text.caption}>Ready to test</span>
                          )}
                        </td>
                        <td className="py-3 px-4 text-right">
                          <div className="flex items-center justify-end gap-2">
                            {canManageCredentials ? (
                              <>
                                <button
                                  type="button"
                                  onClick={() => handleTestConnection(cred.id)}
                                  disabled={testState?.testing}
                                  className={AppTheme.controls.buttonAction}
                                >
                                  <Zap className="h-3 w-3" />
                                  <span>Test</span>
                                </button>
                                <button
                                  type="button"
                                  onClick={() => handleDeleteCredential(cred.id)}
                                  className={AppTheme.controls.iconButtonDanger}
                                  title="Delete credential"
                                >
                                  <Trash2 className="h-3.5 w-3.5" />
                                </button>
                              </>
                            ) : (
                              <span className="text-[11px] text-zinc-500 italic font-mono">Read-only</span>
                            )}
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

      <Dialog
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title="Connect Cloud Provider Credentials"
        description="Credentials are stored with AES-256-GCM encryption and only decrypted during instance provisioning."
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
              <label className={AppTheme.text.label}>Select Provider</label>
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
              <label className={AppTheme.text.label}>Credential Name / Label</label>
              <input
                type="text"
                placeholder="e.g. Production AWS Account, Hetzner Production Token"
                value={credName}
                onChange={(e) => setCredName(e.target.value)}
                required
                className={AppTheme.controls.input}
              />
            </div>

            {selectedProviderSlug === "cloudflare" && (
              <>
                <div className="space-y-1.5">
                  <label className={AppTheme.text.label}>Cloudflare Account ID</label>
                  <input
                    type="text"
                    placeholder="e.g. c0356789abcdef0123456789abcdef01"
                    value={accountId}
                    onChange={(e) => setAccountId(e.target.value.trim())}
                    required
                    className={AppTheme.controls.inputMono}
                  />
                  <p className="text-[11px] text-[#707070]">
                    Found in R2 overview / S3 URL endpoint:
                  </p>
                </div>
                <div className="space-y-1.5">
                  <label className={AppTheme.text.label}>Cloudflare R2 Access Key ID</label>
                  <input
                    type="text"
                    placeholder="e.g. 0123456789abcdef0123456789abcdef"
                    value={apiKey}
                    onChange={(e) => setApiKey(e.target.value.trim())}
                    required
                    className={AppTheme.controls.inputMono}
                  />
                </div>
                <div className="space-y-1.5">
                  <label className={AppTheme.text.label}>Cloudflare Secret Access Key</label>
                  <input
                    type="password"
                    placeholder="e.g. 0123456789abcdef0123456789abcdef01234567"
                    value={apiSecret}
                    onChange={(e) => setApiSecret(e.target.value.trim())}
                    required
                    className={AppTheme.controls.inputMono}
                  />
                </div>
                <div className="space-y-1.5">
                  <label className={AppTheme.text.label}>Default Region / Scope</label>
                  <input
                    type="text"
                    placeholder="auto (Global Anycast Edge)"
                    value={region}
                    onChange={(e) => setRegion(e.target.value)}
                    className={AppTheme.controls.inputMono}
                  />
                </div>
              </>
            )}

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
                    placeholder="API token or secret token"
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
                <label className={AppTheme.text.label}>Simulation Region</label>
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
                Cancel
              </button>
              <button
                type="submit"
                disabled={isSubmitting}
                className={AppTheme.controls.buttonPrimary}
              >
                <Lock className="h-3.5 w-3.5" />
                <span>{isSubmitting ? "Saving..." : "Encrypt & Save"}</span>
              </button>
            </div>
          </form>
        </div>
      </Dialog>
    </div>
  );
}
