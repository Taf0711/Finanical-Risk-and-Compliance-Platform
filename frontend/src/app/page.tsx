'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  Activity, 
  TrendingUp, 
  Shield, 
  BarChart3, 
  Zap, 
  Eye, 
  ChevronRight,
  ArrowRight,
  Star,
  CheckCircle,
  Globe
} from 'lucide-react';

const features = [
  {
    icon: TrendingUp,
    title: 'Real-Time Trading',
    description: 'Execute trades with institutional-grade speed and precision',
    gradient: 'from-primary to-primary/80',
    delay: 0.1
  },
  {
    icon: Shield,
    title: 'Risk Management',
    description: 'Advanced VaR calculations with professional-grade analytics',
    gradient: 'from-accent to-accent/80',
    delay: 0.2
  },
  {
    icon: BarChart3,
    title: 'Portfolio Analytics',
    description: 'Comprehensive insights with Monte Carlo simulations',
    gradient: 'from-success to-success/80',
    delay: 0.3
  }
];

const stats = [
  { value: '$2.4B+', label: 'Assets Monitored', icon: Globe },
  { value: '99.9%', label: 'Uptime Guarantee', icon: CheckCircle },
  { value: '<50ms', label: 'Execution Speed', icon: Zap },
  { value: '24/7', label: 'Risk Monitoring', icon: Eye }
];

