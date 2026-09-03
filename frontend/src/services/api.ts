import axios from "axios";

export const getApiBaseURL = (): string => {
  if (typeof window !== "undefined") {
    
    if (window.location.port === "3000") {
      return `${window.location.protocol}//${window.location.hostname}:8080/api/v1`;
    }
    
    return `${window.location.origin}/api/v1`;
  }
  return process.env.INTERNAL_API_URL ? `${process.env.INTERNAL_API_URL}/api/v1` : (process.env.NEXT_PUBLIC_API_URL || "http://caelus-api:8080/api/v1");
};

export const apiClient = axios.create({
  headers: {
    "Content-Type": "application/json",
  },
  timeout: 15000,
});

apiClient.interceptors.request.use(
  (config) => {
    config.baseURL = getApiBaseURL();
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
