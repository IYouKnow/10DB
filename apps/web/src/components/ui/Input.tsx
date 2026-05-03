import { InputHTMLAttributes } from "react";
import { classNames } from "../../lib/format";

export function Input({ className, ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      className={classNames(
        "w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 shadow-sm outline-none transition focus:border-glow focus:ring-2 focus:ring-glow/30",
        className
      )}
      {...props}
    />
  );
}
