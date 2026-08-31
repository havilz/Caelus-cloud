"use client";

import React, { useState, useEffect } from "react";
import { Building2, ShieldAlert, Check, AlertCircle, RefreshCw, Layers } from "lucide-react";
import { AppTheme } from "@/core/theme";
import { settingsService } from "@/services/settings.service";
import { Organization } from "@/types/settings";

export const OrganizationTab: React.FC = () => {
  const [org, setOrg] = useState<Organization | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [successMsg, setSuccessMsg] = useState<string | null>(null);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  const fetchOrganization = async () => {
    try {
      setIsLoading(true);
      const data = await settingsService.getOrganization();
      setOrg(data);
      setName(data.name || "");
      setSlug(data.slug || "");
    } catch (err: any) {
      setErrorMsg("Gagal memuat informasi organisasi");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchOrganization();
  }, []);

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setIsSaving(true);
      setSuccessMsg(null);
      setErrorMsg(null);
      const updated = await settingsService.updateOrganization({ name, slug });
      setOrg(updated);
      setSuccessMsg("Informasi organisasi berhasil diperbarui");
      setTimeout(() => setSuccessMsg(null), 3000);
    } catch (err: any) {
      setErrorMsg(err.response?.data?.message || "Gagal memperbarui organisasi");
    } finally {
      setIsSaving(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-16 text-zinc-500">
        <RefreshCw className="h-5 w-5 animate-spin mr-2" />
        <span className="text-sm">Memuat informasi organisasi...</span>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Pengaturan Organisasi */}
      <div className={AppTheme.containers.card}>
        <div className="border-b border-[#262626] pb-4 mb-5 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400">
              <Building2 className="h-5 w-5" />
            </div>
            <div>
              <h3 className="text-sm font-semibold text-zinc-100">Profil Organisasi / Workspace</h3>
              <p className="text-xs text-zinc-400">Kelola nama workspace dan identitas URL tim Anda</p>
            </div>
          </div>
          <span className="px-3 py-1 rounded-full text-xs font-semibold uppercase tracking-wider bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            {org?.tier || "Free"} Tier
          </span>
        </div>

        {successMsg && (
          <div className="mb-4 p-3 rounded-lg bg-emerald-950/40 border border-emerald-500/30 text-emerald-400 text-xs flex items-center gap-2">
            <Check className="h-4 w-4 shrink-0" />
            <span>{successMsg}</span>
          </div>
        )}

        {errorMsg && (
          <div className="mb-4 p-3 rounded-lg bg-rose-950/40 border border-rose-500/30 text-rose-400 text-xs flex items-center gap-2">
            <AlertCircle className="h-4 w-4 shrink-0" />
            <span>{errorMsg}</span>
          </div>
        )}

        <form onSubmit={handleUpdate} className="space-y-4 max-w-xl">
          <div>
            <label className="block text-xs font-medium text-zinc-300 mb-1.5">ID Organisasi</label>
            <input
              type="text"
              disabled
              value={org?.id || ""}
              className="w-full bg-[#181818] border border-[#2e2e2e] text-zinc-400 font-mono text-xs rounded-lg px-3 py-2 cursor-not-allowed"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-zinc-300 mb-1.5">Nama Organisasi</label>
            <input
              type="text"
              required
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Contoh: Acme Corporation"
              className="w-full bg-[#141414] border border-[#2e2e2e] text-zinc-200 text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-emerald-500 transition-colors"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-zinc-300 mb-1.5">Slug URL Workspace</label>
            <input
              type="text"
              required
              value={slug}
              onChange={(e) => setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, "-"))}
              placeholder="acme-corp"
              className="w-full bg-[#141414] border border-[#2e2e2e] text-zinc-200 font-mono text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-emerald-500 transition-colors"
            />
            <p className="text-[11px] text-zinc-400 mt-1">
              Digunakan sebagai penanda unik URL routing API dan integrasi webhook
            </p>
          </div>

          <button
            type="submit"
            disabled={isSaving}
            className="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-zinc-950 font-semibold text-xs rounded-lg transition-colors flex items-center gap-2 cursor-pointer disabled:opacity-50"
          >
            {isSaving && <RefreshCw className="h-3.5 w-3.5 animate-spin" />}
            <span>Simpan Perubahan</span>
          </button>
        </form>
      </div>

      {/* Danger Zone */}
      <div className="p-5 rounded-xl bg-rose-950/10 border border-rose-500/20">
        <div className="flex items-center gap-3 border-b border-rose-500/20 pb-4 mb-4">
          <div className="p-2 rounded-lg bg-rose-500/10 border border-rose-500/20 text-rose-400">
            <ShieldAlert className="h-5 w-5" />
          </div>
          <div>
            <h3 className="text-sm font-semibold text-rose-400">Zona Kritis (Danger Zone)</h3>
            <p className="text-xs text-zinc-400">Tindakan berikut bersifat destruktif dan tidak dapat dibatalkan</p>
          </div>
        </div>

        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 py-2">
          <div>
            <p className="text-xs font-medium text-zinc-200">Hapus Workspace Organisasi</p>
            <p className="text-[11px] text-zinc-400">
              Menghapus seluruh konfigurasi VPS, volume storage, jaringan VPC, dan data telemetri terkait secara permanen.
            </p>
          </div>
          <button
            type="button"
            onClick={() => alert("Untuk keamanan tingkat tinggi, hubungi Master Admin untuk menghapus workspace ini.")}
            className="px-3.5 py-1.5 bg-rose-600/20 hover:bg-rose-600/30 text-rose-400 border border-rose-500/30 text-xs font-semibold rounded-lg transition-colors cursor-pointer shrink-0"
          >
            Hapus Organisasi
          </button>
        </div>
      </div>
    </div>
  );
};
