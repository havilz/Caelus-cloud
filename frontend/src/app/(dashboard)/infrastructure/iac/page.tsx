'use client';

import React, { useState, useEffect, useCallback } from 'react';
import {
  Code2,
  Play,
  RotateCcw,
  CheckCircle2,
  AlertTriangle,
  FileCode,
  Layers,
  History,
  Plus,
  Trash2,
  Save,
  Check,
  X,
  FileDiff,
  ChevronRight,
  RefreshCw,
  Server,
  Database,
  Box,
  Zap,
} from 'lucide-react';
import { AppTheme } from '@/core/theme';
import {
  IaCConfiguration,
  IaCPlan,
  IaCState,
  IaCValidationError,
} from '@/types/iac';
import { iacService } from '@/services/iac.service';

const STARTER_TEMPLATES = [
  {
    name: 'Fullstack App & Database',
    description: '1 VPS Web App + 1 MinIO S3 Object Storage + 1 Redis Container',
    yaml: `version: v1

servers:
  - name: production-app-server
    provider: aws
    region: us-east-1
    size: t3.micro
    image: ubuntu-22.04
    tags:
      environment: production
      role: api-backend

storages:
  - name: app-media-bucket
    type: s3
    region: us-east-1
    versioning: true
    access: private

containers:
  - name: redis-cache
    image: redis:7-alpine
    ports:
      - "6379:6379"
    restart_policy: unless-stopped
`,
  },
  {
    name: 'Multi-Cloud HA Cluster',
    description: 'Multi-provider servers across AWS EC2 and DigitalOcean',
    yaml: `version: v1

servers:
  - name: aws-primary-node
    provider: aws
    region: us-east-1
    size: t3.medium
    image: ubuntu-22.04
  - name: do-secondary-node
    provider: digitalocean
    region: nyc1
    size: s-1vcpu-1gb
    image: debian-12

storages:
  - name: disaster-recovery-backup
    type: r2
    region: global
    versioning: true
`,
  },
  {
    name: 'Microservice Containers & Automation',
    description: 'Containerized API gateway, frontend, and alert automation rule',
    yaml: `version: v1

containers:
  - name: api-gateway
    image: traefik:v2.10
    ports:
      - "80:80"
      - "443:443"
  - name: frontend-service
    image: nginx:alpine
    ports:
      - "8080:80"

rules:
  - name: auto-reboot-on-high-cpu
    trigger: cpu_exceeded
    condition:
      threshold: 90
      duration_seconds: 60
    action:
      type: restart_service
`,
  },
];

