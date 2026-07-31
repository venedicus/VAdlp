import { useState } from "react";
import type { app } from "../wailsjs/go/models";
import { Modal } from "./Modal";
import { Check, Field } from "./FormControls";

export function EditQueueTaskModal({
  task,
  t,
  onClose,
  onSave,
}: {
  task: app.QueueTaskDTO;
  t: (id: string) => string;
  onClose: () => void;
  onSave: (cfg: app.ConfigDTO) => Promise<void>;
}) {
  const [cfg, setCfg] = useState<app.ConfigDTO>(task.config);
  const [saving, setSaving] = useState(false);

  const patch = (p: Partial<app.ConfigDTO>) => setCfg((prev) => ({ ...prev, ...p }));

  return (
    <Modal title={t("dialog.edit_task")} onClose={onClose}>
      <Field label={t("form.rate_limit")}>
        <input
          type="text"
          placeholder={t("placeholder.rate")}
          value={cfg.rateLimit}
          onChange={(e) => patch({ rateLimit: e.target.value })}
        />
      </Field>
      <Field label={t("form.quality")}>
        <input
          type="text"
          placeholder={t("placeholder.quality")}
          value={cfg.quality}
          onChange={(e) => patch({ quality: e.target.value })}
        />
      </Field>
      <Check label={t("form.audio_only")} checked={cfg.audioOnly} onChange={(v) => patch({ audioOnly: v })} />
      <div className="btn-row">
        <button
          type="button"
          className="btn btn-primary"
          disabled={saving}
          onClick={async () => {
            setSaving(true);
            try {
              await onSave(cfg);
            } finally {
              setSaving(false);
            }
          }}
        >
          {saving ? t("btn.working") : t("btn.save")}
        </button>
        <button type="button" className="btn" disabled={saving} onClick={onClose}>
          {t("btn.cancel")}
        </button>
      </div>
    </Modal>
  );
}
