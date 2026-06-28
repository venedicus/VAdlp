import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { ActivityPanel } from "./components/ActivityPanel";
import { FormatPickerModal } from "./components/FormatPickerModal";
import { ConfirmModal, Modal, PromptModal } from "./components/Modal";
import { ResizableSplit } from "./components/ResizableSplit";
import { countDepAttention, DependenciesTab, installDependencyWithSave } from "./components/DependenciesTab";
import { useToast } from "./components/Toast";
import { asArray, defaultConfig, defaultSettings, normalizeSettings } from "./lib/defaults";
import { tf } from "./lib/i18nFmt";
import { queueOverallProgress } from "./lib/progressLabels";
import { BASE_WINDOW_HEIGHT, BASE_WINDOW_WIDTH, effectiveUIScale } from "./lib/uiScale";
import { useWindowBounds } from "./hooks/useWindowBounds";
import { AppAPI, eventsOn, waitForWailsRuntime } from "./wailsjs/runtime";
import { BrowserOpenURL, ScreenGetAll, WindowSetSize } from "./wailsjs/runtime/runtime";
import type {
  AppSettingsDTO,
  AppUpdateDTO,
  ConfigDTO,
  DependencyDTO,
  DownloadProgressDTO,
  HealthIssueDTO,
  HistoryItemDTO,
  InstanceDTO,
  LocaleMap,
  ProbeResultDTO,
  QueueTaskDTO,
} from "./types";

const TABS = [
  "download",
  "network",
  "playlist",
  "extras",
  "queue",
  "history",
  "dependencies",
  "settings",
] as const;
type TabId = (typeof TABS)[number];

const TAB_KEYS: Record<TabId, string> = {
  download: "tab.download",
  network: "tab.network",
  playlist: "tab.playlist",
  extras: "tab.extras",
  queue: "tab.queue",
  history: "tab.history",
  dependencies: "tab.dependencies",
  settings: "tab.tools",
};

const AUDIO_FORMATS = ["", "mp3", "m4a", "opus", "wav", "flac", "vorbis", "aac", "alac"];

const COOKIES_BROWSERS = ["chrome", "firefox", "vivaldi", "edge", "brave"] as const;

const UI_SCALES = [
  { value: 0, key: "ui_scale.auto" },
  { value: 0.95, key: "ui_scale.compact" },
  { value: 1.05, key: "ui_scale.comfortable" },
  { value: 1.15, key: "ui_scale.large" },
  { value: 1.25, key: "ui_scale.extra_large" },
] as const;

function Field({
  label,
  children,
  wide,
}: {
  label: string;
  children: React.ReactNode;
  wide?: boolean;
}) {
  return (
    <div className={`form-row${wide ? " form-row-wide" : ""}`}>
      <label>{label}</label>
      {wide ? children : <div>{children}</div>}
    </div>
  );
}

function Check({
  label,
  checked,
  onChange,
}: {
  label: string;
  checked: boolean;
  onChange: (v: boolean) => void;
}) {
  return (
    <label className="check-row">
      <input type="checkbox" checked={checked} onChange={(e) => onChange(e.target.checked)} />
      {label}
    </label>
  );
}

function extractDroppedURL(data: DataTransfer): string {
  const raw = data.getData("text/uri-list") || data.getData("text/plain") || "";
  const line = raw.split(/\r?\n/).find((l) => l.trim() && !l.startsWith("#"));
  return line?.trim() ?? "";
}

function looksLikeDownloadableURL(text: string): boolean {
  if (text.length > 2000 || text.includes("\n")) return false;
  return /^https?:\/\/\S+$/i.test(text.trim());
}

