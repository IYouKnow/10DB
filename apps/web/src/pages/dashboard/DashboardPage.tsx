import { AppShell } from "../../components/layout/AppShell";
import { ProjectCreateForm } from "../../features/projects/ProjectCreateForm";
import { ProjectList } from "../../features/projects/ProjectList";

export function DashboardPage() {
  return (
    <AppShell title="Projects">
      <div className="space-y-6">
        <section>
          <h1 className="text-3xl font-black text-ink">Launch and design PostgreSQL projects visually</h1>
          <p className="mt-2 max-w-3xl text-slate-600">
            One dashboard for provisioning isolated databases, shaping tables visually, previewing SQL, and copying connection strings without touching the CLI.
          </p>
        </section>
        <ProjectCreateForm />
        <ProjectList />
      </div>
    </AppShell>
  );
}
