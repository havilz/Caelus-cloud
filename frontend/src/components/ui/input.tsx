import React from "react";
import { clsx } from "clsx";
import { twMerge } from "tailwind-merge";
import { AppText } from "@/core/theme";

function cn(...inputs: any[]) {
  return twMerge(clsx(inputs));
}

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  readonly label?: string;
  readonly error?: string;
  readonly helperText?: string;
}

export const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, label, error, helperText, type, id, ...props }, ref) => {
    const inputId = id || (label ? label.toLowerCase().replaceAll(/\s+/g, "-") : undefined);

    return (
      <div className="w-full space-y-1.5">
        {label && (
          <label htmlFor={inputId} className={AppText.label}>
            {label}
          </label>
        )}
        <input
          type={type}
          id={inputId}
          ref={ref}
          className={cn(
            "w-full rounded-lg border bg-[#121212] dark:bg-[#121212] light:bg-[#ffffff] px-3.5 py-2 text-xs text-[#ededed] dark:text-[#ededed] light:text-[#111827] placeholder-[#707070] transition-colors",
            "border-[#262626] dark:border-[#262626] light:border-[#d1d5db] focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500",
            error && "border-rose-500 focus:border-rose-500 focus:ring-rose-500",
            className
          )}
          {...props}
        />
        {error && <p className="text-xs text-rose-400 mt-1">{error}</p>}
        {helperText && !error && <p className={AppText.caption}>{helperText}</p>}
      </div>
    );
  }
);

Input.displayName = "Input";
