import { describe, expect, it } from "vitest";
import { tf } from "./i18nFmt";
import type { LocaleMap } from "../types";

describe("tf", () => {
  const locales: LocaleMap = {
    "greeting": "Hello {{.Name}}",
    "plain": "No params here",
  };

  it("returns the raw id when not found in locales", () => {
    expect(tf(locales, "missing.id")).toBe("missing.id");
  });

  it("returns the translation unchanged when no params are given", () => {
    expect(tf(locales, "plain")).toBe("No params here");
  });

  it("substitutes a single placeholder", () => {
    expect(tf(locales, "greeting", { Name: "World" })).toBe("Hello World");
  });

  it("substitutes numeric values by stringifying them", () => {
    const withNumber: LocaleMap = { count: "{{.Count}} items" };
    expect(tf(withNumber, "count", { Count: 5 })).toBe("5 items");
  });

  it("substitutes every occurrence of a repeated placeholder", () => {
    const repeated: LocaleMap = { both: "{{.X}} and {{.X}} again" };
    expect(tf(repeated, "both", { X: "1" })).toBe("1 and 1 again");
  });
});
