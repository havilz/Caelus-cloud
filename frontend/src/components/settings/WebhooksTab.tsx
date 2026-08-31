"use client";

import React, { useState, useEffect } from "react";
import { Send, Plus, Trash2, Check, AlertCircle, RefreshCw, X, Play, Globe } from "lucide-react";
import { AppTheme } from "@/core/theme";
import { settingsService } from "@/services/settings.service";
import { Webhook } from "@/types/settings";

const AVAILABLE_EVENTS = [
  { id: "server.down", label: "Server Offline / Down" },
  { id: "alert.triggered", label: "Threshold Alert Triggered" },
  { id: "backup.failed", label: "Backup Policy Failed" },
  { id: "deployment.finished", label: "IaC Deployment Finished" },
  { id: "security.threat", label: "Sentinel Threat Detected" },
];

export const WebhooksTab: React.FC = () => {
  const [webhooks, setWebhooks] = useState<Webhook[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

  // Create Modal
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [name, setName] = useState("");
  const [url, setUrl] = useState("");
  const [secret, setSecret] = useState("");
  const [selectedEvents, setSelectedEvents] = useState<string[]>([
    "server.down",
    "alert.triggered",
    "backup.failed",
  ]);
  const [isCreating, setIsCreating] = useState(false);
  const [modalError, setModalError] = useState<string | null>(null);

  // Testing Webhook state
  const [testingId, setTestingId] = useState<string | null>(null);

  const fetchWebhooks = async () => {
    try {
      setIsLoading(true);
      const data = await settingsService.listWebhooks();
      setWebhooks(data || []);
    } catch (err: any) {
      setErrorMsg("Gagal memuat daftar webhook");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchWebhooks();
  }, []);

  const handleToggleEvent = (eventId: string) => {
    if (selectedEvents.includes(eventId)) {
      setSelectedEvents(selectedEvents.filter((e) => e !== eventId));
    } else {
      setSelectedEvents([...selectedEvents, eventId]);
    }
  };

  const handleCreateWebhook = async (e: React.FormEvent) => {
    e.preventDefault();
    if (selectedEvents.length === 0) {
      setModalError("Pilih minimal satu event trigger");
      return;
    }

    try {
      setIsCreating(true);
      setModalError(null);
      await settingsService.createWebhook({
        name,
        url,
        secret: secret.trim() ? secret : undefined,
        events: selectedEvents,
      });

      setIsCreateModalOpen(false);
      setName("");
      setUrl("");
      setSecret("");
      setSuccessMsg("Webhook berhasil didaftarkan");
      setTimeout(() => setSuccessMsg(null), 3000);
      fetchWebhooks();
    } catch (err: any) {
      setModalError(err.response?.data?.message || "Gagal membuat webhook");
    } finally {
      setIsCreating(false);
    }
  };

  const handleTestWebhook = async (id: string) => {
    try {
      setTestingId(id);
      const res = await settingsService.testWebhook(id);
      setSuccessMsg(`Test ping berhasil dikirim (HTTP Status: ${res.http_status})`);
      setTimeout(() => setSuccessMsg(null), 4000);
      fetchWebhooks();
    } catch (err: any) {
      setErrorMsg(err.response?.data?.message || "Gagal mengirim test ping");
      setTimeout(() => setErrorMsg(null), 4000);
    } finally {
      setTestingId(null);
    }
  };

  const handleDeleteWebhook = async (id: string, name: string) => {
    if (!confirm(`Hapus webhook "${name}"?`)) return;
    try {
      await settingsService.deleteWebhook(id);
      setSuccessMsg("Webhook berhasil dihapus");
      setTimeout(() => setSuccessMsg(null), 3000);
      fetchWebhooks();
    } catch (err: any) {
      setErrorMsg("Gagal menghapus webhook");
      setTimeout(() => setErrorMsg(null), 3000);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-16 text-zinc-500">
        <RefreshCw className="h-5 w-5 animate-spin mr-2" />
        <span className="text-sm">Memuat integrasi webhook...</span>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header & Aksi */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h3 className="text-sm font-semibold text-zinc-100">Outgoing Webhooks & Alert Dispatchers</h3>
          <p className="text-xs text-zinc-400 mt-0.5">
            Kirimkan notifikasi HTTP POST real-time ke Discord, Slack, Telegram bot, atau server Anda sendiri saat terjadi event penting
          </p>
        </div>
        <button
          type="button"
          onClick={() => setIsCreateModalOpen(true)}
          className="px-3.5 py-2 bg-emerald-600 hover:bg-emerald-500 text-zinc-950 font-semibold text-xs rounded-lg transition-colors flex items-center gap-2 cursor-pointer shrink-0"
        >
          <Plus className="h-4 w-4" />
          <span>Tambah Webhook</span>
        </button>
      </div>

      {successMsg && (
        <div className="p-3 rounded-lg bg-emerald-950/40 border border-emerald-500/30 text-emerald-400 text-xs flex items-center gap-2">
          <Check className="h-4 w-4 shrink-0" />
          <span>{successMsg}</span>
        </div>
      )}

      {errorMsg && (
        <div className="p-3 rounded-lg bg-rose-950/40 border border-rose-500/30 text-rose-400 text-xs flex items-center gap-2">
          <AlertCircle className="h-4 w-4 shrink-0" />
          <span>{errorMsg}</span>
        </div>
      )}

      {/* Daftar Webhook */}
      {webhooks.length === 0 ? (
        <div className={`${AppTheme.containers.card} text-center py-12 space-y-3`}>
          <div className="mx-auto w-10 h-10 rounded-full bg-zinc-800 flex items-center justify-center text-zinc-500">
            <Send className="h-5 w-5" />
          </div>
          <p className="text-xs text-zinc-400">Belum ada webhook yang didaftarkan.</p>
        </div>
      ) : (
        <div className={`${AppTheme.containers.card} overflow-hidden p-0`}>
          <div className="divide-y divide-[#222222]">
            {webhooks.map((wh) => (
              <div key={wh.id} className="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4 hover:bg-[#161616]/50 transition-colors">
                <div className="space-y-1.5">
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-semibold text-zinc-200">{wh.name}</span>
                    <span className="text-[11px] font-mono text-zinc-400 truncate max-w-xs">{wh.url}</span>
                    {wh.last_status && (
                      <span
                        className={`px-2 py-0.2 rounded text-[10px] font-mono font-bold ${
                          wh.last_status >= 200 && wh.last_status < 300
                            ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                            : "bg-rose-500/10 text-rose-400 border border-rose-500/20"
                        }`}
                      >
                        HTTP {wh.last_status}
                      </span>
                    )}
                  </div>
                  <div className="flex flex-wrap items-center gap-1.5 pt-0.5">
                    {wh.events.map((ev) => (
                      <span key={ev} className="px-2 py-0.5 rounded text-[10px] bg-zinc-800 text-zinc-300 border border-zinc-700">
                        {ev}
                      </span>
                    ))}
                  </div>
                </div>

                <div className="flex items-center gap-2 shrink-0">
                  <button
                    type="button"
                    disabled={testingId === wh.id}
                    onClick={() => handleTestWebhook(wh.id)}
                    className="px-2.5 py-1.5 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-xs font-medium border border-zinc-700 flex items-center gap-1.5 transition-colors cursor-pointer disabled:opacity-50"
                    title="Kirim Test Ping"
                  >
                    {testingId === wh.id ? <RefreshCw className="h-3.5 w-3.5 animate-spin" /> : <Play className="h-3.5 w-3.5 text-emerald-400" />}
                    <span>Test Ping</span>
                  </button>

                  <button
                    type="button"
                    onClick={() => handleDeleteWebhook(wh.id, wh.name)}
                    className="p-1.5 rounded-lg text-zinc-400 hover:text-rose-400 hover:bg-rose-500/10 transition-colors cursor-pointer"
                    title="Hapus Webhook"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Modal Tambah Webhook */}
      {isCreateModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-xs p-4">
          <div className="bg-[#141414] border border-[#2e2e2e] rounded-xl w-full max-w-md p-6 shadow-2xl space-y-4">
            <div className="flex items-center justify-between border-b border-[#262626] pb-3">
              <h3 className="text-sm font-semibold text-zinc-100 flex items-center gap-2">
                <Send className="h-4 w-4 text-emerald-400" />
                Daftarkan Webhook Baru
              </h3>
              <button
                type="button"
                onClick={() => setIsCreateModalOpen(false)}
                className="text-zinc-400 hover:text-zinc-200 p-1 rounded-lg transition-colors cursor-pointer"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            {modalError && (
              <div className="p-3 rounded-lg bg-rose-950/40 border border-rose-500/30 text-rose-400 text-xs flex items-center gap-2">
                <AlertCircle className="h-4 w-4 shrink-0" />
                <span>{modalError}</span>
              </div>
            )}

            <form onSubmit={handleCreateWebhook} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-zinc-300 mb-1.5">Nama Webhook</label>
                <input
                  type="text"
                  required
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Contoh: Discord Alerts, Slack Dev Channel"
                  className="w-full bg-[#181818] border border-[#2e2e2e] text-zinc-200 text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-emerald-500 transition-colors"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-zinc-300 mb-1.5">Payload URL Target (HTTPS)</label>
                <input
                  type="url"
                  required
                  value={url}
                  onChange={(e) => setUrl(e.target.value)}
                  placeholder="https://discord.com/api/webhooks/..."
                  className="w-full bg-[#181818] border border-[#2e2e2e] text-zinc-200 text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-emerald-500 transition-colors"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-zinc-300 mb-1.5">Secret Key HMAC (Opsional)</label>
                <input
                  type="password"
                  value={secret}
                  onChange={(e) => setSecret(e.target.value)}
                  placeholder="Kunci rahasia untuk memverifikasi signature X-Caelus-Signature"
                  className="w-full bg-[#181818] border border-[#2e2e2e] text-zinc-200 text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-emerald-500 transition-colors"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-zinc-300 mb-2">Event Triggers</label>
                <div className="space-y-2">
                  {AVAILABLE_EVENTS.map((ev) => (
                    <label key={ev.id} className="flex items-center gap-2.5 text-xs text-zinc-300 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={selectedEvents.includes(ev.id)}
                        onChange={() => handleToggleEvent(ev.id)}
                        className="rounded border-[#2e2e2e] bg-[#181818] text-emerald-500 focus:ring-0 cursor-pointer"
                      />
                      <span>{ev.label}</span>
                      <span className="text-[10px] font-mono text-zinc-400">({ev.id})</span>
                    </label>
                  ))}
                </div>
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setIsCreateModalOpen(false)}
                  className="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-xs rounded-lg transition-colors cursor-pointer"
                >
                  Batal
                </button>
                <button
                  type="submit"
                  disabled={isCreating}
                  className="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-zinc-950 font-semibold text-xs rounded-lg transition-colors flex items-center gap-2 cursor-pointer disabled:opacity-50"
                >
                  {isCreating && <RefreshCw className="h-3.5 w-3.5 animate-spin" />}
                  <span>Daftarkan Webhook</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
