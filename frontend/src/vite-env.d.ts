/// <reference types="vite/client" />

interface Window {
  go?: {
    app?: {
      App?: Record<string, (...args: unknown[]) => unknown>;
    };
  };
  runtime?: Record<string, unknown>;
}
