'use client';

import { motion } from 'framer-motion';
import { 
  ChartBarIcon, 
  ArrowTrendingUpIcon,
  ArrowTrendingDownIcon,
  BanknotesIcon,
  ExclamationTriangleIcon,
  ShieldCheckIcon,
  ClockIcon
} from '@heroicons/react/24/outline';
import { Portfolio } from '@/types';
import { useState } from 'react';

interface PortfolioOverviewProps {
  portfolios: Portfolio[];
  riskUpdates: any;
}

export default function PortfolioOverview({ portfolios, riskUpdates }: PortfolioOverviewProps) {
  const [selectedPortfolio, setSelectedPortfolio] = useState<string | null>(null);

  if (!portfolios || portfolios.length === 0) {
    return (
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6"
      >
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-white">Portfolio Overview</h3>
          <ChartBarIcon className="h-6 w-6 text-blue-400" />
        </div>
        <div className="text-center py-8">
          <BanknotesIcon className="h-12 w-12 text-slate-600 mx-auto mb-3" />
          <p className="text-slate-400 mb-2">No portfolios found</p>
          <p className="text-sm text-slate-500">Create your first portfolio to get started</p>
        </div>
      </motion.div>
    );
  }

  // Calculate total portfolio value and performance
  const totalValue = portfolios.reduce((sum, p) => sum + parseFloat(p.total_value), 0);
  const avgPerformance = portfolios.length > 0 ? totalValue / portfolios.length : 0;

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6"
    >
      <div className="flex items-center justify-between mb-6">
        <div>
          <h3 className="text-lg font-semibold text-white">Portfolio Overview</h3>
          <p className="text-sm text-slate-400 mt-1">
            {portfolios.length} portfolio{portfolios.length !== 1 ? 's' : ''} • Total: ${totalValue.toLocaleString()}
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <div className="bg-blue-500/20 p-2 rounded-lg">
            <ChartBarIcon className="h-6 w-6 text-blue-400" />
          </div>
        </div>
      </div>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
        {portfolios.map((portfolio, index) => {
          const totalValue = parseFloat(portfolio.total_value);
          const riskData = riskUpdates[portfolio.id];
          const isSelected = selectedPortfolio === portfolio.id;
          
          // Calculate some mock performance data for demo
          const mockPerformance = {
            dailyChange: ((Math.random() - 0.5) * 0.1), // -5% to +5%
            weeklyChange: ((Math.random() - 0.5) * 0.2), // -10% to +10%
            positions: Math.floor(Math.random() * 15) + 5, // 5-20 positions
          };
          
          const riskLevel = riskData?.var?.status || 
            (totalValue > 600000 ? 'WARNING' : 
             totalValue > 400000 ? 'MEDIUM' : 'SAFE');
          
          return (
            <motion.div
              key={portfolio.id}
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
              onClick={() => setSelectedPortfolio(isSelected ? null : portfolio.id)}
              className={`bg-slate-700/50 border rounded-lg p-4 cursor-pointer transition-all duration-200 ${
                isSelected 
                  ? 'border-blue-500 bg-blue-500/10 shadow-lg shadow-blue-500/20' 
                  : 'border-slate-600 hover:border-slate-500 hover:bg-slate-700/70'
              }`}
            >
              <div className="flex items-start justify-between mb-3">
                <div className="flex-1 min-w-0">
                  <h4 className="font-semibold text-white truncate text-lg">{portfolio.name}</h4>
                  <p className="text-sm text-slate-400 truncate mt-1">{portfolio.description}</p>
                </div>
                <div className="flex items-center space-x-2 ml-3">
                  {riskLevel === 'SAFE' && <ShieldCheckIcon className="h-5 w-5 text-green-400" />}
                  {riskLevel === 'WARNING' && <ExclamationTriangleIcon className="h-5 w-5 text-yellow-400" />}
                  {(riskLevel === 'CRITICAL' || riskLevel === 'HIGH') && <ExclamationTriangleIcon className="h-5 w-5 text-red-400" />}
                </div>
              </div>
              
              <div className="space-y-3">
                {/* Portfolio Value */}
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-2xl font-bold text-white">
                      ${totalValue.toLocaleString()}
                    </p>
                    <div className="flex items-center space-x-2 mt-1">
                      <span className="text-sm text-slate-400">{portfolio.currency}</span>
                      <div className={`flex items-center text-sm ${
                        mockPerformance.dailyChange >= 0 ? 'text-green-400' : 'text-red-400'
                      }`}>
                        {mockPerformance.dailyChange >= 0 ? (
                          <ArrowTrendingUpIcon className="h-4 w-4 mr-1" />
                        ) : (
                          <ArrowTrendingDownIcon className="h-4 w-4 mr-1" />
                        )}
                        {Math.abs(mockPerformance.dailyChange * 100).toFixed(2)}%
                      </div>
                    </div>
                  </div>
                </div>
                
                {/* Risk Indicators */}
                <div className="flex items-center justify-between text-sm">
                  <div className="flex items-center space-x-3">
                    <div className="flex items-center">
                      <div className={`w-2 h-2 rounded-full mr-2 ${
                        riskLevel === 'SAFE' ? 'bg-green-400' :
                        riskLevel === 'WARNING' || riskLevel === 'MEDIUM' ? 'bg-yellow-400' :
                        'bg-red-400'
                      }`} />
                      <span className="text-slate-300 capitalize">{riskLevel.toLowerCase()}</span>
                    </div>
                    <div className="text-slate-400">
                      {mockPerformance.positions} positions
                    </div>
                  </div>
                  <div className="text-slate-400">
                    <ClockIcon className="h-4 w-4 inline mr-1" />
                    {new Date(portfolio.updated_at).toLocaleDateString()}
                  </div>
                </div>
                
                {/* Expanded Details */}
                {isSelected && (
                  <motion.div
                    initial={{ opacity: 0, height: 0 }}
                    animate={{ opacity: 1, height: 'auto' }}
                    exit={{ opacity: 0, height: 0 }}
                    className="border-t border-slate-600 pt-3 mt-3"
                  >
                    <div className="grid grid-cols-2 gap-4 text-sm">
                      <div>
                        <p className="text-slate-400">Weekly Change</p>
                        <p className={`font-medium ${
                          mockPerformance.weeklyChange >= 0 ? 'text-green-400' : 'text-red-400'
                        }`}>
                          {mockPerformance.weeklyChange >= 0 ? '+' : ''}
                          {(mockPerformance.weeklyChange * 100).toFixed(2)}%
                        </p>
                      </div>
                      <div>
                        <p className="text-slate-400">Risk Score</p>
                        <p className="font-medium text-white">
                          {riskData?.riskScore || Math.floor(Math.random() * 100)}
                        </p>
                      </div>
                    </div>
                    
                    {riskData && (
                      <div className="mt-3 p-2 bg-slate-800/50 rounded-lg">
                        <p className="text-xs text-slate-400 mb-1">Risk Metrics</p>
                        <div className="flex items-center justify-between text-xs">
                          {riskData.var && (
                            <span className="text-slate-300">
                              VaR: ${parseFloat(riskData.var.value || 0).toLocaleString()}
                            </span>
                          )}
                          {riskData.liquidity && (
                            <span className="text-slate-300">
                              Liquidity: {riskData.liquidity.risk_assessment || 'N/A'}
                            </span>
                          )}
                        </div>
                      </div>
                    )}
                  </motion.div>
                )}
              </div>
            </motion.div>
          );
        })}
      </div>
      
      {/* Summary Stats */}
      <div className="border-t border-slate-700 pt-4">
        <div className="grid grid-cols-3 gap-4 text-center">
          <div>
            <p className="text-slate-400 text-sm">Total Value</p>
            <p className="text-white font-semibold">${totalValue.toLocaleString()}</p>
          </div>
          <div>
            <p className="text-slate-400 text-sm">Portfolios</p>
            <p className="text-white font-semibold">{portfolios.length}</p>
          </div>
          <div>
            <p className="text-slate-400 text-sm">Avg. Size</p>
            <p className="text-white font-semibold">${avgPerformance.toLocaleString()}</p>
          </div>
        </div>
      </div>
    </motion.div>
  );
}