import type { ConfigDTO } from "../../types";
import { Check, Field } from "../FormControls";

export function ExtrasTab({
  cfg,
  t,
  updateConfig,
}: {
  cfg: ConfigDTO;
  t: (id: string) => string;
  updateConfig: (patch: Partial<ConfigDTO>) => void;
}) {
  return (
    <div className="form-grid">
      <div className="hint">{t("card.media_extras")}</div>
      <Check label={t("check.write_subs")} checked={cfg.writeSubs} onChange={(v) => updateConfig({ writeSubs: v })} />
      <Check label={t("check.write_auto_sub")} checked={cfg.writeAutoSub} onChange={(v) => updateConfig({ writeAutoSub: v })} />
      <Check label={t("check.embed_subs")} checked={cfg.embedSubs} onChange={(v) => updateConfig({ embedSubs: v })} />
      <Field label={t("form.sub_langs")}>
        <input type="text" value={cfg.subLangs} onChange={(e) => updateConfig({ subLangs: e.target.value })} />
      </Field>
      <Check label={t("check.write_thumb")} checked={cfg.writeThumbnail} onChange={(v) => updateConfig({ writeThumbnail: v })} />
      <Check label={t("check.embed_thumb")} checked={cfg.embedThumbnail} onChange={(v) => updateConfig({ embedThumbnail: v })} />
      <Check label={t("check.embed_meta")} checked={cfg.embedMetadata} onChange={(v) => updateConfig({ embedMetadata: v })} />
      <Check label={t("check.embed_chapters")} checked={cfg.embedChapters} onChange={(v) => updateConfig({ embedChapters: v })} />
      <Check label={t("check.write_info_json")} checked={cfg.writeInfoJSON} onChange={(v) => updateConfig({ writeInfoJSON: v })} />
      <Field label={t("form.load_info_json")}>
        <input type="text" placeholder={t("placeholder.load_info_json")} value={cfg.loadInfoJson} onChange={(e) => updateConfig({ loadInfoJson: e.target.value })} />
      </Field>

      <div className="section-gap hint">{t("card.retries")}</div>
      <Field label={t("form.retries")}>
        <input type="number" min={0} value={cfg.retries} onChange={(e) => updateConfig({ retries: parseInt(e.target.value, 10) || 0 })} />
      </Field>
      <Field label={t("form.frag_retries")}>
        <input type="number" min={0} value={cfg.fragmentRetries} onChange={(e) => updateConfig({ fragmentRetries: parseInt(e.target.value, 10) || 0 })} />
      </Field>
      <Field label={t("form.concurrent_frags")}>
        <input type="number" min={1} value={cfg.concurrentFragments} onChange={(e) => updateConfig({ concurrentFragments: parseInt(e.target.value, 10) || 1 })} />
      </Field>
      <Field label={t("form.socket_timeout")}>
        <input type="number" min={0} placeholder={t("placeholder.socket_timeout")} value={cfg.socketTimeout || ""} onChange={(e) => updateConfig({ socketTimeout: parseInt(e.target.value, 10) || 0 })} />
      </Field>
      <Check label={t("check.no_warnings")} checked={cfg.noWarnings} onChange={(v) => updateConfig({ noWarnings: v })} />
      <Check label={t("check.verbose")} checked={cfg.verbose} onChange={(v) => updateConfig({ verbose: v, quiet: v ? false : cfg.quiet })} />
      <Check label={t("check.quiet")} checked={cfg.quiet} onChange={(v) => updateConfig({ quiet: v, verbose: v ? false : cfg.verbose })} />
      <Check label={t("check.windows_filenames")} checked={cfg.windowsFilenames} onChange={(v) => updateConfig({ windowsFilenames: v })} />
      <Check label={t("check.no_mtime")} checked={cfg.noMtime} onChange={(v) => updateConfig({ noMtime: v })} />
      <Check label={t("check.abort_on_error")} checked={cfg.abortOnError} onChange={(v) => updateConfig({ abortOnError: v })} />
      <Check label={t("check.ignore_errors")} checked={cfg.ignoreErrors} onChange={(v) => updateConfig({ ignoreErrors: v })} />
      <Check label={t("check.sponsorblock")} checked={cfg.sponsorBlockRemove} onChange={(v) => updateConfig({ sponsorBlockRemove: v })} />
      <Field label={t("card.extra_flags")} wide>
        <textarea placeholder={t("placeholder.extra_args")} value={cfg.extraArgs} onChange={(e) => updateConfig({ extraArgs: e.target.value })} />
      </Field>
    </div>
  );
}
