import React, { useState, useEffect } from "react";
import { Breadcrumbs } from "./Breadcrumbs";
import { useAuthStore } from "@/stores/useAuthStore";
import { useRouter } from "next/navigation";
import { LogOut, Building2, ChevronDown, Bell } from "lucide-react";
import { AlertDrawer } from "@/components/monitoring/AlertDrawer";
import { monitoringService } from "@/services/monitoring.service";

export const Header: React.FC = () => {
  const { user, organization, logout } = useAuthStore();
  const router = useRouter();
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const [isAlertDrawerOpen, setIsAlertDrawerOpen] = useState(false);
  const [activeAlertsCount, setActiveAlertsCount] = useState<number>(0);

  const fetchActiveAlertsCount = async () => {
    try {
      const res = await monitoringService.listAlerts("active", 1, 1);
      setActiveAlertsCount(res.meta?.total_items || 0);
    } catch {
      
    }
  };

  useEffect(() => {
    fetchActiveAlertsCount();
    const interval = setInterval(fetchActiveAlertsCount, 30000);
    return () => clearInterval(interval);
  }, []);

  const handleLogout = () => {
    logout();
    router.push("/login");
  };

  return (
    <>
      <header className="sticky top-0 z-30 flex h-16 w-full items-center justify-between border-b border-[#262626] bg-[#121212]/90 px-6 backdrop-blur-sm">
        <div className="flex items-center gap-4">
          <Breadcrumbs />
        </div>

        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={() => setIsAlertDrawerOpen(true)}
            className="relative flex h-9 w-9 items-center justify-center rounded-lg border border-[#262626] bg-[#171717] text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#222222] transition-colors cursor-pointer"
            aria-label="Buka notifikasi alert"
          >
            <Bell className="h-4 w-4" />
            {activeAlertsCount > 0 && (
              <span className="absolute -top-1 -right-1 flex h-4 min-w-4 items-center justify-center rounded-full bg-rose-500 px-1 text-[10px] font-bold text-white shadow-xs">
                {activeAlertsCount > 9 ? "9+" : activeAlertsCount}
              </span>
            )}
          </button>

          <div className="hidden sm:flex items-center gap-2 rounded-md border border-emerald-800/40 bg-emerald-950/40 px-2.5 py-1 text-xs text-emerald-400">
            <span className="relative flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
              <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
            </span>
            <span className="font-medium text-[11px]">System Operational</span>
          </div>

        <div className="relative">
          <button
            type="button"
            onClick={() => setDropdownOpen(!dropdownOpen)}
            className="flex items-center gap-2.5 rounded-lg border border-[#262626] bg-[#171717] px-3 py-1.5 text-xs text-[#ededed] hover:bg-[#222222] transition-colors cursor-pointer"
          >
            <div className="flex h-6 w-6 items-center justify-center rounded-md bg-emerald-500/20 text-emerald-400 font-bold text-xs">
              {user?.full_name?.charAt(0).toUpperCase() || "A"}
            </div>
            <div className="text-left hidden md:block">
              <p className="font-medium leading-tight">{user?.full_name || "User"}</p>
              <p className="text-[10px] text-[#a1a1a1] leading-tight">{organization?.name || "Personal Workspace"}</p>
            </div>
            <ChevronDown className="h-3.5 w-3.5 text-[#707070]" />
          </button>

          {dropdownOpen && (
            <>
              <button
                type="button"
                aria-label="Tutup menu dropdown"
                className="fixed inset-0 z-10 cursor-default bg-transparent border-0 w-full h-full"
                onClick={() => setDropdownOpen(false)}
              />
              <div className="absolute right-0 mt-2 w-56 rounded-xl border border-[#2e2e2e] bg-[#171717] p-2 shadow-2xl z-20 animate-in fade-in zoom-in-95 duration-100">
                <div className="px-3 py-2 border-b border-[#262626]">
                  <p className="text-xs font-semibold text-[#ededed]">
                    {user?.full_name}
                  </p>
                  <p className="text-[11px] text-[#a1a1a1] truncate">{user?.email}</p>
                  <div className="mt-1.5 flex items-center gap-1.5 text-[11px] text-emerald-400 font-medium">
                    <Building2 className="h-3 w-3" />
                    <span>{organization?.name || "Workspace"}</span>
                    <span className="ml-auto text-[10px] uppercase px-1.5 py-0.2 bg-emerald-950 text-emerald-400 rounded border border-emerald-800/50">
                      {organization?.role || "Owner"}
                    </span>
                  </div>
                </div>

                <div className="py-1">
                  <button
                    type="button"
                    onClick={handleLogout}
                    className="w-full flex items-center gap-2 px-3 py-2 text-xs text-rose-400 hover:bg-rose-950/40 rounded-lg transition-colors cursor-pointer"
                  >
                    <LogOut className="h-3.5 w-3.5" />
                    <span>Keluar (Logout)</span>
                  </button>
                </div>
              </div>
            </>
          )}
        </div>
      </div>
    </header>

    <AlertDrawer
      isOpen={isAlertDrawerOpen}
      onClose={() => setIsAlertDrawerOpen(false)}
      onAlertUpdated={fetchActiveAlertsCount}
    />
  </>
);
};
