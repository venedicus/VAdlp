import type { app } from "../../wailsjs/go/models";
import { AppAPI } from "../../wailsjs/runtime";
import { Check, Field } from "../FormControls";

const COOKIES_BROWSERS = ["chrome", "firefox", "vivaldi", "edge", "brave"] as const;

export function NetworkTab({
  cfg,
  t,
  updateConfig,
}: {
  cfg: app.ConfigDTO;
  t: (id: string) => string;
  updateConfig: (patch: Partial<app.ConfigDTO>) => void;
}) {
  return (
    <div className="form-grid">
      <Check label={t("check.cookies_browser")} checked={cfg.useCookiesBrowser} onChange={(v) => updateConfig({ useCookiesBrowser: v })} />
      <Field label={t("form.browser")}>
        <select
          value={cfg.cookiesBrowser || "chrome"}
          disabled={!cfg.useCookiesBrowser}
          onChange={(e) => updateConfig({ cookiesBrowser: e.target.value })}
        >
          {COOKIES_BROWSERS.map((b) => (
            <option key={b} value={b}>
              {b}
            </option>
          ))}
          {!COOKIES_BROWSERS.includes(cfg.cookiesBrowser as (typeof COOKIES_BROWSERS)[number]) &&
            cfg.cookiesBrowser && <option value={cfg.cookiesBrowser}>{cfg.cookiesBrowser}</option>}
        </select>
      </Field>
      <Check label={t("check.cookies_file")} checked={cfg.useCookiesFile} onChange={(v) => updateConfig({ useCookiesFile: v })} />
      <Field label={t("form.cookies_file")}>
        <div className="input-with-actions">
          <input type="text" placeholder={t("placeholder.cookies")} value={cfg.cookiesFile} onChange={(e) => updateConfig({ cookiesFile: e.target.value })} />
          <button
            type="button"
            className="btn btn-sm"
            onClick={async () => {
              const p = await AppAPI.PickFile();
              if (p) updateConfig({ cookiesFile: p, useCookiesFile: true });
            }}
          >
            {t("btn.browse")}
          </button>
        </div>
      </Field>
      <Field label={t("form.proxy")}>
        <input type="text" placeholder={t("placeholder.proxy")} value={cfg.proxy} onChange={(e) => updateConfig({ proxy: e.target.value })} />
      </Field>
      <Field label={t("form.rate_limit")}>
        <input type="text" placeholder={t("placeholder.rate")} value={cfg.rateLimit} onChange={(e) => updateConfig({ rateLimit: e.target.value })} />
      </Field>
      <Field label={t("form.username")}>
        <input type="text" value={cfg.username} onChange={(e) => updateConfig({ username: e.target.value })} />
      </Field>
      <Field label={t("form.password")}>
        <input type="password" value={cfg.password} onChange={(e) => updateConfig({ password: e.target.value })} />
      </Field>
    </div>
  );
}
