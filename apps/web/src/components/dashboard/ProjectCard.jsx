import React from 'react';
import { Link } from 'react-router-dom';
import { GitBranch, KeyRound, Code2, Trash2, Table2 } from 'lucide-react';
import { cn } from '@/lib/utils';

const statusConfig = {
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
  const status = statusConfig[project.status] || statusConfig.creating;
  const timeAgo = formatTimeAgo(project.updatedAt);
  const projectBasePath = `/projects/${project.id}`;

  return (
    <div className="group relative flex flex-col p-5 rounded-2xl border border-border bg-card hover:border-primary/20 transition-all duration-200">
      {/* Header */}
      <div className="flex items-start justify-between mb-4">
        <div>
          <h3 className="font-mono font-semibold text-sm text-foreground">{project.name}</h3>
          <p className="text-xs text-muted-foreground mt-0.5 font-mono">{project.pgDatabaseName}</p>
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
          <div className="w-1.5 h-1.5 rounded-full bg-chart-3" />
          <span>{project.pgHost}:{project.pgPort}</span>
        </div>
      </div>

      <p className="text-[10px] text-muted-foreground/60 mb-4">Updated {timeAgo}</p>

      {/* Actions */}
      <div className="flex items-center gap-2 mt-auto">
        <Link
          to={`${projectBasePath}/board`}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-primary/10 hover:bg-primary/20 text-primary text-xs font-medium transition-colors"
        >
          <GitBranch className="w-3 h-3" />
          Open
        </Link>
        <Link
          to={`${projectBasePath}/credentials`}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-secondary hover:bg-secondary/80 text-muted-foreground hover:text-foreground text-xs font-medium transition-colors"
        >
          <KeyRound className="w-3 h-3" />
          Creds
        </Link>
        <Link
          to={`${projectBasePath}/sql`}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-secondary hover:bg-secondary/80 text-muted-foreground hover:text-foreground text-xs font-medium transition-colors"
        >
          <Code2 className="w-3 h-3" />
          SQL
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
