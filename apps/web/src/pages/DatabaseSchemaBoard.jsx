import React, { useCallback, useEffect, useRef, useState } from 'react';
import { Plus, GitBranch, Code2, Save, Play, ArrowLeft, Table2, Link2, FunctionSquare, KeyRound } from 'lucide-react';
import { Link, useParams } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import TableCard from '../components/schema/TableCard';
import RelationLines from '../components/schema/RelationLines';
import InspectorPanel from '../components/schema/InspectorPanel';
import { useProjects } from '@/lib/ProjectsContext';
import { toast } from 'sonner';

const DEFAULT_TABLE_WIDTH = 224;
const DEFAULT_TABLE_HEIGHT = 220;
const MIN_TABLE_WIDTH = 224;
const MIN_TABLE_HEIGHT = 160;
const MAX_TABLE_WIDTH = 520;
const MAX_TABLE_HEIGHT = 640;

function blueprintToTables(blueprint, appliedTableNames) {
  const applied = new Set(appliedTableNames);
  return (blueprint?.tables ?? []).map((table) => ({
    id: table.id,
    name: table.name,
    x: table.position?.x ?? 120,
    y: table.position?.y ?? 120,
    width: table.width ?? DEFAULT_TABLE_WIDTH,
    height: table.height ?? DEFAULT_TABLE_HEIGHT,
    status: applied.has(table.name) ? 'applied' : 'draft',
    isApplying: false,
    columns: (table.columns ?? []).map((column) => ({
      id: column.id,
      name: column.name,
      type: column.type,
      primaryKey: column.primaryKey,
      nullable: column.nullable,
      unique: column.unique,
      defaultValue: column.default?.value ?? '',
    })),
  }));
}

function tablesToBlueprint(projectId, databaseId, tables) {
  return {
    version: 1,
    projectId,
    databaseId,
    tables: tables.map((table) => ({
      id: table.id,
      name: table.name,
      position: { x: table.x, y: table.y },
      width: table.width,
      height: table.height,
      columns: table.columns.map((column) => ({
        id: column.id,
        name: column.name,
        type: column.type,
        primaryKey: Boolean(column.primaryKey),
        unique: Boolean(column.unique),
        nullable: Boolean(column.nullable),
        default: column.defaultValue
          ? { kind: 'expression', value: column.defaultValue }
          : null,
        config: {},
      })),
      foreignKeys: [],
    })),
  };
}

function withAppliedStatus(tables, appliedTableNames) {
  const applied = new Set(appliedTableNames);
  return tables.map((table) => ({
    ...table,
    status: applied.has(table.name) ? 'applied' : 'draft',
    isApplying: false,
  }));
}

function buildUniqueTableName(tables) {
  const names = new Set(tables.map((table) => table.name));
  if (!names.has('new_table')) {
    return 'new_table';
  }

  let index = 2;
  while (names.has(`new_table_${index}`)) {
    index += 1;
  }
  return `new_table_${index}`;
}

