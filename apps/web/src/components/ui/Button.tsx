import { ButtonHTMLAttributes } from "react";
import { classNames } from "../../lib/format";

export function Button({ className, ...props }: ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      className={classNames(
        "rounded-xl bg-ink px-4 py-2 text-sm font-semibold text-white transition hover:translate-y-[-1px] hover:bg-panel disabled:cursor-not-allowed disabled:opacity-60",
        className
      )}
      {...props}
    />
  );
}
