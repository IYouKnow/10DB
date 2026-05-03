import React, { useCallback, useEffect, useRef, useState } from 'react';
import { Plus, GitBranch, Code2, Save, Play, ArrowLeft, Table2, Link2, FunctionSquare } from 'lucide-react';
import { Link, useParams } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import TableCard from '../components/schema/TableCard';
import RelationLines from '../components/schema/RelationLines';
import InspectorPanel from '../components/schema/InspectorPanel';
import { toast } from 'sonner';

export default function DatabaseSchemaBoard() {
  const { projectId, databaseId } = useParams();
  const boardViewportRef = useRef(null);
  const [tables, setTables] = useState([]);
  const [selectedTable, setSelectedTable] = useState(null);
  const [selectedColumn, setSelectedColumn] = useState(null);
  const [showInspector, setShowInspector] = useState(false);
  const [menu, setMenu] = useState(null);
  const dragging = useRef(null);

  const handleMouseDown = useCallback((event, tableId) => {
    if (event.button !== 0) {
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
    dragging.current = null;
  }, []);

  const handleBoardClick = () => {
    setSelectedTable(null);
    setSelectedColumn(null);
    setMenu(null);
  };

  const addTable = () => {
    const newTable = {
      id: `tbl_${Date.now()}`,
      name: 'new_table',
      x: 120 + Math.random() * 220,
      y: 120 + Math.random() * 120,
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

    setTables((current) => [...current, newTable]);
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

  const handleUpdateTable = (updated) => {
    setTables((current) => current.map((table) => (
      table.id === updated.id ? updated : table
    )));
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
    addTable();
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

  const handleDeleteTable = () => {
    if (!menu?.tableId) {
      return;
    }

    setTables((current) => current.filter((table) => table.id !== menu.tableId));
    if (selectedTable === menu.tableId) {
      setSelectedTable(null);
      setSelectedColumn(null);
      setShowInspector(false);
    }
    setMenu(null);
    toast.success('Table deleted');
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
          <div className="flex-1" />
          <Button size="sm" variant="ghost" className="h-7 text-xs text-muted-foreground" onClick={() => toast.success('Database schema draft saved')}>
            <Save className="mr-1 h-3.5 w-3.5" /> Save Draft
          </Button>
          <Button size="sm" variant="outline" className="h-7 text-xs" onClick={() => toast.info('SQL generation coming next')}>
            <Code2 className="mr-1 h-3.5 w-3.5" /> Generate SQL
          </Button>
          <Button size="sm" className="h-7 bg-primary text-xs text-primary-foreground hover:bg-primary/90" onClick={() => toast.success('Database schema apply coming next')}>
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
            {tables.length === 0 ? (
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
                    selected={selectedTable === table.id}
                  selectedColumn={selectedColumn}
                  onSelect={handleSelectTable}
                  onColumnSelect={handleSelectColumn}
                  onContextMenu={handleTableContextMenu}
                  style={{ left: table.x, top: table.y }}
                  onMouseDown={(event) => handleMouseDown(event, table.id)}
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
                  Add tables now. Relations and functions can come next.
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
          onUpdateTable={handleUpdateTable}
          onClose={() => setShowInspector(false)}
        />
      )}
    </div>
  );
}
