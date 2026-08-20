export type StorageProviderType = 'minio' | 's3' | 'r2' | 'mock';

export interface Bucket {
  id: string;
  organization_id: string;
  name: string;
  provider_type: StorageProviderType;
  region: string;
  is_public: boolean;
  versioning: boolean;
  created_at: string;
  updated_at: string;
}

export interface ObjectItem {
  key: string;
  size: number;
  etag: string;
  content_type: string;
  last_modified: string;
  storage_class?: string;
  metadata?: Record<string, string>;
}

export interface ObjectsListResponse {
  bucket: string;
  prefix: string;
  folders: string[];
  objects: ObjectItem[];
}

export interface CreateBucketInput {
  name: string;
  provider_type: StorageProviderType;
  region: string;
  is_public: boolean;
  versioning: boolean;
}

export type SignedURLOperation = 'download' | 'upload';

export interface GenerateSignedURLInput {
  key: string;
  operation: SignedURLOperation;
  expiry_minutes: number;
}

export interface SignedURLResponse {
  url: string;
  operation: SignedURLOperation;
  expires_in_sec: number;
  expires_at: string;
}
