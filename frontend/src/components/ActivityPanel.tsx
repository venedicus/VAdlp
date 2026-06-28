import { useCallback, useState } from "react";
import { fileProgressLabel, overallProgressLabel, queueOverallLabel, type QueueOverallInfo } from "../lib/progressLabels";
import type { DownloadProgressDTO, LocaleMap } from "../types";

type SectionId = "command" | "progress" | "log";

type Props = {
  command: string;
  logs: string;
  progress: DownloadProgressDTO | null;
  queueOverall: QueueOverallInfo | null;
  locales: LocaleMap;
  t: (id: string) => string;
};

function AccordionSection({
  id,
  title,
  open,
  onToggle,
  children,
  className,
}: {
  id: SectionId;
  title: string;
  open: boolean;
  onToggle: (id: SectionId) => void;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <div className={`activity-accordion${className ? ` ${className}` : ""}`}>
      <button type="button" className="activity-accordion-head" onClick={() => onToggle(id)}>
        <span className={`activity-accordion-chevron${open ? " open" : ""}`}>▸</span>
        {title}
      </button>
      {open && <div className="activity-accordion-body">{children}</div>}
    </div>
  );
}

export function ActivityPanel({ command, logs, progress, queueOverall, locales, t }: Props) {
  const [openSections, setOpenSections] = useState<Record<SectionId, boolean>>({
    command: true,
    progress: true,
    log: true,
  });

  const toggle = useCallback((id: SectionId) => {
    setOpenSections((prev) => ({ ...prev, [id]: !prev[id] }));
  }, []);

  const fileLabel = fileProgressLabel(progress, locales, t);
  const overallLabel = queueOverall ? queueOverallLabel(queueOverall, locales) : overallProgressLabel(progress, locales, t);
  const overallPct = queueOverall ? queueOverall.pct : progress?.overallPct ?? 0;

  return (
    <aside className="activity-panel-inner">
      <AccordionSection id="command" title={t("command.title")} open={openSections.command} onToggle={toggle}>
        <textarea className="command-preview readonly" readOnly value={command} rows={4} />
      </AccordionSection>

      <AccordionSection id="progress" title={t("activity.title")} open={openSections.progress} onToggle={toggle}>
        <div className="progress-block">
          <div className="progress-label">
            <span>{fileLabel}</span>
          </div>
          <div className="progress-track">
            <div className="progress-fill" style={{ width: `${progress?.filePct ?? 0}%` }} />
          </div>
        </div>
        <div className="progress-block">
          <div className="progress-label">
            <span>{overallLabel}</span>
            <span>{Math.round(overallPct)}%</span>
          </div>
          <div className="progress-track">
            <div
              className="progress-fill"
              style={{ width: `${overallPct}%`, opacity: 0.7 }}
            />
          </div>
        </div>
      </AccordionSection>

      <AccordionSection
        id="log"
        title={t("log.title")}
        open={openSections.log}
        onToggle={toggle}
        className="log-accordion"
      >
        <textarea className="log-area readonly" readOnly value={logs} />
      </AccordionSection>
    </aside>
  );
}
