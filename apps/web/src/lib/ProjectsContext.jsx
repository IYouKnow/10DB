import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { useAuth } from './AuthContext';

const ProjectsContext = createContext(null);

function slugify(value) {
  return value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

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
    try {
      const body = await response.json();
      message = body?.message || body?.error?.message || message;
    } catch {
      // Ignore JSON parsing errors and keep the fallback message.
    }
    const error = new Error(message);
    error.status = response.status;
    throw error;
  }

  if (response.status === 204) {
    return null;
  }

  return response.json();
}

export function ProjectsProvider({ children }) {
  const { isAuthenticated, isLoadingAuth } = useAuth();
  const [projects, setProjects] = useState([]);
  const [isLoadingProjects, setIsLoadingProjects] = useState(true);
  const [projectsError, setProjectsError] = useState('');

  const loadProjects = useCallback(async () => {
    if (!isAuthenticated) {
      setProjects([]);
      setProjectsError('');
      setIsLoadingProjects(false);
      return;
    }

    setIsLoadingProjects(true);
    setProjectsError('');

    try {
      const data = await request('/api/v1/projects');
      setProjects(data.projects ?? []);
    } catch (error) {
      setProjectsError(error.message || 'Failed to load projects');
    } finally {
      setIsLoadingProjects(false);
    }
  }, [isAuthenticated]);

  useEffect(() => {
    if (isLoadingAuth) {
      return;
    }
    loadProjects();
  }, [isLoadingAuth, loadProjects]);

  const createProject = useCallback(async ({ name, description = '' }) => {
    const project = await request('/api/v1/projects', {
      method: 'POST',
      body: JSON.stringify({
        name,
        slug: slugify(name),
        description,
      }),
    });

    setProjects((current) => [...current, project]);
    return project;
  }, []);

  const getProject = useCallback(async (projectId) => {
    return request(`/api/v1/projects/${projectId}`);
  }, []);

  const provisionPostgres = useCallback(async (projectId) => {
    const project = await request(`/api/v1/projects/${projectId}/databases/postgres`, {
      method: 'POST',
    });

    setProjects((current) => current.map((entry) => (
      entry.id === project.id ? project : entry
    )));

    return project;
  }, []);

  const removeProvisionedPostgres = useCallback(async (projectId, databaseId) => {
    const project = await request(`/api/v1/projects/${projectId}/databases/${databaseId}`, {
      method: 'DELETE',
    });

    setProjects((current) => current.map((entry) => (
      entry.id === project.id ? project : entry
    )));

    return project;
  }, []);

  const deleteProject = useCallback(async (projectId) => {
    await request(`/api/v1/projects/${projectId}`, {
      method: 'DELETE',
    });

    setProjects((current) => current.filter((project) => project.id !== projectId));
  }, []);

  const value = useMemo(() => ({
    projects,
    isLoadingProjects,
    projectsError,
    loadProjects,
    createProject,
    getProject,
    provisionPostgres,
    removeProvisionedPostgres,
    deleteProject,
  }), [projects, isLoadingProjects, projectsError, loadProjects, createProject, getProject, provisionPostgres, removeProvisionedPostgres, deleteProject]);

  return (
    <ProjectsContext.Provider value={value}>
      {children}
    </ProjectsContext.Provider>
  );
}

export function useProjects() {
  const context = useContext(ProjectsContext);
  if (!context) {
    throw new Error('useProjects must be used within a ProjectsProvider');
  }
  return context;
}
