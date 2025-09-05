'use client';

import { useState, useEffect, useMemo } from 'react';
import { motion } from 'framer-motion';
import { 
  FunnelIcon,
  ArrowsUpDownIcon,
  MagnifyingGlassIcon,
  ArrowDownTrayIcon,
  EyeIcon,
  ArrowUpIcon,
  ArrowDownIcon
} from '@heroicons/react/24/outline';
import { format } from 'date-fns';
import { useRequireAuth } from '@/hooks/useAuth';
import { transactionApi, portfolioApi } from '@/lib/api';
import { Transaction, Portfolio, TransactionFilters, SortOption } from '@/types';
import DashboardLayout from '@/components/DashboardLayout';

export default function TradesPage() {
  const { user, loading } = useRequireAuth();
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [portfolios, setPortfolios] = useState<Portfolio[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [filters, setFilters] = useState<TransactionFilters>({});
  const [sortOption, setSortOption] = useState<SortOption>({ field: 'created_at', direction: 'desc' });
  const [selectedTransaction, setSelectedTransaction] = useState<Transaction | null>(null);
  const [showFilters, setShowFilters] = useState(false);
  
  useEffect(() => {
    if (!loading && user) {
      loadData();
    }
  }, [user, loading, filters]);

  const loadData = async () => {
    try {
      setIsLoading(true);
      const [transactionsData, portfoliosData] = await Promise.all([
        transactionApi.getAll(filters),
        portfolioApi.getAll()
      ]);
      setTransactions(transactionsData || []);
      setPortfolios(portfoliosData);
    } catch (error) {
      console.error('Failed to load data:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const filteredAndSortedTransactions = useMemo(() => {
    let filtered = transactions;

    // Search filter
    if (searchTerm) {
      filtered = filtered.filter(tx => 
        tx.symbol.toLowerCase().includes(searchTerm.toLowerCase()) ||
        tx.transaction_type.toLowerCase().includes(searchTerm.toLowerCase()) ||
        tx.status.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    // Sort
    filtered.sort((a, b) => {
      const aValue = a[sortOption.field as keyof Transaction];
      const bValue = b[sortOption.field as keyof Transaction];
      
      if (sortOption.direction === 'asc') {
        return aValue > bValue ? 1 : -1;
      } else {
        return aValue < bValue ? 1 : -1;
      }
    });

    return filtered;
  }, [transactions, searchTerm, sortOption]);

  const handleFilterChange = (key: keyof TransactionFilters, value: string) => {
    setFilters(prev => ({
      ...prev,
      [key]: value || undefined
    }));
  };

  const clearFilters = () => {
    setFilters({});
    setSearchTerm('');
  };

  const exportTransactions = () => {
    // Create CSV content
    const headers = ['Date', 'Type', 'Symbol', 'Quantity', 'Price', 'Amount', 'Status'];
    const csvContent = [
      headers.join(','),
      ...filteredAndSortedTransactions.map(tx => [
        format(new Date(tx.created_at), 'yyyy-MM-dd HH:mm:ss'),
        tx.transaction_type,
        tx.symbol,
        tx.quantity,
        tx.price,
        tx.amount,
        tx.status
      ].join(','))
    ].join('\n');

    // Download CSV
    const blob = new Blob([csvContent], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `transactions_${format(new Date(), 'yyyy-MM-dd')}.csv`;
    a.click();
    window.URL.revokeObjectURL(url);
  };

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'completed': return 'text-green-400 bg-green-400/10';
      case 'pending': return 'text-yellow-400 bg-yellow-400/10';
      case 'failed': return 'text-red-400 bg-red-400/10';
      default: return 'text-gray-400 bg-gray-400/10';
    }
  };

  const getTransactionTypeColor = (type: string) => {
    return type === 'BUY' ? 'text-green-400' : 'text-red-400';
  };

  if (loading || isLoading) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-white">Transaction History</h1>
            <p className="text-slate-400 mt-1">
              View and analyze all your trading activities
            </p>
          </div>
          <div className="flex items-center space-x-3">
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              onClick={exportTransactions}
              className="flex items-center px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
            >
              <ArrowDownTrayIcon className="h-5 w-5 mr-2" />
              Export CSV
            </motion.button>
          </div>
        </div>

        {/* Search and Filters */}
        <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6">
          <div className="flex flex-col lg:flex-row gap-4">
            {/* Search */}
            <div className="flex-1">
              <div className="relative">
                <MagnifyingGlassIcon className="h-5 w-5 absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-400" />
                <input
                  type="text"
                  placeholder="Search transactions..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="w-full pl-10 pr-4 py-2 bg-slate-700 border border-slate-600 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>

            {/* Filter Toggle */}
            <button
              onClick={() => setShowFilters(!showFilters)}
              className="flex items-center px-4 py-2 bg-slate-700 text-white rounded-lg hover:bg-slate-600 transition-colors"
            >
              <FunnelIcon className="h-5 w-5 mr-2" />
              Filters
            </button>

            {/* Sort */}
            <select
              value={`${sortOption.field}_${sortOption.direction}`}
              onChange={(e) => {
                const [field, direction] = e.target.value.split('_');
                setSortOption({ field, direction: direction as 'asc' | 'desc' });
              }}
              className="px-4 py-2 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="created_at_desc">Latest First</option>
              <option value="created_at_asc">Oldest First</option>
              <option value="amount_desc">Highest Amount</option>
              <option value="amount_asc">Lowest Amount</option>
              <option value="symbol_asc">Symbol A-Z</option>
              <option value="symbol_desc">Symbol Z-A</option>
            </select>
          </div>

          {/* Advanced Filters */}
          {showFilters && (
            <motion.div
              initial={{ height: 0, opacity: 0 }}
              animate={{ height: 'auto', opacity: 1 }}
              exit={{ height: 0, opacity: 0 }}
              className="mt-4 pt-4 border-t border-slate-700 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4"
            >
              <select
                value={filters.portfolio_id || ''}
                onChange={(e) => handleFilterChange('portfolio_id', e.target.value)}
                className="px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">All Portfolios</option>
                {portfolios.map(portfolio => (
                  <option key={portfolio.id} value={portfolio.id}>
                    {portfolio.name}
                  </option>
                ))}
              </select>

              <select
                value={filters.transaction_type || ''}
                onChange={(e) => handleFilterChange('transaction_type', e.target.value)}
                className="px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">All Types</option>
                <option value="BUY">Buy</option>
                <option value="SELL">Sell</option>
              </select>

              <select
                value={filters.status || ''}
                onChange={(e) => handleFilterChange('status', e.target.value)}
                className="px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">All Status</option>
                <option value="COMPLETED">Completed</option>
                <option value="PENDING">Pending</option>
                <option value="FAILED">Failed</option>
              </select>

              <button
                onClick={clearFilters}
                className="px-3 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
              >
                Clear Filters
              </button>
            </motion.div>
          )}
        </div>

        {/* Transactions Table */}
        <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-slate-700/50">
                <tr>
                  <th className="px-6 py-4 text-left text-xs font-medium text-slate-300 uppercase tracking-wider">
                    Date & Time
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-medium text-slate-300 uppercase tracking-wider">
                    Type
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-medium text-slate-300 uppercase tracking-wider">
                    Symbol
                  </th>
                  <th className="px-6 py-4 text-right text-xs font-medium text-slate-300 uppercase tracking-wider">
                    Quantity
                  </th>
                  <th className="px-6 py-4 text-right text-xs font-medium text-slate-300 uppercase tracking-wider">
                    Price
                  </th>
                  <th className="px-6 py-4 text-right text-xs font-medium text-slate-300 uppercase tracking-wider">
                    Amount
                  </th>
                  <th className="px-6 py-4 text-center text-xs font-medium text-slate-300 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-4 text-center text-xs font-medium text-slate-300 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-700">
                {filteredAndSortedTransactions.map((transaction, index) => (
                  <motion.tr
                    key={transaction.id}
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: index * 0.05 }}
                    className="hover:bg-slate-700/30 transition-colors"
                  >
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-white">
                        {format(new Date(transaction.created_at), 'MMM dd, yyyy')}
                      </div>
                      <div className="text-xs text-slate-400">
                        {format(new Date(transaction.created_at), 'HH:mm:ss')}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className={`flex items-center ${getTransactionTypeColor(transaction.transaction_type)}`}>
                        {transaction.transaction_type === 'BUY' ? (
                          <ArrowUpIcon className="h-4 w-4 mr-1" />
                        ) : (
                          <ArrowDownIcon className="h-4 w-4 mr-1" />
                        )}
                        {transaction.transaction_type}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-white">
                      {transaction.symbol}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-right text-white">
                      {parseFloat(transaction.quantity).toLocaleString()}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-right text-white">
                      ${parseFloat(transaction.price).toLocaleString()}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-right font-medium text-white">
                      ${parseFloat(transaction.amount).toLocaleString()}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-center">
                      <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${getStatusColor(transaction.status)}`}>
                        {transaction.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-center">
                      <button
                        onClick={() => setSelectedTransaction(transaction)}
                        className="text-blue-400 hover:text-blue-300 transition-colors"
                      >
                        <EyeIcon className="h-5 w-5" />
                      </button>
                    </td>
                  </motion.tr>
                ))}
              </tbody>
            </table>
          </div>

          {filteredAndSortedTransactions.length === 0 && (
            <div className="text-center py-12">
              <p className="text-slate-400">No transactions found</p>
            </div>
          )}
        </div>

        {/* Transaction Details Modal */}
        {selectedTransaction && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
            onClick={() => setSelectedTransaction(null)}
          >
            <motion.div
              initial={{ scale: 0.9, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              className="bg-slate-800 border border-slate-700 rounded-xl p-6 max-w-md w-full"
              onClick={(e) => e.stopPropagation()}
            >
              <h3 className="text-xl font-bold text-white mb-4">Transaction Details</h3>
              <div className="space-y-3">
                <div className="flex justify-between">
                  <span className="text-slate-400">ID:</span>
                  <span className="text-white font-mono text-sm">{selectedTransaction.id}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-slate-400">Type:</span>
                  <span className={getTransactionTypeColor(selectedTransaction.transaction_type)}>
                    {selectedTransaction.transaction_type}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-slate-400">Symbol:</span>
                  <span className="text-white">{selectedTransaction.symbol}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-slate-400">Quantity:</span>
                  <span className="text-white">{parseFloat(selectedTransaction.quantity).toLocaleString()}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-slate-400">Price:</span>
                  <span className="text-white">${parseFloat(selectedTransaction.price).toLocaleString()}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-slate-400">Total Amount:</span>
                  <span className="text-white font-semibold">${parseFloat(selectedTransaction.amount).toLocaleString()}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-slate-400">Status:</span>
                  <span className={`px-2 py-1 text-xs font-semibold rounded ${getStatusColor(selectedTransaction.status)}`}>
                    {selectedTransaction.status}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-slate-400">Risk Score:</span>
                  <span className="text-white">{selectedTransaction.risk_score}/100</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-slate-400">KYC Verified:</span>
                  <span className={selectedTransaction.kyc_verified ? 'text-green-400' : 'text-red-400'}>
                    {selectedTransaction.kyc_verified ? 'Yes' : 'No'}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-slate-400">AML Checked:</span>
                  <span className={selectedTransaction.aml_checked ? 'text-green-400' : 'text-red-400'}>
                    {selectedTransaction.aml_checked ? 'Yes' : 'No'}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-slate-400">Created:</span>
                  <span className="text-white">{format(new Date(selectedTransaction.created_at), 'MMM dd, yyyy HH:mm:ss')}</span>
                </div>
              </div>
              <button
                onClick={() => setSelectedTransaction(null)}
                className="mt-6 w-full px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
              >
                Close
              </button>
            </motion.div>
          </motion.div>
        )}
      </div>
    </DashboardLayout>
  );
}
