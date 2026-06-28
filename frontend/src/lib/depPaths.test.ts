import { describe, expect, it } from "vitest";
import { defaultSettings } from "./defaults";
import { depSettingsPath, mergeSettingsPatch, patchDepPath } from "./depPaths";

describe("depSettingsPath", () => {
  const settings = { ...defaultSettings(), ytDlpPath: "/bin/yt-dlp", ffmpegPath: "/bin/ffmpeg", denoPath: "/bin/deno" };

  it("returns the right path per dependency id", () => {
    expect(depSettingsPath(settings, "ytdlp")).toBe("/bin/yt-dlp");
    expect(depSettingsPath(settings, "ffmpeg")).toBe("/bin/ffmpeg");
    expect(depSettingsPath(settings, "deno")).toBe("/bin/deno");
  });

  it("returns empty string for an unknown id", () => {
    expect(depSettingsPath(settings, "unknown")).toBe("");
  });
});

describe("patchDepPath", () => {
  const settings = defaultSettings();

  it("patches both the top-level path and the matching config field for ytdlp", () => {
    const patch = patchDepPath(settings, "ytdlp", "/new/yt-dlp");
    expect(patch.ytDlpPath).toBe("/new/yt-dlp");
    expect(patch.config?.ytDlpPath).toBe("/new/yt-dlp");
  });

  it("maps ffmpeg's path onto config.ffmpegLocation, not config.ffmpegPath", () => {
    const patch = patchDepPath(settings, "ffmpeg", "/new/ffmpeg");
    expect(patch.ffmpegPath).toBe("/new/ffmpeg");
    expect(patch.config?.ffmpegLocation).toBe("/new/ffmpeg");
  });

  it("returns an empty patch for an unknown id", () => {
    expect(patchDepPath(settings, "unknown", "/x")).toEqual({});
  });
});

describe("mergeSettingsPatch", () => {
  it("merges top-level fields", () => {
    const settings = defaultSettings();
    const merged = mergeSettingsPatch(settings, { ytDlpPath: "/x" });
    expect(merged.ytDlpPath).toBe("/x");
    expect(merged.denoPath).toBe(settings.denoPath);
  });

  it("merges config without dropping unrelated config fields", () => {
    const settings = { ...defaultSettings(), config: { ...defaultSettings().config, url: "https://keep" } };
    const merged = mergeSettingsPatch(settings, { config: { ...settings.config, ytDlpPath: "/x" } });
    expect(merged.config.url).toBe("https://keep");
    expect(merged.config.ytDlpPath).toBe("/x");
  });
});
