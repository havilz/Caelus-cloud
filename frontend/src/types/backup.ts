export type BackupStatus = 'pending' | 'in_progress' | 'completed' | 'failed';

export interface BackupPolicy {
  id: string;
  organization_id: string;
  server_id: string;
  bucket_id?: string;
  name: string;
  cron_expression: string;
  retention_days: number;
  include_disks: boolean;
  is_active: boolean;
  next_run_at?: string;
  last_run_at?: string;
  created_at: string;
  updated_at: string;
  server_name?: string;
  bucket_name?: string;
}

export interface BackupRecord {
  id: string;
  organization_id: string;
  policy_id?: string;
  server_id: string;
  bucket_id?: string;
  backup_name: string;
  storage_key: string;
  size_bytes: number;
  status: BackupStatus;
  error_message?: string;
  checksum_sha256?: string;
  started_at: string;
  completed_at?: string;
  expires_at?: string;
  created_at: string;
  updated_at: string;
  server_name?: string;
  bucket_name?: string;
}

export interface CreateBackupPolicyInput {
  server_id: string;
  bucket_id?: string;
  name: string;
  cron_expression: string;
  retention_days: number;
  include_disks: boolean;
}

export interface TriggerBackupInput {
  backup_name?: string;
  policy_id?: string;
}
