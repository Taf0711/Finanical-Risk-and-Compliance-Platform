'use client';

import { motion } from 'framer-motion';
import { ArrowUpIcon, ArrowDownIcon } from '@heroicons/react/24/outline';
import { PriceUpdate } from '@/types';

interface PriceMonitorProps {
  priceUpdates: PriceUpdate;
}

export default function PriceMonitor({ priceUpdates }: PriceMonitorProps) {
  const symbols = Object.keys(priceUpdates);

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6"
    >
      <h3 className="text-lg font-semibold text-white mb-4">Live Prices</h3>
      <div className="space-y-3 max-h-64 overflow-y-auto">
        {symbols.slice(0, 8).map((symbol, index) => {
          const data = priceUpdates[symbol];
          const isPositive = data.change >= 0;
          
          return (
            <motion.div
              key={symbol}
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: index * 0.05 }}
              className="flex items-center justify-between p-3 border border-slate-700 rounded-lg hover:bg-slate-700/30 transition-colors"
            >
              <div className="flex items-center space-x-3">
                <div className={`p-1 rounded ${isPositive ? 'bg-green-500/20' : 'bg-red-500/20'}`}>
                  {isPositive ? (
                    <ArrowUpIcon className="w-3 h-3 text-green-400" />
                  ) : (
                    <ArrowDownIcon className="w-3 h-3 text-red-400" />
                  )}
                </div>
                <span className="text-white font-medium">{symbol}</span>
              </div>
              <div className="text-right">
                <div className="text-white font-semibold">
                  ${data.price.toFixed(2)}
                </div>
                <div className={`text-sm ${isPositive ? 'text-green-400' : 'text-red-400'}`}>
                  {isPositive ? '+' : ''}{data.change.toFixed(2)}%
                </div>
              </div>
            </motion.div>
          );
        })}
        {symbols.length === 0 && (
          <div className="text-center py-8 text-slate-400">
            No price data available
          </div>
        )}
      </div>
    </motion.div>
  );
}
