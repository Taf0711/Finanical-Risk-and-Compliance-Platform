'use client';

import React from 'react';
import { motion } from 'framer-motion';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { cn, formatCurrency, formatPercent, getChangeColor, getChangeIcon } from '@/lib/utils';
import { TrendingUp, TrendingDown, Minus } from 'lucide-react';

interface StatCardProps {
  title: string;
  value: string | number;
  change?: number;
  changeType?: 'positive' | 'negative' | 'neutral';
  icon?: React.ReactNode;
  subtitle?: string;
  format?: 'currency' | 'percent' | 'number' | 'text';
  className?: string;
  trend?: number[];
  loading?: boolean;
  // Legacy props for backward compatibility
  trendValue?: string;
  color?: 'blue' | 'green' | 'red' | 'purple' | 'yellow';
}

export default function StatCard({ 
  title, 
  value, 
  change,
  changeType,
  icon,
  subtitle,
  format = 'text',
  className,
  trend, 
  loading = false,
  // Legacy props
  trendValue, 
  color = 'blue' 
}: StatCardProps) {
  const formatValue = (val: string | number) => {
    if (typeof val === 'string') return val;
    
    switch (format) {
      case 'currency':
        return formatCurrency(val);
      case 'percent':
        return formatPercent(val);
      case 'number':
        return val.toLocaleString();
      default:
        return val.toString();
    }
  };

  const getChangeTypeFromValue = (changeVal: number) => {
    if (changeVal > 0) return 'positive';
    if (changeVal < 0) return 'negative';
    return 'neutral';
  };

  const actualChangeType = changeType || (change !== undefined ? getChangeTypeFromValue(change) : 'neutral');

  const changeIcon = change !== undefined ? (
    change > 0 ? <TrendingUp className="w-3 h-3" /> :
    change < 0 ? <TrendingDown className="w-3 h-3" /> :
    <Minus className="w-3 h-3" />
  ) : null;

  const badgeVariant = actualChangeType === 'positive' ? 'success' : 
                      actualChangeType === 'negative' ? 'danger' : 'secondary';

  const colorGradients = {
    blue: 'from-blue-500/20 to-blue-600/20',
    green: 'from-green-500/20 to-green-600/20',
    red: 'from-red-500/20 to-red-600/20',
    purple: 'from-purple-500/20 to-purple-600/20',
    yellow: 'from-yellow-500/20 to-yellow-600/20',
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      whileHover={{ 
        scale: 1.02,
        transition: { duration: 0.2 }
      }}
      whileTap={{ scale: 0.98 }}
      className={cn("group", className)}
    >
      <Card className={cn(
        "relative overflow-hidden transition-all duration-300",
        "hover:shadow-xl hover:shadow-blue-500/10",
        "bg-gradient-to-br from-white/90 to-white/60 dark:from-gray-900/90 dark:to-gray-900/60",
        "backdrop-blur-md border-white/20 dark:border-gray-700/50",
        loading && "animate-pulse"
      )}>
        {/* Gradient overlay */}
        <div className={cn(
          "absolute inset-0 bg-gradient-to-br opacity-0 group-hover:opacity-100 transition-opacity duration-300",
          colorGradients[color]
        )} />
        
        {/* Trend line background */}
        {trend && trend.length > 0 && (
          <div className="absolute top-0 right-0 w-24 h-12 opacity-10">
            <svg className="w-full h-full" viewBox="0 0 100 50">
              <polyline
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                points={trend.map((val, i) => `${(i / (trend.length - 1)) * 100},${50 - (val / Math.max(...trend)) * 40}`).join(' ')}
              />
            </svg>
          </div>
        )}

        <CardContent className="p-4 sm:p-6 relative z-10">
          <div className="flex items-center justify-between mb-3 sm:mb-4">
            <h3 className="text-xs sm:text-sm font-medium text-gray-600 dark:text-gray-400 uppercase tracking-wide">
              {title}
            </h3>
            {icon && (
              <motion.div 
                className={cn(
                  "p-2 sm:p-2.5 rounded-xl border border-white/10 bg-gradient-to-br",
                  colorGradients[color]
                )}
                whileHover={{ rotate: 5, scale: 1.1 }}
                transition={{ duration: 0.2 }}
              >
                <div className={cn(
                  color === 'blue' && "text-blue-600 dark:text-blue-400",
                  color === 'green' && "text-green-600 dark:text-green-400",
                  color === 'red' && "text-red-600 dark:text-red-400",
                  color === 'purple' && "text-purple-600 dark:text-purple-400",
                  color === 'yellow' && "text-yellow-600 dark:text-yellow-400"
                )}>
                  {icon}
                </div>
              </motion.div>
          )}
        </div>
          
          <div className="space-y-3">
            <motion.div
              initial={{ scale: 0.8, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              transition={{ delay: 0.1 }}
            >
              <p className="text-xl sm:text-2xl lg:text-3xl font-bold bg-gradient-to-r from-gray-900 to-gray-700 dark:from-white dark:to-gray-200 bg-clip-text text-transparent">
                {loading ? '---' : formatValue(value)}
              </p>
            </motion.div>
            
            {subtitle && (
              <p className="text-sm text-gray-500 dark:text-gray-400">
                {subtitle}
              </p>
            )}
            
            {/* New change display */}
            {change !== undefined && !loading && (
              <motion.div
                initial={{ opacity: 0, x: -10 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: 0.2 }}
                className="flex items-center gap-2"
              >
                <Badge variant={badgeVariant} className="flex items-center gap-1 px-2 py-1">
                  {changeIcon}
                  <span className="font-medium">
                    {format === 'percent' ? formatPercent(Math.abs(change)) : `${change > 0 ? '+' : ''}${change.toFixed(2)}${format === 'currency' ? '%' : ''}`}
                  </span>
                </Badge>
                <span className="text-xs text-gray-500 dark:text-gray-400">
                  vs previous period
                </span>
              </motion.div>
            )}

            {/* Legacy trend display for backward compatibility */}
            {trendValue && (
              <motion.div
                initial={{ opacity: 0, x: -10 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: 0.2 }}
                className="flex items-center gap-2"
              >
                <Badge variant="secondary" className="flex items-center gap-1 px-2 py-1">
                  <span className="font-medium">{trendValue}</span>
                </Badge>
              </motion.div>
            )}
        </div>
        </CardContent>

        {/* Shimmer effect for loading */}
        {loading && (
          <div className="absolute inset-0 -translate-x-full animate-[shimmer_2s_infinite] bg-gradient-to-r from-transparent via-white/10 to-transparent" />
        )}
      </Card>
    </motion.div>
  );
}