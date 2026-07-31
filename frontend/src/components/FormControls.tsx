export function Field({
  label,
  children,
  wide,
}: {
  label: string;
  children: React.ReactNode;
  wide?: boolean;
}) {
  return (
    <div className={`form-row${wide ? " form-row-wide" : ""}`}>
      <label>{label}</label>
      {wide ? children : <div>{children}</div>}
    </div>
  );
}

export function Check({
  label,
  checked,
  onChange,
}: {
  label: string;
  checked: boolean;
  onChange: (v: boolean) => void;
}) {
  return (
    <label className="check-row">
      <input type="checkbox" checked={checked} onChange={(e) => onChange(e.target.checked)} />
      {label}
    </label>
  );
}

export function extractDroppedURL(data: DataTransfer): string {
  const raw = data.getData("text/uri-list") || data.getData("text/plain") || "";
  const line = raw.split(/\r?\n/).find((l) => l.trim() && !l.startsWith("#"));
  return line?.trim() ?? "";
}

export function looksLikeDownloadableURL(text: string): boolean {
  if (text.length > 2000 || text.includes("\n")) return false;
  return /^https?:\/\/\S+$/i.test(text.trim());
}

export function formatCountdown(ms: number): string {
  if (ms <= 0) return "0:00:00";
  const totalSec = Math.floor(ms / 1000);
  const h = Math.floor(totalSec / 3600);
  const m = Math.floor((totalSec % 3600) / 60);
  const s = totalSec % 60;
  return `${h}:${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
}
