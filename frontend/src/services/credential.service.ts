import { apiClient } from "./api";
import { APIResponse } from "@/types/api";
import {
  Credential,
  CreateCredentialDTO,
  UpdateCredentialDTO,
  TestCredentialResult,
} from "@/types/credential";

export const credentialService = {
  async listCredentials(): Promise<Credential[]> {
    const response = await apiClient.get<APIResponse<Credential[]>>("/credentials");
    return response.data.data || [];
  },

  async getCredential(id: string): Promise<Credential> {
    const response = await apiClient.get<APIResponse<Credential>>(`/credentials/${id}`);
    return response.data.data;
  },

  async createCredential(data: CreateCredentialDTO): Promise<Credential> {
    const response = await apiClient.post<APIResponse<Credential>>("/credentials", data);
    return response.data.data;
  },

  async updateCredential(id: string, data: UpdateCredentialDTO): Promise<Credential> {
    const response = await apiClient.put<APIResponse<Credential>>(`/credentials/${id}`, data);
    return response.data.data;
  },

  async deleteCredential(id: string): Promise<void> {
    await apiClient.delete(`/credentials/${id}`);
  },

  async testCredential(id: string): Promise<TestCredentialResult> {
    const response = await apiClient.post<APIResponse<TestCredentialResult>>(`/credentials/${id}/test`);
    return response.data.data;
  },
};
