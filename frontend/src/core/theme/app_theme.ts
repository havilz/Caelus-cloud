import { AppColors } from "./app_colors";
import { AppText } from "./app_text";
import { AppContainers } from "./app_containers";

export const AppControls = {
  
  input: "w-full px-3 py-2 rounded-lg bg-[#141414] border border-[#262626] text-sm text-[#ededed] placeholder-[#707070] focus:outline-none focus:border-emerald-500 transition-colors",
  inputMono: "w-full px-2.5 py-1.5 rounded-md bg-[#1c1c1c] border border-[#333333] text-xs text-[#ededed] placeholder-[#707070] focus:outline-none focus:border-emerald-500 font-mono transition-colors",
  
  select: "w-full px-3 py-2 rounded-lg bg-[#141414] border border-[#262626] text-sm text-[#ededed] focus:outline-none focus:border-emerald-500 transition-colors",
  selectSm: "px-2.5 py-1.5 rounded-md bg-[#1c1c1c] border border-[#333333] text-xs text-[#ededed] focus:outline-none focus:border-emerald-500 font-mono transition-colors",

  buttonPrimary: "px-4 py-2 rounded-lg bg-emerald-500 hover:bg-emerald-400 text-zinc-950 text-xs font-semibold shadow-sm transition-all disabled:opacity-50 flex items-center justify-center gap-2",
  buttonSecondary: "px-3.5 py-2 rounded-lg bg-[#1a1a1a] hover:bg-[#222222] border border-[#2e2e2e] text-xs font-medium text-[#ededed] flex items-center justify-center gap-2 transition-colors",
  buttonGhost: "px-3 py-1.5 rounded-lg text-xs font-medium text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#222222] transition-colors flex items-center justify-center gap-1.5",
  buttonAction: "px-2.5 py-1 rounded-lg bg-[#222222] hover:bg-[#2e2e2e] text-xs text-emerald-400 border border-[#333333] flex items-center gap-1 transition-colors",
  
  iconButton: "p-1.5 rounded-lg text-[#707070] hover:text-[#ededed] hover:bg-[#262626] transition-colors",
  iconButtonDanger: "p-1.5 rounded-md text-[#707070] hover:text-rose-400 hover:bg-[#262626] disabled:opacity-30 transition-colors",

  badgeActive: "px-2 py-0.5 rounded-md text-[10px] font-semibold uppercase tracking-wider bg-emerald-950/60 text-emerald-400 border border-emerald-800/40",
  badgeInactive: "px-2 py-0.5 rounded-md text-[10px] font-semibold uppercase tracking-wider bg-zinc-900 text-zinc-400 border border-zinc-800",
  badgeMono: "px-2 py-0.5 rounded-md bg-[#1f1f1f] text-[#a1a1a1] border border-[#2e2e2e] text-[10px] font-mono",
  pillSubtle: "px-2.5 py-1 rounded-md bg-[#1a1a1a] border border-[#2a2a2a] text-xs text-[#ededed] flex items-center gap-1",
  pillMono: "px-2.5 py-1 rounded-md bg-[#1a1a1a] border border-[#2a2a2a] text-xs font-mono text-[#ededed]",

  cardSubtle: "p-3 rounded-lg bg-[#141414] border border-[#262626]",
  cardRow: "p-4 rounded-xl bg-[#141414] border border-[#262626] hover:border-[#333333] transition-all",
  cardRowActive: "p-5 rounded-xl bg-[#141414] border border-[#262626] hover:border-[#333333] transition-all space-y-4",

  codeBox: "p-3 rounded-lg bg-[#0d0d0d] border border-[#222222] text-xs font-mono text-emerald-400 overflow-x-auto",
  
  iconBoxEmerald: "p-2 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400",
  iconBoxCyan: "p-2.5 rounded-lg bg-cyan-500/10 border border-cyan-500/20 text-cyan-400",
  iconBoxPurple: "p-2.5 rounded-lg bg-purple-500/10 border border-purple-500/20 text-purple-400",
  iconBoxAmber: "p-2.5 rounded-lg bg-amber-500/10 border border-amber-500/20 text-amber-400",
} as const;

export const AppTheme = {
  colors: AppColors,
  text: AppText,
  containers: AppContainers,
  controls: AppControls,
  mode: "dark",
  transition: "transition-colors duration-200",
} as const;
