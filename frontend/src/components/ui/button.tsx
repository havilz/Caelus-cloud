import React from "react";
import { clsx } from "clsx";
import { twMerge } from "tailwind-merge";

function cn(...inputs: any[]) {
  return twMerge(clsx(inputs));
}

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  readonly variant?: "primary" | "secondary" | "outline" | "danger" | "ghost";
  readonly size?: "sm" | "md" | "lg";
  readonly isLoading?: boolean;
}

export const Button: React.FC<ButtonProps> = ({
  children,
  className,
  variant = "primary",
  size = "md",
  isLoading = false,
  disabled,
  type = "button",
  ...props
}) => {
  const baseStyles = "inline-flex items-center justify-center font-medium rounded-lg transition-all focus:outline-none focus:ring-2 focus:ring-offset-1 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer";
  
  const variants = {
    primary: "bg-emerald-500 hover:bg-emerald-400 text-zinc-950 font-semibold focus:ring-emerald-500 border border-emerald-400/30 shadow-sm",
    secondary: "bg-[#202020] hover:bg-[#282828] dark:bg-[#202020] dark:hover:bg-[#282828] light:bg-[#f3f4f6] light:hover:bg-[#e5e7eb] text-[#ededed] dark:text-[#ededed] light:text-[#111827] border border-[#2e2e2e] dark:border-[#2e2e2e] light:border-[#d1d5db] focus:ring-zinc-600",
    outline: "bg-transparent hover:bg-[#202020] dark:hover:bg-[#202020] light:hover:bg-[#f3f4f6] text-[#ededed] dark:text-[#ededed] light:text-[#111827] border border-[#2e2e2e] dark:border-[#2e2e2e] light:border-[#d1d5db] focus:ring-zinc-600",
    danger: "bg-rose-600 hover:bg-rose-500 text-white border border-rose-500/30 focus:ring-rose-500 shadow-sm",
    ghost: "bg-transparent hover:bg-[#202020] dark:hover:bg-[#202020] light:hover:bg-[#f3f4f6] text-[#a1a1a1] hover:text-[#ededed] dark:hover:text-[#ededed] light:hover:text-[#111827] focus:ring-zinc-600",
  };

  const sizes = {
    sm: "text-xs px-2.5 py-1.5 gap-1.5",
    md: "text-xs px-3.5 py-2 gap-2",
    lg: "text-sm px-4 py-2.5 gap-2.5",
  };

  return (
    <button
      type={type}
      className={cn(baseStyles, variants[variant], sizes[size], className)}
      disabled={disabled || isLoading}
      {...props}
    >
      {isLoading && (
        <svg className="animate-spin -ml-1 mr-2 h-3.5 w-3.5 text-current" fill="none" viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"></path>
        </svg>
      )}
      {children}
    </button>
  );
};
