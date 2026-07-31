/**
 * Formats a view count compactly: 1234 -> "1.2K", 4500000 -> "4.5M".
 * Values below 1000 are returned as-is.
 */
export function formatViewCount(n: number): string {
  if (!Number.isFinite(n) || n <= 0) return "";
  const units: [number, string][] = [
    [1e9, "B"],
    [1e6, "M"],
    [1e3, "K"],
  ];
  for (const [size, suffix] of units) {
    if (n >= size) {
      const scaled = n / size;
      // One decimal below 10 ("1.2M"), none above ("12M").
      return (scaled < 10 ? scaled.toFixed(1).replace(/\.0$/, "") : Math.round(scaled).toString()) + suffix;
    }
  }
  return Math.round(n).toString();
}
