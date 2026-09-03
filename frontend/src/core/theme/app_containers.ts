
export const AppContainers = {
  
  pageWrapper: "max-w-7xl mx-auto space-y-6",
  authWrapper: "min-h-screen w-full flex items-center justify-center p-4 bg-[#0f0f0f]",
  authCardWidth: "w-full max-w-md",

  card: "rounded-xl border border-[#262626] bg-[#171717] text-[#ededed] shadow-sm",
  cardHover: "hover:border-[#383838] transition-colors",
  cardHeader: "p-5 pb-3 border-b border-[#262626]",
  cardContent: "p-5 pt-4",
  cardFooter: "p-5 pt-0 flex items-center",

  modalBackdrop: "fixed inset-0 bg-black/50 backdrop-blur-[2px] transition-opacity z-50 flex items-center justify-center p-4",
  modalDialog: "relative w-full rounded-xl border border-[#2e2e2e] bg-[#171717] text-[#ededed] shadow-2xl z-10 max-h-[90vh] flex flex-col animate-in fade-in zoom-in-95 duration-150",
  modalBodyScroll: "overflow-y-auto flex-1 pr-1 custom-scrollbar",

  metricsGrid: "grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4",
  overviewSplitGrid: "grid grid-cols-1 lg:grid-cols-3 gap-6",
  specsGrid: "grid grid-cols-2 sm:grid-cols-3 gap-4",
} as const;
