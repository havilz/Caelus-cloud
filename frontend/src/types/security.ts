export type ScanType = "full" | "port" | "tls" | "headers" | "host_config" | "vuln";
export type ScanStatus = "pending" | "running" | "completed" | "failed";
export type FindingSeverity = "critical" | "high" | "medium" | "low" | "info";
export type FindingCategory = "network" | "tls" | "http_headers" | "host_config" | "vulnerability";
export type FindingStatus = "open" | "acknowledged" | "resolved" | "false_positive";
export type IncidentStatus = "open" | "investigating" | "mitigated" | "closed";

export interface SecurityScan {
  id: string;
  organization_id: string;
  server_id?: string;
  server_name?: string;
  scan_type: ScanType;
  status: ScanStatus;
  started_at?: string;
  completed_at?: string;
  total_findings: number;
  critical_count: number;
  high_count: number;
  medium_count: number;
  low_count: number;
  score: number;
  error_message?: string;
  created_at: string;
  updated_at: string;
}

export interface SecurityFinding {
  id: string;
  organization_id: string;
  server_id?: string;
  server_name?: string;
  scan_id?: string;
  fingerprint: string;
  category: FindingCategory;
  severity: FindingSeverity;
  title: string;
  description: string;
  evidence?: Record<string, unknown>;
  recommendation?: string;
  remediation_command?: string;
  status: FindingStatus;
  resolved_at?: string;
  first_detected_at: string;
  last_detected_at: string;
}

export interface SecurityPostureOverview {
  overall_score: number;
  grade: "A" | "B" | "C" | "D" | "F";
  total_scans: number;
  open_findings: number;
  critical_count: number;
  high_count: number;
  medium_count: number;
  low_count: number;
  resolved_count: number;
  last_scan_at?: string;
  category_summary: Record<FindingCategory, number>;
}

export interface SecurityIncident {
  id: string;
  organization_id: string;
  title: string;
  severity: FindingSeverity;
  status: IncidentStatus;
  finding_ids: string[];
  summary?: string;
  mitigation_notes?: string;
  created_at: string;
  updated_at: string;
}

export interface TriggerScanInput {
  server_id?: string;
  scan_type: ScanType;
}

export interface ListFindingsFilter {
  server_id?: string;
  category?: FindingCategory;
  severity?: FindingSeverity;
  status?: FindingStatus;
  page?: number;
  limit?: number;
}
