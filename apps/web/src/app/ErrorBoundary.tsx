import { Component, ErrorInfo, ReactNode } from "react";

interface State {
  hasError: boolean;
  errorMessage: string;
}

export class ErrorBoundary extends Component<{ children: ReactNode }, State> {
  state: State = {
    hasError: false,
    errorMessage: ""
  };

  static getDerivedStateFromError(error: Error): State {
    return {
      hasError: true,
      errorMessage: error.message
    };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error("10DB Launch frontend crashed", error, info);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex min-h-screen items-center justify-center bg-slate-100 p-6">
          <div className="max-w-2xl rounded-3xl border border-rose-200 bg-white p-8 shadow-xl">
            <p className="text-sm font-semibold uppercase tracking-[0.18em] text-rose-600">Frontend error</p>
            <h1 className="mt-2 text-2xl font-black text-slate-900">The dashboard crashed while rendering.</h1>
            <p className="mt-3 text-slate-600">This is now being surfaced instead of showing a blank page.</p>
            <pre className="mt-4 overflow-x-auto rounded-2xl bg-slate-950 p-4 text-sm text-white">
              {this.state.errorMessage || "Unknown error"}
            </pre>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
