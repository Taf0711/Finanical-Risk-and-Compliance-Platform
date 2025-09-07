'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { motion, AnimatePresence } from 'framer-motion';
import { cn } from '@/lib/utils';
import {
  LayoutDashboard,
  TrendingUp,
  Shield,
  Bell,
  Settings,
  User,
  LogOut,
  Menu,
  X,
  Activity,
  PieChart,
  BarChart3,
  Wallet,
  Search,
  ChevronDown,
  Plus,
  Minus,
  DollarSign,
  AlertTriangle,
  CheckCircle,
  Clock
} from 'lucide-react';

interface DashboardLayoutProps {
  children: React.ReactNode;
}

const navigation = [
  {
    name: 'Dashboard',
    href: '/dashboard',
    icon: LayoutDashboard,
    description: 'Overview & Analytics'
  },
  {
    name: 'Portfolio',
    href: '/portfolio',
    icon: PieChart,
    description: 'Manage Holdings'
  },
  {
    name: 'Trading',
    href: '/trading',
    icon: TrendingUp,
    description: 'Execute Trades'
  },
  {
    name: 'Risk Analysis',
    href: '/risk',
    icon: Shield,
    description: 'Risk Metrics'
  },
  {
    name: 'Alerts',
    href: '/alerts',
    icon: Bell,
    description: 'Notifications'
  },
  {
    name: 'Reports',
    href: '/reports',
    icon: BarChart3,
    description: 'Analytics'
  }
];

