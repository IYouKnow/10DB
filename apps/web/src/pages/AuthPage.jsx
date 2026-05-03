import React, { useState } from 'react';
import { Database, LoaderCircle } from 'lucide-react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '@/lib/AuthContext';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export default function AuthPage() {
  const location = useLocation();
  const { isAuthenticated, isLoadingAuth, login, register } = useAuth();
  const [mode, setMode] = useState('login');
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  if (!isLoadingAuth && isAuthenticated) {
    const destination = location.state?.from?.pathname || '/';
    return <Navigate to={destination} replace />;
  }

  const submit = async (event) => {
    event.preventDefault();
    setError('');
    setIsSubmitting(true);

    try {
      if (mode === 'login') {
        await login(email, password);
      } else {
        await register(name, email, password);
      }
    } catch (submitError) {
      setError(submitError.message || 'Unable to continue');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-background text-foreground">
      <div className="mx-auto flex min-h-screen max-w-6xl flex-col px-6 py-8 lg:flex-row lg:items-center lg:gap-16">
        <div className="flex-1 py-10">
          <div className="mb-6 flex items-center gap-2.5">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl border border-primary/20 bg-primary/10">
              <Database className="h-4 w-4 text-primary" />
            </div>
            <span className="text-sm font-semibold tracking-tight">
              10DB <span className="text-primary">Launch</span>
            </span>
          </div>

          <h1 className="max-w-xl text-4xl font-bold tracking-tight">Sign in to manage your database projects.</h1>
          <p className="mt-4 max-w-lg text-sm leading-6 text-muted-foreground">
            Create an account to keep projects private to each user, then jump straight into schema design and credentials for the databases you own.
          </p>
        </div>

        <div className="w-full max-w-md">
          <div className="rounded-3xl border border-border bg-card p-6 shadow-2xl shadow-black/20">
            <div className="mb-6 grid grid-cols-2 rounded-xl bg-secondary/50 p-1 text-sm">
              <button
                onClick={() => setMode('login')}
                className={mode === 'login' ? 'rounded-lg bg-background px-3 py-2 font-medium text-foreground' : 'px-3 py-2 text-muted-foreground'}
              >
                Login
              </button>
              <button
                onClick={() => setMode('register')}
                className={mode === 'register' ? 'rounded-lg bg-background px-3 py-2 font-medium text-foreground' : 'px-3 py-2 text-muted-foreground'}
              >
                Create account
              </button>
            </div>

            <form className="space-y-4" onSubmit={submit}>
              {mode === 'register' && (
                <div>
                  <label className="mb-1.5 block text-xs font-medium text-muted-foreground">Full name</label>
                  <Input value={name} onChange={(event) => setName(event.target.value)} placeholder="Pedro Silva" />
                </div>
              )}

              <div>
                <label className="mb-1.5 block text-xs font-medium text-muted-foreground">Email</label>
                <Input
                  type="email"
                  value={email}
                  onChange={(event) => setEmail(event.target.value)}
                  placeholder="you@example.com"
                />
              </div>

              <div>
                <label className="mb-1.5 block text-xs font-medium text-muted-foreground">Password</label>
                <Input
                  type="password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  placeholder={mode === 'login' ? 'Enter your password' : 'At least 8 characters'}
                />
              </div>

              {error && (
                <div className="rounded-xl border border-destructive/20 bg-destructive/5 px-3 py-2 text-sm text-destructive">
                  {error}
                </div>
              )}

              <Button
                type="submit"
                className="w-full bg-primary text-primary-foreground hover:bg-primary/90"
                disabled={isSubmitting}
              >
                {isSubmitting ? (
                  <>
                    <LoaderCircle className="mr-2 h-4 w-4 animate-spin" />
                    {mode === 'login' ? 'Signing in...' : 'Creating account...'}
                  </>
                ) : (
                  mode === 'login' ? 'Login' : 'Create account'
                )}
              </Button>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}
