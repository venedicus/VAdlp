import { useMemo, useState } from "react";
import { asArray } from "../lib/defaults";
import { formatLabel } from "../lib/formatLabel";
import type { MediaEntryDTO, ProbeResultDTO } from "../types";
import { Modal } from "./Modal";

type Props = {
  result: ProbeResultDTO;
  t: (id: string) => string;
  onPick: (formatId: string) => void;
  onClose: () => void;
};

export function FormatPickerModal({ result, t, onPick, onClose }: Props) {
  const entries = asArray(result.Entries);
  const [entryIdx, setEntryIdx] = useState(result.Selected >= 0 ? result.Selected : 0);
  const entry: MediaEntryDTO | undefined = entries[entryIdx];

  const formats = useMemo(() => {
    if (!entry) return [];
    return asArray(entry.Formats).filter((f) => f.ID);
  }, [entry]);

  const [formatId, setFormatId] = useState(formats[0]?.ID ?? "");

  if (!entry || formats.length === 0) {
    return (
      <Modal title={t("format.title")} onClose={onClose}>
        <p>{t("format.none")}</p>
        <div className="btn-row">
          <button type="button" className="btn" onClick={onClose}>
            {t("btn.close")}
          </button>
        </div>
      </Modal>
    );
  }

  return (
    <Modal title={t("format.title")} onClose={onClose} wide>
      {entries.length > 1 && (
        <Field label={t("form.playlist_item")}>
          <select
            value={entryIdx}
            onChange={(e) => {
              const idx = parseInt(e.target.value, 10);
              setEntryIdx(idx);
              const first = entries[idx]?.Formats?.find((f) => f.ID);
              setFormatId(first?.ID ?? "");
            }}
          >
            {entries.map((en, i) => (
              <option key={en.ID || i} value={i}>
                {en.Title || en.URL || `#${i + 1}`}
              </option>
            ))}
          </select>
        </Field>
      )}

      <div className="format-picker-meta">
        <div className="format-picker-title">{entry.Title}</div>
        <div className="format-picker-sub">
          {[entry.Uploader, entry.Duration].filter(Boolean).join(" · ")}
        </div>
        {entry.Thumbnail && (
          <img className="format-picker-thumb" src={entry.Thumbnail} alt="" loading="lazy" />
        )}
      </div>

      <Field label={t("form.format")}>
        <select value={formatId} onChange={(e) => setFormatId(e.target.value)}>
          {formats.map((f) => (
            <option key={f.ID} value={f.ID}>
              {formatLabel(f)}
            </option>
          ))}
        </select>
      </Field>

      <div className="btn-row">
        <button
          type="button"
          className="btn btn-primary"
          disabled={!formatId}
          onClick={() => {
            onPick(formatId);
            onClose();
          }}
        >
          {t("format.use")}
        </button>
        <button type="button" className="btn" onClick={onClose}>
          {t("btn.cancel")}
        </button>
      </div>
    </Modal>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="form-row form-row-wide modal-field">
      <label>{label}</label>
      {children}
    </div>
  );
}
