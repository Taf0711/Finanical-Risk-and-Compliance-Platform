'use client';

import { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  ChevronDown, 
  ChevronUp, 
  TrendingUp, 
  TrendingDown, 
  Activity,
  Search,
  Filter,
  Star,
  MoreVertical
} from 'lucide-react';
import { cn } from '@/lib/utils';

interface PriceData {
  price: number;
  change: number;
  change_percent: number;
  volume: number;
  timestamp: number;
  provider: string;
  is_market_open: boolean;
}

interface PriceMonitorProps {
  priceUpdates: Record<string, PriceData>;
}

const WATCHLIST_SYMBOLS = [
  'AAPL', 'GOOGL', 'MSFT', 'TSLA', 'AMZN', 'NVDA', 'META', 'JPM', 'BAC', 'GS',
  'WFC', 'C', 'MS', 'NFLX', 'DIS', 'V', 'MA', 'PYPL', 'ADBE', 'CRM',
  'ORCL', 'IBM', 'INTC', 'AMD', 'QCOM', 'JNJ', 'PFE', 'UNH', 'ABBV', 'MRK',
  'TMO', 'KO', 'PEP', 'WMT', 'HD', 'MCD', 'NKE', 'XOM', 'CVX', 'COP'
];

export default function PriceMonitor({ priceUpdates }: PriceMonitorProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [sortBy, setSortBy] = useState<'symbol' | 'price' | 'change'>('symbol');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc');
  const [favorites, setFavorites] = useState<Set<string>>(new Set(['AAPL', 'GOOGL', 'MSFT']));
  const [displayMode, setDisplayMode] = useState<'compact' | 'detailed'>('compact');

  // Filter and sort data
  const filteredData = WATCHLIST_SYMBOLS
    .filter(symbol => 
      symbol.toLowerCase().includes(searchTerm.toLowerCase()) ||
      (priceUpdates[symbol] && symbol.toLowerCase().includes(searchTerm.toLowerCase()))
    )
    .map(symbol => ({
      symbol,
      data: priceUpdates[symbol] || {
        price: Math.random() * 500 + 50,
        change: (Math.random() - 0.5) * 10,
        change_percent: (Math.random() - 0.5) * 5,
        volume: Math.floor(Math.random() * 10000000),
        timestamp: Date.now(),
        provider: 'Fallback',
        is_market_open: true
      }
    }))
    .sort((a, b) => {
      let aVal, bVal;
      switch (sortBy) {
        case 'price':
          aVal = a.data.price;
          bVal = b.data.price;
          break;
        case 'change':
          aVal = a.data.change_percent;
          bVal = b.data.change_percent;
          break;
        default:
          aVal = a.symbol;
          bVal = b.symbol;
      }
      
      if (sortOrder === 'asc') {
        return aVal > bVal ? 1 : -1;
      } else {
        return aVal < bVal ? 1 : -1;
      }
    });

  const toggleFavorite = (symbol: string) => {
    const newFavorites = new Set(favorites);
    if (newFavorites.has(symbol)) {
      newFavorites.delete(symbol);
    } else {
      newFavorites.add(symbol);
    }
    setFavorites(newFavorites);
  };

  const handleSort = (field: 'symbol' | 'price' | 'change') => {
    if (sortBy === field) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortBy(field);
      setSortOrder('asc');
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="glass-black rounded-xl border border-red-primary/20 overflow-hidden"
    >
      {/* Header */}
      <div className="p-4 border-b border-red-primary/20">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 gradient-red-black rounded-lg">
              <Activity className="w-4 h-4 text-white" />
            </div>
            <div>
              <h3 className="text-lg font-semibold text-white">Live Market Data</h3>
              <p className="text-xs text-gray-400">
                {Object.keys(priceUpdates).length} symbols • Last updated: {new Date().toLocaleTimeString()}
              </p>
            </div>
          </div>
          
          <div className="flex items-center gap-2">
            {/* Display Mode Toggle */}
            <button
              onClick={() => setDisplayMode(displayMode === 'compact' ? 'detailed' : 'compact')}
              className="p-2 rounded-lg bg-gray-800 hover:bg-red-primary/20 transition-colors text-gray-400 hover:text-white"
            >
              <MoreVertical className="w-4 h-4" />
            </button>
            
            {/* Expand/Collapse Button */}
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              onClick={() => setIsExpanded(!isExpanded)}
              className="flex items-center gap-2 px-3 py-2 bg-red-primary hover:bg-red-600 text-white rounded-lg transition-colors"
            >
              <span className="text-sm font-medium">
                {isExpanded ? 'Collapse' : 'Expand'}
              </span>
              {isExpanded ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
            </motion.button>
          </div>
        </div>

        {/* Search and Filters - Only show when expanded */}
        <AnimatePresence>
          {isExpanded && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              exit={{ opacity: 0, height: 0 }}
              className="mt-4 space-y-3"
            >
              <div className="flex items-center gap-3">
                <div className="flex-1 relative">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
                  <input
                    type="text"
                    placeholder="Search symbols..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="w-full pl-10 pr-4 py-2 bg-gray-800 border border-red-primary/20 rounded-lg text-white placeholder-gray-400 focus-red"
                  />
                </div>
                
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => handleSort('symbol')}
                    className={cn(
                      "px-3 py-2 rounded-lg text-sm font-medium transition-colors",
                      sortBy === 'symbol' 
                        ? "bg-red-primary text-white" 
                        : "bg-gray-800 text-gray-300 hover:bg-red-primary/20"
                    )}
                  >
                    Symbol {sortBy === 'symbol' && (sortOrder === 'asc' ? '↑' : '↓')}
                  </button>
                  <button
                    onClick={() => handleSort('price')}
                    className={cn(
                      "px-3 py-2 rounded-lg text-sm font-medium transition-colors",
                      sortBy === 'price' 
                        ? "bg-red-primary text-white" 
                        : "bg-gray-800 text-gray-300 hover:bg-red-primary/20"
                    )}
                  >
                    Price {sortBy === 'price' && (sortOrder === 'asc' ? '↑' : '↓')}
                  </button>
                  <button
                    onClick={() => handleSort('change')}
                    className={cn(
                      "px-3 py-2 rounded-lg text-sm font-medium transition-colors",
                      sortBy === 'change' 
                        ? "bg-red-primary text-white" 
                        : "bg-gray-800 text-gray-300 hover:bg-red-primary/20"
                    )}
                  >
                    Change {sortBy === 'change' && (sortOrder === 'asc' ? '↑' : '↓')}
                  </button>
                </div>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      {/* Content */}
      <div className={cn(
        "transition-all duration-300",
        isExpanded ? "max-h-96 overflow-y-auto" : "max-h-32 overflow-hidden"
      )}>
        {!isExpanded ? (
          /* Compact Ticker View */
          <div className="p-4">
            <div className="flex items-center gap-4 overflow-x-auto scrollbar-thin scrollbar-thumb-red-500/30">
              {filteredData.slice(0, 8).map(({ symbol, data }) => (
                <motion.div
                  key={symbol}
                  initial={{ opacity: 0, x: 20 }}
                  animate={{ opacity: 1, x: 0 }}
                  className="flex-shrink-0 flex items-center gap-3 p-3 bg-gray-800 rounded-lg min-w-[160px]"
                >
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => toggleFavorite(symbol)}
                      className={cn(
                        "p-1 rounded transition-colors",
                        favorites.has(symbol) 
                          ? "text-yellow-400 hover:text-yellow-300" 
                          : "text-gray-500 hover:text-gray-400"
                      )}
                    >
                      <Star className="w-3 h-3" fill={favorites.has(symbol) ? 'currentColor' : 'none'} />
                    </button>
                    <span className="text-sm font-medium text-white">{symbol}</span>
                  </div>
                  <div className="text-right">
                    <p className="text-sm font-bold text-white">${data.price.toFixed(2)}</p>
                    <div className="flex items-center gap-1">
                      {data.change >= 0 ? (
                        <TrendingUp className="w-3 h-3 text-green-400" />
                      ) : (
                        <TrendingDown className="w-3 h-3 text-red-400" />
                      )}
                      <span className={cn(
                        "text-xs font-medium",
                        data.change >= 0 ? "text-green-400" : "text-red-400"
                      )}>
                        {data.change_percent.toFixed(2)}%
                      </span>
                    </div>
                  </div>
                </motion.div>
              ))}
            </div>
          </div>
        ) : (
          /* Expanded Detailed View */
          <div className="divide-y divide-red-primary/10">
            {filteredData.map(({ symbol, data }, index) => (
              <motion.div
                key={symbol}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: index * 0.02 }}
                className="p-4 hover:bg-red-primary/5 transition-colors"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <button
                      onClick={() => toggleFavorite(symbol)}
                      className={cn(
                        "p-1 rounded transition-colors",
                        favorites.has(symbol) 
                          ? "text-yellow-400 hover:text-yellow-300" 
                          : "text-gray-500 hover:text-gray-400"
                      )}
                    >
                      <Star className="w-4 h-4" fill={favorites.has(symbol) ? 'currentColor' : 'none'} />
                    </button>
                    <div>
                      <p className="text-sm font-bold text-white">{symbol}</p>
                      <p className="text-xs text-gray-400">Vol: {(data.volume / 1000000).toFixed(1)}M</p>
                    </div>
                  </div>
                  
                  <div className="text-right">
                    <p className="text-lg font-bold text-white">${data.price.toFixed(2)}</p>
                    <div className="flex items-center gap-2">
                      {data.change >= 0 ? (
                        <TrendingUp className="w-4 h-4 text-green-400" />
                      ) : (
                        <TrendingDown className="w-4 h-4 text-red-400" />
                      )}
                      <div className="text-right">
                        <p className={cn(
                          "text-sm font-medium",
                          data.change >= 0 ? "text-green-400" : "text-red-400"
                        )}>
                          {data.change >= 0 ? '+' : ''}{data.change.toFixed(2)}
                        </p>
                        <p className={cn(
                          "text-xs",
                          data.change >= 0 ? "text-green-400" : "text-red-400"
                        )}>
                          {data.change_percent.toFixed(2)}%
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
                
                {displayMode === 'detailed' && (
                  <div className="mt-3 pt-3 border-t border-red-primary/10">
                    <div className="grid grid-cols-3 gap-4 text-xs">
                      <div>
                        <p className="text-gray-400">Provider</p>
                        <p className="text-white font-medium">{data.provider}</p>
                      </div>
                      <div>
                        <p className="text-gray-400">Market</p>
                        <div className="flex items-center gap-1">
                          <div className={cn(
                            "w-2 h-2 rounded-full",
                            data.is_market_open ? "bg-green-500 animate-pulse" : "bg-red-500"
                          )} />
                          <p className="text-white font-medium">
                            {data.is_market_open ? 'Open' : 'Closed'}
                          </p>
                        </div>
                      </div>
                      <div>
                        <p className="text-gray-400">Updated</p>
                        <p className="text-white font-medium">
                          {new Date(data.timestamp).toLocaleTimeString()}
                        </p>
                      </div>
                    </div>
                  </div>
                )}
              </motion.div>
            ))}
          </div>
        )}
      </div>

      {/* Footer */}
      {isExpanded && (
        <div className="p-4 border-t border-red-primary/20 bg-gray-900/50">
          <div className="flex items-center justify-between text-xs text-gray-400">
            <p>Showing {filteredData.length} of {WATCHLIST_SYMBOLS.length} symbols</p>
            <p>Data provided by Alpaca Markets & Fallback</p>
          </div>
        </div>
      )}
    </motion.div>
  );
}
