"use client";

import React, { useState, useEffect } from "react";
import { User, Lock, Shield, Key, Check, AlertCircle, RefreshCw } from "lucide-react";
import { AppTheme } from "@/core/theme";
import { settingsService } from "@/services/settings.service";
import { UserProfile } from "@/types/settings";

export const ProfileTab: React.FC = () => {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [fullName, setFullName] = useState("");
  const [avatarUrl, setAvatarUrl] = useState("");
  const [profileSuccess, setProfileSuccess] = useState<string | null>(null);
  const [profileError, setProfileError] = useState<string | null>(null);
  const [isSavingProfile, setIsSavingProfile] = useState(false);

  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [passwordSuccess, setPasswordSuccess] = useState<string | null>(null);
  const [passwordError, setPasswordError] = useState<string | null>(null);
  const [isChangingPassword, setIsChangingPassword] = useState(false);

  const fetchProfile = async () => {
    try {
      setIsLoading(true);
      const data = await settingsService.getProfile();
      setProfile(data);
      setFullName(data.full_name || "");
      setAvatarUrl(data.avatar_url || "");
    } catch (err: any) {
      setProfileError("Failed to load user profile");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchProfile();
  }, []);

  const handleUpdateProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setIsSavingProfile(true);
      setProfileSuccess(null);
      setProfileError(null);
      const updated = await settingsService.updateProfile({
        full_name: fullName,
        avatar_url: avatarUrl || undefined,
      });
      setProfile(updated);
      setProfileSuccess("Profile updated successfully");
      setTimeout(() => setProfileSuccess(null), 3000);
    } catch (err: any) {
      setProfileError(err.response?.data?.message || "Failed to update profile");
    } finally {
      setIsSavingProfile(false);
    }
  };

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    if (newPassword !== confirmPassword) {
      setPasswordError("New password confirmation does not match");
      return;
    }
    if (newPassword.length < 8) {
      setPasswordError("New password must be at least 8 characters");
      return;
    }

    try {
      setIsChangingPassword(true);
      setPasswordSuccess(null);
      setPasswordError(null);
      await settingsService.changePassword({
        old_password: oldPassword,
        new_password: newPassword,
      });
      setPasswordSuccess("Password changed successfully");
      setOldPassword("");
      setNewPassword("");
      setConfirmPassword("");
      setTimeout(() => setPasswordSuccess(null), 3000);
    } catch (err: any) {
      setPasswordError(err.response?.data?.message || "Failed to change password");
    } finally {
      setIsChangingPassword(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-16 text-zinc-500">
        <RefreshCw className="h-5 w-5 animate-spin mr-2" />
        <span className="text-sm">Loading profile settings...</span>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className={AppTheme.containers.card}>
        <div className="border-b border-[#262626] pb-4 mb-5 flex items-center gap-3">
          <div className="p-2 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400">
            <User className="h-5 w-5" />
          </div>
          <div>
            <h3 className="text-sm font-semibold text-zinc-100">Personal Information</h3>
            <p className="text-xs text-zinc-400">Manage your account name and identity details</p>
          </div>
        </div>

        {profileSuccess && (
          <div className="mb-4 p-3 rounded-lg bg-emerald-950/40 border border-emerald-500/30 text-emerald-400 text-xs flex items-center gap-2">
            <Check className="h-4 w-4 shrink-0" />
            <span>{profileSuccess}</span>
          </div>
        )}

        {profileError && (
          <div className="mb-4 p-3 rounded-lg bg-rose-950/40 border border-rose-500/30 text-rose-400 text-xs flex items-center gap-2">
            <AlertCircle className="h-4 w-4 shrink-0" />
            <span>{profileError}</span>
          </div>
        )}

        <form onSubmit={handleUpdateProfile} className="space-y-4 max-w-xl">
          <div>
            <label className="block text-xs font-medium text-zinc-300 mb-1.5">Email Address</label>
            <input
              type="email"
              disabled
              value={profile?.email || ""}
              className="w-full bg-[#181818] border border-[#2e2e2e] text-zinc-400 text-xs rounded-lg px-3 py-2 cursor-not-allowed"
            />
            <p className="text-[11px] text-zinc-400 mt-1">Primary login email cannot be changed directly</p>
          </div>

          <div>
            <label className="block text-xs font-medium text-zinc-300 mb-1.5">Full Name</label>
            <input
              type="text"
              required
              value={fullName}
              onChange={(e) => setFullName(e.target.value)}
              placeholder="e.g. John Doe"
              className="w-full bg-[#141414] border border-[#2e2e2e] text-zinc-200 text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-emerald-500 transition-colors"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-zinc-300 mb-1.5">Avatar URL (Optional)</label>
            <input
              type="url"
              value={avatarUrl}
              onChange={(e) => setAvatarUrl(e.target.value)}
              placeholder="https://example.com/avatar.png"
              className="w-full bg-[#141414] border border-[#2e2e2e] text-zinc-200 text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-emerald-500 transition-colors"
            />
          </div>

          <button
            type="submit"
            disabled={isSavingProfile}
            className="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-zinc-950 font-semibold text-xs rounded-lg transition-colors flex items-center gap-2 cursor-pointer disabled:opacity-50"
          >
            {isSavingProfile && <RefreshCw className="h-3.5 w-3.5 animate-spin" />}
            <span>Save Changes</span>
          </button>
        </form>
      </div>

      <div className={AppTheme.containers.card}>
        <div className="border-b border-[#262626] pb-4 mb-5 flex items-center gap-3">
          <div className="p-2 rounded-lg bg-amber-500/10 border border-amber-500/20 text-amber-400">
            <Lock className="h-5 w-5" />
          </div>
          <div>
            <h3 className="text-sm font-semibold text-zinc-100">Security & Password</h3>
            <p className="text-xs text-zinc-400">Update your password to protect your account</p>
          </div>
        </div>

        {passwordSuccess && (
          <div className="mb-4 p-3 rounded-lg bg-emerald-950/40 border border-emerald-500/30 text-emerald-400 text-xs flex items-center gap-2">
            <Check className="h-4 w-4 shrink-0" />
            <span>{passwordSuccess}</span>
          </div>
        )}

        {passwordError && (
          <div className="mb-4 p-3 rounded-lg bg-rose-950/40 border border-rose-500/30 text-rose-400 text-xs flex items-center gap-2">
            <AlertCircle className="h-4 w-4 shrink-0" />
            <span>{passwordError}</span>
          </div>
        )}

        <form onSubmit={handleChangePassword} className="space-y-4 max-w-xl">
          <div>
            <label className="block text-xs font-medium text-zinc-300 mb-1.5">Current Password</label>
            <input
              type="password"
              required
              value={oldPassword}
              onChange={(e) => setOldPassword(e.target.value)}
              placeholder="Enter current password"
              className="w-full bg-[#141414] border border-[#2e2e2e] text-zinc-200 text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-emerald-500 transition-colors"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-zinc-300 mb-1.5">New Password</label>
            <input
              type="password"
              required
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              placeholder="Minimum 8 characters"
              className="w-full bg-[#141414] border border-[#2e2e2e] text-zinc-200 text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-emerald-500 transition-colors"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-zinc-300 mb-1.5">Confirm New Password</label>
            <input
              type="password"
              required
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder="Repeat new password"
              className="w-full bg-[#141414] border border-[#2e2e2e] text-zinc-200 text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-emerald-500 transition-colors"
            />
          </div>

          <button
            type="submit"
            disabled={isChangingPassword}
            className="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-200 font-medium text-xs rounded-lg border border-zinc-700 transition-colors flex items-center gap-2 cursor-pointer disabled:opacity-50"
          >
            {isChangingPassword && <RefreshCw className="h-3.5 w-3.5 animate-spin" />}
            <Key className="h-3.5 w-3.5" />
            <span>Update Password</span>
          </button>
        </form>
      </div>

      <div className={AppTheme.containers.card}>
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-sky-500/10 border border-sky-500/20 text-sky-400">
              <Shield className="h-5 w-5" />
            </div>
            <div>
              <h3 className="text-sm font-semibold text-zinc-100">Two-Factor Authentication (2FA)</h3>
              <p className="text-xs text-zinc-400 mt-0.5">
                Add an extra layer of protection to your account using an authenticator app (TOTP)
              </p>
            </div>
          </div>
          <span className="px-2.5 py-1 rounded-full text-[11px] font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            JWT Session Protected
          </span>
        </div>
      </div>
    </div>
  );
};
