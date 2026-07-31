export interface DownloadProgressDTO {
  filePct: number;
  overallPct: number;
  speed: string;
  eta: string;
  status: string;
  phase: string;
  plCurrent: number;
  plTotal: number;
  queueIdx: number;
  queueTotal: number;
  taskId?: string;
}

export type LocaleMap = Record<string, string>;
