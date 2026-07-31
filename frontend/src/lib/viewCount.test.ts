import { describe, expect, it } from "vitest";
import { formatViewCount } from "./viewCount";

describe("formatViewCount", () => {
  it("returns small counts unchanged", () => {
    expect(formatViewCount(0)).toBe("");
    expect(formatViewCount(1)).toBe("1");
    expect(formatViewCount(999)).toBe("999");
  });

  it("abbreviates thousands, millions and billions", () => {
    expect(formatViewCount(1000)).toBe("1K");
    expect(formatViewCount(1234)).toBe("1.2K");
    expect(formatViewCount(12345)).toBe("12K");
    expect(formatViewCount(4500000)).toBe("4.5M");
    expect(formatViewCount(412345)).toBe("412K");
    expect(formatViewCount(2000000000)).toBe("2B");
  });

  it("drops a trailing .0", () => {
    expect(formatViewCount(2000)).toBe("2K");
    expect(formatViewCount(3000000)).toBe("3M");
  });

  it("ignores invalid input", () => {
    expect(formatViewCount(NaN)).toBe("");
    expect(formatViewCount(-5)).toBe("");
  });
});
