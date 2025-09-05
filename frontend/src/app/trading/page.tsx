'use client';

import { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import {
  ChartBarIcon,
  BanknotesIcon,
  ArrowUpIcon,
  ArrowDownIcon,
  ClockIcon,
  CheckCircleIcon,
  XCircleIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
} from '@heroicons/react/24/outline';
import { useRequireAuth } from '@/hooks/useAuth';
import DashboardLayout from '@/components/DashboardLayout';
import { tradingApi } from '@/lib/api';

const orderSchema = z.object({
  symbol: z.string().min(1, 'Symbol is required').toUpperCase(),
  quantity: z.number().min(0.001, 'Quantity must be greater than 0'),
  side: z.enum(['buy', 'sell']),
  order_type: z.enum(['market', 'limit', 'stop', 'stop_limit']),
  time_in_force: z.enum(['day', 'gtc', 'ioc', 'fok']),
  limit_price: z.number().optional(),
  stop_price: z.number().optional(),
  portfolio_id: z.string().min(1, 'Portfolio is required'),
});

type OrderFormData = z.infer<typeof orderSchema>;

interface Account {
  id: string;
  account_number: string;
  status: string;
  currency: string;
  cash: string;
  portfolio_value: string;
  buying_power: string;
  equity: string;
  day_trade_count: number;
  pattern_day_trader: boolean;
  trading_blocked: boolean;
}

interface Position {
  symbol: string;
  quantity: string;
  market_value: string;
  cost_basis: string;
  unrealized_pl: string;
  unrealized_plpc: string;
  current_price: string;
  change_today: string;
}

interface Order {
  id: string;
  symbol: string;
  quantity: string;
  side: string;
  order_type: string;
  status: string;
  filled_quantity: string;
  limit_price?: string;
  stop_price?: string;
  submitted_at: string;
  filled_at?: string;
}

export default function TradingPage() {
  const { user, loading } = useRequireAuth();
  const [activeTab, setActiveTab] = useState('trade');
  const [account, setAccount] = useState<Account | null>(null);
  const [positions, setPositions] = useState<Position[]>([]);
  const [orders, setOrders] = useState<Order[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [orderSuccess, setOrderSuccess] = useState('');
  const [orderError, setOrderError] = useState('');

  const orderForm = useForm<OrderFormData>({
    resolver: zodResolver(orderSchema),
    defaultValues: {
      side: 'buy',
      order_type: 'market',
      time_in_force: 'day',
      portfolio_id: '', // Will be set when portfolios are loaded
    },
  });

  const watchOrderType = orderForm.watch('order_type');

  useEffect(() => {
    if (!loading && user) {
      loadTradingData();
    }
  }, [user, loading]);

  const loadTradingData = async () => {
    try {
      setIsLoading(true);
      
      // Load account, positions, and orders in parallel
      const [accountRes, positionsRes, ordersRes] = await Promise.allSettled([
        tradingApi.getAccount(),
        tradingApi.getPositions(),
        tradingApi.getOrders('', 20),
      ]);

      if (accountRes.status === 'fulfilled') {
        setAccount(accountRes.value);
      }

      if (positionsRes.status === 'fulfilled') {
        setPositions(positionsRes.value);
      }

      if (ordersRes.status === 'fulfilled') {
        setOrders(ordersRes.value);
      }
    } catch (error) {
      console.error('Failed to load trading data:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const onOrderSubmit = async (data: OrderFormData) => {
    try {
      setOrderError('');
      setOrderSuccess('');
      
      const order = await tradingApi.placeOrder(data);
      setOrderSuccess(`Order placed successfully! Order ID: ${order.id}`);
      orderForm.reset();
      
      // Refresh data
      loadTradingData();
    } catch (error: any) {
      setOrderError(error.message || 'Failed to place order');
    }
  };

  const cancelOrder = async (orderId: string) => {
    try {
      await tradingApi.cancelOrder(orderId);
      setOrderSuccess('Order cancelled successfully!');
      loadTradingData();
    } catch (error: any) {
      setOrderError(error.message || 'Failed to cancel order');
    }
  };

  const closePosition = async (symbol: string) => {
    try {
      await tradingApi.closePosition(symbol);
      setOrderSuccess(`Position ${symbol} closed successfully!`);
      loadTradingData();
    } catch (error: any) {
      setOrderError(error.message || `Failed to close position ${symbol}`);
    }
  };

  if (loading || isLoading) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center min-h-96">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
        </div>
      </DashboardLayout>
    );
  }

  const getStatusIcon = (status: string) => {
    switch (status.toLowerCase()) {
      case 'filled':
        return <CheckCircleIcon className="h-5 w-5 text-green-400" />;
      case 'cancelled':
      case 'rejected':
        return <XCircleIcon className="h-5 w-5 text-red-400" />;
      case 'pending_new':
      case 'new':
        return <ClockIcon className="h-5 w-5 text-yellow-400" />;
      default:
        return <InformationCircleIcon className="h-5 w-5 text-blue-400" />;
    }
  };

  const tabs = [
    { id: 'trade', label: 'Place Order' },
    { id: 'positions', label: 'Positions' },
    { id: 'orders', label: 'Orders' },
    { id: 'account', label: 'Account' },
  ];

  return (
    <DashboardLayout>
      <div className="p-6">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white">Trading Dashboard</h1>
          <p className="text-slate-400 mt-2">Manage your trades and positions</p>
        </div>

        {/* Account Overview Cards */}
        {account && (
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6"
            >
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-slate-400 text-sm">Portfolio Value</p>
                  <p className="text-2xl font-bold text-white">
                    ${parseFloat(account.portfolio_value).toLocaleString()}
                  </p>
                </div>
                <BanknotesIcon className="h-8 w-8 text-blue-400" />
              </div>
            </motion.div>

            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.1 }}
              className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6"
            >
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-slate-400 text-sm">Buying Power</p>
                  <p className="text-2xl font-bold text-white">
                    ${parseFloat(account.buying_power).toLocaleString()}
                  </p>
                </div>
                <ChartBarIcon className="h-8 w-8 text-green-400" />
              </div>
            </motion.div>

            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
              className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6"
            >
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-slate-400 text-sm">Cash</p>
                  <p className="text-2xl font-bold text-white">
                    ${parseFloat(account.cash).toLocaleString()}
                  </p>
                </div>
                <BanknotesIcon className="h-8 w-8 text-yellow-400" />
              </div>
            </motion.div>

            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.3 }}
              className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6"
            >
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-slate-400 text-sm">Day Trades</p>
                  <p className="text-2xl font-bold text-white">{account.day_trade_count}</p>
                  {account.pattern_day_trader && (
                    <p className="text-xs text-yellow-400">PDT</p>
                  )}
                </div>
                <ExclamationTriangleIcon className="h-8 w-8 text-orange-400" />
              </div>
            </motion.div>
          </div>
        )}

        {/* Success/Error Messages */}
        {orderSuccess && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className="mb-6 p-4 bg-green-500/20 border border-green-500/30 rounded-lg text-green-200 flex items-center"
          >
            <CheckCircleIcon className="h-5 w-5 mr-2" />
            {orderSuccess}
          </motion.div>
        )}

        {orderError && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className="mb-6 p-4 bg-red-500/20 border border-red-500/30 rounded-lg text-red-200 flex items-center"
          >
            <XCircleIcon className="h-5 w-5 mr-2" />
            {orderError}
          </motion.div>
        )}

        {/* Tabs */}
        <div className="mb-6">
          <div className="flex space-x-1 bg-slate-800/30 p-1 rounded-lg">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex-1 px-4 py-2 text-sm rounded-md transition-colors ${
                  activeTab === tab.id
                    ? 'bg-slate-700 text-white'
                    : 'text-slate-400 hover:text-white hover:bg-slate-700/50'
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>
        </div>

        {/* Tab Content */}
        <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6">
          {activeTab === 'trade' && (
            <div>
              <h2 className="text-xl font-semibold text-white mb-6">Place Order</h2>
              <form onSubmit={orderForm.handleSubmit(onOrderSubmit)} className="space-y-6">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div>
                    <label className="block text-sm font-medium text-slate-300 mb-2">
                      Symbol
                    </label>
                    <input
                      {...orderForm.register('symbol')}
                      type="text"
                      placeholder="e.g., AAPL"
                      className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 uppercase"
                    />
                    {orderForm.formState.errors.symbol && (
                      <p className="mt-1 text-sm text-red-400">
                        {orderForm.formState.errors.symbol.message}
                      </p>
                    )}
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-slate-300 mb-2">
                      Quantity
                    </label>
                    <input
                      {...orderForm.register('quantity', { valueAsNumber: true })}
                      type="number"
                      step="0.001"
                      placeholder="0"
                      className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                    {orderForm.formState.errors.quantity && (
                      <p className="mt-1 text-sm text-red-400">
                        {orderForm.formState.errors.quantity.message}
                      </p>
                    )}
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  <div>
                    <label className="block text-sm font-medium text-slate-300 mb-2">
                      Side
                    </label>
                    <select
                      {...orderForm.register('side')}
                      className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="buy">Buy</option>
                      <option value="sell">Sell</option>
                    </select>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-slate-300 mb-2">
                      Order Type
                    </label>
                    <select
                      {...orderForm.register('order_type')}
                      className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="market">Market</option>
                      <option value="limit">Limit</option>
                      <option value="stop">Stop</option>
                      <option value="stop_limit">Stop Limit</option>
                    </select>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-slate-300 mb-2">
                      Time in Force
                    </label>
                    <select
                      {...orderForm.register('time_in_force')}
                      className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="day">Day</option>
                      <option value="gtc">Good Till Cancelled</option>
                      <option value="ioc">Immediate or Cancel</option>
                      <option value="fok">Fill or Kill</option>
                    </select>
                  </div>
                </div>

                {(watchOrderType === 'limit' || watchOrderType === 'stop_limit') && (
                  <div>
                    <label className="block text-sm font-medium text-slate-300 mb-2">
                      Limit Price
                    </label>
                    <input
                      {...orderForm.register('limit_price', { valueAsNumber: true })}
                      type="number"
                      step="0.01"
                      placeholder="0.00"
                      className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </div>
                )}

                {(watchOrderType === 'stop' || watchOrderType === 'stop_limit') && (
                  <div>
                    <label className="block text-sm font-medium text-slate-300 mb-2">
                      Stop Price
                    </label>
                    <input
                      {...orderForm.register('stop_price', { valueAsNumber: true })}
                      type="number"
                      step="0.01"
                      placeholder="0.00"
                      className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </div>
                )}

                <div className="flex justify-end space-x-4">
                  <button
                    type="button"
                    onClick={() => orderForm.reset()}
                    className="px-6 py-3 text-slate-400 hover:text-white transition-colors"
                  >
                    Reset
                  </button>
                  <button
                    type="submit"
                    className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                  >
                    Place Order
                  </button>
                </div>
              </form>
            </div>
          )}

          {activeTab === 'positions' && (
            <div>
              <h2 className="text-xl font-semibold text-white mb-6">Current Positions</h2>
              {positions.length === 0 ? (
                <p className="text-slate-400 text-center py-8">No positions found</p>
              ) : (
                <div className="space-y-4">
                  {positions.map((position, index) => {
                    const isProfit = parseFloat(position.unrealized_pl) >= 0;
                    return (
                      <motion.div
                        key={position.symbol}
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ delay: index * 0.05 }}
                        className="bg-slate-700/50 border border-slate-600 rounded-lg p-4"
                      >
                        <div className="flex items-center justify-between">
                          <div className="flex-1">
                            <div className="flex items-center space-x-3">
                              <h3 className="text-lg font-semibold text-white">{position.symbol}</h3>
                              <span className="text-slate-400">{position.quantity} shares</span>
                            </div>
                            <div className="mt-2 grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                              <div>
                                <p className="text-slate-400">Market Value</p>
                                <p className="text-white font-medium">
                                  ${parseFloat(position.market_value).toLocaleString()}
                                </p>
                              </div>
                              <div>
                                <p className="text-slate-400">Cost Basis</p>
                                <p className="text-white font-medium">
                                  ${parseFloat(position.cost_basis).toLocaleString()}
                                </p>
                              </div>
                              <div>
                                <p className="text-slate-400">Unrealized P&L</p>
                                <p className={`font-medium ${isProfit ? 'text-green-400' : 'text-red-400'}`}>
                                  {isProfit ? '+' : ''}${parseFloat(position.unrealized_pl).toLocaleString()}
                                  ({parseFloat(position.unrealized_plpc).toFixed(2)}%)
                                </p>
                              </div>
                              <div>
                                <p className="text-slate-400">Current Price</p>
                                <p className="text-white font-medium">
                                  ${parseFloat(position.current_price).toFixed(2)}
                                </p>
                              </div>
                            </div>
                          </div>
                          <button
                            onClick={() => closePosition(position.symbol)}
                            className="ml-4 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
                          >
                            Close
                          </button>
                        </div>
                      </motion.div>
                    );
                  })}
                </div>
              )}
            </div>
          )}

          {activeTab === 'orders' && (
            <div>
              <h2 className="text-xl font-semibold text-white mb-6">Recent Orders</h2>
              {orders.length === 0 ? (
                <p className="text-slate-400 text-center py-8">No orders found</p>
              ) : (
                <div className="space-y-4">
                  {orders.map((order, index) => (
                    <motion.div
                      key={order.id}
                      initial={{ opacity: 0, y: 20 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: index * 0.05 }}
                      className="bg-slate-700/50 border border-slate-600 rounded-lg p-4"
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex-1">
                          <div className="flex items-center space-x-3">
                            {getStatusIcon(order.status)}
                            <h3 className="text-lg font-semibold text-white">{order.symbol}</h3>
                            <span className={`px-2 py-1 rounded text-xs font-medium ${
                              order.side === 'buy' ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'
                            }`}>
                              {order.side.toUpperCase()}
                            </span>
                            <span className="text-slate-400 text-sm">{order.order_type.toUpperCase()}</span>
                          </div>
                          <div className="mt-2 grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                            <div>
                              <p className="text-slate-400">Quantity</p>
                              <p className="text-white">{order.quantity}</p>
                            </div>
                            <div>
                              <p className="text-slate-400">Filled</p>
                              <p className="text-white">{order.filled_quantity}</p>
                            </div>
                            <div>
                              <p className="text-slate-400">Status</p>
                              <p className="text-white capitalize">{order.status.replace('_', ' ')}</p>
                            </div>
                            <div>
                              <p className="text-slate-400">Submitted</p>
                              <p className="text-white">
                                {new Date(order.submitted_at).toLocaleString()}
                              </p>
                            </div>
                          </div>
                        </div>
                        {(order.status === 'new' || order.status === 'pending_new') && (
                          <button
                            onClick={() => cancelOrder(order.id)}
                            className="ml-4 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
                          >
                            Cancel
                          </button>
                        )}
                      </div>
                    </motion.div>
                  ))}
                </div>
              )}
            </div>
          )}

          {activeTab === 'account' && account && (
            <div>
              <h2 className="text-xl font-semibold text-white mb-6">Account Details</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="space-y-4">
                  <div>
                    <p className="text-slate-400 text-sm">Account Number</p>
                    <p className="text-white font-medium">{account.account_number}</p>
                  </div>
                  <div>
                    <p className="text-slate-400 text-sm">Status</p>
                    <p className="text-white font-medium capitalize">{account.status}</p>
                  </div>
                  <div>
                    <p className="text-slate-400 text-sm">Currency</p>
                    <p className="text-white font-medium">{account.currency}</p>
                  </div>
                  <div>
                    <p className="text-slate-400 text-sm">Trading Blocked</p>
                    <p className={`font-medium ${account.trading_blocked ? 'text-red-400' : 'text-green-400'}`}>
                      {account.trading_blocked ? 'Yes' : 'No'}
                    </p>
                  </div>
                </div>
                <div className="space-y-4">
                  <div>
                    <p className="text-slate-400 text-sm">Portfolio Value</p>
                    <p className="text-white font-medium">${parseFloat(account.portfolio_value).toLocaleString()}</p>
                  </div>
                  <div>
                    <p className="text-slate-400 text-sm">Equity</p>
                    <p className="text-white font-medium">${parseFloat(account.equity).toLocaleString()}</p>
                  </div>
                  <div>
                    <p className="text-slate-400 text-sm">Buying Power</p>
                    <p className="text-white font-medium">${parseFloat(account.buying_power).toLocaleString()}</p>
                  </div>
                  <div>
                    <p className="text-slate-400 text-sm">Pattern Day Trader</p>
                    <p className={`font-medium ${account.pattern_day_trader ? 'text-yellow-400' : 'text-green-400'}`}>
                      {account.pattern_day_trader ? 'Yes' : 'No'}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </DashboardLayout>
  );
}
