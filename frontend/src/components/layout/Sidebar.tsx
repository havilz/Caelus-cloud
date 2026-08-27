"use client";

import React, { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { clsx } from "clsx";
import { twMerge } from "tailwind-merge";
import {
  LayoutDashboard,
  Server,
  Box,
  Network,
  HardDrive,
  Database,
  Activity,
  ShieldCheck,
  Zap,
  Settings,
  ChevronDown,
  Cloud,
  Code2,
  Globe,
} from "lucide-react";
import { AppText } from "@/core/theme";

function cn(...inputs: any[]) {
  return twMerge(clsx(inputs));
}

interface NavItem {
  readonly name: string;
  readonly href: string;
  readonly icon: React.ElementType;
  readonly badge?: string;
  readonly children?: readonly { readonly name: string; readonly href: string; readonly icon: React.ElementType }[];
}

const navItems: readonly NavItem[] = [
  {
    name: "Overview",
    href: "/overview",
    icon: LayoutDashboard,
  },
  {
    name: "Infrastructure",
    href: "/infrastructure/vps",
    icon: Server,
    children: [
      { name: "VPS & Servers", href: "/infrastructure/vps", icon: Server },
      { name: "Cloud Providers", href: "/infrastructure/providers", icon: Cloud },
      { name: "Declarative IaC", href: "/infrastructure/iac", icon: Code2 },
      { name: "Containers", href: "/infrastructure/containers", icon: Box },
      { name: "Custom Domains", href: "/infrastructure/domains", icon: Globe },
      { name: "Networks", href: "/infrastructure/networks", icon: Network },
      { name: "Volumes", href: "/infrastructure/volumes", icon: HardDrive },
    ],
  },
  {
    name: "Object Storage",
    href: "/storage",
    icon: Database,
  },
  {
    name: "Monitoring",
    href: "/monitoring",
    icon: Activity,
  },
  {
    name: "Security (Sentinel)",
    href: "/security",
    icon: ShieldCheck,
    badge: "Active",
  },
  {
    name: "Automation",
    href: "/automation",
    icon: Zap,
  },
  {
    name: "Settings",
    href: "/settings",
    icon: Settings,
  },
];

export const Sidebar: React.FC = () => {
  const pathname = usePathname();
  const [openSubmenu, setOpenSubmenu] = useState<string | null>("Infrastructure");

  const toggleSubmenu = (name: string) => {
    setOpenSubmenu(openSubmenu === name ? null : name);
  };

  return (
    <aside className="fixed left-0 top-0 z-40 flex h-screen w-64 flex-col border-r border-[#262626] dark:border-[#262626] light:border-[#e5e7eb] bg-[#121212] dark:bg-[#121212] light:bg-[#ffffff] text-[#ededed] dark:text-[#ededed] light:text-[#111827]">
      {/* Brand Header */}
      <div className="flex h-16 items-center gap-3 border-b border-[#262626] dark:border-[#262626] light:border-[#e5e7eb] px-6">
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-emerald-500 text-zinc-950 font-bold shadow-sm">
          <Cloud className="h-4 w-4" />
        </div>
        <div>
          <h1 className="text-xs font-bold tracking-wider text-[#ededed] dark:text-[#ededed] light:text-[#111827] uppercase">
            Caelus Cloud
          </h1>
          <p className="text-[10px] text-[#707070] dark:text-[#707070] light:text-[#9ca3af] font-mono">Control Panel</p>
        </div>
      </div>

      {/* Navigation Links */}
      <div className="flex-1 overflow-y-auto px-3 py-4 space-y-1">
        <p className="px-3 pb-2 text-[10px] font-semibold uppercase tracking-wider text-[#707070] dark:text-[#707070] light:text-[#9ca3af]">
          Platform Menu
        </p>

        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = pathname === item.href || Boolean(item.children?.some((c) => pathname.startsWith(c.href)));
          const hasChildren = Boolean(item.children && item.children.length > 0);
          const isSubmenuOpen = openSubmenu === item.name || isActive;

          if (hasChildren) {
            return (
              <div key={item.name} className="space-y-1">
                <button
                  type="button"
                  onClick={() => toggleSubmenu(item.name)}
                  className={cn(
                    "w-full flex items-center justify-between px-3 py-2 text-xs font-medium rounded-lg transition-colors cursor-pointer",
                    isActive
                      ? "bg-[#1f1f1f] text-emerald-400 border border-[#2e2e2e]"
                      : "text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#1a1a1a]"
                  )}
                >
                  <div className="flex items-center gap-2.5">
                    <Icon className="h-4 w-4" />
                    <span>{item.name}</span>
                  </div>
                  <ChevronDown
                    className={cn(
                      "h-3.5 w-3.5 transition-transform duration-200",
                      isSubmenuOpen && "rotate-180"
                    )}
                  />
                </button>

                {isSubmenuOpen && (
                  <div className="pl-6 pr-1 py-1 space-y-1">
                    {item.children?.map((subItem) => {
                      const SubIcon = subItem.icon;
                      const isSubActive = pathname === subItem.href;
                      return (
                        <Link
                          key={subItem.href}
                          href={subItem.href}
                          className={cn(
                            "flex items-center gap-2.5 px-3 py-1.5 text-xs font-medium rounded-md transition-colors",
                            isSubActive
                              ? "bg-emerald-500/10 text-emerald-400 font-semibold border-l-2 border-emerald-500 pl-2.5"
                              : "text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#1a1a1a]"
                          )}
                        >
                          <SubIcon className="h-3.5 w-3.5" />
                          <span>{subItem.name}</span>
                        </Link>
                      );
                    })}
                  </div>
                )}
              </div>
            );
          }

          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center justify-between px-3 py-2 text-xs font-medium rounded-lg transition-colors",
                isActive
                  ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-semibold"
                  : "text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#1a1a1a]"
              )}
            >
              <div className="flex items-center gap-2.5">
                <Icon className="h-4 w-4" />
                <span>{item.name}</span>
              </div>
              {item.badge && (
                <span className="text-[10px] font-semibold px-2 py-0.5 bg-emerald-950/80 text-emerald-400 border border-emerald-800/40 rounded-md">
                  {item.badge}
                </span>
              )}
            </Link>
          );
        })}
      </div>

      {/* Footer Info */}
      <div className="p-4 border-t border-[#262626] dark:border-[#262626] light:border-[#e5e7eb]">
        <div className="rounded-lg border border-[#262626] dark:border-[#262626] light:border-[#e5e7eb] bg-[#171717] dark:bg-[#171717] light:bg-[#f9fafb] p-3 text-xs">
          <div className="flex items-center justify-between font-medium mb-1">
            <span className="text-[#a1a1a1]">Platform</span>
            <span className={AppText.categoryTag}>Self-Hosted</span>
          </div>
          <p className="text-[10px] text-[#707070] font-mono">Node ID: caelus-core-01</p>
        </div>
      </div>
    </aside>
  );
};
