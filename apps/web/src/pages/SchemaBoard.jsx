import React, { useEffect, useRef, useState } from 'react';
import { Database, GitBranch, Code2, Save, Play, Pencil, ChevronDown, KeyRound } from 'lucide-react';
import { SiMariadb, SiMysql, SiPostgresql, SiRedis, SiSqlite } from 'react-icons/si';
import { useNavigate, useParams } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useProjects } from '@/lib/ProjectsContext';
import { toast } from 'sonner';

const FALLBACK_BOARD_WIDTH = 1400;
const FALLBACK_BOARD_HEIGHT = 900;
const CARD_WIDTH = 320;
const CARD_HEIGHT = 190;
const DEFAULT_CARD_X = 80;
const DEFAULT_CARD_Y = 80;
const CARD_COLUMN_GAP = 360;
const CARD_ROW_GAP = 220;
const CARDS_PER_ROW = 3;

function buildDefaultPosition(index) {
  const column = index % CARDS_PER_ROW;
  const row = Math.floor(index / CARDS_PER_ROW);
  return {
    x: DEFAULT_CARD_X + column * CARD_COLUMN_GAP,
    y: DEFAULT_CARD_Y + row * CARD_ROW_GAP,
  };
}

function clampPosition(position, boardSize, cardSize = { width: CARD_WIDTH, height: CARD_HEIGHT }) {
  return {
    x: Math.max(0, Math.min(position.x, boardSize.width - cardSize.width)),
    y: Math.max(0, Math.min(position.y, boardSize.height - cardSize.height)),
  };
}

function LogoBadge({ children, tone }) {
  return (
    <div className={`flex h-8 w-8 items-center justify-center rounded-xl border shadow-[inset_0_1px_0_rgba(255,255,255,0.04)] ${tone}`}>
      {children}
    </div>
  );
}

function PostgreSQLLogo() {
  return (
    <LogoBadge tone="border-[#336791]/35 bg-[#336791]/16">
      <SiPostgresql className="h-[18px] w-[18px] text-[#7CC4FF]" />
    </LogoBadge>
  );
}

function MySQLLogo() {
  return (
    <LogoBadge tone="border-[#4479A1]/35 bg-[#4479A1]/16">
      <SiMysql className="h-[18px] w-[18px] text-[#7FC8FF]" />
    </LogoBadge>
  );
}

function SQLiteLogo() {
  return (
    <LogoBadge tone="border-[#0F80CC]/35 bg-[#0F80CC]/16">
      <SiSqlite className="h-[18px] w-[18px] text-[#78C9FF]" />
    </LogoBadge>
  );
}

function MariaDBLogo() {
  return (
    <LogoBadge tone="border-[#C0762C]/35 bg-[#C0762C]/16">
      <SiMariadb className="h-[18px] w-[18px] text-[#F5BE78]" />
    </LogoBadge>
  );
}

function RedisLogo() {
  return (
    <LogoBadge tone="border-[#DC382D]/35 bg-[#DC382D]/16">
      <SiRedis className="h-[18px] w-[18px] text-[#FF8D86]" />
    </LogoBadge>
  );
}

function FunctionLogo() {
  return (
    <LogoBadge tone="border-violet-500/35 bg-violet-500/16">
      <svg viewBox="0 0 24 24" className="h-[18px] w-[18px]" aria-hidden="true">
        <path fill="#C4B5FD" d="M9.2 5.5h6v1.9h-4v2.4h3.7v1.9h-3.7v6.8H9.2v-6.8H6.8V9.8h2.4V8.2c0-1 .2-1.8.7-2.4.6-.2 1.3-.3 2.1-.3Z"/>
      </svg>
    </LogoBadge>
  );
}

