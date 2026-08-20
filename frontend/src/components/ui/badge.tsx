import React from "react";
import { clsx } from "clsx";
import { twMerge } from "tailwind-merge";
import { AppColors } from "@/core/theme";

function cn(...inputs: any[]) {
  return twMerge(clsx(inputs));
}

export interface BadgeProps extends React.HTMLAttributes<HTMLDivElement> {
  readonly variant?: "default" | "success" | "warning" | "danger" | "info" | "outline";
}

export const Badge: React.FC<BadgeProps> = ({
  children,
  className,
  variant = "default",
  ...props
}) => {
  const base = "inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md text-[11px] font-medium tracking-wide border";
  
  const variants = {
    default: "bg-[#202020] text-[#a1a1a1] border-[#2e2e2e]",
    success: cn(AppColors.status.running.bg, AppColors.status.running.text, AppColors.status.running.border),
    warning: cn(AppColors.status.restarting.bg, AppColors.status.restarting.text, AppColors.status.restarting.border),
    danger: cn(AppColors.status.danger.bg, AppColors.status.danger.text, AppColors.status.danger.border),
    info: cn(AppColors.status.info.bg, AppColors.status.info.text, AppColors.status.info.border),
    outline: "bg-transparent text-[#a1a1a1] border-[#2e2e2e]",
  };

  return (
    <div className={cn(base, variants[variant], className)} {...props}>
      {children}
    </div>
  );
};
