import React, { useState } from 'react';
import { Link, Outlet, useLocation, useNavigate, useParams } from 'react-router-dom';
import {
  LayoutGrid, GitBranch, Table2, KeyRound, Code2, Database,
  Settings, ChevronDown, Circle, Menu
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { useProjects } from '@/lib/ProjectsContext';
import { useAuth } from '@/lib/AuthContext';
import { Button } from '@/components/ui/button';

const navItems = [
  { label: 'All Projects', icon: LayoutGrid, href: () => '/' },
  { label: 'Schema Board', icon: GitBranch, href: (projectId) => `/projects/${projectId}/board` },
  { label: 'Tables', icon: Table2, href: (projectId) => `/projects/${projectId}/tables` },
  { label: 'Credentials', icon: KeyRound, href: (projectId) => `/projects/${projectId}/credentials` },
  { label: 'SQL Preview', icon: Code2, href: (projectId) => `/projects/${projectId}/sql` },
  { label: 'Backups', icon: Database, href: (projectId) => `/projects/${projectId}/backups` },
  { label: 'Settings', icon: Settings, href: (projectId) => `/projects/${projectId}/settings` },
];

export default function AppLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const { projectId } = useParams();
  const [projectOpen, setProjectOpen] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const { projects } = useProjects();
  const { user, logout } = useAuth();
  const activeProject = projects.find((project) => project.id === projectId) ?? null;

  const handleLogout = async () => {
    await logout();
    navigate('/login', { replace: true });
  };

  return (
    <div className="flex h-screen bg-background text-foreground overflow-hidden">
      {/* Mobile overlay */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black/60 z-20 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Sidebar */}
      <aside className={cn(
        'fixed lg:static inset-y-0 left-0 z-30 w-56 flex flex-col border-r border-border bg-card transition-transform duration-200',
        sidebarOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'
      )}>
        {/* Logo */}
        <div className="h-14 flex items-center gap-2.5 px-4 border-b border-border shrink-0">
          <div className="w-7 h-7 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center">
            <Database className="w-3.5 h-3.5 text-primary" />
          </div>
          <span className="font-semibold text-sm tracking-tight">
            10DB <span className="text-primary">Launch</span>
          </span>
        </div>

        {/* Nav */}
        <nav className="flex-1 px-3 py-4 space-y-0.5 overflow-y-auto">
          {navItems.map((item) => {
            const href = item.href(projectId);
            const active = location.pathname === href;
            return (
              <Link
                key={item.label}
                to={href}
                onClick={() => setSidebarOpen(false)}
                className={cn(
                  'flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors',
                  active
                    ? 'bg-primary/10 text-primary font-medium'
                    : 'text-muted-foreground hover:text-foreground hover:bg-secondary'
                )}
              >
                <item.icon className="w-4 h-4 shrink-0" />
                {item.label}
              </Link>
            );
          })}
        </nav>

        {/* Instance status */}
        <div className="px-4 py-4 border-t border-border shrink-0">
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <Circle className="w-2 h-2 fill-chart-3 text-chart-3" />
            <span className="font-mono">localhost:8080</span>
          </div>
        </div>
      </aside>

      {/* Main area */}
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        {/* Topbar */}
        <header className="h-14 flex items-center gap-4 px-4 border-b border-border bg-card shrink-0">
          <button
            className="lg:hidden text-muted-foreground hover:text-foreground"
            onClick={() => setSidebarOpen(true)}
          >
            <Menu className="w-5 h-5" />
          </button>

          {/* Project selector */}
          <div className="relative">
            <button
              onClick={() => setProjectOpen(!projectOpen)}
              className="flex items-center gap-2 px-3 py-1.5 rounded-lg border border-border bg-secondary/50 hover:bg-secondary text-sm transition-colors"
            >
              <div className="w-2 h-2 rounded-full bg-chart-3" />
              <span className="font-mono font-medium">{activeProject?.name ?? 'Project'}</span>
              <ChevronDown className="w-3.5 h-3.5 text-muted-foreground" />
            </button>

            {projectOpen && (
              <div className="absolute top-full mt-1 left-0 w-52 bg-card border border-border rounded-xl shadow-2xl z-50 py-1">
                {projects.map((project) => (
                  <Link
                    key={project.id}
                    to={`/projects/${project.id}/board`}
                    onClick={() => setProjectOpen(false)}
                    className={cn(
                      'w-full text-left flex items-center gap-2.5 px-3 py-2 text-sm hover:bg-secondary transition-colors',
                      project.id === activeProject?.id ? 'text-primary' : 'text-foreground'
                    )}
                  >
                    <div className={cn('w-2 h-2 rounded-full', project.id === activeProject?.id ? 'bg-chart-3' : 'bg-muted-foreground/40')} />
                    <span className="font-mono">{project.name}</span>
                  </Link>
                ))}
              </div>
            )}
          </div>

          <div className="flex-1" />

          <div className="hidden text-xs text-muted-foreground sm:block">
            {user?.email}
          </div>

          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={handleLogout}
            className="h-8"
          >
            Logout
          </Button>

          <div className="text-xs text-muted-foreground font-mono hidden sm:block">
            PostgreSQL 16.2
          </div>
        </header>

        {/* Page content */}
        <main className="flex-1 overflow-auto">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
