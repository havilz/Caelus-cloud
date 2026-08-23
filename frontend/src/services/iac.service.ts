import { apiClient } from './api';
import { APIResponse } from '@/types/api';
import {
  IaCConfiguration,
  IaCPlan,
  IaCState,
  IaCValidationResponse,
} from '@/types/iac';

export const iacService = {
  async validateYAML(rawYAML: string): Promise<IaCValidationResponse> {
    const response = await apiClient.post<APIResponse<IaCValidationResponse>>('/iac/validate', {
      raw_yaml: rawYAML,
    });
    return response.data.data;
  },

  async listConfigs(): Promise<IaCConfiguration[]> {
    const response = await apiClient.get<APIResponse<IaCConfiguration[]>>('/iac/configs');
    return response.data.data || [];
  },

  async getConfig(id: string): Promise<IaCConfiguration> {
    const response = await apiClient.get<APIResponse<IaCConfiguration>>(`/iac/configs/${id}`);
    return response.data.data;
  },

  async createConfig(data: { name: string; description?: string; raw_yaml: string }): Promise<IaCConfiguration> {
    const response = await apiClient.post<APIResponse<IaCConfiguration>>('/iac/configs', data);
    return response.data.data;
  },

  async updateConfig(id: string, data: { name?: string; description?: string; raw_yaml?: string }): Promise<IaCConfiguration> {
    const response = await apiClient.put<APIResponse<IaCConfiguration>>(`/iac/configs/${id}`, data);
    return response.data.data;
  },

  async deleteConfig(id: string): Promise<void> {
    await apiClient.delete(`/iac/configs/${id}`);
  },

  async generatePlan(id: string): Promise<IaCPlan> {
    const response = await apiClient.post<APIResponse<IaCPlan>>(`/iac/configs/${id}/plan`);
    return response.data.data;
  },

  async getLatestPlan(id: string): Promise<IaCPlan | null> {
    const response = await apiClient.get<APIResponse<IaCPlan>>(`/iac/configs/${id}/plan`);
    return response.data.data || null;
  },

  async applyPlan(planId: string): Promise<IaCState> {
    const response = await apiClient.post<APIResponse<IaCState>>(`/iac/plans/${planId}/apply`);
    return response.data.data;
  },

  async rollbackState(configId: string, targetVersion: number): Promise<IaCState> {
    const response = await apiClient.post<APIResponse<IaCState>>(`/iac/configs/${configId}/rollback`, {
      target_version: targetVersion,
    });
    return response.data.data;
  },

  async listStates(configId: string): Promise<IaCState[]> {
    const response = await apiClient.get<APIResponse<IaCState[]>>(`/iac/configs/${configId}/states`);
    return response.data.data || [];
  },
};
