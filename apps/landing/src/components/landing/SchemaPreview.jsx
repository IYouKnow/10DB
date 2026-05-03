import React from 'react';
import { Key, Link2 } from 'lucide-react';

const tables = [
  {
    name: 'users',
    x: 'left-4 sm:left-8 top-4 sm:top-6',
    columns: [
      { name: 'id', type: 'uuid', pk: true },
      { name: 'email', type: 'varchar(255)' },
      { name: 'full_name', type: 'text' },
      { name: 'created_at', type: 'timestamptz' },
    ],
  },
  {
    name: 'projects',
    x: 'right-4 sm:right-8 top-4 sm:top-6',
    columns: [
      { name: 'id', type: 'serial', pk: true },
      { name: 'user_id', type: 'uuid', fk: true },
      { name: 'name', type: 'varchar(120)' },
      { name: 'status', type: 'text' },
    ],
  },
  {
    name: 'tasks',
    x: 'left-1/2 -translate-x-1/2 bottom-4 sm:bottom-6',
    columns: [
      { name: 'id', type: 'serial', pk: true },
      { name: 'project_id', type: 'int', fk: true },
      { name: 'title', type: 'text' },
      { name: 'done', type: 'boolean' },
    ],
  },
];

function TableCard({ table }) {
  return (
    <div className={`absolute ${table.x} w-52 sm:w-60`}>
      <div className="bg-card border border-border rounded-xl overflow-hidden shadow-2xl shadow-black/20">
        <div className="px-4 py-2.5 border-b border-border bg-secondary/40 flex items-center gap-2">
          <div className="w-2 h-2 rounded-full bg-primary/60" />
          <span className="font-mono text-xs font-semibold text-foreground">{table.name}</span>
        </div>
        <div className="divide-y divide-border/50">
          {table.columns.map((col) => (
            <div key={col.name} className="px-4 py-1.5 flex items-center justify-between">
              <div className="flex items-center gap-2">
                {col.pk && <Key className="w-3 h-3 text-primary" />}
                {col.fk && <Link2 className="w-3 h-3 text-accent" />}
                <span className="font-mono text-xs text-foreground">{col.name}</span>
              </div>
              <span className="font-mono text-[10px] text-muted-foreground">{col.type}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

export default function SchemaPreview() {
  return (
    <div className="relative max-w-4xl mx-auto">
      <div className="rounded-2xl border border-border bg-card/50 backdrop-blur-sm overflow-hidden glow-primary">
        {/* Window chrome */}
        <div className="px-5 py-3 border-b border-border flex items-center gap-2">
          <div className="flex gap-1.5">
            <div className="w-2.5 h-2.5 rounded-full bg-red-500/60" />
            <div className="w-2.5 h-2.5 rounded-full bg-yellow-500/60" />
            <div className="w-2.5 h-2.5 rounded-full bg-green-500/60" />
          </div>
          <span className="ml-4 text-xs text-muted-foreground font-mono">schema-board — my-saas-project</span>
        </div>

        {/* Canvas */}
        <div className="relative h-72 sm:h-80 md:h-96 grid-pattern">
          {/* Connection lines (SVG) */}
          <svg className="absolute inset-0 w-full h-full pointer-events-none" xmlns="http://www.w3.org/2000/svg">
            <defs>
              <linearGradient id="lineGrad" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stopColor="hsl(199 89% 48% / 0.4)" />
                <stop offset="100%" stopColor="hsl(262 83% 58% / 0.4)" />
              </linearGradient>
            </defs>
            {/* users -> projects */}
            <line x1="42%" y1="20%" x2="58%" y2="20%" stroke="url(#lineGrad)" strokeWidth="2" strokeDasharray="6 4" />
            {/* users -> tasks */}
            <line x1="30%" y1="50%" x2="45%" y2="72%" stroke="url(#lineGrad)" strokeWidth="2" strokeDasharray="6 4" />
            {/* projects -> tasks */}
            <line x1="70%" y1="50%" x2="55%" y2="72%" stroke="url(#lineGrad)" strokeWidth="2" strokeDasharray="6 4" />
          </svg>

          {tables.map((table) => (
            <TableCard key={table.name} table={table} />
          ))}
        </div>
      </div>
    </div>
  );
}