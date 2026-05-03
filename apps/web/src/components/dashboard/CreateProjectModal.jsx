import React, { useState } from 'react';
import { X, Database } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { TEMPLATES } from '@/lib/mockData';
import { cn } from '@/lib/utils';

function slugify(str) {
  return str.toLowerCase().replace(/[^a-z0-9]+/g, '_').replace(/^_|_$/g, '');
}

export default function CreateProjectModal({ onClose, onCreate, isCreating = false }) {
  const [name, setName] = useState('');
  const [template, setTemplate] = useState('blank');

  const dbName = slugify(name) || 'my_project';
  const dbUser = `app_${slugify(name) || 'my_project'}`;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4">
      <div className="w-full max-w-lg bg-card border border-border rounded-2xl shadow-2xl">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-border">
          <div className="flex items-center gap-2.5">
            <div className="w-7 h-7 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center">
              <Database className="w-3.5 h-3.5 text-primary" />
            </div>
            <h2 className="font-semibold text-sm">New Project</h2>
          </div>
          <button onClick={onClose} className="text-muted-foreground hover:text-foreground transition-colors">
            <X className="w-4 h-4" />
          </button>
        </div>

        <div className="px-6 py-5 space-y-5">
          {/* Project name */}
          <div>
            <label className="block text-xs font-medium text-muted-foreground mb-1.5">Project name</label>
            <Input
              placeholder="my-saas-app"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="font-mono text-sm"
            />
          </div>

          {/* DB previews */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-muted-foreground mb-1.5">Database name</label>
              <div className="px-3 py-2 rounded-lg border border-border bg-secondary/40 font-mono text-xs text-muted-foreground">
                {dbName}
              </div>
            </div>
            <div>
              <label className="block text-xs font-medium text-muted-foreground mb-1.5">Database user</label>
              <div className="px-3 py-2 rounded-lg border border-border bg-secondary/40 font-mono text-xs text-muted-foreground">
                {dbUser}
              </div>
            </div>
          </div>

          {/* Template selector */}
          <div>
            <label className="block text-xs font-medium text-muted-foreground mb-2">Template</label>
            <div className="grid grid-cols-2 sm:grid-cols-3 gap-2">
              {TEMPLATES.map((t) => (
                <button
                  key={t.id}
                  onClick={() => setTemplate(t.id)}
                  className={cn(
                    'p-3 rounded-xl border text-left transition-all',
                    template === t.id
                      ? 'border-primary/50 bg-primary/8 text-foreground'
                      : 'border-border bg-secondary/20 text-muted-foreground hover:border-border hover:text-foreground'
                  )}
                >
                  <p className="text-xs font-semibold">{t.label}</p>
                  <p className="text-[10px] mt-0.5 opacity-70">{t.description}</p>
                </button>
              ))}
            </div>
          </div>
        </div>

        <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-border">
          <Button variant="ghost" size="sm" onClick={onClose} className="text-muted-foreground">
            Cancel
          </Button>
          <Button
            size="sm"
            disabled={!name.trim() || isCreating}
            onClick={() => onCreate({ name: name.trim(), template })}
            className="bg-primary text-primary-foreground hover:bg-primary/90 font-medium"
          >
            {isCreating ? 'Creating...' : 'Create Project'}
          </Button>
        </div>
      </div>
    </div>
  );
}
