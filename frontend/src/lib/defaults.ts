import type { AppSettingsDTO, ConfigDTO } from "../types";

export function defaultConfig(): ConfigDTO {
  return {
    url: "",
    quality: "best",
    format: "mp4",
    audioOnly: false,
    audioFormat: "",
    outputPath: "",
    outputTemplate: "%(upload_date)s - %(title)s.%(ext)s",
    useCookiesFile: false,
    cookiesFile: "",
    useCookiesBrowser: true,
    cookiesBrowser: "chrome",
    proxy: "",
    rateLimit: "",
    playlistReverse: true,
    continue: true,
    noPart: false,
    playlistStart: 0,
    playlistEnd: 0,
    maxDownloads: 0,
    downloadArchive: "",
    noPlaylist: false,
    flatPlaylist: false,
    writeSubs: false,
    writeAutoSub: false,
    embedSubs: false,
    subLangs: "en.*,ru.*",
    writeThumbnail: false,
    embedThumbnail: false,
    embedMetadata: false,
    embedChapters: false,
    retries: 10,
    fragmentRetries: 10,
    concurrentFragments: 1,
    socketTimeout: 0,
    noWarnings: false,
    verbose: false,
    quiet: false,
    writeInfoJSON: false,
    loadInfoJson: "",
    windowsFilenames: false,
    noMtime: false,
    abortOnError: false,
    ignoreErrors: false,
    extraArgs: "",
    ffmpegLocation: "",
    username: "",
    password: "",
    sponsorBlockRemove: false,
    batchUrls: "",
    ytDlpPath: "",
    denoPath: "",
  };
}

export function defaultSettings(): AppSettingsDTO {
  return {
    config: defaultConfig(),
    ffmpegPath: "",
    sessionPath: "",
    queueParallel: 1,
    language: "en",
    ytDlpPath: "",
    denoPath: "",
    lastProfile: "",
    debugLog: false,
    activityPanelOpen: true,
    uiScale: 0,
    theme: "auto",
    windowWidth: 0,
    windowHeight: 0,
    activityPanelOffset: 0.4,
  };
}

export function normalizeConfig(raw: Partial<ConfigDTO> | null | undefined): ConfigDTO {
  const base = defaultConfig();
  if (!raw || typeof raw !== "object") return base;
  return { ...base, ...raw };
}

export function normalizeSettings(raw: Partial<AppSettingsDTO> | null | undefined): AppSettingsDTO {
  const base = defaultSettings();
  if (!raw || typeof raw !== "object") return base;
  return {
    ...base,
    ...raw,
    config: normalizeConfig(raw.config),
    queueParallel: raw.queueParallel && raw.queueParallel > 0 ? raw.queueParallel : base.queueParallel,
    activityPanelOpen: raw.activityPanelOpen ?? base.activityPanelOpen,
    theme: raw.theme === "light" || raw.theme === "dark" || raw.theme === "auto" ? raw.theme : base.theme,
  };
}

export function asArray<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : [];
}
