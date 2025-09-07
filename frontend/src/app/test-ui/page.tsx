'use client';

import { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { 
  Activity,
  TrendingUp,
  TrendingDown,
  PieChart,
  Shield,
  Bell,
  Search,
  Star,
  ChevronDown,
  ChevronUp,
  Plus,
  Minus
} from 'lucide-react';

export default function TestUIPage() {
  const [isExpanded, setIsExpanded] = useState(false);
  const [currentTime, setCurrentTime] = useState('');
  const [stockPrices, setStockPrices] = useState<Record<string, { price: number; change: number }>>({});

  useEffect(() => {
    // Set initial time and prices on client side only
    setCurrentTime(new Date().toLocaleTimeString());
    
    // Generate initial stock prices
    const symbols = ['AAPL', 'GOOGL', 'MSFT', 'TSLA', 'AMZN', 'NVDA', 'META', 'JPM'];
    const initialPrices: Record<string, { price: number; change: number }> = {};
    symbols.forEach(symbol => {
      initialPrices[symbol] = {
        price: Math.random() * 500 + 50,
        change: Math.random() * 5
      };
    });
    setStockPrices(initialPrices);

    // Update time every second
    const interval = setInterval(() => {
      setCurrentTime(new Date().toLocaleTimeString());
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  return (
    <div className="min-h-screen bg-gray-900 dark">
      {/* Sidebar */}
      <aside className="fixed top-0 left-0 z-50 h-full bg-black border-r border-red-500/20 shadow-2xl w-80">
        <div className="flex flex-col h-full">
          {/* Header */}
          <div className="flex items-center justify-between p-6 border-b border-red-500/20">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-gradient-to-r from-red-600 to-black rounded-xl">
                <Activity className="w-6 h-6 text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold text-white">RiskMonitor</h1>
                <p className="text-xs text-gray-400">Financial Analytics</p>
              </div>
            </div>
          </div>

          {/* User Profile */}
          <div className="p-6 border-b border-red-500/20">
            <div className="p-4 bg-red-500/10 backdrop-blur-sm border border-red-500/20 rounded-xl">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-gradient-to-r from-red-600 to-black rounded-full flex items-center justify-center">
                  <span className="text-white font-bold">DT</span>
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-white truncate">Demo Trader</p>
                  <p className="text-xs text-gray-300 truncate">Premium Account</p>
                </div>
                <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
              </div>
            </div>
          </div>

          {/* Navigation */}
          <nav className="flex-1 p-6 space-y-2 overflow-y-auto">
            {[
              { name: 'Dashboard', icon: Activity, description: 'Overview & Analytics', active: false },
              { name: 'Portfolio', icon: PieChart, description: 'Manage Holdings', active: true },
              { name: 'Trading', icon: TrendingUp, description: 'Execute Trades', active: false },
              { name: 'Risk Analysis', icon: Shield, description: 'Risk Metrics', active: false },
              { name: 'Alerts', icon: Bell, description: 'Notifications', active: false },
            ].map((item) => (
              <motion.div
                key={item.name}
                whileHover={{ x: 4 }}
                whileTap={{ scale: 0.98 }}
                className={`group flex items-center gap-3 p-3 rounded-xl transition-all duration-200 ${
                  item.active
                    ? "bg-red-600 text-white shadow-lg"
                    : "hover:bg-red-500/10 text-gray-300 hover:text-white"
                }`}
              >
                <div className={`p-2 rounded-lg transition-colors ${
                  item.active 
                    ? "bg-white/20" 
                    : "bg-gray-800 group-hover:bg-red-500/20"
                }`}>
                  <item.icon className="w-4 h-4" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium truncate">{item.name}</p>
                  <p className={`text-xs truncate ${
                    item.active ? "text-white/80" : "text-gray-400"
                  }`}>
                    {item.description}
                  </p>
                </div>
              </motion.div>
            ))}
          </nav>

          {/* Quick Actions */}
          <div className="p-6 border-t border-red-500/20">
            <h3 className="text-sm font-medium text-white mb-3">Quick Actions</h3>
            <div className="grid grid-cols-2 gap-2">
              <motion.button
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                className="flex flex-col items-center gap-2 p-3 rounded-lg bg-gray-800 hover:bg-red-500/20 transition-colors"
              >
                <Plus className="w-4 h-4 text-green-400" />
                <span className="text-xs font-medium text-gray-300">Buy</span>
              </motion.button>
              <motion.button
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                className="flex flex-col items-center gap-2 p-3 rounded-lg bg-gray-800 hover:bg-red-500/20 transition-colors"
              >
                <Minus className="w-4 h-4 text-red-400" />
                <span className="text-xs font-medium text-gray-300">Sell</span>
              </motion.button>
            </div>
          </div>
        </div>
      </aside>

      {/* Main content */}
      <div className="pl-80">
        {/* Top bar */}
        <header className="sticky top-0 z-40 bg-black/80 backdrop-blur-xl border-b border-red-500/20">
          <div className="flex items-center justify-between px-6 py-4">
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-3 px-4 py-2 bg-gray-800 rounded-xl border border-red-500/20">
                <Search className="w-4 h-4 text-gray-400" />
                <input
                  type="text"
                  placeholder="Search stocks, portfolios..."
                  className="bg-transparent text-sm text-gray-300 placeholder-gray-400 outline-none w-64 focus:ring-2 focus:ring-red-500"
                />
              </div>
            </div>

            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <div className="flex items-center gap-2 px-3 py-1 bg-red-500/10 backdrop-blur-sm border border-red-500/20 rounded-full">
                  <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
                  <span className="text-sm font-medium text-green-400">Market Open</span>
                </div>
              </div>

              <div className="flex items-center gap-2 px-4 py-2 bg-black/40 backdrop-blur-sm border border-red-500/20 rounded-xl">
                <PieChart className="w-4 h-4 text-red-400" />
                <div>
                  <p className="text-sm font-bold text-white">$929,000</p>
                  <p className="text-xs text-green-400">+2.4%</p>
                </div>
              </div>

              <motion.button
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                className="relative p-2 rounded-lg bg-gray-800 hover:bg-red-500/20 transition-colors"
              >
                <Bell className="w-5 h-5 text-gray-400" />
                <div className="absolute -top-1 -right-1 w-3 h-3 bg-red-500 rounded-full flex items-center justify-center">
                  <span className="text-xs text-white font-bold">3</span>
                </div>
              </motion.button>
            </div>
          </div>
        </header>

        {/* Page content */}
        <main className="p-6">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            className="space-y-6"
          >
            {/* Header */}
            <div>
              <h1 className="text-3xl font-bold text-white">UI Theme Test</h1>
              <p className="text-gray-400 mt-1">Black + Red/White Theme Showcase</p>
            </div>

            {/* Stats Cards */}
            <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-6">
              {[
                { title: 'Total Portfolio Value', value: '$929,000', change: '+2.4%', positive: true },
                { title: 'Total P&L', value: '$22,340', change: '+2.4%', positive: true },
                { title: 'Active Portfolios', value: '3', change: null, positive: null },
                { title: 'Active Alerts', value: '2', change: null, positive: false },
              ].map((stat, index) => (
                <motion.div
                  key={index}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: index * 0.1 }}
                  className="bg-black/40 backdrop-blur-sm rounded-xl p-6 border border-red-500/20"
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-gray-400 text-sm">{stat.title}</p>
                      <p className="text-2xl font-bold text-white mt-1">{stat.value}</p>
                      {stat.change && (
                        <div className="flex items-center gap-1 mt-2">
                          {stat.positive ? (
                            <TrendingUp className="w-4 h-4 text-green-400" />
                          ) : (
                            <TrendingDown className="w-4 h-4 text-red-400" />
                          )}
                          <span className={`text-sm font-medium ${
                            stat.positive ? 'text-green-400' : 'text-red-400'
                          }`}>
                            {stat.change}
                          </span>
                        </div>
                      )}
                    </div>
                    <div className="p-3 bg-gradient-to-r from-red-600 to-black rounded-xl">
                      <PieChart className="w-6 h-6 text-white" />
                    </div>
                  </div>
                </motion.div>
              ))}
            </div>

            {/* Price Monitor Component */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="bg-black/40 backdrop-blur-sm rounded-xl border border-red-500/20 overflow-hidden"
            >
              <div className="p-4 border-b border-red-500/20">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-gradient-to-r from-red-600 to-black rounded-lg">
                      <Activity className="w-4 h-4 text-white" />
                    </div>
                    <div>
                      <h3 className="text-lg font-semibold text-white">Live Market Data</h3>
                      <p className="text-xs text-gray-400">8 symbols • Last updated: {currentTime || 'Loading...'}</p>
                    </div>
                  </div>
                  
                  <motion.button
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                    onClick={() => setIsExpanded(!isExpanded)}
                    className="flex items-center gap-2 px-3 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors"
                  >
                    <span className="text-sm font-medium">
                      {isExpanded ? 'Collapse' : 'Expand'}
                    </span>
                    {isExpanded ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
                  </motion.button>
                </div>
              </div>

              <div className="p-4">
                <div className="flex items-center gap-4 overflow-x-auto">
                  {['AAPL', 'GOOGL', 'MSFT', 'TSLA', 'AMZN', 'NVDA', 'META', 'JPM'].map((symbol, index) => (
                    <motion.div
                      key={symbol}
                      initial={{ opacity: 0, x: 20 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ delay: index * 0.05 }}
                      className="flex-shrink-0 flex items-center gap-3 p-3 bg-gray-800 rounded-lg min-w-[160px]"
                    >
                      <div className="flex items-center gap-2">
                        <Star className="w-3 h-3 text-yellow-400" fill="currentColor" />
                        <span className="text-sm font-medium text-white">{symbol}</span>
                      </div>
                      <div className="text-right">
                        <p className="text-sm font-bold text-white">
                          ${stockPrices[symbol]?.price.toFixed(2) || '0.00'}
                        </p>
                        <div className="flex items-center gap-1">
                          <TrendingUp className="w-3 h-3 text-green-400" />
                          <span className="text-xs font-medium text-green-400">
                            +{stockPrices[symbol]?.change.toFixed(2) || '0.00'}%
                          </span>
                        </div>
                      </div>
                    </motion.div>
                  ))}
                </div>
              </div>
            </motion.div>

            {/* Portfolio Cards */}
            <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
              {[
                { name: 'Growth Portfolio', value: '$450,000', change: '+3.2%', positions: 12 },
                { name: 'Dividend Portfolio', value: '$320,000', change: '+1.8%', positions: 8 },
                { name: 'Tech Portfolio', value: '$159,000', change: '+5.1%', positions: 6 },
              ].map((portfolio, index) => (
                <motion.div
                  key={index}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: index * 0.1 }}
                  className="bg-black/40 backdrop-blur-sm rounded-xl p-6 border border-red-500/20 hover:border-red-500/40 transition-colors"
                >
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="text-lg font-semibold text-white">{portfolio.name}</h3>
                    <div className="p-2 bg-gradient-to-r from-red-600 to-black rounded-lg">
                      <PieChart className="w-4 h-4 text-white" />
                    </div>
                  </div>
                  
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="text-gray-400">Total Value</span>
                      <span className="text-white font-bold">{portfolio.value}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-gray-400">Change</span>
                      <div className="flex items-center gap-1">
                        <TrendingUp className="w-4 h-4 text-green-400" />
                        <span className="text-green-400 font-medium">{portfolio.change}</span>
                      </div>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-gray-400">Positions</span>
                      <span className="text-white">{portfolio.positions}</span>
                    </div>
                  </div>
                  
                  <motion.button
                    whileHover={{ scale: 1.02 }}
                    whileTap={{ scale: 0.98 }}
                    className="w-full mt-4 px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors"
                  >
                    View Details
                  </motion.button>
                </motion.div>
              ))}
            </div>
          </motion.div>
        </main>
      </div>
    </div>
  );
}
