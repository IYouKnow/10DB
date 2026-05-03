import React from 'react';
import { motion } from 'framer-motion';
import { Monitor, Server, Database, HardDrive } from 'lucide-react';

const layers = [
  {
    icon: Monitor,
    label: 'React Frontend',
    detail: 'Schema board, project dashboard, connection manager',
    color: 'border-primary/30 bg-primary/5',
    iconColor: 'text-primary',
  },
  {
    icon: Server,
    label: 'Go Backend',
    detail: 'API server, SQL generation, credential management',
    color: 'border-accent/30 bg-accent/5',
    iconColor: 'text-accent',
  },
  {
    icon: HardDrive,
    label: 'SQLite Control DB',
    detail: 'Project metadata, schema blueprints, user sessions',
    color: 'border-chart-3/30 bg-chart-3/5',
    iconColor: 'text-chart-3',
  },
  {
    icon: Database,
    label: 'PostgreSQL Server',
    detail: 'Shared instance with per-project isolated databases',
    color: 'border-chart-4/30 bg-chart-4/5',
    iconColor: 'text-chart-4',
  },
];

export default function Architecture() {
  return (
    <section id="architecture" className="relative py-28 sm:py-36">
      <div className="max-w-6xl mx-auto px-6">
        <div className="grid lg:grid-cols-2 gap-16 items-center">
          <motion.div
            initial={{ opacity: 0, x: -30 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6 }}
          >
            <p className="text-xs font-mono uppercase tracking-widest text-primary mb-3">Architecture</p>
            <h2 className="text-3xl sm:text-4xl font-bold tracking-tight mb-6">
              Simple stack,
              <br />
              serious power
            </h2>
            <p className="text-muted-foreground leading-relaxed mb-8 max-w-md">
              A lean architecture built for clarity. The Go backend is the gatekeeper — it generates SQL, manages credentials, and talks to PostgreSQL so the frontend never has to.
            </p>

            <div className="font-mono text-sm bg-card border border-border rounded-xl p-5 max-w-md">
              <div className="text-muted-foreground mb-1">
                <span className="text-primary">$</span> docker compose up -d
              </div>
              <div className="text-muted-foreground mb-1">
                <span className="text-chart-3">✓</span> PostgreSQL server ready
              </div>
              <div className="text-muted-foreground mb-1">
                <span className="text-chart-3">✓</span> Go backend listening on :8080
              </div>
              <div className="text-muted-foreground">
                <span className="text-chart-3">✓</span> Dashboard at http://localhost:3000
              </div>
            </div>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, x: 30 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6, delay: 0.15 }}
            className="space-y-4"
          >
            {layers.map((layer, i) => (
              <motion.div
                key={layer.label}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.4, delay: 0.2 + i * 0.1 }}
                className={`flex items-center gap-5 p-5 rounded-2xl border ${layer.color}`}
              >
                <div className="w-12 h-12 rounded-xl bg-card border border-border flex items-center justify-center shrink-0">
                  <layer.icon className={`w-5 h-5 ${layer.iconColor}`} />
                </div>
                <div>
                  <p className="font-semibold text-sm">{layer.label}</p>
                  <p className="text-xs text-muted-foreground mt-0.5">{layer.detail}</p>
                </div>
              </motion.div>
            ))}

            {/* Connection arrows between layers */}
            <div className="flex flex-col items-center gap-0 -mt-3">
              {[0, 1, 2].map((i) => (
                <div key={i} className="w-px h-0 bg-border" />
              ))}
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  );
}