function formatCountdown(ms: number): string {
  if (ms <= 0) return "0:00:00";
  const totalSec = Math.floor(ms / 1000);
  const h = Math.floor(totalSec / 3600);
  const m = Math.floor((totalSec % 3600) / 60);
  const s = totalSec % 60;
  return `${h}:${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
}

export default function App() {
  const { showToast } = useToast();
  const [tab, setTab] = useState<TabId>("download");
  const [locales, setLocales] = useState<LocaleMap>({});
  const [settings, setSettings] = useState<AppSettingsDTO>(defaultSettings);
  const [bootstrapping, setBootstrapping] = useState(true);
  const [bootstrapError, setBootstrapError] = useState<string | null>(null);
  const [queue, setQueue] = useState<QueueTaskDTO[]>([]);
  const [running, setRunning] = useState(false);
  const [version, setVersion] = useState("dev");
  const [toolsDir, setToolsDir] = useState("");
  const [command, setCommand] = useState("");
  const [logs, setLogs] = useState("");
  const [progress, setProgress] = useState<DownloadProgressDTO | null>(null);
  const [history, setHistory] = useState<HistoryItemDTO[]>([]);
  const [deps, setDeps] = useState<DependencyDTO[]>([]);
  const [healthIssues, setHealthIssues] = useState<HealthIssueDTO[]>([]);
  const [presets, setPresets] = useState<string[]>([]);
  const [qualityPresets, setQualityPresets] = useState<{ key: string; value: string }[]>([]);
  const [mergeFormats, setMergeFormats] = useState<string[]>([]);
  const [profiles, setProfiles] = useState<string[]>([]);
  const [journal, setJournal] = useState<string[]>([]);
  const [selectedProfile, setSelectedProfile] = useState("");
  const [selectedQueueId, setSelectedQueueId] = useState<string | null>(null);
  const [showYtDlpModal, setShowYtDlpModal] = useState(false);
  const [startupYtDlpInstalling, setStartupYtDlpInstalling] = useState(false);
  const [startupYtDlpProgress, setStartupYtDlpProgress] = useState(0);
  const [showJournalModal, setShowJournalModal] = useState(false);
  const [dragOver, setDragOver] = useState(false);
  const dragDepthRef = useRef(0);
  const dragQueueIdRef = useRef<string | null>(null);
  const [dragQueueOverId, setDragQueueOverId] = useState<string | null>(null);
  const [probingFormats, setProbingFormats] = useState(false);
  const [formatResult, setFormatResult] = useState<ProbeResultDTO | null>(null);
  const [showHealthModal, setShowHealthModal] = useState(false);
  const [confirmClearHistory, setConfirmClearHistory] = useState(false);
  const [confirmDeleteProfile, setConfirmDeleteProfile] = useState<string | null>(null);
  const [promptSaveProfile, setPromptSaveProfile] = useState<"save" | "saveAs" | "rename" | null>(null);
  const [editingQueueTask, setEditingQueueTask] = useState<QueueTaskDTO | null>(null);
  const [clipboardSuggestion, setClipboardSuggestion] = useState<string | null>(null);
  const clipboardSeenRef = useRef<string>("");
  const [appUpdate, setAppUpdate] = useState<AppUpdateDTO | null>(null);
  const [otherInstances, setOtherInstances] = useState<InstanceDTO[]>([]);
  const [showInstancesModal, setShowInstancesModal] = useState(false);
  const [scheduledQueueAt, setScheduledQueueAt] = useState(0);
  const [scheduleInput, setScheduleInput] = useState("");
  const [now, setNow] = useState(() => Date.now());
  const [taskProgress, setTaskProgress] = useState<Record<string, DownloadProgressDTO>>({});
  const [compactTopBar, setCompactTopBar] = useState(false);
  const [screenSize, setScreenSize] = useState({ w: 1280, h: 800 });
  const [qualityPresetKey, setQualityPresetKey] = useState(0);
  const [historyQuery, setHistoryQuery] = useState("");
  const [historyStatusFilter, setHistoryStatusFilter] = useState<"all" | "completed" | "error" | "cancelled">("all");

  const settingsRef = useRef(settings);
  settingsRef.current = settings;
  const runningRef = useRef(running);
  runningRef.current = running;

  const t = useCallback((id: string) => locales[id] ?? id, [locales]);
  const cfg = settings.config;

  const uiScaleStyle = useMemo(() => {
    const scale = effectiveUIScale(settings.uiScale, screenSize.w, screenSize.h);
    const style: React.CSSProperties = { ["--ui-scale" as string]: String(scale) } as React.CSSProperties;
    if (scale !== 1) {
      style.transform = `scale(${scale})`;
      style.transformOrigin = "top left";
      style.width = `${100 / scale}%`;
      style.height = `${100 / scale}%`;
    }
    return style;
  }, [settings.uiScale, screenSize.w, screenSize.h]);

  const updateConfig = useCallback(
    (patch: Partial<ConfigDTO>) => {
      setSettings((prev) => ({ ...prev, config: { ...prev.config, ...patch } }));
    },
    [],
  );

  const updateSettings = useCallback((patch: Partial<AppSettingsDTO>) => {
    setSettings((prev) => ({ ...prev, ...patch }));
  }, []);

  const persistSettings = useCallback(async (s: AppSettingsDTO) => {
    await AppAPI.SaveSettings(s);
  }, []);

  const refreshPreview = useCallback(async (c: ConfigDTO) => {
    try {
      setCommand(await AppAPI.PreviewCommand(c));
    } catch {
      setCommand("");
    }
  }, []);

  const refreshHistory = useCallback(async () => {
    setHistory(asArray(await AppAPI.GetHistory()));
  }, []);

  const refreshHealth = useCallback(async () => {
    setHealthIssues(asArray(await AppAPI.HealthCheck()));
  }, []);

  const syncRunning = useCallback(async () => {
    const state = await AppAPI.GetState();
    setRunning(state.running);
    return state.running;
  }, []);

  const loadState = useCallback(async () => {
    await waitForWailsRuntime();
    const state = await AppAPI.GetState();
    const normalized = normalizeSettings(state.settings);
    setSettings(normalized);
    setQueue(asArray(state.queue));
    setJournal(asArray(state.journal));
    setRunning(state.running);
    setVersion(state.version);
    setToolsDir(state.toolsDir);
    setScheduledQueueAt(state.scheduledQueueAt || 0);
    setSelectedProfile(normalized.lastProfile || "");
    const loc = await AppAPI.GetLocales(normalized.language || "en");
    setLocales(loc);
    await refreshPreview(normalized.config);
    const [h, p, pr, mf] = await Promise.all([
      AppAPI.GetHistory(),
      AppAPI.ListProfiles(),
      AppAPI.GetPresets(),
      AppAPI.GetMergeFormats(),
    ]);
    setHistory(asArray(h));
    setProfiles(asArray(p));
    setPresets(asArray(pr));
    setMergeFormats(asArray(mf));
    const qp = await AppAPI.GetQualityPresets();
    setQualityPresets(asArray(qp));
    const bootDeps = await AppAPI.CheckDependencies();
    setDeps(asArray(bootDeps));
    await refreshHealth();
    setBootstrapError(null);
    setBootstrapping(false);
  }, [refreshPreview, refreshHealth]);

  useEffect(() => {
    loadState().catch((err) => {
      console.error(err);
      setBootstrapError(err instanceof Error ? err.message : String(err));
      setBootstrapping(false);
    });
  }, [loadState]);

  useEffect(() => {
    if (bootstrapping) return;
    const previewTimer = setTimeout(() => refreshPreview(settings.config), 300);
    return () => clearTimeout(previewTimer);
  }, [settings, refreshPreview, bootstrapping]);

  useEffect(() => {
    if (bootstrapping) return;
    void ScreenGetAll()
      .then((screens) => {
        const primary = screens.find((s) => s.isPrimary) ?? screens[0];
        if (primary) setScreenSize({ w: primary.width, h: primary.height });
      })
      .catch(() => setScreenSize({ w: window.innerWidth, h: window.innerHeight }));
  }, [bootstrapping]);

  useEffect(() => {
    if (bootstrapping) return;
    AppAPI.CheckAppUpdate()
      .then((info) => {
        if (info.updateAvail) setAppUpdate(info);
      })
      .catch(() => {
        /* offline or rate-limited — silently skip */
      });
  }, [bootstrapping]);

  useWindowBounds({
    bootstrapping,
    settings,
    onBoundsChange: updateSettings,
    onCompactChange: setCompactTopBar,
  });

  useEffect(() => {
    if (bootstrapping) return;
    const interval = setInterval(() => {
      persistSettings(settingsRef.current).catch(console.error);
    }, 30_000);
    return () => clearInterval(interval);
  }, [bootstrapping, persistSettings]);

  useEffect(() => {
    if (bootstrapping) return;
    clipboardSeenRef.current = settingsRef.current.config.url || "";
    const interval = setInterval(() => {
      if (!document.hasFocus()) return;
      navigator.clipboard
        .readText()
        .then((text) => {
          const trimmed = text.trim();
          if (!trimmed || trimmed === clipboardSeenRef.current) return;
          clipboardSeenRef.current = trimmed;
          if (looksLikeDownloadableURL(trimmed) && trimmed !== settingsRef.current.config.url) {
            setClipboardSuggestion(trimmed);
          }
        })
        .catch(() => {
          /* clipboard unavailable or permission denied */
        });
    }, 2500);
    return () => clearInterval(interval);
  }, [bootstrapping]);

  useEffect(() => {
    if (bootstrapping) return;
    const saveTimer = setTimeout(() => persistSettings(settings).catch(console.error), 600);
    return () => clearTimeout(saveTimer);
  }, [settings, persistSettings, bootstrapping]);

  useEffect(() => {
    const offs = [
      eventsOn("download:progress", (p) => {
        const prog = p as DownloadProgressDTO;
        if (prog.taskId) {
          setTaskProgress((prev) => {
            const next = { ...prev };
            if (prog.status === "completed" || prog.status === "error" || prog.status === "cancelled") {
              delete next[prog.taskId!];
            } else {
              next[prog.taskId!] = prog;
            }
            return next;
          });
        }
        const showInPanel = !prog.taskId || settingsRef.current.queueParallel <= 1;
        if (showInPanel) {
          setProgress(prog);
        }
        if (prog.status === "running") setRunning(true);
        if (prog.status === "completed" || prog.status === "error" || prog.status === "cancelled") {
          if (showInPanel) setRunning(false);
          refreshHistory().catch(console.error);
        }
      }),
      eventsOn("download:log", (l) => setLogs(l as string)),
      eventsOn("queue:update", (q) => setQueue(asArray(q as QueueTaskDTO[]))),
      eventsOn("journal:add", (entry) => setJournal((prev) => [...prev, entry as string])),
      eventsOn("startup:ytdlp-missing", () => setShowYtDlpModal(true)),
      eventsOn("queue:scheduled", (at) => setScheduledQueueAt((at as number) || 0)),
      eventsOn("startup:other-instances", (list) => {
        const arr = asArray(list as InstanceDTO[]);
        setOtherInstances(arr);
        if (arr.length > 0) setShowInstancesModal(true);
      }),
    ];
    return () => offs.forEach((off) => off());
  }, [refreshHistory]);

  useEffect(() => {
    if (!startupYtDlpInstalling) return;
    const off = eventsOn("install:progress", (d) => {
      const data = d as { id: string; pct: number };
      if (data.id === "ytdlp") setStartupYtDlpProgress(data.pct);
    });
    return () => off();
  }, [startupYtDlpInstalling]);

  const handleStartupYtDlpInstall = useCallback(async () => {
    setStartupYtDlpInstalling(true);
    setStartupYtDlpProgress(0);
    try {
      const { settings: next } = await installDependencyWithSave("ytdlp", settingsRef.current);
      setSettings(next);
      const d = await AppAPI.CheckDependencies();
      setDeps(asArray(d));
      await refreshHealth();
      await refreshPreview(next.config);
      setShowYtDlpModal(false);
    } catch (e) {
      showToast(e instanceof Error ? e.message : String(e), "error");
    } finally {
      setStartupYtDlpInstalling(false);
      setStartupYtDlpProgress(0);
    }
  }, [refreshHealth, refreshPreview, showToast]);

  useEffect(() => {
    if (!running) return;
    const timer = setInterval(() => {
      syncRunning().then((still) => {
        if (!still) refreshHistory().catch(console.error);
      });
    }, 800);
    return () => clearInterval(timer);
  }, [running, syncRunning, refreshHistory]);

  const startDownload = useCallback(async () => {
    const s = settingsRef.current;
    if (!s) return;
    setLogs("");
    setProgress({
      filePct: 0,
      overallPct: 0,
      speed: "",
      eta: "",
      status: "running",
      phase: "idle",
      plCurrent: 0,
      plTotal: 0,
      queueIdx: 0,
      queueTotal: 0,
    });
    setRunning(true);
    try {
      await AppAPI.RunDownload(s.config);
    } catch {
      setRunning(false);
    }
  }, []);

  const handleDepError = useCallback(
    (msg: string) => showToast(msg, "error"),
    [showToast],
  );

  const pasteURL = useCallback(async () => {
    try {
      const text = await navigator.clipboard.readText();
      if (text.trim()) updateConfig({ url: text.trim() });
    } catch {
      /* clipboard unavailable */
    }
  }, [updateConfig]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.ctrlKey && e.key === "v" && tab === "download") {
        e.preventDefault();
        void pasteURL();
      }
      if (e.ctrlKey && e.key === "Enter") {
        e.preventDefault();
        if (runningRef.current) void AppAPI.StopDownload();
        else void startDownload();
      }
      if (e.key === "Escape" && runningRef.current) {
        e.preventDefault();
        void AppAPI.StopDownload();
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [startDownload, tab, pasteURL]);

  useEffect(() => {
    if (!scheduledQueueAt) return;
    const interval = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(interval);
  }, [scheduledQueueAt]);

  const handleScheduleQueue = useCallback(async () => {
    if (!scheduleInput) return;
    const at = new Date(scheduleInput).getTime();
    if (Number.isNaN(at)) return;
    await AppAPI.ScheduleQueueRun(at);
    setScheduledQueueAt(at);
  }, [scheduleInput]);

  const handleCancelSchedule = useCallback(async () => {
    await AppAPI.CancelScheduledQueueRun();
    setScheduledQueueAt(0);
  }, []);

  const refreshOtherInstances = useCallback(async () => {
    const list = await AppAPI.ListOtherInstances();
    setOtherInstances(asArray(list));
  }, []);

  const handleCloseIdleInstances = useCallback(async () => {
    const n = await AppAPI.CloseIdleInstances();
    showToast(tf(locales, "instances.closed_count", { Count: String(n) }));
    setTimeout(() => refreshOtherInstances().catch(console.error), 1000);
  }, [showToast, locales, refreshOtherInstances]);

  const handleKillInstance = useCallback(
    async (pid: number) => {
      await AppAPI.KillInstance(pid);
      setTimeout(() => refreshOtherInstances().catch(console.error), 500);
    },
    [refreshOtherInstances],
  );

  const handleRunQueue = async () => {
    setTaskProgress({});
    setRunning(true);
    try {
      await AppAPI.RunQueue();
    } catch {
      setRunning(false);
    }
  };

  const handleQueueDragStart = (id: string) => {
    dragQueueIdRef.current = id;
  };

  const handleQueueDragOver = (e: React.DragEvent, id: string) => {
    e.preventDefault();
    if (dragQueueIdRef.current && dragQueueIdRef.current !== id) setDragQueueOverId(id);
  };

  const handleQueueDrop = (e: React.DragEvent, targetId: string) => {
    e.preventDefault();
    const draggedId = dragQueueIdRef.current;
    dragQueueIdRef.current = null;
    setDragQueueOverId(null);
    if (!draggedId || draggedId === targetId) return;
    const ids = queue.map((t) => t.id);
    const from = ids.indexOf(draggedId);
    const to = ids.indexOf(targetId);
    if (from < 0 || to < 0) return;
    ids.splice(from, 1);
    ids.splice(to, 0, draggedId);
    const byId = new Map(queue.map((t) => [t.id, t]));
    setQueue(ids.map((id) => byId.get(id)!));
    void AppAPI.ReorderQueue(ids);
  };

  const handleQueueDragEnd = () => {
    dragQueueIdRef.current = null;
    setDragQueueOverId(null);
  };

  useEffect(() => {
    if (tab === "history") refreshHistory().catch(console.error);
  }, [tab, refreshHistory]);

  const fetchFormats = async () => {
    if (!cfg.url.trim()) return;
    setProbingFormats(true);
    try {
      const result = await AppAPI.ProbeFormats(cfg);
      setFormatResult(result);
    } catch (e) {
      console.error(e);
    } finally {
      setProbingFormats(false);
    }
  };

  const loadProfile = async (name: string) => {
    if (!name) return;
    const p = await AppAPI.LoadProfile(name);
    updateConfig(p.config);
    updateSettings({ lastProfile: name });
  };

  const saveProfile = async (name: string, description = "") => {
    await AppAPI.SaveProfile({ name, description, config: cfg });
    const list = await AppAPI.ListProfiles();
    setProfiles(list);
    setSelectedProfile(name);
    updateSettings({ lastProfile: name });
    showToast(t("msg.profile_saved"));
  };

  const runSessionAction = useCallback(
    async (action: "save" | "load" | "resume") => {
      const path =
        settingsRef.current.sessionPath ||
        (await (action === "save" ? AppAPI.PickSaveFile("session.json") : AppAPI.PickFile()));
      if (!path) return;
      updateSettings({ sessionPath: path });
      if (action === "save") {
        await AppAPI.SaveSession(path);
      } else if (action === "load") {
        updateConfig(await AppAPI.LoadSession(path));
      } else {
        updateConfig(await AppAPI.ResumeSession(path));
      }
    },
    [updateSettings, updateConfig],
  );

  const resetWindowSize = useCallback(async () => {
    WindowSetSize(BASE_WINDOW_WIDTH, BASE_WINDOW_HEIGHT);
    const next = { ...settingsRef.current, windowWidth: 0, windowHeight: 0 };
    setSettings(next);
    await AppAPI.SaveSettings(next);
    showToast(t("btn.reset_window_size"));
  }, [showToast, t]);

  const handleExportSettings = useCallback(async () => {
    const path = await AppAPI.PickSaveFile("vadlp-backup.json");
    if (!path) return;
    await AppAPI.ExportSettings(path);
    showToast(t("msg.export_done"));
  }, [showToast, t]);

  const handleImportSettings = useCallback(async () => {
    const path = await AppAPI.PickFile();
    if (!path) return;
    await AppAPI.ImportSettings(path);
    await loadState();
    showToast(t("msg.import_done"));
  }, [showToast, t, loadState]);

  const handleSplitOffset = useCallback(
    (offset: number) => {
      updateSettings({ activityPanelOffset: offset });
    },
    [updateSettings],
  );

  const depAttention = useMemo(() => countDepAttention(deps), [deps]);

  const queueOverall = useMemo(
    () => queueOverallProgress(queue, taskProgress),
    [queue, taskProgress],
  );

  const healthAttention = healthIssues.filter(
    (i) => i.severity === "warning" || i.severity === "critical",
  ).length;

  const isRunning = running || progress?.status === "running";
  const statusKey = isRunning ? "running" : progress?.status === "error" ? "error" : progress?.status === "cancelled" ? "cancelled" : progress?.status === "completed" ? "completed" : "ready";

  const presetLabels: Record<string, string> = useMemo(
    () => ({
      youtube_playlist: t("preset.youtube_playlist"),
      audio_only: t("preset.audio_only"),
      video_best: t("preset.video_best"),
      video_1080: t("preset.video_1080"),
      video_4k: t("preset.video_4k"),
      podcast: t("preset.podcast"),
    }),
    [t],
  );

  const containerOptions = useMemo(() => {
    const opts = [...mergeFormats];
    if (cfg.format && !opts.includes(cfg.format)) opts.push(cfg.format);
    return opts;
  }, [mergeFormats, cfg.format]);

  const filteredHistory = useMemo(() => {
    const q = historyQuery.trim().toLowerCase();
    return history.filter((item) => {
      if (historyStatusFilter !== "all" && item.status !== historyStatusFilter) return false;
      if (q && !item.url.toLowerCase().includes(q)) return false;
      return true;
    });
  }, [history, historyQuery, historyStatusFilter]);

  const onDrop = (e: React.DragEvent) => {
    e.preventDefault();
    dragDepthRef.current = 0;
    setDragOver(false);
    const url = extractDroppedURL(e.dataTransfer);
    if (url) updateConfig({ url });
  };

  const onDragEnter = (e: React.DragEvent) => {
    e.preventDefault();
    dragDepthRef.current += 1;
    setDragOver(true);
  };

  const onDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    dragDepthRef.current = Math.max(0, dragDepthRef.current - 1);
    if (dragDepthRef.current === 0) setDragOver(false);
  };

  const activityOpen = settings.activityPanelOpen;

  useEffect(() => {
    if (settings.theme !== "auto") {
      document.documentElement.setAttribute("data-theme", settings.theme === "light" ? "light" : "dark");
      return;
    }
    const media = window.matchMedia("(prefers-color-scheme: light)");
    const apply = () => document.documentElement.setAttribute("data-theme", media.matches ? "light" : "dark");
    apply();
    media.addEventListener("change", apply);
    return () => media.removeEventListener("change", apply);
  }, [settings.theme]);

  return (
    <div
      className={`app-shell${dragOver ? " drag-over" : ""}`}
      style={uiScaleStyle}
      onDragEnter={onDragEnter}
      onDragOver={(e) => e.preventDefault()}
      onDragLeave={onDragLeave}
      onDrop={onDrop}
    >
      {bootstrapping && (
        <div className="bootstrap-overlay">
          <span className="status-pill running">Loading…</span>
        </div>
      )}
      {bootstrapError && (
        <div className="bootstrap-banner">{bootstrapError}</div>
      )}
      {appUpdate && (
        <div className="clipboard-banner app-update-banner">
          <span className="clipboard-banner-text">
            {tf(locales, "update.available", { Current: appUpdate.current, Latest: appUpdate.latest })}
          </span>
          <button type="button" className="btn btn-sm btn-primary" onClick={() => BrowserOpenURL(appUpdate.url)}>
            {t("update.open_release")}
          </button>
          <button type="button" className="btn btn-sm" onClick={() => setAppUpdate(null)}>
            {t("btn.dismiss")}
          </button>
        </div>
      )}
      {clipboardSuggestion && (
        <div className="clipboard-banner">
          <span className="clipboard-banner-text" title={clipboardSuggestion}>
            {tf(locales, "clipboard.detected", { Url: clipboardSuggestion })}
          </span>
          <button
            type="button"
            className="btn btn-sm btn-primary"
            onClick={async () => {
              await AppAPI.AddToQueue({ ...cfg, url: clipboardSuggestion });
              setClipboardSuggestion(null);
            }}
          >
            {t("btn.add_queue")}
          </button>
          <button type="button" className="btn btn-sm" onClick={() => setClipboardSuggestion(null)}>
            {t("btn.dismiss")}
          </button>
        </div>
      )}
      <header className={`top-bar${compactTopBar ? " top-bar-compact" : ""}`}>
        <div className="app-title">
          VAdlp {!compactTopBar && <span>· yt-dlp GUI</span>}
        </div>
        <span className={`status-pill ${statusKey}`}>{t(`status.${statusKey}`)}</span>
        {progress?.phase && progress.phase !== "idle" && isRunning && (
          <span className="status-pill running">{t(`phase.${progress.phase}`)}</span>
        )}
        <button
          type="button"
          className={`btn btn-sm health-btn${healthAttention ? " health-warn" : ""}`}
          onClick={() => setShowHealthModal(true)}
        >
          {healthAttention ? `${t("health.btn.issues")} (${healthAttention})` : t("health.btn.all_ok")}
        </button>
        <div className="top-bar-spacer" />
        <button
          type="button"
          className="btn btn-sm"
          onClick={() => setShowJournalModal(true)}
        >
          {t("journal.title")} ({journal.length})
        </button>
        <button
          type="button"
          className={`btn btn-sm${activityOpen ? " btn-primary" : ""}`}
          onClick={() => updateSettings({ activityPanelOpen: !activityOpen })}
          title={t("activity.title")}
        >
          {t("activity.title")}
        </button>
        <button type="button" className="btn btn-sm" onClick={() => AppAPI.OpenFolder(cfg.outputPath)}>
          {t("btn.open_folder")}
        </button>
        {isRunning ? (
          <button type="button" className="btn btn-danger" onClick={() => AppAPI.StopDownload()}>
            {t("btn.stop")}
          </button>
        ) : (
          <button type="button" className="btn btn-primary" onClick={() => startDownload()}>
            {t("btn.download")}
          </button>
        )}
      </header>

      <ResizableSplit
        offset={settings.activityPanelOffset > 0.05 ? settings.activityPanelOffset : 0.4}
        onOffsetChange={handleSplitOffset}
        hidden={!activityOpen}
        left={
          <>
          <div className="profile-bar">
            <span className="profile-bar-label">{t("form.saved_profile")}</span>
            <select
              value={selectedProfile}
              onChange={(e) => {
                const name = e.target.value;
                setSelectedProfile(name);
                if (name) loadProfile(name).catch(console.error);
              }}
            >
              <option value="">—</option>
              {profiles.map((p) => (
                <option key={p} value={p}>
                  {p}
                </option>
              ))}
            </select>
            <button type="button" className="btn btn-sm" onClick={() => setPromptSaveProfile("save")}>
              {t("btn.save_profile")}
            </button>
            <button type="button" className="btn btn-sm" onClick={() => setPromptSaveProfile("saveAs")}>
              {t("btn.save_profile_as")}
            </button>
            <button
              type="button"
              className="btn btn-sm"
              disabled={!selectedProfile}
              onClick={() => setPromptSaveProfile("rename")}
            >
              {t("btn.rename_profile")}
            </button>
            <button
              type="button"
              className="btn btn-sm btn-danger"
              disabled={!selectedProfile}
              onClick={() => setConfirmDeleteProfile(selectedProfile)}
            >
              {t("btn.delete_profile")}
            </button>
            <button
              type="button"
              className="btn btn-sm"
              onClick={() => {
                setSelectedProfile("");
                updateConfig(defaultConfig());
              }}
            >
              {t("btn.new_profile")}
            </button>
          </div>
          <div className="tab-bar">
            {TABS.map((id) => (
              <button
                key={id}
                type="button"
                className={`tab-btn${tab === id ? " active" : ""}${id === "dependencies" && depAttention ? " tab-attention" : ""}`}
                onClick={() => setTab(id)}
              >
                {t(TAB_KEYS[id])}
                {id === "dependencies" && depAttention > 0 && (
                  <span className="tab-badge">{depAttention}</span>
                )}
              </button>
            ))}
          </div>

          <div className="tab-content">
            {tab === "download" && (
              <div className="form-grid">
                <Field label={t("card.url")} wide>
                  <div className="input-with-actions">
                    <input
                      type="text"
                      placeholder={t("placeholder.url")}
                      value={cfg.url}
                      onChange={(e) => updateConfig({ url: e.target.value })}
                    />
                    <button type="button" className="btn btn-sm" onClick={pasteURL}>
                      {t("btn.paste")}
                    </button>
                    <button
                      type="button"
                      className="btn btn-sm"
                      disabled={!cfg.url.trim() || probingFormats}
                      onClick={fetchFormats}
                    >
                      {probingFormats ? t("btn.working") : t("btn.fetch_formats")}
                    </button>
                  </div>
                </Field>
                <Field label={t("card.batch")} wide>
                  <textarea
                    placeholder={t("placeholder.batch_urls")}
                    value={cfg.batchUrls}
                    onChange={(e) => updateConfig({ batchUrls: e.target.value })}
                  />
                </Field>
                <Field label={t("card.output")}>
                  <div className="input-with-actions">
                    <input
                      type="text"
                      value={cfg.outputPath}
                      onChange={(e) => updateConfig({ outputPath: e.target.value })}
                    />
                    <button
                      type="button"
                      className="btn btn-sm"
                      onClick={async () => {
                        const p = await AppAPI.PickFolder();
                        if (p) updateConfig({ outputPath: p });
                      }}
                    >
                      {t("btn.browse")}
                    </button>
                  </div>
                </Field>
                <Field label={t("form.filename_template")}>
                  <input
                    type="text"
                    value={cfg.outputTemplate}
                    onChange={(e) => updateConfig({ outputTemplate: e.target.value })}
                  />
                </Field>

                <p className="hint">{t("format.quality_hint")}</p>

                <Field label={t("form.quality_preset")}>
                  <select
                    key={qualityPresetKey}
                    defaultValue=""
                    onChange={(e) => {
                      const val = e.target.value;
                      if (val) updateConfig({ quality: val });
                      setQualityPresetKey((k) => k + 1);
                    }}
                  >
                    <option value="">—</option>
                    {qualityPresets.map((q) => (
                      <option key={q.value} value={q.value}>
                        {t(q.key)}
                      </option>
                    ))}
                  </select>
                </Field>
                <Field label={t("form.quality")}>
                  <input
                    type="text"
                    placeholder={t("placeholder.quality")}
                    value={cfg.quality}
                    onChange={(e) => updateConfig({ quality: e.target.value })}
                  />
                </Field>
                <Field label={t("form.container")}>
                  <select
                    value={containerOptions.includes(cfg.format) ? cfg.format : ""}
                    onChange={(e) => {
                      const val = e.target.value;
                      if (val) updateConfig({ format: val });
                    }}
                  >
                    <option value="">—</option>
                    {containerOptions.map((f) => (
                      <option key={f} value={f}>
                        {f}
                      </option>
                    ))}
                  </select>
                </Field>
                <Field label={t("form.container_custom")}>
                  <input
                    type="text"
                    placeholder={t("form.container_custom")}
                    value={containerOptions.includes(cfg.format) ? "" : cfg.format}
                    onChange={(e) => {
                      const val = e.target.value.trim();
                      updateConfig({ format: val });
                    }}
                  />
                </Field>
                <Check
                  label={t("form.audio_only")}
                  checked={cfg.audioOnly}
                  onChange={(v) => updateConfig({ audioOnly: v })}
                />
                {cfg.audioOnly && (
                  <Field label={t("form.audio_format")}>
                    <select
                      value={cfg.audioFormat}
                      onChange={(e) => updateConfig({ audioFormat: e.target.value })}
                    >
                      {AUDIO_FORMATS.map((f) => (
                        <option key={f || "default"} value={f}>
                          {f || "—"}
                        </option>
                      ))}
                    </select>
                  </Field>
                )}

                <div>
                  <div className="hint hint-spaced">
                    {t("form.quick_presets")}
                  </div>
                  <div className="preset-row">
                    {presets.map((p) => (
                      <button
                        key={p}
                        type="button"
                        className="preset-chip"
                        onClick={async () => {
                          const next = await AppAPI.ApplyPreset(p);
                          updateConfig(next);
                        }}
                      >
                        {presetLabels[p] ?? p}
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            )}

            {tab === "network" && (
              <div className="form-grid">
                <Check label={t("check.cookies_browser")} checked={cfg.useCookiesBrowser} onChange={(v) => updateConfig({ useCookiesBrowser: v })} />
                <Field label={t("form.browser")}>
                  <select
                    value={cfg.cookiesBrowser || "chrome"}
                    disabled={!cfg.useCookiesBrowser}
                    onChange={(e) => updateConfig({ cookiesBrowser: e.target.value })}
                  >
                    {COOKIES_BROWSERS.map((b) => (
                      <option key={b} value={b}>
                        {b}
                      </option>
                    ))}
                    {!COOKIES_BROWSERS.includes(cfg.cookiesBrowser as (typeof COOKIES_BROWSERS)[number]) &&
                      cfg.cookiesBrowser && (
                        <option value={cfg.cookiesBrowser}>{cfg.cookiesBrowser}</option>
                      )}
                  </select>
                </Field>
                <Check label={t("check.cookies_file")} checked={cfg.useCookiesFile} onChange={(v) => updateConfig({ useCookiesFile: v })} />
                <Field label={t("form.cookies_file")}>
                  <div className="input-with-actions">
                    <input type="text" placeholder={t("placeholder.cookies")} value={cfg.cookiesFile} onChange={(e) => updateConfig({ cookiesFile: e.target.value })} />
                    <button
                      type="button"
                      className="btn btn-sm"
                      onClick={async () => {
                        const p = await AppAPI.PickFile();
                        if (p) updateConfig({ cookiesFile: p, useCookiesFile: true });
                      }}
                    >
                      {t("btn.browse")}
                    </button>
                  </div>
                </Field>
                <Field label={t("form.proxy")}>
                  <input type="text" placeholder={t("placeholder.proxy")} value={cfg.proxy} onChange={(e) => updateConfig({ proxy: e.target.value })} />
                </Field>
                <Field label={t("form.rate_limit")}>
                  <input type="text" placeholder={t("placeholder.rate")} value={cfg.rateLimit} onChange={(e) => updateConfig({ rateLimit: e.target.value })} />
                </Field>
                <Field label={t("form.username")}>
                  <input type="text" value={cfg.username} onChange={(e) => updateConfig({ username: e.target.value })} />
                </Field>
                <Field label={t("form.password")}>
                  <input type="password" value={cfg.password} onChange={(e) => updateConfig({ password: e.target.value })} />
                </Field>
              </div>
            )}

            {tab === "playlist" && (
              <div className="form-grid">
                <Check label={t("check.reverse")} checked={cfg.playlistReverse} onChange={(v) => updateConfig({ playlistReverse: v })} />
                <Check label={t("check.continue")} checked={cfg.continue} onChange={(v) => updateConfig({ continue: v })} />
                <Check label={t("check.no_part")} checked={cfg.noPart} onChange={(v) => updateConfig({ noPart: v })} />
                <Check label={t("check.no_playlist")} checked={cfg.noPlaylist} onChange={(v) => updateConfig({ noPlaylist: v })} />
                <Check label={t("check.flat_playlist")} checked={cfg.flatPlaylist} onChange={(v) => updateConfig({ flatPlaylist: v })} />
                <Field label={t("form.playlist_start")}>
                  <input type="number" min={0} value={cfg.playlistStart || ""} onChange={(e) => updateConfig({ playlistStart: parseInt(e.target.value, 10) || 0 })} />
                </Field>
                <Field label={t("form.playlist_end")}>
                  <input type="number" min={0} value={cfg.playlistEnd || ""} onChange={(e) => updateConfig({ playlistEnd: parseInt(e.target.value, 10) || 0 })} />
                </Field>
                <Field label={t("form.max_downloads")}>
                  <input
                    type="number"
                    min={0}
                    placeholder={t("placeholder.max_downloads")}
                    value={cfg.maxDownloads || ""}
                    onChange={(e) => updateConfig({ maxDownloads: parseInt(e.target.value, 10) || 0 })}
                  />
                </Field>
                <Field label={t("form.archive")}>
                  <input type="text" placeholder={t("placeholder.archive")} value={cfg.downloadArchive} onChange={(e) => updateConfig({ downloadArchive: e.target.value })} />
                </Field>
                <Field label={t("form.session_path")}>
                  <div className="input-with-actions">
                    <input type="text" value={settings.sessionPath} onChange={(e) => updateSettings({ sessionPath: e.target.value })} />
                    <button
                      type="button"
                      className="btn btn-sm"
                      onClick={async () => {
                        const p = await AppAPI.PickFile();
                        if (p) updateSettings({ sessionPath: p });
                      }}
                    >
                      {t("btn.browse")}
                    </button>
                  </div>
                </Field>
                <div className="btn-row">
                  <button type="button" className="btn btn-sm" onClick={() => runSessionAction("save")}>
                    {t("btn.save_session")}
                  </button>
                  <button type="button" className="btn btn-sm" onClick={() => runSessionAction("load")}>
                    {t("btn.load_session")}
                  </button>
                  <button type="button" className="btn btn-sm" onClick={() => runSessionAction("resume")}>
                    {t("btn.resume_session")}
                  </button>
                </div>
              </div>
            )}

            {tab === "extras" && (
              <div className="form-grid">
                <div className="hint">{t("card.media_extras")}</div>
                <Check label={t("check.write_subs")} checked={cfg.writeSubs} onChange={(v) => updateConfig({ writeSubs: v })} />
                <Check label={t("check.write_auto_sub")} checked={cfg.writeAutoSub} onChange={(v) => updateConfig({ writeAutoSub: v })} />
                <Check label={t("check.embed_subs")} checked={cfg.embedSubs} onChange={(v) => updateConfig({ embedSubs: v })} />
                <Field label={t("form.sub_langs")}>
                  <input type="text" value={cfg.subLangs} onChange={(e) => updateConfig({ subLangs: e.target.value })} />
                </Field>
                <Check label={t("check.write_thumb")} checked={cfg.writeThumbnail} onChange={(v) => updateConfig({ writeThumbnail: v })} />
                <Check label={t("check.embed_thumb")} checked={cfg.embedThumbnail} onChange={(v) => updateConfig({ embedThumbnail: v })} />
                <Check label={t("check.embed_meta")} checked={cfg.embedMetadata} onChange={(v) => updateConfig({ embedMetadata: v })} />
                <Check label={t("check.embed_chapters")} checked={cfg.embedChapters} onChange={(v) => updateConfig({ embedChapters: v })} />
                <Check label={t("check.write_info_json")} checked={cfg.writeInfoJSON} onChange={(v) => updateConfig({ writeInfoJSON: v })} />
                <Field label={t("form.load_info_json")}>
                  <input type="text" placeholder={t("placeholder.load_info_json")} value={cfg.loadInfoJson} onChange={(e) => updateConfig({ loadInfoJson: e.target.value })} />
                </Field>

                <div className="section-gap hint">{t("card.retries")}</div>
                <Field label={t("form.retries")}>
                  <input type="number" min={0} value={cfg.retries} onChange={(e) => updateConfig({ retries: parseInt(e.target.value, 10) || 0 })} />
                </Field>
                <Field label={t("form.frag_retries")}>
                  <input type="number" min={0} value={cfg.fragmentRetries} onChange={(e) => updateConfig({ fragmentRetries: parseInt(e.target.value, 10) || 0 })} />
                </Field>
                <Field label={t("form.concurrent_frags")}>
                  <input type="number" min={1} value={cfg.concurrentFragments} onChange={(e) => updateConfig({ concurrentFragments: parseInt(e.target.value, 10) || 1 })} />
                </Field>
                <Field label={t("form.socket_timeout")}>
                  <input type="number" min={0} placeholder={t("placeholder.socket_timeout")} value={cfg.socketTimeout || ""} onChange={(e) => updateConfig({ socketTimeout: parseInt(e.target.value, 10) || 0 })} />
                </Field>
                <Check label={t("check.no_warnings")} checked={cfg.noWarnings} onChange={(v) => updateConfig({ noWarnings: v })} />
                <Check label={t("check.verbose")} checked={cfg.verbose} onChange={(v) => updateConfig({ verbose: v, quiet: v ? false : cfg.quiet })} />
                <Check label={t("check.quiet")} checked={cfg.quiet} onChange={(v) => updateConfig({ quiet: v, verbose: v ? false : cfg.verbose })} />
                <Check label={t("check.windows_filenames")} checked={cfg.windowsFilenames} onChange={(v) => updateConfig({ windowsFilenames: v })} />
                <Check label={t("check.no_mtime")} checked={cfg.noMtime} onChange={(v) => updateConfig({ noMtime: v })} />
                <Check label={t("check.abort_on_error")} checked={cfg.abortOnError} onChange={(v) => updateConfig({ abortOnError: v })} />
                <Check label={t("check.ignore_errors")} checked={cfg.ignoreErrors} onChange={(v) => updateConfig({ ignoreErrors: v })} />
                <Check label={t("check.sponsorblock")} checked={cfg.sponsorBlockRemove} onChange={(v) => updateConfig({ sponsorBlockRemove: v })} />
                <Field label={t("card.extra_flags")} wide>
                  <textarea placeholder={t("placeholder.extra_args")} value={cfg.extraArgs} onChange={(e) => updateConfig({ extraArgs: e.target.value })} />
                </Field>
              </div>
            )}

            {tab === "queue" && (
              <div className="form-grid">
                <div className="btn-row">
                  <button type="button" className="btn btn-sm" onClick={() => AppAPI.AddToQueue(cfg)}>
                    {t("btn.add_queue")}
                  </button>
                  <button type="button" className="btn btn-primary btn-sm" disabled={isRunning} onClick={handleRunQueue}>
                    {t("btn.run_queue")}
                  </button>
                  <button type="button" className="btn btn-sm" disabled={!selectedQueueId} onClick={() => selectedQueueId && AppAPI.RemoveFromQueue(selectedQueueId)}>
                    {t("btn.remove")}
                  </button>
                  <button type="button" className="btn btn-sm" onClick={() => AppAPI.RetryFailedQueue()}>
                    {t("btn.retry_failed")}
                  </button>
                  <button type="button" className="btn btn-sm" onClick={() => AppAPI.ClearQueue()}>
                    {t("btn.clear_queue")}
                  </button>
                </div>
                <div className="btn-row">
                  {scheduledQueueAt ? (
                    <>
                      <span className="hint">
                        {tf(locales, "queue.scheduled_in", { Time: formatCountdown(scheduledQueueAt - now) })}
                      </span>
                      <button type="button" className="btn btn-sm btn-danger" onClick={() => handleCancelSchedule()}>
                        {t("btn.cancel_schedule")}
                      </button>
                    </>
                  ) : (
                    <>
                      <input
                        type="datetime-local"
                        value={scheduleInput}
                        onChange={(e) => setScheduleInput(e.target.value)}
                      />
                      <button
                        type="button"
                        className="btn btn-sm"
                        disabled={!scheduleInput}
                        onClick={() => handleScheduleQueue()}
                      >
                        {t("btn.schedule_queue")}
                      </button>
                    </>
                  )}
                </div>
                <div className="queue-list">
                  {queue.length === 0 && <div className="hint">{t("journal.empty")}</div>}
                  {queue.map((task) => {
                    const tp = taskProgress[task.id];
                    return (
                    <div
                      key={task.id}
                      className={`queue-item${selectedQueueId === task.id ? " selected" : ""}${task.status === "running" ? " queue-item-running" : ""}${dragQueueOverId === task.id ? " queue-item-drag-over" : ""}`}
                      onClick={() => setSelectedQueueId(task.id)}
                      draggable
                      onDragStart={() => handleQueueDragStart(task.id)}
                      onDragOver={(e) => handleQueueDragOver(e, task.id)}
                      onDrop={(e) => handleQueueDrop(e, task.id)}
                      onDragEnd={handleQueueDragEnd}
                    >
                      <span className="queue-drag-handle" title={t("queue.drag_hint")} aria-hidden>⠿</span>
                      <span className={`queue-status-dot ${task.status}`} aria-hidden />
                      <span className={`status-pill ${task.status}`}>{t(`status.${task.status}`)}</span>
                      <span className="queue-item-name" title={task.name}>{task.name}</span>
                      {task.status === "running" && tp && (
                        <>
                          <div className="queue-item-progress">
                            <div className="progress-fill" style={{ width: `${tp.filePct}%` }} />
                          </div>
                          {(tp.speed || tp.eta) && (
                            <span className="queue-item-speed">
                              {tf(locales, "progress.speed_eta", { Speed: tp.speed || "—", ETA: tp.eta || "—" })}
                            </span>
                          )}
                        </>
                      )}
                      {task.status === "running" && (
                        <button
                          type="button"
                          className="btn btn-sm btn-danger"
                          onClick={(e) => {
                            e.stopPropagation();
                            void AppAPI.CancelQueueTask(task.id);
                          }}
                        >
                          {t("btn.cancel_task")}
                        </button>
                      )}
                      {(task.status === "queued" || task.status === "paused") && (
                        <button
                          type="button"
                          className="btn btn-sm"
                          onClick={(e) => { e.stopPropagation(); setEditingQueueTask(task); }}
                        >
                          {t("btn.edit_task")}
                        </button>
                      )}
                      {task.status === "queued" && (
                        <button
                          type="button"
                          className="btn btn-sm"
                          onClick={(e) => { e.stopPropagation(); void AppAPI.PauseQueueTask(task.id); }}
                        >
                          {t("btn.pause_task")}
                        </button>
                      )}
                      {task.status === "paused" && (
                        <button
                          type="button"
                          className="btn btn-sm btn-primary"
                          onClick={(e) => { e.stopPropagation(); void AppAPI.ResumeQueueTask(task.id); }}
                        >
                          {t("btn.resume_task")}
                        </button>
                      )}
                      <button type="button" className="btn btn-sm" onClick={(e) => { e.stopPropagation(); AppAPI.MoveQueueItem(task.id, -1); }}>
                        {t("btn.up")}
                      </button>
                      <button type="button" className="btn btn-sm" onClick={(e) => { e.stopPropagation(); AppAPI.MoveQueueItem(task.id, 1); }}>
                        {t("btn.down")}
                      </button>
                    </div>
                    );
                  })}
                </div>
              </div>
            )}

            {tab === "history" && (
              <div className="form-grid">
                <div className="btn-row">
                  <input
                    type="text"
                    className="history-search"
                    placeholder={t("history.search_placeholder")}
                    value={historyQuery}
                    onChange={(e) => setHistoryQuery(e.target.value)}
                  />
                  <select
                    value={historyStatusFilter}
                    onChange={(e) => setHistoryStatusFilter(e.target.value as typeof historyStatusFilter)}
                  >
                    <option value="all">{t("history.filter_all")}</option>
                    <option value="completed">{t("status.completed")}</option>
                    <option value="error">{t("status.error")}</option>
                    <option value="cancelled">{t("status.cancelled")}</option>
                  </select>
                  <button type="button" className="btn btn-danger btn-sm" onClick={() => setConfirmClearHistory(true)}>
                    {t("btn.clear_history")}
                  </button>
                </div>
                <div className="history-list">
                  {filteredHistory.length === 0 && <div className="hint">{t("journal.empty")}</div>}
                  {filteredHistory.map((item, i) => (
                    <div
                      key={`${item.at}-${i}`}
                      className={`history-item ${item.status}`}
                      role="button"
                      tabIndex={0}
                      onClick={() => item.url && updateConfig({ url: item.url })}
                      onKeyDown={(e) => e.key === "Enter" && item.url && updateConfig({ url: item.url })}
                    >
                      <div className="history-url">{item.url}</div>
                      <div className="history-meta">
                        {t(`status.${item.status}`)} · {item.durationSec}s · {new Date(item.at).toLocaleString()}
                        {item.error ? ` · ${item.error}` : ""}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {tab === "dependencies" && (
              <DependenciesTab
                active={tab === "dependencies"}
                deps={deps}
                setDeps={setDeps}
                settings={settings}
                updateSettings={updateSettings}
                locales={locales}
                toolsDir={toolsDir}
                onHealthRefresh={refreshHealth}
                t={t}
                onError={handleDepError}
              />
            )}

            {tab === "settings" && (
              <div className="form-grid">
                <div className="hint">{t("card.settings")}</div>
                <Field label={t("form.language")}>
                  <select
                    value={settings.language || "en"}
                    onChange={async (e) => {
                      const lang = e.target.value;
                      const next = { ...settingsRef.current, language: lang };
                      updateSettings({ language: lang });
                      setLocales(await AppAPI.GetLocales(lang));
                      await AppAPI.SaveSettings(next);
                    }}
                  >
                    <option value="en">{t("lang.en")}</option>
                    <option value="ru">{t("lang.ru")}</option>
                    <option value="es">{t("lang.es")}</option>
                    <option value="pt">{t("lang.pt")}</option>
                    <option value="ja">{t("lang.ja")}</option>
                    <option value="de">{t("lang.de")}</option>
                    <option value="fr">{t("lang.fr")}</option>
                    <option value="pl">{t("lang.pl")}</option>
                    <option value="ko">{t("lang.ko")}</option>
                    <option value="zh-Hant">{t("lang.zh-Hant")}</option>
                    <option value="zh-Hans">{t("lang.zh-Hans")}</option>
                  </select>
                </Field>
                <Field label={t("form.theme")}>
                  <select
                    value={settings.theme}
                    onChange={(e) => updateSettings({ theme: e.target.value as AppSettingsDTO["theme"] })}
                  >
                    <option value="auto">{t("theme.auto")}</option>
                    <option value="dark">{t("theme.dark")}</option>
                    <option value="light">{t("theme.light")}</option>
                  </select>
                </Field>
                <Field label={t("form.ui_scale")}>
                  <select
                    value={String(settings.uiScale || 0)}
                    onChange={(e) => updateSettings({ uiScale: parseFloat(e.target.value) || 0 })}
                  >
                    {UI_SCALES.map((s) => (
                      <option key={s.key} value={String(s.value)}>
                        {t(s.key)}
                      </option>
                    ))}
                  </select>
                </Field>
                <p className="hint">{t("tools.ui_scale_hint")}</p>
                <Field label={t("form.queue_workers")}>
                  <input
                    type="number"
                    min={1}
                    max={32}
                    value={settings.queueParallel}
                    onChange={(e) => updateSettings({ queueParallel: parseInt(e.target.value, 10) || 1 })}
                  />
                </Field>
                <Check
                  label={t("check.debug_log")}
                  checked={settings.debugLog}
                  onChange={(v) => updateSettings({ debugLog: v })}
                />
                <Check
                  label={t("activity.title")}
                  checked={settings.activityPanelOpen}
                  onChange={(v) => updateSettings({ activityPanelOpen: v })}
                />
                <div className="btn-row">
                  <button type="button" className="btn btn-sm" onClick={() => resetWindowSize()}>
                    {t("btn.reset_window_size")}
                  </button>
                </div>
                <p className="hint">{t("tray.hint")}</p>
                <div className="btn-row">
                  <button
                    type="button"
                    className="btn btn-sm"
                    onClick={async () => {
                      await refreshOtherInstances();
                      setShowInstancesModal(true);
                    }}
                  >
                    {t("instances.title")}
                  </button>
                </div>
                <div className="section-gap hint">{t("card.backup")}</div>
                <div className="btn-row">
                  <button type="button" className="btn btn-sm" onClick={() => handleExportSettings()}>
                    {t("btn.export_settings")}
                  </button>
                  <button type="button" className="btn btn-sm" onClick={() => handleImportSettings()}>
                    {t("btn.import_settings")}
                  </button>
                </div>
                <div className="hint">{tf(locales, "tools.app_version", { Version: version })}</div>
              </div>
            )}
          </div>
          </>
        }
        right={
          <ActivityPanel
            command={command}
            logs={logs}
            progress={progress}
            queueOverall={queueOverall}
            locales={locales}
            t={t}
          />
        }
      />

      <footer className={`bottom-bar${dragOver ? " drop-active" : ""}`}>
        <span>{dragOver ? t("drop.hint") : `${t("drop.hint")} · Ctrl+Enter`}</span>
        <span className="bottom-bar-spacer">{version}</span>
      </footer>

      {showJournalModal && (
        <Modal title={t("journal.title")} onClose={() => setShowJournalModal(false)} wide>
          <textarea
            className="log-area readonly"
            readOnly
            value={journal.length ? journal.join("\n") : t("journal.empty")}
            rows={16}
          />
          <div className="btn-row">
            <button type="button" className="btn" onClick={() => setShowJournalModal(false)}>
              {t("btn.close")}
            </button>
          </div>
        </Modal>
      )}

      {showYtDlpModal && (
        <Modal title={t("install.ytdlp.title")} onClose={() => !startupYtDlpInstalling && setShowYtDlpModal(false)}>
          <p>{t("install.ytdlp.body")}</p>
          {startupYtDlpInstalling && (
            <div className="progress-track dep-progress">
              <div className="progress-fill" style={{ width: `${startupYtDlpProgress}%` }} />
            </div>
          )}
          <div className="btn-row">
            <button
              type="button"
              className="btn btn-primary"
              disabled={startupYtDlpInstalling}
              onClick={() => handleStartupYtDlpInstall()}
            >
              {startupYtDlpInstalling ? t("btn.working") : t("btn.install")}
            </button>
            <button
              type="button"
              className="btn"
              disabled={startupYtDlpInstalling}
              onClick={() => setShowYtDlpModal(false)}
            >
              {t("btn.skip")}
            </button>
            <button
              type="button"
              className="btn btn-sm"
              disabled={startupYtDlpInstalling}
              onClick={() => {
                setShowYtDlpModal(false);
                setTab("dependencies");
              }}
            >
              {t("tab.dependencies")}
            </button>
          </div>
        </Modal>
      )}

      {showInstancesModal && (
        <Modal title={t("instances.title")} onClose={() => setShowInstancesModal(false)} wide>
          <p className="hint">{t("instances.hint")}</p>
          {otherInstances.length === 0 ? (
            <p>{t("instances.none")}</p>
          ) : (
            <ul className="health-list instances-list">
              {otherInstances.map((inst) => (
                <li key={inst.pid} className="instances-row">
                  <span className={`status-pill ${inst.busy ? "running" : "ready"}`}>
                    {inst.busy ? t("instances.busy") : t("instances.idle")}
                  </span>
                  <span>{tf(locales, "instances.pid", { Pid: String(inst.pid) })}</span>
                  <span className="hint">{new Date(inst.startedAt).toLocaleString()}</span>
                  <button type="button" className="btn btn-sm btn-danger" onClick={() => handleKillInstance(inst.pid)}>
                    {t("btn.kill_instance")}
                  </button>
                </li>
              ))}
            </ul>
          )}
          <div className="btn-row">
            <button
              type="button"
              className="btn btn-sm"
              disabled={!otherInstances.some((i) => !i.busy)}
              onClick={() => handleCloseIdleInstances()}
            >
              {t("instances.close_idle")}
            </button>
            <button type="button" className="btn btn-sm" onClick={() => refreshOtherInstances()}>
              {t("btn.recheck")}
            </button>
            <button type="button" className="btn" onClick={() => setShowInstancesModal(false)}>
              {t("btn.close")}
            </button>
          </div>
        </Modal>
      )}

      {formatResult && (
        <FormatPickerModal
          result={formatResult}
          t={t}
          onPick={(formatId) => updateConfig({ quality: formatId })}
          onClose={() => setFormatResult(null)}
        />
      )}

      {showHealthModal && (
        <Modal title={t("health.title")} onClose={() => setShowHealthModal(false)} wide>
          {healthIssues.length === 0 ? (
            <p>{t("health.all_ok")}</p>
          ) : (
            <ul className="health-list">
              {healthIssues.map((iss) => (
                <li key={iss.id} className={`health-item health-${iss.severity}`}>
                  {tf(locales, iss.key, iss.params as Record<string, string>)}
                </li>
              ))}
            </ul>
          )}
          <div className="btn-row">
            {healthAttention > 0 && (
              <button
                type="button"
                className="btn btn-sm"
                onClick={() => {
                  setShowHealthModal(false);
                  setTab("dependencies");
                }}
              >
                {t("health.btn.go_deps")}
              </button>
            )}
            <button type="button" className="btn" onClick={() => setShowHealthModal(false)}>
              {t("btn.close")}
            </button>
          </div>
        </Modal>
      )}

      {confirmClearHistory && (
        <ConfirmModal
          title={t("btn.clear_history")}
          message={t("msg.clear_history_confirm")}
          confirmLabel={t("btn.clear_history")}
          cancelLabel={t("btn.cancel")}
          danger
          onConfirm={() => {
            AppAPI.ClearHistory().then(() => setHistory([]));
          }}
          onClose={() => setConfirmClearHistory(false)}
        />
      )}

      {confirmDeleteProfile && (
        <ConfirmModal
          title={t("btn.delete_profile")}
          message={tf(locales, "msg.delete_profile_confirm", { Name: confirmDeleteProfile })}
          confirmLabel={t("btn.delete_profile")}
          cancelLabel={t("btn.cancel")}
          danger
          onConfirm={async () => {
            await AppAPI.DeleteProfile(confirmDeleteProfile);
            const list = await AppAPI.ListProfiles();
            setProfiles(list);
            setSelectedProfile("");
            updateSettings({ lastProfile: "" });
          }}
          onClose={() => setConfirmDeleteProfile(null)}
        />
      )}

      {promptSaveProfile === "save" && (
        <PromptModal
          title={t("dialog.save_profile")}
          label={t("form.profile_name")}
          defaultValue={selectedProfile}
          submitLabel={t("btn.save")}
          cancelLabel={t("btn.cancel")}
          onSubmit={(name) => saveProfile(name).catch(console.error)}
          onClose={() => setPromptSaveProfile(null)}
        />
      )}

      {promptSaveProfile === "saveAs" && (
        <PromptModal
          title={t("dialog.save_profile_as")}
          label={t("form.profile_name")}
          descriptionLabel={t("form.profile_description")}
          submitLabel={t("btn.save")}
          cancelLabel={t("btn.cancel")}
          onSubmit={(name, description) => saveProfile(name, description ?? "").catch(console.error)}
          onClose={() => setPromptSaveProfile(null)}
        />
      )}

      {promptSaveProfile === "rename" && selectedProfile && (
        <PromptModal
          title={t("btn.rename_profile")}
          label={t("form.profile_name")}
          defaultValue={selectedProfile}
          submitLabel={t("btn.save")}
          cancelLabel={t("btn.cancel")}
          onSubmit={async (newName) => {
            await AppAPI.RenameProfile(selectedProfile, newName);
            const list = await AppAPI.ListProfiles();
            setProfiles(list);
            setSelectedProfile(newName);
            updateSettings({ lastProfile: newName });
          }}
          onClose={() => setPromptSaveProfile(null)}
        />
      )}

      {editingQueueTask && (
        <EditQueueTaskModal
          task={editingQueueTask}
          t={t}
          onClose={() => setEditingQueueTask(null)}
          onSave={async (cfg) => {
            await AppAPI.UpdateQueueTask(editingQueueTask.id, cfg);
            setEditingQueueTask(null);
          }}
        />
      )}
    </div>
  );
}

function EditQueueTaskModal({
  task,
  t,
  onClose,
  onSave,
}: {
  task: QueueTaskDTO;
  t: (id: string) => string;
  onClose: () => void;
  onSave: (cfg: ConfigDTO) => Promise<void>;
}) {
  const [cfg, setCfg] = useState<ConfigDTO>(task.config);
  const [saving, setSaving] = useState(false);

  const patch = (p: Partial<ConfigDTO>) => setCfg((prev) => ({ ...prev, ...p }));

  return (
    <Modal title={t("dialog.edit_task")} onClose={onClose}>
      <Field label={t("form.rate_limit")}>
        <input
          type="text"
          placeholder={t("placeholder.rate")}
          value={cfg.rateLimit}
          onChange={(e) => patch({ rateLimit: e.target.value })}
        />
      </Field>
      <Field label={t("form.quality")}>
        <input
          type="text"
          placeholder={t("placeholder.quality")}
          value={cfg.quality}
          onChange={(e) => patch({ quality: e.target.value })}
        />
      </Field>
      <Check label={t("form.audio_only")} checked={cfg.audioOnly} onChange={(v) => patch({ audioOnly: v })} />
      <div className="btn-row">
        <button
          type="button"
          className="btn btn-primary"
          disabled={saving}
          onClick={async () => {
            setSaving(true);
            try {
              await onSave(cfg);
            } finally {
              setSaving(false);
            }
          }}
        >
          {saving ? t("btn.working") : t("btn.save")}
        </button>
        <button type="button" className="btn" disabled={saving} onClick={onClose}>
          {t("btn.cancel")}
        </button>
      </div>
    </Modal>
  );
}
