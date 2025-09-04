'use client';

import { motion } from 'framer-motion';
import { Portfolio } from '@/types';

interface RiskMetricsChartProps {
  portfolios: Portfolio[];
  riskUpdates: any;
}

export default function RiskMetricsChart({ portfolios, riskUpdates }: RiskMetricsChartProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6"
    >
      <h3 className="text-lg font-semibold text-white mb-4">Risk Metrics</h3>
      <div className="space-y-4">
        {portfolios.map((portfolio) => {
          const riskData = riskUpdates[portfolio.id];
          
          return (
            <div key={portfolio.id} className="border border-slate-700 rounded-lg p-4">
              <h4 className="font-medium text-white mb-3">{portfolio.name}</h4>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <div className="text-slate-400 text-sm">Value at Risk (95%)</div>
                  <div className="text-white font-semibold">
                    {riskData?.var?.VaRValue ? 
                      `$${parseFloat(riskData.var.VaRValue).toLocaleString()}` : 
                      'Calculating...'
                    }
                  </div>
                  <div className={`text-xs px-2 py-1 rounded mt-1 inline-block ${
                    riskData?.var?.Status === 'SAFE' ? 'bg-green-500/20 text-green-400' :
                    riskData?.var?.Status === 'WARNING' ? 'bg-yellow-500/20 text-yellow-400' :
                    riskData?.var?.Status === 'CRITICAL' ? 'bg-red-500/20 text-red-400' :
                    'bg-gray-500/20 text-gray-400'
                  }`}>
                    {riskData?.var?.Status || 'Unknown'}
                  </div>
                </div>
                <div>
                  <div className="text-slate-400 text-sm">Liquidity Ratio</div>
                  <div className="text-white font-semibold">
                    {riskData?.liquidity?.LiquidityRatio ? 
                      `${(parseFloat(riskData.liquidity.LiquidityRatio) * 100).toFixed(1)}%` : 
                      'Calculating...'
                    }
                  </div>
                  <div className={`text-xs px-2 py-1 rounded mt-1 inline-block ${
                    riskData?.liquidity?.RiskAssessment === 'LOW_RISK' ? 'bg-green-500/20 text-green-400' :
                    riskData?.liquidity?.RiskAssessment === 'MEDIUM_RISK' ? 'bg-yellow-500/20 text-yellow-400' :
                    riskData?.liquidity?.RiskAssessment === 'HIGH_RISK' ? 'bg-red-500/20 text-red-400' :
                    'bg-gray-500/20 text-gray-400'
                  }`}>
                    {riskData?.liquidity?.RiskAssessment?.replace('_', ' ') || 'Unknown'}
                  </div>
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </motion.div>
  );
}
