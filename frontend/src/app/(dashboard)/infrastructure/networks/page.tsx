"use client";

import React, { useState } from "react";
import {
  Network,
  Plus,
  RefreshCw,
  ShieldCheck,
  Server,
  Globe,
  Radio,
  Trash2,
  CheckCircle2,
  AlertCircle,
  Search,
  Filter,
  ArrowRight,
  Layers,
  Lock,
  ExternalLink,
} from "lucide-react";
import { Dialog } from "@/components/ui/dialog";
import { AppTheme } from "@/core/theme";

interface VirtualNetwork {
  id: string;
  name: string;
  type: "vpc" | "bridge" | "overlay";
  cidr: string;
  gateway: string;
  region: string;
  attachedServers: number;
  status: "active" | "provisioning" | "idle";
  createdAt: string;
}

interface FirewallRule {
  id: string;
  name: string;
  direction: "inbound" | "outbound";
  protocol: "tcp" | "udp" | "icmp" | "all";
  portRange: string;
  source: string;
  action: "allow" | "deny";
  targetNetworkId: string;
}

export default function NetworksManagementPage() {
  const [networks, setNetworks] = useState<VirtualNetwork[]>([]);
  const [firewallRules, setFirewallRules] = useState<FirewallRule[]>([]);
  const [activeTab, setActiveTab] = useState<"networks" | "firewall">("networks");
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [isCreateModalOpen, setIsCreateModalOpen] = useState<boolean>(false);
  const [isRuleModalOpen, setIsRuleModalOpen] = useState<boolean>(false);

  // New Network Form State
  const [newNetName, setNewNetName] = useState<string>("");
  const [newNetType, setNewNetType] = useState<"vpc" | "bridge" | "overlay">("vpc");
  const [newNetCidr, setNewNetCidr] = useState<string>("10.20.0.0/16");
  const [newNetRegion, setNewNetRegion] = useState<string>("ap-southeast-1");

  // New Rule Form State
  const [newRuleName, setNewRuleName] = useState<string>("");
  const [newRuleDirection, setNewRuleDirection] = useState<"inbound" | "outbound">("inbound");
  const [newRuleProtocol, setNewRuleProtocol] = useState<"tcp" | "udp" | "icmp" | "all">("tcp");
  const [newRulePorts, setNewRulePorts] = useState<string>("8080");
  const [newRuleSource, setNewRuleSource] = useState<string>("0.0.0.0/0");
  const [newRuleAction, setNewRuleAction] = useState<"allow" | "deny">("allow");

  const handleCreateNetwork = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newNetName.trim()) return;

    const newNet: VirtualNetwork = {
      id: `net-${Date.now()}`,
      name: newNetName.trim().toLowerCase().replace(/\s+/g, "-"),
      type: newNetType,
      cidr: newNetCidr,
      gateway: newNetCidr.replace(".0.0/16", ".0.1").replace(".0/24", ".1"),
      region: newNetRegion,
      attachedServers: 0,
      status: "active",
      createdAt: new Date().toISOString().split("T")[0],
    };

    setNetworks([newNet, ...networks]);
    setNewNetName("");
    setIsCreateModalOpen(false);
  };

  const handleCreateRule = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newRuleName.trim()) return;

    const newRule: FirewallRule = {
      id: `fw-${Date.now()}`,
      name: newRuleName.trim(),
      direction: newRuleDirection,
      protocol: newRuleProtocol,
      portRange: newRulePorts.trim(),
      source: newRuleSource.trim(),
      action: newRuleAction,
      targetNetworkId: networks[0]?.id || "net-vpc-default",
    };

    setFirewallRules([newRule, ...firewallRules]);
    setNewRuleName("");
    setIsRuleModalOpen(false);
  };

  const handleDeleteNetwork = (id: string) => {
    setNetworks(networks.filter((n) => n.id !== id));
  };

  const handleDeleteRule = (id: string) => {
    setFirewallRules(firewallRules.filter((r) => r.id !== id));
  };

  const filteredNetworks = networks.filter((n) =>
    n.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    n.cidr.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const filteredRules = firewallRules.filter((r) =>
    r.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    r.portRange.toLowerCase().includes(searchQuery.toLowerCase()) ||
    r.source.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-100 flex items-center gap-2.5">
            <Network className="w-6 h-6 text-emerald-400" />
            Virtual Networks & Security Groups
          </h1>
          <p className="text-sm text-slate-400 mt-1">
            Kelola Virtual Private Cloud (VPC), Subnet, Docker Overlay Networks, dan Firewall Ingress/Egress.
          </p>
        </div>
        <div className="flex items-center gap-3">
          {activeTab === "networks" ? (
            <button
              onClick={() => setIsCreateModalOpen(true)}
              className="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-slate-950 bg-emerald-400 hover:bg-emerald-300 rounded-lg transition-colors shadow-sm"
            >
              <Plus className="w-4 h-4" />
              Create Virtual Network
            </button>
          ) : (
            <button
              onClick={() => setIsRuleModalOpen(true)}
              className="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-slate-950 bg-emerald-400 hover:bg-emerald-300 rounded-lg transition-colors shadow-sm"
            >
              <Plus className="w-4 h-4" />
              Add Firewall Rule
            </button>
          )}
        </div>
      </div>

      {/* Top Stats Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="p-4 rounded-xl bg-slate-900/60 border border-slate-800/80">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-slate-400">Total Virtual Networks</span>
            <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-400">
              <Network className="w-4 h-4" />
            </div>
          </div>
          <p className="text-2xl font-bold text-slate-100 mt-2">{networks.length}</p>
          <span className="text-xs text-emerald-400 flex items-center gap-1 mt-1">
            <CheckCircle2 className="w-3 h-3" /> All active & routed
          </span>
        </div>

        <div className="p-4 rounded-xl bg-slate-900/60 border border-slate-800/80">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-slate-400">Attached Compute Nodes</span>
            <div className="p-2 rounded-lg bg-cyan-500/10 text-cyan-400">
              <Server className="w-4 h-4" />
            </div>
          </div>
          <p className="text-2xl font-bold text-slate-100 mt-2">
            {networks.reduce((acc, n) => acc + n.attachedServers, 0)}
          </p>
          <span className="text-xs text-slate-400 mt-1 block">Connected instances</span>
        </div>

        <div className="p-4 rounded-xl bg-slate-900/60 border border-slate-800/80">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-slate-400">Firewall Rules</span>
            <div className="p-2 rounded-lg bg-amber-500/10 text-amber-400">
              <ShieldCheck className="w-4 h-4" />
            </div>
          </div>
          <p className="text-2xl font-bold text-slate-100 mt-2">{firewallRules.length}</p>
          <span className="text-xs text-amber-400 mt-1 block">Ingress & egress filters</span>
        </div>

        <div className="p-4 rounded-xl bg-slate-900/60 border border-slate-800/80">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-slate-400">Default Gateway</span>
            <div className="p-2 rounded-lg bg-purple-500/10 text-purple-400">
              <Globe className="w-4 h-4" />
            </div>
          </div>
          <p className="text-base font-bold text-slate-100 mt-2 font-mono">
            {networks.length > 0 ? networks[0].gateway : "-"}
          </p>
          <span className="text-xs text-slate-400 mt-1 block">
            {networks.length > 0 ? "NAT routing active" : "No active gateway"}
          </span>
        </div>
      </div>

      {/* Navigation Tabs & Search */}
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-4 border-b border-slate-800 pb-4">
        <div className="flex items-center gap-2">
          <button
            onClick={() => setActiveTab("networks")}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
              activeTab === "networks"
                ? "bg-slate-800 text-emerald-400 border border-slate-700 shadow-sm"
                : "text-slate-400 hover:text-slate-200 hover:bg-slate-900"
            }`}
          >
            Virtual Networks ({networks.length})
          </button>
          <button
            onClick={() => setActiveTab("firewall")}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
              activeTab === "firewall"
                ? "bg-slate-800 text-emerald-400 border border-slate-700 shadow-sm"
                : "text-slate-400 hover:text-slate-200 hover:bg-slate-900"
            }`}
          >
            Firewall & Security Groups ({firewallRules.length})
          </button>
        </div>

        <div className="relative">
          <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
          <input
            type="text"
            placeholder={activeTab === "networks" ? "Search networks or CIDR..." : "Search rules or ports..."}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full sm:w-64 pl-9 pr-4 py-2 text-xs rounded-lg bg-slate-900/80 border border-slate-800 text-slate-200 placeholder-slate-500 focus:outline-none focus:border-emerald-500/50"
          />
        </div>
      </div>

      {/* Tab 1: Virtual Networks List */}
      {activeTab === "networks" && (
        filteredNetworks.length === 0 ? (
          <div className="p-12 text-center rounded-xl bg-slate-900/60 border border-slate-800/80">
            <div className="flex flex-col items-center justify-center space-y-3">
              <div className="p-3 rounded-full bg-slate-800/60 border border-slate-700 text-slate-400">
                <Network className="w-6 h-6" />
              </div>
              <div className="space-y-1">
                <p className="text-sm font-semibold text-slate-200">No Virtual Networks Found</p>
                <p className="text-xs text-slate-400 max-w-sm">
                  {searchQuery
                    ? `No networks matching "${searchQuery}".`
                    : "Create software-defined VPC, overlay, or bridge networks to connect your infrastructure nodes."}
                </p>
              </div>
              {!searchQuery && (
                <button
                  onClick={() => setIsCreateModalOpen(true)}
                  className="mt-2 px-3.5 py-2 rounded-lg text-xs font-semibold bg-emerald-500 hover:bg-emerald-400 text-slate-950 transition-colors flex items-center gap-1.5"
                >
                  <Plus className="w-3.5 h-3.5" />
                  <span>Create Network</span>
                </button>
              )}
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {filteredNetworks.map((net) => (
              <div
                key={net.id}
                className="p-5 rounded-xl bg-slate-900/60 border border-slate-800/80 hover:border-slate-700 transition-all flex flex-col justify-between"
              >
                <div>
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-2.5">
                      <div className="p-2 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400">
                        <Network className="w-5 h-5" />
                      </div>
                      <div>
                        <h3 className="text-sm font-semibold text-slate-200">{net.name}</h3>
                        <span className="text-[11px] font-mono uppercase text-slate-400">{net.type}</span>
                      </div>
                    </div>
                    <span className="px-2 py-0.5 rounded text-[10px] font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                      {net.status}
                    </span>
                  </div>

                  <div className="mt-4 space-y-2 text-xs">
                    <div className="flex justify-between py-1 border-b border-slate-800/60">
                      <span className="text-slate-500">CIDR Block</span>
                      <span className="font-mono text-slate-200 font-semibold">{net.cidr}</span>
                    </div>
                    <div className="flex justify-between py-1 border-b border-slate-800/60">
                      <span className="text-slate-500">Gateway</span>
                      <span className="font-mono text-slate-300">{net.gateway}</span>
                    </div>
                    <div className="flex justify-between py-1 border-b border-slate-800/60">
                      <span className="text-slate-500">Region</span>
                      <span className="text-slate-300">{net.region}</span>
                    </div>
                    <div className="flex justify-between py-1">
                      <span className="text-slate-500">Connected Nodes</span>
                      <span className="text-emerald-400 font-medium">{net.attachedServers} instances</span>
                    </div>
                  </div>
                </div>

                <div className="mt-5 pt-3 border-t border-slate-800/60 flex items-center justify-between">
                  <span className="text-[11px] text-slate-500">Created {net.createdAt}</span>
                  <button
                    onClick={() => handleDeleteNetwork(net.id)}
                    className="p-1.5 rounded-lg text-slate-500 hover:text-rose-400 hover:bg-rose-500/10 transition-colors"
                    title="Delete Network"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )
      )}

      {/* Tab 2: Firewall Rules Table */}
      {activeTab === "firewall" && (
        <div className="rounded-xl bg-slate-900/60 border border-slate-800/80 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-slate-950/60 text-slate-400 border-b border-slate-800 font-medium">
                <tr>
                  <th className="px-4 py-3">Rule Name</th>
                  <th className="px-4 py-3">Direction</th>
                  <th className="px-4 py-3">Protocol</th>
                  <th className="px-4 py-3">Port Range</th>
                  <th className="px-4 py-3">Source CIDR</th>
                  <th className="px-4 py-3">Action</th>
                  <th className="px-4 py-3 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60 text-slate-300">
                {filteredRules.length === 0 ? (
                  <tr>
                    <td colSpan={7} className="px-6 py-12 text-center">
                      <div className="flex flex-col items-center justify-center space-y-3">
                        <div className="p-3 rounded-full bg-slate-800/60 border border-slate-700 text-slate-400">
                          <ShieldCheck className="w-6 h-6" />
                        </div>
                        <div className="space-y-1">
                          <p className="text-sm font-semibold text-slate-200">No Firewall Rules Configured</p>
                          <p className="text-xs text-slate-400 max-w-sm">
                            {searchQuery
                              ? `No firewall rules matching "${searchQuery}".`
                              : "Define ingress and egress traffic filtering rules to protect your network boundaries."}
                          </p>
                        </div>
                        {!searchQuery && (
                          <button
                            onClick={() => setIsRuleModalOpen(true)}
                            className="mt-2 px-3.5 py-2 rounded-lg text-xs font-semibold bg-emerald-500 hover:bg-emerald-400 text-slate-950 transition-colors flex items-center gap-1.5"
                          >
                            <Plus className="w-3.5 h-3.5" />
                            <span>Add Firewall Rule</span>
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ) : (
                  filteredRules.map((rule) => (
                    <tr key={rule.id} className="hover:bg-slate-800/30 transition-colors">
                      <td className="px-4 py-3.5 font-medium text-slate-200">
                        <div className="flex items-center gap-2">
                          <Lock className="w-3.5 h-3.5 text-slate-500" />
                          {rule.name}
                        </div>
                      </td>
                      <td className="px-4 py-3.5">
                        <span className="px-2 py-0.5 rounded text-[10px] font-medium bg-slate-800 text-slate-300 uppercase">
                          {rule.direction}
                        </span>
                      </td>
                      <td className="px-4 py-3.5 font-mono uppercase text-slate-300">{rule.protocol}</td>
                      <td className="px-4 py-3.5 font-mono text-emerald-400 font-semibold">{rule.portRange}</td>
                      <td className="px-4 py-3.5 font-mono text-slate-400">{rule.source}</td>
                      <td className="px-4 py-3.5">
                        <span
                          className={`px-2 py-0.5 rounded text-[10px] font-semibold uppercase ${
                            rule.action === "allow"
                              ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                              : "bg-rose-500/10 text-rose-400 border border-rose-500/20"
                          }`}
                        >
                          {rule.action}
                        </span>
                      </td>
                      <td className="px-4 py-3.5 text-right">
                        <button
                          onClick={() => handleDeleteRule(rule.id)}
                          className="p-1.5 rounded-lg text-slate-500 hover:text-rose-400 hover:bg-rose-500/10 transition-colors"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Modal: Create Network */}
      <Dialog
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        title="Create Virtual Network"
        description="Konfigurasikan ruang alamat IP terisolasi untuk server dan kontainer Anda."
      >
        <form onSubmit={handleCreateNetwork} className="space-y-4 mt-2">
          <div>
            <label className="block text-xs font-medium text-slate-300 mb-1">Network Name</label>
            <input
              type="text"
              required
              placeholder="misal: production-vpc"
              value={newNetName}
              onChange={(e) => setNewNetName(e.target.value)}
              className="w-full px-3 py-2 text-xs rounded-lg bg-slate-950 border border-slate-800 text-slate-200 focus:outline-none focus:border-emerald-500/50"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-300 mb-1">Network Topology</label>
            <select
              value={newNetType}
              onChange={(e) => setNewNetType(e.target.value as any)}
              className="w-full px-3 py-2 text-xs rounded-lg bg-slate-950 border border-slate-800 text-slate-200 focus:outline-none focus:border-emerald-500/50"
            >
              <option value="vpc">Virtual Private Cloud (VPC)</option>
              <option value="bridge">Host Bridge Network</option>
              <option value="overlay">Multi-Host Overlay Mesh</option>
            </select>
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-300 mb-1">CIDR Block Range</label>
            <input
              type="text"
              required
              placeholder="10.20.0.0/16"
              value={newNetCidr}
              onChange={(e) => setNewNetCidr(e.target.value)}
              className="w-full px-3 py-2 text-xs rounded-lg bg-slate-950 border border-slate-800 text-slate-200 font-mono focus:outline-none focus:border-emerald-500/50"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-300 mb-1">Region / Datacenter</label>
            <input
              type="text"
              value={newNetRegion}
              onChange={(e) => setNewNetRegion(e.target.value)}
              className="w-full px-3 py-2 text-xs rounded-lg bg-slate-950 border border-slate-800 text-slate-200 focus:outline-none focus:border-emerald-500/50"
            />
          </div>

          <div className="flex justify-end gap-2 pt-3 border-t border-slate-800">
            <button
              type="button"
              onClick={() => setIsCreateModalOpen(false)}
              className="px-4 py-2 text-xs text-slate-400 hover:text-slate-200 bg-slate-800 rounded-lg"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 text-xs font-medium text-slate-950 bg-emerald-400 hover:bg-emerald-300 rounded-lg transition-colors"
            >
              Create Network
            </button>
          </div>
        </form>
      </Dialog>

      {/* Modal: Add Firewall Rule */}
      <Dialog
        isOpen={isRuleModalOpen}
        onClose={() => setIsRuleModalOpen(false)}
        title="Add Firewall Security Rule"
        description="Atur kebijakan izin atau blokir paket jaringan berdasarkan port dan alamat IP."
      >
        <form onSubmit={handleCreateRule} className="space-y-4 mt-2">
          <div>
            <label className="block text-xs font-medium text-slate-300 mb-1">Rule Name</label>
            <input
              type="text"
              required
              placeholder="misal: Allow Web API"
              value={newRuleName}
              onChange={(e) => setNewRuleName(e.target.value)}
              className="w-full px-3 py-2 text-xs rounded-lg bg-slate-950 border border-slate-800 text-slate-200 focus:outline-none focus:border-emerald-500/50"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-slate-300 mb-1">Direction</label>
              <select
                value={newRuleDirection}
                onChange={(e) => setNewRuleDirection(e.target.value as any)}
                className="w-full px-3 py-2 text-xs rounded-lg bg-slate-950 border border-slate-800 text-slate-200"
              >
                <option value="inbound">Inbound (Ingress)</option>
                <option value="outbound">Outbound (Egress)</option>
              </select>
            </div>
            <div>
              <label className="block text-xs font-medium text-slate-300 mb-1">Protocol</label>
              <select
                value={newRuleProtocol}
                onChange={(e) => setNewRuleProtocol(e.target.value as any)}
                className="w-full px-3 py-2 text-xs rounded-lg bg-slate-950 border border-slate-800 text-slate-200"
              >
                <option value="tcp">TCP</option>
                <option value="udp">UDP</option>
                <option value="icmp">ICMP</option>
                <option value="all">ALL Protocols</option>
              </select>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-slate-300 mb-1">Port Range</label>
              <input
                type="text"
                required
                placeholder="80, 443, 8000-8080"
                value={newRulePorts}
                onChange={(e) => setNewRulePorts(e.target.value)}
                className="w-full px-3 py-2 text-xs rounded-lg bg-slate-950 border border-slate-800 text-slate-200 font-mono"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-slate-300 mb-1">Action</label>
              <select
                value={newRuleAction}
                onChange={(e) => setNewRuleAction(e.target.value as any)}
                className="w-full px-3 py-2 text-xs rounded-lg bg-slate-950 border border-slate-800 text-slate-200"
              >
                <option value="allow">ALLOW</option>
                <option value="deny">DENY</option>
              </select>
            </div>
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-300 mb-1">Source CIDR</label>
            <input
              type="text"
              required
              placeholder="0.0.0.0/0 (Public) atau 10.0.0.0/16"
              value={newRuleSource}
              onChange={(e) => setNewRuleSource(e.target.value)}
              className="w-full px-3 py-2 text-xs rounded-lg bg-slate-950 border border-slate-800 text-slate-200 font-mono"
            />
          </div>

          <div className="flex justify-end gap-2 pt-3 border-t border-slate-800">
            <button
              type="button"
              onClick={() => setIsRuleModalOpen(false)}
              className="px-4 py-2 text-xs text-slate-400 hover:text-slate-200 bg-slate-800 rounded-lg"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 text-xs font-medium text-slate-950 bg-emerald-400 hover:bg-emerald-300 rounded-lg transition-colors"
            >
              Add Rule
            </button>
          </div>
        </form>
      </Dialog>
    </div>
  );
}
