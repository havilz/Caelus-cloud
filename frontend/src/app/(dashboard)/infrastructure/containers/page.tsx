'use client';

import React, { useState, useEffect, useRef, useCallback } from 'react';
import {
  Box,
  Play,
  RotateCcw,
  Terminal,
  StopCircle,
  Plus,
  RefreshCw,
  Search,
  Copy,
  Trash2,
  Check,
  CheckCircle2,
  AlertCircle,
  Clock,
  Layers,
  ArrowRight,
  Filter,
  ExternalLink,
} from 'lucide-react';
import { AppTheme } from '@/core/theme';
import { Deployment, DeploymentLog, DeploymentRequest } from '@/types/deployment';
import { deploymentService } from '@/services/deployment.service';
import { networkService, VirtualNetwork } from '@/services/network.service';

const QUICK_IMAGES = [
  { name: 'Nginx Web Server', tag: 'nginx:alpine', port: 80 },
  { name: 'Redis In-Memory Store', tag: 'redis:7-alpine', port: 6379 },
  { name: 'PostgreSQL Database', tag: 'postgres:16-alpine', port: 5432 },
  { name: 'Node.js Runtime', tag: 'node:20-alpine', port: 3000 },
];

export default function ContainersOrchestrationPage() {
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [selectedDeployment, setSelectedDeployment] = useState<Deployment | null>(null);
  const [availableNetworks, setAvailableNetworks] = useState<VirtualNetwork[]>([]);
  const [servers, setServers] = useState<any[]>([]);
  const [logs, setLogs] = useState<DeploymentLog[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isModalOpen, setIsModalOpen] = useState<boolean>(false);
  const [isStreaming, setIsStreaming] = useState<boolean>(false);
  const [streamFilter, setStreamFilter] = useState<'all' | 'stdout' | 'stderr' | 'system'>('all');
  const [autoScroll, setAutoScroll] = useState<boolean>(true);
  const [copied, setCopied] = useState<boolean>(false);

  // Deploy Form States
  const [targetServerId, setTargetServerId] = useState<string>('');
  const [appName, setAppName] = useState<string>('');
  const [imageTag, setImageTag] = useState<string>('nginx:alpine');
  const [containerName, setContainerName] = useState<string>('');
  const [selectedNetwork, setSelectedNetwork] = useState<string>('');
  const [hostPort, setHostPort] = useState<number>(8080);
  const [containerPort, setContainerPort] = useState<number>(80);
  const [envKey, setEnvKey] = useState<string>('');
  const [envVal, setEnvVal] = useState<string>('');
  const [envVars, setEnvVars] = useState<{ [key: string]: string }>({});
  const [restartPolicy, setRestartPolicy] = useState<string>('unless-stopped');
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);

  const terminalEndRef = useRef<HTMLDivElement>(null);

  const fetchDeployments = useCallback(async () => {
    try {
      setIsLoading(true);
      const data = await deploymentService.listDeployments();
      setDeployments(data);
      if (data.length > 0 && !selectedDeployment) {
        setSelectedDeployment(data[0]);
      }
    } catch (err) {
      console.error('Failed to list deployments:', err);
    } finally {
      setIsLoading(false);
    }
  }, [selectedDeployment]);

  useEffect(() => {
    fetchDeployments();
    networkService.listNetworks().then(setAvailableNetworks).catch(console.error);
    import('@/services/server.service').then((m) =>
      m.serverService.listServers().then((res) => {
        if (res && res.data) setServers(res.data);
      }).catch(console.error)
    );
  }, [fetchDeployments]);

  // Load and stream logs for selected deployment
  useEffect(() => {
    if (!selectedDeployment) return;

    // Fetch initial logs
    deploymentService
      .getLogs(selectedDeployment.id, 200)
      .then((data) => setLogs(data))
      .catch(console.error);

    // Setup SSE streaming if deployment is active/running
    const apiBase =
      process.env.NEXT_PUBLIC_API_URL ||
      (typeof window !== 'undefined' ? `${window.location.protocol}//${window.location.hostname}:8080` : '');
    const sseUrl = `${apiBase}/api/v1/deployments/${selectedDeployment.id}/logs/stream`;

    let eventSource: EventSource | null = null;
    try {
      setIsStreaming(true);
      eventSource = new EventSource(sseUrl);

      eventSource.onmessage = (event) => {
        try {
          const newLog: DeploymentLog = JSON.parse(event.data);
          setLogs((prev) => {
            if (prev.some((l) => l.id === newLog.id)) return prev;
            return [...prev, newLog];
          });
        } catch (e) {
          console.error('Failed to parse log line', e);
        }
      };

      eventSource.addEventListener('complete', () => {
        setIsStreaming(false);
        eventSource?.close();
      });

      eventSource.onerror = () => {
        setIsStreaming(false);
        eventSource?.close();
      };
    } catch (err) {
      console.error('SSE initialization error:', err);
      setIsStreaming(false);
    }

    return () => {
      if (eventSource) {
        eventSource.close();
      }
      setIsStreaming(false);
    };
  }, [selectedDeployment]);

  // Auto-scroll terminal
  useEffect(() => {
    if (autoScroll && terminalEndRef.current) {
      terminalEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [logs, autoScroll]);

  const handleAddEnv = () => {
    if (envKey.trim()) {
      setEnvVars((prev) => ({ ...prev, [envKey.trim()]: envVal }));
      setEnvKey('');
      setEnvVal('');
    }
  };

  const handleRemoveEnv = (key: string) => {
    setEnvVars((prev) => {
      const copy = { ...prev };
      delete copy[key];
      return copy;
    });
  };

  const handleDeploy = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!appName.trim() || !imageTag.trim()) return;

    try {
      setIsSubmitting(true);
      const req: DeploymentRequest = {
        server_id: targetServerId ? targetServerId : undefined,
        app_name: appName.trim(),
        image_tag: imageTag.trim(),
        container_name: containerName.trim() || undefined,
        port_bindings: [
          {
            host_port: Number(hostPort),
            container_port: Number(containerPort),
            protocol: 'tcp',
          },
        ],
        environment_variables: envVars,
        restart_policy: restartPolicy,
        network_name: selectedNetwork || undefined,
      };

      const created = await deploymentService.createDeployment(req);
      setDeployments((prev) => [created, ...prev]);
      setSelectedDeployment(created);
      setIsModalOpen(false);
      // Reset form
      setAppName('');
      setImageTag('nginx:alpine');
      setContainerName('');
      setEnvVars({});
    } catch (err: any) {
      alert(`Deployment failed: ${err?.response?.data?.message || err.message}`);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleStopDeployment = async (id: string) => {
    if (!confirm('Are you sure you want to stop this container?')) return;
    try {
      await deploymentService.stopDeployment(id);
      fetchDeployments();
    } catch (err: any) {
      alert(`Failed to stop: ${err?.response?.data?.message || err.message}`);
    }
  };

  const handleRedeploy = async (id: string) => {
    if (!confirm('Are you sure you want to redeploy this container?')) return;
    try {
      const redeployed = await deploymentService.redeployDeployment(id);
      fetchDeployments();
      setSelectedDeployment(redeployed);
    } catch (err: any) {
      alert(`Redeploy failed: ${err?.response?.data?.message || err.message}`);
    }
  };

  const handleDeleteDeployment = async (id: string) => {
    if (!confirm('Are you sure you want to permanently delete this container and remove its container instance?')) return;
    try {
      await deploymentService.deleteDeployment(id);
      setDeployments((prev) => prev.filter((d) => d.id !== id));
      if (selectedDeployment?.id === id) {
        setSelectedDeployment(null);
        setLogs([]);
      }
      fetchDeployments();
    } catch (err: any) {
      alert(`Delete failed: ${err?.response?.data?.message || err.message}`);
    }
  };

  const handleRollback = async (id: string) => {
    if (!confirm('Are you sure you want to rollback this container?')) return;
    try {
      const rolled = await deploymentService.rollbackDeployment(id);
      fetchDeployments();
      setSelectedDeployment(rolled);
    } catch (err: any) {
      alert(`Rollback failed: ${err?.response?.data?.message || err.message}`);
    }
  };

  const handleCopyLogs = () => {
    const text = logs.map((l) => `[${l.timestamp}] [${l.stream}] ${l.message}`).join('\n');
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'running':
        return (
          <span className="flex items-center gap-1.5 px-2 py-0.5 text-[10px] font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 rounded">
            <span className="h-1.5 w-1.5 rounded-full bg-emerald-400"></span> RUNNING
          </span>
        );
      case 'pulling':
      case 'deploying':
      case 'building':
        return (
          <span className="flex items-center gap-1.5 px-2 py-0.5 text-[10px] font-semibold bg-sky-500/10 text-sky-400 border border-sky-500/30 rounded">
            <RefreshCw className="h-2.5 w-2.5 animate-spin" /> {status.toUpperCase()}
          </span>
        );
      case 'stopped':
        return (
          <span className="flex items-center gap-1.5 px-2 py-0.5 text-[10px] font-semibold bg-zinc-500/10 text-zinc-400 border border-zinc-500/30 rounded">
            STOPPED
          </span>
        );
      case 'failed':
        return (
          <span className="flex items-center gap-1.5 px-2 py-0.5 text-[10px] font-semibold bg-red-500/10 text-red-400 border border-red-500/30 rounded">
            FAILED
          </span>
        );
      default:
        return (
          <span className="flex items-center gap-1.5 px-2 py-0.5 text-[10px] font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/30 rounded">
            {status.toUpperCase()}
          </span>
        );
    }
  };

  const filteredLogs = logs.filter((l) => {
    if (streamFilter === 'all') return true;
    return l.stream === streamFilter;
  });

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-[#262626] pb-5">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-[#ededed]">Container Orchestration & Deployments</h1>
            <span className="px-2 py-0.5 text-[10px] font-medium bg-purple-500/10 text-purple-400 border border-purple-500/20 rounded">
              Phase 6.3 Active
            </span>
          </div>
          <p className="text-xs text-[#a1a1a1] mt-1">
            Build, pull, deploy, and monitor containerized applications with live streaming log terminals.
          </p>
        </div>

        <button
          onClick={() => setIsModalOpen(true)}
          className="flex items-center gap-1.5 px-4 py-2 text-xs font-semibold bg-emerald-500 hover:bg-emerald-400 text-zinc-950 rounded-lg shadow-sm transition-colors cursor-pointer"
        >
          <Plus className="h-3.5 w-3.5" />
          Deploy New Container
        </button>
      </div>

      {/* Main Grid: Containers List + Live Terminal */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left Column: Container Deployments (5 cols) */}
        <div className="lg:col-span-5 space-y-4">
          <div className="bg-[#171717] border border-[#262626] rounded-xl p-4 space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-xs font-bold text-[#ededed] uppercase tracking-wider">Active Deployments</span>
              <button onClick={fetchDeployments} className="text-[#707070] hover:text-[#ededed] p-1">
                <RefreshCw className={`h-3.5 w-3.5 ${isLoading ? 'animate-spin' : ''}`} />
              </button>
            </div>

            <div className="space-y-2 max-h-[600px] overflow-y-auto">
              {deployments.map((d) => (
                <div
                  key={d.id}
                  onClick={() => setSelectedDeployment(d)}
                  className={`p-3.5 rounded-xl border transition-all cursor-pointer ${
                    selectedDeployment?.id === d.id
                      ? 'bg-purple-950/10 border-purple-500/40 shadow-sm'
                      : 'bg-[#121212] border-[#262626] hover:border-[#383838]'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Box className="h-4 w-4 text-purple-400" />
                      <span className="text-xs font-bold text-[#ededed] font-mono">{d.app_name}</span>
                    </div>
                    {getStatusBadge(d.status)}
                  </div>

                  <div className="flex items-center justify-between mt-1.5">
                    <p className="text-[11px] font-mono text-[#a1a1a1] truncate">{d.image_tag}</p>
                    {d.network_name && (
                      <span className="px-1.5 py-0.5 text-[9px] font-mono bg-cyan-950/60 text-cyan-400 border border-cyan-500/30 rounded">
                        VPC: {d.network_name}
                      </span>
                    )}
                  </div>

                  <div className="mt-2 pt-2 border-t border-[#222222] flex items-center justify-between text-[10px] text-[#707070]">
                    <span>Port: {d.port_bindings?.[0]?.host_port || 80}:{d.port_bindings?.[0]?.container_port || 80}</span>
                    <div className="flex items-center gap-1.5">
                      {d.status === 'running' && (
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleStopDeployment(d.id);
                          }}
                          className="px-2 py-0.5 text-[10px] font-medium text-amber-400 hover:text-amber-300 bg-amber-500/10 hover:bg-amber-500/20 border border-amber-500/30 rounded transition-colors"
                        >
                          Stop
                        </button>
                      )}
                      {(d.status === 'stopped' || d.status === 'failed') && (
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleRedeploy(d.id);
                          }}
                          className="px-2 py-0.5 text-[10px] font-medium text-emerald-400 hover:text-emerald-300 bg-emerald-500/10 hover:bg-emerald-500/20 border border-emerald-500/30 rounded transition-colors"
                        >
                          Redeploy
                        </button>
                      )}
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDeleteDeployment(d.id);
                        }}
                        className="px-2 py-0.5 text-[10px] font-medium text-red-400 hover:text-red-300 bg-red-500/10 hover:bg-red-500/20 border border-red-500/30 rounded transition-colors flex items-center gap-1"
                        title="Delete container"
                      >
                        <Trash2 className="h-2.5 w-2.5" />
                        Delete
                      </button>
                    </div>
                  </div>
                </div>
              ))}

              {deployments.length === 0 && !isLoading && (
                <div className="p-8 text-center text-xs text-[#707070]">
                  No container deployments yet. Click &quot;Deploy New Container&quot; to begin.
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Right Column: Web-based Streaming Terminal (7 cols) */}
        <div className="lg:col-span-7 space-y-4">
          <div className="bg-[#0f1218] border border-[#22272e] rounded-xl overflow-hidden shadow-lg flex flex-col h-[650px]">
            {/* Terminal Header & Toolbar (Sleek 2-Tier Layout) */}
            <div className="bg-[#13171f] border-b border-[#22272e] flex flex-col">
              {/* Top Tier: Container Identity & Host Endpoint Link */}
              <div className="px-4 py-2.5 flex items-center justify-between border-b border-[#1c2128]">
                <div className="flex items-center gap-2.5">
                  <div className="flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-slate-800/80 border border-slate-700/60">
                    <Terminal className="h-3.5 w-3.5 text-emerald-400" />
                    <span className="text-xs font-mono font-semibold text-slate-200">
                      {selectedDeployment ? selectedDeployment.app_name : 'Terminal Logs'}
                    </span>
                  </div>
                  {selectedDeployment && (
                    <span className="text-[11px] font-mono text-slate-500 hidden sm:inline truncate max-w-[200px]">
                      {selectedDeployment.image_tag}
                    </span>
                  )}
                </div>

                <div className="flex items-center gap-2.5">
                  {selectedDeployment && selectedDeployment.port_bindings?.[0]?.host_port && (
                    <a
                      href={`http://localhost:${selectedDeployment.port_bindings[0].host_port}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="px-2.5 py-1 text-[11px] font-medium bg-purple-500/15 hover:bg-purple-500/25 text-purple-300 border border-purple-500/40 rounded-lg flex items-center gap-1.5 transition-all shadow-sm cursor-pointer"
                      title="Open Application in New Tab"
                    >
                      <span className="font-mono">localhost:{selectedDeployment.port_bindings[0].host_port}</span>
                      <ExternalLink className="h-3 w-3 text-purple-400" />
                    </a>
                  )}
                  {isStreaming && (
                    <span className="flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-emerald-950/60 border border-emerald-500/30 text-[10px] font-mono text-emerald-400">
                      <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
                      <span>Live SSE</span>
                    </span>
                  )}
                </div>
              </div>

              {/* Bottom Tier: Stream Filter Pills & Log Viewer Actions */}
              <div className="px-4 py-1.5 bg-[#0d1017] flex items-center justify-between text-xs">
                {/* Stream Filter Pills */}
                <div className="flex items-center gap-1">
                  {(['all', 'stdout', 'stderr', 'system'] as const).map((stream) => (
                    <button
                      key={stream}
                      type="button"
                      onClick={() => setStreamFilter(stream)}
                      className={`px-2 py-0.5 rounded text-[10px] font-mono uppercase tracking-wider transition-colors cursor-pointer ${
                        streamFilter === stream
                          ? 'bg-slate-800 text-emerald-400 font-semibold border border-slate-700 shadow-xs'
                          : 'text-slate-500 hover:text-slate-300 hover:bg-slate-850'
                      }`}
                    >
                      {stream}
                    </button>
                  ))}
                </div>

                {/* Toolbar Action Group */}
                <div className="flex items-center gap-1.5">
                  <span className="text-[10px] font-mono text-slate-500 mr-1.5">
                    {filteredLogs.length} lines
                  </span>

                  <button
                    type="button"
                    onClick={() => setAutoScroll(!autoScroll)}
                    className={`px-2 py-0.5 text-[10px] font-medium rounded border transition-colors cursor-pointer ${
                      autoScroll
                        ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30'
                        : 'text-slate-500 border-slate-800 hover:text-slate-300'
                    }`}
                    title="Auto-scroll terminal output"
                  >
                    Auto-Scroll
                  </button>

                  <button
                    type="button"
                    onClick={handleCopyLogs}
                    className="p-1 text-slate-500 hover:text-slate-200 border border-slate-800 hover:bg-slate-800/60 rounded transition-colors cursor-pointer"
                    title="Copy logs to clipboard"
                  >
                    {copied ? <Check className="h-3 w-3 text-emerald-400" /> : <Copy className="h-3 w-3" />}
                  </button>

                  <button
                    type="button"
                    onClick={() => setLogs([])}
                    className="px-2 py-0.5 text-[10px] font-medium text-slate-500 hover:text-slate-200 border border-slate-800 hover:bg-slate-800/60 rounded transition-colors cursor-pointer"
                    title="Clear view"
                  >
                    Clear View
                  </button>
                </div>
              </div>
            </div>

            {/* Terminal Body */}
            <div className="flex-1 p-4 overflow-y-auto font-mono text-xs text-[#c9d1d9] space-y-1 selection:bg-emerald-500/30">
              {filteredLogs.map((log) => {
                let streamColor = 'text-emerald-400';
                if (log.stream === 'stderr') streamColor = 'text-red-400';
                if (log.stream === 'system') streamColor = 'text-purple-400';

                return (
                  <div key={log.id} className="leading-relaxed hover:bg-[#161b22] px-1 rounded flex items-start gap-2">
                    <span className="text-[10px] text-[#545d68] flex-shrink-0 select-none">
                      {new Date(log.timestamp).toLocaleTimeString()}
                    </span>
                    <span className={`text-[10px] font-bold uppercase flex-shrink-0 select-none ${streamColor}`}>
                      [{log.stream}]
                    </span>
                    <span className="break-all whitespace-pre-wrap">{log.message}</span>
                  </div>
                );
              })}

              {filteredLogs.length === 0 && (
                <div className="text-center text-[#545d68] pt-24 select-none">
                  No log entries available for this container.
                </div>
              )}
              <div ref={terminalEndRef} />
            </div>

            {/* Terminal Status Footer */}
            <div className="px-4 py-1.5 bg-[#161b22] border-t border-[#22272e] flex items-center justify-between text-[10px] text-[#707070] font-mono">
              <span>Lines: {filteredLogs.length}</span>
              <span>Container ID: {selectedDeployment ? selectedDeployment.id.slice(0, 8) : 'N/A'}</span>
            </div>
          </div>
        </div>
      </div>

      {/* Deploy Container Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
          <div className="bg-[#171717] border border-[#262626] rounded-xl w-full max-w-xl p-6 space-y-5 shadow-2xl">
            <div className="flex items-center justify-between border-b border-[#262626] pb-3">
              <div className="flex items-center gap-2">
                <Box className="h-5 w-5 text-purple-400" />
                <h2 className="text-sm font-bold text-[#ededed]">Deploy Docker Container</h2>
              </div>
              <button onClick={() => setIsModalOpen(false)} className="text-[#707070] hover:text-[#ededed]">
                &times;
              </button>
            </div>

            <form onSubmit={handleDeploy} className="space-y-4">
              {/* Target Server / Host Node */}
              <div>
                <label className="text-[11px] font-medium text-[#a1a1a1]">Target Server / Host Node</label>
                <select
                  value={targetServerId}
                  onChange={(e) => setTargetServerId(e.target.value)}
                  className="w-full bg-[#121212] border border-[#2e2e2e] rounded-lg px-3 py-1.5 text-xs text-[#ededed] focus:outline-none focus:border-purple-500 font-mono mt-1"
                >
                  <option value="">Local Host (Current Machine)</option>
                  {servers.map((s: any) => (
                    <option key={s.id} value={s.id}>
                      {s.name} ({s.ip_address || s.ipAddress || "Agent Node"} - {s.status})
                    </option>
                  ))}
                </select>
              </div>

              {/* Quick Image Selector */}
              <div className="space-y-1.5">
                <label className="text-[11px] font-medium text-[#a1a1a1]">Quick Image Presets</label>
                <div className="grid grid-cols-2 gap-2">
                  {QUICK_IMAGES.map((img) => (
                    <button
                      key={img.tag}
                      type="button"
                      onClick={() => {
                        setImageTag(img.tag);
                        setContainerPort(img.port);
                        if (!appName) setAppName(img.name.toLowerCase().replace(/\s+/g, '-'));
                      }}
                      className="p-2 text-left bg-[#121212] hover:bg-[#1f1f1f] border border-[#262626] hover:border-purple-500/40 rounded-lg text-xs transition-colors cursor-pointer"
                    >
                      <p className="font-semibold text-[#ededed]">{img.name}</p>
                      <p className="text-[10px] text-[#707070] font-mono">{img.tag}</p>
                    </button>
                  ))}
                </div>
              </div>

              {/* App Name & Container Name */}
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-[11px] font-medium text-[#a1a1a1]">App Name *</label>
                  <input
                    type="text"
                    required
                    value={appName}
                    onChange={(e) => setAppName(e.target.value)}
                    placeholder="e.g. backend-api"
                    className="w-full bg-[#121212] border border-[#2e2e2e] rounded-lg px-3 py-1.5 text-xs text-[#ededed] focus:outline-none focus:border-purple-500"
                  />
                </div>
                <div>
                  <label className="text-[11px] font-medium text-[#a1a1a1]">Docker Image *</label>
                  <input
                    type="text"
                    required
                    value={imageTag}
                    onChange={(e) => setImageTag(e.target.value)}
                    placeholder="e.g. nginx:alpine"
                    className="w-full bg-[#121212] border border-[#2e2e2e] rounded-lg px-3 py-1.5 text-xs text-[#ededed] font-mono focus:outline-none focus:border-purple-500"
                  />
                </div>
              </div>

              {/* Port Mappings */}
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-[11px] font-medium text-[#a1a1a1]">Host Port</label>
                  <input
                    type="number"
                    value={hostPort}
                    onChange={(e) => setHostPort(Number(e.target.value))}
                    className="w-full bg-[#121212] border border-[#2e2e2e] rounded-lg px-3 py-1.5 text-xs text-[#ededed] font-mono focus:outline-none focus:border-purple-500"
                  />
                </div>
                <div>
                  <label className="text-[11px] font-medium text-[#a1a1a1]">Container Port</label>
                  <input
                    type="number"
                    value={containerPort}
                    onChange={(e) => setContainerPort(Number(e.target.value))}
                    className="w-full bg-[#121212] border border-[#2e2e2e] rounded-lg px-3 py-1.5 text-xs text-[#ededed] font-mono focus:outline-none focus:border-purple-500"
                  />
                </div>
              </div>

              {/* VPC Network Selector */}
              <div>
                <label className="text-[11px] font-medium text-[#a1a1a1]">VPC Network / Isolation</label>
                <select
                  value={selectedNetwork}
                  onChange={(e) => setSelectedNetwork(e.target.value)}
                  className="w-full bg-[#121212] border border-[#2e2e2e] rounded-lg px-3 py-1.5 text-xs text-[#ededed] focus:outline-none focus:border-purple-500"
                >
                  <option value="">Default Host Bridge (caelus-network)</option>
                  {availableNetworks.map((net) => (
                    <option key={net.id} value={net.name}>
                      {net.name} ({net.cidr} - {net.region})
                    </option>
                  ))}
                </select>
              </div>

              {/* Environment Variables */}
              <div className="space-y-2">
                <label className="text-[11px] font-medium text-[#a1a1a1]">Environment Variables</label>
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={envKey}
                    onChange={(e) => setEnvKey(e.target.value)}
                    placeholder="KEY (e.g. NODE_ENV)"
                    className="flex-1 bg-[#121212] border border-[#2e2e2e] rounded-lg px-3 py-1 text-xs text-[#ededed] font-mono"
                  />
                  <input
                    type="text"
                    value={envVal}
                    onChange={(e) => setEnvVal(e.target.value)}
                    placeholder="VALUE"
                    className="flex-1 bg-[#121212] border border-[#2e2e2e] rounded-lg px-3 py-1 text-xs text-[#ededed] font-mono"
                  />
                  <button
                    type="button"
                    onClick={handleAddEnv}
                    className="px-3 py-1 text-xs bg-zinc-800 hover:bg-zinc-700 text-[#ededed] rounded-lg border border-zinc-700 cursor-pointer"
                  >
                    Add
                  </button>
                </div>

                {Object.keys(envVars).length > 0 && (
                  <div className="flex flex-wrap gap-1.5 pt-1">
                    {Object.entries(envVars).map(([k, v]) => (
                      <span
                        key={k}
                        className="inline-flex items-center gap-1.5 px-2 py-0.5 text-[10px] font-mono bg-[#1f1f1f] text-emerald-400 border border-[#333333] rounded"
                      >
                        {k}={v}
                        <button type="button" onClick={() => handleRemoveEnv(k)} className="hover:text-red-400">
                          &times;
                        </button>
                      </span>
                    ))}
                  </div>
                )}
              </div>

              {/* Actions */}
              <div className="flex items-center justify-end gap-3 pt-3 border-t border-[#262626]">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="px-4 py-1.5 text-xs text-[#a1a1a1] hover:text-[#ededed]"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isSubmitting}
                  className="px-4 py-1.5 text-xs font-semibold bg-emerald-500 hover:bg-emerald-400 text-zinc-950 rounded-lg shadow cursor-pointer disabled:opacity-50"
                >
                  {isSubmitting ? 'Deploying...' : 'Launch Container'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
