"use client";

import React, { useState, useMemo, useRef, useEffect } from "react";
import {
  Terminal,
  Search,
  Trash2,
  Copy,
  Check,
  Download,
  ArrowDown,
  Filter,
} from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { LogEntry } from "@/types/monitoring";

interface LogViewerProps {
  logs: LogEntry[];
  onClearLogs?: () => void;
  title?: string;
  isStreaming?: boolean;
}

export const LogViewer: React.FC<LogViewerProps> = ({
  logs,
  onClearLogs,
  title = "System Console & Log Stream",
  isStreaming = true,
}) => {
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedLevel, setSelectedLevel] = useState<string>("ALL");
  const [autoScroll, setAutoScroll] = useState(true);
  const [copied, setCopied] = useState(false);
  const terminalEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (autoScroll && terminalEndRef.current) {
      terminalEndRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [logs, autoScroll]);

  const filteredLogs = useMemo(() => {
    return logs.filter((log) => {
      const matchLevel =
        selectedLevel === "ALL" || log.level === selectedLevel;
      const matchQuery =
        searchQuery === "" ||
        log.line.toLowerCase().includes(searchQuery.toLowerCase()) ||
        log.service?.toLowerCase().includes(searchQuery.toLowerCase());
      return matchLevel && matchQuery;
    });
  }, [logs, selectedLevel, searchQuery]);

  const handleCopy = () => {
    const text = filteredLogs
      .map((l) => `[${l.timestamp}] [${l.level}] ${l.service ? `(${l.service}) ` : ""}${l.line}`)
      .join("\n");
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleDownload = () => {
    const text = filteredLogs
      .map((l) => `[${l.timestamp}] [${l.level}] ${l.service ? `(${l.service}) ` : ""}${l.line}`)
      .join("\n");
    const blob = new Blob([text], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `caelus-logs-${new Date().toISOString().slice(0, 19)}.txt`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const getLevelBadge = (level: string) => {
    switch (level) {
      case "ERROR":
        return <span className="text-rose-400 font-bold">[ERROR]</span>;
      case "WARN":
        return <span className="text-amber-400 font-bold">[WARN]</span>;
      case "DEBUG":
        return <span className="text-purple-400 font-bold">[DEBUG]</span>;
      default:
        return <span className="text-emerald-400 font-bold">[INFO]</span>;
    }
  };

  return (
    <Card className="border-[#262626] bg-[#0d0d0d] overflow-hidden flex flex-col h-[520px]">
      <CardHeader className="p-3.5 border-b border-[#1f1f1f] bg-[#121212]">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            <Terminal className="h-4 w-4 text-emerald-400" />
            <CardTitle className="text-xs font-semibold text-[#ededed]">{title}</CardTitle>
            {isStreaming && (
              <span className="flex items-center gap-1 text-[10px] text-emerald-400 font-mono bg-emerald-500/10 px-2 py-0.5 rounded border border-emerald-500/20">
                <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse" />
                STREAMING
              </span>
            )}
            <span className="text-[11px] text-[#707070] font-mono">
              ({filteredLogs.length} / {logs.length} lines)
            </span>
          </div>

          <div className="flex items-center gap-1.5">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setAutoScroll(!autoScroll)}
              className={`h-7 px-2.5 text-[11px] ${
                autoScroll ? "text-emerald-400 border-emerald-500/30 bg-emerald-950/20" : "text-[#a1a1a1]"
              }`}
            >
              <ArrowDown className="h-3 w-3 mr-1" />
              Auto-scroll
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleCopy}
              className="h-7 px-2.5 text-[11px] text-[#a1a1a1]"
            >
              {copied ? <Check className="h-3 w-3 mr-1 text-emerald-400" /> : <Copy className="h-3 w-3 mr-1" />}
              {copied ? "Tersalin" : "Copy"}
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleDownload}
              className="h-7 px-2.5 text-[11px] text-[#a1a1a1]"
            >
              <Download className="h-3 w-3 mr-1" />
              Export
            </Button>
            {onClearLogs && (
              <Button
                variant="outline"
                size="sm"
                onClick={onClearLogs}
                className="h-7 px-2.5 text-[11px] text-rose-400 hover:bg-rose-950/30 hover:border-rose-900/50"
              >
                <Trash2 className="h-3 w-3 mr-1" />
                Clear
              </Button>
            )}
          </div>
        </div>

        <div className="flex flex-col sm:flex-row items-center gap-2 pt-2 mt-1">
          <div className="relative flex-1 w-full">
            <Search className="absolute left-2.5 top-2 h-3.5 w-3.5 text-[#707070]" />
            <input
              type="text"
              placeholder="Search log keywords..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full h-7.5 rounded-md border border-[#262626] bg-[#171717] pl-8 pr-3 text-[11px] text-[#ededed] placeholder-[#606060] focus:border-emerald-500 focus:outline-none"
            />
          </div>

          <div className="flex items-center gap-1 bg-[#171717] p-0.5 rounded-md border border-[#262626] self-start sm:self-auto">
            {["ALL", "INFO", "WARN", "ERROR", "DEBUG"].map((lvl) => (
              <button
                key={lvl}
                type="button"
                onClick={() => setSelectedLevel(lvl)}
                className={`px-2 py-0.5 text-[10px] font-bold rounded transition-colors cursor-pointer ${
                  selectedLevel === lvl
                    ? "bg-[#282828] text-emerald-400 shadow-sm"
                    : "text-[#707070] hover:text-[#a1a1a1]"
                }`}
              >
                {lvl}
              </button>
            ))}
          </div>
        </div>
      </CardHeader>

      <CardContent className="p-3 flex-1 overflow-y-auto font-mono text-[11px] leading-relaxed text-[#c0c0c0] select-text">
        {filteredLogs.length === 0 ? (
          <div className="flex h-full items-center justify-center text-center text-[#555555]">
            <div>
              <Filter className="h-6 w-6 mx-auto mb-2 opacity-40" />
              <p>No log entries match the selected filter.</p>
            </div>
          </div>
        ) : (
          <div className="space-y-1">
            {filteredLogs.map((log) => (
              <div
                key={log.id}
                className="flex items-start gap-2 hover:bg-[#151515] px-1.5 py-0.5 rounded transition-colors"
              >
                <span className="text-[#555555] shrink-0 select-none">
                  {new Date(log.timestamp).toLocaleTimeString("en-US", {
                    hour: "2-digit",
                    minute: "2-digit",
                    second: "2-digit",
                    fractionalSecondDigits: 3,
                  })}
                </span>
                <span className="shrink-0">{getLevelBadge(log.level)}</span>
                {log.service && (
                  <span className="text-[#888888] shrink-0 font-semibold">[{log.service}]</span>
                )}
                <span className="break-all text-[#d8d8d8]">{log.line}</span>
              </div>
            ))}
            <div ref={terminalEndRef} />
          </div>
        )}
      </CardContent>
    </Card>
  );
};
