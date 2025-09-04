'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { motion } from 'framer-motion';
import Link from 'next/link';
import { 
  ChartBarIcon, 
  ShieldCheckIcon, 
  BoltIcon,
  ArrowRightIcon
} from '@heroicons/react/24/outline';
import { useAuth } from '@/hooks/useAuth';

export default function HomePage() {
  const { isAuthenticated } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (isAuthenticated) {
      router.push('/dashboard');
    }
  }, [isAuthenticated, router]);

  if (isAuthenticated) {
    return (
      <div className="min-h-screen bg-slate-900 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-blue-900 to-slate-900">
      {/* Background Pattern */}
      <div className="absolute inset-0 bg-[url('/grid.svg')] bg-center [mask-image:linear-gradient(180deg,white,rgba(255,255,255,0))]" />
      
      <div className="relative">
        {/* Navigation */}
        <nav className="flex items-center justify-between p-6">
          <div className="flex items-center">
            <div className="w-10 h-10 bg-gradient-to-r from-blue-500 to-purple-600 rounded-xl flex items-center justify-center">
              <ChartBarIcon className="w-6 h-6 text-white" />
            </div>
            <span className="ml-3 text-2xl font-bold text-white">RiskMonitor</span>
          </div>
          <div className="flex items-center space-x-4">
            <Link 
              href="/login"
              className="text-blue-100 hover:text-white transition-colors"
            >
              Sign In
            </Link>
            <Link 
              href="/register"
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
            >
              Get Started
            </Link>
          </div>
        </nav>

        {/* Hero Section */}
        <div className="max-w-7xl mx-auto px-6 pt-20 pb-32">
          <div className="text-center">
            <motion.h1
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="text-5xl md:text-7xl font-bold text-white mb-6"
            >
              Advanced Portfolio
              <span className="bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
                {' '}Risk Management
              </span>
            </motion.h1>
            
            <motion.p
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
              className="text-xl text-blue-100 mb-8 max-w-3xl mx-auto"
            >
              Monitor, analyze, and manage portfolio risk in real-time with advanced 
              Monte Carlo simulations, VaR calculations, and compliance monitoring.
            </motion.p>

            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.4 }}
              className="flex flex-col sm:flex-row gap-4 justify-center"
            >
              <Link 
                href="/register"
                className="inline-flex items-center px-8 py-4 bg-gradient-to-r from-blue-600 to-purple-600 text-white font-semibold rounded-lg hover:shadow-xl transition-all duration-200"
              >
                Start Free Trial
                <ArrowRightIcon className="ml-2 w-5 h-5" />
              </Link>
              <Link 
                href="/demo"
                className="inline-flex items-center px-8 py-4 bg-white/10 backdrop-blur-sm border border-white/20 text-white font-semibold rounded-lg hover:bg-white/20 transition-all duration-200"
              >
                View Demo
              </Link>
            </motion.div>
          </div>

          {/* Features */}
          <motion.div
            initial={{ opacity: 0, y: 40 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.6 }}
            className="grid md:grid-cols-3 gap-8 mt-24"
          >
            <div className="bg-white/5 backdrop-blur-sm border border-white/10 rounded-2xl p-8 text-center">
              <div className="w-16 h-16 bg-gradient-to-r from-blue-500 to-blue-600 rounded-2xl flex items-center justify-center mx-auto mb-4">
                <ChartBarIcon className="w-8 h-8 text-white" />
              </div>
              <h3 className="text-xl font-semibold text-white mb-2">Real-Time Analytics</h3>
              <p className="text-blue-100/80">
                Monitor portfolio performance and risk metrics with live data updates and interactive charts.
              </p>
            </div>

            <div className="bg-white/5 backdrop-blur-sm border border-white/10 rounded-2xl p-8 text-center">
              <div className="w-16 h-16 bg-gradient-to-r from-purple-500 to-purple-600 rounded-2xl flex items-center justify-center mx-auto mb-4">
                <ShieldCheckIcon className="w-8 h-8 text-white" />
              </div>
              <h3 className="text-xl font-semibold text-white mb-2">Risk Management</h3>
              <p className="text-blue-100/80">
                Advanced VaR calculations, stress testing, and Monte Carlo simulations for comprehensive risk analysis.
              </p>
            </div>

            <div className="bg-white/5 backdrop-blur-sm border border-white/10 rounded-2xl p-8 text-center">
              <div className="w-16 h-16 bg-gradient-to-r from-green-500 to-green-600 rounded-2xl flex items-center justify-center mx-auto mb-4">
                <BoltIcon className="w-8 h-8 text-white" />
              </div>
              <h3 className="text-xl font-semibold text-white mb-2">Compliance Ready</h3>
              <p className="text-blue-100/80">
                Built-in KYC/AML checks, regulatory reporting, and automated compliance monitoring.
              </p>
            </div>
          </motion.div>
        </div>

        {/* Footer */}
        <div className="border-t border-white/10 py-8">
          <div className="max-w-7xl mx-auto px-6 text-center">
            <p className="text-blue-100/60">
              © 2024 Financial Risk Monitor. All rights reserved.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}