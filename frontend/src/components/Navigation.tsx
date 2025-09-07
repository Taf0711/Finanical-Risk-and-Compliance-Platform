'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  LayoutDashboard,
  TrendingUp,
  PieChart,
  Shield,
  Bell,
  BarChart3,
  Settings,
  User,
  LogOut,
  Menu,
  X,
  Activity,
  Search,
  Wallet,
  AlertTriangle,
  Plus,
  Minus,
  ChevronDown,
  Zap,
  Eye,
  DollarSign
} from 'lucide-react';

const navigationItems = [
  {
    name: 'Dashboard',
    href: '/dashboard',
    icon: LayoutDashboard,
    description: 'Overview & Analytics',
    color: 'text-primary'
  },
  {
    name: 'Portfolio',
    href: '/portfolio',
    icon: PieChart,
    description: 'Manage Holdings',
    color: 'text-success'
  },
  {
    name: 'Trading',
    href: '/trading',
    icon: TrendingUp,
    description: 'Execute Trades',
    color: 'text-warning'
  },
  {
    name: 'Risk Analysis',
    href: '/risk',
    icon: Shield,
    description: 'Risk Metrics',
    color: 'text-danger'
  },
  {
    name: 'Analytics',
    href: '/analytics',
    icon: BarChart3,
    description: 'Advanced Charts',
    color: 'text-accent'
  },
  {
    name: 'Alerts',
    href: '/alerts',
    icon: Bell,
    description: 'Notifications',
    color: 'text-warning'
  }
];

const quickActions = [
  {
    name: 'Buy',
    icon: Plus,
    action: 'buy',
    color: 'text-success',
    bgColor: 'hover:bg-success/10'
  },
  {
    name: 'Sell',
    icon: Minus,
    action: 'sell',
    color: 'text-danger',
    bgColor: 'hover:bg-danger/10'
  },
  {
    name: 'Alert',
    icon: AlertTriangle,
    action: 'alert',
    color: 'text-warning',
    bgColor: 'hover:bg-warning/10'
  }
];

interface NavigationProps {
  className?: string;
}

