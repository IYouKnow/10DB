import React, { useState } from 'react';
import { Copy, Eye, EyeOff, CheckCheck } from 'lucide-react';
import { toast } from 'sonner';

const creds = {
  host: 'localhost',
  port: '5432',
  database: 'my_saas_app',
  user: 'app_my_saas',
  password: 'xK9#mL2$vP4nQ7wR',
};

const connString = `postgresql://${creds.user}:${creds.password}@${creds.host}:${creds.port}/${creds.database}`;
const envBlock = `DB_HOST=${creds.host}
DB_PORT=${creds.port}
DB_NAME=${creds.database}
DB_USER=${creds.user}
DB_PASSWORD=${creds.password}
DATABASE_URL="${connString}"`;

function CopyButton({ value }) {
  const [copied, setCopied] = useState(false);
  const copy = () => {
    navigator.clipboard.writeText(value);
    setCopied(true);
    toast.success('Copied to clipboard');
    setTimeout(() => setCopied(false), 2000);
  };
  return (
    <button onClick={copy} className="p-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors">
      {copied ? <CheckCheck className="w-3.5 h-3.5 text-chart-3" /> : <Copy className="w-3.5 h-3.5" />}
    </button>
  );
}

function CredRow({ label, value, secret }) {
  const [revealed, setRevealed] = useState(false);
  const display = secret && !revealed ? '••••••••••••••••' : value;
  return (
    <div className="flex items-center justify-between py-3 border-b border-border last:border-0">
      <span className="text-xs font-medium text-muted-foreground w-24 shrink-0">{label}</span>
      <span className="font-mono text-sm text-foreground flex-1 truncate">{display}</span>
      <div className="flex items-center gap-1 ml-3 shrink-0">
        {secret && (
          <button
            onClick={() => setRevealed(!revealed)}
            className="p-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors"
          >
            {revealed ? <EyeOff className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
          </button>
        )}
        <CopyButton value={value} />
      </div>
    </div>
  );
}

export default function Credentials() {
  const [connRevealed, setConnRevealed] = useState(false);

  return (
    <div className="p-6 max-w-2xl mx-auto space-y-6">
      <div>
        <h1 className="text-xl font-bold text-foreground">Credentials</h1>
        <p className="text-sm text-muted-foreground mt-0.5 font-mono">my-saas-app</p>
      </div>

      {/* Connection details */}
      <div className="rounded-2xl border border-border bg-card p-5">
        <h2 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-3">Connection Details</h2>
        <CredRow label="Host" value={creds.host} />
        <CredRow label="Port" value={creds.port} />
        <CredRow label="Database" value={creds.database} />
        <CredRow label="User" value={creds.user} />
        <CredRow label="Password" value={creds.password} secret />
      </div>

      {/* Connection string */}
      <div className="rounded-2xl border border-border bg-card p-5">
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Connection String</h2>
          <div className="flex items-center gap-1">
            <button
              onClick={() => setConnRevealed(!connRevealed)}
              className="p-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors"
            >
              {connRevealed ? <EyeOff className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
            </button>
            <CopyButton value={connString} />
          </div>
        </div>
        <div className="font-mono text-xs text-muted-foreground bg-secondary/40 rounded-xl p-4 break-all">
          {connRevealed ? connString : 'postgresql://app_my_saas:••••••••@localhost:5432/my_saas_app'}
        </div>
      </div>

      {/* .env block */}
      <div className="rounded-2xl border border-border bg-card p-5">
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">.env Example</h2>
          <CopyButton value={envBlock} />
        </div>
        <pre className="font-mono text-xs text-muted-foreground bg-secondary/40 rounded-xl p-4 overflow-x-auto whitespace-pre">
          {envBlock}
        </pre>
      </div>
    </div>
  );
}