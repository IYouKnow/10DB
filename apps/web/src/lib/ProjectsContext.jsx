import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { useAuth } from './AuthContext';
import { request } from './api';

const ProjectsContext = createContext(null);

function slugify(value) {
  return value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');
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

  const provisionPostgres = useCallback(async (projectId, input = {}) => {
    const project = await request(`/api/v1/projects/${projectId}/databases/postgres`, {
      method: 'POST',
      body: JSON.stringify(input),
    });

    setProjects((current) => current.map((entry) => (
      entry.id === project.id ? project : entry
    )));

    return project;
  }, []);

  const updateProjectDatabase = useCallback(async (projectId, databaseId, input) => {
    const project = await request(`/api/v1/projects/${projectId}/databases/${databaseId}`, {
      method: 'PATCH',
      body: JSON.stringify(input),
    });

    setProjects((current) => current.map((entry) => (
      entry.id === project.id ? project : entry
    )));

    return project;
  }, []);

  const getDatabaseSchema = useCallback(async (projectId, databaseId) => {
    return request(`/api/v1/projects/${projectId}/databases/${databaseId}/schema`);
  }, []);

  const listDatabaseApiKeys = useCallback(async (databaseId) => {
    const data = await request(`/api/v1/databases/${databaseId}/api-keys`);
    return data.apiKeys ?? [];
  }, []);

  const listDraftTableColumns = useCallback(async (tableId) => {
    const data = await request(`/api/v1/tables/${tableId}/columns`);
    return data.columns ?? [];
  }, []);

  const createDraftTableColumn = useCallback(async (tableId, input) => {
    return request(`/api/v1/tables/${tableId}/columns`, {
      method: 'POST',
      body: JSON.stringify(input),
    });
  }, []);

  const updateDraftTableColumn = useCallback(async (tableId, columnId, input) => {
    return request(`/api/v1/tables/${tableId}/columns/${columnId}`, {
      method: 'PATCH',
      body: JSON.stringify(input),
    });
  }, []);

  const deleteDraftTableColumn = useCallback(async (tableId, columnId) => {
    return request(`/api/v1/tables/${tableId}/columns/${columnId}`, {
      method: 'DELETE',
    });
  }, []);

  const createDatabaseApiKey = useCallback(async (databaseId, input) => {
    return request(`/api/v1/databases/${databaseId}/api-keys`, {
      method: 'POST',
      body: JSON.stringify(input),
    });
  }, []);

  const revokeDatabaseApiKey = useCallback(async (databaseId, keyId) => {
    return request(`/api/v1/databases/${databaseId}/api-keys/${keyId}`, {
      method: 'DELETE',
    });
  }, []);

  const saveDatabaseSchema = useCallback(async (projectId, databaseId, blueprint) => {
    return request(`/api/v1/projects/${projectId}/databases/${databaseId}/schema`, {
      method: 'PUT',
      body: JSON.stringify(blueprint),
    });
  }, []);

  const listDatabaseTables = useCallback(async (projectId, databaseId) => {
    const data = await request(`/api/v1/projects/${projectId}/databases/${databaseId}/tables`);
    return data.tables ?? [];
  }, []);

  const applyDatabaseTable = useCallback(async (projectId, databaseId, tableId) => {
    return request(`/api/v1/projects/${projectId}/databases/${databaseId}/schema/tables/${tableId}/apply`, {
      method: 'POST',
    });
  }, []);

  const deleteDatabaseTable = useCallback(async (projectId, databaseId, tableId) => {
    return request(`/api/v1/projects/${projectId}/databases/${databaseId}/schema/tables/${tableId}`, {
      method: 'DELETE',
    });
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
    updateProjectDatabase,
    listDatabaseApiKeys,
    listDraftTableColumns,
    createDraftTableColumn,
    updateDraftTableColumn,
    deleteDraftTableColumn,
    createDatabaseApiKey,
    revokeDatabaseApiKey,
    getDatabaseSchema,
    saveDatabaseSchema,
    listDatabaseTables,
    applyDatabaseTable,
    deleteDatabaseTable,
    removeProvisionedPostgres,
    deleteProject,
  }), [projects, isLoadingProjects, projectsError, loadProjects, createProject, getProject, provisionPostgres, updateProjectDatabase, listDatabaseApiKeys, listDraftTableColumns, createDraftTableColumn, updateDraftTableColumn, deleteDraftTableColumn, createDatabaseApiKey, revokeDatabaseApiKey, getDatabaseSchema, saveDatabaseSchema, listDatabaseTables, applyDatabaseTable, deleteDatabaseTable, removeProvisionedPostgres, deleteProject]);

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
