import { describe, expect, it } from "vitest";
import { youtubeEmbedURL } from "./BrowsePreviewModal";

describe("youtubeEmbedURL", () => {
  const EMBED = "https://www.youtube-nocookie.com/embed/oV9rvDllKEg";

  it("uses a bare 11-character video id", () => {
    expect(youtubeEmbedURL({ id: "oV9rvDllKEg" })).toBe(EMBED);
  });

  it("extracts the id from watch, short, embed and shorts URLs", () => {
    const urls = [
      "https://www.youtube.com/watch?v=oV9rvDllKEg",
      "https://youtube.com/watch?v=oV9rvDllKEg&list=PL1",
      "https://m.youtube.com/watch?v=oV9rvDllKEg",
      "https://youtu.be/oV9rvDllKEg",
      "https://www.youtube.com/embed/oV9rvDllKEg",
      "https://www.youtube.com/shorts/oV9rvDllKEg",
      "https://www.youtube.com/live/oV9rvDllKEg",
    ];
    for (const url of urls) {
      expect(youtubeEmbedURL({ url }), url).toBe(EMBED);
    }
  });

  it("prefers the id field over the URL", () => {
    expect(youtubeEmbedURL({ id: "oV9rvDllKEg", url: "https://example.com/x" })).toBe(EMBED);
  });

  it("returns empty for non-YouTube and malformed input", () => {
    expect(youtubeEmbedURL({})).toBe("");
    expect(youtubeEmbedURL({ url: "" })).toBe("");
    expect(youtubeEmbedURL({ url: "not a url" })).toBe("");
    expect(youtubeEmbedURL({ url: "https://vimeo.com/12345" })).toBe("");
    expect(youtubeEmbedURL({ url: "https://www.youtube.com/@someone" })).toBe("");
    expect(youtubeEmbedURL({ url: "https://www.youtube.com/playlist?list=PL1" })).toBe("");
  });

  it("rejects ids of the wrong length", () => {
    expect(youtubeEmbedURL({ id: "tooshort" })).toBe("");
    expect(youtubeEmbedURL({ id: "waaaaaytoolongid" })).toBe("");
  });
});
