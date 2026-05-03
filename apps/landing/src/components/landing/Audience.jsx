import React from 'react';
import { motion } from 'framer-motion';
import { Rocket, GraduationCap, Briefcase, Bot } from 'lucide-react';

const personas = [
  {
    icon: Rocket,
    label: 'Indie Builders',
    description: 'Ship your side project with a real database instead of stitching together free tiers.',
  },
  {
    icon: GraduationCap,
    label: 'Students',
    description: 'Learn PostgreSQL visually without wrestling with psql, roles, and pg_hba.conf.',
  },
  {
    icon: Briefcase,
    label: 'Freelancers',
    description: 'Spin up isolated databases per client. Clean separation, no extra billing.',
  },
  {
    icon: Bot,
    label: 'AI-Assisted Builders',
    description: 'Pair with your AI coding tool. Design the model visually, grab the connection string, and let AI write the code.',
  },
];

export default function Audience() {
  return (
    <section className="relative py-28 sm:py-36 overflow-hidden">
      <div className="absolute inset-0 grid-pattern pointer-events-none" />

      <div className="relative max-w-6xl mx-auto px-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5 }}
          className="text-center mb-20"
        >
          <p className="text-xs font-mono uppercase tracking-widest text-primary mb-3">Built for</p>
          <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold tracking-tight">
            People who build things
          </h2>
        </motion.div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
          {personas.map((persona, i) => (
            <motion.div
              key={persona.label}
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: i * 0.1 }}
              className="text-center group"
            >
              <div className="p-8 rounded-2xl border border-border bg-card/30 hover:bg-card/60 hover:border-primary/15 transition-all duration-300 h-full flex flex-col items-center">
                <div className="w-14 h-14 rounded-2xl bg-primary/10 border border-primary/15 flex items-center justify-center mb-5 group-hover:scale-110 transition-transform">
                  <persona.icon className="w-6 h-6 text-primary" />
                </div>
                <h3 className="font-semibold mb-2">{persona.label}</h3>
                <p className="text-sm text-muted-foreground leading-relaxed">{persona.description}</p>
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}