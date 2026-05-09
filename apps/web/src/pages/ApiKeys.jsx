import React, { useEffect, useMemo, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { ArrowLeft, CheckCheck, Copy, KeyRound, Plus, Trash2 } from 'lucide-react';
import { toast } from 'sonner';
import { useProjects } from '@/lib/ProjectsContext';

function formatDate(value) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
}

function formatPermission(value) {
  switch (value) {
    case 'read':
      return 'Read only';
    case 'write':
      return 'Write only';
    case 'read_write':
      return 'Read & write';
    default:
      return value;
  }
}

function CopyButton({ value }) {
  const [copied, setCopied] = useState(false);

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      toast.success('Copied to clipboard');
      setTimeout(() => setCopied(false), 2000);
    } catch (error) {
      toast.error(error.message || 'Failed to copy');
    }
  };

  return (
    <button
      type="button"
      onClick={copy}
      className="rounded-lg p-1.5 text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
      aria-label="Copy value"
    >
      {copied ? <CheckCheck className="h-3.5 w-3.5 text-chart-3" /> : <Copy className="h-3.5 w-3.5" />}
    </button>
  );
}

function CreateKeyModal({ open, onClose, onCreate, isCreating, createdKey }) {
  const [name, setName] = useState('');
  const [permission, setPermission] = useState('read_write');

  useEffect(() => {
    if (!open) {
      setName('');
      setPermission('read_write');
    }
  }, [open]);

  if (!open) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4">
      <div className="w-full max-w-lg rounded-2xl border border-border bg-card shadow-2xl">
        <div className="flex items-center justify-between border-b border-border px-6 py-4">
          <div className="flex items-center gap-2.5">
            <div className="flex h-8 w-8 items-center justify-center rounded-xl border border-primary/20 bg-primary/10">
              <KeyRound className="h-4 w-4 text-primary" />
            </div>
            <div>
              <h2 className="text-sm font-semibold text-foreground">Create API Key</h2>
              <p className="text-xs text-muted-foreground">The full key is only shown once.</p>
            </div>
          </div>
          <button type="button" onClick={onClose} className="text-muted-foreground transition-colors hover:text-foreground">
            x
          </button>
        </div>

        {createdKey ? (
          <div className="space-y-4 px-6 py-5">
            <div className="rounded-xl border border-chart-3/20 bg-chart-3/10 p-4 text-sm text-foreground">
              Save this key now. You will not be able to see it again after closing this modal.
            </div>
            <div>
              <div className="mb-1.5 text-xs font-medium text-muted-foreground">Name</div>
              <div className="text-sm text-foreground">{createdKey.name}</div>
            </div>
            <div>
              <div className="mb-1.5 text-xs font-medium text-muted-foreground">Permission</div>
              <div className="text-sm text-foreground">{formatPermission(createdKey.permission)}</div>
            </div>
            <div>
              <div className="mb-1.5 text-xs font-medium text-muted-foreground">API Key</div>
              <div className="flex items-center gap-2 rounded-xl bg-secondary/40 p-3">
                <div className="flex-1 break-all font-mono text-xs text-foreground">{createdKey.key}</div>
                <CopyButton value={createdKey.key} />
              </div>
            </div>
          </div>
        ) : (
          <div className="space-y-4 px-6 py-5">
            <div>
              <label className="mb-1.5 block text-xs font-medium text-muted-foreground">Key name</label>
              <input
                value={name}
                onChange={(event) => setName(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === 'Enter' && name.trim()) {
                    event.preventDefault();
                    onCreate(name.trim(), permission);
                  }
                }}
                className="h-10 w-full rounded-xl border border-input bg-background px-3 text-sm text-foreground outline-none ring-offset-background transition-colors placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring"
                placeholder="My app key"
                autoFocus
              />
            </div>
            <div>
              <label className="mb-1.5 block text-xs font-medium text-muted-foreground">Permission</label>
              <select
                value={permission}
                onChange={(event) => setPermission(event.target.value)}
                className="h-10 w-full rounded-xl border border-input bg-background px-3 text-sm text-foreground outline-none ring-offset-background transition-colors focus-visible:ring-2 focus-visible:ring-ring"
              >
                <option value="read">Read only</option>
                <option value="write">Write only</option>
                <option value="read_write">Read & write</option>
              </select>
            </div>
          </div>
        )}

        <div className="flex items-center justify-end gap-3 border-t border-border px-6 py-4">
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
          >
            {createdKey ? 'Close' : 'Cancel'}
          </button>
          {!createdKey ? (
            <button
              type="button"
              onClick={() => onCreate(name.trim(), permission)}
              disabled={isCreating || !name.trim()}
              className="rounded-lg bg-primary px-3 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {isCreating ? 'Creating...' : 'Create key'}
            </button>
          ) : null}
        </div>
      </div>
    </div>
  );
}

