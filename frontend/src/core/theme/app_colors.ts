
export const AppColors = {
  
  brand: {
    primary: "bg-emerald-500 hover:bg-emerald-400 text-zinc-950 font-semibold shadow-sm",
    primaryLight: "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20",
    accent: "text-emerald-400",
    accentHover: "hover:text-emerald-300",
    border: "border-emerald-500/30",
    ring: "focus:ring-emerald-500",
  },

  bg: {
    canvas: "bg-[#0f0f0f] dark:bg-[#0f0f0f] light:bg-[#ffffff]",
    surface: "bg-[#171717] dark:bg-[#171717] light:bg-[#f9fafb]",
    surfaceElevated: "bg-[#1f1f1f] dark:bg-[#1f1f1f] light:bg-[#f3f4f6]",
    surfaceSubtle: "bg-[#141414] dark:bg-[#141414] light:bg-[#f4f4f5]",
    surfaceHover: "hover:bg-[#222222] dark:hover:bg-[#222222] light:hover:bg-[#f4f4f5]",
  },

  border: {
    subtle: "border-[#262626] dark:border-[#262626] light:border-[#e5e7eb]",
    strong: "border-[#333333] dark:border-[#333333] light:border-[#d1d5db]",
    hover: "hover:border-[#404040] dark:hover:border-[#404040] light:hover:border-[#9ca3af]",
  },

  text: {
    primary: "text-[#ededed] dark:text-[#ededed] light:text-[#111827]",
    secondary: "text-[#a1a1a1] dark:text-[#a1a1a1] light:text-[#4b5563]",
    muted: "text-[#707070] dark:text-[#707070] light:text-[#9ca3af]",
    inverse: "text-zinc-950 dark:text-zinc-950 light:text-white",
  },

  status: {
    running: {
      bg: "bg-emerald-950/60 dark:bg-emerald-950/60 light:bg-emerald-50",
      text: "text-emerald-400 dark:text-emerald-400 light:text-emerald-700",
      border: "border-emerald-800/40 dark:border-emerald-800/40 light:border-emerald-200",
      dot: "bg-emerald-400",
    },
    stopped: {
      bg: "bg-zinc-900/80 dark:bg-zinc-900/80 light:bg-zinc-100",
      text: "text-zinc-400 dark:text-zinc-400 light:text-zinc-600",
      border: "border-zinc-800 dark:border-zinc-800 light:border-zinc-300",
      dot: "bg-zinc-500",
    },
    restarting: {
      bg: "bg-amber-950/60 dark:bg-amber-950/60 light:bg-amber-50",
      text: "text-amber-400 dark:text-amber-400 light:text-amber-700",
      border: "border-amber-800/40 dark:border-amber-800/40 light:border-amber-200",
      dot: "bg-amber-400",
    },
    danger: {
      bg: "bg-rose-950/60 dark:bg-rose-950/60 light:bg-rose-50",
      text: "text-rose-400 dark:text-rose-400 light:text-rose-700",
      border: "border-rose-800/40 dark:border-rose-800/40 light:border-rose-200",
      dot: "bg-rose-400",
    },
    info: {
      bg: "bg-cyan-950/60 dark:bg-cyan-950/60 light:bg-cyan-50",
      text: "text-cyan-400 dark:text-cyan-400 light:text-cyan-700",
      border: "border-cyan-800/40 dark:border-cyan-800/40 light:border-cyan-200",
      dot: "bg-cyan-400",
    },
  },
} as const;
