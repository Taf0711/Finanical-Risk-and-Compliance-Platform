'use client';

import { useState, ReactNode } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { motion } from 'framer-motion';
import {
  HomeIcon,
  ChartBarIcon,
  ArrowsRightLeftIcon,
  BellIcon,
  Cog6ToothIcon,
  ArrowRightOnRectangleIcon,
  Bars3Icon,
  XMarkIcon,
  UserCircleIcon
} from '@heroicons/react/24/outline';
import { useAuth } from '@/hooks/useAuth';
import clsx from 'clsx';

interface DashboardLayoutProps {
  children: ReactNode;
}

const navigation = [
  { name: 'Dashboard', href: '/dashboard', icon: HomeIcon },
  { name: 'Risk Analytics', href: '/risk', icon: ChartBarIcon },
  { name: 'Transactions', href: '/trades', icon: ArrowsRightLeftIcon },
  { name: 'Alerts', href: '/alerts', icon: BellIcon },
  { name: 'Settings', href: '/settings', icon: Cog6ToothIcon },
];

export default function DashboardLayout({ children }: DashboardLayoutProps) {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const pathname = usePathname();
  const { user, logout } = useAuth();

  return (
    <div className="min-h-screen bg-slate-900">
      {/* Mobile sidebar backdrop */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 z-40 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        >
          <div className="absolute inset-0 bg-black/50" />
        </div>
      )}

      {/* Sidebar */}
      <motion.div
        initial={false}
        animate={{
          x: sidebarOpen ? 0 : '-100%',
        }}
        className={clsx(
          'fixed inset-y-0 left-0 z-50 w-64 bg-slate-800/95 backdrop-blur-lg border-r border-slate-700 lg:translate-x-0 lg:static lg:inset-0',
          'lg:translate-x-0'
        )}
      >
        <div className="flex h-full flex-col">
          {/* Logo */}
          <div className="flex items-center justify-between h-16 px-6 border-b border-slate-700">
            <div className="flex items-center">
              <div className="w-8 h-8 bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
                <ChartBarIcon className="w-5 h-5 text-white" />
              </div>
              <span className="ml-3 text-xl font-bold text-white">RiskMonitor</span>
            </div>
            <button
              onClick={() => setSidebarOpen(false)}
              className="lg:hidden text-slate-400 hover:text-white"
            >
              <XMarkIcon className="w-6 h-6" />
            </button>
          </div>

          {/* Navigation */}
          <nav className="flex-1 px-4 py-6 space-y-2">
            {navigation.map((item) => {
              const isActive = pathname === item.href;
              return (
                <Link
                  key={item.name}
                  href={item.href}
                  className={clsx(
                    'flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-colors',
                    isActive
                      ? 'bg-blue-600 text-white'
                      : 'text-slate-300 hover:text-white hover:bg-slate-700'
                  )}
                >
                  <item.icon className="w-5 h-5 mr-3" />
                  {item.name}
                </Link>
              );
            })}
          </nav>

          {/* User Menu */}
          <div className="border-t border-slate-700 p-4">
            <div className="flex items-center mb-4">
              <UserCircleIcon className="w-8 h-8 text-slate-400" />
              <div className="ml-3">
                <p className="text-sm font-medium text-white">
                  {user?.first_name} {user?.last_name}
                </p>
                <p className="text-xs text-slate-400">{user?.email}</p>
              </div>
            </div>
            <button
              onClick={logout}
              className="flex items-center w-full px-3 py-2 text-sm text-slate-300 hover:text-white hover:bg-slate-700 rounded-lg transition-colors"
            >
              <ArrowRightOnRectangleIcon className="w-5 h-5 mr-3" />
              Sign Out
            </button>
          </div>
        </div>
      </motion.div>

      {/* Main content */}
      <div className="lg:pl-64">
        {/* Top bar */}
        <div className="sticky top-0 z-30 bg-slate-800/95 backdrop-blur-lg border-b border-slate-700">
          <div className="flex items-center justify-between h-16 px-6">
            <button
              onClick={() => setSidebarOpen(true)}
              className="lg:hidden text-slate-400 hover:text-white"
            >
              <Bars3Icon className="w-6 h-6" />
            </button>

            <div className="flex items-center space-x-4">
              {/* Notifications */}
              <button className="relative text-slate-400 hover:text-white transition-colors">
                <BellIcon className="w-6 h-6" />
                <span className="absolute -top-1 -right-1 w-3 h-3 bg-red-500 rounded-full"></span>
              </button>

              {/* User avatar */}
              <div className="w-8 h-8 bg-gradient-to-r from-blue-500 to-purple-600 rounded-full flex items-center justify-center">
                <span className="text-white text-sm font-medium">
                  {user?.first_name?.charAt(0)}{user?.last_name?.charAt(0)}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Page content */}
        <main className="min-h-screen">
          {children}
        </main>
      </div>
    </div>
  );
}