export default function ApiKeys() {
  const { projectId, databaseId } = useParams();
  const { getProject, listDatabaseApiKeys, createDatabaseApiKey, revokeDatabaseApiKey } = useProjects();
  const [project, setProject] = useState(null);
  const [apiKeys, setApiKeys] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isCreating, setIsCreating] = useState(false);
  const [creatingOpen, setCreatingOpen] = useState(false);
  const [createdKey, setCreatedKey] = useState(null);
  const [revokingId, setRevokingId] = useState('');

  const load = async () => {
    setIsLoading(true);
    try {
      const [projectData, keys] = await Promise.all([
        getProject(projectId),
        listDatabaseApiKeys(databaseId),
      ]);
      setProject(projectData);
      setApiKeys(keys);
    } catch (error) {
      toast.error(error.message || 'Failed to load API keys');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, [databaseId, projectId]);

  const database = useMemo(
    () => project?.databases?.find((entry) => entry.id === databaseId) ?? null,
    [databaseId, project],
  );

  const handleCreate = async (name, permission) => {
    setIsCreating(true);
    try {
      const secret = await createDatabaseApiKey(databaseId, { name, permission });
      setCreatedKey(secret);
      setApiKeys((current) => [
        {
          id: secret.id,
          name: secret.name,
          keyPrefix: secret.keyPrefix,
          permission: secret.permission,
          createdAt: secret.createdAt,
          revokedAt: null,
        },
        ...current,
      ]);
      toast.success('API key created');
    } catch (error) {
      toast.error(error.message || 'Failed to create API key');
    } finally {
      setIsCreating(false);
    }
  };

  const handleRevoke = async (keyId) => {
    setRevokingId(keyId);
    try {
      await revokeDatabaseApiKey(databaseId, keyId);
      setApiKeys((current) => current.map((key) => (
        key.id === keyId ? { ...key, revokedAt: new Date().toISOString() } : key
      )));
      toast.success('API key revoked');
    } catch (error) {
      toast.error(error.message || 'Failed to revoke API key');
    } finally {
      setRevokingId('');
    }
  };

  return (
    <div className="mx-auto flex max-w-5xl flex-col gap-6 p-6">
      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <Link
            to={`/projects/${projectId}/databases/${databaseId}/schema`}
            className="inline-flex h-8 items-center gap-1.5 rounded-lg border border-border px-2.5 text-xs text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
          >
            <ArrowLeft className="h-3.5 w-3.5" />
            Database Board
          </Link>
          <div>
            <h1 className="text-xl font-bold text-foreground">API Keys</h1>
            <p className="mt-0.5 text-sm text-muted-foreground">
              {database?.name ?? 'Database'} - <span className="font-mono">{database?.pgDatabaseName ?? databaseId}</span>
            </p>
          </div>
        </div>

        <button
          type="button"
          onClick={() => {
            setCreatedKey(null);
            setCreatingOpen(true);
          }}
          className="inline-flex items-center gap-2 rounded-lg bg-primary px-3 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
        >
          <Plus className="h-4 w-4" />
          Create key
        </button>
      </div>

      <div className="rounded-2xl border border-border bg-card p-5">
        <div className="mb-4 flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-xl border border-primary/20 bg-primary/10">
            <KeyRound className="h-4 w-4 text-primary" />
          </div>
          <div>
            <h2 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Usage</h2>
            <p className="text-xs text-muted-foreground">Use HTTPS with a bearer token instead of direct PostgreSQL credentials.</p>
          </div>
        </div>

        <div className="space-y-3 text-sm text-muted-foreground">
          <div>
            <div className="mb-1 text-xs font-medium uppercase tracking-wider">API URL</div>
            <div className="rounded-xl bg-secondary/40 p-3 font-mono text-xs text-foreground">https://your-domain.com/data/TABLE_NAME</div>
          </div>
          <div>
            <div className="mb-1 text-xs font-medium uppercase tracking-wider">Header</div>
            <div className="rounded-xl bg-secondary/40 p-3 font-mono text-xs text-foreground">Authorization: Bearer YOUR_API_KEY</div>
          </div>
          <div>
            <div className="mb-1 text-xs font-medium uppercase tracking-wider">Example</div>
            <pre className="overflow-x-auto rounded-xl bg-secondary/40 p-3 font-mono text-xs text-foreground">{`fetch("/data/users", {
  headers: { Authorization: "Bearer YOUR_API_KEY" }
})`}</pre>
          </div>
        </div>
      </div>

      <div className="rounded-2xl border border-border bg-card p-5">
        <div className="mb-4">
          <h2 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Existing Keys</h2>
          <p className="text-xs text-muted-foreground">Only the prefix is visible after creation.</p>
        </div>

        {isLoading ? (
          <div className="text-sm text-muted-foreground">Loading API keys...</div>
        ) : apiKeys.length === 0 ? (
          <div className="rounded-xl border border-dashed border-border p-5 text-sm text-muted-foreground">
            No API keys created yet.
          </div>
        ) : (
          <div className="overflow-hidden rounded-xl border border-border">
            <table className="min-w-full divide-y divide-border text-sm">
              <thead className="bg-secondary/30">
                <tr className="text-left text-xs uppercase tracking-wider text-muted-foreground">
                  <th className="px-4 py-3 font-medium">Name</th>
                  <th className="px-4 py-3 font-medium">Prefix</th>
                  <th className="px-4 py-3 font-medium">Permission</th>
                  <th className="px-4 py-3 font-medium">Created</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3 font-medium">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {apiKeys.map((key) => (
                  <tr key={key.id}>
                    <td className="px-4 py-3 text-foreground">{key.name}</td>
                    <td className="px-4 py-3 font-mono text-xs text-muted-foreground">{key.keyPrefix}</td>
                    <td className="px-4 py-3 text-muted-foreground">{formatPermission(key.permission)}</td>
                    <td className="px-4 py-3 text-muted-foreground">{formatDate(key.createdAt)}</td>
                    <td className="px-4 py-3">
                      <span className={key.revokedAt ? 'text-destructive' : 'text-chart-3'}>
                        {key.revokedAt ? `Revoked ${formatDate(key.revokedAt)}` : 'Active'}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <button
                        type="button"
                        disabled={Boolean(key.revokedAt) || revokingId === key.id}
                        onClick={() => handleRevoke(key.id)}
                        className="inline-flex items-center gap-1.5 rounded-lg border border-destructive/20 px-2.5 py-1.5 text-xs font-medium text-destructive transition-colors hover:bg-destructive/10 disabled:cursor-not-allowed disabled:opacity-60"
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                        {revokingId === key.id ? 'Revoking...' : 'Revoke'}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <CreateKeyModal
        open={creatingOpen}
        onClose={() => setCreatingOpen(false)}
        onCreate={handleCreate}
        isCreating={isCreating}
        createdKey={createdKey}
      />
    </div>
  );
}
