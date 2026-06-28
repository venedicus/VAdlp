import { describe, expect, it } from "vitest";
import { asArray, defaultConfig, defaultSettings, normalizeConfig, normalizeSettings } from "./defaults";

describe("asArray", () => {
  it("passes through arrays", () => {
    expect(asArray([1, 2, 3])).toEqual([1, 2, 3]);
  });

  it("falls back to [] for null/undefined", () => {
    expect(asArray(null)).toEqual([]);
    expect(asArray(undefined)).toEqual([]);
  });
});

describe("normalizeConfig", () => {
  it("fills in missing fields from defaults", () => {
    const result = normalizeConfig({ url: "https://example.com" });
    expect(result.url).toBe("https://example.com");
    expect(result.quality).toBe(defaultConfig().quality);
  });

  it("returns defaults for null/non-object input", () => {
    expect(normalizeConfig(null)).toEqual(defaultConfig());
    expect(normalizeConfig(undefined)).toEqual(defaultConfig());
  });
});

describe("normalizeSettings", () => {
  it("returns defaults for null input", () => {
    expect(normalizeSettings(null)).toEqual(defaultSettings());
  });

  it("clamps queueParallel to at least 1", () => {
    const result = normalizeSettings({ queueParallel: 0 });
    expect(result.queueParallel).toBe(1);
  });

  it("keeps a valid queueParallel value", () => {
    const result = normalizeSettings({ queueParallel: 4 });
    expect(result.queueParallel).toBe(4);
  });

  it("accepts only known theme values, falling back to default otherwise", () => {
    expect(normalizeSettings({ theme: "light" }).theme).toBe("light");
    expect(normalizeSettings({ theme: "dark" }).theme).toBe("dark");
    expect(normalizeSettings({ theme: "auto" }).theme).toBe("auto");
    // @ts-expect-error intentionally invalid value to test the fallback
    expect(normalizeSettings({ theme: "neon" }).theme).toBe(defaultSettings().theme);
  });

  it("normalizes the nested config", () => {
    const result = normalizeSettings({ config: { ...defaultConfig(), url: "https://x" } });
    expect(result.config.url).toBe("https://x");
    expect(result.config.quality).toBe(defaultConfig().quality);
  });
});
