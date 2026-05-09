import React, { useEffect, useMemo, useState } from 'react';
import { Loader2, Pencil, Plus, Save, Trash2, X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useProjects } from '@/lib/ProjectsContext';
import { cn } from '@/lib/utils';

const COLUMN_TYPES = ['text', 'integer', 'boolean', 'uuid', 'timestamp', 'jsonb'];

function emptyColumnDraft() {
  return {
    name: '',
    type: 'text',
    nullable: true,
    primaryKey: false,
    defaultValue: '',
  };
}

function columnToDraft(column) {
  return {
    name: column.name ?? '',
    type: column.type ?? 'text',
    nullable: Boolean(column.nullable),
    primaryKey: Boolean(column.primaryKey),
    defaultValue: column.defaultValue ?? '',
  };
}

function normalizeColumns(columns) {
  return columns.map((column) => ({
    id: column.id,
    name: column.name,
    type: column.type,
    primaryKey: Boolean(column.primaryKey),
    nullable: Boolean(column.nullable),
    unique: false,
    defaultValue: column.defaultValue ?? '',
  }));
}

function ErrorBanner({ message }) {
  if (!message) {
    return null;
  }

  return (
    <div className="rounded-xl border border-destructive/20 bg-destructive/10 px-3 py-2 text-xs text-destructive">
      {message}
    </div>
  );
}

function Toggle({ checked, onChange, label }) {
  return (
    <label className="flex items-center justify-between gap-3">
      <span className="text-xs text-muted-foreground">{label}</span>
      <button
        type="button"
        onClick={() => onChange(!checked)}
        className={cn(
          'relative h-5 w-9 rounded-full transition-colors',
          checked ? 'bg-primary' : 'border border-border bg-secondary',
        )}
      >
        <span
          className={cn(
            'absolute top-0.5 h-4 w-4 rounded-full bg-white transition-transform',
            checked ? 'left-4.5' : 'left-0.5',
          )}
        />
      </button>
    </label>
  );
}

function ColumnForm({ value, onChange, disabled }) {
  return (
    <div className="space-y-3">
      <div>
        <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">Name</label>
        <Input
          value={value.name}
          onChange={(event) => onChange({ ...value, name: event.target.value })}
          className="h-8 font-mono text-xs"
          disabled={disabled}
          placeholder="column_name"
        />
      </div>

      <div>
        <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">Type</label>
        <select
          value={value.type}
          onChange={(event) => onChange({ ...value, type: event.target.value })}
          disabled={disabled}
          className="h-8 w-full rounded-lg border border-input bg-background px-2 text-xs font-mono text-foreground outline-none focus:ring-2 focus:ring-ring"
        >
          {COLUMN_TYPES.map((type) => (
            <option key={type} value={type}>{type}</option>
          ))}
        </select>
      </div>

      <div>
        <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">Default Value</label>
        <Input
          value={value.defaultValue}
          onChange={(event) => onChange({ ...value, defaultValue: event.target.value })}
          className="h-8 font-mono text-xs"
          disabled={disabled}
          placeholder="Optional expression"
        />
      </div>

      <div className="space-y-2 rounded-xl border border-border bg-secondary/20 p-3">
        <Toggle
          checked={value.nullable}
          onChange={(next) => onChange({ ...value, nullable: next })}
          label="Nullable"
        />
        <Toggle
          checked={value.primaryKey}
          onChange={(next) => onChange({ ...value, primaryKey: next, nullable: next ? false : value.nullable })}
          label="Primary key"
        />
      </div>
    </div>
  );
}

