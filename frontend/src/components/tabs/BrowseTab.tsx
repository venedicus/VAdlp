import { useCallback, useRef, useState } from "react";
import type { app, browse } from "../../wailsjs/go/models";
import { AppAPI } from "../../wailsjs/runtime";
import { asArray } from "../../lib/defaults";
import { formatViewCount } from "../../lib/viewCount";

type Props = {
  config: app.ConfigDTO;
  t: (id: string) => string;
  /** Opens the format picker for a single video, then queues it. */
  onPickFormats: (url: string) => void;
  onAddToQueue: (url: string) => void;
  onPreview: (item: browse.Item) => void;
};

const LIMITS = [15, 30, 50, 100];

export function BrowseTab({ config, t, onPickFormats, onAddToQueue, onPreview }: Props) {
  const [query, setQuery] = useState("");
  const [limit, setLimit] = useState(30);
  const [result, setResult] = useState<browse.Result | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  // Breadcrumb of listings the user drilled into, so Back returns to the
  // search results rather than clearing them.
  const [trail, setTrail] = useState<browse.Result[]>([]);
  // Guards against an earlier, slower request overwriting a newer one.
  const requestSeq = useRef(0);

  const run = useCallback(
    async (fn: () => Promise<browse.Result>, pushTrail: boolean) => {
      const seq = ++requestSeq.current;
      setLoading(true);
      setError(null);
      try {
        const res = await fn();
        if (seq !== requestSeq.current) return;
        setResult((prev) => {
          if (pushTrail && prev) setTrail((tr) => [...tr, prev]);
          return res;
        });
      } catch (e) {
        if (seq !== requestSeq.current) return;
        setError(e instanceof Error ? e.message : String(e));
      } finally {
        if (seq === requestSeq.current) setLoading(false);
      }
    },
    [],
  );

  const search = useCallback(() => {
    if (!query.trim()) return;
    setTrail([]);
    void run(() => AppAPI.BrowseSearch(config, query, limit), false);
  }, [config, query, limit, run]);

  const openListing = useCallback(
    (url: string) => void run(() => AppAPI.BrowseOpen(config, url), true),
    [config, run],
  );

  const goBack = useCallback(() => {
    setTrail((tr) => {
      if (tr.length === 0) return tr;
      // Cancel any in-flight request so it cannot land after we go back.
      requestSeq.current++;
      setLoading(false);
      setResult(tr[tr.length - 1]);
      return tr.slice(0, -1);
    });
  }, []);

  const items = asArray(result?.items);

  return (
    <div className="form-grid browse-tab">
      <div className="btn-row browse-searchbar">
        {trail.length > 0 && (
          <button type="button" className="btn btn-sm" onClick={goBack}>
            ← {t("browse.back")}
          </button>
        )}
        <input
          type="text"
          className="browse-search-input"
          placeholder={t("browse.search_placeholder")}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") search();
          }}
        />
        <select
          value={limit}
          onChange={(e) => setLimit(parseInt(e.target.value, 10))}
          title={t("browse.limit")}
        >
          {LIMITS.map((n) => (
            <option key={n} value={n}>
              {n}
            </option>
          ))}
        </select>
        <button
          type="button"
          className="btn btn-primary btn-sm"
          onClick={search}
          disabled={loading || !query.trim()}
        >
          {loading ? t("browse.searching") : t("browse.search")}
        </button>
      </div>

      {error && <div className="browse-error">{t("browse.error")}: {error}</div>}

      {result?.title && !loading && (
        <div className="browse-listing-title">{result.title}</div>
      )}

      {!result && !loading && !error && <div className="hint">{t("browse.empty")}</div>}
      {result && !loading && items.length === 0 && (
        <div className="hint">{t("browse.no_results")}</div>
      )}

      <div className="browse-grid">
        {items.map((item, i) => (
          <BrowseCard
            key={item.id || item.url || i}
            item={item}
            t={t}
            onOpen={() => openListing(item.url)}
            onPickFormats={() => onPickFormats(item.url)}
            onAddToQueue={() => onAddToQueue(item.url)}
            onPreview={() => onPreview(item)}
          />
        ))}
      </div>
    </div>
  );
}

function BrowseCard({
  item,
  t,
  onOpen,
  onPickFormats,
  onAddToQueue,
  onPreview,
}: {
  item: browse.Item;
  t: (id: string) => string;
  onOpen: () => void;
  onPickFormats: () => void;
  onAddToQueue: () => void;
  onPreview: () => void;
}) {
  const isVideo = item.kind === "video";
  const meta = [
    item.uploader,
    item.viewCount > 0 ? `${formatViewCount(item.viewCount)} ${t("browse.views")}` : "",
  ]
    .filter(Boolean)
    .join(" · ");

  return (
    <div className="browse-card">
      <div className="browse-thumb-wrap">
        {item.thumbnail ? (
          <img className="browse-thumb" src={item.thumbnail} alt="" loading="lazy" />
        ) : (
          <div className="browse-thumb browse-thumb-empty" />
        )}
        {item.live && <span className="browse-badge browse-badge-live">{t("browse.live")}</span>}
        {!item.live && item.duration && <span className="browse-badge">{item.duration}</span>}
        {!isVideo && (
          <span className="browse-badge browse-badge-kind">
            {t(item.kind === "playlist" ? "browse.kind_playlist" : "browse.kind_channel")}
          </span>
        )}
      </div>
      <div className="browse-card-title" title={item.title}>
        {item.title}
      </div>
      {meta && <div className="browse-card-meta">{meta}</div>}
      <div className="browse-card-actions">
        {isVideo ? (
          <>
            <button type="button" className="btn btn-primary btn-sm" onClick={onPickFormats}>
              {t("browse.download")}
            </button>
            <button type="button" className="btn btn-sm" onClick={onAddToQueue}>
              {t("browse.add_queue")}
            </button>
            <button type="button" className="btn btn-sm" onClick={onPreview}>
              {t("browse.preview")}
            </button>
          </>
        ) : (
          <>
            <button type="button" className="btn btn-primary btn-sm" onClick={onOpen}>
              {t("browse.open_channel")}
            </button>
            <button type="button" className="btn btn-sm" onClick={onAddToQueue}>
              {t("browse.add_queue")}
            </button>
          </>
        )}
      </div>
    </div>
  );
}
