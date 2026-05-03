import React from 'react';
import { motion } from 'framer-motion';
import { FolderPlus, LayoutGrid, Zap, Cable } from 'lucide-react';

const steps = [
  {
    icon: FolderPlus,
    number: '01',
    title: 'Create a project',
    description: 'Name it, and 10DB provisions an isolated PostgreSQL database and user behind the scenes.',
  },
  {
    icon: LayoutGrid,
    number: '02',
    title: 'Design your schema',
    description: 'Drag tables onto the visual board. Add columns, set types, define relationships — all visually.',
  },
  {
    icon: Zap,
    number: '03',
    title: 'Apply with one click',
    description: 'The backend generates proper PostgreSQL DDL, runs migrations, and handles everything safely.',
  },
  {
    icon: Cable,
    number: '04',
    title: 'Grab your connection string',
    description: 'Copy a ready-to-use connection string. Plug it into your app, ORM, or any PostgreSQL client.',
  },
];

export default function HowItWorks() {
  return (
    <section id="how-it-works" className="relative py-28 sm:py-36">
      <div className="max-w-6xl mx-auto px-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5 }}
          className="text-center mb-20"
        >
          <p className="text-xs font-mono uppercase tracking-widest text-primary mb-3">Workflow</p>
          <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold tracking-tight">
            Four steps. Zero friction.
          </h2>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {steps.map((step, i) => (
            <motion.div
              key={step.number}
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: i * 0.1 }}
              className="group relative"
            >
              <div className="p-6 rounded-2xl border border-border bg-card/50 hover:bg-card hover:border-primary/20 transition-all duration-300 h-full">
                <div className="flex items-start justify-between mb-6">
                  <div className="w-11 h-11 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center group-hover:bg-primary/20 transition-colors">
                    <step.icon className="w-5 h-5 text-primary" />
                  </div>
                  <span className="font-mono text-xs text-muted-foreground/40">{step.number}</span>
                </div>
                <h3 className="text-base font-semibold mb-2">{step.title}</h3>
                <p className="text-sm text-muted-foreground leading-relaxed">{step.description}</p>
              </div>

              {i < steps.length - 1 && (
                <div className="hidden lg:block absolute top-1/2 -right-3 w-6 h-px bg-border" />
              )}
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}