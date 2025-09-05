'use client';

import { motion } from 'framer-motion';
import { 
  ArrowUpIcon, 
  ArrowDownIcon, 
  ChartBarIcon,
  SignalIcon,
  ClockIcon
} from '@heroicons/react/24/outline';
import { PriceUpdate } from '@/types';
import { useState, useEffect } from 'react';

interface PriceMonitorProps {
  priceUpdates: PriceUpdate;
}

export default function PriceMonitor({ priceUpdates }: PriceMonitorProps) {
  const [lastUpdate, setLastUpdate] = useState(new Date());
  const [mockData, setMockData] = useState<PriceUpdate>({});

  // Generate mock data if no real data is available (for testing)
  useEffect(() => {
    const generateMockData = () => {
      const symbols = ['AAPL', 'GOOGL', 'MSFT', 'TSLA', 'AMZN', 'NVDA', 'META', 'BTC', 'ETH'];
      const mock: PriceUpdate = {};
      
      symbols.forEach(symbol => {
        const basePrice = symbol === 'BTC' ? 45000 : symbol === 'ETH' ? 3000 : 150;
        const change = (Math.random() - 0.5) * 10; // Random change between -5% and +5%
        
        mock[symbol] = {
          price: basePrice + (basePrice * change / 100),
          change: change,
          volume: Math.floor(Math.random() * 1000000),
          timestamp: Date.now()
        };
      });
      
      return mock;
    };

    // If no real price updates, use mock data
    if (!priceUpdates || Object.keys(priceUpdates).length === 0) {
      setMockData(generateMockData());
      
      // Update mock data every 5 seconds
      const interval = setInterval(() => {
        setMockData(generateMockData());
        setLastUpdate(new Date());
      }, 5000);
      
      return () => clearInterval(interval);
    }
  }, [priceUpdates]);

  // Update timestamp when new price data is received
  useEffect(() => {
    if (priceUpdates && Object.keys(priceUpdates).length > 0) {
      console.log('PriceMonitor received new price updates:', priceUpdates);
      setLastUpdate(new Date());
    }
  }, [priceUpdates]);

  // Use real data if available, otherwise use mock data
  const activeData = (priceUpdates && Object.keys(priceUpdates).length > 0) ? priceUpdates : mockData;
  const symbols = Object.keys(activeData);
  const priceData = activeData;

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6"
    >
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center space-x-3">
          <div className="bg-green-500/20 p-2 rounded-lg">
            <ChartBarIcon className="h-6 w-6 text-green-400" />
          </div>
          <div>
            <h3 className="text-lg font-semibold text-white">Live Market Data</h3>
            <div className="flex items-center space-x-2 text-sm text-slate-400">
              <SignalIcon className="w-4 h-4" />
              <span>Real-time prices</span>
            </div>
          </div>
        </div>
        <div className="flex items-center space-x-2">
          <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
          <span className="text-sm text-green-400 font-medium">Live</span>
        </div>
      </div>
      
      <div className="space-y-3 max-h-80 overflow-y-auto">
        {symbols.length === 0 ? (
          <div className="text-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-green-400 mx-auto mb-4"></div>
            <p className="text-slate-400">Connecting to real-time market data...</p>
          </div>
        ) : (
          symbols.slice(0, 10).map((symbol, index) => {
            const data = priceData[symbol];
            if (!data) return null;
            
            const isPositive = (data.change || 0) >= 0;
            const isCrypto = symbol === 'BTC' || symbol === 'ETH';
          
            return (
              <motion.div
                key={symbol}
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: index * 0.05 }}
                className="flex items-center justify-between p-4 border border-slate-700 rounded-lg hover:bg-slate-700/30 transition-all duration-200 group"
              >
                <div className="flex items-center space-x-4">
                  <div className={`p-2 rounded-lg transition-colors ${
                    isPositive ? 'bg-green-500/20 group-hover:bg-green-500/30' : 'bg-red-500/20 group-hover:bg-red-500/30'
                  }`}>
                    {isPositive ? (
                      <ArrowUpIcon className="w-5 h-5 text-green-400" />
                    ) : (
                      <ArrowDownIcon className="w-5 h-5 text-red-400" />
                    )}
                  </div>
                  <div>
                    <div className="flex items-center space-x-2">
                      <span className="text-white font-bold text-lg">{symbol}</span>
                      {isCrypto && (
                        <span className="px-2 py-1 text-xs bg-orange-500/20 text-orange-400 rounded-full">
                          CRYPTO
                        </span>
                      )}
                    </div>
                    <div className="text-slate-400 text-sm">
                      Vol: {data.volume || 'N/A'}
                    </div>
                  </div>
                </div>
                
                <div className="text-right">
                  <div className="text-white font-bold text-xl">
                    ${data.price.toLocaleString(undefined, { 
                      minimumFractionDigits: isCrypto ? 0 : 2, 
                      maximumFractionDigits: isCrypto ? 0 : 2 
                    })}
                  </div>
                  <div className={`text-sm font-semibold flex items-center justify-end space-x-1 ${
                    isPositive ? 'text-green-400' : 'text-red-400'
                  }`}>
                    {isPositive ? (
                      <ArrowUpIcon className="w-3 h-3" />
                    ) : (
                      <ArrowDownIcon className="w-3 h-3" />
                    )}
                    <span>
                      {isPositive ? '+' : ''}{(data.change || 0).toFixed(2)}%
                    </span>
                  </div>
                </div>
              </motion.div>
            );
          })
        )}
      </div>
      
      {/* Footer with last update time */}
      {symbols.length > 0 && (
        <div className="border-t border-slate-700 pt-4 mt-4">
          <div className="flex items-center justify-between text-sm">
            <div className="flex items-center space-x-2 text-slate-400">
              <ClockIcon className="w-4 h-4" />
              <span>Last updated: {lastUpdate.toLocaleTimeString()}</span>
            </div>
            <div className="text-slate-400">
              {symbols.length} symbols
            </div>
          </div>
        </div>
      )}
    </motion.div>
  );
}
