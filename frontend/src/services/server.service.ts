import { apiClient } from "./api";
import { APIResponse, PaginatedResponse } from "@/types/api";
import { CreateServerDTO, ResizeServerDTO, Server } from "@/types/server";

export const serverService = {
  async listServers(page = 1, limit = 20): Promise<PaginatedResponse<Server>> {
    const response = await apiClient.get<PaginatedResponse<Server>>(`/servers?page=${page}&limit=${limit}`);
    return response.data;
  },

  async getServer(id: string): Promise<Server> {
    const response = await apiClient.get<APIResponse<Server>>(`/servers/${id}`);
    return response.data.data;
  },

  async createServer(data: CreateServerDTO): Promise<Server> {
    const response = await apiClient.post<APIResponse<Server>>("/servers", data);
    return response.data.data;
  },

  async rebootServer(id: string): Promise<void> {
    await apiClient.post(`/servers/${id}/reboot`);
  },

  async shutdownServer(id: string): Promise<void> {
    await apiClient.post(`/servers/${id}/shutdown`);
  },

  async startServer(id: string): Promise<void> {
    await apiClient.post(`/servers/${id}/start`);
  },

  async resizeServer(id: string, data: ResizeServerDTO): Promise<void> {
    await apiClient.patch(`/servers/${id}/resize`, data);
  },

  async deleteServer(id: string): Promise<void> {
    await apiClient.delete(`/servers/${id}`);
  },
};