export default function InspectorPanel({
  tables,
  selectedTable,
  selectedColumn,
  onSelectColumn,
  onColumnsChange,
  onClose,
}) {
  const table = useMemo(() => tables.find((entry) => entry.id === selectedTable) ?? null, [selectedTable, tables]);
  const tableId = table?.id ?? '';
  const { listDraftTableColumns, createDraftTableColumn, updateDraftTableColumn, deleteDraftTableColumn } = useProjects();
  const [columns, setColumns] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [loadingError, setLoadingError] = useState('');
  const [createDraft, setCreateDraft] = useState(emptyColumnDraft());
  const [createError, setCreateError] = useState('');
  const [isCreating, setIsCreating] = useState(false);
  const [editingId, setEditingId] = useState('');
  const [editDraft, setEditDraft] = useState(null);
  const [editError, setEditError] = useState('');
  const [isSavingEdit, setIsSavingEdit] = useState(false);
  const [deletingId, setDeletingId] = useState('');

  useEffect(() => {
    if (!tableId) {
      return;
    }

    let mounted = true;
    setIsLoading(true);
    setLoadingError('');

    void listDraftTableColumns(tableId)
      .then((data) => {
        if (!mounted) {
          return;
        }
        setColumns(data);
        onColumnsChange?.(tableId, normalizeColumns(data));
      })
      .catch((error) => {
        if (mounted) {
          setLoadingError(error.message || 'Failed to load columns');
        }
      })
      .finally(() => {
        if (mounted) {
          setIsLoading(false);
        }
      });

    return () => {
      mounted = false;
    };
  }, [listDraftTableColumns, onColumnsChange, tableId]);

  useEffect(() => {
    if (!selectedColumn?.colId) {
      setEditingId('');
      setEditDraft(null);
      setEditError('');
      return;
    }

    const column = columns.find((entry) => entry.id === selectedColumn.colId);
    if (!column) {
      return;
    }
    setEditingId(column.id);
    setEditDraft(columnToDraft(column));
    setEditError('');
  }, [columns, selectedColumn]);

  if (!table) {
    return null;
  }

  const syncColumns = (nextColumns) => {
    setColumns(nextColumns);
    onColumnsChange?.(table.id, normalizeColumns(nextColumns));
  };

  const handleCreate = async () => {
    setIsCreating(true);
    setCreateError('');
    try {
      const created = await createDraftTableColumn(table.id, createDraft);
      const nextColumns = [...columns, created];
      syncColumns(nextColumns);
      setCreateDraft(emptyColumnDraft());
      onSelectColumn?.(table.id, created.id);
    } catch (error) {
      setCreateError(error.message || 'Failed to add column');
    } finally {
      setIsCreating(false);
    }
  };

  const handleSaveEdit = async () => {
    if (!editingId || !editDraft) {
      return;
    }

    setIsSavingEdit(true);
    setEditError('');
    try {
      const updated = await updateDraftTableColumn(table.id, editingId, editDraft);
      const nextColumns = columns.map((column) => (
        column.id === editingId ? updated : column
      ));
      syncColumns(nextColumns);
    } catch (error) {
      setEditError(error.message || 'Failed to save column');
    } finally {
      setIsSavingEdit(false);
    }
  };

  const handleDelete = async (columnId) => {
    setDeletingId(columnId);
    setEditError('');
    setCreateError('');
    try {
      await deleteDraftTableColumn(table.id, columnId);
      const nextColumns = columns.filter((column) => column.id !== columnId);
      syncColumns(nextColumns);
      if (editingId === columnId) {
        setEditingId('');
        setEditDraft(null);
      }
    } catch (error) {
      setEditError(error.message || 'Failed to delete column');
    } finally {
      setDeletingId('');
    }
  };

  const selectedColumnRecord = columns.find((column) => column.id === editingId) ?? null;

  return (
    <div className="flex w-80 shrink-0 flex-col border-l border-border bg-card">
      <div className="flex items-center justify-between border-b border-border px-4 py-3">
        <div>
          <div className="text-xs font-semibold text-foreground">Table Editor</div>
          <div className="mt-0.5 text-[11px] text-muted-foreground">Draft columns only</div>
        </div>
        <button onClick={onClose} className="text-muted-foreground transition-colors hover:text-foreground">
          <X className="h-3.5 w-3.5" />
        </button>
      </div>

      <div className="space-y-5 overflow-y-auto px-4 py-4">
        <div className="rounded-xl border border-border bg-secondary/20 p-3">
          <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Table</div>
          <div className="mt-1 font-mono text-sm text-foreground">{table.name}</div>
          <div className="mt-3 flex items-center justify-between">
            <span className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Status</span>
            <span className={cn(
              'rounded-md px-2 py-1 text-[10px] font-medium uppercase tracking-wider',
              table.status === 'applied' ? 'bg-chart-3/10 text-chart-3' : 'bg-amber-500/10 text-amber-300',
            )}>
              {table.status}
            </span>
          </div>
        </div>

        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Columns</div>
              <div className="mt-0.5 text-[11px] text-muted-foreground">Select one to edit or delete it.</div>
            </div>
            {isLoading ? <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" /> : null}
          </div>

          <ErrorBanner message={loadingError} />

          <div className="space-y-1">
            {columns.map((column) => (
              <button
                key={column.id}
                type="button"
                onClick={() => onSelectColumn?.(table.id, column.id)}
                className={cn(
                  'flex w-full items-center justify-between rounded-xl border px-3 py-2 text-left transition-colors',
                  editingId === column.id ? 'border-primary/30 bg-primary/10' : 'border-border hover:bg-secondary/40',
                )}
              >
                <div className="min-w-0">
                  <div className="truncate font-mono text-xs text-foreground">{column.name}</div>
                  <div className="mt-0.5 text-[11px] text-muted-foreground">
                    {column.type}
                    {column.primaryKey ? ' • primary key' : ''}
                    {column.nullable ? ' • nullable' : ' • required'}
                  </div>
                </div>
                <Pencil className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
              </button>
            ))}

            {!isLoading && columns.length === 0 ? (
              <div className="rounded-xl border border-dashed border-border px-3 py-4 text-xs text-muted-foreground">
                No draft columns yet.
              </div>
            ) : null}
          </div>
        </div>

        <div className="space-y-3 rounded-2xl border border-border p-4">
          <div>
            <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Add Column</div>
            <div className="mt-0.5 text-[11px] text-muted-foreground">Create a new draft column for this table.</div>
          </div>
          <ErrorBanner message={createError} />
          <ColumnForm value={createDraft} onChange={setCreateDraft} disabled={isCreating} />
          <Button size="sm" onClick={handleCreate} disabled={isCreating || !createDraft.name.trim()}>
            <Plus className="h-3.5 w-3.5" />
            {isCreating ? 'Adding...' : 'Add Column'}
          </Button>
        </div>

        <div className="space-y-3 rounded-2xl border border-border p-4">
          <div>
            <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Edit Column</div>
            <div className="mt-0.5 text-[11px] text-muted-foreground">
              {selectedColumnRecord ? `Editing ${selectedColumnRecord.name}` : 'Select a column from the list above.'}
            </div>
          </div>
          <ErrorBanner message={editError} />
          {editDraft && selectedColumnRecord ? (
            <>
              <ColumnForm value={editDraft} onChange={setEditDraft} disabled={isSavingEdit || deletingId === editingId} />
              <div className="flex items-center gap-2">
                <Button size="sm" onClick={handleSaveEdit} disabled={isSavingEdit}>
                  <Save className="h-3.5 w-3.5" />
                  {isSavingEdit ? 'Saving...' : 'Save Changes'}
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => handleDelete(editingId)}
                  disabled={deletingId === editingId}
                  className="border-destructive/20 text-destructive hover:bg-destructive/10"
                >
                  <Trash2 className="h-3.5 w-3.5" />
                  {deletingId === editingId ? 'Deleting...' : 'Delete'}
                </Button>
              </div>
            </>
          ) : (
            <div className="rounded-xl border border-dashed border-border px-3 py-4 text-xs text-muted-foreground">
              Pick a column to edit its fields.
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
