export interface APIResponse<T = any> {
  success: boolean;
  message?: string;
  data: T;
  errors?: any;
}

export interface PaginatedMeta {
  page: number;
  limit: number;
  total_items: number;
  total_pages: number;
}

export interface PaginatedResponse<T = any> {
  success: boolean;
  message?: string;
  data: T[];
  meta: PaginatedMeta;
}
