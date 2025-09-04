'use client';

import { motion } from 'framer-motion';
import { format } from 'date-fns';
import { ArrowUpIcon, ArrowDownIcon } from '@heroicons/react/24/outline';
import { Transaction } from '@/types';

interface RecentTransactionsProps {
  transactions: Transaction[];
}

export default function RecentTransactions({ transactions }: RecentTransactionsProps) {
  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'completed': return 'text-green-400 bg-green-400/10';
      case 'pending': return 'text-yellow-400 bg-yellow-400/10';
      case 'failed': return 'text-red-400 bg-red-400/10';
      default: return 'text-gray-400 bg-gray-400/10';
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6"
    >
      <h3 className="text-lg font-semibold text-white mb-4">Recent Transactions</h3>
      <div className="space-y-3">
        {transactions.slice(0, 5).map((transaction, index) => (
          <motion.div
            key={transaction.id}
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: index * 0.1 }}
            className="flex items-center justify-between p-3 border border-slate-700 rounded-lg hover:bg-slate-700/30 transition-colors"
          >
            <div className="flex items-center space-x-3">
              <div className={`p-2 rounded-lg ${
                transaction.transaction_type === 'BUY' ? 'bg-green-500/20' : 'bg-red-500/20'
              }`}>
                {transaction.transaction_type === 'BUY' ? (
                  <ArrowUpIcon className="w-4 h-4 text-green-400" />
                ) : (
                  <ArrowDownIcon className="w-4 h-4 text-red-400" />
                )}
              </div>
              <div>
                <div className="text-white font-medium">{transaction.symbol}</div>
                <div className="text-slate-400 text-sm">
                  {parseFloat(transaction.quantity).toLocaleString()} shares
                </div>
              </div>
            </div>
            <div className="text-right">
              <div className="text-white font-semibold">
                ${parseFloat(transaction.amount).toLocaleString()}
              </div>
              <div className="flex items-center space-x-2">
                <span className={`px-2 py-1 text-xs rounded ${getStatusColor(transaction.status)}`}>
                  {transaction.status}
                </span>
              </div>
            </div>
          </motion.div>
        ))}
        {transactions.length === 0 && (
          <div className="text-center py-8 text-slate-400">
            No recent transactions
          </div>
        )}
      </div>
    </motion.div>
  );
}
