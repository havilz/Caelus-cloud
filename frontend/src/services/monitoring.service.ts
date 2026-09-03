import { apiClient } from "./api";
import { ServerMetric, Alert, AlertRule } from "@/types/monitoring";
import { APIResponse, PaginatedResponse } from "@/types/api";

export const monitoringService = {
  
  async getLiveMetrics(serverId: string): Promise<ServerMetric> {
    const res = await apiClient.get<APIResponse<ServerMetric>>(`/servers/${serverId}/metrics/live`);
    return res.data.data;
  },

  async getMetricHistory(serverId: string, duration = "1h"): Promise<ServerMetric[]> {
    const res = await apiClient.get<APIResponse<ServerMetric[]>>(
      `/servers/${serverId}/metrics/history?duration=${duration}`
    );
    return res.data.data || [];
  },

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

  async acknowledgeAlert(alertId: string): Promise<void> {
    await apiClient.post(`/alerts/${alertId}/acknowledge`);
  },

  async resolveAlert(alertId: string): Promise<void> {
    await apiClient.post(`/alerts/${alertId}/resolve`);
  },

  async listAlertRules(): Promise<AlertRule[]> {
    const res = await apiClient.get<APIResponse<AlertRule[]>>("/alerts/rules");
    return res.data.data || [];
  },

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

  async deleteAlertRule(ruleId: string): Promise<void> {
    await apiClient.delete(`/alerts/rules/${ruleId}`);
  },
};
