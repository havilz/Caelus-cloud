import { apiClient } from "./api";
import { APIResponse } from "@/types/api";
import { Provider } from "@/types/server";

export const providerService = {
  async listProviders(): Promise<Provider[]> {
    const response = await apiClient.get<APIResponse<Provider[]>>("/providers");
    return response.data.data || [];
  },
};
