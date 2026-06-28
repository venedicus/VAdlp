/** UI scale presets, keyed by screen size. */
export const UIScaleAuto = 0;
export const UIScaleCompact = 0.95;
export const UIScaleComfortable = 1.05;
export const UIScaleLarge = 1.15;

export function autoUIScale(screenW: number, screenH: number): number {
  if (screenW < 1360 || screenH < 820) return UIScaleCompact;
  if (screenW < 1680 || screenH < 980) return UIScaleComfortable;
  return UIScaleLarge;
}

export function effectiveUIScale(stored: number, screenW: number, screenH: number): number {
  if (stored <= 0 || stored === UIScaleAuto) {
    return autoUIScale(screenW, screenH);
  }
  return stored;
}

export const BASE_WINDOW_WIDTH = 1240;
export const BASE_WINDOW_HEIGHT = 820;
