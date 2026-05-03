import React from 'react';
import Navbar from './components/landing/Navbar';
import Hero from './components/landing/Hero';
import HowItWorks from './components/landing/HowItWorks';
import Features from './components/landing/Features';
import Audience from './components/landing/Audience';
import Architecture from './components/landing/Architecture';
import CTA from './components/landing/CTA';
import Footer from './components/landing/Footer';

export default function Home() {
  return (
    <div className="min-h-screen bg-background text-foreground font-sans antialiased">
      <Navbar />
      <Hero />
      <HowItWorks />
      <Features />
      <Audience />
      <Architecture />
      <CTA />
      <Footer />
    </div>
  );
}
