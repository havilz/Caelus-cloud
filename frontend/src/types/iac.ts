export type IaCAction = 'create' | 'update' | 'delete' | 'noop';

export type IaCStatus = 'draft' | 'planned' | 'applying' | 'applied' | 'failed' | 'drifted' | 'rolled_back';

export type ResourceType = 'server' | 'storage' | 'container' | 'rule';

export interface IaCValidationError {
  line: number;
  column?: number;
  field: string;
  message: string;
}

export interface IaCValidationResponse {
  valid: boolean;
  errors?: IaCValidationError[];
  manifest?: any;
}

export interface IaCChange {
  resource_type: ResourceType;
  resource_name: string;
  action: IaCAction;
  before?: Record<string, any>;
  after?: Record<string, any>;
  changed_fields?: string[];
  reason?: string;
}

export interface IaCSummary {
  create: number;
  update: number;
  delete: number;
  noop: number;
  total: number;
}

export interface IaCPlan {
  id: string;
  configuration_id: string;
  target_version: number;
  changes: IaCChange[];
  summary: IaCSummary;
  status: IaCStatus;
  error_message?: string;
  created_at: string;
  executed_at?: string;
}

export interface IaCState {
  id: string;
  configuration_id: string;
  version: number;
  state_data: Record<string, any>;
  hash: string;
  applied_at: string;
  applied_by?: string;
  created_at: string;
}

export interface IaCConfiguration {
  id: string;
  organization_id: string;
  name: string;
  description: string;
  raw_yaml: string;
  status: IaCStatus;
  current_version: number;
  created_at: string;
  updated_at: string;
}
