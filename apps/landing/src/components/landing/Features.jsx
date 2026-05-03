import React from 'react';
import { motion } from 'framer-motion';
import {
  Eye,
  ShieldCheck,
  Box,
  RefreshCcw,
  Code2,
  Server,
} from 'lucide-react';

const features = [
  {
    icon: Eye,
    title: 'Visual Schema Board',
    description: 'Design tables, columns, types, and foreign keys on a drag-and-drop canvas. See your data model at a glance.',
  },
  {
    icon: ShieldCheck,
    title: 'Admin Creds Stay Server-Side',
    description: 'PostgreSQL superuser credentials never touch the browser. The Go backend handles all privileged operations.',
  },
  {
    icon: Box,
    title: 'Isolated Databases',
    description: 'Every project gets its own PostgreSQL database and database user. Full isolation with zero configuration.',
  },
  {
    icon: RefreshCcw,
    title: 'One-Click Migrations',
    description: 'Edit your schema visually and apply changes. 10DB generates the DDL diff and runs it against your database.',
  },
  {
    icon: Code2,
    title: 'Instant Connection Strings',
    description: 'Get a ready-to-paste connection string for every project. Works with any PostgreSQL client, ORM, or framework.',
  },
  {
    icon: Server,
    title: 'Fully Self-Hosted',
    description: 'Run it on your VPS, homelab, or cloud instance. Your data stays on your hardware, always.',
  },
];

export default function Features() {
  return (
    <section id="features" className="relative py-28 sm:py-36 overflow-hidden">
      <div className="absolute inset-0 grid-pattern pointer-events-none" />
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-primary/4 rounded-full blur-3xl pointer-events-none" />

      <div className="relative max-w-6xl mx-auto px-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5 }}
          className="text-center mb-20"
        >
          <p className="text-xs font-mono uppercase tracking-widest text-primary mb-3">Features</p>
          <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold tracking-tight mb-4">
            PostgreSQL without the ceremony
          </h2>
          <p className="text-muted-foreground text-lg max-w-xl mx-auto">
            Everything you need to go from idea to database, minus the infrastructure headaches.
          </p>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {features.map((feature, i) => (
            <motion.div
              key={feature.title}
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: i * 0.08 }}
              className="group"
            >
              <div className="p-6 rounded-2xl border border-border bg-card/30 hover:bg-card/60 hover:border-primary/15 transition-all duration-300 h-full">
                <div className="w-10 h-10 rounded-xl bg-primary/10 border border-primary/15 flex items-center justify-center mb-5 group-hover:bg-primary/20 transition-colors">
                  <feature.icon className="w-5 h-5 text-primary" />
                </div>
                <h3 className="text-base font-semibold mb-2">{feature.title}</h3>
                <p className="text-sm text-muted-foreground leading-relaxed">{feature.description}</p>
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}