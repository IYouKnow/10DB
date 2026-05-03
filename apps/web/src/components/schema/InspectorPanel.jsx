import React, { useState, useEffect } from 'react';
import { Trash2, Plus, X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { PG_TYPES } from '@/lib/mockData';
import { cn } from '@/lib/utils';

export default function InspectorPanel({ tables, selectedTable, selectedColumn, onUpdateTable, onClose }) {
  const table = tables.find(t => t.id === selectedTable);
  const col = table?.columns.find(c => c.id === selectedColumn?.colId);

  const [tableName, setTableName] = useState('');
  const [colState, setColState] = useState(null);

  useEffect(() => {
    if (table) setTableName(table.name);
  }, [table?.id]);

  useEffect(() => {
    if (col) setColState({ ...col });
  }, [col?.id]);

  if (!table) return null;

  const updateColField = (field, value) => {
    const updated = { ...colState, [field]: value };
    setColState(updated);
    const newColumns = table.columns.map(c => c.id === col.id ? updated : c);
    onUpdateTable({ ...table, columns: newColumns });
  };

  const addColumn = () => {
    const newCol = {
      id: `c_${Date.now()}`,
      name: 'new_column',
      type: 'text',
      primaryKey: false,
      nullable: true,
      unique: false,
      defaultValue: '',
    };
    onUpdateTable({ ...table, columns: [...table.columns, newCol] });
  };

  const deleteColumn = (colId) => {
    onUpdateTable({ ...table, columns: table.columns.filter(c => c.id !== colId) });
  };

  const renameTable = () => {
    onUpdateTable({ ...table, name: tableName });
  };

  return (
    <div className="w-64 flex flex-col border-l border-border bg-card overflow-y-auto shrink-0">
      <div className="flex items-center justify-between px-4 py-3 border-b border-border shrink-0">
        <span className="text-xs font-semibold text-foreground">Inspector</span>
        <button onClick={onClose} className="text-muted-foreground hover:text-foreground transition-colors">
          <X className="w-3.5 h-3.5" />
        </button>
      </div>

      <div className="flex-1 px-4 py-4 space-y-5 overflow-y-auto">
        {/* Table name */}
        <div>
          <label className="block text-[10px] font-semibold uppercase tracking-wider text-muted-foreground mb-1.5">Table</label>
          <div className="flex gap-2">
            <Input
              value={tableName}
              onChange={e => setTableName(e.target.value)}
              onBlur={renameTable}
              className="font-mono text-xs h-8"
            />
          </div>
        </div>

        {/* Columns list */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <label className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Columns</label>
            <button
              onClick={addColumn}
              className="flex items-center gap-1 text-[10px] text-primary hover:text-primary/80 transition-colors"
            >
              <Plus className="w-3 h-3" /> Add
            </button>
          </div>
          <div className="space-y-0.5">
            {table.columns.map((c) => (
              <div
                key={c.id}
                className={cn(
                  'flex items-center justify-between px-2 py-1.5 rounded-lg cursor-pointer group',
                  selectedColumn?.colId === c.id ? 'bg-primary/10 text-primary' : 'hover:bg-secondary text-muted-foreground'
                )}
              >
                <span className="font-mono text-[11px] truncate">{c.name}</span>
                <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  <button
                    onClick={(e) => { e.stopPropagation(); deleteColumn(c.id); }}
                    className="p-0.5 hover:text-destructive transition-colors"
                  >
                    <Trash2 className="w-2.5 h-2.5" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Column editor */}
        {col && colState && (
          <div className="space-y-3">
            <div className="h-px bg-border" />
            <label className="block text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
              Edit Column
            </label>

            <div>
              <label className="text-[10px] text-muted-foreground mb-1 block">Name</label>
              <Input
                value={colState.name}
                onChange={e => setColState({ ...colState, name: e.target.value })}
                onBlur={() => updateColField('name', colState.name)}
                className="font-mono text-xs h-8"
              />
            </div>

            <div>
              <label className="text-[10px] text-muted-foreground mb-1 block">Type</label>
              <select
                value={colState.type}
                onChange={e => updateColField('type', e.target.value)}
                className="w-full h-8 px-2 rounded-md border border-input bg-background text-xs font-mono text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
              >
                {PG_TYPES.map(t => <option key={t} value={t}>{t}</option>)}
              </select>
            </div>

            <div>
              <label className="text-[10px] text-muted-foreground mb-1 block">Default</label>
              <Input
                value={colState.defaultValue}
                onChange={e => setColState({ ...colState, defaultValue: e.target.value })}
                onBlur={() => updateColField('defaultValue', colState.defaultValue)}
                className="font-mono text-xs h-8"
                placeholder="NULL"
              />
            </div>

            <div className="space-y-2">
              {[
                { key: 'primaryKey', label: 'Primary Key' },
                { key: 'nullable', label: 'Nullable' },
                { key: 'unique', label: 'Unique' },
              ].map(({ key, label }) => (
                <label key={key} className="flex items-center justify-between cursor-pointer">
                  <span className="text-xs text-muted-foreground">{label}</span>
                  <button
                    onClick={() => updateColField(key, !colState[key])}
                    className={cn(
                      'w-8 h-4 rounded-full transition-colors relative',
                      colState[key] ? 'bg-primary' : 'bg-secondary border border-border'
                    )}
                  >
                    <div className={cn(
                      'absolute top-0.5 w-3 h-3 rounded-full bg-white transition-transform',
                      colState[key] ? 'left-4.5 translate-x-0' : 'left-0.5'
                    )} />
                  </button>
                </label>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}