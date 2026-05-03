import React, { useEffect, useRef, useState } from 'react';
import { Database, GitBranch, Code2, Save, Play } from 'lucide-react';
import { useNavigate, useParams } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { useProjects } from '@/lib/ProjectsContext';
import { toast } from 'sonner';

export default function SchemaBoard() {
  const navigate = useNavigate();
  const { projectId } = useParams();
  const boardRef = useRef(null);
  const draggingRef = useRef(null);
  const suppressDatabaseClickRef = useRef(false);
  const [project, setProject] = useState(null);
  const [isLoadingProject, setIsLoadingProject] = useState(true);
  const [menu, setMenu] = useState(null);
  const [isProvisioning, setIsProvisioning] = useState(false);
  const [isRemovingDatabase, setIsRemovingDatabase] = useState(false);
  const [databasePosition, setDatabasePosition] = useState({ x: 80, y: 80 });
  const { getProject, provisionPostgres, removeProvisionedPostgres } = useProjects();

  useEffect(() => {
    let mounted = true;

    async function loadProject() {
      setIsLoadingProject(true);
      try {
        const data = await getProject(projectId);
        if (mounted) {
          setProject(data);
        }
      } finally {
        if (mounted) {
          setIsLoadingProject(false);
        }
      }
    }

    loadProject();

    return () => {
      mounted = false;
    };
  }, [getProject, projectId]);

  useEffect(() => {
    const closeMenu = () => setMenu(null);
    window.addEventListener('click', closeMenu);
    return () => window.removeEventListener('click', closeMenu);
  }, []);

  useEffect(() => {
    const handlePointerMove = (event) => {
      if (!draggingRef.current) {
        return;
      }

      const bounds = boardRef.current?.getBoundingClientRect();
      if (!bounds) {
        return;
      }

      const nextX = event.clientX - bounds.left - draggingRef.current.offsetX;
      const nextY = event.clientY - bounds.top - draggingRef.current.offsetY;
      const deltaX = nextX - draggingRef.current.startX;
      const deltaY = nextY - draggingRef.current.startY;

      if (Math.abs(deltaX) > 4 || Math.abs(deltaY) > 4) {
        draggingRef.current.moved = true;
      }

      setDatabasePosition({
        x: Math.max(0, Math.min(nextX, 1400 - 320)),
        y: Math.max(0, Math.min(nextY, 900 - 170)),
      });
    };

    const handlePointerUp = () => {
      if (draggingRef.current?.moved) {
        suppressDatabaseClickRef.current = true;
      }
      draggingRef.current = null;
    };

    window.addEventListener('mousemove', handlePointerMove);
    window.addEventListener('mouseup', handlePointerUp);

    return () => {
      window.removeEventListener('mousemove', handlePointerMove);
      window.removeEventListener('mouseup', handlePointerUp);
    };
  }, []);

  const handleContextMenu = (event) => {
    event.preventDefault();
    const bounds = boardRef.current?.getBoundingClientRect();
    setMenu({
      x: event.clientX - (bounds?.left ?? 0),
      y: event.clientY - (bounds?.top ?? 0),
    });
  };

  const handleProvisionPostgres = async () => {
    setIsProvisioning(true);
    try {
      const data = await provisionPostgres(projectId);
      setProject(data);
      setMenu(null);
      toast.success('PostgreSQL database added to this project');
    } catch (error) {
      toast.error(error.message || 'Failed to add PostgreSQL database');
    } finally {
      setIsProvisioning(false);
    }
  };

  const handleRemovePostgres = async () => {
    setIsRemovingDatabase(true);
    try {
      const data = await removeProvisionedPostgres(projectId);
      setProject(data);
      setMenu(null);
      toast.success('PostgreSQL database removed from this project');
    } catch (error) {
      toast.error(error.message || 'Failed to remove PostgreSQL database');
    } finally {
      setIsRemovingDatabase(false);
    }
  };

  const handleDatabaseMouseDown = (event) => {
    if (event.button !== 0) {
      return;
    }

    const bounds = boardRef.current?.getBoundingClientRect();
    if (!bounds) {
      return;
    }

    draggingRef.current = {
      offsetX: event.clientX - bounds.left - databasePosition.x,
      offsetY: event.clientY - bounds.top - databasePosition.y,
      startX: databasePosition.x,
      startY: databasePosition.y,
      moved: false,
    };
  };

  const handleDatabaseClick = () => {
    if (suppressDatabaseClickRef.current) {
      suppressDatabaseClickRef.current = false;
      return;
    }
    navigate(`/projects/${projectId}/database/schema`);
  };

  return (
    <div className="flex h-full overflow-hidden">
      {/* Board area */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Toolbar */}
        <div className="flex items-center gap-2 px-4 py-3 border-b border-border bg-card shrink-0">
          <Button size="sm" variant="outline" className="h-7 text-xs text-muted-foreground" disabled>
            <Database className="w-3.5 h-3.5 mr-1" /> Add Database
          </Button>
          <Button size="sm" variant="outline" className="h-7 text-xs text-muted-foreground" disabled>
            <GitBranch className="w-3.5 h-3.5 mr-1" /> Add Relation
          </Button>
          <div className="flex-1" />
          <Button size="sm" variant="ghost" className="h-7 text-xs text-muted-foreground" onClick={() => toast.success('Draft saved')}>
            <Save className="w-3.5 h-3.5 mr-1" /> Save Draft
          </Button>
          <Button size="sm" variant="outline" className="h-7 text-xs" onClick={() => toast.info('SQL generated')} disabled={!project?.pgDatabaseName}>
            <Code2 className="w-3.5 h-3.5 mr-1" /> Generate SQL
          </Button>
          <Button size="sm" className="h-7 text-xs bg-primary text-primary-foreground hover:bg-primary/90" onClick={() => toast.success('Schema applied!')} disabled={!project?.pgDatabaseName}>
            <Play className="w-3.5 h-3.5 mr-1" /> Apply Schema
          </Button>
        </div>

        {/* Canvas */}
        <div
          ref={boardRef}
          className="flex-1 relative overflow-auto"
          style={{ backgroundImage: 'radial-gradient(hsl(220 15% 14%) 1px, transparent 1px)', backgroundSize: '24px 24px' }}
          onContextMenu={handleContextMenu}
        >
          <div className="relative" style={{ width: 1400, height: 900 }}>
            {isLoadingProject ? (
              <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
                Loading project board...
              </div>
            ) : project?.pgDatabaseName ? (
              <div
                className="group absolute w-80 cursor-grab rounded-2xl border border-border bg-card p-5 shadow-2xl shadow-black/20 transition-colors hover:border-primary/25 active:cursor-grabbing"
                style={{ left: databasePosition.x, top: databasePosition.y }}
                onMouseDown={handleDatabaseMouseDown}
                onClick={handleDatabaseClick}
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-xl border border-primary/20 bg-primary/10">
                    <Database className="h-5 w-5 text-primary" />
                  </div>
                  <div className="flex-1">
                    <h2 className="text-sm font-semibold text-foreground">PostgreSQL Database</h2>
                    <p className="mt-1 font-mono text-xs text-muted-foreground">{project.pgDatabaseName}</p>
                    <p className="mt-2 text-xs text-muted-foreground">
                      Database provisioned. Next we can make this board add tables, relations, and schema objects inside it.
                    </p>
                  </div>
                  <button
                    type="button"
                    onClick={(event) => {
                      event.stopPropagation();
                      handleRemovePostgres();
                    }}
                    onMouseDown={(event) => event.stopPropagation()}
                    disabled={isRemovingDatabase}
                    className="rounded-lg border border-destructive/20 px-2.5 py-1.5 text-[11px] font-medium text-destructive transition-colors hover:bg-destructive/10 disabled:cursor-not-allowed disabled:opacity-60"
                  >
                    {isRemovingDatabase ? 'Removing...' : 'Delete'}
                  </button>
                </div>

                <div className="mt-4 flex items-center gap-2 border-t border-border pt-3">
                  <div className="rounded-md bg-secondary px-2 py-1 text-[11px] text-muted-foreground">
                    PostgreSQL
                  </div>
                  <div className="rounded-md bg-secondary px-2 py-1 text-[11px] text-muted-foreground">
                    Ready
                  </div>
                </div>
              </div>
            ) : (
              <div className="flex h-full items-center justify-center p-10">
                <div className="max-w-md rounded-3xl border border-dashed border-border bg-card/80 px-8 py-10 text-center shadow-2xl shadow-black/10 backdrop-blur">
                  <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-2xl border border-primary/20 bg-primary/10">
                    <Database className="h-5 w-5 text-primary" />
                  </div>
                  <h2 className="mt-4 text-lg font-semibold text-foreground">Start by adding a database</h2>
                  <p className="mt-2 text-sm leading-6 text-muted-foreground">
                    This project is just a container for now. Right-click anywhere on the board and choose PostgreSQL to add the first database.
                  </p>
                </div>
              </div>
            )}

            {menu && (
              <div
                className="absolute z-20 min-w-48 rounded-xl border border-border bg-popover p-1.5 shadow-xl shadow-black/35"
                style={{ left: menu.x, top: menu.y }}
                onClick={(event) => event.stopPropagation()}
              >
                <button
                  onClick={handleProvisionPostgres}
                  disabled={isProvisioning || Boolean(project?.pgDatabaseName)}
                  className="flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm transition-colors hover:bg-secondary disabled:cursor-not-allowed disabled:opacity-60"
                >
                  <div className="flex h-5 w-5 items-center justify-center text-primary">
                    <Database className="h-3.5 w-3.5" />
                  </div>
                  <div className="font-medium text-foreground">PostgreSQL database</div>
                </button>
                <div className="px-2.5 pt-1 pb-0.5 text-[11px] text-muted-foreground">
                  {isRemovingDatabase
                    ? 'Dropping database and role from PostgreSQL...'
                    : isProvisioning
                    ? 'Provisioning database...'
                    : project?.pgDatabaseName
                      ? 'This project already has a PostgreSQL database. You can remove it from here.'
                      : 'Add a PostgreSQL database to this project'}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
