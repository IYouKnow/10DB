import React from 'react';
import { Download, RefreshCw, HardDrive } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { toast } from 'sonner';

const backups = [
  { id: 1, name: 'my_saas_app_2026-05-03_00-00.sql.gz', size: '2.4 MB', created: '2026-05-03 00:00:01', type: 'auto' },
  { id: 2, name: 'my_saas_app_2026-05-02_00-00.sql.gz', size: '2.3 MB', created: '2026-05-02 00:00:01', type: 'auto' },
  { id: 3, name: 'my_saas_app_2026-05-01_manual.sql.gz', size: '2.2 MB', created: '2026-05-01 14:32:00', type: 'manual' },
  { id: 4, name: 'my_saas_app_2026-04-30_00-00.sql.gz', size: '2.1 MB', created: '2026-04-30 00:00:01', type: 'auto' },
];

export default function Backups() {
  return (
    <div className="p-6 max-w-3xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-foreground">Backups</h1>
          <p className="text-sm text-muted-foreground mt-0.5">Automated daily backups for my-saas-app</p>
        </div>
        <Button size="sm" className="bg-primary text-primary-foreground hover:bg-primary/90" onClick={() => toast.success('Backup started')}>
          <RefreshCw className="w-3.5 h-3.5 mr-1.5" /> Create Backup
        </Button>
      </div>

      <div className="rounded-2xl border border-border bg-card overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="border-b border-border bg-secondary/30">
              <th className="px-5 py-2.5 text-left text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">File</th>
              <th className="px-5 py-2.5 text-left text-[10px] font-semibold uppercase tracking-wider text-muted-foreground hidden sm:table-cell">Size</th>
              <th className="px-5 py-2.5 text-left text-[10px] font-semibold uppercase tracking-wider text-muted-foreground hidden md:table-cell">Created</th>
              <th className="px-5 py-2.5 text-left text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Type</th>
              <th className="px-5 py-2.5" />
            </tr>
          </thead>
          <tbody>
            {backups.map(b => (
              <tr key={b.id} className="border-b border-border/50 last:border-0 hover:bg-secondary/20 transition-colors">
                <td className="px-5 py-3">
                  <div className="flex items-center gap-2">
                    <HardDrive className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
                    <span className="font-mono text-xs text-foreground truncate max-w-[180px] sm:max-w-xs">{b.name}</span>
                  </div>
                </td>
                <td className="px-5 py-3 font-mono text-xs text-muted-foreground hidden sm:table-cell">{b.size}</td>
                <td className="px-5 py-3 font-mono text-xs text-muted-foreground hidden md:table-cell">{b.created}</td>
                <td className="px-5 py-3">
                  <span className={`text-[10px] px-2 py-0.5 rounded-full border font-medium ${b.type === 'auto' ? 'bg-primary/8 border-primary/20 text-primary' : 'bg-chart-2/10 border-chart-2/20 text-chart-2'}`}>
                    {b.type}
                  </span>
                </td>
                <td className="px-5 py-3 text-right">
                  <button className="text-muted-foreground hover:text-foreground transition-colors" onClick={() => toast.success('Download started')}>
                    <Download className="w-3.5 h-3.5" />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}