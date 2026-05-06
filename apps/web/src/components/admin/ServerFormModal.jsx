import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

const SSL_OPTIONS = ['disable', 'require', 'verify-ca', 'verify-full'];

function buildInitialState(server) {
  return {
    name: server?.name ?? '',
    host: server?.host ?? '',
    port: String(server?.port ?? 5432),
    admin_username: server?.adminUsername ?? '',
    admin_password: '',
    ssl_mode: server?.sslMode ?? 'disable',
    default_database: server?.defaultDatabase ?? 'postgres',
    is_active: server?.isActive ?? true,
    is_default: server?.isDefault ?? false,
  };
}

export default function ServerFormModal({ mode, server, isSaving, error, onClose, onSubmit }) {
  const [form, setForm] = useState(() => buildInitialState(server));

  useEffect(() => {
    setForm(buildInitialState(server));
  }, [server]);

  const handleSubmit = async (event) => {
    event.preventDefault();

    const payload = {
      name: form.name.trim(),
      host: form.host.trim(),
      port: Number(form.port),
      admin_username: form.admin_username.trim(),
      ssl_mode: form.ssl_mode.trim(),
      default_database: form.default_database.trim(),
      is_active: form.is_active,
      is_default: form.is_default,
    };

    if (mode === 'create' || form.admin_password.trim()) {
      payload.admin_password = form.admin_password;
    }

    await onSubmit(payload);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 px-4 py-8">
      <div className="w-full max-w-2xl rounded-3xl border border-border bg-card p-6 shadow-2xl shadow-black/30">
        <div className="mb-6 flex items-start justify-between gap-4">
          <div>
            <h2 className="text-xl font-semibold">
              {mode === 'create' ? 'Add database server' : 'Edit database server'}
            </h2>
            <p className="mt-1 text-sm text-muted-foreground">
              {mode === 'create'
                ? 'Register a PostgreSQL admin server for new database provisioning.'
                : 'Leave the password blank to keep the stored credentials unchanged.'}
            </p>
          </div>
          <Button type="button" variant="ghost" size="sm" onClick={onClose}>
            Close
          </Button>
        </div>

        <form className="space-y-5" onSubmit={handleSubmit}>
          <div className="grid gap-4 md:grid-cols-2">
            <Field label="Name">
              <Input
                value={form.name}
                onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))}
                placeholder="Primary cluster"
              />
            </Field>
            <Field label="Host">
              <Input
                value={form.host}
                onChange={(event) => setForm((current) => ({ ...current, host: event.target.value }))}
                placeholder="db.internal"
              />
            </Field>
            <Field label="Port">
              <Input
                type="number"
                value={form.port}
                onChange={(event) => setForm((current) => ({ ...current, port: event.target.value }))}
              />
            </Field>
            <Field label="SSL mode">
              <select
                value={form.ssl_mode}
                onChange={(event) => setForm((current) => ({ ...current, ssl_mode: event.target.value }))}
                className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
              >
                {SSL_OPTIONS.map((option) => (
                  <option key={option} value={option}>{option}</option>
                ))}
              </select>
            </Field>
            <Field label="Admin username">
              <Input
                value={form.admin_username}
                onChange={(event) => setForm((current) => ({ ...current, admin_username: event.target.value }))}
                placeholder="postgres"
              />
            </Field>
            <Field label={mode === 'create' ? 'Admin password' : 'Admin password (optional)'}>
              <Input
                type="password"
                value={form.admin_password}
                onChange={(event) => setForm((current) => ({ ...current, admin_password: event.target.value }))}
                placeholder={mode === 'create' ? 'Required' : 'Only fill to change'}
              />
            </Field>
            <Field label="Default database">
              <Input
                value={form.default_database}
                onChange={(event) => setForm((current) => ({ ...current, default_database: event.target.value }))}
                placeholder="postgres"
              />
            </Field>
          </div>

          <div className="flex flex-col gap-3 rounded-2xl border border-border bg-background/60 p-4 text-sm md:flex-row md:items-center md:justify-between">
            <label className="flex items-center gap-3">
              <input
                type="checkbox"
                checked={form.is_active}
                onChange={(event) => setForm((current) => ({ ...current, is_active: event.target.checked }))}
              />
              <span>Server is active</span>
            </label>
            <label className="flex items-center gap-3">
              <input
                type="checkbox"
                checked={form.is_default}
                onChange={(event) => setForm((current) => ({ ...current, is_default: event.target.checked }))}
              />
              <span>Use as default provisioning server</span>
            </label>
          </div>

          {error ? (
            <div className="rounded-2xl border border-destructive/20 bg-destructive/10 px-4 py-3 text-sm text-destructive">
              {error}
            </div>
          ) : null}

          <div className="flex justify-end gap-3">
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={isSaving}>
              {isSaving ? 'Saving...' : mode === 'create' ? 'Create server' : 'Save changes'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}

function Field({ label, children }) {
  return (
    <label className="space-y-2">
      <div className="text-sm font-medium text-foreground">{label}</div>
      {children}
    </label>
  );
}
