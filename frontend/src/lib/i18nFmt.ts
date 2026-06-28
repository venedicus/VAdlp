import type { LocaleMap } from "../types";

export function tf(locales: LocaleMap, id: string, params?: Record<string, string | number>): string {
  let s = locales[id] ?? id;
  if (!params) return s;
  for (const [k, v] of Object.entries(params)) {
    s = s.replaceAll(`{{.${k}}}`, String(v)).replaceAll(`{{${k}}}`, String(v));
  }
  return s;
}