export default function DatabaseSchemaBoard() {
  const { projectId, databaseId } = useParams();
  const boardViewportRef = useRef(null);
  const [tables, setTables] = useState([]);
  const [appliedTableNames, setAppliedTableNames] = useState([]);
  const [selectedTable, setSelectedTable] = useState(null);
  const [selectedColumn, setSelectedColumn] = useState(null);
  const [showInspector, setShowInspector] = useState(false);
  const [menu, setMenu] = useState(null);
  const [isLoadingSchema, setIsLoadingSchema] = useState(true);
  const [isSavingDraft, setIsSavingDraft] = useState(false);
  const [deletingTableId, setDeletingTableId] = useState(null);
  const dragging = useRef(null);
  const resizing = useRef(null);
  const { getDatabaseSchema, saveDatabaseSchema, listDatabaseTables, applyDatabaseTable, deleteDatabaseTable } = useProjects();

  const persistTables = useCallback(async (nextTables, options = {}) => {
    setTables(withAppliedStatus(nextTables, appliedTableNames));
    setIsSavingDraft(true);
    try {
      await saveDatabaseSchema(projectId, databaseId, tablesToBlueprint(projectId, databaseId, nextTables));
      if (!options.silent) {
        toast.success('Draft saved');
      }
    } catch (error) {
      toast.error(error.message || 'Failed to save draft');
    } finally {
      setIsSavingDraft(false);
    }
  }, [appliedTableNames, databaseId, projectId, saveDatabaseSchema]);

  useEffect(() => {
    let mounted = true;

    async function loadSchema() {
      setIsLoadingSchema(true);
      try {
        const [schema, liveTables] = await Promise.all([
          getDatabaseSchema(projectId, databaseId),
          listDatabaseTables(projectId, databaseId),
        ]);
        if (!mounted) {
          return;
        }
        const appliedNames = liveTables.map((table) => table.name);
        setAppliedTableNames(appliedNames);
        const blueprint = schema?.blueprint ?? schema;
        setTables(blueprintToTables(blueprint, appliedNames));
      } catch (error) {
        if (mounted) {
          toast.error(error.message || 'Failed to load database schema');
        }
      } finally {
        if (mounted) {
          setIsLoadingSchema(false);
        }
      }
    }

    loadSchema();

    return () => {
      mounted = false;
    };
  }, [databaseId, getDatabaseSchema, listDatabaseTables, projectId]);

  const handleMouseDown = useCallback((event, tableId) => {
    if (event.button !== 0) {
      return;
    }
    if (resizing.current) {
      return;
    }

    const table = tables.find((entry) => entry.id === tableId);
    if (!table) {
      return;
    }

    dragging.current = {
      tableId,
      startX: event.clientX - table.x,
      startY: event.clientY - table.y,
    };
    event.preventDefault();
  }, [tables]);

  const handleMouseMove = useCallback((event) => {
    if (resizing.current) {
      const { tableId, startWidth, startHeight, startClientX, startClientY } = resizing.current;
      const nextWidth = Math.max(MIN_TABLE_WIDTH, Math.min(MAX_TABLE_WIDTH, startWidth + (event.clientX - startClientX)));
      const nextHeight = Math.max(MIN_TABLE_HEIGHT, Math.min(MAX_TABLE_HEIGHT, startHeight + (event.clientY - startClientY)));
      setTables((current) => current.map((table) => (
        table.id === tableId
          ? { ...table, width: nextWidth, height: nextHeight }
          : table
      )));
      return;
    }

    if (!dragging.current) {
      return;
    }

    const { tableId, startX, startY } = dragging.current;
    setTables((current) => current.map((table) => (
      table.id === tableId
        ? { ...table, x: Math.max(0, event.clientX - startX), y: Math.max(0, event.clientY - startY) }
        : table
    )));
  }, []);

  const handleMouseUp = useCallback(() => {
    if (resizing.current) {
      const nextTables = tables.map((table) => ({ ...table }));
      void persistTables(nextTables, { silent: true });
      resizing.current = null;
      return;
    }
    if (dragging.current) {
      const nextTables = tables.map((table) => ({ ...table }));
      void persistTables(nextTables, { silent: true });
    }
    dragging.current = null;
  }, [persistTables, tables]);

  const handleBoardClick = () => {
    setSelectedTable(null);
    setSelectedColumn(null);
    setMenu(null);
  };

  const addTable = async () => {
    const tableName = buildUniqueTableName(tables);
    const newTable = {
      id: `tbl_${Date.now()}`,
      name: tableName,
      x: 120 + Math.random() * 220,
      y: 120 + Math.random() * 120,
      width: DEFAULT_TABLE_WIDTH,
      height: DEFAULT_TABLE_HEIGHT,
      status: 'draft',
      isApplying: false,
      columns: [
        {
          id: `c_${Date.now()}`,
          name: 'id',
          type: 'uuid',
          primaryKey: true,
          nullable: false,
          unique: true,
          defaultValue: 'gen_random_uuid()',
        },
      ],
    };

    const nextTables = [...tables, newTable];
    await persistTables(nextTables, { silent: true });
    setSelectedTable(newTable.id);
    setSelectedColumn(null);
    setShowInspector(true);
  };

  const handleSelectTable = (id) => {
    setSelectedTable(id);
    setSelectedColumn(null);
    setShowInspector(true);
  };

  const handleSelectColumn = (tableId, colId) => {
    setSelectedTable(tableId);
    setSelectedColumn({ tableId, colId });
    setShowInspector(true);
  };

  const handleColumnsChange = useCallback((tableId, nextColumns) => {
    setTables((current) => current.map((table) => (
      table.id === tableId ? { ...table, columns: nextColumns } : table
    )));
  }, []);

  const handleResizeStart = useCallback((event, tableId) => {
    if (event.button !== 0) {
      return;
    }
    const table = tables.find((entry) => entry.id === tableId);
    if (!table) {
      return;
    }
    resizing.current = {
      tableId,
      startWidth: table.width ?? DEFAULT_TABLE_WIDTH,
      startHeight: table.height ?? DEFAULT_TABLE_HEIGHT,
      startClientX: event.clientX,
      startClientY: event.clientY,
    };
    event.preventDefault();
    event.stopPropagation();
  }, [tables]);

  const handleUpdateTable = (updated) => {
    const nextTables = tables.map((table) => (
      table.id === updated.id ? updated : table
    ));
    void persistTables(nextTables, { silent: true });
  };

  useEffect(() => {
    const closeMenu = () => setMenu(null);
    window.addEventListener('click', closeMenu);
    return () => window.removeEventListener('click', closeMenu);
  }, []);

  const getMenuPosition = (event) => {
    const bounds = boardViewportRef.current?.getBoundingClientRect();
    return {
      x: (event.clientX - (bounds?.left ?? 0)) + (boardViewportRef.current?.scrollLeft ?? 0),
      y: (event.clientY - (bounds?.top ?? 0)) + (boardViewportRef.current?.scrollTop ?? 0),
    };
  };

  const handleContextMenu = (event) => {
    event.preventDefault();
    setMenu({
      kind: 'board',
      ...getMenuPosition(event),
    });
  };

  const handleAddTableFromMenu = () => {
    void addTable();
    setMenu(null);
  };

  const handleTableContextMenu = (event, tableId) => {
    event.preventDefault();
    setSelectedTable(tableId);
    setSelectedColumn(null);
    setShowInspector(true);
    setMenu({
      kind: 'table',
      tableId,
      ...getMenuPosition(event),
    });
  };

  const handleDeleteTable = async () => {
    if (!menu?.tableId) {
      return;
    }

    const targetTableId = menu.tableId;
    const targetTable = tables.find((table) => table.id === targetTableId);
    if (!targetTable) {
      return;
    }

    setDeletingTableId(targetTable.id);
    setMenu(null);

    if (targetTable.status === 'applied') {
      try {
        await deleteDatabaseTable(projectId, databaseId, targetTable.id);
      } catch (error) {
        setDeletingTableId(null);
        toast.error(error.message || 'Failed to delete table from PostgreSQL');
        return;
      }
    }

    const nextTables = tables.filter((table) => table.id !== targetTableId);
    await persistTables(nextTables, { silent: true });
    const liveTables = await listDatabaseTables(projectId, databaseId);
    const appliedNames = liveTables.map((table) => table.name);
    setAppliedTableNames(appliedNames);
    setTables(withAppliedStatus(nextTables, appliedNames));
    if (selectedTable === targetTableId) {
      setSelectedTable(null);
      setSelectedColumn(null);
      setShowInspector(false);
    }
    setDeletingTableId(null);
    toast.success(targetTable.status === 'applied' ? 'Table deleted from PostgreSQL and draft' : 'Draft table deleted');
  };

  const handleApplyTable = async (tableId) => {
    setTables((current) => current.map((table) => (
      table.id === tableId ? { ...table, isApplying: true } : table
    )));
    try {
      await saveDatabaseSchema(projectId, databaseId, tablesToBlueprint(projectId, databaseId, tables));
      await applyDatabaseTable(projectId, databaseId, tableId);
      const liveTables = await listDatabaseTables(projectId, databaseId);
      const appliedNames = liveTables.map((table) => table.name);
      setAppliedTableNames(appliedNames);
      setTables((current) => withAppliedStatus(current, appliedNames));
      toast.success('Table applied to PostgreSQL');
    } catch (error) {
      setTables((current) => current.map((table) => (
        table.id === tableId ? { ...table, isApplying: false } : table
      )));
      toast.error(error.message || 'Failed to apply table');
    }
  };

  return (
    <div className="flex h-full overflow-hidden">
      <div className="flex flex-1 flex-col overflow-hidden">
        <div className="flex items-center gap-2 border-b border-border bg-card px-4 py-3 shrink-0">
          <Link
            to={`/projects/${projectId}/board`}
            className="inline-flex h-7 items-center gap-1.5 rounded-lg border border-border px-2.5 text-xs text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
          >
            <ArrowLeft className="h-3.5 w-3.5" />
            Project Board
          </Link>
          <Button size="sm" variant="outline" onClick={addTable} className="h-7 text-xs">
            <Plus className="mr-1 h-3.5 w-3.5" /> Add Table
          </Button>
          <Button size="sm" variant="outline" className="h-7 text-xs text-muted-foreground" disabled>
            <GitBranch className="mr-1 h-3.5 w-3.5" /> Add Relation
          </Button>
          <Link
            to={`/projects/${projectId}/databases/${databaseId}/api-keys`}
            className="inline-flex h-7 items-center gap-1.5 rounded-md border border-border px-2.5 text-xs text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
          >
            <KeyRound className="h-3.5 w-3.5" />
            API Keys
          </Link>
          <div className="flex-1" />
          <Button size="sm" variant="ghost" className="h-7 text-xs text-muted-foreground" onClick={() => void persistTables(tables)}>
            <Save className="mr-1 h-3.5 w-3.5" /> Save Draft
          </Button>
          <Button size="sm" variant="outline" className="h-7 text-xs" onClick={() => toast.info('SQL generation coming next')}>
            <Code2 className="mr-1 h-3.5 w-3.5" /> Generate SQL
          </Button>
          <Button size="sm" className="h-7 bg-primary text-xs text-primary-foreground hover:bg-primary/90" onClick={() => toast.info('Use each table Apply button for now')}>
            <Play className="mr-1 h-3.5 w-3.5" /> Apply Schema
          </Button>
        </div>

        <div
          ref={boardViewportRef}
          className="relative flex-1 overflow-auto"
          style={{ backgroundImage: 'radial-gradient(hsl(220 15% 14%) 1px, transparent 1px)', backgroundSize: '24px 24px' }}
          onMouseMove={handleMouseMove}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
          onClick={handleBoardClick}
          onContextMenu={handleContextMenu}
        >
          <div className="relative" style={{ width: 1600, height: 960 }}>
            {isLoadingSchema ? (
              <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
                Loading database schema...
              </div>
            ) : tables.length === 0 ? (
              <div className="flex h-full items-center justify-center p-10">
                <div className="max-w-md rounded-3xl border border-dashed border-border bg-card/80 px-8 py-10 text-center shadow-2xl shadow-black/10 backdrop-blur">
                  <h2 className="text-lg font-semibold text-foreground">Database schema board</h2>
                  <p className="mt-2 text-sm leading-6 text-muted-foreground">
                    This is where tables will live. Add your first table to start shaping the schema inside database `{databaseId}`.
                  </p>
                </div>
              </div>
            ) : (
              <>
                <RelationLines tables={tables} />
                {tables.map((table) => (
                  <TableCard
                    key={table.id}
                    table={table}
                    deleting={deletingTableId === table.id}
                    selected={selectedTable === table.id}
                  selectedColumn={selectedColumn}
                  onSelect={handleSelectTable}
                  onColumnSelect={handleSelectColumn}
                  onContextMenu={handleTableContextMenu}
                  onApply={handleApplyTable}
                  style={{ left: table.x, top: table.y, width: table.width, height: table.height }}
                  onMouseDown={(event) => handleMouseDown(event, table.id)}
                  onResizeMouseDown={(event) => handleResizeStart(event, table.id)}
                />
              ))}
            </>
            )}

            {menu?.kind === 'board' && (
              <div
                className="absolute z-20 min-w-52 rounded-xl border border-border bg-popover p-1.5 shadow-xl shadow-black/35"
                style={{ left: menu.x, top: menu.y }}
                onClick={(event) => event.stopPropagation()}
              >
                <button
                  onClick={handleAddTableFromMenu}
                  className="flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm transition-colors hover:bg-secondary"
                >
                  <div className="flex h-5 w-5 items-center justify-center text-primary">
                    <Table2 className="h-3.5 w-3.5" />
                  </div>
                  <div className="font-medium text-foreground">Add table</div>
                </button>
                <button
                  disabled
                  className="flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm text-muted-foreground opacity-60"
                >
                  <div className="flex h-5 w-5 items-center justify-center">
                    <Link2 className="h-3.5 w-3.5" />
                  </div>
                  <div className="font-medium">Add relation</div>
                </button>
                <button
                  disabled
                  className="flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm text-muted-foreground opacity-60"
                >
                  <div className="flex h-5 w-5 items-center justify-center">
                    <FunctionSquare className="h-3.5 w-3.5" />
                  </div>
                  <div className="font-medium">Add function</div>
                </button>
                <div className="px-2.5 pt-1 pb-0.5 text-[11px] text-muted-foreground">
                  {isSavingDraft ? 'Saving draft...' : 'Tables save to the control DB immediately. Apply sends them to PostgreSQL.'}
                </div>
              </div>
            )}

            {menu?.kind === 'table' && (
              <div
                className="absolute z-20 min-w-44 rounded-xl border border-border bg-popover p-1.5 shadow-xl shadow-black/35"
                style={{ left: menu.x, top: menu.y }}
                onClick={(event) => event.stopPropagation()}
              >
                <button
                  onClick={handleDeleteTable}
                  className="flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm transition-colors hover:bg-secondary"
                >
                  <div className="flex h-5 w-5 items-center justify-center text-destructive">
                    <Table2 className="h-3.5 w-3.5" />
                  </div>
                  <div className="font-medium text-foreground">Delete table</div>
                </button>
                <div className="px-2.5 pt-1 pb-0.5 text-[11px] text-muted-foreground">
                  More table actions can live here next.
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {showInspector && (
        <InspectorPanel
          tables={tables}
          selectedTable={selectedTable}
          selectedColumn={selectedColumn}
          onSelectColumn={handleSelectColumn}
          onColumnsChange={handleColumnsChange}
          onUpdateTable={handleUpdateTable}
          onClose={() => setShowInspector(false)}
        />
      )}
    </div>
  );
}
