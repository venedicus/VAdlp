import { describe, expect, it } from "vitest";
import { autoUIScale, effectiveUIScale, UIScaleAuto, UIScaleCompact, UIScaleComfortable, UIScaleLarge } from "./uiScale";

describe("autoUIScale", () => {
  it("picks compact for small screens", () => {
    expect(autoUIScale(1280, 720)).toBe(UIScaleCompact);
  });

  it("picks comfortable for medium screens", () => {
    expect(autoUIScale(1440, 900)).toBe(UIScaleComfortable);
  });

  it("picks large for big screens", () => {
    expect(autoUIScale(1920, 1080)).toBe(UIScaleLarge);
  });
});

describe("effectiveUIScale", () => {
  it("falls back to auto-detected scale when stored is auto (0)", () => {
    expect(effectiveUIScale(UIScaleAuto, 1280, 720)).toBe(autoUIScale(1280, 720));
  });

  it("falls back to auto-detected scale for any non-positive stored value", () => {
    expect(effectiveUIScale(-1, 1920, 1080)).toBe(autoUIScale(1920, 1080));
  });

  it("uses the stored scale when explicitly set", () => {
    expect(effectiveUIScale(1.25, 1280, 720)).toBe(1.25);
  });
});
