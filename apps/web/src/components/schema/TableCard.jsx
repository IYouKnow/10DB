
import React from 'react';
import { Grip, Key, Link2 } from 'lucide-react';
import { cn } from '@/lib/utils';

const HEADER_HEIGHT = 46;

export default function TableCard({ table, selected, onSelect, onColumnSelect, selectedColumn, style, onMouseDown, onContextMenu, onApply, onResizeMouseDown }) {
  return (
    <div
      style={style}
      className={cn(
        'absolute overflow-hidden rounded-xl border bg-card shadow-2xl shadow-black/30 cursor-grab select-none transition-shadow',
        selected ? 'border-primary/60 shadow-primary/10' : 'border-border hover:border-border/80'
      )}
      onMouseDown={onMouseDown}
      onContextMenu={(e) => {
        e.stopPropagation();
        onContextMenu?.(e, table.id);
      }}
      onClick={(e) => { e.stopPropagation(); onSelect(table.id); }}
    >
      {/* Table header */}
      <div className={cn(
        'px-3 py-2 border-b flex items-center gap-2 rounded-t-xl',
        selected ? 'border-primary/30 bg-primary/5' : 'border-border bg-secondary/40'
      )}>
        <div className={cn('w-2 h-2 rounded-full shrink-0', selected ? 'bg-primary' : 'bg-primary/50')} />
        <div className="min-w-0">
          <div className="truncate font-mono text-xs font-semibold text-foreground">{table.name}</div>
          <div className="truncate font-mono text-[10px] text-muted-foreground/70">{table.id}</div>
        </div>
        <span className={cn(
          'ml-auto shrink-0 rounded-md px-1.5 py-0.5 text-[9px] font-medium uppercase tracking-[0.08em]',
          table.status === 'applied'
            ? 'bg-emerald-500/12 text-emerald-300'
            : 'bg-amber-500/12 text-amber-300'
        )}>
          {table.status === 'applied' ? 'Applied' : 'Draft'}
        </span>
        {table.status !== 'applied' && (
          <button
            type="button"
            onClick={(event) => {
              event.stopPropagation();
              onApply?.(table.id);
            }}
            className="shrink-0 rounded-md border border-primary/25 px-1.5 py-0.5 text-[9px] font-semibold text-primary transition-colors hover:bg-primary/10"
          >
            {table.isApplying ? 'Applying...' : 'Apply'}
          </button>
        )}
        <span className="shrink-0 text-[10px] text-muted-foreground/50 font-mono">{table.columns.length} cols</span>
      </div>

      {/* Columns */}
      <div
        className="divide-y divide-border/40 overflow-y-auto"
        style={{ maxHeight: `calc(100% - ${HEADER_HEIGHT}px)` }}
      >
        {table.columns.map((col) => (
          <div
            key={col.id}
            onClick={(e) => { e.stopPropagation(); onColumnSelect(table.id, col.id); }}
            className={cn(
              'px-3 py-1.5 flex items-center justify-between cursor-pointer transition-colors',
              selectedColumn?.colId === col.id && selectedColumn?.tableId === table.id
                ? 'bg-primary/8 text-primary'
                : 'hover:bg-secondary/50'
            )}
          >
            <div className="flex items-center gap-1.5 min-w-0">
              {col.primaryKey && <Key className="w-2.5 h-2.5 text-primary shrink-0" />}
              {col.fk && !col.primaryKey && <Link2 className="w-2.5 h-2.5 text-accent shrink-0" />}
              {!col.primaryKey && !col.fk && <div className="w-2.5" />}
              <span className="font-mono text-[11px] text-foreground truncate">{col.name}</span>
            </div>
            <span className="font-mono text-[10px] text-muted-foreground ml-2 shrink-0">{col.type}</span>
          </div>
        ))}
      </div>

      <button
        type="button"
        onMouseDown={(event) => {
          event.stopPropagation();
          onResizeMouseDown?.(event);
        }}
        className="absolute bottom-1 right-1 flex h-5 w-5 cursor-se-resize items-center justify-center rounded-md text-muted-foreground/60 transition-colors hover:bg-secondary/70 hover:text-foreground"
        aria-label="Resize table"
      >
        <Grip className="h-3 w-3 rotate-45" />
      </button>
    </div>
  );
}
