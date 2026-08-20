import { apiClient } from "./api";
import { ServerMetric, Alert, AlertRule } from "@/types/monitoring";
import { APIResponse, PaginatedResponse } from "@/types/api";

export const monitoringService = {
  // Mengambil rekaman snapshot metrik telemetri terbaru untuk server tertentu.
  async getLiveMetrics(serverId: string): Promise<ServerMetric> {
    const res = await apiClient.get<APIResponse<ServerMetric>>(`/servers/${serverId}/metrics/live`);
    return res.data.data;
  },

  // Mengambil riwayat deret waktu metrik server dalam rentang waktu tertentu (1h, 6h, 24h, 7d).
  async getMetricHistory(serverId: string, duration = "1h"): Promise<ServerMetric[]> {
    const res = await apiClient.get<APIResponse<ServerMetric[]>>(
      `/servers/${serverId}/metrics/history?duration=${duration}`
    );
    return res.data.data || [];
  },

  // Mengambil daftar insiden alert organisasi dengan paginasi dan filter status.
  async listAlerts(
    status?: string,
    page = 1,
    limit = 20
  ): Promise<PaginatedResponse<Alert>> {
    const params = new URLSearchParams();
    if (status && status !== "all") {
      params.append("status", status);
    }
    params.append("page", page.toString());
    params.append("limit", limit.toString());

    const res = await apiClient.get<PaginatedResponse<Alert>>(`/alerts?${params.toString()}`);
    return res.data;
  },

  // Mengubah status alert menjadi Acknowledged (telah ditinjau).
  async acknowledgeAlert(alertId: string): Promise<void> {
    await apiClient.post(`/alerts/${alertId}/acknowledge`);
  },

  // Mengubah status alert menjadi Resolved (telah terselesaikan).
  async resolveAlert(alertId: string): Promise<void> {
    await apiClient.post(`/alerts/${alertId}/resolve`);
  },

  // Mengambil seluruh aturan ambang batas alert milik organisasi.
  async listAlertRules(): Promise<AlertRule[]> {
    const res = await apiClient.get<APIResponse<AlertRule[]>>("/alerts/rules");
    return res.data.data || [];
  },

  // Membuat aturan threshold alert baru.
  async createAlertRule(data: {
    server_id?: string;
    name: string;
    metric_name: string;
    operator: string;
    threshold: number;
    duration_seconds?: number;
    severity?: string;
  }): Promise<AlertRule> {
    const res = await apiClient.post<APIResponse<AlertRule>>("/alerts/rules", data);
    return res.data.data;
  },

  // Menghapus aturan alert berdasarkan ID.
  async deleteAlertRule(ruleId: string): Promise<void> {
    await apiClient.delete(`/alerts/rules/${ruleId}`);
  },
};
