import { apiClient } from './api';
import { APIResponse } from '@/types/api';
import { Deployment, DeploymentLog, DeploymentRequest } from '@/types/deployment';

export const deploymentService = {
  async createDeployment(data: DeploymentRequest): Promise<Deployment> {
    const response = await apiClient.post<APIResponse<Deployment>>('/deployments', data);
    return response.data.data;
  },

  async listDeployments(serverId?: string): Promise<Deployment[]> {
    const params = serverId ? { server_id: serverId } : {};
    const response = await apiClient.get<APIResponse<Deployment[]>>('/deployments', { params });
    return response.data.data || [];
  },

  async getDeployment(id: string): Promise<Deployment> {
    const response = await apiClient.get<APIResponse<Deployment>>(`/deployments/${id}`);
    return response.data.data;
  },

  async getLogs(id: string, limit = 200): Promise<DeploymentLog[]> {
    const response = await apiClient.get<APIResponse<DeploymentLog[]>>(`/deployments/${id}/logs`, {
      params: { limit },
    });
    return response.data.data || [];
  },

  async stopDeployment(id: string): Promise<void> {
    await apiClient.post(`/deployments/${id}/stop`);
  },

  async rollbackDeployment(id: string): Promise<Deployment> {
    const response = await apiClient.post<APIResponse<Deployment>>(`/deployments/${id}/rollback`);
    return response.data.data;
  },
};
