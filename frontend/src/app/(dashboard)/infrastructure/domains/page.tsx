"use client";

import React, { useState, useEffect } from "react";
import {
  Globe,
  Plus,
  Search,
  RefreshCw,
  ExternalLink,
  Shield,
  ShieldCheck,
  AlertCircle,
  Clock,
  Trash2,
  Settings2,
  Server,
  Box,
  CheckCircle2,
  Lock,
  ArrowUpRight,
} from "lucide-react";
import { CustomDomain, domainService } from "@/services/domain.service";
import { AddDomainModal } from "@/components/domains/AddDomainModal";
import { DnsInstructionModal } from "@/components/domains/DnsInstructionModal";

export default function DomainsPage() {
  const [domains, setDomains] = useState<CustomDomain[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");

  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [selectedDnsDomain, setSelectedDnsDomain] = useState<CustomDomain | null>(null);
  const [verifyingId, setVerifyingId] = useState<string | null>(null);

  useEffect(() => {
    loadDomains();
  }, []);

  const loadDomains = async () => {
    try {
      setLoading(true);
      const data = await domainService.listDomains();
      setDomains(data);
    } catch (err: any) {
      console.error("Failed to load domains:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleQuickVerify = async (domain: CustomDomain) => {
    try {
      setVerifyingId(domain.id);
      const res = await domainService.verifyDomain(domain.id);
      if (res.verified) {
        await loadDomains();
      } else {
        setSelectedDnsDomain(domain);
      }
    } catch (err: any) {
      console.error("Verification failed:", err);
      setSelectedDnsDomain(domain);
    } finally {
      setVerifyingId(null);
    }
  };

  const handleDelete = async (id: string, name: string) => {
    if (!confirm(`Are you sure you want to remove custom domain '${name}' and its reverse proxy routing?`)) {
      return;
    }
    try {
      await domainService.deleteDomain(id);
      setDomains((prev) => prev.filter((d) => d.id !== id));
    } catch (err: any) {
      alert(err?.response?.data?.message || err?.message || "Failed to delete domain");
    }
  };

  const filteredDomains = domains.filter((d) => {
    const matchesSearch =
      d.domain_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      d.target_id.toLowerCase().includes(searchQuery.toLowerCase()) ||
      (d.server_name && d.server_name.toLowerCase().includes(searchQuery.toLowerCase()));

    const matchesStatus = statusFilter === "all" || d.status === statusFilter;

    return matchesSearch && matchesStatus;
  });

  const totalDomains = domains.length;
  const activeDomains = domains.filter((d) => d.status === "active").length;
  const pendingDomains = domains.filter((d) => d.status === "pending_dns").length;
  const activeSSL = domains.filter((d) => d.ssl_status === "active").length;

  return (
    <div className="min-h-screen bg-[#0a0a0a] text-[#ededed] p-6 lg:p-8 space-y-6">
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-[#262626] pb-6">
        <div>
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              <Globe className="h-5 w-5" />
            </div>
            <div>
              <h1 className="text-xl font-bold text-[#ededed] tracking-tight">Custom Domains & Ingress</h1>
              <p className="text-xs text-zinc-400">
                Map custom public domains to containers and services with automated TLS/SSL
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={loadDomains}
            disabled={loading}
            className="flex items-center gap-2 rounded-lg border border-[#262626] bg-[#141414] px-3.5 py-2 text-xs font-medium text-zinc-300 hover:bg-[#1a1a1a] transition-colors"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${loading ? "animate-spin" : ""}`} />
            <span>Refresh</span>
          </button>
          <button
            onClick={() => setIsAddModalOpen(true)}
            className="flex items-center gap-2 rounded-lg bg-emerald-500 px-4 py-2 text-xs font-medium text-zinc-950 hover:bg-emerald-400 transition-colors shadow-sm"
          >
            <Plus className="h-4 w-4" />
            <span>Add Custom Domain</span>
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="rounded-xl border border-[#262626] bg-[#121212] p-4 flex items-center justify-between">
          <div>
            <p className="text-xs text-zinc-400 font-medium">Total Domains</p>
            <h3 className="text-2xl font-bold text-[#ededed] mt-1">{totalDomains}</h3>
          </div>
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-zinc-800/60 text-zinc-300">
            <Globe className="h-5 w-5" />
          </div>
        </div>

        <div className="rounded-xl border border-[#262626] bg-[#121212] p-4 flex items-center justify-between">
          <div>
            <p className="text-xs text-zinc-400 font-medium">Active Ingress Routes</p>
            <h3 className="text-2xl font-bold text-emerald-400 mt-1">{activeDomains}</h3>
          </div>
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-400">
            <CheckCircle2 className="h-5 w-5" />
          </div>
        </div>

        <div className="rounded-xl border border-[#262626] bg-[#121212] p-4 flex items-center justify-between">
          <div>
            <p className="text-xs text-zinc-400 font-medium">Active SSL / TLS</p>
            <h3 className="text-2xl font-bold text-sky-400 mt-1">{activeSSL}</h3>
          </div>
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-sky-500/10 text-sky-400">
            <ShieldCheck className="h-5 w-5" />
          </div>
        </div>

        <div className="rounded-xl border border-[#262626] bg-[#121212] p-4 flex items-center justify-between">
          <div>
            <p className="text-xs text-zinc-400 font-medium">Pending DNS Setup</p>
            <h3 className="text-2xl font-bold text-amber-400 mt-1">{pendingDomains}</h3>
          </div>
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-amber-500/10 text-amber-400">
            <Clock className="h-5 w-5" />
          </div>
        </div>
      </div>

      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500" />
          <input
            type="text"
            placeholder="Search by domain, container name, or server..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full rounded-lg border border-[#262626] bg-[#121212] pl-10 pr-4 py-2 text-xs text-[#ededed] placeholder-zinc-500 focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500"
          />
        </div>

        <div className="flex items-center gap-2">
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="rounded-lg border border-[#262626] bg-[#121212] px-3 py-2 text-xs text-[#ededed] focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500"
          >
            <option value="all">All Statuses</option>
            <option value="active">Active Routes</option>
            <option value="pending_dns">Pending DNS</option>
            <option value="error">Error</option>
          </select>
        </div>
      </div>

      <div className="rounded-xl border border-[#262626] bg-[#121212] overflow-hidden shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="border-b border-[#262626] bg-[#161616] text-zinc-400 uppercase tracking-wider font-medium">
              <tr>
                <th className="px-5 py-3.5">Domain Name</th>
                <th className="px-5 py-3.5">Target Host / Server</th>
                <th className="px-5 py-3.5">Routing Target</th>
                <th className="px-5 py-3.5">DNS Status</th>
                <th className="px-5 py-3.5">SSL / HTTPS</th>
                <th className="px-5 py-3.5 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#202020]">
              {filteredDomains.map((d) => (
                <tr key={d.id} className="hover:bg-[#161616]/70 transition-colors">
                  <td className="px-5 py-4">
                    <div className="flex items-center gap-2">
                      <Globe className="h-4 w-4 text-emerald-400 shrink-0" />
                      <span className="font-semibold text-sm text-[#ededed] font-mono">{d.domain_name}</span>
                      {d.status === "active" && (
                        <a
                          href={`https://${d.domain_name}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-zinc-500 hover:text-emerald-400 transition-colors"
                          title="Open in new tab"
                        >
                          <ArrowUpRight className="h-3.5 w-3.5" />
                        </a>
                      )}
                    </div>
                  </td>

                  <td className="px-5 py-4">
                    <div className="flex items-center gap-2">
                      <Server className="h-3.5 w-3.5 text-zinc-400" />
                      <div>
                        <span className="text-zinc-200 font-medium">{d.server_name || "Control Plane Host"}</span>
                        <span className="block text-[11px] text-zinc-500 font-mono">
                          {d.server_public_ip || "127.0.0.1"}
                        </span>
                      </div>
                    </div>
                  </td>

                  <td className="px-5 py-4">
                    <div className="flex items-center gap-2">
                      <Box className="h-3.5 w-3.5 text-zinc-400" />
                      <div>
                        <span className="text-zinc-200 font-medium uppercase text-[11px] bg-zinc-800/60 px-1.5 py-0.5 rounded mr-1.5">
                          {d.target_type}
                        </span>
                        <span className="font-mono text-zinc-300">
                          {d.target_id}:{d.target_port}
                        </span>
                      </div>
                    </div>
                  </td>

                  <td className="px-5 py-4">
                    {d.status === "active" ? (
                      <span className="inline-flex items-center gap-1.5 rounded-full bg-emerald-500/10 px-2.5 py-1 text-[11px] font-medium text-emerald-400 border border-emerald-500/20">
                        <span className="h-1.5 w-1.5 rounded-full bg-emerald-400" />
                        Active
                      </span>
                    ) : d.status === "pending_dns" ? (
                      <span className="inline-flex items-center gap-1.5 rounded-full bg-amber-500/10 px-2.5 py-1 text-[11px] font-medium text-amber-400 border border-amber-500/20">
                        <span className="h-1.5 w-1.5 rounded-full bg-amber-400 animate-pulse" />
                        Pending DNS
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1.5 rounded-full bg-rose-500/10 px-2.5 py-1 text-[11px] font-medium text-rose-400 border border-rose-500/20">
                        <span className="h-1.5 w-1.5 rounded-full bg-rose-400" />
                        Error
                      </span>
                    )}
                  </td>

                  <td className="px-5 py-4">
                    {d.ssl_status === "active" ? (
                      <span className="inline-flex items-center gap-1.5 rounded-full bg-sky-500/10 px-2.5 py-1 text-[11px] font-medium text-sky-400 border border-sky-500/20">
                        <Lock className="h-3 w-3" />
                        Active TLS
                      </span>
                    ) : d.ssl_status === "pending" ? (
                      <span className="inline-flex items-center gap-1.5 rounded-full bg-zinc-800 px-2.5 py-1 text-[11px] font-medium text-zinc-400">
                        <Shield className="h-3 w-3" />
                        Auto-Provisioning
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1.5 rounded-full bg-zinc-800/40 px-2.5 py-1 text-[11px] font-medium text-zinc-500">
                        HTTP Only
                      </span>
                    )}
                  </td>

                  <td className="px-5 py-4 text-right">
                    <div className="flex items-center justify-end gap-2">
                      <button
                        onClick={() => handleQuickVerify(d)}
                        disabled={verifyingId === d.id}
                        className="rounded-lg border border-[#262626] bg-[#181818] px-2.5 py-1.5 text-xs text-zinc-300 hover:bg-[#222222] hover:text-emerald-400 transition-colors"
                        title="Verify DNS records now"
                      >
                        <RefreshCw className={`h-3.5 w-3.5 ${verifyingId === d.id ? "animate-spin text-emerald-400" : ""}`} />
                      </button>

                      <button
                        onClick={() => setSelectedDnsDomain(d)}
                        className="rounded-lg border border-[#262626] bg-[#181818] px-2.5 py-1.5 text-xs text-zinc-300 hover:bg-[#222222] hover:text-sky-400 transition-colors"
                        title="View DNS Records & Setup Guide"
                      >
                        DNS Records
                      </button>

                      <button
                        onClick={() => handleDelete(d.id, d.domain_name)}
                        className="rounded-lg border border-[#262626] bg-[#181818] p-1.5 text-zinc-400 hover:bg-rose-500/10 hover:border-rose-500/30 hover:text-rose-400 transition-colors"
                        title="Delete custom domain"
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}

              {filteredDomains.length === 0 && !loading && (
                <tr>
                  <td colSpan={6} className="px-6 py-12 text-center">
                    <div className="flex flex-col items-center justify-center max-w-sm mx-auto">
                      <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-zinc-900 border border-zinc-800 text-zinc-500 mb-3">
                        <Globe className="h-6 w-6" />
                      </div>
                      <h4 className="text-sm font-semibold text-zinc-200">No Custom Domains Found</h4>
                      <p className="mt-1 text-xs text-zinc-500 text-center leading-relaxed">
                        Connect your first domain name to route incoming traffic directly to your containers with automatic SSL certificates.
                      </p>
                      <button
                        onClick={() => setIsAddModalOpen(true)}
                        className="mt-4 flex items-center gap-2 rounded-lg bg-emerald-500 px-4 py-2 text-xs font-medium text-zinc-950 hover:bg-emerald-400 transition-colors"
                      >
                        <Plus className="h-3.5 w-3.5" />
                        <span>Add First Domain</span>
                      </button>
                    </div>
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      <AddDomainModal
        isOpen={isAddModalOpen}
        onClose={() => setIsAddModalOpen(false)}
        onSuccess={loadDomains}
      />

      <DnsInstructionModal
        domain={selectedDnsDomain}
        isOpen={!!selectedDnsDomain}
        onClose={() => setSelectedDnsDomain(null)}
        onVerified={() => {
          loadDomains();
        }}
      />
    </div>
  );
}
