import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { ActivityPanel } from "./components/ActivityPanel";
import { FormatPickerModal } from "./components/FormatPickerModal";
import { ConfirmModal, Modal, PromptModal } from "./components/Modal";
import { ResizableSplit } from "./components/ResizableSplit";
import { countDepAttention, DependenciesTab, installDependencyWithSave } from "./components/DependenciesTab";
import { useToast } from "./components/Toast";
import { EditQueueTaskModal } from "./components/EditQueueTaskModal";
import { extractDroppedURL, looksLikeDownloadableURL } from "./components/FormControls";
import { DownloadTab } from "./components/tabs/DownloadTab";
import { NetworkTab } from "./components/tabs/NetworkTab";
import { PlaylistTab } from "./components/tabs/PlaylistTab";
import { ExtrasTab } from "./components/tabs/ExtrasTab";
import { QueueTab } from "./components/tabs/QueueTab";
import { HistoryTab } from "./components/tabs/HistoryTab";
import { SettingsTab } from "./components/tabs/SettingsTab";
import { asArray, defaultConfig, defaultSettings, normalizeSettings } from "./lib/defaults";
import { tf } from "./lib/i18nFmt";
import { queueOverallProgress } from "./lib/progressLabels";
import { BASE_WINDOW_HEIGHT, BASE_WINDOW_WIDTH, effectiveUIScale } from "./lib/uiScale";
import { useWindowBounds } from "./hooks/useWindowBounds";
import { AppAPI, eventsOn, waitForWailsRuntime } from "./wailsjs/runtime";
import { BrowserOpenURL, ScreenGetAll, WindowSetSize } from "./wailsjs/runtime/runtime";
import { app, downloader } from "./wailsjs/go/models";
import type { DownloadProgressDTO, LocaleMap } from "./lib/eventTypes";

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

