import { ShieldAlert } from 'lucide-react';
import { Outlet } from 'react-router-dom';
import { useAuth } from '@/lib/AuthContext';

const LoadingFallback = () => (
  <div className="fixed inset-0 flex items-center justify-center">
    <div className="h-8 w-8 animate-spin rounded-full border-4 border-slate-200 border-t-slate-800"></div>
  </div>
);

const AccessDenied = () => (
  <div className="min-h-screen bg-background text-foreground">
    <div className="mx-auto flex min-h-screen max-w-2xl items-center px-6 py-10">
      <div className="w-full rounded-3xl border border-border bg-card p-8 shadow-2xl shadow-black/20">
        <div className="mb-5 flex h-12 w-12 items-center justify-center rounded-2xl border border-destructive/20 bg-destructive/10 text-destructive">
          <ShieldAlert className="h-6 w-6" />
        </div>
        <h1 className="text-2xl font-semibold">Admin access required</h1>
        <p className="mt-3 text-sm leading-6 text-muted-foreground">
          Your account can sign in, but it does not have permission to view the admin dashboard.
        </p>
      </div>
    </div>
  </div>
);

export default function AdminRoute() {
  const { user, isLoadingAuth } = useAuth();

  if (isLoadingAuth) {
    return <LoadingFallback />;
  }

  if (user?.role !== 'admin') {
    return <AccessDenied />;
  }

  return <Outlet />;
}
