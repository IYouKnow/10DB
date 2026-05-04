import React, { createContext, useContext, useEffect, useMemo, useState } from 'react';

const AuthContext = createContext(null);

async function request(path, options = {}) {
  const response = await fetch(path, {
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers ?? {}),
    },
    ...options,
  });

  if (!response.ok) {
    let message = 'Request failed';
    let code = '';
    try {
      const body = await response.json();
      message = body?.message || body?.error || body?.error?.message || message;
      code = body?.code || '';
      if (body?.details && typeof body.details === 'object') {
        const firstDetail = Object.values(body.details).find((value) => typeof value === 'string');
        if (firstDetail) {
          message = String(firstDetail);
        }
      }
    } catch {
      // Keep fallback values when the response is not JSON.
    }
    const error = new Error(message);
    error.code = code;
    throw error;
  }

  if (response.status === 204) {
    return null;
  }

  return response.json();
}

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [isLoadingAuth, setIsLoadingAuth] = useState(true);

  useEffect(() => {
    let mounted = true;

    async function loadSession() {
      try {
        const currentUser = await request('/api/v1/auth/me');
        if (mounted) {
          setUser(currentUser);
        }
      } catch {
        if (mounted) {
          setUser(null);
        }
      } finally {
        if (mounted) {
          setIsLoadingAuth(false);
        }
      }
    }

    loadSession();

    return () => {
      mounted = false;
    };
  }, []);

  const value = useMemo(() => ({
    user,
    isAuthenticated: Boolean(user),
    isLoadingAuth,
    async login(email, password) {
      const data = await request('/api/v1/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      });
      setUser(data.user);
      return data.user;
    },
    async register(name, email, password) {
      const data = await request('/api/v1/auth/register', {
        method: 'POST',
        body: JSON.stringify({ name, email, password }),
      });
      setUser(data.user);
      return data.user;
    },
    async logout() {
      await request('/api/v1/auth/logout', { method: 'POST' });
      setUser(null);
    },
  }), [user, isLoadingAuth]);

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
