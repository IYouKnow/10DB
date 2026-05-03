import React from 'react';
import { motion } from 'framer-motion';
import { ArrowRight, Database } from 'lucide-react';
import { Button } from '@/components/ui/button';

export default function CTA() {
  return (
    <section className="relative py-28 sm:py-36">
      <div className="max-w-6xl mx-auto px-6">
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="relative rounded-3xl border border-primary/20 bg-gradient-to-br from-primary/5 via-card to-accent/5 p-12 sm:p-16 text-center overflow-hidden"
        >
          <div className="absolute top-0 left-1/2 -translate-x-1/2 w-80 h-40 bg-primary/10 rounded-full blur-3xl pointer-events-none" />

          <div className="relative">
            <div className="w-16 h-16 rounded-2xl bg-primary/10 border border-primary/20 flex items-center justify-center mx-auto mb-8">
              <Database className="w-7 h-7 text-primary" />
            </div>

            <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold tracking-tight mb-4">
              Stop configuring.
              <br />
              <span className="text-primary">Start building.</span>
            </h2>

            <p className="text-muted-foreground text-lg max-w-lg mx-auto mb-10">
              Get 10DB Launch running on your server in under two minutes. Design your first schema and have a production-ready PostgreSQL database.
            </p>

            <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
              <Button size="lg" className="bg-primary text-primary-foreground hover:bg-primary/90 font-semibold px-10 h-13 text-base glow-primary">
                Get Started Free
                <ArrowRight className="w-4 h-4 ml-2" />
              </Button>
              <Button
                size="lg"
                variant="ghost"
                className="text-muted-foreground hover:text-foreground font-medium"
              >
                Read the Docs
              </Button>
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  );
}