'use client';

import { motion } from 'framer-motion';
import { ArrowUpIcon, ArrowDownIcon } from '@heroicons/react/24/outline';
import clsx from 'clsx';

interface StatCardProps {
  title: string;
  value: string;
  icon: React.ComponentType<{ className?: string }>;
  trend?: 'up' | 'down';
  trendValue?: string;
  color?: 'blue' | 'green' | 'red' | 'purple' | 'yellow';
}

const colorClasses = {
  blue: 'from-blue-500 to-blue-600',
  green: 'from-green-500 to-green-600',
  red: 'from-red-500 to-red-600',
  purple: 'from-purple-500 to-purple-600',
  yellow: 'from-yellow-500 to-yellow-600',
};

export default function StatCard({ 
  title, 
  value, 
  icon: Icon, 
  trend, 
  trendValue, 
  color = 'blue' 
}: StatCardProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      whileHover={{ y: -2 }}
      className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6 hover:border-slate-600 transition-all duration-200"
    >
      <div className="flex items-center justify-between">
        <div className="flex-1">
          <p className="text-slate-400 text-sm font-medium">{title}</p>
          <p className="text-2xl font-bold text-white mt-1">{value}</p>
          {trend && trendValue && (
            <div className={clsx(
              'flex items-center mt-2 text-sm',
              trend === 'up' ? 'text-green-400' : 'text-red-400'
            )}>
              {trend === 'up' ? (
                <ArrowUpIcon className="w-4 h-4 mr-1" />
              ) : (
                <ArrowDownIcon className="w-4 h-4 mr-1" />
              )}
              {trendValue}
            </div>
          )}
        </div>
        <div className={clsx(
          'w-12 h-12 rounded-lg bg-gradient-to-r flex items-center justify-center',
          colorClasses[color]
        )}>
          <Icon className="w-6 h-6 text-white" />
        </div>
      </div>
    </motion.div>
  );
}
