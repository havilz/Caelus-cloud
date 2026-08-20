import { apiClient } from './api';
import { APIResponse, PaginatedResponse } from '@/types/api';
import {
  AutomationRule,
  CreateRulePayload,
  UpdateRulePayload,
  RuleExecutionLog,
  ExecutionStatus,
} from '@/types/automation';

export const automationService = {
  // Rules
  async listRules(page = 1, limit = 50): Promise<PaginatedResponse<AutomationRule>> {
    const res = await apiClient.get<PaginatedResponse<AutomationRule>>('/automation/rules', {
      params: { page, limit },
    });
    return res.data;
  },

  async getRule(id: string): Promise<AutomationRule> {
    const res = await apiClient.get<APIResponse<AutomationRule>>(`/automation/rules/${id}`);
    return res.data.data!;
  },

  async createRule(payload: CreateRulePayload): Promise<AutomationRule> {
    const res = await apiClient.post<APIResponse<AutomationRule>>('/automation/rules', payload);
    return res.data.data!;
  },

  async updateRule(id: string, payload: UpdateRulePayload): Promise<AutomationRule> {
    const res = await apiClient.put<APIResponse<AutomationRule>>(`/automation/rules/${id}`, payload);
    return res.data.data!;
  },

  async deleteRule(id: string): Promise<void> {
    await apiClient.delete(`/automation/rules/${id}`);
  },

  async testRule(id: string, mockData?: Record<string, unknown>): Promise<RuleExecutionLog> {
    const res = await apiClient.post<APIResponse<RuleExecutionLog>>(`/automation/rules/${id}/test`, {
      mock_data: mockData || {},
    });
    return res.data.data!;
  },

  // Execution Logs
  async listLogs(
    page = 1,
    limit = 20,
    ruleId?: string,
    status?: ExecutionStatus
  ): Promise<PaginatedResponse<RuleExecutionLog>> {
    const res = await apiClient.get<PaginatedResponse<RuleExecutionLog>>('/automation/logs', {
      params: {
        page,
        limit,
        rule_id: ruleId || undefined,
        status: status || undefined,
      },
    });
    return res.data;
  },
};
