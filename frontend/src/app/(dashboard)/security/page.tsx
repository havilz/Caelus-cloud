"use client";

import React, { useState, useEffect, useCallback } from "react";
import { AppTheme } from "@/core/theme";
import { securityService } from "@/services/security.service";
import { serverService } from "@/services/server.service";
import {
  SecurityPostureOverview,
  SecurityFinding,
  SecurityScan,
  ScanType,
  FindingCategory,
  FindingSeverity,
  FindingStatus,
} from "@/types/security";
import { Server } from "@/types/server";
import { SecurityScoreBadge } from "@/components/security/SecurityScoreBadge";
import { FindingDetailModal } from "@/components/security/FindingDetailModal";
import {
  ShieldCheck,
  Search,
  RotateCw,
  Play,
  Filter,
  CheckCircle2,
  Clock,
  ExternalLink,
  ChevronRight,
} from "lucide-react";

export default function SecurityPage() {
  const [posture, setPosture] = useState<SecurityPostureOverview | null>(null);
  const [findings, setFindings] = useState<SecurityFinding[]>([]);
  const [scans, setScans] = useState<SecurityScan[]>([]);
  const [servers, setServers] = useState<Server[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isScanning, setIsScanning] = useState(false);

  // Tab State
  const [activeTab, setActiveTab] = useState<"findings" | "scans">("findings");

  // Scan Launcher State
  const [selectedServerId, setSelectedServerId] = useState<string>("all");
  const [selectedScanType, setSelectedScanType] = useState<ScanType>("full");

  // Filters
  const [searchQuery, setSearchQuery] = useState("");
  const [categoryFilter, setCategoryFilter] = useState<FindingCategory | "all">("all");
  const [severityFilter, setSeverityFilter] = useState<FindingSeverity | "all">("all");
  const [statusFilter, setStatusFilter] = useState<FindingStatus | "all">("open");

  // Selected Finding for Modal
  const [selectedFinding, setSelectedFinding] = useState<SecurityFinding | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const loadData = useCallback(async () => {
    try {
      setIsLoading(true);
      const [postureRes, findingsRes, scansRes, serversRes] = await Promise.all([
        securityService.getPostureOverview().catch(() => null),
        securityService
          .listFindings({
            category: categoryFilter !== "all" ? categoryFilter : undefined,
            severity: severityFilter !== "all" ? severityFilter : undefined,
            status: statusFilter !== "all" ? statusFilter : undefined,
            limit: 50,
          })
          .catch(() => ({ data: [], meta: { page: 1, limit: 50, total_items: 0, total_pages: 1 } })),
        securityService.listScans(1, 20).catch(() => ({ data: [], meta: { page: 1, limit: 20, total_items: 0, total_pages: 1 } })),
        serverService.listServers(1, 100).catch(() => ({ data: [] })),
      ]);

      if (postureRes) setPosture(postureRes);
      setFindings(findingsRes.data || []);
      setScans(scansRes.data || []);
      setServers(serversRes.data || []);
    } catch (err) {
      console.error("Gagal memuat data keamanan Sentinel:", err);
    } finally {
      setIsLoading(false);
    }
  }, [categoryFilter, severityFilter, statusFilter]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleTriggerScan = async () => {
    try {
      setIsScanning(true);
      await securityService.triggerScan({
        server_id: selectedServerId !== "all" ? selectedServerId : undefined,
        scan_type: selectedScanType,
      });
      // Refresh setelah jeda
      setTimeout(() => {
        loadData();
        setIsScanning(false);
      }, 1500);
    } catch (err) {
      console.error("Gagal memicu scan:", err);
      setIsScanning(false);
    }
  };

  const handleUpdateFindingStatus = async (findingId: string, status: FindingStatus) => {
    await securityService.updateFindingStatus(findingId, status);
    loadData();
  };

  const filteredFindings = findings.filter((f) =>
    f.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
    f.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
    (f.server_name && f.server_name.toLowerCase().includes(searchQuery.toLowerCase()))
  );

  const getSeverityBadge = (severity: FindingSeverity) => {
    switch (severity) {
      case "critical":
        return "bg-rose-950/80 text-rose-400 border-rose-800/60";
      case "high":
        return "bg-orange-950/80 text-orange-400 border-orange-800/60";
      case "medium":
        return "bg-amber-950/80 text-amber-400 border-amber-800/60";
      case "low":
        return "bg-emerald-950/80 text-emerald-400 border-emerald-800/60";
      default:
        return "bg-zinc-900 text-zinc-400 border-zinc-800";
    }
  };

  return (
    <div className={AppTheme.containers.pageWrapper}>
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className={AppTheme.text.h1}>Sentinel Security Hub</h1>
          <p className={AppTheme.text.subtitle}>
            Pusat audit keamanan infrastruktur, pemindaian kerentanan modular, dan mitigasi otomatis Caelus.
          </p>
        </div>
        <button
          onClick={() => loadData()}
          disabled={isLoading}
          className={AppTheme.controls.buttonSecondary}
        >
          <RotateCw className={`w-3.5 h-3.5 ${isLoading ? "animate-spin" : ""}`} />
          Segarkan Data
        </button>
      </div>

      {/* Posture Score Badge */}
      {posture && (
        <SecurityScoreBadge
          score={posture.overall_score}
          grade={posture.grade}
          criticalCount={posture.critical_count}
          highCount={posture.high_count}
          openFindings={posture.open_findings}
        />
      )}

      {/* Quick Scan Launcher Card */}
      <div className={AppTheme.containers.card}>
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div className="space-y-1">
            <h2 className="text-sm font-semibold text-zinc-200 flex items-center gap-2">
              <ShieldCheck className="w-4 h-4 text-emerald-400" />
              Jalankan Pemindaian Keamanan Baru
            </h2>
            <p className="text-xs text-zinc-400">
              Pilih target server dan modul scanner untuk memulai audit kepatuhan dan analisis celah risiko.
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-3">
            {/* Target Server Select */}
            <select
              value={selectedServerId}
              onChange={(e) => setSelectedServerId(e.target.value)}
              className={AppTheme.controls.selectSm}
            >
              <option value="all">Semua Server Organisasi</option>
              {servers.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name} ({s.ip_address || "Custom Node"})
                </option>
              ))}
            </select>

            {/* Scan Type Select */}
            <select
              value={selectedScanType}
              onChange={(e) => setSelectedScanType(e.target.value as ScanType)}
              className={AppTheme.controls.selectSm}
            >
              <option value="full">Audit Lengkap (Full Suite)</option>
              <option value="port">Port & Service Exposure</option>
              <option value="tls">TLS/SSL & Cipher Suites</option>
              <option value="headers">HTTP Security Headers</option>
              <option value="host_config">Host & Container Hardening</option>
              <option value="vuln">System CVE Vulnerabilities</option>
            </select>

            {/* Trigger Button */}
            <button
              onClick={handleTriggerScan}
              disabled={isScanning}
              className={AppTheme.controls.buttonPrimary}
            >
              <Play className={`w-3.5 h-3.5 ${isScanning ? "animate-spin" : ""}`} />
              {isScanning ? "Memindai..." : "Mulai Pemindaian"}
            </button>
          </div>
        </div>
      </div>

      {/* Tab Navigation */}
      <div className="flex items-center gap-2 border-b border-[#262626] pb-3">
        <button
          onClick={() => setActiveTab("findings")}
          className={`px-4 py-2 rounded-lg text-xs font-semibold transition-colors flex items-center gap-2 ${
            activeTab === "findings"
              ? "bg-[#1f1f1f] text-emerald-400 border border-[#333333]"
              : "text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Filter className="w-3.5 h-3.5" />
          Temuan Keamanan ({findings.length})
        </button>
        <button
          onClick={() => setActiveTab("scans")}
          className={`px-4 py-2 rounded-lg text-xs font-semibold transition-colors flex items-center gap-2 ${
            activeTab === "scans"
              ? "bg-[#1f1f1f] text-emerald-400 border border-[#333333]"
              : "text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Clock className="w-3.5 h-3.5" />
          Riwayat Pemindaian ({scans.length})
        </button>
      </div>

      {/* Tab Content: Findings */}
      {activeTab === "findings" && (
        <div className="space-y-4">
          {/* Filters Bar */}
          <div className="flex flex-col sm:flex-row items-center justify-between gap-3 bg-[#111111] p-3 rounded-xl border border-[#222222]">
            <div className="relative w-full sm:w-72">
              <Search className="w-4 h-4 absolute left-3 top-2.5 text-zinc-500" />
              <input
                type="text"
                placeholder="Cari temuan, CVE, judul..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-9 pr-3 py-1.5 rounded-lg bg-[#181818] border border-[#2e2e2e] text-xs text-zinc-200 placeholder-zinc-500 focus:outline-none focus:border-emerald-500"
              />
            </div>

            <div className="flex items-center gap-2 w-full sm:w-auto overflow-x-auto">
              <select
                value={severityFilter}
                onChange={(e) => setSeverityFilter(e.target.value as FindingSeverity | "all")}
                className={AppTheme.controls.selectSm}
              >
                <option value="all">Semua Severity</option>
                <option value="critical">Critical</option>
                <option value="high">High</option>
                <option value="medium">Medium</option>
                <option value="low">Low</option>
              </select>

              <select
                value={categoryFilter}
                onChange={(e) => setCategoryFilter(e.target.value as FindingCategory | "all")}
                className={AppTheme.controls.selectSm}
              >
                <option value="all">Semua Kategori</option>
                <option value="network">Network & Ports</option>
                <option value="tls">TLS / SSL</option>
                <option value="http_headers">HTTP Headers</option>
                <option value="host_config">Host Hardening</option>
                <option value="vulnerability">Vulnerabilities</option>
              </select>

              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value as FindingStatus | "all")}
                className={AppTheme.controls.selectSm}
              >
                <option value="all">Semua Status</option>
                <option value="open">Open</option>
                <option value="acknowledged">Acknowledged</option>
                <option value="resolved">Resolved</option>
                <option value="false_positive">False Positive</option>
              </select>
            </div>
          </div>

          {/* Findings Table */}
          <div className={AppTheme.containers.card}>
            {filteredFindings.length === 0 ? (
              <div className="text-center py-12 space-y-3">
                <CheckCircle2 className="w-10 h-10 text-emerald-400 mx-auto opacity-70" />
                <h3 className="text-sm font-semibold text-zinc-300">
                  Tidak Ada Temuan Keamanan yang Cocok
                </h3>
                <p className="text-xs text-zinc-500 max-w-sm mx-auto">
                  Infrastruktur Anda berada dalam postur aman atau filter yang Anda gunakan tidak menemukan data aktif.
                </p>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs">
                  <thead className="text-[11px] uppercase tracking-wider text-zinc-500 border-b border-[#222222]">
                    <tr>
                      <th className="pb-3 font-semibold">Tingkat Risiko</th>
                      <th className="pb-3 font-semibold">Judul Temuan</th>
                      <th className="pb-3 font-semibold">Kategori</th>
                      <th className="pb-3 font-semibold">Server Target</th>
                      <th className="pb-3 font-semibold">Status</th>
                      <th className="pb-3 font-semibold text-right">Aksi</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-[#1c1c1c]">
                    {filteredFindings.map((f) => (
                      <tr
                        key={f.id}
                        onClick={() => {
                          setSelectedFinding(f);
                          setIsModalOpen(true);
                        }}
                        className="hover:bg-[#161616] cursor-pointer transition-colors group"
                      >
                        <td className="py-3.5">
                          <span
                            className={`px-2 py-0.5 rounded-md text-[10px] font-bold uppercase tracking-wider border ${getSeverityBadge(
                              f.severity
                            )}`}
                          >
                            {f.severity}
                          </span>
                        </td>
                        <td className="py-3.5 font-medium text-zinc-200 group-hover:text-emerald-400 transition-colors">
                          {f.title}
                        </td>
                        <td className="py-3.5 text-zinc-400">
                          <span className={AppTheme.controls.badgeMono}>
                            {f.category.replace("_", " ")}
                          </span>
                        </td>
                        <td className="py-3.5 text-zinc-400 font-mono">
                          {f.server_name || "Semua Server"}
                        </td>
                        <td className="py-3.5">
                          <span
                            className={`px-2 py-0.5 rounded-md text-[10px] uppercase font-semibold ${
                              f.status === "resolved"
                                ? "bg-emerald-950/60 text-emerald-400 border border-emerald-800/40"
                                : f.status === "acknowledged"
                                ? "bg-amber-950/60 text-amber-400 border border-amber-800/40"
                                : "bg-zinc-900 text-zinc-400 border border-zinc-800"
                            }`}
                          >
                            {f.status}
                          </span>
                        </td>
                        <td className="py-3.5 text-right">
                          <button className="text-zinc-500 group-hover:text-zinc-300 p-1">
                            <ChevronRight className="w-4 h-4" />
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Tab Content: Scans History */}
      {activeTab === "scans" && (
        <div className={AppTheme.containers.card}>
          {scans.length === 0 ? (
            <div className="text-center py-12 space-y-3">
              <Clock className="w-10 h-10 text-zinc-600 mx-auto" />
              <h3 className="text-sm font-semibold text-zinc-300">Belum Ada Riwayat Pemindaian</h3>
              <p className="text-xs text-zinc-500">
                Jalankan pemindaian pertama Anda menggunakan bilah di atas untuk memeriksa postur keamanan.
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead className="text-[11px] uppercase tracking-wider text-zinc-500 border-b border-[#222222]">
                  <tr>
                    <th className="pb-3 font-semibold">Tipe Scan</th>
                    <th className="pb-3 font-semibold">Status</th>
                    <th className="pb-3 font-semibold">Skor</th>
                    <th className="pb-3 font-semibold">Total Temuan</th>
                    <th className="pb-3 font-semibold">Critical / High</th>
                    <th className="pb-3 font-semibold">Waktu Mulai</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-[#1c1c1c]">
                  {scans.map((s) => (
                    <tr key={s.id} className="hover:bg-[#161616] transition-colors">
                      <td className="py-3.5 font-medium text-zinc-200 flex items-center gap-2">
                        <span className={AppTheme.controls.badgeMono}>
                          {s.scan_type.toUpperCase()}
                        </span>
                        <span className="text-zinc-400 font-mono text-[11px]">
                          {s.server_name || "Semua Server"}
                        </span>
                      </td>
                      <td className="py-3.5">
                        <span
                          className={`px-2 py-0.5 rounded-md text-[10px] uppercase font-semibold ${
                            s.status === "completed"
                              ? "bg-emerald-950/60 text-emerald-400 border border-emerald-800/40"
                              : s.status === "running"
                              ? "bg-cyan-950/60 text-cyan-400 border border-cyan-800/40 animate-pulse"
                              : "bg-rose-950/60 text-rose-400 border border-rose-800/40"
                          }`}
                        >
                          {s.status}
                        </span>
                      </td>
                      <td className="py-3.5 font-bold text-zinc-200">
                        {s.score} / 100
                      </td>
                      <td className="py-3.5 text-zinc-300">{s.total_findings} temuan</td>
                      <td className="py-3.5 space-x-2">
                        <span className="text-rose-400 font-bold">{s.critical_count} crit</span>
                        <span className="text-zinc-600">/</span>
                        <span className="text-orange-400 font-bold">{s.high_count} high</span>
                      </td>
                      <td className="py-3.5 text-zinc-400 font-mono text-[11px]">
                        {s.started_at ? new Date(s.started_at).toLocaleString("id-ID") : "-"}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* Finding Detail Modal */}
      <FindingDetailModal
        finding={selectedFinding}
        isOpen={isModalOpen}
        onClose={() => {
          setIsModalOpen(false);
          setSelectedFinding(null);
        }}
        onUpdateStatus={handleUpdateFindingStatus}
      />
    </div>
  );
}
