import { Component, type ErrorInfo, type ReactNode } from "react";

type Props = { children: ReactNode };
type State = { error: Error | null };

export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null };

  static getDerivedStateFromError(error: Error): State {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error("VAdlp UI error:", error, info.componentStack);
  }

  render() {
    if (this.state.error) {
      return (
        <div className="app-shell error-boundary">
          <div className="bubble-box" data-title="Error">
            <div className="bubble-box-inner">
              <p className="error-boundary-message">The interface failed to render.</p>
              <pre className="error-boundary-stack">{this.state.error.message}</pre>
              <button type="button" className="btn btn-primary" onClick={() => window.location.reload()}>
                Reload
              </button>
            </div>
          </div>
        </div>
      );
    }
    return this.props.children;
  }
}
