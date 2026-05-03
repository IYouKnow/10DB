import { ReactNode } from "react";
import { Link } from "react-router-dom";
import { useLogout } from "../../api/auth";
import { Button } from "../ui/Button";

export function AppShell({ title, children }: { title: string; children: ReactNode }) {
  const logout = useLogout();
  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top_left,_rgba(141,210,200,0.45),_transparent_35%),linear-gradient(180deg,_#eef5ff_0%,_#dbe7fb_100%)]">
      <header className="border-b border-white/60 bg-white/60 backdrop-blur">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-4">
          <Link to="/" className="text-xl font-black tracking-tight text-ink">
            10DB Launch
          </Link>
          <div className="flex items-center gap-3">
            <span className="text-sm text-slate-600">{title}</span>
            <Button className="bg-ember text-ink hover:bg-ember/90" onClick={() => logout.mutate()}>
              Log out
            </Button>
          </div>
        </div>
      </header>
      <main className="mx-auto max-w-7xl px-6 py-8">{children}</main>
    </div>
  );
}
