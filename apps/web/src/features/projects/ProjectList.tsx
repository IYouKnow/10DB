import { Link } from "react-router-dom";
import { useDeleteProject, useProjects } from "../../api/projects";
import { Card } from "../../components/ui/Card";
import { formatDateTime } from "../../lib/format";

export function ProjectList() {
  const projects = useProjects();
  const deleteProject = useDeleteProject();
  const items = projects.data?.projects ?? [];

  if (projects.isLoading) {
    return <Card>Loading projects...</Card>;
  }

  return (
    <div className="grid gap-4">
      {items.map((project) => (
        <Card key={project.id} className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div>
            <Link className="text-lg font-bold text-ink hover:text-ember" to={`/projects/${project.id}`}>
              {project.name}
            </Link>
            <p className="text-sm text-slate-500">{project.slug}</p>
            <p className="mt-2 text-sm text-slate-600">{project.description || "No description yet."}</p>
          </div>
          <div className="flex items-center gap-4">
            <div className="text-right text-xs text-slate-500">
              <div className="font-semibold uppercase tracking-[0.16em]">{project.status}</div>
              <div>{formatDateTime(project.createdAt)}</div>
            </div>
            <button
              className="rounded-xl border border-rose-200 px-3 py-2 text-sm font-semibold text-rose-600 hover:bg-rose-50"
              onClick={() => {
                if (window.confirm(`Delete project ${project.name}? This removes its database and role.`)) {
                  deleteProject.mutate(project.id);
                }
              }}
            >
              Delete
            </button>
          </div>
        </Card>
      ))}
      {items.length === 0 ? (
        <Card>
          <p className="text-slate-600">No projects yet. Create your first isolated PostgreSQL app database above.</p>
        </Card>
      ) : null}
    </div>
  );
}
