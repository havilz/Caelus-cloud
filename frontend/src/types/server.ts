export type ServerStatus = "running" | "stopped" | "restarting" | "error" | "provisioning";

export interface Provider {
  id: string;
  name: string;
  slug: string;
  is_active: boolean;
  created_at?: string;
}

export interface Server {
  id: string;
  organization_id: string;
  credential_id?: string;
  provider_id: string;
  external_server_id?: string;
  name: string;
  hostname?: string;
  ip_address?: string;
  status: ServerStatus;
  os_type: string;
  cpu_cores: number;
  memory_mb: number;
  disk_gb: number;
  region: string;
  created_at: string;
  updated_at: string;
  provider?: Provider;
}

export interface CreateServerDTO {
  provider_id: string;
  credential_id?: string;
  name: string;
  region: string;
  os_type: string;
  plan_id?: string;
  cpu_cores: number;
  memory_mb: number;
  disk_gb: number;
}

export interface ResizeServerDTO {
  cpu_cores: number;
  memory_mb: number;
  disk_gb: number;
  plan_id?: string;
}
