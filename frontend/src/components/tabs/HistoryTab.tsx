import { useMemo, useState } from "react";
import type { HistoryItemDTO } from "../../types";

export function HistoryTab({
  history,
  t,
  onRequestClear,
  onUseUrl,
}: {
  history: HistoryItemDTO[];
  t: (id: string) => string;
  onRequestClear: () => void;
  onUseUrl: (url: string) => void;
}) {
  const [historyQuery, setHistoryQuery] = useState("");
  const [historyStatusFilter, setHistoryStatusFilter] = useState<"all" | "completed" | "error" | "cancelled">("all");

  const filteredHistory = useMemo(() => {
    const q = historyQuery.trim().toLowerCase();
    return history.filter((item) => {
      if (historyStatusFilter !== "all" && item.status !== historyStatusFilter) return false;
      if (q && !item.url.toLowerCase().includes(q)) return false;
      return true;
    });
  }, [history, historyQuery, historyStatusFilter]);

  return (
    <div className="form-grid">
      <div className="btn-row">
        <input
          type="text"
          className="history-search"
          placeholder={t("history.search_placeholder")}
          value={historyQuery}
          onChange={(e) => setHistoryQuery(e.target.value)}
        />
        <select
          value={historyStatusFilter}
          onChange={(e) => setHistoryStatusFilter(e.target.value as typeof historyStatusFilter)}
        >
          <option value="all">{t("history.filter_all")}</option>
          <option value="completed">{t("status.completed")}</option>
          <option value="error">{t("status.error")}</option>
          <option value="cancelled">{t("status.cancelled")}</option>
        </select>
        <button type="button" className="btn btn-danger btn-sm" onClick={onRequestClear}>
          {t("btn.clear_history")}
        </button>
      </div>
      <div className="history-list">
        {filteredHistory.length === 0 && <div className="hint">{t("journal.empty")}</div>}
        {filteredHistory.map((item, i) => (
          <div
            key={`${item.at}-${i}`}
            className={`history-item ${item.status}`}
            role="button"
            tabIndex={0}
            onClick={() => item.url && onUseUrl(item.url)}
            onKeyDown={(e) => e.key === "Enter" && item.url && onUseUrl(item.url)}
          >
            <div className="history-url">{item.url}</div>
            <div className="history-meta">
              {t(`status.${item.status}`)} · {item.durationSec}s · {new Date(item.at).toLocaleString()}
              {item.error ? ` · ${item.error}` : ""}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
