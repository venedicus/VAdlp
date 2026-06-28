import { describe, expect, it } from "vitest";
import { formatLabel } from "./formatLabel";
import type { FormatDTO } from "../types";

const base: FormatDTO = {
  ID: "137",
  Ext: "mp4",
  Resolution: "1920x1080",
  FPS: 30,
  Vcodec: "avc1",
  Acodec: "none",
  Filesize: 0,
  TBR: 0,
  Note: "",
};

describe("formatLabel", () => {
  it("always includes the format id", () => {
    expect(formatLabel(base)).toContain("137");
  });

  it("omits codecs reported as 'none'", () => {
    const label = formatLabel(base);
    expect(label).not.toContain("none");
  });

  it("includes both codecs when present", () => {
    const label = formatLabel({ ...base, Acodec: "mp4a" });
    expect(label).toContain("avc1");
    expect(label).toContain("mp4a");
  });

  it("rounds the bitrate and appends k", () => {
    expect(formatLabel({ ...base, TBR: 128.7 })).toContain("129k");
  });

  it("formats filesize in human-readable units", () => {
    expect(formatLabel({ ...base, Filesize: 500 })).toContain("500 B");
    expect(formatLabel({ ...base, Filesize: 1024 * 1024 * 5 })).toContain("5.0 MB");
  });

  it("includes a trailing note when present", () => {
    expect(formatLabel({ ...base, Note: "DASH video" })).toContain("DASH video");
  });
});
