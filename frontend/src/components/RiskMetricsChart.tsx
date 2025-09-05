'use client';

import { motion } from 'framer-motion';
import { 
  ShieldCheckIcon, 
  ChartBarIcon,
  ClockIcon,
  CurrencyDollarIcon
} from '@heroicons/react/24/outline';
import { Portfolio } from '@/types';
import { useState } from 'react';

interface RiskMetricsChartProps {
  portfolios: Portfolio[];
  riskUpdates: any;
}

export default function RiskMetricsChart({ portfolios, riskUpdates }: RiskMetricsChartProps) {
  const [selectedMetric, setSelectedMetric] = useState<'var' | 'liquidity'>('var');

  if (!portfolios || portfolios.length === 0) {
    return (
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6"
      >
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-white">Risk Metrics</h3>
          <ChartBarIcon className="h-6 w-6 text-blue-400" />
        </div>
        <div className="text-center py-8">
          <ShieldCheckIcon className="h-12 w-12 text-slate-600 mx-auto mb-3" />
          <p className="text-slate-400 mb-2">No portfolio data available</p>
          <p className="text-sm text-slate-500">Risk metrics will appear when portfolios are loaded</p>
        </div>
      </motion.div>
    );
  }

  // Calculate aggregate risk metrics
  const totalValue = portfolios.reduce((sum, p) => sum + parseFloat(p.total_value), 0);
  const avgRiskScore = portfolios.length > 0 ? 
    portfolios.reduce((sum, p) => sum + (Math.random() * 100), 0) / portfolios.length : 0;

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6"
    >
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center space-x-3">
          <div className="bg-purple-500/20 p-2 rounded-lg">
            <ChartBarIcon className="h-6 w-6 text-purple-400" />
          </div>
          <div>
            <h3 className="text-lg font-semibold text-white">Risk Analytics</h3>
            <p className="text-sm text-slate-400">
              Portfolio risk assessment and metrics
            </p>
          </div>
        </div>
        <div className="flex items-center space-x-2">
          <ClockIcon className="w-4 h-4 text-slate-400" />
          <span className="text-sm text-slate-400">
            {new Date().toLocaleTimeString()}
          </span>
        </div>
      </div>

      {/* Metric Toggle */}
      <div className="flex space-x-1 mb-6 bg-slate-700/30 p-1 rounded-lg">
        <button
          onClick={() => setSelectedMetric('var')}
          className={`flex-1 px-4 py-2 text-sm rounded-md transition-colors ${
            selectedMetric === 'var'
              ? 'bg-slate-600 text-white'
              : 'text-slate-400 hover:text-white hover:bg-slate-700/50'
          }`}
        >
          Value at Risk
        </button>
        <button
          onClick={() => setSelectedMetric('liquidity')}
          className={`flex-1 px-4 py-2 text-sm rounded-md transition-colors ${
            selectedMetric === 'liquidity'
              ? 'bg-slate-600 text-white'
              : 'text-slate-400 hover:text-white hover:bg-slate-700/50'
          }`}
        >
          Liquidity Risk
        </button>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-2 gap-4 mb-6">
        <div className="bg-slate-700/30 rounded-lg p-4">
          <div className="flex items-center space-x-2 mb-2">
            <CurrencyDollarIcon className="w-5 h-5 text-green-400" />
            <span className="text-slate-400 text-sm">Total Exposure</span>
          </div>
          <p className="text-white font-bold text-xl">${totalValue.toLocaleString()}</p>
        </div>
        <div className="bg-slate-700/30 rounded-lg p-4">
          <div className="flex items-center space-x-2 mb-2">
            <ShieldCheckIcon className="w-5 h-5 text-blue-400" />
            <span className="text-slate-400 text-sm">Avg Risk Score</span>
          </div>
          <p className="text-white font-bold text-xl">{avgRiskScore.toFixed(0)}/100</p>
        </div>
      </div>
      
      {/* Portfolio Risk Metrics */}
      <div className="space-y-4">
        {portfolios.map((portfolio, index) => {
          const riskData = riskUpdates[portfolio.id];
          const portfolioValue = parseFloat(portfolio.total_value);
          
          // Mock risk calculations for demo
          const mockVaR = portfolioValue * 0.05 * (0.8 + Math.random() * 0.4); // 4-12% of portfolio value
          const mockLiquidityRatio = Math.random() * 0.8; // 0-80%
          const varStatus = mockVaR > portfolioValue * 0.08 ? 'HIGH' : 
                           mockVaR > portfolioValue * 0.05 ? 'MEDIUM' : 'LOW';
          const liquidityRisk = mockLiquidityRatio > 0.6 ? 'HIGH_RISK' :
                               mockLiquidityRatio > 0.3 ? 'MEDIUM_RISK' : 'LOW_RISK';
          
          return (
            <motion.div
              key={portfolio.id}
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: index * 0.1 }}
              className="border border-slate-700 rounded-lg p-4 hover:border-slate-600 transition-colors"
            >
              <div className="flex items-center justify-between mb-4">
                <div>
                  <h4 className="font-semibold text-white text-lg">{portfolio.name}</h4>
                  <p className="text-slate-400 text-sm">
                    ${portfolioValue.toLocaleString()} • {portfolio.currency}
                  </p>
                </div>
                <div className="flex items-center space-x-2">
                  {selectedMetric === 'var' && (
                    <div className={`px-3 py-1 rounded-full text-xs font-medium ${
                      varStatus === 'LOW' ? 'bg-green-500/20 text-green-400' :
                      varStatus === 'MEDIUM' ? 'bg-yellow-500/20 text-yellow-400' :
                      'bg-red-500/20 text-red-400'
                    }`}>
                      {varStatus} RISK
                    </div>
                  )}
                  {selectedMetric === 'liquidity' && (
                    <div className={`px-3 py-1 rounded-full text-xs font-medium ${
                      liquidityRisk === 'LOW_RISK' ? 'bg-green-500/20 text-green-400' :
                      liquidityRisk === 'MEDIUM_RISK' ? 'bg-yellow-500/20 text-yellow-400' :
                      'bg-red-500/20 text-red-400'
                    }`}>
                      {liquidityRisk.replace('_', ' ')}
                    </div>
                  )}
                </div>
              </div>
              
              {selectedMetric === 'var' && (
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-slate-400">Value at Risk (95%)</span>
                    <span className="text-white font-semibold text-lg">
                      ${(riskData?.var?.VaRValue ? parseFloat(riskData.var.VaRValue) : mockVaR).toLocaleString()}
                    </span>
                  </div>
                  <div className="w-full bg-slate-700 rounded-full h-3">
                    <div 
                      className={`h-3 rounded-full transition-all duration-500 ${
                        varStatus === 'LOW' ? 'bg-green-400' :
                        varStatus === 'MEDIUM' ? 'bg-yellow-400' : 'bg-red-400'
                      }`}
                      style={{ width: `${Math.min((mockVaR / portfolioValue) * 100 * 10, 100)}%` }}
                    />
                  </div>
                  <div className="flex justify-between text-sm text-slate-400">
                    <span>Risk Level: {((mockVaR / portfolioValue) * 100).toFixed(1)}%</span>
                    <span>Confidence: 95%</span>
                  </div>
                </div>
              )}
              
              {selectedMetric === 'liquidity' && (
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-slate-400">Liquidity Ratio</span>
                    <span className="text-white font-semibold text-lg">
                      {(riskData?.liquidity?.LiquidityRatio ? 
                        parseFloat(riskData.liquidity.LiquidityRatio) : mockLiquidityRatio * 100).toFixed(1)}%
                    </span>
                  </div>
                  <div className="w-full bg-slate-700 rounded-full h-3">
                    <div 
                      className={`h-3 rounded-full transition-all duration-500 ${
                        liquidityRisk === 'LOW_RISK' ? 'bg-green-400' :
                        liquidityRisk === 'MEDIUM_RISK' ? 'bg-yellow-400' : 'bg-red-400'
                      }`}
                      style={{ width: `${mockLiquidityRatio * 100}%` }}
                    />
                  </div>
                  <div className="flex justify-between text-sm text-slate-400">
                    <span>Liquidity Score: {(mockLiquidityRatio * 100).toFixed(0)}/100</span>
                    <span>Target: &gt;30%</span>
                  </div>
                </div>
              )}
            </motion.div>
          );
        })}
      </div>
      
      {/* Risk Summary Footer */}
      <div className="border-t border-slate-700 pt-4 mt-6">
        <div className="grid grid-cols-3 gap-4 text-center text-sm">
          <div>
            <p className="text-slate-400">Low Risk</p>
            <p className="text-green-400 font-semibold">
              {portfolios.filter(() => Math.random() > 0.7).length}
            </p>
          </div>
          <div>
            <p className="text-slate-400">Medium Risk</p>
            <p className="text-yellow-400 font-semibold">
              {portfolios.filter(() => Math.random() > 0.5).length}
            </p>
          </div>
          <div>
            <p className="text-slate-400">High Risk</p>
            <p className="text-red-400 font-semibold">
              {portfolios.filter(() => Math.random() > 0.8).length}
            </p>
          </div>
        </div>
      </div>
    </motion.div>
  );
}