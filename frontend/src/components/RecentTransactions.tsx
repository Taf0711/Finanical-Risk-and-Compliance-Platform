'use client';

import { motion } from 'framer-motion';
import { format } from 'date-fns';
import { 
  ArrowUpIcon, 
  ArrowDownIcon, 
  ClockIcon,
  CheckCircleIcon,
  XCircleIcon,
  ExclamationTriangleIcon,
  ArrowsRightLeftIcon,
  ShieldCheckIcon,
  EyeIcon
} from '@heroicons/react/24/outline';
import { Transaction } from '@/types';
import { useState } from 'react';

interface RecentTransactionsProps {
  transactions: Transaction[];
}

export default function RecentTransactions({ transactions }: RecentTransactionsProps) {
  const [selectedTransaction, setSelectedTransaction] = useState<string | null>(null);

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'completed': return 'text-green-400 bg-green-400/10 border-green-400/20';
      case 'pending': return 'text-yellow-400 bg-yellow-400/10 border-yellow-400/20';
      case 'failed': return 'text-red-400 bg-red-400/10 border-red-400/20';
      default: return 'text-slate-400 bg-slate-400/10 border-slate-400/20';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status.toLowerCase()) {
      case 'completed': return CheckCircleIcon;
      case 'pending': return ClockIcon;
      case 'failed': return XCircleIcon;
      default: return ExclamationTriangleIcon;
    }
  };

  const getRiskScoreColor = (score: number) => {
    if (score >= 25) return 'text-red-400';
    if (score >= 15) return 'text-yellow-400';
    return 'text-green-400';
  };

  if (!transactions || transactions.length === 0) {
    return (
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6"
      >
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-white">Recent Transactions</h3>
          <ArrowsRightLeftIcon className="h-6 w-6 text-blue-400" />
        </div>
        <div className="text-center py-8">
          <ArrowsRightLeftIcon className="h-12 w-12 text-slate-600 mx-auto mb-3" />
          <p className="text-slate-400 mb-2">No recent transactions</p>
          <p className="text-sm text-slate-500">Transaction history will appear here</p>
        </div>
      </motion.div>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6"
    >
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center space-x-3">
          <div className="bg-blue-500/20 p-2 rounded-lg">
            <ArrowsRightLeftIcon className="h-6 w-6 text-blue-400" />
          </div>
          <div>
            <h3 className="text-lg font-semibold text-white">Recent Transactions</h3>
            <p className="text-sm text-slate-400">
              {transactions.length} transaction{transactions.length !== 1 ? 's' : ''}
            </p>
          </div>
        </div>
        <button className="text-blue-400 hover:text-blue-300 text-sm font-medium transition-colors">
          View All
        </button>
      </div>
      
      <div className="space-y-3 max-h-96 overflow-y-auto">
        {transactions.slice(0, 8).map((transaction, index) => {
          const StatusIcon = getStatusIcon(transaction.status);
          const isSelected = selectedTransaction === transaction.id;
          const isBuy = transaction.transaction_type === 'BUY';
          
          return (
            <motion.div
              key={transaction.id}
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: index * 0.05 }}
              onClick={() => setSelectedTransaction(isSelected ? null : transaction.id)}
              className={`border rounded-lg p-4 cursor-pointer transition-all duration-200 ${
                isSelected 
                  ? 'border-blue-500 bg-blue-500/5' 
                  : 'border-slate-700 hover:border-slate-600 hover:bg-slate-700/30'
              }`}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  {/* Transaction Type Icon */}
                  <div className={`p-2 rounded-lg ${
                    isBuy ? 'bg-green-500/20' : 'bg-red-500/20'
                  }`}>
                    {isBuy ? (
                      <ArrowUpIcon className="w-5 h-5 text-green-400" />
                    ) : (
                      <ArrowDownIcon className="w-5 h-5 text-red-400" />
                    )}
                  </div>
                  
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center space-x-2">
                      <h4 className="text-white font-semibold">{transaction.symbol}</h4>
                      <span className={`px-2 py-1 text-xs rounded-full font-medium ${
                        isBuy ? 'text-green-400 bg-green-400/10' : 'text-red-400 bg-red-400/10'
                      }`}>
                        {transaction.transaction_type}
                      </span>
                    </div>
                    <p className="text-slate-400 text-sm">
                      {parseFloat(transaction.quantity).toLocaleString()} shares @ ${parseFloat(transaction.price).toFixed(2)}
                    </p>
                  </div>
                </div>
                
                <div className="text-right">
                  <div className="text-white font-semibold text-lg">
                    ${parseFloat(transaction.amount).toLocaleString()}
                  </div>
                  <div className="flex items-center justify-end space-x-2 mt-1">
                    <StatusIcon className="w-4 h-4 text-slate-400" />
                    <span className={`px-2 py-1 text-xs rounded border ${getStatusColor(transaction.status)}`}>
                      {transaction.status}
                    </span>
                  </div>
                </div>
              </div>
              
              {/* Expanded Details */}
              {isSelected && (
                <motion.div
                  initial={{ opacity: 0, height: 0 }}
                  animate={{ opacity: 1, height: 'auto' }}
                  exit={{ opacity: 0, height: 0 }}
                  className="border-t border-slate-700 pt-4 mt-4"
                >
                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <p className="text-slate-400 mb-1">Transaction ID</p>
                      <p className="text-white font-mono text-xs truncate">
                        {transaction.id}
                      </p>
                    </div>
                    <div>
                      <p className="text-slate-400 mb-1">Executed</p>
                      <p className="text-white">
                        {format(new Date(transaction.executed_at), 'MMM dd, yyyy HH:mm')}
                      </p>
                    </div>
                    <div>
                      <p className="text-slate-400 mb-1">Risk Score</p>
                      <div className="flex items-center space-x-2">
                        <span className={`font-semibold ${getRiskScoreColor(transaction.risk_score)}`}>
                          {transaction.risk_score}
                        </span>
                        <div className="w-16 bg-slate-700 rounded-full h-2">
                          <div 
                            className={`h-2 rounded-full ${
                              transaction.risk_score >= 25 ? 'bg-red-400' :
                              transaction.risk_score >= 15 ? 'bg-yellow-400' : 'bg-green-400'
                            }`}
                            style={{ width: `${Math.min(transaction.risk_score, 30) / 30 * 100}%` }}
                          />
                        </div>
                      </div>
                    </div>
                    <div>
                      <p className="text-slate-400 mb-1">Currency</p>
                      <p className="text-white">{transaction.currency}</p>
                    </div>
                  </div>
                  
                  {/* Compliance Info */}
                  <div className="mt-4 p-3 bg-slate-800/50 rounded-lg">
                    <p className="text-slate-400 text-xs mb-2">Compliance Status</p>
                    <div className="flex items-center space-x-4 text-xs">
                      <div className="flex items-center space-x-1">
                        {transaction.kyc_verified ? (
                          <CheckCircleIcon className="w-4 h-4 text-green-400" />
                        ) : (
                          <XCircleIcon className="w-4 h-4 text-red-400" />
                        )}
                        <span className="text-slate-300">KYC</span>
                      </div>
                      <div className="flex items-center space-x-1">
                        {transaction.aml_checked ? (
                          <ShieldCheckIcon className="w-4 h-4 text-green-400" />
                        ) : (
                          <ExclamationTriangleIcon className="w-4 h-4 text-yellow-400" />
                        )}
                        <span className="text-slate-300">AML</span>
                      </div>
                      {(transaction as any).requires_review && (
                        <div className="flex items-center space-x-1">
                          <EyeIcon className="w-4 h-4 text-yellow-400" />
                          <span className="text-yellow-400">Review Required</span>
                        </div>
                      )}
                    </div>
                  </div>
                </motion.div>
              )}
            </motion.div>
          );
        })}
      </div>
      
      {/* Summary Footer */}
      {transactions.length > 0 && (
        <div className="border-t border-slate-700 pt-4 mt-4">
          <div className="flex items-center justify-between text-sm">
            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-1">
                <ArrowUpIcon className="w-4 h-4 text-green-400" />
                <span className="text-slate-400">
                  {transactions.filter(t => t.transaction_type === 'BUY').length} buys
                </span>
              </div>
              <div className="flex items-center space-x-1">
                <ArrowDownIcon className="w-4 h-4 text-red-400" />
                <span className="text-slate-400">
                  {transactions.filter(t => t.transaction_type === 'SELL').length} sells
                </span>
              </div>
            </div>
            <div className="text-slate-400">
              Total Volume: ${transactions.reduce((sum, t) => sum + parseFloat(t.amount), 0).toLocaleString()}
            </div>
          </div>
        </div>
      )}
    </motion.div>
  );
}