import { useCallback, useEffect, useRef } from "react";

type Props = {
  offset: number;
  onOffsetChange: (offset: number) => void;
  left: React.ReactNode;
  right: React.ReactNode;
  hidden?: boolean;
};

const MIN_LEFT = 0.35;
const MAX_LEFT = 0.82;

export function ResizableSplit({ offset, onOffsetChange, left, right, hidden }: Props) {
  const dragging = useRef(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const clampOffset = (v: number) => Math.min(MAX_LEFT, Math.max(MIN_LEFT, v));

  const onPointerDown = useCallback((e: React.PointerEvent) => {
    dragging.current = true;
    (e.target as HTMLElement).setPointerCapture(e.pointerId);
  }, []);

  const onPointerMove = useCallback(
    (e: React.PointerEvent) => {
      if (!dragging.current || !containerRef.current) return;
      const rect = containerRef.current.getBoundingClientRect();
      const next = (e.clientX - rect.left) / rect.width;
      onOffsetChange(clampOffset(next));
    },
    [onOffsetChange],
  );

  const onPointerUp = useCallback(() => {
    dragging.current = false;
  }, []);

  useEffect(() => {
    const stop = () => {
      dragging.current = false;
    };
    window.addEventListener("pointerup", stop);
    return () => window.removeEventListener("pointerup", stop);
  }, []);

  if (hidden) {
    return (
      <div className="main-layout activity-hidden">
        <div className="left-panel">{left}</div>
      </div>
    );
  }

  const leftPct = `${(clampOffset(offset) * 100).toFixed(2)}%`;

  return (
    <div
      ref={containerRef}
      className="main-layout resizable-split"
      style={{ gridTemplateColumns: `${leftPct} 6px 1fr` }}
    >
      <div className="left-panel">{left}</div>
      <div
        className="split-handle"
        role="separator"
        aria-orientation="vertical"
        onPointerDown={onPointerDown}
        onPointerMove={onPointerMove}
        onPointerUp={onPointerUp}
      />
      <div className="right-panel">{right}</div>
    </div>
  );
}
