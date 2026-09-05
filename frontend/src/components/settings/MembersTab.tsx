"use client";

import React, { useState, useEffect } from "react";
import { Users, UserPlus, Mail, Shield, Trash2, Check, AlertCircle, RefreshCw, X } from "lucide-react";
import { AppTheme } from "@/core/theme";
import { settingsService } from "@/services/settings.service";
import { OrganizationMember, OrganizationInvitation, OrganizationRole } from "@/types/settings";
import { useRoleGuard } from "@/hooks/useRoleGuard";

export const MembersTab: React.FC = () => {
  const { canManageTeam, canChangeRoles } = useRoleGuard();
  const [members, setMembers] = useState<OrganizationMember[]>([]);
  const [invitations, setInvitations] = useState<OrganizationInvitation[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

  const [isInviteModalOpen, setIsInviteModalOpen] = useState(false);
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteRole, setInviteRole] = useState<OrganizationRole>("member");
  const [isInviting, setIsInviting] = useState(false);
  const [inviteModalError, setInviteModalError] = useState<string | null>(null);

  const fetchMembers = async () => {
    try {
      setIsLoading(true);
      const data = await settingsService.listMembers();
      setMembers(data.members || []);
      setInvitations(data.invitations || []);
    } catch (err: any) {
      setErrorMsg("Failed to load team members list");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchMembers();
  }, []);

  const handleInvite = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setIsInviting(true);
      setInviteModalError(null);
      await settingsService.inviteMember({
        email: inviteEmail,
        role: inviteRole,
      });
      setIsInviteModalOpen(false);
      setInviteEmail("");
      setInviteRole("member");
      setSuccessMsg("Team member invitation sent successfully");
      setTimeout(() => setSuccessMsg(null), 3000);
      fetchMembers();
    } catch (err: any) {
      setInviteModalError(err.response?.data?.message || "Failed to send invitation");
    } finally {
      setIsInviting(false);
    }
  };

  const handleRoleChange = async (userId: string, newRole: OrganizationRole) => {
    try {
      await settingsService.updateMemberRole(userId, newRole);
      setSuccessMsg("Member role updated successfully");
      setTimeout(() => setSuccessMsg(null), 3000);
      fetchMembers();
    } catch (err: any) {
      setErrorMsg(err.response?.data?.message || "Failed to update role");
      setTimeout(() => setErrorMsg(null), 3000);
    }
  };

  const handleRemoveMember = async (userId: string, name: string) => {
    if (!confirm(`Remove ${name} from this organization?`)) return;
    try {
      await settingsService.removeMember(userId);
      setSuccessMsg("Member removed successfully");
      setTimeout(() => setSuccessMsg(null), 3000);
      fetchMembers();
    } catch (err: any) {
      setErrorMsg(err.response?.data?.message || "Failed to remove member");
      setTimeout(() => setErrorMsg(null), 3000);
    }
  };

  const handleDeleteInvitation = async (invId: string) => {
    try {
      await settingsService.deleteInvitation(invId);
      setSuccessMsg("Invitation cancelled successfully");
      setTimeout(() => setSuccessMsg(null), 3000);
      fetchMembers();
    } catch (err: any) {
      setErrorMsg("Failed to cancel invitation");
      setTimeout(() => setErrorMsg(null), 3000);
    }
  };

  const getRoleBadge = (role: OrganizationRole) => {
    switch (role) {
      case "owner":
        return <span className="px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/20">Owner</span>;
      case "admin":
        return <span className="px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">Admin</span>;
      case "member":
        return <span className="px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-sky-500/10 text-sky-400 border border-sky-500/20">Member</span>;
      case "viewer":
        return <span className="px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-zinc-500/10 text-zinc-400 border border-zinc-500/20">Viewer</span>;
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-16 text-zinc-500">
        <RefreshCw className="h-5 w-5 animate-spin mr-2" />
        <span className="text-sm">Loading organization members...</span>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h3 className="text-sm font-semibold text-zinc-100">Team Members & Access Control (RBAC)</h3>
          <p className="text-xs text-zinc-400 mt-0.5">Manage collaborators who have access to this organization workspace</p>
        </div>
        {canManageTeam && (
          <button
            type="button"
            onClick={() => setIsInviteModalOpen(true)}
            className="px-3.5 py-2 bg-emerald-600 hover:bg-emerald-500 text-zinc-950 font-semibold text-xs rounded-lg transition-colors flex items-center gap-2 cursor-pointer shrink-0"
          >
            <UserPlus className="h-4 w-4" />
            <span>Invite Member</span>
          </button>
        )}
      </div>

      {successMsg && (
        <div className="p-3 rounded-lg bg-emerald-950/40 border border-emerald-500/30 text-emerald-400 text-xs flex items-center gap-2">
          <Check className="h-4 w-4 shrink-0" />
          <span>{successMsg}</span>
        </div>
      )}

      {errorMsg && (
        <div className="p-3 rounded-lg bg-rose-950/40 border border-rose-500/30 text-rose-400 text-xs flex items-center gap-2">
          <AlertCircle className="h-4 w-4 shrink-0" />
          <span>{errorMsg}</span>
        </div>
      )}

      <div className={`${AppTheme.containers.card} overflow-hidden p-0`}>
        <div className="p-4 border-b border-[#262626] bg-[#161616]">
          <h4 className="text-xs font-semibold text-zinc-200 uppercase tracking-wider">Active Members ({members.length})</h4>
        </div>
        <div className="divide-y divide-[#222222]">
          {members.map((member) => (
            <div key={member.id} className="p-4 flex items-center justify-between gap-4 hover:bg-[#161616]/50 transition-colors">
              <div className="flex items-center gap-3">
                <div className="h-9 w-9 rounded-full bg-zinc-800 border border-zinc-700 flex items-center justify-center text-xs font-bold text-emerald-400">
                  {member.user?.full_name?.charAt(0).toUpperCase() || "U"}
                </div>
                <div>
                  <p className="text-xs font-medium text-zinc-200">{member.user?.full_name || "User"}</p>
                  <p className="text-[11px] text-zinc-400">{member.user?.email}</p>
                </div>
              </div>

              <div className="flex items-center gap-3">
                {!canChangeRoles || member.role === "owner" ? (
                  getRoleBadge(member.role)
                ) : (
                  <select
                    value={member.role}
                    onChange={(e) => handleRoleChange(member.user_id, e.target.value as OrganizationRole)}
                    className="bg-[#141414] border border-[#2e2e2e] text-zinc-300 text-xs rounded-lg px-2.5 py-1 focus:outline-none focus:border-emerald-500 cursor-pointer"
                  >
                    <option value="admin">Admin</option>
                    <option value="member">Member</option>
                    <option value="viewer">Viewer</option>
                  </select>
                )}

                {canManageTeam && member.role !== "owner" && (
                  <button
                    type="button"
                    onClick={() => handleRemoveMember(member.user_id, member.user?.full_name || member.user?.email || "")}
                    className="p-1.5 rounded-lg text-zinc-400 hover:text-rose-400 hover:bg-rose-500/10 transition-colors cursor-pointer"
                    title="Remove Member"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>

      {invitations.length > 0 && (
        <div className={`${AppTheme.containers.card} overflow-hidden p-0`}>
          <div className="p-4 border-b border-[#262626] bg-[#161616]">
            <h4 className="text-xs font-semibold text-amber-400 uppercase tracking-wider">
              Pending Invitations ({invitations.length})
            </h4>
          </div>
          <div className="divide-y divide-[#222222]">
            {invitations.map((inv) => (
              <div key={inv.id} className="p-4 flex items-center justify-between gap-4 hover:bg-[#161616]/50 transition-colors">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-lg bg-amber-500/10 border border-amber-500/20 text-amber-400">
                    <Mail className="h-4 w-4" />
                  </div>
                  <div>
                    <p className="text-xs font-medium text-zinc-200">{inv.email}</p>
                    <p className="text-[11px] text-zinc-400">Expires: {new Date(inv.expires_at).toLocaleDateString()}</p>
                  </div>
                </div>

                <div className="flex items-center gap-3">
                  {getRoleBadge(inv.role)}
                  {canManageTeam && (
                    <button
                      type="button"
                      onClick={() => handleDeleteInvitation(inv.id)}
                      className="p-1.5 rounded-lg text-zinc-400 hover:text-rose-400 hover:bg-rose-500/10 transition-colors cursor-pointer"
                      title="Cancel Invitation"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {isInviteModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-xs p-4">
          <div className="bg-[#141414] border border-[#2e2e2e] rounded-xl w-full max-w-md p-6 shadow-2xl space-y-4">
            <div className="flex items-center justify-between border-b border-[#262626] pb-3">
              <h3 className="text-sm font-semibold text-zinc-100 flex items-center gap-2">
                <UserPlus className="h-4 w-4 text-emerald-400" />
                Invite New Team Member
              </h3>
              <button
                type="button"
                onClick={() => setIsInviteModalOpen(false)}
                className="text-zinc-400 hover:text-zinc-200 p-1 rounded-lg transition-colors cursor-pointer"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            {inviteModalError && (
              <div className="p-3 rounded-lg bg-rose-950/40 border border-rose-500/30 text-rose-400 text-xs flex items-center gap-2">
                <AlertCircle className="h-4 w-4 shrink-0" />
                <span>{inviteModalError}</span>
              </div>
            )}

            <form onSubmit={handleInvite} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-zinc-300 mb-1.5">Invitee Email Address</label>
                <input
                  type="email"
                  required
                  value={inviteEmail}
                  onChange={(e) => setInviteEmail(e.target.value)}
                  placeholder="colleague@company.com"
                  className="w-full bg-[#181818] border border-[#2e2e2e] text-zinc-200 text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-emerald-500 transition-colors"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-zinc-300 mb-1.5">Role (RBAC Role)</label>
                <select
                  value={inviteRole}
                  onChange={(e) => setInviteRole(e.target.value as OrganizationRole)}
                  className="w-full bg-[#181818] border border-[#2e2e2e] text-zinc-200 text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-emerald-500 cursor-pointer"
                >
                  <option value="admin">Admin — Full access to resources & settings</option>
                  <option value="member">Member — Can manage VPS, Storage & Deployments</option>
                  <option value="viewer">Viewer — Read-only access to metrics & audit logs</option>
                </select>
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setIsInviteModalOpen(false)}
                  className="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-xs rounded-lg transition-colors cursor-pointer"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isInviting}
                  className="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-zinc-950 font-semibold text-xs rounded-lg transition-colors flex items-center gap-2 cursor-pointer disabled:opacity-50"
                >
                  {isInviting && <RefreshCw className="h-3.5 w-3.5 animate-spin" />}
                  <span>Send Invitation</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
