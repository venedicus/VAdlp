import type {
  AppSettingsDTO,
  AppUpdateDTO,
  ConfigDTO,
  DependencyDTO,
  DownloadProgressDTO,
  HealthIssueDTO,
  HistoryItemDTO,
  InstallGuardDTO,
  InstanceDTO,
  LocaleMap,
  ProbeResultDTO,
  QueueTaskDTO,
} from "../types";
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
      settings: AppSettingsDTO;
      queue: QueueTaskDTO[];
      journal: string[];
      running: boolean;
      version: string;
      toolsDir: string;
      scheduledQueueAt: number;
    }>,
  SaveSettings: (s: AppSettingsDTO) => Generated.SaveSettings(toGen(s)),
  PreviewCommand: (cfg: ConfigDTO) => Generated.PreviewCommand(toGen(cfg)) as Promise<string>,
  ApplyPreset: (id: string) => Generated.ApplyPreset(id) as Promise<ConfigDTO>,
  RunDownload: (cfg: ConfigDTO) => Generated.RunDownload(toGen(cfg)),
  StopDownload: () => Generated.StopDownload(),
  AddToQueue: (cfg: ConfigDTO) => Generated.AddToQueue(toGen(cfg)) as Promise<QueueTaskDTO>,
  RemoveFromQueue: (id: string) => Generated.RemoveFromQueue(id),
  MoveQueueItem: (id: string, dir: number) => Generated.MoveQueueItem(id, dir),
  ReorderQueue: (ids: string[]) => Generated.ReorderQueue(ids),
  ExportSettings: (path: string) => Generated.ExportSettings(path),
  CheckAppUpdate: () => Generated.CheckAppUpdate() as Promise<AppUpdateDTO>,
  ListOtherInstances: () => Generated.ListOtherInstances() as Promise<InstanceDTO[]>,
  CloseIdleInstances: () => Generated.CloseIdleInstances() as Promise<number>,
  KillInstance: (pid: number) => Generated.KillInstance(pid),
  ScheduleQueueRun: (atUnixMillis: number) => Generated.ScheduleQueueRun(atUnixMillis),
  CancelScheduledQueueRun: () => Generated.CancelScheduledQueueRun(),
  GetScheduledQueueRun: () => Generated.GetScheduledQueueRun() as Promise<number>,
  ImportSettings: (path: string) => Generated.ImportSettings(path),
  UpdateQueueTask: (id: string, cfg: ConfigDTO) => Generated.UpdateQueueTask(id, toGen(cfg)),
  ClearQueue: () => Generated.ClearQueue(),
  RetryFailedQueue: () => Generated.RetryFailedQueue(),
  RunQueue: () => Generated.RunQueue(),
  GetHistory: () => Generated.GetHistory() as Promise<HistoryItemDTO[]>,
  ClearHistory: () => Generated.ClearHistory(),
  ListProfiles: () => Generated.ListProfiles() as Promise<string[]>,
  LoadProfile: (name: string) =>
    Generated.LoadProfile(name) as Promise<{ name: string; description: string; config: ConfigDTO }>,
  SaveProfile: (p: { name: string; description: string; config: ConfigDTO }) => Generated.SaveProfile(toGen(p)),
  DeleteProfile: (name: string) => Generated.DeleteProfile(name),
  RenameProfile: (old: string, n: string) => Generated.RenameProfile(old, n),
  CheckDependencies: () => Generated.CheckDependencies() as Promise<DependencyDTO[]>,
  ResolveDependenciesLocal: () => Generated.ResolveDependenciesLocal() as Promise<DependencyDTO[]>,
  CheckInstallGuard: (id: string) => Generated.CheckInstallGuard(id) as Promise<InstallGuardDTO>,
  InstallDependency: (id: string) => Generated.InstallDependency(id) as Promise<string>,
  UpdateDependency: (id: string) => Generated.UpdateDependency(id) as Promise<string>,
  ProbeFormats: (cfg: ConfigDTO) => Generated.ProbeFormats(toGen(cfg)) as Promise<ProbeResultDTO>,
  HealthCheck: () => Generated.HealthCheck() as Promise<HealthIssueDTO[]>,
  OpenFolder: (path: string) => Generated.OpenFolder(path),
  PickFolder: () => Generated.PickFolder() as Promise<string>,
  SaveSession: (path: string) => Generated.SaveSession(path),
  LoadSession: (path: string) => Generated.LoadSession(path) as Promise<ConfigDTO>,
  ResumeSession: (path: string) => Generated.ResumeSession(path) as Promise<ConfigDTO>,
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