export default function Navigation({ className = '' }: NavigationProps) {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const pathname = usePathname();

  const isActive = (href: string) => pathname === href;

  // Close mobile menu when route changes
  useEffect(() => {
    setIsMobileMenuOpen(false);
  }, [pathname]);

  // Mock user data - replace with real user context
  const user = {
    name: 'John Trader',
    email: 'john@finrisk.pro',
    role: 'Premium Account',
    avatar: null,
    portfolioValue: '$2,847,392',
    portfolioChange: '+12.5%',
    isOnline: true
  };

  return (
    <>
      {/* Mobile Menu Button */}
      <motion.button
        whileHover={{ scale: 1.05 }}
        whileTap={{ scale: 0.95 }}
        onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
        className="fixed top-4 left-4 z-50 p-3 glass-card rounded-xl lg:hidden no-print"
      >
        <AnimatePresence mode="wait">
          {isMobileMenuOpen ? (
            <motion.div
              key="close"
              initial={{ rotate: -90, opacity: 0 }}
              animate={{ rotate: 0, opacity: 1 }}
              exit={{ rotate: 90, opacity: 0 }}
              transition={{ duration: 0.2 }}
            >
              <X className="w-5 h-5 text-foreground" />
            </motion.div>
          ) : (
            <motion.div
              key="menu"
              initial={{ rotate: 90, opacity: 0 }}
              animate={{ rotate: 0, opacity: 1 }}
              exit={{ rotate: -90, opacity: 0 }}
              transition={{ duration: 0.2 }}
            >
              <Menu className="w-5 h-5 text-foreground" />
            </motion.div>
          )}
        </AnimatePresence>
      </motion.button>

      {/* Sidebar */}
      <motion.aside 
        initial={false}
        animate={{
          x: isMobileMenuOpen ? 0 : -280,
        }}
        transition={{ type: "spring", damping: 25, stiffness: 200 }}
        className={`sidebar sidebar-mobile ${isMobileMenuOpen ? 'open' : ''} ${className}`}
      >
        <div className="flex flex-col h-full">
          {/* Logo Section */}
          <div className="flex items-center gap-3 p-6 border-b border-white/10">
            <motion.div 
              whileHover={{ scale: 1.05, rotate: 5 }}
              className="w-10 h-10 rounded-xl bg-gradient-primary flex items-center justify-center shadow-glow-primary"
            >
              <Activity className="w-6 h-6 text-white" />
            </motion.div>
            <div>
              <h1 className="text-xl font-bold gradient-text">FinRisk Pro</h1>
              <p className="text-2xs text-foreground-muted">Financial Analytics</p>
            </div>
          </div>

          {/* User Profile Section */}
          <div className="p-4 border-b border-white/10">
            <motion.div 
              whileHover={{ scale: 1.02 }}
              className="glass-card p-4 rounded-xl cursor-pointer"
              onClick={() => setUserMenuOpen(!userMenuOpen)}
            >
              <div className="flex items-center gap-3">
                <div className="relative">
                  <div className="w-10 h-10 rounded-full bg-gradient-primary flex items-center justify-center">
                    <User className="w-5 h-5 text-white" />
                  </div>
                  {user.isOnline && (
                    <div className="absolute -bottom-1 -right-1 w-4 h-4 bg-success rounded-full border-2 border-background flex items-center justify-center">
                      <div className="w-2 h-2 bg-background rounded-full animate-pulse" />
                    </div>
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-foreground truncate">{user.name}</p>
                  <p className="text-2xs text-foreground-muted truncate">{user.role}</p>
                </div>
                <motion.div
                  animate={{ rotate: userMenuOpen ? 180 : 0 }}
                  transition={{ duration: 0.2 }}
                >
                  <ChevronDown className="w-4 h-4 text-foreground-muted" />
                </motion.div>
              </div>
              
              <AnimatePresence>
                {userMenuOpen && (
                  <motion.div
                    initial={{ height: 0, opacity: 0 }}
                    animate={{ height: 'auto', opacity: 1 }}
                    exit={{ height: 0, opacity: 0 }}
                    transition={{ duration: 0.2 }}
                    className="mt-4 pt-4 border-t border-white/10 overflow-hidden"
                  >
                    <div className="flex items-center justify-between text-2xs">
                      <span className="text-foreground-muted">Portfolio Value</span>
                      <div className="text-right">
                        <div className="font-semibold text-foreground">{user.portfolioValue}</div>
                        <div className="text-success">{user.portfolioChange}</div>
                      </div>
                    </div>
                  </motion.div>
                )}
              </AnimatePresence>
            </motion.div>
          </div>

          {/* Navigation Menu */}
          <nav className="flex-1 p-4 space-y-2 overflow-y-auto scrollbar-thin">
            <div className="mb-6">
              <h3 className="text-2xs font-semibold text-foreground-muted uppercase tracking-wider mb-3 px-2">
                Main Menu
              </h3>
              {navigationItems.map((item, index) => {
                const Icon = item.icon;
                const active = isActive(item.href);
                
                return (
                  <motion.div
                    key={item.name}
                    initial={{ opacity: 0, x: -20 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: index * 0.05 }}
                  >
                    <Link
                      href={item.href}
                      className={`nav-item ${active ? 'active' : ''} group`}
                      onClick={() => setIsMobileMenuOpen(false)}
                    >
                      <div className={`p-2 rounded-lg transition-all duration-200 ${
                        active 
                          ? 'bg-primary/20 shadow-glow-primary/20' 
                          : 'bg-glass group-hover:bg-glass-strong'
                      }`}>
                        <Icon className={`w-5 h-5 ${active ? 'text-primary' : 'text-foreground-muted group-hover:text-foreground'}`} />
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className={`text-sm font-medium truncate ${
                          active ? 'text-primary' : 'text-foreground group-hover:text-foreground'
                        }`}>
                          {item.name}
                        </p>
                        <p className={`text-2xs truncate ${
                          active ? 'text-primary/80' : 'text-foreground-muted'
                        }`}>
                          {item.description}
                        </p>
                      </div>
                      {active && (
                        <motion.div
                          layoutId="activeIndicator"
                          className="w-1 h-8 bg-primary rounded-full shadow-glow-primary"
                        />
                      )}
                    </Link>
                  </motion.div>
                );
              })}
            </div>

            {/* Quick Actions */}
            <div>
              <h3 className="text-2xs font-semibold text-foreground-muted uppercase tracking-wider mb-3 px-2">
                Quick Actions
              </h3>
              <div className="grid grid-cols-3 gap-2">
                {quickActions.map((action, index) => {
                  const Icon = action.icon;
                  return (
                    <motion.button
                      key={action.name}
                      whileHover={{ scale: 1.05, y: -2 }}
                      whileTap={{ scale: 0.95 }}
                      initial={{ opacity: 0, y: 20 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: 0.3 + index * 0.1 }}
                      className={`flex flex-col items-center gap-2 p-3 rounded-xl glass-card transition-all duration-200 ${action.bgColor}`}
                    >
                      <Icon className={`w-4 h-4 ${action.color}`} />
                      <span className="text-2xs font-medium text-foreground-muted">{action.name}</span>
                    </motion.button>
                  );
                })}
              </div>
            </div>
          </nav>

          {/* Market Status */}
          <div className="p-4 border-t border-white/10">
            <motion.div 
              whileHover={{ scale: 1.02 }}
              className="glass-card p-3 rounded-xl"
            >
              <div className="flex items-center justify-between mb-2">
                <span className="text-2xs text-foreground-muted uppercase tracking-wider">Market Status</span>
                <div className="flex items-center gap-1">
                  <div className="status-online" />
                  <span className="text-2xs text-success font-medium">OPEN</span>
                </div>
              </div>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <TrendingUp className="w-3 h-3 text-success" />
                  <span className="text-xs text-foreground">SPY</span>
                </div>
                <div className="text-right">
                  <div className="text-xs font-semibold text-foreground">$428.50</div>
                  <div className="text-2xs text-success">+0.8%</div>
                </div>
              </div>
            </motion.div>
          </div>

          {/* Bottom Actions */}
          <div className="p-4 border-t border-white/10">
            <div className="flex items-center justify-between">
              <motion.button
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                className="p-2 rounded-lg glass-card hover:bg-glass-strong transition-colors"
              >
                <Settings className="w-4 h-4 text-foreground-muted hover:text-foreground transition-colors" />
              </motion.button>
              
              <motion.button
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium text-danger hover:bg-danger/10 transition-colors"
              >
                <LogOut className="w-4 h-4" />
                <span className="hidden sm:inline">Sign Out</span>
              </motion.button>
            </div>
          </div>
        </div>
      </motion.aside>

      {/* Mobile Overlay */}
      <AnimatePresence>
        {isMobileMenuOpen && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 bg-black/50 backdrop-blur-sm z-40 lg:hidden"
            onClick={() => setIsMobileMenuOpen(false)}
          />
        )}
      </AnimatePresence>
    </>
  );
}