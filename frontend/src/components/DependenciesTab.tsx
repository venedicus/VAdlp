import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Modal } from "./Modal";
import { tf } from "../lib/i18nFmt";
import { depSettingsPath, mergeSettingsPatch, patchDepPath } from "../lib/depPaths";
import { AppAPI, eventsOn } from "../wailsjs/runtime";
import type { AppSettingsDTO, DependencyDTO, LocaleMap } from "../types";

type DepDialog = { id: string; mode: "install" | "update" } | null;

type GuardNotice = { id: string; message: string } | null;

function depKey(id: string): "ytdlp" | "ffmpeg" | "deno" {
  if (id === "ytdlp") return "ytdlp";
  if (id === "ffmpeg") return "ffmpeg";
  return "deno";
}

export function countDepAttention(deps: DependencyDTO[]): number {
  let n = 0;
  for (const d of deps) {
    if (d.status === "missing" && (d.level === "required" || d.level === "recommended")) n++;
    else if (d.status === "outdated" || d.updateAvail) n++;
    else if (d.status === "error") n++;
    else if (d.status === "unknown") n++;
  }
  return n;
}

function markChecking(deps: DependencyDTO[], ids?: string[]): DependencyDTO[] {
  const all = !ids || ids.length === 0;
  const set = all ? null : new Set(ids);
  return deps.map((d) =>
    all || set!.has(d.id) ? { ...d, status: "checking" as const } : d,
  );
}

type Props = {
  active: boolean;
  deps: DependencyDTO[];
  setDeps: (deps: DependencyDTO[]) => void;
  settings: AppSettingsDTO;
  updateSettings: (patch: Partial<AppSettingsDTO>) => void;
  locales: LocaleMap;
  toolsDir: string;
  onHealthRefresh: () => Promise<void>;
  t: (id: string) => string;
  onError?: (message: string) => void;
};