export default function DeclarativeIaCPage() {
  const [configs, setConfigs] = useState<IaCConfiguration[]>([]);
  const [selectedConfig, setSelectedConfig] = useState<IaCConfiguration | null>(null);
  const [rawYAML, setRawYAML] = useState<string>(STARTER_TEMPLATES[0].yaml);
  const [configName, setConfigName] = useState<string>('production-stack');
  const [configDesc, setConfigDesc] = useState<string>('Primary cloud infrastructure stack');
  
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isSaving, setIsSaving] = useState<boolean>(false);
  const [isPlanning, setIsPlanning] = useState<boolean>(false);
  const [isApplying, setIsApplying] = useState<boolean>(false);
  const [validationErrors, setValidationErrors] = useState<IaCValidationError[]>([]);
  const [isValidating, setIsValidating] = useState<boolean>(false);
  const [currentPlan, setCurrentPlan] = useState<IaCPlan | null>(null);
  const [statesHistory, setStatesHistory] = useState<IaCState[]>([]);
  const [activeTab, setActiveTab] = useState<'editor' | 'diff' | 'history'>('editor');
  const [applySuccessMessage, setApplySuccessMessage] = useState<string | null>(null);

  const fetchConfigs = useCallback(async () => {
    try {
      setIsLoading(true);
      const data = await iacService.listConfigs();
      setConfigs(data);
      if (data.length > 0 && !selectedConfig) {
        setSelectedConfig(data[0]);
        setRawYAML(data[0].raw_yaml);
        setConfigName(data[0].name);
        setConfigDesc(data[0].description || '');
      }
    } catch (err) {
      console.error('Failed to load IaC configurations:', err);
    } finally {
      setIsLoading(false);
    }
  }, [selectedConfig]);

  useEffect(() => {
    fetchConfigs();
  }, [fetchConfigs]);

  // Real-time YAML validation
  useEffect(() => {
    const timer = setTimeout(async () => {
      if (!rawYAML.trim()) {
        setValidationErrors([]);
        return;
      }
      setIsValidating(true);
      try {
        const res = await iacService.validateYAML(rawYAML);
        if (!res.valid && res.errors) {
          setValidationErrors(res.errors);
        } else {
          setValidationErrors([]);
        }
      } catch (err) {
        console.error('Validation error:', err);
      } finally {
        setIsValidating(false);
      }
    }, 400);

    return () => clearTimeout(timer);
  }, [rawYAML]);

  // Load state history when tab changes
  useEffect(() => {
    if (selectedConfig && activeTab === 'history') {
      iacService.listStates(selectedConfig.id).then(setStatesHistory).catch(console.error);
    }
  }, [selectedConfig, activeTab]);

  const handleSelectConfig = (cfg: IaCConfiguration) => {
    setSelectedConfig(cfg);
    setRawYAML(cfg.raw_yaml);
    setConfigName(cfg.name);
    setConfigDesc(cfg.description || '');
    setCurrentPlan(null);
    setApplySuccessMessage(null);
  };

  const handleSaveConfig = async () => {
    try {
      setIsSaving(true);
      if (selectedConfig) {
        const updated = await iacService.updateConfig(selectedConfig.id, {
          name: configName,
          description: configDesc,
          raw_yaml: rawYAML,
        });
        setSelectedConfig(updated);
        setConfigs((prev) => prev.map((c) => (c.id === updated.id ? updated : c)));
      } else {
        const created = await iacService.createConfig({
          name: configName,
          description: configDesc,
          raw_yaml: rawYAML,
        });
        setSelectedConfig(created);
        setConfigs((prev) => [created, ...prev]);
      }
    } catch (err: any) {
      alert(`Failed to save: ${err?.response?.data?.message || err.message}`);
    } finally {
      setIsSaving(false);
    }
  };

  const handleGeneratePlan = async () => {
    if (!selectedConfig) {
      await handleSaveConfig();
    }
    if (!selectedConfig) return;

    try {
      setIsPlanning(true);
      const plan = await iacService.generatePlan(selectedConfig.id);
      setCurrentPlan(plan);
      setActiveTab('diff');
    } catch (err: any) {
      alert(`Plan error: ${err?.response?.data?.message || err.message}`);
    } finally {
      setIsPlanning(false);
    }
  };

  const handleApplyPlan = async () => {
    if (!currentPlan) return;
    try {
      setIsApplying(true);
      const state = await iacService.applyPlan(currentPlan.id);
      setApplySuccessMessage(`Successfully applied state Version ${state.version}!`);
      setCurrentPlan(null);
      if (selectedConfig) {
        const updated = await iacService.getConfig(selectedConfig.id);
        setSelectedConfig(updated);
        setConfigs((prev) => prev.map((c) => (c.id === updated.id ? updated : c)));
      }
    } catch (err: any) {
      alert(`Apply failed: ${err?.response?.data?.message || err.message}`);
    } finally {
      setIsApplying(false);
    }
  };

  const handleRollback = async (version: number) => {
    if (!selectedConfig) return;
    if (!confirm(`Are you sure you want to rollback to state Version ${version}?`)) return;

    try {
      setIsApplying(true);
      const restored = await iacService.rollbackState(selectedConfig.id, version);
      alert(`State successfully rolled back to Version ${restored.version} (mirrored from v${version})`);
      const updated = await iacService.getConfig(selectedConfig.id);
      setSelectedConfig(updated);
      const history = await iacService.listStates(selectedConfig.id);
      setStatesHistory(history);
    } catch (err: any) {
      alert(`Rollback failed: ${err?.response?.data?.message || err.message}`);
    } finally {
      setIsApplying(false);
    }
  };

  const handleLoadTemplate = (tpl: (typeof STARTER_TEMPLATES)[0]) => {
    if (confirm(`Load template "${tpl.name}"? This will overwrite the current editor content.`)) {
      setRawYAML(tpl.yaml);
      setConfigName(tpl.name.toLowerCase().replace(/\s+/g, '-'));
      setConfigDesc(tpl.description);
    }
  };

  const handleCreateNew = () => {
    setSelectedConfig(null);
    setRawYAML(STARTER_TEMPLATES[0].yaml);
    setConfigName('new-infrastructure-stack');
    setConfigDesc('Declarative configuration');
    setCurrentPlan(null);
    setApplySuccessMessage(null);
  };

  const getActionBadge = (action: string) => {
    switch (action) {
      case 'create':
        return <span className="px-2 py-0.5 text-[10px] font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 rounded">+ CREATE</span>;
      case 'update':
        return <span className="px-2 py-0.5 text-[10px] font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/30 rounded">~ UPDATE</span>;
      case 'delete':
        return <span className="px-2 py-0.5 text-[10px] font-semibold bg-red-500/10 text-red-400 border border-red-500/30 rounded">- DELETE</span>;
      default:
        return <span className="px-2 py-0.5 text-[10px] font-semibold bg-zinc-500/10 text-zinc-400 border border-zinc-500/30 rounded">= NOOP</span>;
    }
  };

  const getResourceIcon = (type: string) => {
    switch (type) {
      case 'server':
        return <Server className="h-4 w-4 text-sky-400" />;
      case 'storage':
        return <Database className="h-4 w-4 text-emerald-400" />;
      case 'container':
        return <Box className="h-4 w-4 text-purple-400" />;
      case 'rule':
        return <Zap className="h-4 w-4 text-amber-400" />;
      default:
        return <Layers className="h-4 w-4 text-zinc-400" />;
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-[#262626] pb-5">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-[#ededed]">Declarative Infrastructure as Code (IaC)</h1>
            <span className="px-2 py-0.5 text-[10px] font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded">
              Phase 6.2 Active
            </span>
          </div>
          <p className="text-xs text-[#a1a1a1] mt-1">
            Define, plan, diff, and apply multi-cloud infrastructure declaratively with automatic rollback safety.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={handleCreateNew}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-[#1e1e1e] hover:bg-[#2a2a2a] text-[#ededed] border border-[#333333] rounded-lg transition-colors cursor-pointer"
          >
            <Plus className="h-3.5 w-3.5" />
            New Stack
          </button>
          <button
            onClick={handleSaveConfig}
            disabled={isSaving}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-zinc-800 hover:bg-zinc-700 text-[#ededed] border border-zinc-700 rounded-lg transition-colors cursor-pointer disabled:opacity-50"
          >
            <Save className="h-3.5 w-3.5" />
            {isSaving ? 'Saving...' : 'Save Draft'}
          </button>
          <button
            onClick={handleGeneratePlan}
            disabled={isPlanning || validationErrors.length > 0}
            className="flex items-center gap-1.5 px-4 py-1.5 text-xs font-semibold bg-emerald-500 hover:bg-emerald-400 text-zinc-950 rounded-lg shadow-sm transition-colors cursor-pointer disabled:opacity-50"
          >
            <Play className="h-3.5 w-3.5 fill-current" />
            {isPlanning ? 'Planning...' : 'Generate Plan'}
          </button>
        </div>
      </div>

      {/* Main Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* Left Sidebar: Stacks & Templates */}
        <div className="lg:col-span-1 space-y-4">
          {/* Stacks List */}
          <div className="bg-[#171717] border border-[#262626] rounded-xl p-4 space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-xs font-bold text-[#ededed] uppercase tracking-wider">Saved Stacks</span>
              <span className="text-[10px] text-[#707070]">{configs.length} available</span>
            </div>

            <div className="space-y-1.5 max-h-52 overflow-y-auto">
              {configs.map((c) => (
                <button
                  key={c.id}
                  onClick={() => handleSelectConfig(c)}
                  className={`w-full text-left p-2.5 rounded-lg text-xs transition-colors flex items-center justify-between cursor-pointer ${
                    selectedConfig?.id === c.id
                      ? 'bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 font-medium'
                      : 'bg-[#121212] border border-[#222222] text-[#a1a1a1] hover:text-[#ededed] hover:border-[#333333]'
                  }`}
                >
                  <div className="truncate mr-2">
                    <p className="font-mono truncate">{c.name}</p>
                    <p className="text-[10px] text-[#707070]">v{c.current_version} • {c.status}</p>
                  </div>
                  <ChevronRight className="h-3.5 w-3.5 flex-shrink-0 text-[#707070]" />
                </button>
              ))}
              {configs.length === 0 && !isLoading && (
                <p className="text-xs text-[#707070] italic">No saved configurations yet.</p>
              )}
            </div>
          </div>

          {/* Starter Templates */}
          <div className="bg-[#171717] border border-[#262626] rounded-xl p-4 space-y-3">
            <span className="text-xs font-bold text-[#ededed] uppercase tracking-wider">Starter Templates</span>
            <div className="space-y-2">
              {STARTER_TEMPLATES.map((tpl) => (
                <div
                  key={tpl.name}
                  onClick={() => handleLoadTemplate(tpl)}
                  className="p-2.5 bg-[#121212] hover:bg-[#1a1a1a] border border-[#222222] hover:border-emerald-500/40 rounded-lg cursor-pointer transition-colors"
                >
                  <p className="text-xs font-semibold text-[#ededed]">{tpl.name}</p>
                  <p className="text-[10px] text-[#707070] mt-0.5 line-clamp-2">{tpl.description}</p>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Right Area: Tabs & Workspaces */}
        <div className="lg:col-span-3 space-y-4">
          {/* Top Bar / Metadata */}
          <div className="bg-[#171717] border border-[#262626] rounded-xl p-4 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
            <div className="flex-1 w-full sm:w-auto">
              <input
                type="text"
                value={configName}
                onChange={(e) => setConfigName(e.target.value)}
                placeholder="Stack name (e.g. production-api)"
                className="w-full bg-[#121212] border border-[#2e2e2e] rounded-lg px-3 py-1.5 text-xs text-[#ededed] font-mono focus:outline-none focus:border-emerald-500"
              />
            </div>

            {/* View Tabs */}
            <div className="flex items-center bg-[#121212] p-1 border border-[#262626] rounded-lg">
              <button
                onClick={() => setActiveTab('editor')}
                className={`flex items-center gap-1.5 px-3 py-1 text-xs rounded-md font-medium transition-colors cursor-pointer ${
                  activeTab === 'editor' ? 'bg-[#222222] text-emerald-400' : 'text-[#707070] hover:text-[#ededed]'
                }`}
              >
                <Code2 className="h-3.5 w-3.5" />
                YAML Editor
              </button>
              <button
                onClick={() => setActiveTab('diff')}
                className={`flex items-center gap-1.5 px-3 py-1 text-xs rounded-md font-medium transition-colors cursor-pointer ${
                  activeTab === 'diff' ? 'bg-[#222222] text-emerald-400' : 'text-[#707070] hover:text-[#ededed]'
                }`}
              >
                <FileDiff className="h-3.5 w-3.5" />
                Visual Diff {currentPlan && `(${currentPlan.changes.length})`}
              </button>
              <button
                onClick={() => setActiveTab('history')}
                className={`flex items-center gap-1.5 px-3 py-1 text-xs rounded-md font-medium transition-colors cursor-pointer ${
                  activeTab === 'history' ? 'bg-[#222222] text-emerald-400' : 'text-[#707070] hover:text-[#ededed]'
                }`}
              >
                <History className="h-3.5 w-3.5" />
                State History
              </button>
            </div>
          </div>

          {applySuccessMessage && (
            <div className="p-3 bg-emerald-500/10 border border-emerald-500/30 rounded-xl flex items-center justify-between text-xs text-emerald-400">
              <div className="flex items-center gap-2">
                <CheckCircle2 className="h-4 w-4" />
                <span>{applySuccessMessage}</span>
              </div>
              <button onClick={() => setApplySuccessMessage(null)} className="text-emerald-400/60 hover:text-emerald-400">
                <X className="h-3.5 w-3.5" />
              </button>
            </div>
          )}

          {/* TAB 1: YAML EDITOR */}
          {activeTab === 'editor' && (
            <div className="bg-[#171717] border border-[#262626] rounded-xl overflow-hidden shadow-sm">
              <div className="flex items-center justify-between px-4 py-2.5 bg-[#121212] border-b border-[#262626]">
                <div className="flex items-center gap-2 text-xs text-[#a1a1a1]">
                  <FileCode className="h-4 w-4 text-emerald-400" />
                  <span className="font-mono">schema.yaml</span>
                </div>
                <div className="flex items-center gap-2 text-[10px]">
                  {isValidating ? (
                    <span className="text-[#707070] flex items-center gap-1">
                      <RefreshCw className="h-3 w-3 animate-spin" /> Validating...
                    </span>
                  ) : validationErrors.length > 0 ? (
                    <span className="text-red-400 flex items-center gap-1 font-semibold">
                      <AlertTriangle className="h-3 w-3" /> {validationErrors.length} syntax error(s)
                    </span>
                  ) : (
                    <span className="text-emerald-400 flex items-center gap-1 font-semibold">
                      <Check className="h-3 w-3" /> YAML Valid
                    </span>
                  )}
                </div>
              </div>

              <div className="relative">
                <textarea
                  value={rawYAML}
                  onChange={(e) => setRawYAML(e.target.value)}
                  placeholder="Define your declarative infrastructure..."
                  rows={20}
                  className="w-full bg-[#101010] p-4 text-xs font-mono text-[#ededed] focus:outline-none resize-y selection:bg-emerald-500/30"
                  spellCheck={false}
                />
              </div>

              {validationErrors.length > 0 && (
                <div className="p-3 bg-red-950/20 border-t border-red-900/30 text-xs text-red-400 space-y-1">
                  <p className="font-semibold flex items-center gap-1.5">
                    <AlertTriangle className="h-3.5 w-3.5" /> Validation Errors:
                  </p>
                  {validationErrors.map((err, idx) => (
                    <p key={idx} className="font-mono text-[11px] pl-5">
                      Line {err.line}: {err.message}
                    </p>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* TAB 2: VISUAL DIFF VIEWER */}
          {activeTab === 'diff' && (
            <div className="space-y-4">
              {currentPlan ? (
                <div className="space-y-4">
                  {/* Summary Cards */}
                  <div className="grid grid-cols-4 gap-3">
                    <div className="p-3 bg-[#171717] border border-emerald-500/20 rounded-xl text-center">
                      <p className="text-[10px] text-[#707070] uppercase font-bold">To Create</p>
                      <p className="text-lg font-bold text-emerald-400">+{currentPlan.summary.create}</p>
                    </div>
                    <div className="p-3 bg-[#171717] border border-amber-500/20 rounded-xl text-center">
                      <p className="text-[10px] text-[#707070] uppercase font-bold">To Update</p>
                      <p className="text-lg font-bold text-amber-400">~{currentPlan.summary.update}</p>
                    </div>
                    <div className="p-3 bg-[#171717] border border-red-500/20 rounded-xl text-center">
                      <p className="text-[10px] text-[#707070] uppercase font-bold">To Delete</p>
                      <p className="text-lg font-bold text-red-400">-{currentPlan.summary.delete}</p>
                    </div>
                    <div className="p-3 bg-[#171717] border border-zinc-500/20 rounded-xl text-center">
                      <p className="text-[10px] text-[#707070] uppercase font-bold">No-op / Unchanged</p>
                      <p className="text-lg font-bold text-zinc-400">={currentPlan.summary.noop}</p>
                    </div>
                  </div>

                  {/* Changes List */}
                  <div className="bg-[#171717] border border-[#262626] rounded-xl overflow-hidden">
                    <div className="px-4 py-3 bg-[#121212] border-b border-[#262626] flex items-center justify-between">
                      <span className="text-xs font-bold text-[#ededed]">Planned Resource Changes</span>
                      <button
                        onClick={handleApplyPlan}
                        disabled={isApplying}
                        className="px-4 py-1.5 text-xs font-semibold bg-emerald-500 hover:bg-emerald-400 text-zinc-950 rounded-lg shadow transition-colors cursor-pointer disabled:opacity-50"
                      >
                        {isApplying ? 'Applying Plan...' : 'Apply This Plan'}
                      </button>
                    </div>

                    <div className="divide-y divide-[#222222]">
                      {currentPlan.changes.map((change, idx) => (
                        <div key={idx} className="p-4 space-y-2.5">
                          <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2">
                              {getResourceIcon(change.resource_type)}
                              <span className="text-xs font-bold text-[#ededed] font-mono">{change.resource_name}</span>
                              <span className="text-[10px] text-[#707070] uppercase">({change.resource_type})</span>
                            </div>
                            <div>{getActionBadge(change.action)}</div>
                          </div>

                          {change.reason && (
                            <p className="text-xs text-[#a1a1a1] italic">{change.reason}</p>
                          )}

                          {/* Attribute Diff Breakdown */}
                          {change.action === 'update' && (
                            <div className="bg-[#101010] p-3 rounded-lg border border-[#222222] font-mono text-[11px] space-y-1">
                              <p className="text-[#707070] font-bold">Modified Fields: {change.changed_fields?.join(', ')}</p>
                              <div className="grid grid-cols-2 gap-4 pt-2">
                                <div>
                                  <p className="text-red-400 font-semibold mb-1">- Before</p>
                                  <pre className="text-[10px] text-red-300/80 overflow-x-auto">{JSON.stringify(change.before, null, 2)}</pre>
                                </div>
                                <div>
                                  <p className="text-emerald-400 font-semibold mb-1">+ After</p>
                                  <pre className="text-[10px] text-emerald-300/80 overflow-x-auto">{JSON.stringify(change.after, null, 2)}</pre>
                                </div>
                              </div>
                            </div>
                          )}

                          {change.action === 'create' && change.after && (
                            <div className="bg-[#101010] p-3 rounded-lg border border-[#222222] font-mono text-[10px] text-emerald-300/80">
                              <pre className="overflow-x-auto">{JSON.stringify(change.after, null, 2)}</pre>
                            </div>
                          )}
                        </div>
                      ))}

                      {currentPlan.changes.length === 0 && (
                        <div className="p-8 text-center text-xs text-[#707070]">
                          No changes detected. Desired state matches current infrastructure.
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              ) : (
                <div className="bg-[#171717] border border-[#262626] rounded-xl p-12 text-center space-y-3">
                  <FileDiff className="h-8 w-8 text-[#707070] mx-auto" />
                  <p className="text-xs text-[#a1a1a1]">No execution plan generated yet.</p>
                  <button
                    onClick={handleGeneratePlan}
                    disabled={isPlanning}
                    className="px-4 py-1.5 text-xs font-semibold bg-emerald-500 hover:bg-emerald-400 text-zinc-950 rounded-lg cursor-pointer"
                  >
                    Generate Execution Plan
                  </button>
                </div>
              )}
            </div>
          )}

          {/* TAB 3: STATE HISTORY & ROLLBACK */}
          {activeTab === 'history' && (
            <div className="bg-[#171717] border border-[#262626] rounded-xl overflow-hidden">
              <div className="px-4 py-3 bg-[#121212] border-b border-[#262626]">
                <span className="text-xs font-bold text-[#ededed]">State Snapshot History</span>
              </div>

              <div className="divide-y divide-[#222222]">
                {statesHistory.map((st) => (
                  <div key={st.id} className="p-4 flex items-center justify-between gap-4">
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <span className="text-xs font-bold text-emerald-400 font-mono">Version {st.version}</span>
                        <span className="text-[10px] text-[#707070] font-mono">SHA: {st.hash.slice(0, 12)}...</span>
                      </div>
                      <p className="text-[10px] text-[#707070]">
                        Applied at: {new Date(st.applied_at).toLocaleString()}
                      </p>
                    </div>

                    <button
                      onClick={() => handleRollback(st.version)}
                      disabled={isApplying || selectedConfig?.current_version === st.version}
                      className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-[#1e1e1e] hover:bg-[#2a2a2a] text-[#ededed] border border-[#333333] rounded-lg transition-colors cursor-pointer disabled:opacity-30"
                    >
                      <RotateCcw className="h-3.5 w-3.5" />
                      Rollback to this
                    </button>
                  </div>
                ))}

                {statesHistory.length === 0 && (
                  <div className="p-8 text-center text-xs text-[#707070]">
                    No state history recorded yet. Apply a plan to create the first state version.
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
