import type {
  app,
  downloader,
} from "./go/models";
import type { DownloadProgressDTO, LocaleMap } from "../lib/eventTypes";
import * as Generated from "./go/app/App";
import { EventsOn } from "./runtime/runtime";

function toGen(v: unknown): never {
  return v as never;
}

export function waitForWailsRuntime(timeoutMs = 8000): Promise<void> {
  if (typeof window !== "undefined" && window.go?.app?.App) {
    return Promise.resolve();
  }
  return new Promise((resolve, reject) => {
    const started = Date.now();
    const tick = () => {
      if (window.go?.app?.App) {
        resolve();
        return;
      }
      if (Date.now() - started >= timeoutMs) {
        reject(new Error("Wails runtime not available"));
        return;
      }
      requestAnimationFrame(tick);
    };
    tick();
  });
}

export const AppAPI = {
  GetState: () =>
    Generated.GetState() as Promise<{
      settings: app.AppSettingsDTO;
      queue: app.QueueTaskDTO[];
      journal: string[];
      running: boolean;
      version: string;
      toolsDir: string;
      scheduledQueueAt: number;
    }>,
  SaveSettings: (s: app.AppSettingsDTO) => Generated.SaveSettings(toGen(s)),
  PreviewCommand: (cfg: app.ConfigDTO) => Generated.PreviewCommand(toGen(cfg)) as Promise<string>,
  ApplyPreset: (id: string) => Generated.ApplyPreset(id) as Promise<app.ConfigDTO>,
  RunDownload: (cfg: app.ConfigDTO) => Generated.RunDownload(toGen(cfg)),
  StopDownload: () => Generated.StopDownload(),
  AddToQueue: (cfg: app.ConfigDTO) => Generated.AddToQueue(toGen(cfg)) as Promise<app.QueueTaskDTO>,
  RemoveFromQueue: (id: string) => Generated.RemoveFromQueue(id),
  MoveQueueItem: (id: string, dir: number) => Generated.MoveQueueItem(id, dir),
  ReorderQueue: (ids: string[]) => Generated.ReorderQueue(ids),
  ExportSettings: (path: string) => Generated.ExportSettings(path),
  CheckAppUpdate: () => Generated.CheckAppUpdate() as Promise<app.AppUpdateDTO>,
  ListOtherInstances: () => Generated.ListOtherInstances() as Promise<app.InstanceDTO[]>,
  CloseIdleInstances: () => Generated.CloseIdleInstances() as Promise<number>,
  KillInstance: (pid: number) => Generated.KillInstance(pid),
  ScheduleQueueRun: (atUnixMillis: number) => Generated.ScheduleQueueRun(atUnixMillis),
  CancelScheduledQueueRun: () => Generated.CancelScheduledQueueRun(),
  GetScheduledQueueRun: () => Generated.GetScheduledQueueRun() as Promise<number>,
  ImportSettings: (path: string) => Generated.ImportSettings(path),
  UpdateQueueTask: (id: string, cfg: app.ConfigDTO) => Generated.UpdateQueueTask(id, toGen(cfg)),
  ClearQueue: () => Generated.ClearQueue(),
  RetryFailedQueue: () => Generated.RetryFailedQueue(),
  RunQueue: () => Generated.RunQueue(),
  GetHistory: () => Generated.GetHistory() as Promise<app.HistoryItemDTO[]>,
  ClearHistory: () => Generated.ClearHistory(),
  ListProfiles: () => Generated.ListProfiles() as Promise<string[]>,
  LoadProfile: (name: string) =>
    Generated.LoadProfile(name) as Promise<{ name: string; description: string; config: app.ConfigDTO }>,
  SaveProfile: (p: { name: string; description: string; config: app.ConfigDTO }) => Generated.SaveProfile(toGen(p)),
  DeleteProfile: (name: string) => Generated.DeleteProfile(name),
  RenameProfile: (old: string, n: string) => Generated.RenameProfile(old, n),
  CheckDependencies: () => Generated.CheckDependencies() as Promise<app.DependencyDTO[]>,
  ResolveDependenciesLocal: () => Generated.ResolveDependenciesLocal() as Promise<app.DependencyDTO[]>,
  CheckInstallGuard: (id: string) => Generated.CheckInstallGuard(id) as Promise<app.InstallGuardDTO>,
  InstallDependency: (id: string) => Generated.InstallDependency(id) as Promise<string>,
  UpdateDependency: (id: string) => Generated.UpdateDependency(id) as Promise<string>,
  ProbeFormats: (cfg: app.ConfigDTO) => Generated.ProbeFormats(toGen(cfg)) as Promise<downloader.ProbeResult>,
  HealthCheck: () => Generated.HealthCheck() as Promise<app.HealthIssueDTO[]>,
  OpenFolder: (path: string) => Generated.OpenFolder(path),
  PickFolder: () => Generated.PickFolder() as Promise<string>,
  SaveSession: (path: string) => Generated.SaveSession(path),
  LoadSession: (path: string) => Generated.LoadSession(path) as Promise<app.ConfigDTO>,
  ResumeSession: (path: string) => Generated.ResumeSession(path) as Promise<app.ConfigDTO>,
  GetLocales: (lang: string) => Generated.GetLocales(lang) as Promise<LocaleMap>,
  GetPresets: () => Generated.GetPresets() as Promise<string[]>,
  GetQualityPresets: () => Generated.GetQualityPresets() as Promise<{ key: string; value: string }[]>,
  GetMergeFormats: () => Generated.GetMergeFormats() as Promise<string[]>,
  PickFile: () => Generated.PickFile() as Promise<string>,
  PickSaveFile: (defaultName: string) => Generated.PickSaveFile(defaultName) as Promise<string>,
  CancelQueueTask: (id: string) => Generated.CancelQueueTask(id) as Promise<boolean>,
  PauseQueueTask: (id: string) => Generated.PauseQueueTask(id),
  ResumeQueueTask: (id: string) => Generated.ResumeQueueTask(id),
};

export function eventsOn(event: string, cb: (...args: unknown[]) => void) {
  if (typeof EventsOn === "function") {
    return EventsOn(event, cb);
  }
  return () => {};
}

export type { DownloadProgressDTO };
