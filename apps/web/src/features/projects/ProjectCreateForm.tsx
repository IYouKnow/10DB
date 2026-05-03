import { FormEvent, useState } from "react";
import { useCreateProject } from "../../api/projects";
import { Button } from "../../components/ui/Button";
import { Card } from "../../components/ui/Card";
import { Input } from "../../components/ui/Input";

export function ProjectCreateForm() {
  const createProject = useCreateProject();
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [description, setDescription] = useState("");

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    createProject.mutate(
      { name, slug, description },
      {
        onSuccess: () => {
          setName("");
          setSlug("");
          setDescription("");
        }
      }
    );
  }

  return (
    <Card>
      <h2 className="text-lg font-bold text-ink">Create project</h2>
      <p className="mt-1 text-sm text-slate-600">Each project gets its own PostgreSQL database and database user.</p>
      <form className="mt-5 grid gap-3 md:grid-cols-3" onSubmit={handleSubmit}>
        <Input placeholder="Project name" value={name} onChange={(event) => setName(event.target.value)} />
        <Input placeholder="slug" value={slug} onChange={(event) => setSlug(event.target.value.toLowerCase())} />
        <Input placeholder="Short description" value={description} onChange={(event) => setDescription(event.target.value)} />
        <div className="md:col-span-3">
          {createProject.isError ? <p className="mb-3 text-sm text-rose-600">{createProject.error.message}</p> : null}
          <Button disabled={createProject.isPending} type="submit">
            {createProject.isPending ? "Provisioning..." : "Create project"}
          </Button>
        </div>
      </form>
    </Card>
  );
}
