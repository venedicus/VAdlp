import { app } from "../wailsjs/go/models";

export function depSettingsPath(settings: app.AppSettingsDTO, id: string): string {
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
  settings: app.AppSettingsDTO,
  id: string,
  path: string,
): Partial<app.AppSettingsDTO> {
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
  settings: app.AppSettingsDTO,
  patch: Partial<app.AppSettingsDTO>,
): app.AppSettingsDTO {
  const config = patch.config ? { ...settings.config, ...patch.config } : settings.config;
  return new app.AppSettingsDTO({ ...settings, ...patch, config });
}
