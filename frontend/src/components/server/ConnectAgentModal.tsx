"use client";

import React, { useState } from "react";
import { Dialog } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Copy, Check, Terminal, RefreshCw, CheckCircle2 } from "lucide-react";
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

  const handleCopy = (text: string, type: string) => {
    navigator.clipboard.writeText(text);
    setCopiedType(type);
    setTimeout(() => setCopiedType(null), 2000);
  };

  return (
    <Dialog
      isOpen={isOpen}
      onClose={onClose}
      title="Connect Caelus Agent to Your VPS"
      description={`Run the following command in VPS terminal "${server.name}" to start telemetry streaming & automation`}
      maxWidth="lg"
    >
      <div className="space-y-4 py-1">
        <div className="p-3.5 rounded-xl bg-[#141414] border border-[#262626] flex items-center justify-between">
          <div>
            <span className="text-[10px] text-[#707070] font-semibold uppercase tracking-wider">Target Server</span>
            <h4 className="text-sm font-bold text-[#ededed]">{server.name}</h4>
            <p className="text-xs font-mono text-[#a1a1a1]">IP: {server.ip_address || "BYOS / Host Server"}</p>
          </div>
          <div className="text-right">
            <span className="text-[10px] text-[#707070] font-semibold uppercase tracking-wider">Agent Status</span>
            <div className="flex items-center gap-1.5 mt-0.5 justify-end">
              {server.status === "running" ? (
                <span className="px-2 py-0.5 rounded-md text-[10px] font-semibold bg-emerald-950/60 text-emerald-400 border border-emerald-800/40 flex items-center gap-1">
                  <CheckCircle2 className="w-3 h-3" />
                  Connected & Active
                </span>
              ) : (
                <span className="px-2 py-0.5 rounded-md text-[10px] font-semibold bg-amber-950/60 text-amber-400 border border-amber-800/40 flex items-center gap-1">
                  <RefreshCw className="w-3 h-3 animate-spin" />
                  Waiting for Connection...
                </span>
              )}
            </div>
          </div>
        </div>

        <div className="flex border-b border-[#262626] gap-2 pt-1">
          <button
            type="button"
            onClick={() => setCopiedType(null)}
            className="pb-2 text-xs font-semibold border-b-2 border-emerald-500 text-emerald-400 flex items-center gap-1.5 cursor-pointer"
          >
            <Terminal className="w-3.5 h-3.5" />
            Method 1: Automatic (VPS Recommended)
          </button>
        </div>

        <div className="space-y-3 pt-1">
          <div className="flex items-center justify-between">
            <span className="text-xs text-[#a1a1a1]">
              Run the following one-line command in your VPS terminal to install and start the agent automatically via <code>systemd</code>:
            </span>
            <span className="text-[10px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 px-2 py-0.5 rounded whitespace-nowrap">
              Auto-Start Service
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
                  <span className="text-emerald-400 text-[11px]">Copied!</span>
                </>
              ) : (
                <>
                  <Copy className="w-3.5 h-3.5" />
                  <span className="text-[11px]">Copy Command</span>
                </>
              )}
            </button>
          </div>

          <div className="p-2.5 rounded-lg bg-[#141414] border border-[#222222] text-[11px] text-[#888888] space-y-1">
            <p className="font-semibold text-[#cccccc]">Note:</p>
            <p>• The command above starts the agent in the background as a system service immediately (no additional commands required).</p>
            <p>• To check status on your VPS: <code>sudo systemctl status caelus-agent</code></p>
          </div>
        </div>

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

        <div className="flex items-center justify-end gap-3 pt-3 border-t border-[#262626]">
          <Button variant="outline" type="button" onClick={onClose}>
            Close
          </Button>
        </div>
      </div>
    </Dialog>
  );
};
