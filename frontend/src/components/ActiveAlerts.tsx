'use client';

import { motion } from 'framer-motion';
import { format } from 'date-fns';
import { ExclamationTriangleIcon, InformationCircleIcon, XCircleIcon } from '@heroicons/react/24/outline';
import { Alert } from '@/types';

interface ActiveAlertsProps {
  alerts: Alert[];
}

export default function ActiveAlerts({ alerts }: ActiveAlertsProps) {
  const getSeverityIcon = (severity: string) => {
    switch (severity.toLowerCase()) {
      case 'critical':
      case 'high':
        return XCircleIcon;
      case 'medium':
        return ExclamationTriangleIcon;
      case 'low':
      default:
        return InformationCircleIcon;
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity.toLowerCase()) {
      case 'critical':
      case 'high':
        return 'text-red-400 bg-red-400/10';
      case 'medium':
        return 'text-yellow-400 bg-yellow-400/10';
      case 'low':
      default:
        return 'text-blue-400 bg-blue-400/10';
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6"
    >
      <h3 className="text-lg font-semibold text-white mb-4">Active Alerts</h3>
      <div className="space-y-3 max-h-96 overflow-y-auto">
        {alerts.slice(0, 10).map((alert, index) => {
          const SeverityIcon = getSeverityIcon(alert.severity);
          
          return (
            <motion.div
              key={alert.id}
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: index * 0.1 }}
              className="flex items-start space-x-3 p-3 border border-slate-700 rounded-lg hover:bg-slate-700/30 transition-colors"
            >
              <div className={`p-1 rounded ${getSeverityColor(alert.severity)}`}>
                <SeverityIcon className="w-4 h-4" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="text-white font-medium truncate">{alert.title}</div>
                <div className="text-slate-400 text-sm mt-1 line-clamp-2">
                  {alert.description}
                </div>
                <div className="flex items-center justify-between mt-2">
                  <span className={`px-2 py-1 text-xs rounded ${getSeverityColor(alert.severity)}`}>
                    {alert.severity}
                  </span>
                  <span className="text-slate-500 text-xs">
                    {format(new Date(alert.created_at), 'MMM dd, HH:mm')}
                  </span>
                </div>
              </div>
            </motion.div>
          );
        })}
        {alerts.length === 0 && (
          <div className="text-center py-8 text-slate-400">
            No active alerts
          </div>
        )}
      </div>
    </motion.div>
  );
}
