import { Provider } from "./server";

export interface Credential {
  id: string;
  organization_id: string;
  provider_id: string;
  name: string;
  metadata?: Record<string, any>;
  created_at: string;
  updated_at: string;
  provider?: Provider;
}

export interface CreateCredentialDTO {
  provider_id: string;
  name: string;
  api_key?: string;
  api_secret?: string;
  ssh_key?: string;
  metadata?: Record<string, any>;
}

export interface UpdateCredentialDTO {
  name?: string;
  api_key?: string;
  api_secret?: string;
  ssh_key?: string;
  metadata?: Record<string, any>;
}

export interface TestCredentialResult {
  provider: string;
  status: "connected" | "failed";
  server_count?: number;
}
