import { useMemo, useState } from "react";
import type { ConfigDTO } from "../../types";
import { AppAPI } from "../../wailsjs/runtime";
import { Check, Field } from "../FormControls";

const AUDIO_FORMATS = ["", "mp3", "m4a", "opus", "wav", "flac", "vorbis", "aac", "alac"];

export function DownloadTab({
  cfg,
  t,
  presets,
  qualityPresets,
  mergeFormats,
  probingFormats,
  onProbeFormats,
  updateConfig,
}: {
  cfg: ConfigDTO;
  t: (id: string) => string;
  presets: string[];
  qualityPresets: { key: string; value: string }[];
  mergeFormats: string[];
  probingFormats: boolean;
  onProbeFormats: () => void;
  updateConfig: (patch: Partial<ConfigDTO>) => void;
}) {
  const [qualityPresetKey, setQualityPresetKey] = useState(0);

  const presetLabels: Record<string, string> = useMemo(
    () => ({
      youtube_playlist: t("preset.youtube_playlist"),
      audio_only: t("preset.audio_only"),
      video_best: t("preset.video_best"),
      video_1080: t("preset.video_1080"),
      video_4k: t("preset.video_4k"),
      podcast: t("preset.podcast"),
    }),
    [t],
  );

  const containerOptions = useMemo(() => {
    const opts = [...mergeFormats];
    if (cfg.format && !opts.includes(cfg.format)) opts.push(cfg.format);
    return opts;
  }, [mergeFormats, cfg.format]);

  const pasteURL = async () => {
    try {
      const text = await navigator.clipboard.readText();
      if (text.trim()) updateConfig({ url: text.trim() });
    } catch {
      /* clipboard unavailable */
    }
  };

  return (
    <div className="form-grid">
      <Field label={t("card.url")} wide>
        <div className="input-with-actions">
          <input
            type="text"
            placeholder={t("placeholder.url")}
            value={cfg.url}
            onChange={(e) => updateConfig({ url: e.target.value })}
          />
          <button type="button" className="btn btn-sm" onClick={pasteURL}>
            {t("btn.paste")}
          </button>
          <button
            type="button"
            className="btn btn-sm"
            disabled={!cfg.url.trim() || probingFormats}
            onClick={onProbeFormats}
          >
            {probingFormats ? t("btn.working") : t("btn.fetch_formats")}
          </button>
        </div>
      </Field>
      <Field label={t("card.batch")} wide>
        <textarea
          placeholder={t("placeholder.batch_urls")}
          value={cfg.batchUrls}
          onChange={(e) => updateConfig({ batchUrls: e.target.value })}
        />
      </Field>
      <Field label={t("card.output")}>
        <div className="input-with-actions">
          <input
            type="text"
            value={cfg.outputPath}
            onChange={(e) => updateConfig({ outputPath: e.target.value })}
          />
          <button
            type="button"
            className="btn btn-sm"
            onClick={async () => {
              const p = await AppAPI.PickFolder();
              if (p) updateConfig({ outputPath: p });
            }}
          >
            {t("btn.browse")}
          </button>
        </div>
      </Field>
      <Field label={t("form.filename_template")}>
        <input
          type="text"
          value={cfg.outputTemplate}
          onChange={(e) => updateConfig({ outputTemplate: e.target.value })}
        />
      </Field>

      <p className="hint">{t("format.quality_hint")}</p>

      <Field label={t("form.quality_preset")}>
        <select
          key={qualityPresetKey}
          defaultValue=""
          onChange={(e) => {
            const val = e.target.value;
            if (val) updateConfig({ quality: val });
            setQualityPresetKey((k) => k + 1);
          }}
        >
          <option value="">—</option>
          {qualityPresets.map((q) => (
            <option key={q.value} value={q.value}>
              {t(q.key)}
            </option>
          ))}
        </select>
      </Field>
      <Field label={t("form.quality")}>
        <input
          type="text"
          placeholder={t("placeholder.quality")}
          value={cfg.quality}
          onChange={(e) => updateConfig({ quality: e.target.value })}
        />
      </Field>
      <Field label={t("form.container")}>
        <select
          value={containerOptions.includes(cfg.format) ? cfg.format : ""}
          onChange={(e) => {
            const val = e.target.value;
            if (val) updateConfig({ format: val });
          }}
        >
          <option value="">—</option>
          {containerOptions.map((f) => (
            <option key={f} value={f}>
              {f}
            </option>
          ))}
        </select>
      </Field>
      <Field label={t("form.container_custom")}>
        <input
          type="text"
          placeholder={t("form.container_custom")}
          value={containerOptions.includes(cfg.format) ? "" : cfg.format}
          onChange={(e) => {
            const val = e.target.value.trim();
            updateConfig({ format: val });
          }}
        />
      </Field>
      <Check
        label={t("form.audio_only")}
        checked={cfg.audioOnly}
        onChange={(v) => updateConfig({ audioOnly: v })}
      />
      {cfg.audioOnly && (
        <Field label={t("form.audio_format")}>
          <select
            value={cfg.audioFormat}
            onChange={(e) => updateConfig({ audioFormat: e.target.value })}
          >
            {AUDIO_FORMATS.map((f) => (
              <option key={f || "default"} value={f}>
                {f || "—"}
              </option>
            ))}
          </select>
        </Field>
      )}

      <div>
        <div className="hint hint-spaced">{t("form.quick_presets")}</div>
        <div className="preset-row">
          {presets.map((p) => (
            <button
              key={p}
              type="button"
              className="preset-chip"
              onClick={async () => {
                const next = await AppAPI.ApplyPreset(p);
                updateConfig(next);
              }}
            >
              {presetLabels[p] ?? p}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
