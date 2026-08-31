import { apiClient } from "./api";
import { APIResponse, PaginatedResponse } from "@/types/api";
import {
  UserProfile,
  Organization,
  OrganizationMember,
  OrganizationInvitation,
  OrganizationRole,
  APIKey,
  Webhook,
  AuditLog,
} from "@/types/settings";

export const settingsService = {
  // Profile & Security
  async getProfile(): Promise<UserProfile> {
    const response = await apiClient.get<APIResponse<UserProfile>>("/settings/profile");
    return response.data.data;
  },

  async updateProfile(data: { full_name: string; avatar_url?: string }): Promise<UserProfile> {
    const response = await apiClient.put<APIResponse<UserProfile>>("/settings/profile", data);
    return response.data.data;
  },

  async changePassword(data: { old_password: string; new_password: string }): Promise<void> {
    await apiClient.post("/settings/change-password", data);
  },

  // Organization
  async getOrganization(): Promise<Organization> {
    const response = await apiClient.get<APIResponse<Organization>>("/settings/organization");
    return response.data.data;
  },

  async updateOrganization(data: { name: string; slug: string }): Promise<Organization> {
    const response = await apiClient.put<APIResponse<Organization>>("/settings/organization", data);
    return response.data.data;
  },

  // Members & Invitations
  async listMembers(): Promise<{ members: OrganizationMember[]; invitations: OrganizationInvitation[] }> {
    const response = await apiClient.get<APIResponse<{ members: OrganizationMember[]; invitations: OrganizationInvitation[] }>>(
      "/settings/members"
    );
    return response.data.data;
  },

  async inviteMember(data: { email: string; role: OrganizationRole }): Promise<OrganizationInvitation> {
    const response = await apiClient.post<APIResponse<OrganizationInvitation>>("/settings/members/invite", data);
    return response.data.data;
  },

  async updateMemberRole(userId: string, role: OrganizationRole): Promise<void> {
    await apiClient.put(`/settings/members/${userId}/role`, { role });
  },

  async removeMember(userId: string): Promise<void> {
    await apiClient.delete(`/settings/members/${userId}`);
  },

  async deleteInvitation(invitationId: string): Promise<void> {
    await apiClient.delete(`/settings/invitations/${invitationId}`);
  },

  // API Keys
  async listAPIKeys(): Promise<APIKey[]> {
    const response = await apiClient.get<APIResponse<APIKey[]>>("/settings/api-keys");
    return response.data.data || [];
  },

  async createAPIKey(data: { name: string; scopes: string[]; expires_in_days?: number }): Promise<APIKey> {
    const response = await apiClient.post<APIResponse<APIKey>>("/settings/api-keys", data);
    return response.data.data;
  },

  async deleteAPIKey(id: string): Promise<void> {
    await apiClient.delete(`/settings/api-keys/${id}`);
  },

  // Webhooks
  async listWebhooks(): Promise<Webhook[]> {
    const response = await apiClient.get<APIResponse<Webhook[]>>("/settings/webhooks");
    return response.data.data || [];
  },

  async createWebhook(data: { name: string; url: string; secret?: string; events: string[] }): Promise<Webhook> {
    const response = await apiClient.post<APIResponse<Webhook>>("/settings/webhooks", data);
    return response.data.data;
  },

  async updateWebhook(
    id: string,
    data: { name: string; url: string; secret?: string; events: string[]; is_active: boolean }
  ): Promise<Webhook> {
    const response = await apiClient.put<APIResponse<Webhook>>(`/settings/webhooks/${id}`, data);
    return response.data.data;
  },

  async testWebhook(id: string): Promise<{ http_status: number }> {
    const response = await apiClient.post<APIResponse<{ http_status: number }>>(`/settings/webhooks/${id}/test`);
    return response.data.data;
  },

  async deleteWebhook(id: string): Promise<void> {
    await apiClient.delete(`/settings/webhooks/${id}`);
  },

  // Audit Logs
  async listAuditLogs(page = 1, limit = 20): Promise<{ data: AuditLog[]; total: number; page: number; limit: number }> {
    const response = await apiClient.get<PaginatedResponse<AuditLog>>(`/settings/audit-logs?page=${page}&limit=${limit}`);
    return {
      data: response.data.data || [],
      total: response.data.meta?.total_items || 0,
      page: response.data.meta?.page || page,
      limit: response.data.meta?.limit || limit,
    };
  },
};
