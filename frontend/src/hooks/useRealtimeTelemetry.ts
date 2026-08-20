"use client";

import { useEffect, useState, useRef, useCallback } from "react";
import { ServerMetric, Alert } from "@/types/monitoring";
import { useAuthStore } from "@/stores/useAuthStore";

interface UseRealtimeTelemetryProps {
  serverId?: string;
  orgId?: string;
  onMetricUpdate?: (metric: ServerMetric) => void;
  onAlert?: (alert: Alert) => void;
}

export function useRealtimeTelemetry({
  serverId,
  orgId,
  onMetricUpdate,
  onAlert,
}: UseRealtimeTelemetryProps = {}) {
  const [isConnected, setIsConnected] = useState(false);
  const [latestMetric, setLatestMetric] = useState<ServerMetric | null>(null);
  const [latestAlert, setLatestAlert] = useState<Alert | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const accessToken = useAuthStore((state) => state.accessToken);

  const connect = useCallback(() => {
    const token =
      accessToken ||
      (typeof window !== "undefined"
        ? localStorage.getItem("caelus_access_token")
        : null);

    if (!token) return;

    // Bersihkan timeout reconnect sebelumnya jika ada
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }

    const wsUrl = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080";
    const endpoint = `${wsUrl}/api/v1/ws?token=${encodeURIComponent(token)}`;

    try {
      const ws = new WebSocket(endpoint);
      wsRef.current = ws;

      ws.onopen = () => {
        setIsConnected(true);
        // Subscribe ke topik server tertentu jika serverId disediakan
        if (serverId) {
          ws.send(JSON.stringify({ type: "subscribe", topic: `server:${serverId}` }));
        }
        // Subscribe ke topik organisasi jika orgId disediakan
        if (orgId) {
          ws.send(JSON.stringify({ type: "subscribe", topic: `org:${orgId}` }));
        }
      };

      ws.onmessage = (event) => {
        try {
          const parsed = JSON.parse(event.data);
          if (parsed.event === "metrics.updated") {
            const metric = parsed.data as ServerMetric;
            setLatestMetric(metric);
            onMetricUpdate?.(metric);
          } else if (parsed.event === "alert.created") {
            const alert = parsed.data as Alert;
            setLatestAlert(alert);
            onAlert?.(alert);
          }
        } catch {
          // Abaikan pesan yang bukan format JSON terstruktur
        }
      };

      ws.onclose = () => {
        setIsConnected(false);
        // Jadwalkan rekoneksi otomatis setelah 3 detik
        reconnectTimeoutRef.current = setTimeout(() => {
          connect();
        }, 3000);
      };

      ws.onerror = () => {
        setIsConnected(false);
        ws.close();
      };
    } catch {
      setIsConnected(false);
    }
  }, [accessToken, serverId, orgId, onMetricUpdate, onAlert]);

  useEffect(() => {
    connect();

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [connect]);

  return {
    isConnected,
    latestMetric,
    latestAlert,
  };
}
