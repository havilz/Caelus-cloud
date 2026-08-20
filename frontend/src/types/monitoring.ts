export type AlertSeverity = "critical" | "warning" | "info";
export type AlertStatus = "active" | "acknowledged" | "resolved";

export interface ServerMetric {
  id: number;
  server_id: string;
  cpu_usage_pct: number;
  memory_used_mb: number;
  memory_total_mb: number;
  memory_usage_pct: number;
  disk_used_gb: number;
  disk_total_gb: number;
  disk_usage_pct: number;
  network_in_kb: number;
  network_out_kb: number;
  network_in_rate_kbps: number;
  network_out_rate_kbps: number;
  uptime_seconds: number;
  containers_count: number;
  docker_available: boolean;
  containers_json?: string;
  recorded_at: string;
}

export interface ContainerMetric {
  id: string;
  names: string[];
  image: string;
  state: string;
  status: string;
  created: number;
  cpu_usage_pct: number;
  memory_usage_mb: number;
  memory_limit_mb: number;
}

export interface Alert {
  id: string;
  organization_id: string;
  server_id: string;
  rule_id?: string;
  alert_type: string;
  severity: AlertSeverity;
  title: string;
  message: string;
  status: AlertStatus;
  current_value?: number;
  threshold_value?: number;
  acknowledged_at?: string;
  acknowledged_by?: string;
  resolved_at?: string;
  resolved_by?: string;
  triggered_at: string;
  created_at: string;
  server?: {
    id: string;
    name: string;
    ip_address?: string;
  };
}

export interface AlertRule {
  id: string;
  organization_id: string;
  server_id?: string;
  name: string;
  metric_name: string;
  operator: string;
  threshold: number;
  duration_seconds: number;
  severity: AlertSeverity;
  is_enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface LogEntry {
  id: string;
  timestamp: string;
  line: string;
  level: "INFO" | "WARN" | "ERROR" | "DEBUG";
  service?: string;
}
