export interface ConfigDTO {
  url: string;
  quality: string;
  format: string;
  audioOnly: boolean;
  audioFormat: string;
  outputPath: string;
  outputTemplate: string;
  useCookiesFile: boolean;
  cookiesFile: string;
  useCookiesBrowser: boolean;
  cookiesBrowser: string;
  proxy: string;
  rateLimit: string;
  playlistReverse: boolean;
  continue: boolean;
  noPart: boolean;
  playlistStart: number;
  playlistEnd: number;
  maxDownloads: number;
  downloadArchive: string;
  noPlaylist: boolean;
  flatPlaylist: boolean;
  writeSubs: boolean;
  writeAutoSub: boolean;
  embedSubs: boolean;
  subLangs: string;
  writeThumbnail: boolean;
  embedThumbnail: boolean;
  embedMetadata: boolean;
  embedChapters: boolean;
  retries: number;
  fragmentRetries: number;
  concurrentFragments: number;
  socketTimeout: number;
  noWarnings: boolean;
  verbose: boolean;
  quiet: boolean;
  writeInfoJSON: boolean;
  loadInfoJson: string;
  windowsFilenames: boolean;
  noMtime: boolean;
  abortOnError: boolean;
  ignoreErrors: boolean;
  extraArgs: string;
  ffmpegLocation: string;
  username: string;
  password: string;
  sponsorBlockRemove: boolean;
  batchUrls: string;
  ytDlpPath: string;
  denoPath: string;
}

export interface AppSettingsDTO {
  config: ConfigDTO;
  ffmpegPath: string;
  sessionPath: string;
  queueParallel: number;
  language: string;
  ytDlpPath: string;
  denoPath: string;
  lastProfile: string;
  debugLog: boolean;
  activityPanelOpen: boolean;
  uiScale: number;
  theme: "dark" | "light" | "auto";
  windowWidth: number;
  windowHeight: number;
  activityPanelOffset: number;
}

export interface AppStateDTO {
  settings: AppSettingsDTO;
  queue: QueueTaskDTO[];
  journal: string[];
  running: boolean;
  version: string;
  toolsDir: string;
  scheduledQueueAt: number;
}

export interface QueueTaskDTO {
  id: string;
  name: string;
  config: ConfigDTO;
  status: string;
}

export interface DependencyDTO {
  id: string;
  level: string;
  status: string;
  path: string;
  version: string;
  latestVer: string;
  updateAvail: boolean;
  source: string;
  error: string;
}

export interface InstallGuardDTO {
  alreadyInstalled: boolean;
  foundElsewhere: boolean;
  path: string;
}

export interface HistoryItemDTO {
  at: string;
  url: string;
  status: string;
  output: string;
  error: string;
  durationSec: number;
}

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

export interface FormatDTO {
  ID: string;
  Ext: string;
  Resolution: string;
  FPS: number;
  Vcodec: string;
  Acodec: string;
  Filesize: number;
  TBR: number;
  Note: string;
}

export interface MediaEntryDTO {
  Title: string;
  ID: string;
  URL: string;
  Duration: string;
  Uploader: string;
  Thumbnail: string;
  Formats: FormatDTO[];
}

export interface ProbeResultDTO {
  Title: string;
  Kind: string;
  Entries: MediaEntryDTO[];
  Selected: number;
}

export interface HealthIssueDTO {
  id: string;
  severity: string;
  key: string;
  params: Record<string, string | number>;
}

export interface AppUpdateDTO {
  current: string;
  latest: string;
  updateAvail: boolean;
  url: string;
}

export interface InstanceDTO {
  pid: number;
  startedAt: string;
  busy: boolean;
}
