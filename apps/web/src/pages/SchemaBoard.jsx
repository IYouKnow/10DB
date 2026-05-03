import React, { useState, useRef, useCallback } from 'react';
import { Plus, GitBranch, Code2, Save, Play } from 'lucide-react';
import { Button } from '@/components/ui/button';
import TableCard from '../components/schema/TableCard';
import RelationLines from '../components/schema/RelationLines';
import InspectorPanel from '../components/schema/InspectorPanel';
import { mockTables, PG_TYPES } from '@/lib/mockData';
import { toast } from 'sonner';

export default function SchemaBoard() {
  const [tables, setTables] = useState(mockTables);
  const [selectedTable, setSelectedTable] = useState(null);
  const [selectedColumn, setSelectedColumn] = useState(null);
  const [showInspector, setShowInspector] = useState(false);
  const dragging = useRef(null);
  const boardRef = useRef(null);

  const handleMouseDown = useCallback((e, tableId) => {
    if (e.button !== 0) return;
    const table = tables.find(t => t.id === tableId);
    dragging.current = {
      tableId,
      startX: e.clientX - table.x,
      startY: e.clientY - table.y,
    };
    e.preventDefault();
  }, [tables]);

  const handleMouseMove = useCallback((e) => {
    if (!dragging.current) return;
    const { tableId, startX, startY } = dragging.current;
    setTables(prev => prev.map(t =>
      t.id === tableId
        ? { ...t, x: Math.max(0, e.clientX - startX), y: Math.max(0, e.clientY - startY) }
        : t
    ));
  }, []);

  const handleMouseUp = useCallback(() => {
    dragging.current = null;
  }, []);

  const handleBoardClick = () => {
    setSelectedTable(null);
    setSelectedColumn(null);
  };

  const addTable = () => {
    const newTable = {
      id: `tbl_${Date.now()}`,
      name: `new_table`,
      x: 120 + Math.random() * 200,
      y: 120 + Math.random() * 100,
      columns: [
        { id: `c_${Date.now()}`, name: 'id', type: 'uuid', primaryKey: true, nullable: false, unique: true, defaultValue: 'gen_random_uuid()' },
      ],
    };
    setTables(prev => [...prev, newTable]);
    setSelectedTable(newTable.id);
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
    setTables(prev => prev.map(t => t.id === updated.id ? updated : t));
  };

  return (
    <div className="flex h-full overflow-hidden">
      {/* Board area */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Toolbar */}
        <div className="flex items-center gap-2 px-4 py-3 border-b border-border bg-card shrink-0">
          <Button size="sm" variant="outline" onClick={addTable} className="h-7 text-xs">
            <Plus className="w-3.5 h-3.5 mr-1" /> Add Table
          </Button>
          <Button size="sm" variant="outline" className="h-7 text-xs text-muted-foreground">
            <GitBranch className="w-3.5 h-3.5 mr-1" /> Add Relation
          </Button>
          <div className="flex-1" />
          <Button size="sm" variant="ghost" className="h-7 text-xs text-muted-foreground" onClick={() => toast.success('Draft saved')}>
            <Save className="w-3.5 h-3.5 mr-1" /> Save Draft
          </Button>
          <Button size="sm" variant="outline" className="h-7 text-xs" onClick={() => toast.info('SQL generated')}>
            <Code2 className="w-3.5 h-3.5 mr-1" /> Generate SQL
          </Button>
          <Button size="sm" className="h-7 text-xs bg-primary text-primary-foreground hover:bg-primary/90" onClick={() => toast.success('Schema applied!')}>
            <Play className="w-3.5 h-3.5 mr-1" /> Apply Schema
          </Button>
        </div>

        {/* Canvas */}
        <div
          ref={boardRef}
          className="flex-1 relative overflow-auto"
          style={{ backgroundImage: 'radial-gradient(hsl(220 15% 14%) 1px, transparent 1px)', backgroundSize: '24px 24px' }}
          onMouseMove={handleMouseMove}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
          onClick={handleBoardClick}
        >
          <div className="relative" style={{ width: 1400, height: 900 }}>
            <RelationLines tables={tables} />
            {tables.map(table => (
              <TableCard
                key={table.id}
                table={table}
                selected={selectedTable === table.id}
                selectedColumn={selectedColumn}
                onSelect={handleSelectTable}
                onColumnSelect={handleSelectColumn}
                style={{ left: table.x, top: table.y }}
                onMouseDown={(e) => handleMouseDown(e, table.id)}
              />
            ))}
          </div>
        </div>
      </div>

      {/* Inspector */}
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