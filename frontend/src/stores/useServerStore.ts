import { create } from "zustand";
import { Server, CreateServerDTO, ResizeServerDTO } from "@/types/server";
import { serverService } from "@/services/server.service";

interface ServerState {
  servers: Server[];
  totalServers: number;
  currentPage: number;
  totalPages: number;
  isLoading: boolean;
  error: string | null;

  fetchServers: (page?: number) => Promise<void>;
  createServer: (data: CreateServerDTO) => Promise<Server>;
  rebootServer: (id: string) => Promise<void>;
  shutdownServer: (id: string) => Promise<void>;
  startServer: (id: string) => Promise<void>;
  resizeServer: (id: string, data: ResizeServerDTO) => Promise<void>;
  deleteServer: (id: string) => Promise<void>;
}

export const useServerStore = create<ServerState>((set, get) => ({
  servers: [],
  totalServers: 0,
  currentPage: 1,
  totalPages: 1,
  isLoading: false,
  error: null,

  fetchServers: async (page = 1) => {
    set({ isLoading: true, error: null });
    try {
      const res = await serverService.listServers(page);
      set({
        servers: res.data || [],
        totalServers: res.meta.total_items,
        currentPage: res.meta.page,
        totalPages: res.meta.total_pages,
        isLoading: false,
      });
    } catch (err: any) {
      set({
        error: err.response?.data?.message || "Gagal memuat daftar server",
        isLoading: false,
      });
    }
  },

  createServer: async (data: CreateServerDTO) => {
    set({ isLoading: true, error: null });
    try {
      const newServer = await serverService.createServer(data);
      await get().fetchServers(get().currentPage);
      set({ isLoading: false });
      return newServer;
    } catch (err: any) {
      set({
        error: err.response?.data?.message || err.response?.data?.errors || "Gagal membuat server",
        isLoading: false,
      });
      throw err;
    }
  },

  rebootServer: async (id: string) => {
    await serverService.rebootServer(id);
    await get().fetchServers(get().currentPage);
  },

  shutdownServer: async (id: string) => {
    await serverService.shutdownServer(id);
    await get().fetchServers(get().currentPage);
  },

  startServer: async (id: string) => {
    await serverService.startServer(id);
    await get().fetchServers(get().currentPage);
  },

  resizeServer: async (id: string, data: ResizeServerDTO) => {
    await serverService.resizeServer(id, data);
    await get().fetchServers(get().currentPage);
  },

  deleteServer: async (id: string) => {
    await serverService.deleteServer(id);
    await get().fetchServers(get().currentPage);
  },
}));
