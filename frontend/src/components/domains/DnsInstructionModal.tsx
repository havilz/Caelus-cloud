"use client";

import React, { useState } from "react";
import { X, Globe, Copy, Check, RefreshCw, ShieldCheck, AlertCircle, CheckCircle2, Server, HelpCircle } from "lucide-react";
import { CustomDomain, domainService, VerifyDomainResponse } from "@/services/domain.service";

interface DnsInstructionModalProps {
  domain: CustomDomain | null;
  isOpen: boolean;
  onClose: () => void;
  onVerified: () => void;
}

export const DnsInstructionModal: React.FC<DnsInstructionModalProps> = ({
  domain,
  isOpen,
  onClose,
  onVerified,
}) => {
  const [copiedField, setCopiedField] = useState<string | null>(null);
  const [verifying, setVerifying] = useState(false);
  const [result, setResult] = useState<VerifyDomainResponse | null>(null);

  if (!isOpen || !domain) return null;

  const handleCopy = (text: string, field: string) => {
    navigator.clipboard.writeText(text);
    setCopiedField(field);
    setTimeout(() => setCopiedField(null), 2000);
  };

  const handleVerifyNow = async () => {
    try {
      setVerifying(true);
      setResult(null);
      const res = await domainService.verifyDomain(domain.id);
      setResult(res);
      if (res.verified) {
        onVerified();
      }
    } catch (err: any) {
      console.error("Verification error:", err);
    } finally {
      setVerifying(false);
    }
  };

  const isSubdomain = domain.domain_name.split(".").length > 2;
  const aHost = isSubdomain ? domain.domain_name.split(".")[0] : "@";
  const targetIP = domain.server_public_ip || "127.0.0.1";

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4 overflow-y-auto">
      <div className="w-full max-w-xl rounded-xl border border-[#262626] bg-[#121212] p-6 shadow-2xl animate-in fade-in zoom-in-95 duration-200">
        <div className="flex items-center justify-between border-b border-[#262626] pb-4 mb-5">
          <div className="flex items-center gap-2.5">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-sky-500/10 text-sky-400 border border-sky-500/20">
              <Globe className="h-5 w-5" />
            </div>
            <div>
              <h2 className="text-base font-semibold text-[#ededed]">DNS Setup & Verification</h2>
              <p className="text-xs text-zinc-400 font-mono">{domain.domain_name}</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="rounded-lg p-1.5 text-zinc-400 hover:bg-[#1a1a1a] hover:text-zinc-200 transition-colors"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        <div className="space-y-4">
          <div className="rounded-lg border border-[#262626] bg-[#181818] p-3.5 text-xs text-zinc-300">
            <p className="font-medium text-zinc-200 mb-1">How to configure your DNS records:</p>
            <p className="text-zinc-400 leading-relaxed">
              Sign in to your domain registrar (e.g. Cloudflare, Namecheap, Niagahoster) and add the following DNS records for <strong className="text-emerald-400">{domain.domain_name}</strong>:
            </p>
          </div>

          {/* Record A */}
          <div className="rounded-lg border border-[#262626] bg-[#161616] p-4 space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-xs font-semibold text-emerald-400 uppercase tracking-wider">
                1. Required Record A (IPv4 Pointing)
              </span>
              <span className="text-[11px] text-zinc-500">Directs traffic to your server</span>
            </div>

            <div className="grid grid-cols-12 gap-2 text-xs">
              <div className="col-span-3 rounded-lg border border-[#262626] bg-[#121212] p-2.5">
                <span className="text-[10px] text-zinc-500 uppercase block font-mono">Type</span>
                <span className="font-mono font-medium text-[#ededed]">A</span>
              </div>
              <div className="col-span-4 rounded-lg border border-[#262626] bg-[#121212] p-2.5 flex items-center justify-between">
                <div>
                  <span className="text-[10px] text-zinc-500 uppercase block font-mono">Name / Host</span>
                  <span className="font-mono font-medium text-[#ededed]">{aHost}</span>
                </div>
                <button
                  onClick={() => handleCopy(aHost, "a_host")}
                  className="text-zinc-500 hover:text-zinc-300 p-1"
                >
                  {copiedField === "a_host" ? <Check className="h-3.5 w-3.5 text-emerald-400" /> : <Copy className="h-3.5 w-3.5" />}
                </button>
              </div>
              <div className="col-span-5 rounded-lg border border-[#262626] bg-[#121212] p-2.5 flex items-center justify-between">
                <div>
                  <span className="text-[10px] text-zinc-500 uppercase block font-mono">Target IP</span>
                  <span className="font-mono font-medium text-[#ededed]">{targetIP}</span>
                </div>
                <button
                  onClick={() => handleCopy(targetIP, "a_ip")}
                  className="text-zinc-500 hover:text-zinc-300 p-1"
                >
                  {copiedField === "a_ip" ? <Check className="h-3.5 w-3.5 text-emerald-400" /> : <Copy className="h-3.5 w-3.5" />}
                </button>
              </div>
            </div>
          </div>

          {/* Record TXT (Optional Verification) */}
          <div className="rounded-lg border border-[#262626] bg-[#161616] p-4 space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-xs font-semibold text-sky-400 uppercase tracking-wider">
                2. Ownership Verification TXT (Optional)
              </span>
              <span className="text-[11px] text-zinc-500">Proves domain ownership</span>
            </div>

            <div className="grid grid-cols-12 gap-2 text-xs">
              <div className="col-span-3 rounded-lg border border-[#262626] bg-[#121212] p-2.5">
                <span className="text-[10px] text-zinc-500 uppercase block font-mono">Type</span>
                <span className="font-mono font-medium text-[#ededed]">TXT</span>
              </div>
              <div className="col-span-4 rounded-lg border border-[#262626] bg-[#121212] p-2.5 flex items-center justify-between">
                <div>
                  <span className="text-[10px] text-zinc-500 uppercase block font-mono">Host</span>
                  <span className="font-mono font-medium text-[#ededed] truncate">_caelus-verify</span>
                </div>
                <button
                  onClick={() => handleCopy(`_caelus-verify`, "txt_host")}
                  className="text-zinc-500 hover:text-zinc-300 p-1"
                >
                  {copiedField === "txt_host" ? <Check className="h-3.5 w-3.5 text-emerald-400" /> : <Copy className="h-3.5 w-3.5" />}
                </button>
              </div>
              <div className="col-span-5 rounded-lg border border-[#262626] bg-[#121212] p-2.5 flex items-center justify-between">
                <div className="truncate mr-1">
                  <span className="text-[10px] text-zinc-500 uppercase block font-mono">Value</span>
                  <span className="font-mono font-medium text-[#ededed] truncate">{domain.verification_token}</span>
                </div>
                <button
                  onClick={() => handleCopy(domain.verification_token, "txt_val")}
                  className="text-zinc-500 hover:text-zinc-300 p-1 shrink-0"
                >
                  {copiedField === "txt_val" ? <Check className="h-3.5 w-3.5 text-emerald-400" /> : <Copy className="h-3.5 w-3.5" />}
                </button>
              </div>
            </div>
          </div>

          {result && (
            <div
              className={`rounded-lg border p-3.5 text-xs ${
                result.verified
                  ? "border-emerald-500/30 bg-emerald-500/10 text-emerald-300"
                  : "border-amber-500/30 bg-amber-500/10 text-amber-300"
              }`}
            >
              <div className="flex items-start gap-2.5">
                {result.verified ? (
                  <CheckCircle2 className="h-4 w-4 text-emerald-400 shrink-0 mt-0.5" />
                ) : (
                  <AlertCircle className="h-4 w-4 text-amber-400 shrink-0 mt-0.5" />
                )}
                <div>
                  <p className="font-medium">{result.verified ? "DNS Verified Successfully!" : "Verification In Progress"}</p>
                  <p className="mt-0.5 text-zinc-400 leading-relaxed">{result.message}</p>
                </div>
              </div>
            </div>
          )}

          <div className="flex items-center justify-between pt-2">
            <div className="flex items-center gap-1.5 text-[11px] text-zinc-500">
              <HelpCircle className="h-3.5 w-3.5" />
              <span>DNS propagation typically takes 1 - 5 minutes</span>
            </div>
            <div className="flex items-center gap-2">
              <button
                type="button"
                onClick={onClose}
                className="rounded-lg border border-[#262626] bg-[#181818] px-4 py-2 text-xs font-medium text-zinc-300 hover:bg-[#222222] transition-colors"
              >
                Close
              </button>
              <button
                type="button"
                onClick={handleVerifyNow}
                disabled={verifying}
                className="flex items-center gap-2 rounded-lg bg-emerald-500 px-4 py-2 text-xs font-medium text-zinc-950 hover:bg-emerald-400 disabled:opacity-50 transition-colors shadow-sm"
              >
                <RefreshCw className={`h-3.5 w-3.5 ${verifying ? "animate-spin" : ""}`} />
                <span>{verifying ? "Checking DNS..." : "Verify DNS Now"}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
