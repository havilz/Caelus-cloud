"use client";

import React, { useState } from "react";
import { X, ShieldAlert, Check } from "lucide-react";
import { Button } from "@/components/ui/button";
import { monitoringService } from "@/services/monitoring.service";

interface CreateAlertRuleModalProps {
  isOpen: boolean;
  onClose: () => void;
  onRuleCreated: () => void;
  serverId?: string;
}

export const CreateAlertRuleModal: React.FC<CreateAlertRuleModalProps> = ({
  isOpen,
  onClose,
  onRuleCreated,
  serverId,
}) => {
  const [name, setName] = useState("");
  const [metricName, setMetricName] = useState("cpu_usage");
  const [operator, setOperator] = useState(">");
  const [threshold, setThreshold] = useState("85");
  const [severity, setSeverity] = useState("warning");
  const [durationSeconds, setDurationSeconds] = useState("60");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name) {
      setError("Nama aturan alert wajib diisi.");
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      await monitoringService.createAlertRule({
        server_id: serverId || undefined,
        name,
        metric_name: metricName,
        operator,
        threshold: parseFloat(threshold),
        duration_seconds: parseInt(durationSeconds, 10),
        severity,
      });
      onRuleCreated();
      onClose();
    } catch (err: any) {
      setError(err.response?.data?.message || "Gagal membuat aturan alert.");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div
        className="fixed inset-0 bg-black/70 backdrop-blur-xs transition-opacity"
        onClick={onClose}
      />

      <div className="relative w-full max-w-md rounded-2xl border border-[#2e2e2e] bg-[#171717] p-6 shadow-2xl z-10 animate-in fade-in zoom-in-95 duration-150">
        <div className="flex items-center justify-between pb-4 border-b border-[#262626]">
          <div className="flex items-center gap-2.5">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              <ShieldAlert className="h-4 w-4" />
            </div>
            <div>
              <h3 className="text-sm font-semibold text-[#ededed]">Buat Aturan Alert Baru</h3>
              <p className="text-[11px] text-[#a1a1a1]">Konfigurasi ambang batas evaluasi otomatis</p>
            </div>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="text-[#707070] hover:text-[#ededed] transition-colors"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        {error && (
          <div className="mt-4 p-2.5 rounded-lg bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="mt-4 space-y-4 text-xs">
          <div>
            <label className="block text-[#a1a1a1] mb-1 font-medium">Nama Aturan</label>
            <input
              type="text"
              required
              placeholder="Contoh: Beban CPU Kritis Node 1"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full rounded-lg border border-[#262626] bg-[#121212] px-3 py-2 text-[#ededed] placeholder-[#666666] focus:border-emerald-500 focus:outline-none"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-[#a1a1a1] mb-1 font-medium">Target Metrik</label>
              <select
                value={metricName}
                onChange={(e) => setMetricName(e.target.value)}
                className="w-full rounded-lg border border-[#262626] bg-[#121212] px-3 py-2 text-[#ededed] focus:border-emerald-500 focus:outline-none cursor-pointer"
              >
                <option value="cpu_usage">CPU Usage (%)</option>
                <option value="memory_usage">RAM Memory (%)</option>
                <option value="disk_usage">SSD Disk (%)</option>
              </select>
            </div>

            <div>
              <label className="block text-[#a1a1a1] mb-1 font-medium">Operator & Ambang Batas</label>
              <div className="flex gap-2">
                <select
                  value={operator}
                  onChange={(e) => setOperator(e.target.value)}
                  className="w-16 rounded-lg border border-[#262626] bg-[#121212] px-2 py-2 text-[#ededed] focus:border-emerald-500 focus:outline-none cursor-pointer text-center font-bold"
                >
                  <option value=">">&gt;</option>
                  <option value=">=">&gt;=</option>
                  <option value="<">&lt;</option>
                  <option value="<=">&lt;=</option>
                </select>
                <input
                  type="number"
                  required
                  min="1"
                  max="100"
                  value={threshold}
                  onChange={(e) => setThreshold(e.target.value)}
                  className="flex-1 rounded-lg border border-[#262626] bg-[#121212] px-3 py-2 text-[#ededed] focus:border-emerald-500 focus:outline-none font-mono"
                />
              </div>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-[#a1a1a1] mb-1 font-medium">Tingkat Keparahan (Severity)</label>
              <select
                value={severity}
                onChange={(e) => setSeverity(e.target.value)}
                className="w-full rounded-lg border border-[#262626] bg-[#121212] px-3 py-2 text-[#ededed] focus:border-emerald-500 focus:outline-none cursor-pointer"
              >
                <option value="warning">Warning (Peringatan)</option>
                <option value="critical">Critical (Kritis)</option>
                <option value="info">Info (Informasi)</option>
              </select>
            </div>

            <div>
              <label className="block text-[#a1a1a1] mb-1 font-medium">Durasi Pemicu (Detik)</label>
              <input
                type="number"
                required
                min="10"
                step="10"
                value={durationSeconds}
                onChange={(e) => setDurationSeconds(e.target.value)}
                className="w-full rounded-lg border border-[#262626] bg-[#121212] px-3 py-2 text-[#ededed] focus:border-emerald-500 focus:outline-none font-mono"
              />
            </div>
          </div>

          <div className="pt-4 border-t border-[#262626] flex items-center justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={onClose}
              className="text-[#a1a1a1]"
            >
              Batal
            </Button>
            <Button
              type="submit"
              size="sm"
              disabled={isLoading}
              className="bg-emerald-500 text-black hover:bg-emerald-400 font-semibold"
            >
              {isLoading ? "Menyimpan..." : "Buat Aturan"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
};
