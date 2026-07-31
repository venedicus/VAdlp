import type { browse } from "../wailsjs/go/models";
import { BrowserOpenURL } from "../wailsjs/runtime/runtime";
import { Modal } from "./Modal";

/**
 * Plays a video in YouTube's official embed player.
 *
 * The embed is the only sanctioned way to show YouTube inside the app —
 * loading the watch page in a frame is blocked by X-Frame-Options, and the
 * player exposes no stream URLs, so downloads still go through yt-dlp.
 * Uploaders can disable embedding, hence the "open in browser" fallback.
 */
export function BrowsePreviewModal({
  item,
  t,
  onClose,
  onDownload,
}: {
  item: browse.Item;
  t: (id: string) => string;
  onClose: () => void;
  onDownload: () => void;
}) {
  const embedURL = youtubeEmbedURL(item);

  return (
    <Modal title={item.title || t("browse.preview")} onClose={onClose} wide>
      {embedURL ? (
        <div className="browse-preview-frame">
          <iframe
            src={embedURL}
            title={item.title || "preview"}
            allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
            allowFullScreen
            referrerPolicy="strict-origin-when-cross-origin"
          />
        </div>
      ) : (
        <p>{t("browse.preview_unavailable")}</p>
      )}

      <div className="browse-preview-meta">
        {[item.uploader, item.duration].filter(Boolean).join(" · ")}
      </div>

      <div className="btn-row">
        <button type="button" className="btn btn-primary" onClick={onDownload}>
          {t("browse.download")}
        </button>
        {item.url && (
          <button type="button" className="btn" onClick={() => BrowserOpenURL(item.url)}>
            {t("browse.open_external")}
          </button>
        )}
        <button type="button" className="btn" onClick={onClose}>
          {t("btn.close")}
        </button>
      </div>
    </Modal>
  );
}

/**
 * Builds the privacy-preserving embed URL for a YouTube video id.
 * Returns "" for non-YouTube items, where no embed player applies.
 */
export function youtubeEmbedURL(item: { id?: string; url?: string }): string {
  const id = youtubeVideoID(item);
  return id ? `https://www.youtube-nocookie.com/embed/${id}` : "";
}

function youtubeVideoID(item: { id?: string; url?: string }): string {
  const id = (item.id ?? "").trim();
  if (/^[A-Za-z0-9_-]{11}$/.test(id)) return id;

  const raw = (item.url ?? "").trim();
  if (!raw) return "";
  let parsed: URL;
  try {
    parsed = new URL(raw);
  } catch {
    return "";
  }
  const host = parsed.hostname.replace(/^www\./, "");
  if (host === "youtu.be") {
    const seg = parsed.pathname.slice(1);
    return /^[A-Za-z0-9_-]{11}$/.test(seg) ? seg : "";
  }
  if (host === "youtube.com" || host === "m.youtube.com" || host === "youtube-nocookie.com") {
    const v = parsed.searchParams.get("v") ?? "";
    if (/^[A-Za-z0-9_-]{11}$/.test(v)) return v;
    const m = parsed.pathname.match(/^\/(?:embed|shorts|live)\/([A-Za-z0-9_-]{11})/);
    if (m) return m[1];
  }
  return "";
}