export default function SchemaBoard() {
  const navigate = useNavigate();
  const { projectId } = useParams();
  const boardRef = useRef(null);
  const toolbarAddRef = useRef(null);
  const draggingRef = useRef(null);
  const suppressDatabaseClickRef = useRef(false);
  const [project, setProject] = useState(null);
  const [isLoadingProject, setIsLoadingProject] = useState(true);
  const [menu, setMenu] = useState(null);
  const [isProvisioning, setIsProvisioning] = useState(false);
  const [removingDatabaseId, setRemovingDatabaseId] = useState(null);
  const [editingDatabaseId, setEditingDatabaseId] = useState(null);
  const [draftDatabaseName, setDraftDatabaseName] = useState('PostgreSQL Database');
  const [showAddDatabaseModal, setShowAddDatabaseModal] = useState(false);
  const [renameDraft, setRenameDraft] = useState('');
  const [isRenamingDatabase, setIsRenamingDatabase] = useState(false);
  const [boardSize, setBoardSize] = useState({
    width: FALLBACK_BOARD_WIDTH,
    height: FALLBACK_BOARD_HEIGHT,
  });
  const [databasePositions, setDatabasePositions] = useState({});
  const { getProject, provisionPostgres, updateProjectDatabase, removeProvisionedPostgres } = useProjects();

  useEffect(() => {
    let mounted = true;

    async function loadProject() {
      setIsLoadingProject(true);
      try {
        const data = await getProject(projectId);
        if (mounted) {
          setProject(data);
          setDatabasePositions((current) => {
            const next = { ...current };
            (data.databases ?? []).forEach((database, index) => {
              if (!next[database.id]) {
                const defaultPosition = buildDefaultPosition(index);
                const clampedPosition = clampPosition(defaultPosition, boardSize);
                next[database.id] = {
                  x: database.positionX ?? clampedPosition.x,
                  y: database.positionY ?? clampedPosition.y,
                };
              }
            });
            return next;
          });
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
  }, [boardSize, getProject, projectId]);

  useEffect(() => {
    const closeMenu = () => setMenu(null);
    window.addEventListener('click', closeMenu);
    return () => window.removeEventListener('click', closeMenu);
  }, []);

  useEffect(() => {
    const element = boardRef.current;
    if (!element) {
      return undefined;
    }

    const updateBoardSize = () => {
      setBoardSize({
        width: Math.max(element.clientWidth, CARD_WIDTH + 32),
        height: Math.max(element.clientHeight, CARD_HEIGHT + 32),
      });
    };

    updateBoardSize();

    const observer = new ResizeObserver(updateBoardSize);
    observer.observe(element);

    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    setDatabasePositions((current) => {
      const next = {};
      Object.entries(current).forEach(([id, position]) => {
        next[id] = clampPosition(position, boardSize);
      });
      return next;
    });
  }, [boardSize]);

  useEffect(() => {
    const handlePointerMove = (event) => {
      const dragState = draggingRef.current;
      if (!dragState) {
        return;
      }

      const bounds = boardRef.current?.getBoundingClientRect();
      if (!bounds) {
        return;
      }

      const nextX = event.clientX - bounds.left - dragState.offsetX;
      const nextY = event.clientY - bounds.top - dragState.offsetY;
      const deltaX = nextX - dragState.startX;
      const deltaY = nextY - dragState.startY;

      if (Math.abs(deltaX) > 4 || Math.abs(deltaY) > 4) {
        dragState.moved = true;
      }

      setDatabasePositions((current) => ({
        ...current,
        [dragState.databaseID]: clampPosition(
          { x: nextX, y: nextY },
          boardSize,
          { width: dragState.cardWidth, height: dragState.cardHeight },
        ),
      }));
    };

    const handlePointerUp = () => {
      if (draggingRef.current?.moved) {
        suppressDatabaseClickRef.current = true;
      }
      draggingRef.current = null;
      document.body.style.userSelect = '';
    };

    window.addEventListener('mousemove', handlePointerMove);
    window.addEventListener('mouseup', handlePointerUp);
    window.addEventListener('blur', handlePointerUp);

    return () => {
      window.removeEventListener('mousemove', handlePointerMove);
      window.removeEventListener('mouseup', handlePointerUp);
      window.removeEventListener('blur', handlePointerUp);
      document.body.style.userSelect = '';
    };
  }, [boardSize]);

  const handleContextMenu = (event) => {
    event.preventDefault();
    const bounds = boardRef.current?.getBoundingClientRect();
    openAddMenuAt({
      x: event.clientX - (bounds?.left ?? 0),
      y: event.clientY - (bounds?.top ?? 0),
    });
  };

  const openAddMenuAt = (position) => {
    setDraftDatabaseName(`PostgreSQL Database ${(project?.databases?.length ?? 0) + 1}`);
    setMenu({
      kind: 'add',
      ...position,
    });
  };

  const openAddDatabaseModal = () => {
    setDraftDatabaseName(`PostgreSQL Database ${(project?.databases?.length ?? 0) + 1}`);
    setMenu(null);
    setShowAddDatabaseModal(true);
  };

  const handleToolbarAddClick = (event) => {
    event.stopPropagation();
    const buttonBounds = toolbarAddRef.current?.getBoundingClientRect();
    const boardBounds = boardRef.current?.getBoundingClientRect();
    if (!buttonBounds || !boardBounds) {
      return;
    }

    openAddMenuAt({
      x: Math.max(12, buttonBounds.left - boardBounds.left),
      y: Math.max(12, buttonBounds.bottom - boardBounds.top + 10),
    });
  };

  const handleProvisionPostgres = async () => {
    setIsProvisioning(true);
    try {
      const data = await provisionPostgres(projectId, { name: draftDatabaseName.trim() });
      setProject(data);
      setDatabasePositions((current) => {
        const next = { ...current };
        (data.databases ?? []).forEach((database, index) => {
          if (!next[database.id]) {
            const defaultPosition = buildDefaultPosition(index);
            const clampedPosition = clampPosition(defaultPosition, boardSize);
            next[database.id] = {
              x: database.positionX ?? clampedPosition.x,
              y: database.positionY ?? clampedPosition.y,
            };
          }
        });
        return next;
      });
      setMenu(null);
      setShowAddDatabaseModal(false);
      setDraftDatabaseName(`PostgreSQL Database ${(data.databases?.length ?? 0) + 1}`);
      toast.success('PostgreSQL database added to this project');
    } catch (error) {
      toast.error(error.message || 'Failed to add PostgreSQL database');
    } finally {
      setIsProvisioning(false);
    }
  };

  const startRenamingDatabase = (database) => {
    setEditingDatabaseId(database.id);
    setRenameDraft(database.name);
  };

  const handleRenameDatabase = async (databaseId) => {
    const nextName = renameDraft.trim();
    if (!nextName) {
      toast.error('Database name is required');
      return;
    }

    setIsRenamingDatabase(true);
    try {
      const data = await updateProjectDatabase(projectId, databaseId, { name: nextName });
      setProject(data);
      setEditingDatabaseId(null);
      toast.success('Database renamed');
    } catch (error) {
      toast.error(error.message || 'Failed to rename database');
    } finally {
      setIsRenamingDatabase(false);
    }
  };

  const handleRemovePostgres = async (databaseId) => {
    setRemovingDatabaseId(databaseId);
    try {
      const data = await removeProvisionedPostgres(projectId, databaseId);
      setProject(data);
      setDatabasePositions((current) => {
        const next = {};
        const remainingIds = new Set((data.databases ?? []).map((database) => database.id));
        Object.entries(current).forEach(([id, position]) => {
          if (remainingIds.has(id)) {
            next[id] = position;
          }
        });
        return next;
      });
      setMenu(null);
      toast.success('PostgreSQL database removed from this project');
    } catch (error) {
      toast.error(error.message || 'Failed to remove PostgreSQL database');
    } finally {
      setRemovingDatabaseId(null);
    }
  };

  const handleDatabaseMouseDown = (event, databaseId) => {
    if (event.button !== 0) {
      return;
    }

    const bounds = boardRef.current?.getBoundingClientRect();
    if (!bounds) {
      return;
    }

    const position = databasePositions[databaseId] ?? { x: 80, y: 80 };
    draggingRef.current = {
      databaseID: databaseId,
      offsetX: event.clientX - bounds.left - position.x,
      offsetY: event.clientY - bounds.top - position.y,
      startX: position.x,
      startY: position.y,
      cardWidth: event.currentTarget.offsetWidth || CARD_WIDTH,
      cardHeight: event.currentTarget.offsetHeight || CARD_HEIGHT,
      moved: false,
    };
    document.body.style.userSelect = 'none';
    event.preventDefault();
  };

  const handleDatabaseClick = (databaseId) => {
    if (suppressDatabaseClickRef.current) {
      suppressDatabaseClickRef.current = false;
      return;
    }
    navigate(`/projects/${projectId}/databases/${databaseId}/schema`);
  };

  const handleDatabaseMouseUp = (event, databaseId) => {
    if (event.button !== 0) {
      return;
    }

    const dragState = draggingRef.current;
    if (!dragState || dragState.databaseID !== databaseId) {
      return;
    }

    draggingRef.current = null;
    document.body.style.userSelect = '';

    if (dragState.moved) {
      suppressDatabaseClickRef.current = true;
      return;
    }

    handleDatabaseClick(databaseId);
  };

  return (
    <div className="flex h-full overflow-hidden">
      {/* Board area */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Toolbar */}
        <div className="flex items-center gap-2 px-4 py-3 border-b border-border bg-card shrink-0">
          <Button
            ref={toolbarAddRef}
            size="sm"
            variant="outline"
            className="h-7 text-xs"
            onClick={handleToolbarAddClick}
          >
            <Database className="w-3.5 h-3.5 mr-1" /> Add Object
            <ChevronDown className="w-3.5 h-3.5 ml-1 text-muted-foreground" />
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
          className="flex-1 relative overflow-hidden"
          style={{ backgroundImage: 'radial-gradient(hsl(220 15% 14%) 1px, transparent 1px)', backgroundSize: '24px 24px' }}
          onContextMenu={handleContextMenu}
        >
          <div className="relative" style={{ width: boardSize.width, height: boardSize.height }}>
            {isLoadingProject ? (
              <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
                Loading project board...
              </div>
            ) : project?.databases?.length ? (
              project.databases.map((database) => {
                const position = databasePositions[database.id] ?? { x: 80, y: 80 };
                return (
                  <div
                    key={database.id}
                    draggable={false}
                    className="group absolute w-80 cursor-grab rounded-2xl border border-border bg-card p-5 shadow-2xl shadow-black/20 transition-colors hover:border-primary/25 active:cursor-grabbing"
                    style={{ left: position.x, top: position.y }}
                    onMouseDown={(event) => handleDatabaseMouseDown(event, database.id)}
                    onMouseUp={(event) => handleDatabaseMouseUp(event, database.id)}
                    onDragStart={(event) => event.preventDefault()}
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div className="flex h-10 w-10 items-center justify-center rounded-xl border border-primary/20 bg-primary/10">
                        <Database className="h-5 w-5 text-primary" />
                      </div>
                      <div className="flex-1">
                        {editingDatabaseId === database.id ? (
                          <div className="space-y-2">
                            <Input
                              value={renameDraft}
                              onChange={(event) => setRenameDraft(event.target.value)}
                              onMouseDown={(event) => event.stopPropagation()}
                              onClick={(event) => event.stopPropagation()}
                              onKeyDown={(event) => {
                                if (event.key === 'Enter') {
                                  event.preventDefault();
                                  handleRenameDatabase(database.id);
                                }
                                if (event.key === 'Escape') {
                                  setEditingDatabaseId(null);
                                }
                              }}
                              className="h-8 text-xs font-medium"
                              autoFocus
                            />
                            <div className="flex items-center gap-2">
                              <Button
                                size="sm"
                                className="h-7 px-2.5"
                                onMouseDown={(event) => event.stopPropagation()}
                                onClick={(event) => {
                                  event.stopPropagation();
                                  handleRenameDatabase(database.id);
                                }}
                                disabled={isRenamingDatabase}
                              >
                                {isRenamingDatabase ? 'Saving...' : 'Save'}
                              </Button>
                              <Button
                                size="sm"
                                variant="ghost"
                                className="h-7 px-2.5 text-muted-foreground"
                                onMouseDown={(event) => event.stopPropagation()}
                                onClick={(event) => {
                                  event.stopPropagation();
                                  setEditingDatabaseId(null);
                                }}
                              >
                                Cancel
                              </Button>
                            </div>
                          </div>
                        ) : (
                          <div className="flex items-start justify-between gap-2">
                            <h2 className="text-sm font-semibold text-foreground">{database.name}</h2>
                            <div className="flex items-center gap-1">
                              <button
                                type="button"
                                onMouseDown={(event) => event.stopPropagation()}
                                onClick={(event) => {
                                  event.stopPropagation();
                                  startRenamingDatabase(database);
                                }}
                                className="rounded-md p-1 text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                                aria-label="Rename database"
                              >
                                <Pencil className="h-3.5 w-3.5" />
                              </button>
                            </div>
                          </div>
                        )}
                        <p className="mt-1 font-mono text-xs text-muted-foreground">{database.pgDatabaseName}</p>
                        <p className="mt-2 text-xs text-muted-foreground">
                          Database provisioned. Click to enter the inner schema board for tables and relations.
                        </p>
                      </div>
                      <button
                        type="button"
                        onClick={(event) => {
                          event.stopPropagation();
                          handleRemovePostgres(database.id);
                        }}
                        onMouseDown={(event) => event.stopPropagation()}
                        disabled={removingDatabaseId === database.id}
                        className="rounded-lg border border-destructive/20 px-2.5 py-1.5 text-[11px] font-medium text-destructive transition-colors hover:bg-destructive/10 disabled:cursor-not-allowed disabled:opacity-60"
                      >
                        {removingDatabaseId === database.id ? 'Removing...' : 'Delete'}
                      </button>
                    </div>

                    <div className="mt-4 flex items-center gap-2 border-t border-border pt-3">
                      <div className="rounded-md bg-secondary px-2 py-1 text-[11px] text-muted-foreground">
                        PostgreSQL
                      </div>
                      <div className="rounded-md bg-secondary px-2 py-1 text-[11px] text-muted-foreground">
                        {database.status}
                      </div>
                      <button
                        type="button"
                        onClick={(event) => {
                          event.stopPropagation();
                          navigate(`/projects/${projectId}/databases/${database.id}/api-keys`);
                        }}
                        onMouseDown={(event) => event.stopPropagation()}
                        className="ml-auto rounded-md border border-border px-2 py-1 text-[11px] text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                      >
                        API Keys
                      </button>
                    </div>
                  </div>
                );
              })
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

            {menu?.kind === 'add' && (
              <div
                className="absolute z-20 min-w-48 rounded-xl border border-border bg-popover p-1.5 shadow-xl shadow-black/35"
                style={{ left: menu.x, top: menu.y }}
                onClick={(event) => event.stopPropagation()}
              >
                <div className="rounded-lg border border-border/60 bg-card/60 px-2.5 py-2">
                  <div className="text-[11px] font-medium uppercase tracking-[0.14em] text-muted-foreground">Add object</div>
                  <div className="mt-2 space-y-1">
                    <button
                      type="button"
                      onClick={openAddDatabaseModal}
                      className="flex w-full items-center gap-2 rounded-lg bg-secondary px-2.5 py-2 text-left text-sm transition-colors hover:bg-secondary/80"
                    >
                      <PostgreSQLLogo />
                      <div className="font-medium text-foreground">PostgreSQL database</div>
                    </button>
                    <button
                      type="button"
                      disabled
                      className="flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm text-muted-foreground opacity-60"
                    >
                      <MySQLLogo />
                      <div className="font-medium">MySQL database</div>
                    </button>
                    <button
                      type="button"
                      disabled
                      className="flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm text-muted-foreground opacity-60"
                    >
                      <SQLiteLogo />
                      <div className="font-medium">SQLite database</div>
                    </button>
                    <button
                      type="button"
                      disabled
                      className="flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm text-muted-foreground opacity-60"
                    >
                      <MariaDBLogo />
                      <div className="font-medium">MariaDB database</div>
                    </button>
                    <button
                      type="button"
                      disabled
                      className="flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm text-muted-foreground opacity-60"
                    >
                      <RedisLogo />
                      <div className="font-medium">Redis cache</div>
                    </button>
                    <button
                      type="button"
                      disabled
                      className="flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm text-muted-foreground opacity-60"
                    >
                      <FunctionLogo />
                      <div className="font-medium">Function</div>
                    </button>
                  </div>
                </div>
                <div className="px-2.5 pt-1 pb-0.5 text-[11px] text-muted-foreground">
                  Choose an object type first. Settings open in the next step.
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {showAddDatabaseModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4">
          <div className="w-full max-w-md rounded-2xl border border-border bg-card shadow-2xl">
            <div className="flex items-center justify-between border-b border-border px-6 py-4">
              <div className="flex items-center gap-2.5">
                <div className="flex h-7 w-7 items-center justify-center rounded-lg border border-primary/20 bg-primary/10">
                  <Database className="h-3.5 w-3.5 text-primary" />
                </div>
                <h2 className="text-sm font-semibold">Add PostgreSQL Database</h2>
              </div>
              <button
                type="button"
                onClick={() => setShowAddDatabaseModal(false)}
                className="text-muted-foreground transition-colors hover:text-foreground"
              >
                ✕
              </button>
            </div>

            <div className="space-y-4 px-6 py-5">
              <div>
                <label className="mb-1.5 block text-xs font-medium text-muted-foreground">Database name</label>
                <Input
                  value={draftDatabaseName}
                  onChange={(event) => setDraftDatabaseName(event.target.value)}
                  onKeyDown={(event) => {
                    if (event.key === 'Enter') {
                      event.preventDefault();
                      handleProvisionPostgres();
                    }
                  }}
                  className="font-mono text-sm"
                  placeholder="PostgreSQL Database"
                  autoFocus
                />
              </div>

              <div className="rounded-xl border border-border bg-secondary/20 px-3 py-3 text-xs text-muted-foreground">
                Engine: PostgreSQL
              </div>
            </div>

            <div className="flex items-center justify-end gap-3 border-t border-border px-6 py-4">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setShowAddDatabaseModal(false)}
                className="text-muted-foreground"
              >
                Cancel
              </Button>
              <Button
                size="sm"
                onClick={handleProvisionPostgres}
                disabled={isProvisioning || !draftDatabaseName.trim()}
                className="bg-primary text-primary-foreground hover:bg-primary/90"
              >
                {isProvisioning ? 'Adding...' : 'Add database'}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