export default function LandingPage() {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(true);
  const [currentTime, setCurrentTime] = useState('');

  useEffect(() => {
    const timer = setTimeout(() => {
      setIsLoading(false);
    }, 2000);

    const timeInterval = setInterval(() => {
      setCurrentTime(new Date().toLocaleTimeString('en-US', {
        hour12: false,
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      }));
    }, 1000);

    return () => {
      clearTimeout(timer);
      clearInterval(timeInterval);
    };
  }, []);

  const handleEnterDashboard = () => {
    router.push('/dashboard');
  };

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background relative overflow-hidden">
        {/* Animated Background */}
        <div className="absolute inset-0 bg-mesh-gradient opacity-30" />
        
        {/* Loading Animation */}
        <motion.div 
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.5 }}
          className="text-center z-10"
        >
          <motion.div
            animate={{ rotate: 360 }}
            transition={{ duration: 2, repeat: Infinity, ease: "linear" }}
            className="w-16 h-16 mx-auto mb-8 rounded-2xl bg-gradient-primary flex items-center justify-center shadow-glow-primary"
          >
            <Activity className="w-8 h-8 text-white" />
          </motion.div>
          
          <motion.h1 
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3 }}
            className="text-4xl font-bold gradient-text mb-4"
          >
            FinRisk Pro
          </motion.h1>
          
          <motion.p 
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.5 }}
            className="text-foreground-muted text-lg mb-8"
          >
            Initializing Advanced Risk Analytics...
          </motion.p>
          
          {/* Loading Dots */}
          <div className="flex items-center justify-center gap-2">
            {[0, 1, 2].map((i) => (
              <motion.div
                key={i}
                animate={{ 
                  scale: [1, 1.2, 1],
                  opacity: [0.5, 1, 0.5]
                }}
                transition={{
                  duration: 1.5,
                  repeat: Infinity,
                  delay: i * 0.2
                }}
                className="w-3 h-3 bg-primary rounded-full"
              />
            ))}
          </div>
        </motion.div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background relative overflow-hidden">
      {/* Animated Background Grid */}
      <div className="absolute inset-0 grid-pattern opacity-20" />
      <div className="absolute inset-0 bg-mesh-gradient opacity-40" />
      
      {/* Header */}
      <motion.header 
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        className="relative z-10 p-6 flex items-center justify-between"
      >
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-primary flex items-center justify-center shadow-glow-primary">
            <Activity className="w-6 h-6 text-white" />
          </div>
          <div>
            <h1 className="text-xl font-bold gradient-text">FinRisk Pro</h1>
            <p className="text-2xs text-foreground-muted">Financial Analytics Platform</p>
          </div>
        </div>
        
        <div className="flex items-center gap-4">
          <div className="hidden md:flex items-center gap-2 px-3 py-1 glass-card rounded-full">
            <div className="w-2 h-2 bg-success rounded-full animate-pulse" />
            <span className="text-sm text-success font-medium">LIVE</span>
            <span className="text-sm text-foreground-muted">{currentTime}</span>
          </div>
          
          <motion.button
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            onClick={handleEnterDashboard}
            className="btn-primary flex items-center gap-2"
          >
            <span>Enter Platform</span>
            <ArrowRight className="w-4 h-4" />
          </motion.button>
        </div>
      </motion.header>

      {/* Main Content */}
      <main className="relative z-10 pt-20 pb-32">
        <div className="container mx-auto px-6 text-center">
          {/* Hero Section */}
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8 }}
            className="mb-20"
          >
            <motion.div
              animate={{ 
                boxShadow: [
                  '0 0 20px rgba(0, 212, 255, 0.2)',
                  '0 0 40px rgba(0, 212, 255, 0.4)',
                  '0 0 20px rgba(0, 212, 255, 0.2)'
                ]
              }}
              transition={{ duration: 3, repeat: Infinity }}
              className="inline-flex items-center gap-2 px-4 py-2 glass-card rounded-full mb-8 border border-primary/20"
            >
              <Star className="w-4 h-4 text-primary" />
              <span className="text-sm text-foreground-muted">Institutional-Grade Risk Analytics</span>
            </motion.div>
            
            <h1 className="text-5xl md:text-7xl font-bold mb-6 leading-tight">
              <span className="gradient-text">Advanced Financial</span>
              <br />
              <span className="text-foreground">Risk Monitoring</span>
            </h1>
            
            <p className="text-xl text-foreground-muted max-w-3xl mx-auto mb-10 leading-relaxed">
              Professional-grade risk management platform with real-time analytics, 
              Monte Carlo simulations, and institutional-level compliance monitoring.
            </p>
            
            <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
              <motion.button
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                onClick={handleEnterDashboard}
                className="btn-primary flex items-center gap-2 text-lg px-8 py-4"
              >
                <span>Launch Dashboard</span>
                <ChevronRight className="w-5 h-5" />
              </motion.button>
              
              <motion.button
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                className="btn-secondary flex items-center gap-2 text-lg px-8 py-4"
              >
                <Eye className="w-5 h-5" />
                <span>View Demo</span>
              </motion.button>
            </div>
          </motion.div>

          {/* Stats Section */}
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.2 }}
            className="grid grid-cols-2 md:grid-cols-4 gap-6 mb-20"
          >
            {stats.map((stat, index) => {
              const Icon = stat.icon;
              return (
                <motion.div
                  key={stat.label}
                  initial={{ opacity: 0, scale: 0.8 }}
                  animate={{ opacity: 1, scale: 1 }}
                  transition={{ delay: 0.3 + index * 0.1 }}
                  className="glass-card p-6 hover-lift"
                >
                  <Icon className="w-8 h-8 text-primary mx-auto mb-4" />
                  <div className="text-3xl font-bold gradient-text mb-2">{stat.value}</div>
                  <div className="text-sm text-foreground-muted">{stat.label}</div>
                </motion.div>
              );
            })}
          </motion.div>

          {/* Features Section */}
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.4 }}
            className="grid md:grid-cols-3 gap-8 mb-20"
          >
            {features.map((feature, index) => {
              const Icon = feature.icon;
              return (
                <motion.div
                  key={feature.title}
                  initial={{ opacity: 0, y: 30 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: feature.delay }}
                  whileHover={{ y: -10 }}
                  className="glass-card p-8 hover-lift relative overflow-hidden group"
                >
                  <div className={`absolute top-0 left-0 w-full h-1 bg-gradient-to-r ${feature.gradient}`} />
                  
                  <div className={`w-16 h-16 mx-auto mb-6 rounded-2xl bg-gradient-to-r ${feature.gradient} flex items-center justify-center shadow-glow-primary group-hover:shadow-glow-primary/50 transition-all duration-300`}>
                    <Icon className="w-8 h-8 text-white" />
                  </div>
                  
                  <h3 className="text-xl font-bold text-foreground mb-4">{feature.title}</h3>
                  <p className="text-foreground-muted leading-relaxed">{feature.description}</p>
                </motion.div>
              );
            })}
          </motion.div>

          {/* CTA Section */}
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.6 }}
            className="glass-card p-12 relative overflow-hidden"
          >
            <div className="absolute inset-0 bg-gradient-to-r from-primary/5 to-accent/5" />
            <div className="relative z-10">
              <h2 className="text-3xl font-bold text-foreground mb-4">
                Ready to Transform Your Risk Management?
              </h2>
              <p className="text-foreground-muted mb-8 text-lg max-w-2xl mx-auto">
                Join institutional investors using our advanced analytics platform 
                for professional-grade risk monitoring and portfolio optimization.
              </p>
              
              <motion.button
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                onClick={handleEnterDashboard}
                className="btn-primary text-lg px-10 py-4 shadow-glow-primary"
              >
                Get Started Now
              </motion.button>
            </div>
          </motion.div>
        </div>
      </main>

      {/* Footer */}
      <motion.footer
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.8 }}
        className="relative z-10 border-t border-white/10 py-8"
      >
        <div className="container mx-auto px-6 text-center">
          <p className="text-foreground-muted text-sm">
            © 2024 FinRisk Pro. Professional Financial Risk Analytics Platform.
          </p>
        </div>
      </motion.footer>
    </div>
  );
}