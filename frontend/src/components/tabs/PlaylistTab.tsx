import type { app } from "../../wailsjs/go/models";
import { AppAPI } from "../../wailsjs/runtime";
import { Check, Field } from "../FormControls";

export function PlaylistTab({
  cfg,
  settings,
  t,
  updateConfig,
  updateSettings,
  runSessionAction,
}: {
  cfg: app.ConfigDTO;
  settings: app.AppSettingsDTO;
  t: (id: string) => string;
  updateConfig: (patch: Partial<app.ConfigDTO>) => void;
  updateSettings: (patch: Partial<app.AppSettingsDTO>) => void;
  runSessionAction: (action: "save" | "load" | "resume") => Promise<void>;
}) {
  return (
    <div className="form-grid">
      <Check label={t("check.reverse")} checked={cfg.playlistReverse} onChange={(v) => updateConfig({ playlistReverse: v })} />
      <Check label={t("check.continue")} checked={cfg.continue} onChange={(v) => updateConfig({ continue: v })} />
      <Check label={t("check.no_part")} checked={cfg.noPart} onChange={(v) => updateConfig({ noPart: v })} />
      <Check label={t("check.no_playlist")} checked={cfg.noPlaylist} onChange={(v) => updateConfig({ noPlaylist: v })} />
      <Check label={t("check.flat_playlist")} checked={cfg.flatPlaylist} onChange={(v) => updateConfig({ flatPlaylist: v })} />
      <Field label={t("form.playlist_start")}>
        <input type="number" min={0} value={cfg.playlistStart || ""} onChange={(e) => updateConfig({ playlistStart: parseInt(e.target.value, 10) || 0 })} />
      </Field>
      <Field label={t("form.playlist_end")}>
        <input type="number" min={0} value={cfg.playlistEnd || ""} onChange={(e) => updateConfig({ playlistEnd: parseInt(e.target.value, 10) || 0 })} />
      </Field>
      <Field label={t("form.max_downloads")}>
        <input
          type="number"
          min={0}
          placeholder={t("placeholder.max_downloads")}
          value={cfg.maxDownloads || ""}
          onChange={(e) => updateConfig({ maxDownloads: parseInt(e.target.value, 10) || 0 })}
        />
      </Field>
      <Field label={t("form.archive")}>
        <input type="text" placeholder={t("placeholder.archive")} value={cfg.downloadArchive} onChange={(e) => updateConfig({ downloadArchive: e.target.value })} />
      </Field>
      <Field label={t("form.session_path")}>
        <div className="input-with-actions">
          <input type="text" value={settings.sessionPath} onChange={(e) => updateSettings({ sessionPath: e.target.value })} />
          <button
            type="button"
            className="btn btn-sm"
            onClick={async () => {
              const p = await AppAPI.PickFile();
              if (p) updateSettings({ sessionPath: p });
            }}
          >
            {t("btn.browse")}
          </button>
        </div>
      </Field>
      <div className="btn-row">
        <button type="button" className="btn btn-sm" onClick={() => runSessionAction("save")}>
          {t("btn.save_session")}
        </button>
        <button type="button" className="btn btn-sm" onClick={() => runSessionAction("load")}>
          {t("btn.load_session")}
        </button>
        <button type="button" className="btn btn-sm" onClick={() => runSessionAction("resume")}>
          {t("btn.resume_session")}
        </button>
      </div>
    </div>
  );
}
