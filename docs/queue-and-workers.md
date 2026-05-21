# Queue and parallel workers

## Queue states

Each row: `queued` → `running` → `completed` | `error` | `cancelled`.

- **Cancel (×)** on a running row calls `downloader.CancelJob(taskID)`.
- **Retry failed** resets `error` / `cancelled` to `queued`.

## Workers (`QueueParallel` in Tools)

| Workers | Activity panel (right) | Queue list |
|--------|-------------------------|------------|
| **1** | Shows log, progress, badges for the active job | Status updates per task |
| **>1** | Stays on the last focused job; does not switch per task | Each row shows its own status/color |

Parallel runs use a semaphore (`QueueParallel` concurrent goroutines). Overall progress in the activity panel is only reliable with **one worker**; with more workers, rely on per-row status.

## Defaults

- `QueueParallel` defaults to **1** if unset or invalid (validated 1–32 on save).
