import { apiClient } from "./api";

export type DomainStatus = "pending_dns" | "verifying" | "active" | "error";
export type SSLStatus = "none" | "pending" | "active" | "expired" | "error";
export type IngressTargetType = "container" | "port" | "service" | "storage";

export interface CustomDomain {
  id: string;
  organization_id: string;
  server_id?: string;
  server_name?: string;
  server_public_ip?: string;
  domain_name: string;
  target_type: IngressTargetType;
  target_id: string;
  target_port: number;
  status: DomainStatus;
  verification_token: string;
  ssl_status: SSLStatus;
  auto_ssl: boolean;
  cloudflare_dns_managed: boolean;
  cloudflare_record_id?: string;
  error_message?: string;
  last_checked_at?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateDomainRequest {
  server_id?: string;
  domain_name: string;
  target_type: IngressTargetType;
  target_id: string;
  target_port: number;
  auto_ssl: boolean;
  cloudflare_dns_managed: boolean;
}

export interface VerifyDomainResponse {
  domain_id: string;
  domain_name: string;
  status: DomainStatus;
  verified: boolean;
  expected_ip: string;
  resolved_ips: string[];
  expected_txt: string;
  resolved_txt: string[];
  ssl_status: SSLStatus;
  message: string;
}

export const domainService = {
  async listDomains(): Promise<CustomDomain[]> {
    const res = await apiClient.get("/domains");
    return res.data?.data || [];
  },

  async getDomain(id: string): Promise<CustomDomain> {
    const res = await apiClient.get(`/domains/${id}`);
    return res.data?.data;
  },

  async createDomain(data: CreateDomainRequest): Promise<CustomDomain> {
    const res = await apiClient.post("/domains", data);
    return res.data?.data;
  },

  async deleteDomain(id: string): Promise<void> {
    await apiClient.delete(`/domains/${id}`);
  },

  async verifyDomain(id: string): Promise<VerifyDomainResponse> {
    const res = await apiClient.post(`/domains/${id}/verify`);
    return res.data?.data;
  },
};
