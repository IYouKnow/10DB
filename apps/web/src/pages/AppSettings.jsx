import React, { useState } from 'react';
import { AlertTriangle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { toast } from 'sonner';

function SettingRow({ label, description, children }) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 py-4 border-b border-border last:border-0">
      <div>
        <p className="text-sm font-medium text-foreground">{label}</p>
        {description && <p className="text-xs text-muted-foreground mt-0.5">{description}</p>}
      </div>
      <div className="sm:w-64 shrink-0">{children}</div>
    </div>
  );
}

export default function AppSettings() {
  const [instanceUrl, setInstanceUrl] = useState('http://localhost:8080');
  const [pgHost, setPgHost] = useState('localhost');
  const [adminEmail, setAdminEmail] = useState('admin@example.com');
  const [storagePath, setStoragePath] = useState('/var/lib/10db');
  const [confirmDelete, setConfirmDelete] = useState('');

  return (
    <div className="p-6 max-w-2xl mx-auto space-y-8">
      <div>
        <h1 className="text-xl font-bold text-foreground">Settings</h1>
        <p className="text-sm text-muted-foreground mt-0.5">Instance configuration for my-saas-app</p>
      </div>

      {/* Instance settings */}
      <div className="rounded-2xl border border-border bg-card px-5">
        <h2 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground pt-4 pb-2">Instance</h2>
        <SettingRow label="Instance URL" description="Backend API endpoint">
          <Input value={instanceUrl} onChange={e => setInstanceUrl(e.target.value)} className="font-mono text-xs h-8" />
        </SettingRow>
        <SettingRow label="PostgreSQL Host" description="Managed PostgreSQL server address">
          <Input value={pgHost} onChange={e => setPgHost(e.target.value)} className="font-mono text-xs h-8" />
        </SettingRow>
        <SettingRow label="Admin Email" description="Notifications and alerts">
          <Input value={adminEmail} onChange={e => setAdminEmail(e.target.value)} className="text-xs h-8" />
        </SettingRow>
        <SettingRow label="Storage Path" description="Where backups and metadata are stored">
          <Input value={storagePath} onChange={e => setStoragePath(e.target.value)} className="font-mono text-xs h-8" />
        </SettingRow>
        <div className="py-4 flex justify-end">
          <Button size="sm" className="bg-primary text-primary-foreground hover:bg-primary/90 h-8 text-xs" onClick={() => toast.success('Settings saved')}>
            Save Changes
          </Button>
        </div>
      </div>

      {/* Danger zone */}
      <div className="rounded-2xl border border-destructive/30 bg-destructive/5 px-5">
        <div className="flex items-center gap-2 pt-4 pb-2">
          <AlertTriangle className="w-4 h-4 text-destructive" />
          <h2 className="text-xs font-semibold uppercase tracking-wider text-destructive">Danger Zone</h2>
        </div>

        <div className="py-4 border-b border-destructive/20">
          <div className="flex items-center justify-between gap-4">
            <div>
              <p className="text-sm font-medium text-foreground">Reset Project</p>
              <p className="text-xs text-muted-foreground mt-0.5">Drop all tables and restart from a blank schema. Data will be lost.</p>
            </div>
            <Button size="sm" variant="outline" className="border-destructive/40 text-destructive hover:bg-destructive hover:text-white shrink-0 h-8 text-xs" onClick={() => toast.error('Reset cancelled — confirmation required')}>
              Reset
            </Button>
          </div>
        </div>

        <div className="py-4">
          <div>
            <p className="text-sm font-medium text-foreground mb-1">Delete Project</p>
            <p className="text-xs text-muted-foreground mb-3">Permanently deletes the project, database, user, and all backups.</p>
            <div className="flex gap-2">
              <Input
                placeholder='Type "my-saas-app" to confirm'
                value={confirmDelete}
                onChange={e => setConfirmDelete(e.target.value)}
                className="text-xs h-8 flex-1"
              />
              <Button
                size="sm"
                disabled={confirmDelete !== 'my-saas-app'}
                className="bg-destructive hover:bg-destructive/90 text-white h-8 text-xs shrink-0"
                onClick={() => toast.error('Project deletion simulated')}
              >
                Delete
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}