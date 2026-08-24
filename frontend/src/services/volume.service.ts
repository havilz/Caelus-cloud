import { apiClient } from './api';
import { APIResponse } from '@/types/api';

export interface StorageVolume {
  id: string;
  organization_id?: string;
  server_id?: string | null;
  name: string;
  size_gb: number;
  sizeGB?: number;
  type: 'nvme' | 'ssd' | 'docker-volume';
  fs_type: 'ext4' | 'xfs' | 'btrfs';
  fsType?: 'ext4' | 'xfs' | 'btrfs';
  mount_path: string;
  mountPath?: string;
  status: 'available' | 'in-use' | 'mounting';
  attached_container_name?: string;
  attachedContainerName?: string;
  attachedServerName?: string | null;
  attachedServerId?: string | null;
  iops?: number;
  created_at?: string;
  createdAt?: string;
  updated_at?: string;
}

export interface StoragePoolStats {
  total_bytes: number;
  used_bytes: number;
  free_bytes: number;
  total_gb: number;
  used_gb: number;
  free_gb: number;
  usage_percent: number;
  storage_path: string;
}

export interface CreateVolumeRequest {
  server_id?: string | null;
  name: string;
  size_gb: number;
  type: 'nvme' | 'ssd' | 'docker-volume';
  fs_type: 'ext4' | 'xfs' | 'btrfs';
  mount_path: string;
}

export const volumeService = {
  async getStoragePoolStats(): Promise<StoragePoolStats> {
    const response = await apiClient.get<APIResponse<StoragePoolStats>>('/volumes/stats');
    return response.data.data;
  },

  async listVolumes(): Promise<StorageVolume[]> {
    const response = await apiClient.get<APIResponse<StorageVolume[]>>('/volumes');
    return response.data.data || [];
  },

  async createVolume(data: CreateVolumeRequest): Promise<StorageVolume> {
    const response = await apiClient.post<APIResponse<StorageVolume>>('/volumes', data);
    return response.data.data;
  },

  async deleteVolume(id: string): Promise<void> {
    await apiClient.delete(`/volumes/${id}`);
  },
};
