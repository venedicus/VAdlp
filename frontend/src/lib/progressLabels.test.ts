import { describe, expect, it } from "vitest";
import {
  fileProgressLabel,
  overallProgressLabel,
  queueOverallLabel,
  queueOverallProgress,
} from "./progressLabels";
import type { DownloadProgressDTO, LocaleMap } from "../types";

const locales: LocaleMap = {
  "progress.file": "This file",
  "progress.overall": "Everything in this run",
  "progress.overall_queue": "Queue: task {{.Current}} of {{.Total}}",
  "progress.overall_playlist": "Playlist: item {{.Current}} of {{.Total}}",
  "progress.overall_queue_playlist":
    "Queue: task {{.QueueCurrent}} of {{.QueueTotal}} · playlist item {{.PlCurrent}} of {{.PlTotal}}",
  "progress.queue_done": "Whole run: {{.Done}} of {{.Total}} done",
  "progress.speed_eta": "{{.Speed}}  ETA {{.ETA}}",
};
const t = (id: string) => locales[id] ?? id;

const baseProgress: DownloadProgressDTO = {
  filePct: 0,
  overallPct: 0,
  speed: "",
  eta: "",
  status: "running",
  phase: "downloading",
  plCurrent: 0,
  plTotal: 0,
  queueIdx: 0,
  queueTotal: 0,
};

describe("fileProgressLabel", () => {
  it("shows the generic label without progress", () => {
    expect(fileProgressLabel(null, locales, t)).toBe("This file");
  });

  it("shows speed/eta once available", () => {
    const label = fileProgressLabel({ ...baseProgress, speed: "1.2MiB/s", eta: "00:05" }, locales, t);
    expect(label).toBe("1.2MiB/s  ETA 00:05");
  });
});

describe("overallProgressLabel (single download / no active queue)", () => {
  it("falls back to the generic label with no playlist/queue info", () => {
    expect(overallProgressLabel(baseProgress, locales, t)).toBe("Everything in this run");
  });

  it("shows playlist position when present", () => {
    const label = overallProgressLabel({ ...baseProgress, plCurrent: 3, plTotal: 10 }, locales, t);
    expect(label).toBe("Playlist: item 3 of 10");
  });

  it("shows queue position when present", () => {
    const label = overallProgressLabel({ ...baseProgress, queueIdx: 2, queueTotal: 5 }, locales, t);
    expect(label).toBe("Queue: task 2 of 5");
  });

  it("combines queue and playlist position when both are present", () => {
    const label = overallProgressLabel(
      { ...baseProgress, queueIdx: 2, queueTotal: 5, plCurrent: 3, plTotal: 10 },
      locales,
      t,
    );
    expect(label).toBe("Queue: task 2 of 5 · playlist item 3 of 10");
  });
});

describe("queueOverallProgress", () => {
  it("returns null when nothing in the queue is running", () => {
    const queue = [
      { id: "a", status: "queued" },
      { id: "b", status: "completed" },
    ];
    expect(queueOverallProgress(queue, {})).toBeNull();
  });

  it("returns null for an empty queue", () => {
    expect(queueOverallProgress([], {})).toBeNull();
  });

  it("excludes paused tasks from the total", () => {
    const queue = [
      { id: "a", status: "running" },
      { id: "b", status: "paused" },
    ];
    const info = queueOverallProgress(queue, { a: { filePct: 50 } });
    expect(info?.total).toBe(1);
  });

  it("counts completed/error/cancelled as fully done", () => {
    const queue = [
      { id: "a", status: "running" },
      { id: "b", status: "completed" },
      { id: "c", status: "error" },
      { id: "d", status: "cancelled" },
    ];
    const info = queueOverallProgress(queue, { a: { filePct: 0 } });
    expect(info?.total).toBe(4);
    expect(info?.done).toBe(3);
  });

  it("blends the running task's own file percentage as a fraction, not a duplicate of 100%", () => {
    // 1 of 2 queue items finished, the other is 50% through its own file —
    // this is exactly the aggregation that fixes the old "duplicate progress bar" bug:
    // overall must reflect 1.5/2 = 75%, not just mirror the active file's 50%.
    const queue = [
      { id: "a", status: "completed" },
      { id: "b", status: "running" },
    ];
    const info = queueOverallProgress(queue, { b: { filePct: 50 } });
    expect(info?.pct).toBe(75);
  });

  it("treats a running task with no progress yet as 0% contribution", () => {
    const queue = [{ id: "a", status: "running" }];
    const info = queueOverallProgress(queue, {});
    expect(info?.pct).toBe(0);
  });
});

describe("queueOverallLabel", () => {
  it("formats the done/total counts", () => {
    expect(queueOverallLabel({ pct: 50, done: 2, total: 4 }, locales)).toBe("Whole run: 2 of 4 done");
  });
});
