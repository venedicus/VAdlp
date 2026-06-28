import type { FormatDTO } from "../types";

export function formatLabel(f: FormatDTO): string {
  const parts: string[] = [f.ID];
  if (f.Resolution) parts.push(f.Resolution);
  if (f.Ext) parts.push(f.Ext);
  if (f.Vcodec && f.Vcodec !== "none") parts.push(f.Vcodec);
  if (f.Acodec && f.Acodec !== "none") parts.push(f.Acodec);
  if (f.TBR > 0) parts.push(`${Math.round(f.TBR)}k`);
  if (f.Filesize > 0) parts.push(humanSize(f.Filesize));
  if (f.Note) parts.push(f.Note);
  return parts.join(" · ");
}

function humanSize(n: number): string {
  const units = ["B", "KB", "MB", "GB", "TB"];
  let i = 0;
  let v = n;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i++;
  }
  return i === 0 ? `${n} B` : `${v.toFixed(1)} ${units[i]}`;
}
