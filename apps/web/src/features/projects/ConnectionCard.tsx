import { useProjectConnection, useResetProject } from "../../api/projects";
import { Button } from "../../components/ui/Button";
import { Card } from "../../components/ui/Card";

export function ConnectionCard({ projectId }: { projectId: string }) {
  const connection = useProjectConnection(projectId);
  const resetProject = useResetProject();

  if (connection.isLoading) {
    return <Card>Loading connection details...</Card>;
  }

  return (
    <Card className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-bold text-ink">Connection details</h2>
          <p className="text-sm text-slate-600">Keep admin credentials server-side. Use the project-specific DSN below.</p>
        </div>
        <Button
          className="bg-slate-200 text-ink hover:bg-slate-300"
          onClick={() => {
            if (window.confirm("Reset the public schema for this project database?")) {
              resetProject.mutate(projectId);
            }
          }}
        >
          Reset schema
        </Button>
      </div>
      <div className="rounded-2xl bg-panel p-4 text-sm text-white">
        <pre className="overflow-x-auto whitespace-pre-wrap">{connection.data?.dsn}</pre>
      </div>
      <div>
        <p className="mb-2 text-sm font-semibold text-slate-700">`.env` example</p>
        <div className="rounded-2xl border border-slate-200 bg-white p-4 text-sm text-slate-700">
          <pre className="overflow-x-auto whitespace-pre-wrap">{connection.data?.envExample}</pre>
        </div>
      </div>
    </Card>
  );
}
