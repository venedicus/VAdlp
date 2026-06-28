import { useEffect, useRef, useState } from "react";

type ModalProps = {
  title: string;
  onClose: () => void;
  children: React.ReactNode;
  wide?: boolean;
};

export function Modal({ title, onClose, children, wide }: ModalProps) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div
        className={`modal${wide ? " modal-wide" : ""}`}
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
      >
        <h2>{title}</h2>
        {children}
      </div>
    </div>
  );
}

type ConfirmModalProps = {
  title: string;
  message: string;
  confirmLabel: string;
  cancelLabel: string;
  danger?: boolean;
  onConfirm: () => void;
  onClose: () => void;
};

export function ConfirmModal({
  title,
  message,
  confirmLabel,
  cancelLabel,
  danger,
  onConfirm,
  onClose,
}: ConfirmModalProps) {
  return (
    <Modal title={title} onClose={onClose}>
      <p>{message}</p>
      <div className="btn-row">
        <button
          type="button"
          className={`btn${danger ? " btn-danger" : " btn-primary"}`}
          onClick={() => {
            onConfirm();
            onClose();
          }}
        >
          {confirmLabel}
        </button>
        <button type="button" className="btn" onClick={onClose}>
          {cancelLabel}
        </button>
      </div>
    </Modal>
  );
}

type PromptModalProps = {
  title: string;
  label: string;
  defaultValue?: string;
  descriptionLabel?: string;
  defaultDescription?: string;
  submitLabel: string;
  cancelLabel: string;
  onSubmit: (value: string, description?: string) => void;
  onClose: () => void;
};

export function PromptModal({
  title,
  label,
  defaultValue = "",
  descriptionLabel,
  defaultDescription = "",
  submitLabel,
  cancelLabel,
  onSubmit,
  onClose,
}: PromptModalProps) {
  const [value, setValue] = useState(defaultValue);
  const [description, setDescription] = useState(defaultDescription);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    inputRef.current?.focus();
    inputRef.current?.select();
  }, []);

  const submit = () => {
    if (!value.trim()) return;
    onSubmit(value.trim(), descriptionLabel ? description.trim() : undefined);
    onClose();
  };

  return (
    <Modal title={title} onClose={onClose}>
      <Field label={label}>
        <input
          ref={inputRef}
          type="text"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter" && value.trim()) submit();
          }}
        />
      </Field>
      {descriptionLabel && (
        <Field label={descriptionLabel}>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={2}
          />
        </Field>
      )}
      <div className="btn-row">
        <button type="button" className="btn btn-primary" disabled={!value.trim()} onClick={submit}>
          {submitLabel}
        </button>
        <button type="button" className="btn" onClick={onClose}>
          {cancelLabel}
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
