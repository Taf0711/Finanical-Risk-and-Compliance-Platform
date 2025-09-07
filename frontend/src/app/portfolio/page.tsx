'use client';

import { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import Navigation from '@/components/Navigation';
import { 
  PieChart, 
  TrendingUp, 
  TrendingDown, 
  DollarSign, 
  Plus,
  Search,
  Filter,
  MoreVertical,
  ArrowUpRight,
  ArrowDownRight,
  Target,
  Zap,
  AlertTriangle,
  CheckCircle,
  BarChart3,
  Activity,
  Eye,
  Settings,
  Download,
  Upload,
  RefreshCw
} from 'lucide-react';
import { formatCurrency, formatPercentage, cn } from '@/lib/utils';

// Mock portfolio data
const portfolios = [
  {
    id: 1,
    name: 'Growth Portfolio',
    value: 2847392,
    change: 318492,
    changePercent: 12.5,
    positions: 24,
    riskScore: 7.2,
    allocation: {
      stocks: 75,
      bonds: 15,
      commodities: 5,
      cash: 5
    },
    topHoldings: [
      { symbol: 'AAPL', name: 'Apple Inc.', weight: 15.2, value: 432000, change: 2.1 },
      { symbol: 'MSFT', name: 'Microsoft Corp.', weight: 12.8, value: 364000, change: 1.8 },
      { symbol: 'GOOGL', name: 'Alphabet Inc.', weight: 10.5, value: 299000, change: -0.5 },
      { symbol: 'NVDA', name: 'NVIDIA Corp.', weight: 8.9, value: 253000, change: 4.2 },
      { symbol: 'TSLA', name: 'Tesla Inc.', weight: 7.3, value: 208000, change: -2.1 }
    ]
  },
  {
    id: 2,
    name: 'Conservative Portfolio',
    value: 1234567,
    change: -23456,
    changePercent: -1.9,
    positions: 18,
    riskScore: 3.8,
    allocation: {
      stocks: 40,
      bonds: 50,
      commodities: 5,
      cash: 5
    },
    topHoldings: [
      { symbol: 'VTI', name: 'Vanguard Total Stock', weight: 20.1, value: 248000, change: 0.8 },
      { symbol: 'BND', name: 'Vanguard Bond Index', weight: 25.4, value: 314000, change: -0.2 },
      { symbol: 'VOO', name: 'Vanguard S&P 500', weight: 15.7, value: 194000, change: 1.1 }
    ]
  }
];

const sectorAllocation = [
  { name: 'Technology', percentage: 35.2, value: 1001000, color: 'text-primary' },
  { name: 'Healthcare', percentage: 15.8, value: 450000, color: 'text-success' },
  { name: 'Financial', percentage: 12.4, value: 353000, color: 'text-warning' },
  { name: 'Consumer', percentage: 11.1, value: 316000, color: 'text-accent' },
  { name: 'Energy', percentage: 8.9, value: 253000, color: 'text-danger' },
  { name: 'Others', percentage: 16.6, value: 474000, color: 'text-foreground-muted' }
];

export default function PortfolioPage() {
  const [selectedPortfolio, setSelectedPortfolio] = useState(0);
  const [searchQuery, setSearchQuery] = useState('');
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [isRefreshing, setIsRefreshing] = useState(false);

  const currentPortfolio = portfolios[selectedPortfolio];

  const handleRefresh = async () => {
    setIsRefreshing(true);
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 2000));
    setIsRefreshing(false);
  };

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
                <h1 className="text-3xl font-bold gradient-text">Portfolio Management</h1>
                <p className="text-sm text-foreground-muted mt-1">
                  Manage your investment portfolios and track performance
                </p>
              </div>
            </div>
            
            <div className="flex items-center gap-4">
              {/* Refresh Button */}
              <motion.button
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                onClick={handleRefresh}
                disabled={isRefreshing}
                className="p-3 glass-card rounded-xl hover:bg-glass-strong transition-all duration-200 disabled:opacity-50"
              >
                <RefreshCw className={`w-5 h-5 text-foreground ${isRefreshing ? 'animate-spin' : ''}`} />
              </motion.button>
              
              {/* Search */}
              <div className="relative hidden md:block">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-foreground-muted" />
                <input
                  type="text"
                  placeholder="Search holdings..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="input-primary pl-10 pr-4 w-80"
                />
              </div>
              
              {/* Add Portfolio Button */}
              <motion.button
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                className="btn-primary flex items-center gap-2"
              >
                <Plus className="w-4 h-4" />
                <span>New Portfolio</span>
              </motion.button>
            </div>
          </div>
        </motion.header>

        {/* Content */}
        <div className="p-6 space-y-8">
          {/* Portfolio Selector */}
          <motion.section
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
          >
            <div className="flex items-center gap-4 mb-6">
              <h2 className="text-xl font-bold text-foreground">Select Portfolio</h2>
              <div className="flex items-center gap-2">
                {portfolios.map((_, index) => (
                  <motion.button
                    key={index}
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                    onClick={() => setSelectedPortfolio(index)}
                    className={`w-3 h-3 rounded-full transition-all duration-200 ${
                      selectedPortfolio === index ? 'bg-primary' : 'bg-foreground-muted/30'
                    }`}
                  />
                ))}
              </div>
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {portfolios.map((portfolio, index) => (
                <motion.div
                  key={portfolio.id}
                  initial={{ opacity: 0, scale: 0.9 }}
                  animate={{ opacity: 1, scale: 1 }}
                  transition={{ delay: 0.1 + index * 0.05 }}
                  whileHover={{ scale: 1.02, y: -5 }}
                  onClick={() => setSelectedPortfolio(index)}
                  className={`stat-card cursor-pointer transition-all duration-300 ${
                    selectedPortfolio === index ? 'ring-2 ring-primary shadow-glow-primary' : ''
                  }`}
                >
                  <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center gap-3">
                      <div className="w-12 h-12 rounded-xl bg-gradient-primary flex items-center justify-center">
                        <PieChart className="w-6 h-6 text-white" />
                      </div>
                      <div>
                        <h3 className="text-lg font-bold text-foreground">{portfolio.name}</h3>
                        <p className="text-sm text-foreground-muted">{portfolio.positions} positions</p>
                      </div>
                    </div>
                    <div className="text-right">
                      <p className="text-2xl font-bold text-foreground">
                        {formatCurrency(portfolio.value)}
                      </p>
                      <div className="flex items-center gap-1 justify-end">
                        {portfolio.change > 0 ? (
                          <ArrowUpRight className="w-4 h-4 text-success" />
                        ) : (
                          <ArrowDownRight className="w-4 h-4 text-danger" />
                        )}
                        <span className={`text-sm font-medium ${
                          portfolio.change > 0 ? 'text-success' : 'text-danger'
                        }`}>
                          {portfolio.change > 0 ? '+' : ''}{formatPercentage(portfolio.changePercent)}
                        </span>
                      </div>
                    </div>
                  </div>
                  
                  <div className="grid grid-cols-4 gap-4 text-center">
                    <div>
                      <p className="text-2xs text-foreground-muted mb-1">Risk Score</p>
                      <p className="text-lg font-bold text-foreground">{portfolio.riskScore}/10</p>
                    </div>
                    <div>
                      <p className="text-2xs text-foreground-muted mb-1">Stocks</p>
                      <p className="text-lg font-bold text-primary">{portfolio.allocation.stocks}%</p>
                    </div>
                    <div>
                      <p className="text-2xs text-foreground-muted mb-1">Bonds</p>
                      <p className="text-lg font-bold text-success">{portfolio.allocation.bonds}%</p>
                    </div>
                    <div>
                      <p className="text-2xs text-foreground-muted mb-1">Other</p>
                      <p className="text-lg font-bold text-warning">{portfolio.allocation.commodities + portfolio.allocation.cash}%</p>
                    </div>
                  </div>
                </motion.div>
              ))}
            </div>
          </motion.section>

          {/* Portfolio Overview */}
          <div className="grid grid-cols-1 xl:grid-cols-3 gap-8">
            {/* Holdings */}
            <motion.section
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: 0.2 }}
              className="xl:col-span-2"
            >
              <div className="chart-card">
                <div className="flex items-center justify-between mb-6">
                  <div>
                    <h3 className="text-xl font-bold text-foreground mb-1">Top Holdings</h3>
                    <p className="text-sm text-foreground-muted">{currentPortfolio.name}</p>
                  </div>
                  <div className="flex items-center gap-2">
                    <motion.button
                      whileHover={{ scale: 1.05 }}
                      whileTap={{ scale: 0.95 }}
                      className="p-2 rounded-lg glass-card hover:bg-glass-strong transition-colors"
                    >
                      <Download className="w-4 h-4 text-foreground-muted" />
                    </motion.button>
                    <motion.button
                      whileHover={{ scale: 1.05 }}
                      whileTap={{ scale: 0.95 }}
                      className="p-2 rounded-lg glass-card hover:bg-glass-strong transition-colors"
                    >
                      <Settings className="w-4 h-4 text-foreground-muted" />
                    </motion.button>
                  </div>
                </div>
                
                <div className="space-y-4">
                  {currentPortfolio.topHoldings.map((holding, index) => (
                    <motion.div
                      key={holding.symbol}
                      initial={{ opacity: 0, y: 20 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: 0.3 + index * 0.05 }}
                      whileHover={{ scale: 1.01, x: 5 }}
                      className="flex items-center justify-between p-4 rounded-xl glass-card hover:bg-glass-strong transition-all duration-200 cursor-pointer group"
                    >
                      <div className="flex items-center gap-4">
                        <div className="w-12 h-12 rounded-xl bg-gradient-primary flex items-center justify-center group-hover:shadow-glow-primary/30 transition-all duration-300">
                          <span className="text-white font-bold text-sm">{holding.symbol.slice(0, 2)}</span>
                        </div>
                        <div>
                          <div className="flex items-center gap-2">
                            <p className="font-bold text-foreground">{holding.symbol}</p>
                            <span className="badge-neutral text-2xs">{holding.weight.toFixed(1)}%</span>
                          </div>
                          <p className="text-sm text-foreground-muted truncate max-w-48">{holding.name}</p>
                        </div>
                      </div>
                      
                      <div className="text-right">
                        <p className="text-lg font-bold text-foreground">
                          {formatCurrency(holding.value)}
                        </p>
                        <div className="flex items-center gap-1 justify-end">
                          {holding.change > 0 ? (
                            <ArrowUpRight className="w-3 h-3 text-success" />
                          ) : (
                            <ArrowDownRight className="w-3 h-3 text-danger" />
                          )}
                          <span className={`text-sm font-medium ${
                            holding.change > 0 ? 'text-success' : 'text-danger'
                          }`}>
                            {holding.change > 0 ? '+' : ''}{holding.change.toFixed(1)}%
                          </span>
                        </div>
                      </div>
                    </motion.div>
                  ))}
                </div>
                
                <div className="mt-6 pt-4 border-t border-white/10">
                  <motion.button
                    whileHover={{ scale: 1.02 }}
                    whileTap={{ scale: 0.98 }}
                    className="w-full btn-secondary text-sm py-3"
                  >
                    View All Holdings ({currentPortfolio.positions})
                  </motion.button>
                </div>
              </div>
            </motion.section>

            {/* Sector Allocation */}
            <motion.section
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: 0.3 }}
            >
              <div className="chart-card">
                <div className="flex items-center justify-between mb-6">
                  <div>
                    <h3 className="text-xl font-bold text-foreground mb-1">Sector Allocation</h3>
                    <p className="text-sm text-foreground-muted">Diversification breakdown</p>
                  </div>
                  <motion.button
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                    className="p-2 rounded-lg glass-card hover:bg-glass-strong transition-colors"
                  >
                    <BarChart3 className="w-4 h-4 text-foreground-muted" />
                  </motion.button>
                </div>
                
                <div className="space-y-4">
                  {sectorAllocation.map((sector, index) => (
                    <motion.div
                      key={sector.name}
                      initial={{ opacity: 0, x: 20 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ delay: 0.4 + index * 0.05 }}
                      className="space-y-2"
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <div className={`w-3 h-3 rounded-full bg-current ${sector.color}`} />
                          <span className="text-sm font-medium text-foreground">{sector.name}</span>
                        </div>
                        <div className="text-right">
                          <span className="text-sm font-bold text-foreground">{sector.percentage}%</span>
                          <p className="text-2xs text-foreground-muted">{formatCurrency(sector.value)}</p>
                        </div>
                      </div>
                      <div className="w-full bg-glass rounded-full h-2">
                        <motion.div
                          initial={{ width: 0 }}
                          animate={{ width: `${sector.percentage}%` }}
                          transition={{ delay: 0.5 + index * 0.1, duration: 0.8 }}
                          className={`h-2 rounded-full ${
                            sector.color === 'text-primary' ? 'bg-primary' :
                            sector.color === 'text-success' ? 'bg-success' :
                            sector.color === 'text-warning' ? 'bg-warning' :
                            sector.color === 'text-accent' ? 'bg-accent' :
                            sector.color === 'text-danger' ? 'bg-danger' :
                            'bg-foreground-muted'
                          }`}
                        />
                      </div>
                    </motion.div>
                  ))}
                </div>
                
                <div className="mt-6 pt-4 border-t border-white/10">
                  <div className="grid grid-cols-2 gap-4 text-center">
                    <div>
                      <p className="text-2xs text-foreground-muted mb-1">Diversification</p>
                      <p className="text-lg font-bold text-success">Good</p>
                    </div>
                    <div>
                      <p className="text-2xs text-foreground-muted mb-1">Risk Level</p>
                      <p className="text-lg font-bold text-warning">Medium</p>
                    </div>
                  </div>
                </div>
              </div>
            </motion.section>
          </div>

          {/* Performance Metrics */}
          <motion.section
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.4 }}
          >
            <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
              {[
                { title: 'Total Return', value: '12.5%', change: '+2.1%', icon: TrendingUp, color: 'text-success' },
                { title: 'Sharpe Ratio', value: '1.85', change: '+0.12', icon: Target, color: 'text-primary' },
                { title: 'Max Drawdown', value: '-8.2%', change: '+1.1%', icon: ArrowDownRight, color: 'text-danger' },
                { title: 'Beta', value: '1.12', change: '-0.05', icon: Activity, color: 'text-warning' }
              ].map((metric, index) => {
                const Icon = metric.icon;
                return (
                  <motion.div
                    key={metric.title}
                    initial={{ opacity: 0, scale: 0.9 }}
                    animate={{ opacity: 1, scale: 1 }}
                    transition={{ delay: 0.5 + index * 0.05 }}
                    whileHover={{ scale: 1.05, y: -5 }}
                    className="metric-card"
                  >
                    <div className="flex items-center justify-between mb-4">
                      <Icon className={`w-6 h-6 ${metric.color}`} />
                      <span className="text-2xs text-foreground-muted">{metric.change}</span>
                    </div>
                    <div>
                      <p className="text-2xs text-foreground-muted mb-1">{metric.title}</p>
                      <p className="text-2xl font-bold text-foreground">{metric.value}</p>
                    </div>
                  </motion.div>
                );
              })}
            </div>
          </motion.section>
        </div>
      </main>
    </div>
  );
}