export default function App() {
  const { showToast } = useToast();
  const [tab, setTab] = useState<TabId>("download");
  const [locales, setLocales] = useState<LocaleMap>({});
  const [settings, setSettings] = useState<app.AppSettingsDTO>(defaultSettings);
  const [bootstrapping, setBootstrapping] = useState(true);
  const [bootstrapError, setBootstrapError] = useState<string | null>(null);
  const [queue, setQueue] = useState<app.QueueTaskDTO[]>([]);
  const [running, setRunning] = useState(false);
  const [version, setVersion] = useState("dev");
  const [toolsDir, setToolsDir] = useState("");
  const [command, setCommand] = useState("");
  const [logs, setLogs] = useState("");
  const [progress, setProgress] = useState<DownloadProgressDTO | null>(null);
  const [history, setHistory] = useState<app.HistoryItemDTO[]>([]);
  const [deps, setDeps] = useState<app.DependencyDTO[]>([]);
  const [healthIssues, setHealthIssues] = useState<app.HealthIssueDTO[]>([]);
  const [presets, setPresets] = useState<string[]>([]);
  const [qualityPresets, setQualityPresets] = useState<{ key: string; value: string }[]>([]);
  const [mergeFormats, setMergeFormats] = useState<string[]>([]);
  const [profiles, setProfiles] = useState<string[]>([]);
  const [journal, setJournal] = useState<string[]>([]);
  const [selectedProfile, setSelectedProfile] = useState("");
  const [showYtDlpModal, setShowYtDlpModal] = useState(false);
  const [startupYtDlpInstalling, setStartupYtDlpInstalling] = useState(false);
  const [startupYtDlpProgress, setStartupYtDlpProgress] = useState(0);
  const [showJournalModal, setShowJournalModal] = useState(false);
  const [dragOver, setDragOver] = useState(false);
  const dragDepthRef = useRef(0);
  const [probingFormats, setProbingFormats] = useState(false);
  const [formatResult, setFormatResult] = useState<downloader.ProbeResult | null>(null);
  const [showHealthModal, setShowHealthModal] = useState(false);
  const [confirmClearHistory, setConfirmClearHistory] = useState(false);
  const [confirmDeleteProfile, setConfirmDeleteProfile] = useState<string | null>(null);
  const [promptSaveProfile, setPromptSaveProfile] = useState<"save" | "saveAs" | "rename" | null>(null);
  const [editingQueueTask, setEditingQueueTask] = useState<app.QueueTaskDTO | null>(null);
  const [clipboardSuggestion, setClipboardSuggestion] = useState<string | null>(null);
  const clipboardSeenRef = useRef<string>("");
  const [appUpdate, setAppUpdate] = useState<app.AppUpdateDTO | null>(null);
  const [otherInstances, setOtherInstances] = useState<app.InstanceDTO[]>([]);
  const [showInstancesModal, setShowInstancesModal] = useState(false);
  const [scheduledQueueAt, setScheduledQueueAt] = useState(0);
  const [scheduleInput, setScheduleInput] = useState("");
  const [now, setNow] = useState(() => Date.now());
  const [taskProgress, setTaskProgress] = useState<Record<string, DownloadProgressDTO>>({});
  const [compactTopBar, setCompactTopBar] = useState(false);
  const [screenSize, setScreenSize] = useState({ w: 1280, h: 800 });

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
    (patch: Partial<app.ConfigDTO>) => {
      setSettings((prev) => new app.AppSettingsDTO({ ...prev, config: { ...prev.config, ...patch } }));
    },
    [],
  );

  const updateSettings = useCallback((patch: Partial<app.AppSettingsDTO>) => {
    setSettings((prev) => new app.AppSettingsDTO({ ...prev, ...patch }));
  }, []);

  const persistSettings = useCallback(async (s: app.AppSettingsDTO) => {
    await AppAPI.SaveSettings(s);
  }, []);

  const refreshPreview = useCallback(async (c: app.ConfigDTO) => {
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
      eventsOn("queue:update", (q) => setQueue(asArray(q as app.QueueTaskDTO[]))),
      eventsOn("journal:add", (entry) => setJournal((prev) => [...prev, entry as string])),
      eventsOn("startup:ytdlp-missing", () => setShowYtDlpModal(true)),
      eventsOn("queue:scheduled", (at) => setScheduledQueueAt((at as number) || 0)),
      eventsOn("startup:other-instances", (list) => {
        const arr = asArray(list as app.InstanceDTO[]);
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

  const handleRunQueue = useCallback(async () => {
    setTaskProgress({});
    setRunning(true);
    try {
      await AppAPI.RunQueue();
    } catch {
      setRunning(false);
    }
  }, []);

  const handleReorderQueue = useCallback(
    (ids: string[]) => {
      const byId = new Map(queue.map((t) => [t.id, t]));
      setQueue(ids.map((id) => byId.get(id)!));
      void AppAPI.ReorderQueue(ids);
    },
    [queue],
  );

  const handleLanguageChange = useCallback(
    async (lang: string) => {
      const next = new app.AppSettingsDTO({ ...settingsRef.current, language: lang });
      updateSettings({ language: lang });
      setLocales(await AppAPI.GetLocales(lang));
      await AppAPI.SaveSettings(next);
    },
    [updateSettings],
  );

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
    const next = new app.AppSettingsDTO({ ...settingsRef.current, windowWidth: 0, windowHeight: 0 });
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
              <DownloadTab
                cfg={cfg}
                t={t}
                presets={presets}
                qualityPresets={qualityPresets}
                mergeFormats={mergeFormats}
                probingFormats={probingFormats}
                onProbeFormats={fetchFormats}
                updateConfig={updateConfig}
              />
            )}

            {tab === "network" && (
              <NetworkTab cfg={cfg} t={t} updateConfig={updateConfig} />
            )}

            {tab === "playlist" && (
              <PlaylistTab
                cfg={cfg}
                settings={settings}
                t={t}
                updateConfig={updateConfig}
                updateSettings={updateSettings}
                runSessionAction={runSessionAction}
              />
            )}

            {tab === "extras" && (
              <ExtrasTab cfg={cfg} t={t} updateConfig={updateConfig} />
            )}

            {tab === "queue" && (
              <QueueTab
                cfg={cfg}
                queue={queue}
                taskProgress={taskProgress}
                isRunning={isRunning}
                scheduledQueueAt={scheduledQueueAt}
                scheduleInput={scheduleInput}
                now={now}
                locales={locales}
                t={t}
                onRunQueue={handleRunQueue}
                onScheduleInputChange={setScheduleInput}
                onScheduleQueue={handleScheduleQueue}
                onCancelSchedule={handleCancelSchedule}
                onReorderQueue={handleReorderQueue}
                onEditTask={setEditingQueueTask}
              />
            )}

            {tab === "history" && (
              <HistoryTab
                history={history}
                t={t}
                onRequestClear={() => setConfirmClearHistory(true)}
                onUseUrl={(url) => updateConfig({ url })}
              />
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
              <SettingsTab
                settings={settings}
                version={version}
                locales={locales}
                t={t}
                updateSettings={updateSettings}
                onLanguageChange={handleLanguageChange}
                onResetWindowSize={resetWindowSize}
                onExportSettings={handleExportSettings}
                onImportSettings={handleImportSettings}
                onShowInstances={async () => {
                  await refreshOtherInstances();
                  setShowInstancesModal(true);
                }}
              />
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
