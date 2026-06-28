import type { AppSettingsDTO } from "../types";

export function depSettingsPath(settings: AppSettingsDTO, id: string): string {
  switch (id) {
    case "ytdlp":
      return settings.ytDlpPath;
    case "ffmpeg":
      return settings.ffmpegPath;
    case "deno":
      return settings.denoPath;
    default:
      return "";
  }
}

export function patchDepPath(
  settings: AppSettingsDTO,
  id: string,
  path: string,
): Partial<AppSettingsDTO> {
  switch (id) {
    case "ytdlp":
      return { ytDlpPath: path, config: { ...settings.config, ytDlpPath: path } };
    case "ffmpeg":
      return { ffmpegPath: path, config: { ...settings.config, ffmpegLocation: path } };
    case "deno":
      return { denoPath: path, config: { ...settings.config, denoPath: path } };
    default:
      return {};
  }
}

export function mergeSettingsPatch(
  settings: AppSettingsDTO,
  patch: Partial<AppSettingsDTO>,
): AppSettingsDTO {
  const next = { ...settings, ...patch };
  if (patch.config) {
    next.config = { ...settings.config, ...patch.config };
  }
  return next;
}