export function DependenciesTab({
  active,
  deps,
  setDeps,
  settings,
  updateSettings,
  locales,
  toolsDir,
  onHealthRefresh,
  t,
  onError,
}: Props) {
  const [depChecking, setDepChecking] = useState(false);
  const [depInstalling, setDepInstalling] = useState<string | null>(null);
  const [depProgress, setDepProgress] = useState(0);
  const [recheckingId, setRecheckingId] = useState<string | null>(null);
  const [depDialog, setDepDialog] = useState<DepDialog>(null);
  const [guardNotice, setGuardNotice] = useState<GuardNotice>(null);
  const pathDebounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const settingsRef = useRef(settings);
  settingsRef.current = settings;
  const depsRef = useRef(deps);
  depsRef.current = deps;

  const persistDepPath = useCallback(
    async (id: string, path: string) => {
      const s = settingsRef.current;
      const patch = patchDepPath(s, id, path);
      const next = mergeSettingsPatch(s, patch);
      updateSettings(patch);
      await AppAPI.SaveSettings(next);
      return next;
    },
    [updateSettings],
  );

  const refreshDepsLocal = useCallback(async () => {
    const d = await AppAPI.ResolveDependenciesLocal();
    setDeps(Array.isArray(d) ? d : []);
  }, [setDeps]);

  const refreshDepsFull = useCallback(async () => {
    setDepChecking(true);
    setDeps(markChecking(depsRef.current));
    try {
      const d = await AppAPI.CheckDependencies();
      setDeps(Array.isArray(d) ? d : []);
      await onHealthRefresh();
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      onError?.(msg);
      try {
        await refreshDepsLocal();
      } catch {
        /* keep checking state cleared in finally */
      }
    } finally {
      setDepChecking(false);
    }
  }, [setDeps, onHealthRefresh, refreshDepsLocal, onError]);

  useEffect(() => {
    if (!active) return;
    refreshDepsFull().catch(console.error);
  }, [active, refreshDepsFull]);

  useEffect(() => {
    const off = eventsOn("install:progress", (d) => {
      const data = d as { id: string; pct: number };
      setDepInstalling(data.id);
      setDepProgress(data.pct);
    });
    return () => off();
  }, []);

  const adoptPath = useCallback(
    async (id: string, path: string) => {
      await persistDepPath(id, path);
      await refreshDepsFull();
    },
    [persistDepPath, refreshDepsFull],
  );

  const setDepPath = useCallback(
    (id: string, path: string) => {
      const patch = patchDepPath(settingsRef.current, id, path);
      updateSettings(patch);
      if (pathDebounceRef.current) clearTimeout(pathDebounceRef.current);
      pathDebounceRef.current = setTimeout(() => {
        const next = mergeSettingsPatch(settingsRef.current, patch);
        AppAPI.SaveSettings(next)
          .then(() => refreshDepsLocal())
          .catch(console.error);
      }, 500);
    },
    [updateSettings, refreshDepsLocal],
  );

  const runInstallOrUpdate = useCallback(
    async (id: string, mode: "install" | "update") => {
      setDepInstalling(id);
      setDepProgress(0);
      try {
        const path =
          mode === "install"
            ? await AppAPI.InstallDependency(id)
            : await AppAPI.UpdateDependency(id);
        if (path) {
          await persistDepPath(id, path);
        }
        await refreshDepsFull();
      } catch (e) {
        const msg = e instanceof Error ? e.message : String(e);
        onError?.(msg);
        throw e;
      } finally {
        setDepInstalling(null);
        setDepProgress(0);
      }
    },
    [persistDepPath, refreshDepsFull],
  );

  const openInstallDialog = useCallback(
    async (id: string, mode: "install" | "update") => {
      if (mode === "install") {
        const guard = await AppAPI.CheckInstallGuard(id);
        if (guard.alreadyInstalled && guard.path) {
          setGuardNotice({
            id,
            message: tf(locales, "install.already", { Path: guard.path }),
          });
          await adoptPath(id, guard.path);
          return;
        }
        if (guard.foundElsewhere && guard.path) {
          setGuardNotice({
            id,
            message: tf(locales, "install.found_elsewhere", { Path: guard.path }),
          });
          await adoptPath(id, guard.path);
          return;
        }
      }
      setDepDialog({ id, mode });
    },
    [locales, adoptPath],
  );

  const runDepAction = useCallback(async () => {
    if (!depDialog) return;
    const { id, mode } = depDialog;
    setDepDialog(null);
    try {
      await runInstallOrUpdate(id, mode);
    } catch {
      /* journal entry from backend */
    }
  }, [depDialog, runInstallOrUpdate]);

  const recheckOne = useCallback(
    async (id: string) => {
      setRecheckingId(id);
      setDeps(markChecking(depsRef.current, [id]));
      try {
        await refreshDepsLocal();
      } finally {
        setRecheckingId(null);
      }
    },
    [setDeps, refreshDepsLocal],
  );

  const primaryAction = (dep: DependencyDTO) => {
    if (dep.status === "missing" || (dep.status === "error" && !dep.path)) {
      return { label: t(`btn.install_${depKey(dep.id)}`), mode: "install" as const, disabled: false };
    }
    if (dep.status === "outdated" || dep.updateAvail || (dep.status === "error" && dep.path)) {
      return { label: t(`btn.update_${depKey(dep.id)}`), mode: "update" as const, disabled: false };
    }
    if (dep.source === "system" && (dep.status === "found" || dep.status === "unknown")) {
      return { label: t("btn.install_managed"), mode: "install" as const, disabled: false };
    }
    if (dep.status === "unknown") {
      return { label: t("dep.status.unknown"), mode: "install" as const, disabled: true };
    }
    return { label: t("btn.up_to_date"), mode: "install" as const, disabled: true };
  };

  const depCards = useMemo(
    () =>
      deps.map((dep) => {
        const details: string[] = [];
        if (dep.version) details.push(tf(locales, "dep.detail.version", { Version: dep.version }));
        if (dep.latestVer && dep.updateAvail) {
          details.push(tf(locales, "dep.detail.latest", { Latest: dep.latestVer }));
        }
        if (dep.source) {
          details.push(
            tf(locales, "dep.detail.source", { Source: t(`dep.source.${dep.source}`) }),
          );
        }
        if (dep.path) details.push(tf(locales, "dep.detail.path", { Path: dep.path }));
        return { dep, details };
      }),
    [deps, locales, t],
  );

  const cardBusy = (id: string) =>
    depChecking || depInstalling === id || recheckingId === id;

  return (
    <div className="dep-grid">
      <p className="hint">{t("dep.tab_hint")}</p>
      <p className="hint">Tools dir: {toolsDir}</p>
      <p className="dep-warning">{t("dep.outdated_warning")}</p>
      <div className="btn-row">
        <button
          type="button"
          className="btn btn-sm"
          disabled={depChecking || !!depInstalling}
          onClick={() => refreshDepsFull()}
        >
          {depChecking ? t("btn.working") : t("btn.recheck_all")}
        </button>
      </div>

      {depCards.map(({ dep, details }) => {
        const action = primaryAction(dep);
        const customPath = depSettingsPath(settings, dep.id);
        const busy = cardBusy(dep.id);
        const displayStatus = dep.status === "checking" ? "checking" : dep.status;
        return (
          <div key={dep.id} className="dep-card">
            <div className="dep-card-header">
              <span className="dep-card-title">{t(`dep.${depKey(dep.id)}.title`)}</span>
              <span className={`dep-level dep-level-${dep.level}`}>{t(`dep.level.${dep.level}`)}</span>
              <span className={`dep-badge ${displayStatus}`}>{t(`dep.status.${displayStatus}`)}</span>
            </div>
            <p className="dep-desc">{t(`dep.${depKey(dep.id)}.desc`)}</p>
            {details.length > 0 && (
              <div className="dep-meta">
                {details.map((line) => (
                  <div key={line}>{line}</div>
                ))}
              </div>
            )}
            {dep.error && <div className="dep-error">{dep.error}</div>}
            {dep.source === "system" && (
              <p className="dep-source-hint">{t("dep.system_source_hint")}</p>
            )}
            <div className="dep-path-row">
              <input
                type="text"
                className="dep-path-input"
                placeholder={t("dep.path_placeholder")}
                value={customPath}
                readOnly
              />
              <button
                type="button"
                className="btn btn-sm"
                disabled={busy}
                onClick={async () => {
                  const p = await AppAPI.PickFile();
                  if (p) setDepPath(dep.id, p);
                }}
              >
                {t("btn.browse")}
              </button>
              <button
                type="button"
                className="btn btn-sm"
                disabled={busy || !customPath}
                onClick={() => setDepPath(dep.id, "")}
              >
                {t("btn.clear_path")}
              </button>
              <button
                type="button"
                className="btn btn-sm"
                disabled={busy}
                onClick={() => recheckOne(dep.id)}
              >
                {recheckingId === dep.id ? t("btn.working") : t("btn.recheck")}
              </button>
            </div>
            {depInstalling === dep.id && (
              <div className="progress-track dep-progress">
                <div className="progress-fill" style={{ width: `${depProgress}%` }} />
              </div>
            )}
            <div className="btn-row">
              <button
                type="button"
                className={`btn btn-sm${!action.disabled && action.mode === "install" ? " btn-primary" : ""}${!action.disabled && action.mode === "update" ? " btn-primary" : ""}`}
                disabled={action.disabled || !!depInstalling || depChecking}
                onClick={() => openInstallDialog(dep.id, action.mode)}
              >
                {depInstalling === dep.id ? t("btn.working") : action.label}
              </button>
            </div>
          </div>
        );
      })}

      {guardNotice && (
        <Modal
          title={t(`install.${depKey(guardNotice.id)}.title`)}
          onClose={() => setGuardNotice(null)}
        >
          <p>{guardNotice.message}</p>
          <div className="btn-row">
            <button type="button" className="btn" onClick={() => setGuardNotice(null)}>
              {t("btn.close")}
            </button>
          </div>
        </Modal>
      )}

      {depDialog && (
        <Modal
          title={t(
            depDialog.mode === "update"
              ? `dialog.${depKey(depDialog.id)}_update`
              : `install.${depKey(depDialog.id)}.title`,
          )}
          onClose={() => !depInstalling && setDepDialog(null)}
        >
          {depDialog.mode === "install" ? (
            <p>{t(`install.${depKey(depDialog.id)}.body`)}</p>
          ) : depDialog.id === "ytdlp" ? (
            <p>{t("install.ytdlp.update_hint")}</p>
          ) : null}
          {depInstalling === depDialog.id && (
            <div className="progress-track dep-progress">
              <div className="progress-fill" style={{ width: `${depProgress}%` }} />
            </div>
          )}
          <div className="btn-row">
            <button
              type="button"
              className="btn btn-primary"
              disabled={!!depInstalling}
              onClick={() => runDepAction()}
            >
              {depInstalling === depDialog.id
                ? t("btn.working")
                : t(
                    depDialog.mode === "update"
                      ? `btn.update_${depKey(depDialog.id)}`
                      : "btn.install",
                  )}
            </button>
            <button
              type="button"
              className="btn"
              disabled={!!depInstalling}
              onClick={() => setDepDialog(null)}
            >
              {t("btn.close")}
            </button>
          </div>
        </Modal>
      )}
    </div>
  );
}

export async function installDependencyWithSave(
  id: string,
  settings: AppSettingsDTO,
): Promise<{ path: string; settings: AppSettingsDTO }> {
  const path = await AppAPI.InstallDependency(id);
  if (!path) {
    return { path: "", settings };
  }
  const next = mergeSettingsPatch(settings, patchDepPath(settings, id, path));
  await AppAPI.SaveSettings(next);
  return { path, settings: next };
}
