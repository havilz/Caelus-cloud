"use client";

import React, { useState, useEffect } from "react";
import { History, Search, RefreshCw, AlertCircle, Eye, Clock, User, Globe, ChevronLeft, ChevronRight, X } from "lucide-react";
import { AppTheme } from "@/core/theme";
import { settingsService } from "@/services/settings.service";
import { AuditLog } from "@/types/settings";

export const AuditLogsTab: React.FC = () => {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [searchFilter, setSearchFilter] = useState("");

  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null);

  const fetchLogs = async (currentPage = 1) => {
    try {
      setIsLoading(true);
      const res = await settingsService.listAuditLogs(currentPage, 20);
      setLogs(res.data || []);
      setTotal(res.total);
      setPage(res.page);
    } catch (err: any) {
      setErrorMsg("Gagal memuat log audit aktivitas");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchLogs(page);
  }, [page]);

  const filteredLogs = logs.filter((log) => {
    if (!searchFilter.trim()) return true;
    const term = searchFilter.toLowerCase();
    return (
      log.action?.toLowerCase().includes(term) ||
      log.resource_type?.toLowerCase().includes(term) ||
      log.ip_address?.toLowerCase().includes(term)
    );
  });

  const totalPages = Math.ceil(total / 20) || 1;

  const getActionBadge = (action: string) => {
    const act = action.toLowerCase();
    if (act.includes("delete") || act.includes("remove") || act.includes("destroy")) {
      return <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-rose-500/10 text-rose-400 border border-rose-500/20">{action}</span>;
    }
    if (act.includes("create") || act.includes("deploy") || act.includes("add")) {
      return <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">{action}</span>;
    }
    if (act.includes("update") || act.includes("edit") || act.includes("change")) {
      return <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/20">{action}</span>;
    }
    return <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-sky-500/10 text-sky-400 border border-sky-500/20">{action}</span>;
  };

  return (
    <div className="space-y-6">
      {}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h3 className="text-sm font-semibold text-zinc-100">Log Audit Aktivitas Global</h3>
          <p className="text-xs text-zinc-400 mt-0.5">
            Rekam jejak tidak dapat diubah dari setiap tindakan administrasi, modifikasi resource, dan sesi pengguna
          </p>
        </div>

        <div className="flex items-center gap-2">
          <div className="relative">
            <Search className="h-3.5 w-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" />
            <input
              type="text"
              value={searchFilter}
              onChange={(e) => setSearchFilter(e.target.value)}
              placeholder="Cari aksi / resource..."
              className="bg-[#141414] border border-[#2e2e2e] text-zinc-200 text-xs rounded-lg pl-8 pr-3 py-1.5 focus:outline-none focus:border-emerald-500 transition-colors w-48"
            />
          </div>
          <button
            type="button"
            onClick={() => fetchLogs(page)}
            className="p-1.5 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 rounded-lg border border-zinc-700 transition-colors cursor-pointer"
            title="Refresh Log"
          >
            <RefreshCw className={`h-4 w-4 ${isLoading ? "animate-spin" : ""}`} />
          </button>
        </div>
      </div>

      {errorMsg && (
        <div className="p-3 rounded-lg bg-rose-950/40 border border-rose-500/30 text-rose-400 text-xs flex items-center gap-2">
          <AlertCircle className="h-4 w-4 shrink-0" />
          <span>{errorMsg}</span>
        </div>
      )}

      {}
      {isLoading ? (
        <div className="flex items-center justify-center py-16 text-zinc-500">
          <RefreshCw className="h-5 w-5 animate-spin mr-2" />
          <span className="text-sm">Memuat log audit...</span>
        </div>
      ) : filteredLogs.length === 0 ? (
        <div className={`${AppTheme.containers.card} text-center py-12 space-y-3`}>
          <div className="mx-auto w-10 h-10 rounded-full bg-zinc-800 flex items-center justify-center text-zinc-500">
            <History className="h-5 w-5" />
          </div>
          <p className="text-xs text-zinc-400">Belum ada riwayat aktivitas yang tercatat.</p>
        </div>
      ) : (
        <div className={`${AppTheme.containers.card} overflow-hidden p-0`}>
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs text-zinc-300">
              <thead className="bg-[#161616] text-[11px] font-semibold text-zinc-400 uppercase tracking-wider border-b border-[#262626]">
                <tr>
                  <th className="py-3 px-4">Aksi</th>
                  <th className="py-3 px-4">Tipe Resource</th>
                  <th className="py-3 px-4">Alamat IP</th>
                  <th className="py-3 px-4">Waktu</th>
                  <th className="py-3 px-4 text-right">Detail</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#222222]">
                {filteredLogs.map((log) => (
                  <tr key={log.id} className="hover:bg-[#161616]/50 transition-colors">
                    <td className="py-3 px-4">{getActionBadge(log.action)}</td>
                    <td className="py-3 px-4 font-mono text-[11px] text-zinc-300">{log.resource_type}</td>
                    <td className="py-3 px-4 font-mono text-[11px] text-zinc-400">{log.ip_address || "-"}</td>
                    <td className="py-3 px-4 text-[11px] text-zinc-400">{new Date(log.created_at).toLocaleString()}</td>
                    <td className="py-3 px-4 text-right">
                      <button
                        type="button"
                        onClick={() => setSelectedLog(log)}
                        className="p-1 text-zinc-400 hover:text-emerald-400 hover:bg-emerald-500/10 rounded transition-colors cursor-pointer"
                        title="Lihat Payload"
                      >
                        <Eye className="h-4 w-4" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {}
          <div className="p-3 border-t border-[#262626] bg-[#161616] flex items-center justify-between text-xs text-zinc-400">
            <span>
              Halaman {page} dari {totalPages} ({total} total aksi)
            </span>
            <div className="flex items-center gap-1.5">
              <button
                type="button"
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                className="p-1 rounded bg-[#1f1f1f] border border-[#2e2e2e] text-zinc-300 hover:bg-[#2a2a2a] disabled:opacity-40 disabled:cursor-not-allowed cursor-pointer"
              >
                <ChevronLeft className="h-4 w-4" />
              </button>
              <button
                type="button"
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
                className="p-1 rounded bg-[#1f1f1f] border border-[#2e2e2e] text-zinc-300 hover:bg-[#2a2a2a] disabled:opacity-40 disabled:cursor-not-allowed cursor-pointer"
              >
                <ChevronRight className="h-4 w-4" />
              </button>
            </div>
          </div>
        </div>
      )}

      {}
      {selectedLog && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-xs p-4">
          <div className="bg-[#141414] border border-[#2e2e2e] rounded-xl w-full max-w-lg p-6 shadow-2xl space-y-4">
            <div className="flex items-center justify-between border-b border-[#262626] pb-3">
              <h3 className="text-sm font-semibold text-zinc-100 flex items-center gap-2">
                <History className="h-4 w-4 text-emerald-400" />
                Detail Log Audit
              </h3>
              <button
                type="button"
                onClick={() => setSelectedLog(null)}
                className="text-zinc-400 hover:text-zinc-200 p-1 rounded-lg transition-colors cursor-pointer"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <div className="space-y-3 text-xs">
              <div className="grid grid-cols-2 gap-2 bg-[#181818] p-3 rounded-lg border border-[#262626]">
                <div>
                  <p className="text-[10px] uppercase text-zinc-500 font-semibold">Aksi</p>
                  <p className="font-semibold text-zinc-200">{selectedLog.action}</p>
                </div>
                <div>
                  <p className="text-[10px] uppercase text-zinc-500 font-semibold">Tipe Resource</p>
                  <p className="font-mono text-zinc-300">{selectedLog.resource_type}</p>
                </div>
                <div>
                  <p className="text-[10px] uppercase text-zinc-500 font-semibold">Alamat IP</p>
                  <p className="font-mono text-zinc-300">{selectedLog.ip_address || "-"}</p>
                </div>
                <div>
                  <p className="text-[10px] uppercase text-zinc-500 font-semibold">Timestamp</p>
                  <p className="text-zinc-300">{new Date(selectedLog.created_at).toLocaleString()}</p>
                </div>
              </div>

              {selectedLog.user_agent && (
                <div>
                  <p className="text-[11px] text-zinc-400 font-medium mb-1">User Agent:</p>
                  <p className="p-2 rounded bg-[#181818] border border-[#262626] font-mono text-[10px] text-zinc-400 break-all">
                    {selectedLog.user_agent}
                  </p>
                </div>
              )}

              <div>
                <p className="text-[11px] text-zinc-400 font-medium mb-1">Payload JSON:</p>
                <pre className="p-3 rounded-lg bg-[#0e0e0e] border border-[#262626] font-mono text-[11px] text-emerald-400/90 overflow-x-auto max-h-48">
                  {JSON.stringify(selectedLog.payload || {}, null, 2)}
                </pre>
              </div>
            </div>

            <div className="flex justify-end pt-2">
              <button
                type="button"
                onClick={() => setSelectedLog(null)}
                className="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-xs rounded-lg transition-colors cursor-pointer"
              >
                Tutup
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
