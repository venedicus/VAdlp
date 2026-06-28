import { useEffect, useRef } from "react";
import { WindowGetSize } from "../wailsjs/runtime/runtime";
import type { AppSettingsDTO } from "../types";

type Args = {
  bootstrapping: boolean;
  settings: AppSettingsDTO;
  onBoundsChange: (patch: Partial<AppSettingsDTO>) => void;
  onCompactChange: (compact: boolean) => void;
};

export function useWindowBounds({ bootstrapping, settings, onBoundsChange, onCompactChange }: Args) {
  const settingsRef = useRef(settings);
  settingsRef.current = settings;
  const saveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (bootstrapping) return;

    const applySize = (w: number, h: number) => {
      onCompactChange(w < 1100);
      if (saveTimerRef.current) clearTimeout(saveTimerRef.current);
      saveTimerRef.current = setTimeout(() => {
        const cur = settingsRef.current;
        if (Math.abs(cur.windowWidth - w) < 2 && Math.abs(cur.windowHeight - h) < 2) return;
        onBoundsChange({ windowWidth: w, windowHeight: h });
      }, 400);
    };

    const onResize = () => {
      void WindowGetSize()
        .then((s) => applySize(s.w, s.h))
        .catch(() => applySize(window.innerWidth, window.innerHeight));
    };

    onResize();
    window.addEventListener("resize", onResize);
    return () => {
      window.removeEventListener("resize", onResize);
      if (saveTimerRef.current) clearTimeout(saveTimerRef.current);
    };
  }, [bootstrapping, onBoundsChange, onCompactChange]);
}
