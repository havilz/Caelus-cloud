import React, { useEffect } from "react";
import { X } from "lucide-react";
import { clsx } from "clsx";
import { twMerge } from "tailwind-merge";
import { AppContainers, AppText } from "@/core/theme";

function cn(...inputs: any[]) {
  return twMerge(clsx(inputs));
}

export interface DialogProps {
  readonly isOpen: boolean;
  readonly onClose: () => void;
  readonly title: string;
  readonly description?: string;
  readonly children: React.ReactNode;
  readonly maxWidth?: "sm" | "md" | "lg" | "xl";
}

export const Dialog: React.FC<DialogProps> = ({
  isOpen,
  onClose,
  title,
  description,
  children,
  maxWidth = "md",
}) => {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    if (isOpen) {
      document.body.style.overflow = "hidden";
      window.addEventListener("keydown", handleKeyDown);
    }
    return () => {
      document.body.style.overflow = "unset";
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const maxWidths = {
    sm: "max-w-sm",
    md: "max-w-md",
    lg: "max-w-lg",
    xl: "max-w-xl",
  };

  return (
    <div className={AppContainers.modalBackdrop}>
      {/* Backdrop button - Semi-transparent opacity (50%) */}
      <button
        type="button"
        aria-label="Tutup modal backdrop"
        className="fixed inset-0 bg-black/50 backdrop-blur-[2px] transition-opacity border-0 w-full h-full cursor-default"
        onClick={onClose}
      />

      {/* Modal Dialog Box */}
      <div
        className={cn(
          AppContainers.modalDialog,
          "p-6",
          maxWidths[maxWidth]
        )}
      >
        {/* Header */}
        <div className="flex items-center justify-between pb-4 border-b border-[#262626] shrink-0">
          <div>
            <h3 className={AppText.h4}>{title}</h3>
            {description && <p className={AppText.subtitle}>{description}</p>}
          </div>
          <button
            type="button"
            aria-label="Tutup modal"
            onClick={onClose}
            className="rounded-lg p-1.5 text-[#a1a1a1] hover:bg-[#222222] hover:text-[#ededed] transition-colors cursor-pointer"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        {/* Scrollable Modal Body */}
        <div className={cn(AppContainers.modalBodyScroll, "mt-4")}>
          {children}
        </div>
      </div>
    </div>
  );
};
