'use client';

import { motion, AnimatePresence } from 'framer-motion';
import { format } from 'date-fns';
import { 
  ExclamationTriangleIcon, 
  InformationCircleIcon, 
  XCircleIcon,
  BellIcon,
  ShieldExclamationIcon,
  EyeSlashIcon,
  CheckCircleIcon,
  ClockIcon
} from '@heroicons/react/24/outline';
import { Alert } from '@/types';
import { useState } from 'react';

interface ActiveAlertsProps {
  alerts: Alert[];
}

export default function ActiveAlerts({ alerts }: ActiveAlertsProps) {
  const [filter, setFilter] = useState<string>('all');
  const [dismissedAlerts, setDismissedAlerts] = useState<Set<string>>(new Set());

  const getSeverityIcon = (severity: string) => {
    switch (severity.toLowerCase()) {
      case 'critical':
        return XCircleIcon;
      case 'high':
        return ExclamationTriangleIcon;
      case 'medium':
        return ShieldExclamationIcon;
      case 'low':
      default:
        return InformationCircleIcon;
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity.toLowerCase()) {
      case 'critical':
        return 'text-red-400 bg-red-400/10 border-red-400/20';
      case 'high':
        return 'text-orange-400 bg-orange-400/10 border-orange-400/20';
      case 'medium':
        return 'text-yellow-400 bg-yellow-400/10 border-yellow-400/20';
      case 'low':
      default:
        return 'text-blue-400 bg-blue-400/10 border-blue-400/20';
    }
  };

  const getAlertTypeIcon = (alertType: string) => {
    switch (alertType.toLowerCase()) {
      case 'risk_breach':
        return ShieldExclamationIcon;
      case 'suspicious_activity':
        return EyeSlashIcon;
      case 'compliance_violation':
        return ExclamationTriangleIcon;
      default:
        return BellIcon;
    }
  };

  const handleDismissAlert = (alertId: string) => {
    setDismissedAlerts(prev => new Set([...prev, alertId]));
  };

  // Filter alerts based on selected filter and dismissed status
  const filteredAlerts = alerts
    .filter(alert => !dismissedAlerts.has(alert.id))
    .filter(alert => {
      if (filter === 'all') return true;
      return alert.severity.toLowerCase() === filter;
    });

  const alertCounts = {
    all: alerts.filter(a => !dismissedAlerts.has(a.id)).length,
    critical: alerts.filter(a => !dismissedAlerts.has(a.id) && a.severity.toLowerCase() === 'critical').length,
    high: alerts.filter(a => !dismissedAlerts.has(a.id) && a.severity.toLowerCase() === 'high').length,
    medium: alerts.filter(a => !dismissedAlerts.has(a.id) && a.severity.toLowerCase() === 'medium').length,
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6"
    >
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center space-x-3">
          <div className="bg-red-500/20 p-2 rounded-lg">
            <BellIcon className="h-6 w-6 text-red-400" />
          </div>
          <div>
            <h3 className="text-lg font-semibold text-white">Active Alerts</h3>
            <p className="text-sm text-slate-400">
              {alertCounts.all} active alert{alertCounts.all !== 1 ? 's' : ''}
            </p>
          </div>
        </div>
        {alertCounts.all > 0 && (
          <div className="flex items-center space-x-2">
            <div className="w-2 h-2 bg-red-400 rounded-full animate-pulse"></div>
            <span className="text-sm text-red-400 font-medium">Live</span>
          </div>
        )}
      </div>

      {/* Filter Tabs */}
      {alertCounts.all > 0 && (
        <div className="flex space-x-1 mb-4 bg-slate-700/30 p-1 rounded-lg">
          {[
            { key: 'all', label: 'All', count: alertCounts.all },
            { key: 'critical', label: 'Critical', count: alertCounts.critical },
            { key: 'high', label: 'High', count: alertCounts.high },
            { key: 'medium', label: 'Medium', count: alertCounts.medium },
          ].map(({ key, label, count }) => (
            <button
              key={key}
              onClick={() => setFilter(key)}
              className={`flex-1 px-3 py-2 text-sm rounded-md transition-colors ${
                filter === key
                  ? 'bg-slate-600 text-white'
                  : 'text-slate-400 hover:text-white hover:bg-slate-700/50'
              }`}
            >
              {label} {count > 0 && <span className="ml-1">({count})</span>}
            </button>
          ))}
        </div>
      )}

      {/* Alerts List */}
      <div className="space-y-3 max-h-96 overflow-y-auto">
        <AnimatePresence>
          {filteredAlerts.slice(0, 10).map((alert, index) => {
            const SeverityIcon = getSeverityIcon(alert.severity);
            const TypeIcon = getAlertTypeIcon(alert.alert_type);
            
            return (
              <motion.div
                key={alert.id}
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: 20, height: 0 }}
                transition={{ delay: index * 0.05 }}
                className={`relative border rounded-lg p-4 hover:bg-slate-700/20 transition-all duration-200 ${getSeverityColor(alert.severity)}`}
              >
                {/* Alert Content */}
                <div className="flex items-start space-x-3">
                  <div className={`p-2 rounded-lg ${getSeverityColor(alert.severity)}`}>
                    <SeverityIcon className="w-5 h-5" />
                  </div>
                  
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center space-x-2">
                        <h4 className="text-white font-semibold truncate">{alert.title}</h4>
                        <TypeIcon className="w-4 h-4 text-slate-400" />
                      </div>
                      <button
                        onClick={() => handleDismissAlert(alert.id)}
                        className="text-slate-500 hover:text-slate-300 transition-colors"
                      >
                        <CheckCircleIcon className="w-4 h-4" />
                      </button>
                    </div>
                    
                    <p className="text-slate-300 text-sm mb-3 line-clamp-2">
                      {alert.description}
                    </p>
                    
                    {/* Alert Metadata */}
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-3">
                        <span className={`px-2 py-1 text-xs rounded-full font-medium ${getSeverityColor(alert.severity)}`}>
                          {alert.severity.toUpperCase()}
                        </span>
                        <span className="text-slate-400 text-xs">
                          {alert.source.replace('_', ' ')}
                        </span>
                      </div>
                      <div className="flex items-center space-x-1 text-slate-500 text-xs">
                        <ClockIcon className="w-3 h-3" />
                        <span>{format(new Date(alert.created_at), 'MMM dd, HH:mm')}</span>
                      </div>
                    </div>
                    
                    {/* Additional Context */}
                    {alert.triggered_by && (
                      <div className="mt-2 p-2 bg-slate-800/50 rounded-lg">
                        <p className="text-xs text-slate-400 mb-1">Trigger Details:</p>
                        <div className="text-xs text-slate-300">
                          {alert.triggered_by.symbol && (
                            <span className="inline-block mr-3">
                              Symbol: <span className="text-white font-mono">{alert.triggered_by.symbol}</span>
                            </span>
                          )}
                          {alert.triggered_by.amount && (
                            <span className="inline-block mr-3">
                              Amount: <span className="text-white font-mono">
                                ${parseFloat(alert.triggered_by.amount).toLocaleString()}
                              </span>
                            </span>
                          )}
                          {alert.triggered_by.breach_ratio && (
                            <span className="inline-block">
                              Breach: <span className="text-white font-mono">
                                {((alert.triggered_by.breach_ratio - 1) * 100).toFixed(1)}%
                              </span>
                            </span>
                          )}
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              </motion.div>
            );
          })}
        </AnimatePresence>
        
        {/* Empty State */}
        {filteredAlerts.length === 0 && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="text-center py-8"
          >
            {alertCounts.all === 0 ? (
              <div>
                <CheckCircleIcon className="h-12 w-12 text-green-400 mx-auto mb-3" />
                <p className="text-slate-400 mb-2">All clear!</p>
                <p className="text-sm text-slate-500">No active alerts at this time</p>
              </div>
            ) : (
              <div>
                <BellIcon className="h-12 w-12 text-slate-600 mx-auto mb-3" />
                <p className="text-slate-400 mb-2">No alerts for this filter</p>
                <p className="text-sm text-slate-500">Try selecting a different severity level</p>
              </div>
            )}
          </motion.div>
        )}
        
        {/* Show More Button */}
        {filteredAlerts.length > 10 && (
          <div className="text-center pt-4 border-t border-slate-700">
            <button className="text-blue-400 hover:text-blue-300 text-sm font-medium transition-colors">
              View {filteredAlerts.length - 10} more alerts
            </button>
          </div>
        )}
      </div>
    </motion.div>
  );
}