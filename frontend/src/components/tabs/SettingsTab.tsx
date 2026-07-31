import type { app } from "../../wailsjs/go/models";
import type { LocaleMap } from "../../lib/eventTypes";
import { tf } from "../../lib/i18nFmt";
import { Check, Field } from "../FormControls";

const UI_SCALES = [
  { value: 0, key: "ui_scale.auto" },
  { value: 0.95, key: "ui_scale.compact" },
  { value: 1.05, key: "ui_scale.comfortable" },
  { value: 1.15, key: "ui_scale.large" },
  { value: 1.25, key: "ui_scale.extra_large" },
] as const;

export function SettingsTab({
  settings,
  version,
  locales,
  t,
  updateSettings,
  onLanguageChange,
  onResetWindowSize,
  onExportSettings,
  onImportSettings,
  onShowInstances,
}: {
  settings: app.AppSettingsDTO;
  version: string;
  locales: LocaleMap;
  t: (id: string) => string;
  updateSettings: (patch: Partial<app.AppSettingsDTO>) => void;
  onLanguageChange: (lang: string) => Promise<void>;
  onResetWindowSize: () => void;
  onExportSettings: () => void;
  onImportSettings: () => void;
  onShowInstances: () => void;
}) {
  return (
    <div className="form-grid">
      <div className="hint">{t("card.settings")}</div>
      <Field label={t("form.language")}>
        <select
          value={settings.language || "en"}
          onChange={(e) => void onLanguageChange(e.target.value)}
        >
          <option value="en">{t("lang.en")}</option>
          <option value="ru">{t("lang.ru")}</option>
          <option value="es">{t("lang.es")}</option>
          <option value="pt">{t("lang.pt")}</option>
          <option value="ja">{t("lang.ja")}</option>
          <option value="de">{t("lang.de")}</option>
          <option value="fr">{t("lang.fr")}</option>
          <option value="pl">{t("lang.pl")}</option>
          <option value="ko">{t("lang.ko")}</option>
          <option value="zh-Hant">{t("lang.zh-Hant")}</option>
          <option value="zh-Hans">{t("lang.zh-Hans")}</option>
        </select>
      </Field>
      <Field label={t("form.theme")}>
        <select
          value={settings.theme}
          onChange={(e) => updateSettings({ theme: e.target.value as app.AppSettingsDTO["theme"] })}
        >
          <option value="auto">{t("theme.auto")}</option>
          <option value="dark">{t("theme.dark")}</option>
          <option value="light">{t("theme.light")}</option>
        </select>
      </Field>
      <Field label={t("form.ui_scale")}>
        <select
          value={String(settings.uiScale || 0)}
          onChange={(e) => updateSettings({ uiScale: parseFloat(e.target.value) || 0 })}
        >
          {UI_SCALES.map((s) => (
            <option key={s.key} value={String(s.value)}>
              {t(s.key)}
            </option>
          ))}
        </select>
      </Field>
      <p className="hint">{t("tools.ui_scale_hint")}</p>
      <Field label={t("form.queue_workers")}>
        <input
          type="number"
          min={1}
          max={32}
          value={settings.queueParallel}
          onChange={(e) => updateSettings({ queueParallel: parseInt(e.target.value, 10) || 1 })}
        />
      </Field>
      <Check
        label={t("check.debug_log")}
        checked={settings.debugLog}
        onChange={(v) => updateSettings({ debugLog: v })}
      />
      <Check
        label={t("activity.title")}
        checked={settings.activityPanelOpen}
        onChange={(v) => updateSettings({ activityPanelOpen: v })}
      />
      <div className="btn-row">
        <button type="button" className="btn btn-sm" onClick={onResetWindowSize}>
          {t("btn.reset_window_size")}
        </button>
      </div>
      <p className="hint">{t("tray.hint")}</p>
      <div className="btn-row">
        <button type="button" className="btn btn-sm" onClick={onShowInstances}>
          {t("instances.title")}
        </button>
      </div>
      <div className="section-gap hint">{t("card.backup")}</div>
      <div className="btn-row">
        <button type="button" className="btn btn-sm" onClick={onExportSettings}>
          {t("btn.export_settings")}
        </button>
        <button type="button" className="btn btn-sm" onClick={onImportSettings}>
          {t("btn.import_settings")}
        </button>
      </div>
      <div className="hint">{tf(locales, "tools.app_version", { Version: version })}</div>
    </div>
  );
}
