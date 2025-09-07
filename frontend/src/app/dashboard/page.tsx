'use client';

import { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import Navigation from '@/components/Navigation';
import { 
  TrendingUp, 
  TrendingDown, 
  DollarSign, 
  PieChart, 
  Activity,
  Bell,
  Search,
  Filter,
  MoreVertical,
  ArrowUpRight,
  ArrowDownRight,
  Star,
  Eye,
  Zap,
  AlertTriangle,
  CheckCircle,
  Clock,
  Globe,
  Shield,
  BarChart3,
  Wallet,
  Target,
  Gauge
} from 'lucide-react';
import { withAuth, useAuth } from '@/contexts/AuthContext';

// Mock data - replace with real API calls
const portfolioStats = [
  {
    id: 'total-value',
    title: 'Total Portfolio Value',
    value: '$2,847,392',
    change: '+12.5%',
    changeValue: '+$318,492',
    positive: true,
    icon: DollarSign,
    gradient: 'from-primary to-primary/80',
    description: '24h change'
  },
  {
    id: 'pnl',
    title: 'Unrealized P&L',
    value: '$145,892',
    change: '+8.2%',
    changeValue: '+$11,045',
    positive: true,
    icon: TrendingUp,
    gradient: 'from-success to-success/80',
    description: 'Since last week'
  },
  {
    id: 'positions',
    title: 'Active Positions',
    value: '24',
    change: '+3',
    changeValue: 'New positions',
    positive: true,
    icon: PieChart,
    gradient: 'from-warning to-warning/80',
    description: 'Across 6 sectors'
  },
  {
    id: 'risk-score',
    title: 'Risk Score',
    value: '7.2/10',
    change: '-0.3',
    changeValue: 'Lower risk',
    positive: true,
    icon: Shield,
    gradient: 'from-danger to-danger/80',
    description: 'Risk assessment'
  },
];

const marketData = [
  { symbol: 'AAPL', name: 'Apple Inc.', price: 175.43, change: 2.34, changePercent: 1.35, volume: '52.3M', sector: 'Technology' },
  { symbol: 'GOOGL', name: 'Alphabet Inc.', price: 2847.92, change: -15.67, changePercent: -0.55, volume: '1.2M', sector: 'Technology' },
  { symbol: 'MSFT', name: 'Microsoft Corp.', price: 378.85, change: 8.92, changePercent: 2.41, volume: '28.7M', sector: 'Technology' },
  { symbol: 'TSLA', name: 'Tesla Inc.', price: 248.50, change: -12.30, changePercent: -4.72, volume: '89.1M', sector: 'Automotive' },
  { symbol: 'NVDA', name: 'NVIDIA Corp.', price: 456.78, change: 23.45, changePercent: 5.41, volume: '45.6M', sector: 'Technology' },
  { symbol: 'AMZN', name: 'Amazon.com Inc.', price: 3247.89, change: 45.67, changePercent: 1.43, volume: '3.4M', sector: 'E-commerce' },
];

const recentTransactions = [
  { id: 1, type: 'BUY', symbol: 'AAPL', quantity: 100, price: 175.43, time: '2 min ago', status: 'completed', total: 17543 },
  { id: 2, type: 'SELL', symbol: 'TSLA', quantity: 50, price: 248.50, time: '15 min ago', status: 'completed', total: 12425 },
  { id: 3, type: 'BUY', symbol: 'NVDA', quantity: 25, price: 456.78, time: '1 hour ago', status: 'pending', total: 11419.5 },
  { id: 4, type: 'SELL', symbol: 'GOOGL', quantity: 10, price: 2847.92, time: '2 hours ago', status: 'completed', total: 28479.2 },
];

const alerts = [
  { id: 1, type: 'warning', title: 'Position Concentration Alert', message: 'TSLA position exceeds 15% of portfolio', time: '5 min ago', severity: 'medium' },
  { id: 2, type: 'success', title: 'Stop Loss Executed', message: 'Stop loss triggered for AAPL position at $170', time: '1 hour ago', severity: 'low' },
  { id: 3, type: 'info', title: 'Market Volatility', message: 'VIX increased by 12% - consider hedging', time: '2 hours ago', severity: 'medium' },
  { id: 4, type: 'danger', title: 'Risk Threshold Breach', message: 'Portfolio VaR exceeded 95% confidence level', time: '3 hours ago', severity: 'high' },
];

function DashboardPage() {
  const { user } = useAuth();
  const [currentTime, setCurrentTime] = useState('');
  const [marketStatus, setMarketStatus] = useState('OPEN');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedTimeframe, setSelectedTimeframe] = useState('24h');

  useEffect(() => {
    const updateTime = () => {
      setCurrentTime(new Date().toLocaleTimeString('en-US', {
        hour12: false,
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      }));
    };
    
    updateTime();
    const interval = setInterval(updateTime, 1000);
    
    // Simulate market status
    const hour = new Date().getHours();
    setMarketStatus(hour >= 9 && hour < 16 ? 'OPEN' : 'CLOSED');
    
    return () => clearInterval(interval);
  }, []);

  const filteredMarketData = marketData.filter(stock =>
    stock.symbol.toLowerCase().includes(searchQuery.toLowerCase()) ||
    stock.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="min-h-screen bg-background">
      <Navigation />
      
      {/* Main Content */}
      <main className="lg:pl-72 min-h-screen">
        {/* Header */}
        <motion.header 
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="sticky top-0 z-30 glass-card border-b border-white/10 p-6"
        >
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-6">
              <div>
                <h1 className="text-3xl font-bold gradient-text">Dashboard</h1>
                <p className="text-sm text-foreground-muted mt-1">
                  Welcome back, John • {currentTime}
                </p>
              </div>
            </div>
            
            <div className="flex items-center gap-4">
              {/* Market Status */}
              <motion.div 
                whileHover={{ scale: 1.02 }}
                className={`flex items-center gap-2 px-4 py-2 glass-card rounded-xl ${
                  marketStatus === 'OPEN' ? 'border-success/30' : 'border-danger/30'
                }`}
              >
                <div className={`w-3 h-3 rounded-full ${
                  marketStatus === 'OPEN' ? 'bg-success animate-pulse' : 'bg-danger'
                }`} />
                <span className={`text-sm font-medium ${
                  marketStatus === 'OPEN' ? 'text-success' : 'text-danger'
                }`}>
                  Market {marketStatus}
                </span>
              </motion.div>
              
              {/* Search */}
              <div className="relative hidden md:block">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-foreground-muted" />
                <input
                  type="text"
                  placeholder="Search stocks, portfolios..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="input-primary pl-10 pr-4 w-80"
                />
              </div>
              
              {/* Notifications */}
              <motion.button 
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                className="relative p-3 glass-card rounded-xl hover:bg-glass-strong transition-all duration-200"
              >
                <Bell className="w-5 h-5 text-foreground" />
                <motion.div 
                  initial={{ scale: 0 }}
                  animate={{ scale: 1 }}
                  className="absolute -top-1 -right-1 w-5 h-5 bg-danger rounded-full flex items-center justify-center"
                >
                  <span className="text-2xs text-white font-bold">{alerts.length}</span>
                </motion.div>
              </motion.button>
            </div>
          </div>
        </motion.header>

        {/* Dashboard Content */}
        <div className="p-6 space-y-8">
          {/* Portfolio Stats Grid */}
          <motion.section
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
          >
            <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6">
              {portfolioStats.map((stat, index) => {
                const Icon = stat.icon;
                return (
                  <motion.div 
                    key={stat.id}
                    initial={{ opacity: 0, scale: 0.9 }}
                    animate={{ opacity: 1, scale: 1 }}
                    transition={{ delay: 0.1 + index * 0.05 }}
                    whileHover={{ scale: 1.02, y: -5 }}
                    className="stat-card group relative overflow-hidden"
                  >
                    <div className={`absolute top-0 left-0 w-full h-1 bg-gradient-to-r ${stat.gradient}`} />
                    
                    <div className="flex items-center justify-between mb-6">
                      <div className={`w-14 h-14 rounded-2xl bg-gradient-to-r ${stat.gradient} flex items-center justify-center shadow-lg group-hover:shadow-xl transition-all duration-300`}>
                        <Icon className="w-7 h-7 text-white" />
                      </div>
                      <motion.button
                        whileHover={{ scale: 1.1 }}
                        whileTap={{ scale: 0.9 }}
                        className="p-2 rounded-lg hover:bg-glass-strong transition-colors"
                      >
                        <MoreVertical className="w-4 h-4 text-foreground-muted" />
                      </motion.button>
                    </div>
                    
                    <div className="space-y-3">
                      <div>
                        <p className="text-sm text-foreground-muted mb-1">{stat.title}</p>
                        <p className="text-3xl font-bold text-foreground">{stat.value}</p>
                      </div>
                      
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          {stat.positive ? (
                            <ArrowUpRight className="w-4 h-4 text-success" />
                          ) : (
                            <ArrowDownRight className="w-4 h-4 text-danger" />
                          )}
                          <span className={`text-sm font-medium ${
                            stat.positive ? 'text-success' : 'text-danger'
                          }`}>
                            {stat.change}
                          </span>
                        </div>
                        <span className="text-2xs text-foreground-muted">
                          {stat.description}
                        </span>
                      </div>
                      
                      <p className="text-xs text-foreground-muted">
                        {stat.changeValue}
                      </p>
                    </div>
                  </motion.div>
                );
              })}
            </div>
          </motion.section>

          {/* Market Data & Alerts */}
          <div className="grid grid-cols-1 xl:grid-cols-3 gap-8">
            {/* Market Data */}
            <motion.section
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: 0.2 }}
              className="xl:col-span-2"
            >
              <div className="chart-card">
                <div className="flex items-center justify-between mb-6">
                  <div>
                    <h3 className="text-xl font-bold text-foreground mb-1">Live Market Data</h3>
                    <p className="text-sm text-foreground-muted">Real-time stock prices and movements</p>
                  </div>
                  <div className="flex items-center gap-2">
                    <motion.button
                      whileHover={{ scale: 1.05 }}
                      whileTap={{ scale: 0.95 }}
                      className="p-2 rounded-lg glass-card hover:bg-glass-strong transition-colors"
                    >
                      <Filter className="w-4 h-4 text-foreground-muted" />
                    </motion.button>
                    <motion.button
                      whileHover={{ scale: 1.05 }}
                      whileTap={{ scale: 0.95 }}
                      className="p-2 rounded-lg glass-card hover:bg-glass-strong transition-colors"
                    >
                      <Eye className="w-4 h-4 text-foreground-muted" />
                    </motion.button>
                  </div>
                </div>
                
                <div className="space-y-3">
                  {filteredMarketData.map((stock, index) => (
                    <motion.div 
                      key={stock.symbol}
                      initial={{ opacity: 0, y: 20 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: 0.3 + index * 0.05 }}
                      whileHover={{ scale: 1.01, x: 5 }}
                      className="flex items-center justify-between p-4 rounded-xl glass-card hover:bg-glass-strong transition-all duration-200 cursor-pointer group"
                    >
                      <div className="flex items-center gap-4">
                        <div className="w-12 h-12 rounded-xl bg-gradient-primary flex items-center justify-center group-hover:shadow-glow-primary/30 transition-all duration-300">
                          <Star className="w-6 h-6 text-white" />
                        </div>
                        <div>
                          <div className="flex items-center gap-2">
                            <p className="font-bold text-foreground">{stock.symbol}</p>
                            <span className="badge-neutral text-2xs">{stock.sector}</span>
                          </div>
                          <p className="text-xs text-foreground-muted truncate max-w-32">{stock.name}</p>
                          <p className="text-2xs text-foreground-subtle">Vol: {stock.volume}</p>
                        </div>
                      </div>
                      
                      <div className="text-right">
                        <p className="text-lg font-bold text-foreground">${stock.price.toFixed(2)}</p>
                        <div className="flex items-center gap-1 justify-end">
                          {stock.change > 0 ? (
                            <ArrowUpRight className="w-3 h-3 text-success" />
                          ) : (
                            <ArrowDownRight className="w-3 h-3 text-danger" />
                          )}
                          <span className={`text-sm font-medium ${
                            stock.change > 0 ? 'text-success' : 'text-danger'
                          }`}>
                            {stock.changePercent > 0 ? '+' : ''}{stock.changePercent.toFixed(2)}%
                          </span>
                        </div>
                        <p className="text-2xs text-foreground-muted">
                          {stock.change > 0 ? '+' : ''}${stock.change.toFixed(2)}
                        </p>
                      </div>
                    </motion.div>
                  ))}
                </div>
              </div>
            </motion.section>

            {/* Alerts Panel */}
            <motion.section
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: 0.3 }}
            >
              <div className="chart-card">
                <div className="flex items-center justify-between mb-6">
                  <div>
                    <h3 className="text-xl font-bold text-foreground mb-1">Risk Alerts</h3>
                    <p className="text-sm text-foreground-muted">Recent notifications</p>
                  </div>
                  <motion.button
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                    className="p-2 rounded-lg glass-card hover:bg-glass-strong transition-colors"
                  >
                    <Zap className="w-4 h-4 text-warning" />
                  </motion.button>
                </div>
                
                <div className="space-y-3 max-h-96 overflow-y-auto scrollbar-thin">
                  {alerts.map((alert, index) => (
                    <motion.div 
                      key={alert.id}
                      initial={{ opacity: 0, x: 20 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ delay: 0.4 + index * 0.05 }}
                      whileHover={{ scale: 1.02 }}
                      className={`p-4 rounded-xl border transition-all duration-200 cursor-pointer ${
                        alert.type === 'success' ? 'bg-success/5 border-success/20 hover:bg-success/10' :
                        alert.type === 'warning' ? 'bg-warning/5 border-warning/20 hover:bg-warning/10' :
                        alert.type === 'danger' ? 'bg-danger/5 border-danger/20 hover:bg-danger/10' :
                        'bg-primary/5 border-primary/20 hover:bg-primary/10'
                      }`}
                    >
                      <div className="flex items-start gap-3">
                        <div className={`p-2 rounded-lg ${
                          alert.type === 'success' ? 'bg-success/20' :
                          alert.type === 'warning' ? 'bg-warning/20' :
                          alert.type === 'danger' ? 'bg-danger/20' :
                          'bg-primary/20'
                        }`}>
                          {alert.type === 'success' ? <CheckCircle className="w-4 h-4 text-success" /> :
                           alert.type === 'warning' ? <AlertTriangle className="w-4 h-4 text-warning" /> :
                           alert.type === 'danger' ? <AlertTriangle className="w-4 h-4 text-danger" /> :
                           <Bell className="w-4 h-4 text-primary" />}
                        </div>
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium text-foreground mb-1">{alert.title}</p>
                          <p className="text-2xs text-foreground-muted leading-relaxed">{alert.message}</p>
                          <div className="flex items-center gap-2 mt-2">
                            <Clock className="w-3 h-3 text-foreground-subtle" />
                            <span className="text-2xs text-foreground-subtle">{alert.time}</span>
                          </div>
                        </div>
                      </div>
                    </motion.div>
                  ))}
                </div>
              </div>
            </motion.section>
          </div>

          {/* Recent Transactions */}
          <motion.section
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.4 }}
          >
            <div className="chart-card">
              <div className="flex items-center justify-between mb-6">
                <div>
                  <h3 className="text-xl font-bold text-foreground mb-1">Recent Transactions</h3>
                  <p className="text-sm text-foreground-muted">Latest trading activity</p>
                </div>
                <motion.button
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  className="btn-primary text-sm px-6 py-2"
                >
                  View All
                </motion.button>
              </div>
              
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-white/10">
                      <th className="text-left py-4 px-4 text-sm font-medium text-foreground-muted">Type</th>
                      <th className="text-left py-4 px-4 text-sm font-medium text-foreground-muted">Symbol</th>
                      <th className="text-left py-4 px-4 text-sm font-medium text-foreground-muted">Quantity</th>
                      <th className="text-left py-4 px-4 text-sm font-medium text-foreground-muted">Price</th>
                      <th className="text-left py-4 px-4 text-sm font-medium text-foreground-muted">Total</th>
                      <th className="text-left py-4 px-4 text-sm font-medium text-foreground-muted">Time</th>
                      <th className="text-left py-4 px-4 text-sm font-medium text-foreground-muted">Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {recentTransactions.map((transaction, index) => (
                      <motion.tr 
                        key={transaction.id}
                        initial={{ opacity: 0, y: 10 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ delay: 0.5 + index * 0.05 }}
                        className="table-row"
                      >
                        <td className="py-4 px-4">
                          <span className={`px-3 py-1 rounded-full text-2xs font-semibold ${
                            transaction.type === 'BUY' 
                              ? 'bg-success/20 text-success border border-success/30' 
                              : 'bg-danger/20 text-danger border border-danger/30'
                          }`}>
                            {transaction.type}
                          </span>
                        </td>
                        <td className="py-4 px-4 font-semibold text-foreground">{transaction.symbol}</td>
                        <td className="py-4 px-4 text-foreground">{transaction.quantity}</td>
                        <td className="py-4 px-4 text-foreground">${transaction.price.toFixed(2)}</td>
                        <td className="py-4 px-4 font-semibold text-foreground">${transaction.total.toLocaleString()}</td>
                        <td className="py-4 px-4 text-foreground-muted">{transaction.time}</td>
                        <td className="py-4 px-4">
                          <span className={`px-3 py-1 rounded-full text-2xs font-semibold ${
                            transaction.status === 'completed' 
                              ? 'bg-success/20 text-success border border-success/30' 
                              : 'bg-warning/20 text-warning border border-warning/30'
                          }`}>
                            {transaction.status}
                          </span>
                        </td>
                      </motion.tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </motion.section>
        </div>
      </main>
    </div>
  );
}

export default withAuth(DashboardPage);