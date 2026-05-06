import React, { useState } from 'react';
import { Database, Plus, Search } from 'lucide-react';
import { Link, useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import ProjectCard from '../components/dashboard/ProjectCard';
import CreateProjectModal from '../components/dashboard/CreateProjectModal';
import { useProjects } from '@/lib/ProjectsContext';
import { useAuth } from '@/lib/AuthContext';

export default function Dashboard() {
  const [search, setSearch] = useState('');
  const [showCreate, setShowCreate] = useState(false);
  const [isCreatingProject, setIsCreatingProject] = useState(false);
  const [createProjectError, setCreateProjectError] = useState('');
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const {
    projects,
    isLoadingProjects,
    projectsError,
    createProject,
    deleteProject,
  } = useProjects();

  const filtered = projects.filter(p =>
    p.name.toLowerCase().includes(search.toLowerCase())
  );

  const handleCreate = async ({ name }) => {
    setCreateProjectError('');
    setIsCreatingProject(true);

    try {
      const project = await createProject({ name });
      setShowCreate(false);
      navigate(`/projects/${project.id}/board`);
    } catch (error) {
      setCreateProjectError(error.message || 'Failed to create project');
    } finally {
      setIsCreatingProject(false);
    }
  };

  const handleDelete = async (id) => {
    await deleteProject(id);
  };

  const handleLogout = async () => {
    await logout();
    navigate('/login', { replace: true });
  };

  return (
    <div className="min-h-screen bg-background text-foreground">
      <header className="border-b border-border bg-card">
        <div className="mx-auto flex h-14 max-w-6xl items-center gap-4 px-6">
          <div className="flex items-center gap-2.5">
            <div className="flex h-7 w-7 items-center justify-center rounded-lg border border-primary/20 bg-primary/10">
              <Database className="h-3.5 w-3.5 text-primary" />
            </div>
            <span className="text-sm font-semibold tracking-tight">
              10DB <span className="text-primary">Launch</span>
            </span>
          </div>

          <div className="flex-1" />

          <div className="hidden text-xs text-muted-foreground sm:block">
            {user?.email}
          </div>

          {user?.role === 'admin' ? (
            <Link
              to="/admin"
              className="hidden rounded-lg border border-border bg-secondary/20 px-3 py-1.5 text-xs text-foreground transition-colors hover:bg-secondary/60 sm:inline-flex"
            >
              Admin
            </Link>
          ) : null}

          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={handleLogout}
            className="h-8"
          >
            Logout
          </Button>

          <div className="hidden text-xs font-mono text-muted-foreground sm:block">
            Create or pick a project to enter the workspace
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-6 py-8">
        {/* Page header */}
        <div className="mb-8 flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
          <div>
            <h1 className="text-xl font-bold text-foreground">Projects</h1>
            <p className="mt-0.5 text-sm text-muted-foreground">{projects.length} databases provisioned</p>
          </div>
          <Button
            onClick={() => setShowCreate(true)}
            className="bg-primary font-medium text-primary-foreground hover:bg-primary/90"
            size="sm"
          >
            <Plus className="mr-1.5 h-4 w-4" />
            Create Project
          </Button>
        </div>

        {/* Search */}
        <div className="relative mb-6 max-w-sm">
          <Search className="absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search projects..."
            value={search}
            onChange={e => setSearch(e.target.value)}
            className="h-8 pl-9 text-sm"
          />
        </div>

        {/* Project grid */}
        {isLoadingProjects ? (
          <div className="py-20 text-center text-sm text-muted-foreground">
            Loading projects...
          </div>
        ) : projectsError ? (
          <div className="rounded-2xl border border-destructive/20 bg-destructive/5 px-4 py-3 text-sm text-destructive">
            {projectsError}
          </div>
        ) : filtered.length === 0 ? (
          <div className="py-20 text-center text-sm text-muted-foreground">
            No projects found.
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {filtered.map(project => (
              <ProjectCard
                key={project.id}
                project={project}
                onDelete={handleDelete}
              />
            ))}
          </div>
        )}

        {showCreate && (
          <CreateProjectModal
            onClose={() => {
              setCreateProjectError('');
              setShowCreate(false);
            }}
            onCreate={handleCreate}
            isCreating={isCreatingProject}
            error={createProjectError}
          />
        )}
      </main>
    </div>
  );
}
