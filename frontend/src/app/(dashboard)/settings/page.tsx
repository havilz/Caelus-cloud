"use client";

import React, { useState } from "react";
import { User, Building2, Users, Key, Send, History } from "lucide-react";
import { AppTheme } from "@/core/theme";
import { ProfileTab } from "@/components/settings/ProfileTab";
import { OrganizationTab } from "@/components/settings/OrganizationTab";
import { MembersTab } from "@/components/settings/MembersTab";
import { ApiKeysTab } from "@/components/settings/ApiKeysTab";
import { WebhooksTab } from "@/components/settings/WebhooksTab";
import { AuditLogsTab } from "@/components/settings/AuditLogsTab";

type SettingsTab = "profile" | "organization" | "members" | "api-keys" | "webhooks" | "audit-logs";

export default function SettingsPage() {
  const [activeTab, setActiveTab] = useState<SettingsTab>("profile");

  const tabs: { id: SettingsTab; label: string; icon: React.ElementType }[] = [
    { id: "profile", label: "Profil & Keamanan", icon: User },
    { id: "organization", label: "Organisasi", icon: Building2 },
    { id: "members", label: "Anggota & Role", icon: Users },
    { id: "api-keys", label: "API Keys (PAT)", icon: Key },
    { id: "webhooks", label: "Webhooks Notifikasi", icon: Send },
    { id: "audit-logs", label: "Audit Logs", icon: History },
  ];

  return (
    <div className={AppTheme.containers.pageWrapper}>
      {}
      <div className="flex flex-col gap-1 pb-6 border-b border-[#262626]">
        <h1 className="text-xl font-bold tracking-tight text-[#ededed]">Pengaturan Sistem & Workspace</h1>
        <p className="text-xs text-[#707070]">
          Kelola profil pengguna, keamanan akun, struktur organisasi, hak akses tim, token API, dan log audit platform
        </p>
      </div>

      {}
      <div className="flex items-center gap-1 border-b border-[#262626] overflow-x-auto no-scrollbar py-2">
        {tabs.map((tab) => {
          const Icon = tab.icon;
          const isActive = activeTab === tab.id;
          return (
            <button
              key={tab.id}
              type="button"
              onClick={() => setActiveTab(tab.id)}
              className={`flex items-center gap-2 px-3.5 py-2 text-xs font-medium rounded-lg transition-all cursor-pointer whitespace-nowrap ${
                isActive
                  ? "bg-[#1f1f1f] text-emerald-400 border border-[#2e2e2e] shadow-xs"
                  : "text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#181818]"
              }`}
            >
              <Icon className="h-4 w-4" />
              <span>{tab.label}</span>
            </button>
          );
        })}
      </div>

      {}
      <div className="pt-2">
        {activeTab === "profile" && <ProfileTab />}
        {activeTab === "organization" && <OrganizationTab />}
        {activeTab === "members" && <MembersTab />}
        {activeTab === "api-keys" && <ApiKeysTab />}
        {activeTab === "webhooks" && <WebhooksTab />}
        {activeTab === "audit-logs" && <AuditLogsTab />}
      </div>
    </div>
  );
}
