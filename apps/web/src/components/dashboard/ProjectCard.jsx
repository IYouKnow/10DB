import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { GitBranch, Trash2, Table2, Database } from 'lucide-react';
import { cn } from '@/lib/utils';

const statusConfig = {
  draft: { label: 'Draft', color: 'text-muted-foreground bg-secondary border-border' },
  ready: { label: 'Ready', color: 'text-chart-3 bg-chart-3/10 border-chart-3/20' },
  creating: { label: 'Creating', color: 'text-chart-4 bg-chart-4/10 border-chart-4/20' },
  apply_failed: { label: 'Apply failed', color: 'text-destructive bg-destructive/10 border-destructive/20' },
  deleting: { label: 'Deleting', color: 'text-muted-foreground bg-secondary border-border' },
};

function formatTimeAgo(value) {
  const timestamp = new Date(value).getTime();
  const diffMs = Date.now() - timestamp;

  if (Number.isNaN(timestamp) || diffMs < 0) {
    return "just now";
  }

  const minute = 60 * 1000;
  const hour = 60 * minute;
  const day = 24 * hour;

  if (diffMs < hour) {
    const minutes = Math.max(1, Math.floor(diffMs / minute));
    return `${minutes}m ago`;
  }

  if (diffMs < day) {
    const hours = Math.floor(diffMs / hour);
    return `${hours}h ago`;
  }

  const days = Math.floor(diffMs / day);
  return `${days}d ago`;
}

export default function ProjectCard({ project, onDelete }) {
  const navigate = useNavigate();
  const status = statusConfig[project.status] || statusConfig.creating;
  const timeAgo = formatTimeAgo(project.updatedAt);
  const projectBasePath = `/projects/${project.id}`;
  const projectBoardPath = `${projectBasePath}/board`;

  const openProject = () => {
    navigate(projectBoardPath);
  };

  const handleCardKeyDown = (event) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      openProject();
    }
  };

  const stopCardClick = (event) => {
    event.stopPropagation();
  };

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={openProject}
      onKeyDown={handleCardKeyDown}
      className="group relative flex cursor-pointer flex-col rounded-2xl border border-border bg-card p-5 transition-all duration-200 hover:border-primary/20 focus:outline-none focus:ring-2 focus:ring-ring"
    >
      {/* Header */}
      <div className="flex items-start justify-between mb-4">
        <div>
          <h3 className="font-mono font-semibold text-sm text-foreground">{project.name}</h3>
          <p className="text-xs text-muted-foreground mt-0.5 font-mono">{project.databases?.length ? `${project.databases.length} database${project.databases.length === 1 ? '' : 's'}` : 'No database added yet'}</p>
        </div>
        <span className={cn('text-[10px] font-medium px-2 py-0.5 rounded-full border', status.color)}>
          {status.label}
        </span>
      </div>

      {/* Stats */}
      <div className="flex items-center gap-4 mb-4 text-xs text-muted-foreground">
        <div className="flex items-center gap-1.5">
          <Table2 className="w-3 h-3" />
          <span>{project.slug}</span>
        </div>
        <div className="flex items-center gap-1.5">
          <div className={cn('w-1.5 h-1.5 rounded-full', project.databases?.length ? 'bg-chart-3' : 'bg-muted-foreground/40')} />
          <span>{project.databases?.length ? `${project.databases.length} active DB cards` : 'Right-click in board to add DB'}</span>
        </div>
      </div>

      <p className="text-[10px] text-muted-foreground/60 mb-4">Updated {timeAgo}</p>

      {/* Actions */}
      <div className="flex items-center gap-2 mt-auto" onClick={stopCardClick}>
        <Link
          to={projectBoardPath}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-primary/10 hover:bg-primary/20 text-primary text-xs font-medium transition-colors"
        >
          <GitBranch className="w-3 h-3" />
          Open
        </Link>
        <Link
          to={projectBoardPath}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-secondary hover:bg-secondary/80 text-muted-foreground hover:text-foreground text-xs font-medium transition-colors"
        >
          <Database className="w-3 h-3" />
          Databases
        </Link>
        <button
          onClick={() => onDelete(project.id)}
          className="ml-auto p-1.5 rounded-lg text-muted-foreground/50 hover:text-destructive hover:bg-destructive/10 transition-colors"
        >
          <Trash2 className="w-3.5 h-3.5" />
        </button>
      </div>
    </div>
  );
}
