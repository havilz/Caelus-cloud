import React from "react";
import { ServerStatus } from "@/types/server";
import { Badge } from "@/components/ui/badge";
import { AppColors } from "@/core/theme";

interface ServerStatusBadgeProps {
  readonly status: ServerStatus;
}

export const ServerStatusBadge: React.FC<ServerStatusBadgeProps> = ({ status }) => {
  switch (status) {
    case "running":
      return (
        <Badge variant="success">
          <span className={`h-1.5 w-1.5 rounded-full ${AppColors.status.running.dot} animate-pulse`} />
          <span>Running</span>
        </Badge>
      );
    case "stopped":
      return (
        <Badge variant="outline">
          <span className={`h-1.5 w-1.5 rounded-full ${AppColors.status.stopped.dot}`} />
          <span>Stopped</span>
        </Badge>
      );
    case "restarting":
      return (
        <Badge variant="warning">
          <span className={`h-1.5 w-1.5 rounded-full ${AppColors.status.restarting.dot} animate-ping`} />
          <span>Restarting</span>
        </Badge>
      );
    case "error":
      return (
        <Badge variant="danger">
          <span className={`h-1.5 w-1.5 rounded-full ${AppColors.status.danger.dot}`} />
          <span>Error</span>
        </Badge>
      );
    default:
      return (
        <Badge variant="warning">
          <span className={`h-1.5 w-1.5 rounded-full ${AppColors.status.restarting.dot}`} />
          <span>{status}</span>
        </Badge>
      );
  }
};
