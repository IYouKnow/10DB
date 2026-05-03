import React, { useState } from 'react';
import { Search, Plus, Pencil, Trash2 } from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { mockTableData } from '@/lib/mockData';
import { cn } from '@/lib/utils';
import { toast } from 'sonner';

const tableNames = Object.keys(mockTableData);

export default function Tables() {
  const [selectedTable, setSelectedTable] = useState(tableNames[0]);
  const [search, setSearch] = useState('');

  const data = mockTableData[selectedTable];
  const filtered = data.rows.filter(row =>
    Object.values(row).some(v => String(v).toLowerCase().includes(search.toLowerCase()))
  );

  return (
    <div className="flex flex-col h-full overflow-hidden p-6 gap-4">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-xl font-bold text-foreground">Tables</h1>
          <p className="text-sm text-muted-foreground mt-0.5">Browse and edit table data</p>
        </div>
        <Button size="sm" className="bg-primary text-primary-foreground hover:bg-primary/90 w-fit" onClick={() => toast.info('Add row coming soon')}>
          <Plus className="w-3.5 h-3.5 mr-1.5" /> Add Row
        </Button>
      </div>

      {/* Controls */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
        {/* Table selector */}
        <div className="flex gap-1.5 flex-wrap">
          {tableNames.map(name => (
            <button
              key={name}
              onClick={() => { setSelectedTable(name); setSearch(''); }}
              className={cn(
                'px-3 py-1.5 rounded-lg border text-xs font-mono font-medium transition-colors',
                selectedTable === name
                  ? 'border-primary/50 bg-primary/10 text-primary'
                  : 'border-border bg-secondary/30 text-muted-foreground hover:text-foreground'
              )}
            >
              {name}
            </button>
          ))}
        </div>

        <div className="sm:ml-auto relative">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground" />
          <Input
            placeholder="Filter rows..."
            value={search}
            onChange={e => setSearch(e.target.value)}
            className="pl-8 h-8 text-xs w-52"
          />
        </div>
      </div>

      {/* Table */}
      <div className="flex-1 overflow-auto rounded-2xl border border-border bg-card">
        <table className="w-full min-w-max">
          <thead>
            <tr className="border-b border-border bg-secondary/30">
              {data.columns.map(col => (
                <th key={col} className="px-4 py-2.5 text-left font-mono text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
                  {col}
                </th>
              ))}
              <th className="px-4 py-2.5 text-right font-mono text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
                Actions
              </th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((row, i) => (
              <tr key={i} className="border-b border-border/50 last:border-0 hover:bg-secondary/20 transition-colors">
                {data.columns.map(col => (
                  <td key={col} className="px-4 py-3 font-mono text-xs text-foreground max-w-[200px] truncate">
                    {typeof row[col] === 'boolean'
                      ? <span className={cn('px-2 py-0.5 rounded-full text-[10px] border', row[col] ? 'bg-chart-3/10 border-chart-3/20 text-chart-3' : 'bg-secondary border-border text-muted-foreground')}>{String(row[col])}</span>
                      : String(row[col])
                    }
                  </td>
                ))}
                <td className="px-4 py-3 text-right">
                  <div className="flex items-center justify-end gap-1">
                    <button className="p-1 rounded text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors" onClick={() => toast.info('Edit coming soon')}>
                      <Pencil className="w-3 h-3" />
                    </button>
                    <button className="p-1 rounded text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors" onClick={() => toast.info('Delete coming soon')}>
                      <Trash2 className="w-3 h-3" />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {filtered.length === 0 && (
          <div className="text-center py-12 text-muted-foreground text-sm">No rows found.</div>
        )}
      </div>
    </div>
  );
}