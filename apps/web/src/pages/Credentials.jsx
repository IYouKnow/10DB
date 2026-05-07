import React, { useEffect, useMemo, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { ArrowLeft, CheckCheck, Copy, Eye, EyeOff, KeyRound } from 'lucide-react';
import { toast } from 'sonner';
import { useProjects } from '@/lib/ProjectsContext';

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

function CredRow({ label, value, secret = false }) {
  const [revealed, setRevealed] = useState(false);
  const display = secret && !revealed ? '****************' : value;

  return (
    <div className="flex items-center gap-3 border-b border-border py-3 last:border-0">
      <div className="w-28 shrink-0 text-xs font-medium text-muted-foreground">{label}</div>
      <div className="flex-1 truncate font-mono text-sm text-foreground">{display}</div>
      <div className="flex shrink-0 items-center gap-1">
        {secret ? (
          <button
            type="button"
            onClick={() => setRevealed((current) => !current)}
            className="rounded-lg p-1.5 text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
            aria-label={revealed ? 'Hide secret' : 'Show secret'}
          >
            {revealed ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
          </button>
        ) : null}
        <CopyButton value={value} />
      </div>
    </div>
  );
}

export default function Credentials() {
  const { projectId, databaseId } = useParams();
  const { getDatabaseCredentials, getProject } = useProjects();
  const [credentials, setCredentials] = useState(null);
  const [project, setProject] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [urlVisible, setUrlVisible] = useState(false);

  useEffect(() => {
    let mounted = true;

    async function load() {
      setIsLoading(true);
      try {
        const [credentialsData, projectData] = await Promise.all([
          getDatabaseCredentials(databaseId),
          getProject(projectId),
        ]);
        if (!mounted) {
          return;
        }
        setCredentials(credentialsData);
        setProject(projectData);
      } catch (error) {
        if (mounted) {
          toast.error(error.message || 'Failed to load credentials');
        }
      } finally {
        if (mounted) {
          setIsLoading(false);
        }
      }
    }

    load();

    return () => {
      mounted = false;
    };
  }, [databaseId, getDatabaseCredentials, getProject, projectId]);

  const database = useMemo(
    () => project?.databases?.find((entry) => entry.id === databaseId) ?? null,
    [databaseId, project],
  );

  return (
    <div className="mx-auto flex max-w-3xl flex-col gap-6 p-6">
      <div className="flex items-center gap-3">
        <Link
          to={`/projects/${projectId}/databases/${databaseId}/schema`}
          className="inline-flex h-8 items-center gap-1.5 rounded-lg border border-border px-2.5 text-xs text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
        >
          <ArrowLeft className="h-3.5 w-3.5" />
          Database Board
        </Link>
        <div>
          <h1 className="text-xl font-bold text-foreground">Credentials</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            {database?.name ?? 'Database'} - <span className="font-mono">{credentials?.database ?? databaseId}</span>
          </p>
        </div>
      </div>

      {isLoading ? (
        <div className="rounded-2xl border border-border bg-card p-6 text-sm text-muted-foreground">
          Loading credentials...
        </div>
      ) : credentials ? (
        <>
          <div className="rounded-2xl border border-border bg-card p-5">
            <div className="mb-3 flex items-center gap-2">
              <div className="flex h-8 w-8 items-center justify-center rounded-xl border border-primary/20 bg-primary/10">
                <KeyRound className="h-4 w-4 text-primary" />
              </div>
              <div>
                <h2 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Main Credentials</h2>
                <p className="text-xs text-muted-foreground">One active username/password per database.</p>
              </div>
            </div>
            <CredRow label="Username" value={credentials.username} />
            <CredRow label="Password" value={credentials.password} secret />
            <CredRow label="Database" value={credentials.database} />
            <CredRow label="Host" value={credentials.host} />
            <CredRow label="Port" value={String(credentials.port)} />
            <CredRow label="SSL Mode" value={credentials.sslMode} />
          </div>

          <div className="rounded-2xl border border-border bg-card p-5">
            <div className="mb-3 flex items-center justify-between gap-3">
              <div>
                <h2 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">DATABASE_URL</h2>
                <p className="text-xs text-muted-foreground">Use this directly in your app config.</p>
              </div>
              <div className="flex items-center gap-1">
                <button
                  type="button"
                  onClick={() => setUrlVisible((current) => !current)}
                  className="rounded-lg p-1.5 text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                  aria-label={urlVisible ? 'Hide DATABASE_URL' : 'Show DATABASE_URL'}
                >
                  {urlVisible ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
                </button>
                <CopyButton value={credentials.databaseUrl} />
              </div>
            </div>
            <div className="break-all rounded-xl bg-secondary/40 p-4 font-mono text-xs text-muted-foreground">
              {urlVisible ? credentials.databaseUrl : `postgres://${credentials.username}:********@${credentials.host}:${credentials.port}/${credentials.database}?sslmode=${credentials.sslMode}`}
            </div>
          </div>
        </>
      ) : (
        <div className="rounded-2xl border border-border bg-card p-6 text-sm text-muted-foreground">
          Credentials are not available for this database yet.
        </div>
      )}
    </div>
  );
}
