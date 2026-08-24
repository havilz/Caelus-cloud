export type DeploymentStatus =
  | 'queued'
  | 'pulling'
  | 'building'
  | 'deploying'
  | 'running'
  | 'failed'
  | 'stopped'
  | 'rolled_back';

export interface PortBinding {
  host_port: number;
  container_port: number;
  protocol: 'tcp' | 'udp' | string;
}

export interface VolumeBinding {
  host_path: string;
  container_path: string;
  mode: 'rw' | 'ro' | string;
}

export interface Deployment {
  id: string;
  organization_id: string;
  server_id?: string;
  app_name: string;
  image_tag: string;
  container_name: string;
  environment_variables?: Record<string, string>;
  port_bindings?: PortBinding[];
  volume_bindings?: VolumeBinding[];
  restart_policy: string;
  network_name?: string;
  status: DeploymentStatus;
  error_message?: string;
  created_at: string;
  updated_at: string;
  finished_at?: string;
}

export interface DeploymentLog {
  id: number;
  deployment_id: string;
  timestamp: string;
  stream: 'stdout' | 'stderr' | 'system' | string;
  message: string;
}

export interface DeploymentRequest {
  server_id?: string;
  app_name: string;
  image_tag: string;
  container_name?: string;
  network_name?: string;
  environment_variables?: Record<string, string>;
  port_bindings?: PortBinding[];
  volume_bindings?: VolumeBinding[];
  restart_policy?: string;
}