export default function DashboardLayout({ children }: DashboardLayoutProps) {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [darkMode, setDarkMode] = useState(true); // Default to dark mode for black theme
  const pathname = usePathname();

  useEffect(() => {
    // Set dark mode by default
    document.documentElement.classList.add('dark');
  }, []);

  const toggleDarkMode = () => {
    setDarkMode(!darkMode);
    document.documentElement.classList.toggle('dark');
  };

  return (
    <div className="min-h-screen bg-gray-900">
      {/* Mobile sidebar overlay */}
      <AnimatePresence>
        {sidebarOpen && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-50 lg:hidden"
          >
            <div 
              className="absolute inset-0 bg-black/50 backdrop-blur-sm"
              onClick={() => setSidebarOpen(false)}
            />
          </motion.div>
        )}
      </AnimatePresence>

      {/* Sidebar */}
      <motion.aside
        initial={false}
        animate={{
          x: sidebarOpen ? 0 : -280,
        }}
        className={cn(
          "fixed top-0 left-0 z-50 h-full bg-black-primary border-r border-red-primary/20 shadow-2xl transition-all duration-300 ease-in-out",
          "w-70 lg:translate-x-0 lg:static lg:inset-0"
        )}
      >
        <div className="flex flex-col h-full">
          {/* Header */}
          <div className="flex items-center justify-between p-6 border-b border-red-primary/20">
            <div className="flex items-center gap-3">
              <div className="p-2 gradient-red-black rounded-xl">
                <Activity className="w-6 h-6 text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold text-white">
                  RiskMonitor
                </h1>
                <p className="text-xs text-gray-400">Financial Analytics</p>
              </div>
            </div>
            <button
              onClick={() => setSidebarOpen(false)}
              className="lg:hidden p-2 rounded-lg hover:bg-red-primary/10 transition-colors text-gray-400 hover:text-white"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* User Profile */}
          <div className="p-6 border-b border-red-primary/20">
            <div className="p-4 glass-red rounded-xl">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 gradient-red-black rounded-full flex items-center justify-center">
                  <User className="w-5 h-5 text-white" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-white truncate">
                    Demo Trader
                  </p>
                  <p className="text-xs text-gray-300 truncate">
                    Premium Account
                  </p>
                </div>
                <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
              </div>
            </div>
          </div>

          {/* Navigation */}
          <nav className="flex-1 p-6 space-y-2 overflow-y-auto">
            {navigation.map((item) => {
              const isActive = pathname === item.href;
              return (
                <Link key={item.name} href={item.href}>
                  <motion.div
                    whileHover={{ x: 4 }}
                    whileTap={{ scale: 0.98 }}
                    className={cn(
                      "group flex items-center gap-3 p-3 rounded-xl transition-all duration-200",
                      isActive
                        ? "bg-red-primary text-white shadow-lg"
                        : "hover:bg-red-primary/10 text-gray-300 hover:text-white"
                    )}
                  >
                    <div className={cn(
                      "p-2 rounded-lg transition-colors",
                      isActive 
                        ? "bg-white/20" 
                        : "bg-gray-800 group-hover:bg-red-primary/20"
                    )}>
                      <item.icon className="w-4 h-4" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium truncate">{item.name}</p>
                      <p className={cn(
                        "text-xs truncate",
                        isActive ? "text-white/80" : "text-gray-400"
                      )}>
                        {item.description}
                      </p>
                    </div>
                  </motion.div>
                </Link>
              );
            })}
          </nav>

          {/* Quick Actions */}
          <div className="p-6 border-t border-red-primary/20">
            <h3 className="text-sm font-medium text-white mb-3">Quick Actions</h3>
            <div className="grid grid-cols-2 gap-2">
              <motion.button
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                className="flex flex-col items-center gap-2 p-3 rounded-lg bg-gray-800 hover:bg-red-primary/20 transition-colors"
              >
                <Plus className="w-4 h-4 text-green-400" />
                <span className="text-xs font-medium text-gray-300">Buy</span>
              </motion.button>
              <motion.button
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                className="flex flex-col items-center gap-2 p-3 rounded-lg bg-gray-800 hover:bg-red-primary/20 transition-colors"
              >
                <Minus className="w-4 h-4 text-red-400" />
                <span className="text-xs font-medium text-gray-300">Sell</span>
              </motion.button>
            </div>
          </div>

          {/* Footer */}
          <div className="p-6 border-t border-red-primary/20">
            <div className="flex items-center justify-between">
              <motion.button
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                className="p-2 rounded-lg bg-gray-800 hover:bg-red-primary/20 transition-colors text-gray-400 hover:text-white"
              >
                <Settings className="w-4 h-4" />
              </motion.button>
              <motion.button
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium text-red-400 hover:bg-red-primary/10 transition-colors"
              >
                <LogOut className="w-4 h-4" />
                <span>Logout</span>
              </motion.button>
            </div>
          </div>
        </div>
      </motion.aside>

      {/* Main content */}
      <div className="lg:pl-70 transition-all duration-300">
        {/* Top bar */}
        <header className="sticky top-0 z-40 bg-black-primary/80 backdrop-blur-xl border-b border-red-primary/20">
          <div className="flex items-center justify-between px-6 py-4">
            <div className="flex items-center gap-4">
              <button
                onClick={() => setSidebarOpen(true)}
                className="lg:hidden p-2 rounded-lg hover:bg-red-primary/10 transition-colors text-gray-400 hover:text-white"
              >
                <Menu className="w-5 h-5" />
              </button>
              
              {/* Search */}
              <div className="hidden sm:flex items-center gap-3 px-4 py-2 bg-gray-800 rounded-xl border border-red-primary/20">
                <Search className="w-4 h-4 text-gray-400" />
                <input
                  type="text"
                  placeholder="Search stocks, portfolios..."
                  className="bg-transparent text-sm text-gray-300 placeholder-gray-400 outline-none w-64 focus-red"
                />
              </div>
            </div>

            <div className="flex items-center gap-4">
              {/* Market Status */}
              <div className="hidden lg:flex items-center gap-2">
                <div className="flex items-center gap-2 px-3 py-1 glass-red rounded-full">
                  <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
                  <span className="text-sm font-medium text-green-400">Market Open</span>
                </div>
              </div>

              {/* Portfolio Value */}
              <div className="hidden md:flex items-center gap-2 px-4 py-2 glass-black rounded-xl">
                <Wallet className="w-4 h-4 text-red-400" />
                <div>
                  <p className="text-sm font-bold text-white">$929,000</p>
                  <p className="text-xs text-green-400">+2.4%</p>
                </div>
              </div>

              {/* Notifications */}
              <motion.button
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                className="relative p-2 rounded-lg bg-gray-800 hover:bg-red-primary/20 transition-colors"
              >
                <Bell className="w-5 h-5 text-gray-400" />
                <div className="absolute -top-1 -right-1 w-3 h-3 bg-red-500 rounded-full flex items-center justify-center">
                  <span className="text-xs text-white font-bold">3</span>
                </div>
              </motion.button>
            </div>
          </div>
        </header>

        {/* Page content */}
        <main className="p-6 max-w-none overflow-x-hidden">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            className="w-full"
          >
            {children}
          </motion.div>
        </main>
      </div>
    </div>
  );
}
