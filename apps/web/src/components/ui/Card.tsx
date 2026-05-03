import { HTMLAttributes } from "react";
import { classNames } from "../../lib/format";

export function Card({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={classNames("rounded-3xl border border-white/60 bg-white/90 p-5 shadow-panel backdrop-blur", className)} {...props} />;
}
