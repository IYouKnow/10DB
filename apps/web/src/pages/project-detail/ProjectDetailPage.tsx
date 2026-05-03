import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { useProject } from "../../api/projects";
import { useSaveSchema, useSchema, useValidateSchema } from "../../api/schemas";
import { useApplySchema, useSqlPreview } from "../../api/sql";
import { AppShell } from "../../components/layout/AppShell";
import { DataBrowser } from "../../features/data-browser/DataBrowser";
import { ConnectionCard } from "../../features/projects/ConnectionCard";
import { SchemaBuilder } from "../../features/schema-builder/SchemaBuilder";
import { SqlPreviewCard } from "../../features/sql-preview/SqlPreviewCard";
import { DEFAULT_BLUEPRINT } from "../../lib/constants";
import { SchemaBlueprint } from "../../lib/types";

export function ProjectDetailPage() {
  const { projectId = "" } = useParams();
  const project = useProject(projectId);
  const schema = useSchema(projectId);
  const saveSchema = useSaveSchema(projectId);
  const validateSchema = useValidateSchema(projectId);
  const previewSql = useSqlPreview(projectId);
  const applySchema = useApplySchema(projectId);
  const [blueprint, setBlueprint] = useState<SchemaBlueprint>(DEFAULT_BLUEPRINT(projectId));
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({});
  const [sql, setSql] = useState("");

  useEffect(() => {
    if (schema.data) {
      if ("tables" in schema.data) {
        setBlueprint(schema.data as SchemaBlueprint);
      } else {
        setBlueprint((schema.data as { blueprint: SchemaBlueprint }).blueprint);
      }
    }
  }, [schema.data]);

  if (project.isLoading) {
    return <div className="flex min-h-screen items-center justify-center">Loading project...</div>;
  }

  if (project.isError || !project.data) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-center">
          <p className="text-slate-700">Project not found.</p>
          <Link className="text-ember" to="/">
            Back to dashboard
          </Link>
        </div>
      </div>
    );
  }

  return (
    <AppShell title={project.data.name} wide>
      <div className="space-y-5">
        <section className="rounded-[2rem] border border-white/45 bg-white/65 px-6 py-5 shadow-panel backdrop-blur">
          <div className="flex flex-wrap items-end justify-between gap-4">
            <div>
              <Link to="/" className="text-sm font-semibold text-ember">
                Back to projects
              </Link>
              <h1 className="mt-2 text-3xl font-black tracking-tight text-ink">{project.data.name}</h1>
              <p className="mt-2 max-w-4xl text-slate-600">
                This screen is the schema board. Build the structure visually here, then use the smaller support panels below only when you need SQL, credentials, or table browsing.
              </p>
            </div>
            <div className="flex flex-wrap gap-2">
              <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] text-slate-600">{project.data.slug}</span>
              <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] text-slate-600">{project.data.status}</span>
              <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] text-slate-600">{project.data.pgDatabaseName}</span>
            </div>
          </div>
        </section>

        <SchemaBuilder
          blueprint={blueprint}
          onChange={setBlueprint}
          onSave={() => {
            validateSchema.mutate(blueprint, {
              onSuccess: (result) => {
                setValidationErrors(result.errors);
                if (result.valid) {
                  saveSchema.mutate(blueprint);
                }
              }
            });
          }}
          onPreview={() => {
            previewSql.mutate(blueprint, {
              onSuccess: (result) => setSql(result.sql),
              onError: () => setSql("")
            });
          }}
          validationErrors={validationErrors}
        />

        <div className="grid gap-6 xl:grid-cols-[minmax(0,1.25fr)_minmax(0,0.75fr)]">
          <div className="space-y-6">
            <SqlPreviewCard
              sql={sql}
              isApplying={applySchema.isPending}
              onApply={() =>
                applySchema.mutate(undefined, {
                  onSuccess: () => {
                    void project.refetch();
                  }
                })
              }
            />

            <ConnectionCard projectId={projectId} />
          </div>

          <DataBrowser projectId={projectId} />
        </div>
      </div>
    </AppShell>
  );
}
