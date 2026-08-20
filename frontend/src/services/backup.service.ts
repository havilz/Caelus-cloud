import { apiClient } from './api';
import { APIResponse, PaginatedResponse } from '@/types/api';
import {
  BackupPolicy,
  BackupRecord,
  CreateBackupPolicyInput,
  TriggerBackupInput,
} from '@/types/backup';

export const backupService = {
  // 1. Policies
  async listPolicies(): Promise<BackupPolicy[]> {
    const res = await apiClient.get<APIResponse<BackupPolicy[]>>('/backups/policies');
    return res.data.data || [];
  },

  async createPolicy(input: CreateBackupPolicyInput): Promise<BackupPolicy> {
    const res = await apiClient.post<APIResponse<BackupPolicy>>('/backups/policies', input);
    return res.data.data!;
  },

  async deletePolicy(id: string): Promise<void> {
    await apiClient.delete(`/backups/policies/${id}`);
  },

  // 2. Records & Execution
  async triggerBackup(serverId: string, input?: TriggerBackupInput): Promise<BackupRecord> {
    const res = await apiClient.post<APIResponse<BackupRecord>>(
      `/backups/trigger/${serverId}`,
      input || {}
    );
    return res.data.data!;
  },

  async listRecords(page = 1, limit = 20): Promise<PaginatedResponse<BackupRecord>> {
    const res = await apiClient.get<PaginatedResponse<BackupRecord>>('/backups/records', {
      params: { page, limit },
    });
    return res.data;
  },

  async deleteRecord(id: string): Promise<void> {
    await apiClient.delete(`/backups/records/${id}`);
  },
};
