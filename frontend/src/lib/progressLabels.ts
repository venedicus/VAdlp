import { tf } from "./i18nFmt";
import type { DownloadProgressDTO, LocaleMap } from "../types";

export function fileProgressLabel(
  progress: DownloadProgressDTO | null,
  locales: LocaleMap,
  t: (id: string) => string,
): string {
  if (!progress) return t("progress.file");
  if (progress.speed || progress.eta) {
    return tf(locales, "progress.speed_eta", {
      Speed: progress.speed || "—",
      ETA: progress.eta || "—",
    });
  }
  return t("progress.file");
}

export type QueueOverallInfo = { pct: number; done: number; total: number };

/**
 * Aggregates whole-queue progress straight from queue task statuses + the
 * per-task progress map, independent of any single task's own file%. This is
 * what makes the "Overall" bar a true batch indicator instead of mirroring
 * whichever task happens to be reporting right now.
 */
export function queueOverallProgress(
  queue: { id: string; status: string }[],
  taskProgress: Record<string, { filePct: number }>,
): QueueOverallInfo | null {
  if (!queue.some((t) => t.status === "running")) return null;
  let total = 0;
  let done = 0;
  for (const task of queue) {
    if (task.status === "paused") continue;
    total++;
    if (task.status === "completed" || task.status === "error" || task.status === "cancelled") {
      done += 1;
    } else if (task.status === "running") {
      const tp = taskProgress[task.id];
      done += tp ? Math.min(1, Math.max(0, tp.filePct / 100)) : 0;
    }
  }
  if (total === 0) return null;
  return { pct: (done / total) * 100, done: Math.round(done), total };
}

export function queueOverallLabel(
  info: QueueOverallInfo,
  locales: LocaleMap,
): string {
  return tf(locales, "progress.queue_done", {
    Done: String(info.done),
    Total: String(info.total),
  });
}

export function overallProgressLabel(
  progress: DownloadProgressDTO | null,
  locales: LocaleMap,
  t: (id: string) => string,
): string {
  if (!progress) return t("progress.overall");
  const hasPlaylist = progress.plTotal > 0 && progress.plCurrent > 0;
  const hasQueue = progress.queueTotal > 1 && progress.queueIdx > 0;
  if (hasPlaylist && hasQueue) {
    return tf(locales, "progress.overall_queue_playlist", {
      QueueCurrent: String(progress.queueIdx),
      QueueTotal: String(progress.queueTotal),
      PlCurrent: String(progress.plCurrent),
      PlTotal: String(progress.plTotal),
    });
  }
  if (hasPlaylist) {
    return tf(locales, "progress.overall_playlist", {
      Current: String(progress.plCurrent),
      Total: String(progress.plTotal),
    });
  }
  if (hasQueue) {
    return tf(locales, "progress.overall_queue", {
      Current: String(progress.queueIdx),
      Total: String(progress.queueTotal),
    });
  }
  return t("progress.overall");
}
