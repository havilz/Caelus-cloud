import axios from "axios";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL;

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
  timeout: 15000,
});

apiClient.interceptors.request.use(
  (config) => {
    if (typeof window !== "undefined") {
      const token = localStorage.getItem("caelus_access_token");
      if (token && config.headers) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    }
    return config;
  },
  (error) => Promise.reject(error)
);

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401 && typeof window !== "undefined") {
      if (!window.location.pathname.startsWith("/login") && !window.location.pathname.startsWith("/register")) {
        localStorage.removeItem("caelus_access_token");
        localStorage.removeItem("caelus_refresh_token");
        localStorage.removeItem("caelus_user");
        localStorage.removeItem("caelus_org");
        window.location.href = "/login";
      }
    }
    return Promise.reject(error);
  }
);
