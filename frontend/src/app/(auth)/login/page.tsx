"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Cloud, Lock, Mail, Eye, EyeOff, ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { authService } from "@/services/auth.service";
import { useAuthStore } from "@/stores/useAuthStore";
import { AppContainers, AppText } from "@/core/theme";

export default function LoginPage() {
  const router = useRouter();
  const { setAuth, isAuthenticated, initialize } = useAuthStore();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
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
    setIsLoading(true);

    try {
      const data = await authService.login({ email, password });
      setAuth(data);
      router.push("/overview");
    } catch (err: any) {
      setError(
        err.response?.data?.message ||
        err.response?.data?.errors ||
        "Failed to log in. Please check your email and password."
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
          <p className={AppText.subtitle}>Unified Infrastructure & VPS Control Panel</p>
        </div>

        <Card className="shadow-lg">
          <CardHeader className="text-center pb-2">
            <CardTitle className="text-base">Sign In to Your Account</CardTitle>
            <CardDescription>Enter your email and password to access the control panel</CardDescription>
          </CardHeader>

          <CardContent className="pt-4">
            {error && (
              <div className="mb-4 rounded-lg border border-rose-800/60 bg-rose-950/40 p-3 text-xs text-rose-300">
                {error}
              </div>
            )}

            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-1.5">
                <label htmlFor="login-email" className={AppText.label}>
                  Email Address
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-[#707070]">
                    <Mail className="h-4 w-4" />
                  </div>
                  <input
                    id="login-email"
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
                <label htmlFor="login-password" className={AppText.label}>
                  Password
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-[#707070]">
                    <Lock className="h-4 w-4" />
                  </div>
                  <input
                    id="login-password"
                    type={showPassword ? "text" : "password"}
                    required
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="Enter your password"
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

              <Button type="submit" className="w-full mt-2" isLoading={isLoading}>
                <span>Sign In to Dashboard</span>
                <ArrowRight className="h-4 w-4 ml-1" />
              </Button>
            </form>

            <div className="mt-6 text-center text-xs text-[#a1a1a1] border-t border-[#262626] dark:border-[#262626] light:border-[#e5e7eb] pt-4">
              Don&apos;t have an account?{" "}
              <Link href="/register" className="font-semibold text-emerald-400 hover:text-emerald-300 hover:underline">
                Create New Account
              </Link>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
