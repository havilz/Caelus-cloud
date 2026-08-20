import { apiClient } from "./api";
import { APIResponse } from "@/types/api";
import { AuthData, LoginInput, RegisterInput, TokenPair } from "@/types/auth";

export const authService = {
  async register(input: RegisterInput): Promise<AuthData> {
    const response = await apiClient.post<APIResponse<AuthData>>("/auth/register", input);
    return response.data.data;
  },

  async login(input: LoginInput): Promise<AuthData> {
    const response = await apiClient.post<APIResponse<AuthData>>("/auth/login", input);
    return response.data.data;
  },

  async refreshToken(refreshToken: string): Promise<TokenPair> {
    const response = await apiClient.post<APIResponse<TokenPair>>("/auth/refresh", {
      refresh_token: refreshToken,
    });
    return response.data.data;
  },
};
