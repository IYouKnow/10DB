import { CheckCircle2, CircleOff, Plus, ServerCrash, ShieldCheck, Trash2 } from 'lucide-react';
import { useEffect, useState } from 'react';
import { toast } from 'sonner';
import ServerFormModal from '@/components/admin/ServerFormModal';
import { Button } from '@/components/ui/button';
import { request } from '@/lib/api';

export default function AdminServers() {
  const [servers, setServers] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [pageError, setPageError] = useState('');
  const [isSaving, setIsSaving] = useState(false);
  const [formError, setFormError] = useState('');
  const [modalState, setModalState] = useState({ open: false, mode: 'create', server: null });
  const [busyServerId, setBusyServerId] = useState('');

  const loadServers = async () => {
    setPageError('');
    try {
      const data = await request('/api/admin/servers');
      setServers(data.servers ?? []);
    } catch (error) {
      setPageError(error.message || 'Failed to load database servers');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadServers();
  }, []);

  const openCreate = () => {
    setFormError('');
    setModalState({ open: true, mode: 'create', server: null });
  };

  const openEdit = (server) => {
    setFormError('');
    setModalState({ open: true, mode: 'edit', server });
  };

  const closeModal = () => {
    if (isSaving) {
      return;
    }
    setFormError('');
    setModalState((current) => ({ ...current, open: false }));
  };

  const handleSubmit = async (payload) => {
    setIsSaving(true);
    setFormError('');

    try {
      if (modalState.mode === 'create') {
        await request('/api/admin/servers', {
          method: 'POST',
          body: JSON.stringify(payload),
        });
        toast.success('Database server created.');
      } else {
        await request(`/api/admin/servers/${modalState.server.id}`, {
          method: 'PATCH',
          body: JSON.stringify(payload),
        });
        toast.success('Database server updated.');
      }

      await loadServers();
      setModalState({ open: false, mode: 'create', server: null });
    } catch (error) {
      setFormError(error.message || 'Failed to save database server');
    } finally {
      setIsSaving(false);
    }
  };

  const runServerAction = async (serverId, action) => {
    setBusyServerId(serverId);
    try {
      await action();
    } catch (error) {
      toast.error(error.message || 'Action failed');
    } finally {
      setBusyServerId('');
    }
  };

  const handleTest = async (server) => {
    await runServerAction(server.id, async () => {
      const result = await request(`/api/admin/servers/${server.id}/test`, { method: 'POST' });
      if (result.success) {
        toast.success(result.message);
      } else {
        toast.error(result.message);
      }
    });
  };

  const handleSetDefault = async (server) => {
    await runServerAction(server.id, async () => {
      await request(`/api/admin/servers/${server.id}/set-default`, { method: 'POST' });
      await loadServers();
      toast.success(`"${server.name}" is now the default server.`);
    });
  };

  const handleToggleActive = async (server) => {
    await runServerAction(server.id, async () => {
      await request(`/api/admin/servers/${server.id}`, {
        method: 'PATCH',
        body: JSON.stringify({ is_active: !server.isActive }),
      });
      await loadServers();
      toast.success(server.isActive ? 'Server disabled.' : 'Server enabled.');
    });
  };

  const handleDelete = async (server) => {
    if (!window.confirm(`Delete "${server.name}"? This cannot be undone.`)) {
      return;
    }

    await runServerAction(server.id, async () => {
      await request(`/api/admin/servers/${server.id}`, { method: 'DELETE' });
      setServers((current) => current.filter((entry) => entry.id !== server.id));
      toast.success('Database server deleted.');
    });
  };

  return (
    <div className="mx-auto max-w-7xl px-6 py-8">
      <div className="mb-8 flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Database Servers</h1>
          <p className="mt-2 max-w-2xl text-sm text-muted-foreground">
            Manage the PostgreSQL admin servers used when 10DB Launch provisions new databases.
          </p>
        </div>

        <Button onClick={openCreate}>
          <Plus className="h-4 w-4" />
          Add Server
        </Button>
      </div>

      {pageError ? (
        <div className="mb-6 rounded-3xl border border-destructive/20 bg-destructive/10 px-6 py-4 text-sm text-destructive">
          {pageError}
        </div>
      ) : null}

      {isLoading ? (
        <div className="rounded-3xl border border-border bg-card px-6 py-12 text-center text-sm text-muted-foreground">
          Loading database servers...
        </div>
      ) : servers.length === 0 ? (
        <div className="rounded-3xl border border-border bg-card px-6 py-12 text-center">
          <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl border border-primary/20 bg-primary/10">
            <ServerCrash className="h-6 w-6 text-primary" />
          </div>
          <h2 className="text-lg font-semibold">No database servers yet</h2>
          <p className="mt-2 text-sm text-muted-foreground">
            Add the first PostgreSQL admin server to enable default provisioning for new databases.
          </p>
        </div>
      ) : (
        <div className="overflow-hidden rounded-3xl border border-border bg-card shadow-xl shadow-black/10">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-border text-sm">
              <thead className="bg-background/60">
                <tr className="text-left text-muted-foreground">
                  <th className="px-4 py-3 font-medium">Server</th>
                  <th className="px-4 py-3 font-medium">Connection</th>
                  <th className="px-4 py-3 font-medium">Engine</th>
                  <th className="px-4 py-3 font-medium">SSL</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3 font-medium">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {servers.map((server) => {
                  const isBusy = busyServerId === server.id;
                  return (
                    <tr key={server.id} className="align-top">
                      <td className="px-4 py-4">
                        <div className="flex flex-wrap items-center gap-2">
                          <div className="font-medium text-foreground">{server.name}</div>
                          {server.isDefault ? (
                            <span className="rounded-full border border-primary/20 bg-primary/10 px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.2em] text-primary">
                              Default
                            </span>
                          ) : null}
                        </div>
                        <div className="mt-1 text-xs text-muted-foreground">
                          Admin user: {server.adminUsername}
                        </div>
                      </td>
                      <td className="px-4 py-4 text-muted-foreground">
                        <div>{server.host}</div>
                        <div className="mt-1 text-xs">Port {server.port} · DB {server.defaultDatabase}</div>
                      </td>
                      <td className="px-4 py-4 text-muted-foreground">{server.engine}</td>
                      <td className="px-4 py-4 text-muted-foreground">{server.sslMode}</td>
                      <td className="px-4 py-4">
                        <div className="flex flex-col gap-2">
                          <StatusBadge active={server.isActive} />
                          <div className="text-xs text-muted-foreground">
                            Password stored: {server.hasPassword ? 'Yes' : 'No'}
                          </div>
                        </div>
                      </td>
                      <td className="px-4 py-4">
                        <div className="flex flex-wrap gap-2">
                          <Button variant="outline" size="sm" onClick={() => openEdit(server)}>
                            Edit
                          </Button>
                          <Button variant="outline" size="sm" disabled={isBusy} onClick={() => handleTest(server)}>
                            <ShieldCheck className="h-3.5 w-3.5" />
                            Test
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            disabled={isBusy || server.isDefault}
                            onClick={() => handleSetDefault(server)}
                          >
                            <CheckCircle2 className="h-3.5 w-3.5" />
                            Set Default
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            disabled={isBusy}
                            onClick={() => handleToggleActive(server)}
                          >
                            <CircleOff className="h-3.5 w-3.5" />
                            {server.isActive ? 'Disable' : 'Enable'}
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            disabled={isBusy}
                            onClick={() => handleDelete(server)}
                            className="border-destructive/20 text-destructive hover:bg-destructive/10"
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                            Delete
                          </Button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {modalState.open ? (
        <ServerFormModal
          mode={modalState.mode}
          server={modalState.server}
          isSaving={isSaving}
          error={formError}
          onClose={closeModal}
          onSubmit={handleSubmit}
        />
      ) : null}
    </div>
  );
}

function StatusBadge({ active }) {
  return (
    <span
      className={[
        'inline-flex w-fit items-center gap-2 rounded-full border px-2.5 py-1 text-xs font-medium',
        active
          ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-300'
          : 'border-amber-500/20 bg-amber-500/10 text-amber-300',
      ].join(' ')}
    >
      <span className={active ? 'h-2 w-2 rounded-full bg-emerald-400' : 'h-2 w-2 rounded-full bg-amber-400'} />
      {active ? 'Active' : 'Inactive'}
    </span>
  );
}
