"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Cloud, Lock, Mail, User, Building2, Eye, EyeOff, ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { authService } from "@/services/auth.service";
import { useAuthStore } from "@/stores/useAuthStore";
import { AppContainers, AppText } from "@/core/theme";

export default function RegisterPage() {
  const router = useRouter();
  const { setAuth, isAuthenticated, initialize } = useAuthStore();

  const [fullName, setFullName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [orgName, setOrgName] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    initialize();
  }, [initialize]);

  useEffect(() => {
    if (isAuthenticated) {
      router.push("/overview");
    }
  }, [isAuthenticated, router]);

  const handleSubmit = async (e: React.SyntheticEvent) => {
    e.preventDefault();
    setError(null);

    if (password.length < 8) {
      setError("Password must be at least 8 characters");
      return;
    }

    setIsLoading(true);

    try {
      const data = await authService.register({
        full_name: fullName,
        email,
        password,
        organization_name: orgName || undefined,
      });
      setAuth(data);
      router.push("/overview");
    } catch (err: any) {
      setError(
        err.response?.data?.message ||
        err.response?.data?.errors ||
        "Failed to register account. Please try again."
      );
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className={AppContainers.authWrapper}>
      <div className={AppContainers.authCardWidth}>
        <div className="flex flex-col items-center text-center mb-6">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-emerald-500 text-zinc-950 font-bold shadow-sm mb-3">
            <Cloud className="h-5 w-5" />
          </div>
          <h1 className={AppText.h2}>Caelus Cloud</h1>
          <p className={AppText.subtitle}>Initialize New Account & Workspace</p>
        </div>

        <Card className="shadow-lg">
          <CardHeader className="text-center pb-2">
            <CardTitle className="text-base">Create New Account</CardTitle>
            <CardDescription>Manage and monitor all your servers in one unified control panel</CardDescription>
          </CardHeader>

          <CardContent className="pt-4">
            {error && (
              <div className="mb-4 rounded-lg border border-rose-800/60 bg-rose-950/40 p-3 text-xs text-rose-300">
                {error}
              </div>
            )}

            <form onSubmit={handleSubmit} className="space-y-3.5">
              <div className="space-y-1.5">
                <label htmlFor="reg-name" className={AppText.label}>
                  Full Name
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-[#707070]">
                    <User className="h-4 w-4" />
                  </div>
                  <input
                    id="reg-name"
                    type="text"
                    required
                    value={fullName}
                    onChange={(e) => setFullName(e.target.value)}
                    placeholder="Your Full Name"
                    className="w-full rounded-lg border border-[#262626] dark:border-[#262626] light:border-[#d1d5db] bg-[#121212] dark:bg-[#121212] light:bg-[#ffffff] pl-9 pr-3.5 py-2 text-xs text-[#ededed] dark:text-[#ededed] light:text-[#111827] placeholder-[#707070] focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500 transition-colors"
                  />
                </div>
              </div>

              <div className="space-y-1.5">
                <label htmlFor="reg-email" className={AppText.label}>
                  Email Address
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-[#707070]">
                    <Mail className="h-4 w-4" />
                  </div>
                  <input
                    id="reg-email"
                    type="email"
                    required
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="admin@caelus.cloud"
                    className="w-full rounded-lg border border-[#262626] dark:border-[#262626] light:border-[#d1d5db] bg-[#121212] dark:bg-[#121212] light:bg-[#ffffff] pl-9 pr-3.5 py-2 text-xs text-[#ededed] dark:text-[#ededed] light:text-[#111827] placeholder-[#707070] focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500 transition-colors"
                  />
                </div>
              </div>

              <div className="space-y-1.5">
                <label htmlFor="reg-org" className={AppText.label}>
                  Workspace / Organization Name
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-[#707070]">
                    <Building2 className="h-4 w-4" />
                  </div>
                  <input
                    id="reg-org"
                    type="text"
                    value={orgName}
                    onChange={(e) => setOrgName(e.target.value)}
                    placeholder="My Production Workspace (Optional)"
                    className="w-full rounded-lg border border-[#262626] dark:border-[#262626] light:border-[#d1d5db] bg-[#121212] dark:bg-[#121212] light:bg-[#ffffff] pl-9 pr-3.5 py-2 text-xs text-[#ededed] dark:text-[#ededed] light:text-[#111827] placeholder-[#707070] focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500 transition-colors"
                  />
                </div>
              </div>

              <div className="space-y-1.5">
                <label htmlFor="reg-password" className={AppText.label}>
                  Password
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-[#707070]">
                    <Lock className="h-4 w-4" />
                  </div>
                  <input
                    id="reg-password"
                    type={showPassword ? "text" : "password"}
                    required
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="Minimum 8 characters"
                    className="w-full rounded-lg border border-[#262626] dark:border-[#262626] light:border-[#d1d5db] bg-[#121212] dark:bg-[#121212] light:bg-[#ffffff] pl-9 pr-10 py-2 text-xs text-[#ededed] dark:text-[#ededed] light:text-[#111827] placeholder-[#707070] focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500 transition-colors"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute inset-y-0 right-0 flex items-center pr-3 text-[#707070] hover:text-[#ededed] cursor-pointer"
                  >
                    {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </button>
                </div>
              </div>

              <Button type="submit" className="w-full mt-4" isLoading={isLoading}>
                <span>Create Account & Sign In</span>
                <ArrowRight className="h-4 w-4 ml-1" />
              </Button>
            </form>

            <div className="mt-6 text-center text-xs text-[#a1a1a1] border-t border-[#262626] dark:border-[#262626] light:border-[#e5e7eb] pt-4">
              Already have an account?{" "}
              <Link href="/login" className="font-semibold text-emerald-400 hover:text-emerald-300 hover:underline">
                Sign In Here
              </Link>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
