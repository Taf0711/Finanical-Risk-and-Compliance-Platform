'use client';

import { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { 
  ChartBarIcon, 
  ExclamationTriangleIcon,
  ArrowTrendingUpIcon,
  ArrowTrendingDownIcon,
  BanknotesIcon,
  ShieldCheckIcon,
  ClockIcon,
  BellIcon
} from '@heroicons/react/24/outline';
import { useRequireAuth } from '@/hooks/useAuth';
import { dashboardApi, WebSocketClient } from '@/lib/api';
import { DashboardData, WebSocketMessage, PriceUpdate, RiskUpdate } from '@/types';
import DashboardLayout from '@/components/DashboardLayout';
import StatCard from '@/components/StatCard';
import RiskMetricsChart from '@/components/RiskMetricsChart';
import PortfolioOverview from '@/components/PortfolioOverview';
import RecentTransactions from '@/components/RecentTransactions';
import ActiveAlerts from '@/components/ActiveAlerts';
import PriceMonitor from '@/components/PriceMonitor';

export default function DashboardPage() {
  const { user, loading } = useRequireAuth();
  const [dashboardData, setDashboardData] = useState<DashboardData | null>(null);
  const [priceUpdates, setPriceUpdates] = useState<PriceUpdate>({});
  const [riskUpdates, setRiskUpdates] = useState<any>({});
  const [wsClient, setWsClient] = useState<WebSocketClient | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    if (!loading && user) {
      loadDashboardData();
      initializeWebSocket();
    }

    return () => {
      if (wsClient) {
        wsClient.disconnect();
      }
    };
  }, [user, loading]);

  const loadDashboardData = async () => {
    try {
      setIsLoading(true);
      const data = await dashboardApi.getDashboardData();
      setDashboardData(data);
    } catch (error) {
      console.error('Failed to load dashboard data:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const initializeWebSocket = () => {
    if (!user) return;

    const client = new WebSocketClient(user.id);
    client.connect(
      (message: WebSocketMessage) => {
        switch (message.type) {
          case 'price_update':
            setPriceUpdates(message.data);
            break;
          case 'risk_update':
            setRiskUpdates((prev: any) => ({
              ...prev,
              [message.data.portfolio_id]: message.data
            }));
            break;
          case 'new_alert':
          case 'aml_alert':
            // Refresh dashboard data to get new alerts
            loadDashboardData();
            break;
          case 'new_transaction':
            // Refresh dashboard data to get new transactions
            loadDashboardData();
            break;
        }
      },
      (error) => {
        console.error('WebSocket error:', error);
      }
    );
    setWsClient(client);
  };

  if (loading || isLoading) {
    return (
      <div className="min-h-screen bg-slate-900 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
          <p className="text-slate-400">Loading dashboard...</p>
        </div>
      </div>
    );
  }

  if (!dashboardData) {
    return (
      <div className="min-h-screen bg-slate-900 flex items-center justify-center">
        <div className="text-center">
          <ExclamationTriangleIcon className="h-12 w-12 text-red-500 mx-auto mb-4" />
          <p className="text-slate-400">Failed to load dashboard data</p>
          <button 
            onClick={loadDashboardData}
            className="mt-4 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  const totalPnLPercent = parseFloat(dashboardData.total_pnl_percent);
  const isPositivePnL = totalPnLPercent >= 0;

  return (
    <DashboardLayout>
      <div className="p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-white">
              Welcome back, {user?.first_name}
            </h1>
            <p className="text-slate-400 mt-1">
              Monitor your portfolio risk and performance in real-time
            </p>
          </div>
          <div className="flex items-center space-x-3">
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              onClick={loadDashboardData}
              className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
            >
              <ArrowTrendingUpIcon className="h-5 w-5 mr-2" />
              Refresh
            </motion.button>
          </div>
        </div>

        {/* Key Metrics */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <StatCard
            title="Total Portfolio Value"
            value={`$${parseFloat(dashboardData.total_value).toLocaleString()}`}
            icon={BanknotesIcon}
            trend={isPositivePnL ? 'up' : 'down'}
            trendValue={`${Math.abs(totalPnLPercent).toFixed(2)}%`}
            color="blue"
          />
          <StatCard
            title="Total P&L"
            value={`$${parseFloat(dashboardData.total_pnl).toLocaleString()}`}
            icon={isPositivePnL ? ArrowTrendingUpIcon : ArrowTrendingDownIcon}
            trend={isPositivePnL ? 'up' : 'down'}
            trendValue={`${totalPnLPercent.toFixed(2)}%`}
            color={isPositivePnL ? 'green' : 'red'}
          />
          <StatCard
            title="Active Portfolios"
            value={dashboardData.portfolios.length.toString()}
            icon={ChartBarIcon}
            color="purple"
          />
          <StatCard
            title="Active Alerts"
            value={dashboardData.active_alerts.length.toString()}
            icon={BellIcon}
            color={dashboardData.active_alerts.length > 0 ? 'red' : 'green'}
          />
        </div>

        {/* Main Content Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Left Column - Portfolio & Risk */}
          <div className="lg:col-span-2 space-y-6">
            {/* Portfolio Overview */}
            <PortfolioOverview 
              portfolios={dashboardData.portfolios}
              riskUpdates={riskUpdates}
            />

            {/* Risk Metrics Chart */}
            <RiskMetricsChart 
              portfolios={dashboardData.portfolios}
              riskUpdates={riskUpdates}
            />

            {/* Recent Transactions */}
            <RecentTransactions 
              transactions={dashboardData.recent_transactions}
            />
          </div>

          {/* Right Column - Alerts & Live Data */}
          <div className="space-y-6">
            {/* Active Alerts */}
            <ActiveAlerts 
              alerts={dashboardData.active_alerts}
            />

            {/* Price Monitor */}
            <PriceMonitor 
              priceUpdates={priceUpdates}
            />

            {/* Risk Status Summary */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6"
            >
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold text-white">Risk Status</h3>
                <ShieldCheckIcon className="h-6 w-6 text-green-400" />
              </div>
              
              <div className="space-y-3">
                {dashboardData.portfolios.map((portfolio) => {
                  const riskData = riskUpdates[portfolio.id];
                  const varStatus = riskData?.var?.status || 'UNKNOWN';
                  const liquidityStatus = riskData?.liquidity?.risk_assessment || 'UNKNOWN';
                  
                  return (
                    <div key={portfolio.id} className="flex items-center justify-between py-2 border-b border-slate-700 last:border-b-0">
                      <div>
                        <p className="text-white font-medium">{portfolio.name}</p>
                        <p className="text-slate-400 text-sm">
                          VaR: {varStatus} | Liquidity: {liquidityStatus}
                        </p>
                      </div>
                      <div className="flex space-x-2">
                        <div className={`w-3 h-3 rounded-full ${
                          varStatus === 'SAFE' ? 'bg-green-400' :
                          varStatus === 'WARNING' ? 'bg-yellow-400' :
                          varStatus === 'CRITICAL' ? 'bg-red-400' : 'bg-gray-400'
                        }`} />
                        <div className={`w-3 h-3 rounded-full ${
                          liquidityStatus === 'LOW_RISK' ? 'bg-green-400' :
                          liquidityStatus === 'MEDIUM_RISK' ? 'bg-yellow-400' :
                          liquidityStatus === 'HIGH_RISK' ? 'bg-red-400' : 'bg-gray-400'
                        }`} />
                      </div>
                    </div>
                  );
                })}
              </div>
            </motion.div>

            {/* System Status */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6"
            >
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold text-white">System Status</h3>
                <ClockIcon className="h-6 w-6 text-blue-400" />
              </div>
              
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-slate-400">WebSocket</span>
                  <div className="flex items-center">
                    <div className={`w-2 h-2 rounded-full mr-2 ${
                      wsClient ? 'bg-green-400' : 'bg-red-400'
                    }`} />
                    <span className="text-sm text-white">
                      {wsClient ? 'Connected' : 'Disconnected'}
                    </span>
                  </div>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-slate-400">Data Updates</span>
                  <span className="text-sm text-green-400">Live</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-slate-400">Last Update</span>
                  <span className="text-sm text-white">
                    {new Date().toLocaleTimeString()}
                  </span>
                </div>
              </div>
            </motion.div>
          </div>
        </div>
      </div>
    </DashboardLayout>
  );
}
