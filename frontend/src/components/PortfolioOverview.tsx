'use client';

import { motion } from 'framer-motion';
import { Portfolio } from '@/types';

interface PortfolioOverviewProps {
  portfolios: Portfolio[];
  riskUpdates: any;
}

export default function PortfolioOverview({ portfolios, riskUpdates }: PortfolioOverviewProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6"
    >
      <h3 className="text-lg font-semibold text-white mb-4">Portfolio Overview</h3>
      <div className="space-y-4">
        {portfolios.map((portfolio) => {
          const totalPnL = portfolio.positions?.reduce((sum, pos) => 
            sum + parseFloat(pos.pnl || '0'), 0
          ) || 0;
          const totalValue = parseFloat(portfolio.total_value);
          const pnlPercent = totalValue > 0 ? (totalPnL / totalValue) * 100 : 0;
          const isPositive = pnlPercent >= 0;

          return (
            <div key={portfolio.id} className="border border-slate-700 rounded-lg p-4">
              <div className="flex items-center justify-between mb-2">
                <h4 className="font-medium text-white">{portfolio.name}</h4>
                <div className="text-right">
                  <div className="text-white font-semibold">
                    ${totalValue.toLocaleString()}
                  </div>
                  <div className={`text-sm ${isPositive ? 'text-green-400' : 'text-red-400'}`}>
                    {isPositive ? '+' : ''}${totalPnL.toLocaleString()} ({pnlPercent.toFixed(2)}%)
                  </div>
                </div>
              </div>
              <p className="text-slate-400 text-sm">{portfolio.description}</p>
              <div className="flex items-center justify-between mt-2 text-sm">
                <span className="text-slate-400">
                  {portfolio.positions?.length || 0} positions
                </span>
                <span className="text-slate-400">
                  {portfolio.currency}
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </motion.div>
  );
}
