import { apiClient } from './api';
import { APIResponse, PaginatedResponse } from '@/types/api';
import {
  SecurityPostureOverview,
  SecurityScan,
  SecurityFinding,
  SecurityIncident,
  TriggerScanInput,
  ListFindingsFilter,
  FindingStatus,
  IncidentStatus,
} from '@/types/security';

export const securityService = {
  // Posture Overview
  async getPostureOverview(): Promise<SecurityPostureOverview> {
    const res = await apiClient.get<APIResponse<SecurityPostureOverview>>('/security/overview');
    return res.data.data!;
  },

  // Scans
  async triggerScan(payload: TriggerScanInput): Promise<SecurityScan> {
    const res = await apiClient.post<APIResponse<SecurityScan>>('/security/scans', payload);
    return res.data.data!;
  },

  async listScans(page = 1, limit = 20, serverId?: string): Promise<PaginatedResponse<SecurityScan>> {
    const res = await apiClient.get<PaginatedResponse<SecurityScan>>('/security/scans', {
      params: { page, limit, server_id: serverId || undefined },
    });
    return res.data;
  },

  async getScan(id: string): Promise<SecurityScan> {
    const res = await apiClient.get<APIResponse<SecurityScan>>(`/security/scans/${id}`);
    return res.data.data!;
  },

  // Findings
  async listFindings(filter: ListFindingsFilter = {}): Promise<PaginatedResponse<SecurityFinding>> {
    const res = await apiClient.get<PaginatedResponse<SecurityFinding>>('/security/findings', {
      params: {
        page: filter.page || 1,
        limit: filter.limit || 20,
        server_id: filter.server_id || undefined,
        category: filter.category || undefined,
        severity: filter.severity || undefined,
        status: filter.status || undefined,
      },
    });
    return res.data;
  },

  async getFinding(id: string): Promise<SecurityFinding> {
    const res = await apiClient.get<APIResponse<SecurityFinding>>(`/security/findings/${id}`);
    return res.data.data!;
  },

  async updateFindingStatus(id: string, status: FindingStatus): Promise<void> {
    await apiClient.patch(`/security/findings/${id}/status`, { status });
  },

  // Incidents
  async listIncidents(page = 1, limit = 20, status?: IncidentStatus): Promise<PaginatedResponse<SecurityIncident>> {
    const res = await apiClient.get<PaginatedResponse<SecurityIncident>>('/security/incidents', {
      params: { page, limit, status: status || undefined },
    });
    return res.data;
  },

  async createIncident(payload: {
    title: string;
    summary?: string;
    severity?: string;
    finding_ids?: string[];
  }): Promise<SecurityIncident> {
    const res = await apiClient.post<APIResponse<SecurityIncident>>('/security/incidents', payload);
    return res.data.data!;
  },

  async updateIncidentStatus(id: string, status: IncidentStatus, notes?: string): Promise<void> {
    await apiClient.patch(`/security/incidents/${id}/status`, { status, notes });
  },
};
