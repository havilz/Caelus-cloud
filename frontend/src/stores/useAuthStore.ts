import { create } from "zustand";
import { AuthData, Organization, User } from "@/types/auth";

interface AuthState {
  user: User | null;
  organization: Organization | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  setAuth: (data: AuthData) => void;
  logout: () => void;
  initialize: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  organization: null,
  accessToken: null,
  refreshToken: null,
  isAuthenticated: false,
  isLoading: true,

  setAuth: (data: AuthData) => {
    if (typeof window !== "undefined") {
      localStorage.setItem("caelus_access_token", data.tokens.access_token);
      localStorage.setItem("caelus_refresh_token", data.tokens.refresh_token);
      localStorage.setItem("caelus_user", JSON.stringify(data.user));
      if (data.organization) {
        localStorage.setItem("caelus_org", JSON.stringify(data.organization));
      }
    }
    set({
      user: data.user,
      organization: data.organization || null,
      accessToken: data.tokens.access_token,
      refreshToken: data.tokens.refresh_token,
      isAuthenticated: true,
      isLoading: false,
    });
  },

  logout: () => {
    if (typeof window !== "undefined") {
      localStorage.removeItem("caelus_access_token");
      localStorage.removeItem("caelus_refresh_token");
      localStorage.removeItem("caelus_user");
      localStorage.removeItem("caelus_org");
    }
    set({
      user: null,
      organization: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
      isLoading: false,
    });
  },

  initialize: () => {
    if (typeof window !== "undefined") {
      const token = localStorage.getItem("caelus_access_token");
      const refresh = localStorage.getItem("caelus_refresh_token");
      const userStr = localStorage.getItem("caelus_user");
      const orgStr = localStorage.getItem("caelus_org");

      if (token && userStr) {
        try {
          const user = JSON.parse(userStr);
          const organization = orgStr ? JSON.parse(orgStr) : null;
          set({
            user,
            organization,
            accessToken: token,
            refreshToken: refresh,
            isAuthenticated: true,
            isLoading: false,
          });
          return;
        } catch {
          localStorage.clear();
        }
      }
    }
    set({ isLoading: false, isAuthenticated: false });
  },
}));
