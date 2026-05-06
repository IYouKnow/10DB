import { useEffect, useState } from 'react';
import { request } from '@/lib/api';

const cards = [
  { key: 'total_users', label: 'Users' },
  { key: 'total_projects', label: 'Projects' },
  { key: 'total_databases', label: 'Databases' },
  { key: 'total_database_servers', label: 'Database Servers' },
  { key: 'active_database_servers', label: 'Active Servers' },
  { key: 'failed_databases', label: 'Failed Databases' },
];

export default function AdminOverview() {
  const [overview, setOverview] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    let mounted = true;

    async function loadOverview() {
      try {
        const data = await request('/api/admin/overview');
        if (mounted) {
          setOverview(data);
        }
      } catch (loadError) {
        if (mounted) {
          setError(loadError.message || 'Failed to load overview');
        }
      } finally {
        if (mounted) {
          setIsLoading(false);
        }
      }
    }

    loadOverview();

    return () => {
      mounted = false;
    };
  }, []);

  return (
    <div className="mx-auto max-w-7xl px-6 py-8">
      <div className="mb-8">
        <h1 className="text-2xl font-semibold">Overview</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          A lightweight snapshot of users, projects, provisioned databases, and server capacity.
        </p>
      </div>

      {isLoading ? (
        <div className="rounded-3xl border border-border bg-card px-6 py-12 text-center text-sm text-muted-foreground">
          Loading overview...
        </div>
      ) : error ? (
        <div className="rounded-3xl border border-destructive/20 bg-destructive/10 px-6 py-4 text-sm text-destructive">
          {error}
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {cards.map((card) => (
            <div key={card.key} className="rounded-3xl border border-border bg-card p-5 shadow-xl shadow-black/10">
              <div className="text-sm text-muted-foreground">{card.label}</div>
              <div className="mt-4 text-3xl font-semibold tracking-tight">
                {overview?.[card.key] ?? 0}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
