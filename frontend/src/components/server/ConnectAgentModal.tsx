"use client";

import React, { useState } from "react";
import { Dialog } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Copy, Check, Terminal, ShieldCheck, RefreshCw, CheckCircle2 } from "lucide-react";
import { AppText, AppTheme } from "@/core/theme";
import { Server } from "@/types/server";

interface ConnectAgentModalProps {
  readonly isOpen: boolean;
  readonly onClose: () => void;
  readonly server: Server | null;
}

export const ConnectAgentModal: React.FC<ConnectAgentModalProps> = ({
  isOpen,
  onClose,
  server,
}) => {
  const [copiedType, setCopiedType] = useState<string | null>(null);

  if (!server) return null;

  const apiEndpoint = typeof window !== "undefined" ? `${window.location.protocol}//${window.location.hostname}:8080` : "http://localhost:8080";
  const agentSecret = "caelus_agent_sec_" + server.id.replace(/-/g, "").substring(0, 16);

  const oneLineCommand = `curl -sSL ${apiEndpoint}/install.sh | sudo bash -s -- --server-id="${server.id}" --secret="${agentSecret}" --api="${apiEndpoint}"`;
  const manualGoCommand = `export SERVER_ID="${server.id}" AGENT_SECRET="${agentSecret}" API_ENDPOINT="${apiEndpoint}" && caelus-agent`;

  const handleCopy = (text: string, type: string) => {
    navigator.clipboard.writeText(text);
    setCopiedType(type);
    setTimeout(() => setCopiedType(null), 2000);
  };

  return (
    <Dialog
      isOpen={isOpen}
      onClose={onClose}
      title="Hubungkan Caelus Agent ke VPS Anda"
      description={`Jalankan perintah berikut pada terminal VPS "${server.name}" untuk memulai streaming telemetri & otomasi`}
      maxWidth="lg"
    >
      <div className="space-y-4 py-1">
        {/* Server Identity Header */}
        <div className="p-3.5 rounded-xl bg-[#141414] border border-[#262626] flex items-center justify-between">
          <div>
            <span className="text-[10px] text-[#707070] font-semibold uppercase tracking-wider">Server Target</span>
            <h4 className="text-sm font-bold text-[#ededed]">{server.name}</h4>
            <p className="text-xs font-mono text-[#a1a1a1]">IP: {server.ip_address || "BYOS / Host Server"}</p>
          </div>
          <div className="text-right">
            <span className="text-[10px] text-[#707070] font-semibold uppercase tracking-wider">Status Agent</span>
            <div className="flex items-center gap-1.5 mt-0.5 justify-end">
              {server.status === "running" ? (
                <span className="px-2 py-0.5 rounded-md text-[10px] font-semibold bg-emerald-950/60 text-emerald-400 border border-emerald-800/40 flex items-center gap-1">
                  <CheckCircle2 className="w-3 h-3" />
                  Terhubung & Aktif
                </span>
              ) : (
                <span className="px-2 py-0.5 rounded-md text-[10px] font-semibold bg-amber-950/60 text-amber-400 border border-amber-800/40 flex items-center gap-1">
                  <RefreshCw className="w-3 h-3 animate-spin" />
                  Menunggu Koneksi...
                </span>
              )}
            </div>
          </div>
        </div>

        {/* Option 1: 1-Line Script (Recommended for Ubuntu/Linux) */}
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Terminal className="w-4 h-4 text-emerald-400" />
              <span className="text-xs font-semibold text-[#ededed]">1. Perintah 1-Line Auto Installer (Rekomendasi Linux / Ubuntu VPS)</span>
            </div>
            <span className="text-[10px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 px-2 py-0.5 rounded">
              Systemd Auto-Start
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
                  <span className="text-[11px]">Salin</span>
                </>
              )}
            </button>
          </div>
          <p className="text-[11px] text-[#707070]">
            Script ini akan mengunduh binary daemon, membuat service systemd di <code>/etc/systemd/system/caelus-agent.service</code>, dan langsung mengaktifkannya.
          </p>
        </div>

        {/* Option 2: Manual Go / Binary Run */}
        <div className="space-y-2 pt-2 border-t border-[#202020]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <ShieldCheck className="w-4 h-4 text-cyan-400" />
              <span className="text-xs font-semibold text-[#ededed]">2. Jalankan Manual via Binary / Terminal Lokal</span>
            </div>
          </div>

          <div className="relative group">
            <pre className="p-3 rounded-xl bg-[#0d0d0d] border border-[#222222] text-xs font-mono text-[#a1a1a1] overflow-x-auto whitespace-pre-wrap break-all select-all">
              {manualGoCommand}
            </pre>
            <button
              onClick={() => handleCopy(manualGoCommand, "manual")}
              className="absolute top-2.5 right-2.5 px-2.5 py-1 rounded-lg bg-[#1f1f1f] hover:bg-[#2a2a2a] text-[#ededed] border border-[#333333] text-xs flex items-center gap-1.5 transition-colors cursor-pointer shadow-md"
            >
              {copiedType === "manual" ? (
                <>
                  <Check className="w-3.5 h-3.5 text-emerald-400" />
                  <span className="text-emerald-400 text-[11px]">Tersalin!</span>
                </>
              ) : (
                <>
                  <Copy className="w-3.5 h-3.5" />
                  <span className="text-[11px]">Salin</span>
                </>
              )}
            </button>
          </div>
        </div>

        {/* Credentials Details Box */}
        <div className="grid grid-cols-2 gap-2.5 p-3 rounded-xl bg-[#121212] border border-[#262626] text-[11px]">
          <div>
            <span className="text-[#707070] uppercase tracking-wider block mb-0.5">Server UUID</span>
            <span className="font-mono text-emerald-400 select-all">{server.id}</span>
          </div>
          <div>
            <span className="text-[#707070] uppercase tracking-wider block mb-0.5">Ingestion Secret</span>
            <span className="font-mono text-[#a1a1a1] select-all">{agentSecret}</span>
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 pt-3 border-t border-[#262626]">
          <Button variant="outline" type="button" onClick={onClose}>
            Tutup
          </Button>
        </div>
      </div>
    </Dialog>
  );
};
