import { apiClient } from './api';
import { APIResponse, PaginatedResponse } from '@/types/api';
import {
  Bucket,
  CreateBucketInput,
  ObjectsListResponse,
  ObjectItem,
  GenerateSignedURLInput,
  SignedURLResponse,
} from '@/types/storage';

export const storageService = {
  // 1. Buckets
  async listBuckets(page = 1, limit = 20): Promise<PaginatedResponse<Bucket>> {
    const res = await apiClient.get<PaginatedResponse<Bucket>>('/storage/buckets', {
      params: { page, limit },
    });
    return res.data;
  },

  async getBucket(name: string): Promise<Bucket> {
    const res = await apiClient.get<APIResponse<Bucket>>(`/storage/buckets/${name}`);
    return res.data.data!;
  },

  async createBucket(input: CreateBucketInput): Promise<Bucket> {
    const res = await apiClient.post<APIResponse<Bucket>>('/storage/buckets', input);
    return res.data.data!;
  },

  async deleteBucket(name: string): Promise<void> {
    await apiClient.delete(`/storage/buckets/${name}`);
  },

  // 2. Objects & Explorer
  async listObjects(
    bucketName: string,
    prefix = '',
    delimiter = '/',
    maxKeys = 1000
  ): Promise<ObjectsListResponse> {
    const res = await apiClient.get<APIResponse<ObjectsListResponse>>(
      `/storage/buckets/${bucketName}/objects`,
      {
        params: { prefix, delimiter, max_keys: maxKeys },
      }
    );
    return res.data.data!;
  },

  async uploadObject(
    bucketName: string,
    file: File,
    key?: string,
    onProgress?: (percent: number) => void
  ): Promise<ObjectItem> {
    const formData = new FormData();
    formData.append('file', file);
    if (key) {
      formData.append('key', key);
    }

    const res = await apiClient.post<APIResponse<ObjectItem>>(
      `/storage/buckets/${bucketName}/objects`,
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
        onUploadProgress: (progressEvent) => {
          if (progressEvent.total && onProgress) {
            const percent = Math.round(
              (progressEvent.loaded * 100) / progressEvent.total
            );
            onProgress(percent);
          }
        },
      }
    );
    return res.data.data!;
  },

  async downloadObject(bucketName: string, key: string): Promise<Blob> {
    const res = await apiClient.get(`/storage/buckets/${bucketName}/objects/download`, {
      params: { key },
      responseType: 'blob',
    });
    return res.data;
  },

  async deleteObject(bucketName: string, key: string): Promise<void> {
    await apiClient.delete(`/storage/buckets/${bucketName}/objects`, {
      params: { key },
    });
  },

  async deleteObjects(bucketName: string, keys: string[]): Promise<void> {
    await apiClient.delete(`/storage/buckets/${bucketName}/objects`, {
      params: { keys: keys.join(',') },
    });
  },

  async generateSignedURL(
    bucketName: string,
    input: GenerateSignedURLInput
  ): Promise<SignedURLResponse> {
    const res = await apiClient.post<APIResponse<SignedURLResponse>>(
      `/storage/buckets/${bucketName}/objects/signed-url`,
      input
    );
    return res.data.data!;
  },
};
