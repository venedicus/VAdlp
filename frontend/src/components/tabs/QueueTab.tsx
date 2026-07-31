import { useRef, useState } from "react";
import type { ConfigDTO, DownloadProgressDTO, LocaleMap, QueueTaskDTO } from "../../types";
import { AppAPI } from "../../wailsjs/runtime";
import { tf } from "../../lib/i18nFmt";
import { formatCountdown } from "../FormControls";

export function QueueTab({
  cfg,
  queue,
  taskProgress,
  isRunning,
  scheduledQueueAt,
  scheduleInput,
  now,
  locales,
  t,
  onRunQueue,
  onScheduleInputChange,
  onScheduleQueue,
  onCancelSchedule,
  onReorderQueue,
  onEditTask,
}: {
  cfg: ConfigDTO;
  queue: QueueTaskDTO[];
  taskProgress: Record<string, DownloadProgressDTO>;
  isRunning: boolean;
  scheduledQueueAt: number;
  scheduleInput: string;
  now: number;
  locales: LocaleMap;
  t: (id: string) => string;
  onRunQueue: () => void;
  onScheduleInputChange: (v: string) => void;
  onScheduleQueue: () => void;
  onCancelSchedule: () => void;
  onReorderQueue: (ids: string[]) => void;
  onEditTask: (task: QueueTaskDTO) => void;
}) {
  const dragQueueIdRef = useRef<string | null>(null);
  const [selectedQueueId, setSelectedQueueId] = useState<string | null>(null);
  const [dragQueueOverId, setDragQueueOverId] = useState<string | null>(null);

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
    const ids = queue.map((task) => task.id);
    const from = ids.indexOf(draggedId);
    const to = ids.indexOf(targetId);
    if (from < 0 || to < 0) return;
    ids.splice(from, 1);
    ids.splice(to, 0, draggedId);
    onReorderQueue(ids);
  };

  const handleQueueDragEnd = () => {
    dragQueueIdRef.current = null;
    setDragQueueOverId(null);
  };

  return (
    <div className="form-grid">
      <div className="btn-row">
        <button type="button" className="btn btn-sm" onClick={() => AppAPI.AddToQueue(cfg)}>
          {t("btn.add_queue")}
        </button>
        <button type="button" className="btn btn-primary btn-sm" disabled={isRunning} onClick={onRunQueue}>
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
            <button type="button" className="btn btn-sm btn-danger" onClick={onCancelSchedule}>
              {t("btn.cancel_schedule")}
            </button>
          </>
        ) : (
          <>
            <input
              type="datetime-local"
              value={scheduleInput}
              onChange={(e) => onScheduleInputChange(e.target.value)}
            />
            <button
              type="button"
              className="btn btn-sm"
              disabled={!scheduleInput}
              onClick={onScheduleQueue}
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
              <span className="queue-drag-handle" title={t("queue.drag_hint")} aria-hidden>
                ⠿
              </span>
              <span className={`queue-status-dot ${task.status}`} aria-hidden />
              <span className={`status-pill ${task.status}`}>{t(`status.${task.status}`)}</span>
              <span className="queue-item-name" title={task.name}>
                {task.name}
              </span>
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
                  onClick={(e) => {
                    e.stopPropagation();
                    onEditTask(task);
                  }}
                >
                  {t("btn.edit_task")}
                </button>
              )}
              {task.status === "queued" && (
                <button
                  type="button"
                  className="btn btn-sm"
                  onClick={(e) => {
                    e.stopPropagation();
                    void AppAPI.PauseQueueTask(task.id);
                  }}
                >
                  {t("btn.pause_task")}
                </button>
              )}
              {task.status === "paused" && (
                <button
                  type="button"
                  className="btn btn-sm btn-primary"
                  onClick={(e) => {
                    e.stopPropagation();
                    void AppAPI.ResumeQueueTask(task.id);
                  }}
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
  );
}
