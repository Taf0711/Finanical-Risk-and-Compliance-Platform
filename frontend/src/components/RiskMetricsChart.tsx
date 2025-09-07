'use client';

import { motion } from 'framer-motion';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  AreaChart,
  Area,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  RadialBarChart,
  RadialBar,
  Legend
} from 'recharts';
import {
  TrendingUp,
  TrendingDown,
  AlertTriangle,
  Shield,
  Target,
  Activity,
  PieChart as PieChartIcon,
  BarChart3,
  TrendingUp as LineChartIcon
} from 'lucide-react';
import { useState } from 'react';

interface RiskMetricsChartProps {
  className?: string;
}

// Mock data for different chart types
const riskTimeSeriesData = [
  { date: '2024-01', var: 0.15, liquidity: 0.85, volatility: 0.25, stress: 0.12 },
  { date: '2024-02', var: 0.18, liquidity: 0.82, volatility: 0.28, stress: 0.15 },
  { date: '2024-03', var: 0.16, liquidity: 0.88, volatility: 0.22, stress: 0.11 },
  { date: '2024-04', var: 0.20, liquidity: 0.78, volatility: 0.32, stress: 0.18 },
  { date: '2024-05', var: 0.14, liquidity: 0.90, volatility: 0.20, stress: 0.10 },
  { date: '2024-06', var: 0.17, liquidity: 0.85, volatility: 0.26, stress: 0.14 },
];

const portfolioComposition = [
  { name: 'Equities', value: 45, color: '#3B82F6' },
  { name: 'Bonds', value: 25, color: '#10B981' },
  { name: 'Commodities', value: 15, color: '#F59E0B' },
  { name: 'Cash', value: 10, color: '#6B7280' },
  { name: 'Alternatives', value: 5, color: '#8B5CF6' },
];

const riskMetrics = [
  { metric: 'VaR (95%)', value: 85, max: 100, color: '#EF4444' },
  { metric: 'Liquidity', value: 72, max: 100, color: '#3B82F6' },
  { metric: 'Concentration', value: 45, max: 100, color: '#F59E0B' },
  { metric: 'Volatility', value: 68, max: 100, color: '#8B5CF6' },
];

const stressTestResults = [
  { scenario: 'Market Crash', impact: -15.2, probability: 5 },
  { scenario: 'Interest Rate Shock', impact: -8.7, probability: 15 },
  { scenario: 'Inflation Spike', impact: -12.3, probability: 25 },
  { scenario: 'Currency Crisis', impact: -6.8, probability: 10 },
  { scenario: 'Sector Rotation', impact: -4.5, probability: 35 },
];

type ChartType = 'timeseries' | 'composition' | 'metrics' | 'stress';

export default function RiskMetricsChart({ className }: RiskMetricsChartProps) {
  const [activeChart, setActiveChart] = useState<ChartType>('timeseries');
  const [selectedMetric, setSelectedMetric] = useState<string>('var');

  const chartOptions = [
    { key: 'timeseries' as ChartType, label: 'Time Series', icon: LineChartIcon, color: 'blue' },
    { key: 'composition' as ChartType, label: 'Portfolio Mix', icon: PieChartIcon, color: 'green' },
    { key: 'metrics' as ChartType, label: 'Risk Metrics', icon: BarChart3, color: 'purple' },
    { key: 'stress' as ChartType, label: 'Stress Tests', icon: AlertTriangle, color: 'red' },
  ];

  const CustomTooltip = ({ active, payload, label }: any) => {
    if (active && payload && payload.length) {
      return (
        <div className="bg-white/95 dark:bg-gray-800/95 backdrop-blur-sm border border-gray-200 dark:border-gray-700 rounded-lg p-3 shadow-lg">
          <p className="text-sm font-medium text-gray-900 dark:text-white">{label}</p>
          {payload.map((entry: any, index: number) => (
            <p key={index} className="text-sm" style={{ color: entry.color }}>
              {`${entry.dataKey}: ${(entry.value * 100).toFixed(1)}%`}
            </p>
          ))}
        </div>
      );
    }
    return null;
  };

  const renderTimeSeriesChart = () => (
    <div className="space-y-4">
      {/* Metric Selector */}
      <div className="flex flex-wrap gap-2">
        {[
          { key: 'var', label: 'VaR', color: 'red' },
          { key: 'liquidity', label: 'Liquidity', color: 'blue' },
          { key: 'volatility', label: 'Volatility', color: 'purple' },
          { key: 'stress', label: 'Stress', color: 'orange' }
        ].map(({ key, label, color }) => (
          <button
            key={key}
            onClick={() => setSelectedMetric(key)}
            className={cn(
              "px-3 py-1 rounded-full text-xs font-medium transition-all",
              selectedMetric === key
                ? "bg-blue-500 text-white shadow-lg"
                : "bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600"
            )}
          >
            {label}
          </button>
        ))}
      </div>

      {/* Chart */}
      <ResponsiveContainer width="100%" height={300}>
        <AreaChart data={riskTimeSeriesData}>
          <defs>
            <linearGradient id="colorVar" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#EF4444" stopOpacity={0.3}/>
              <stop offset="95%" stopColor="#EF4444" stopOpacity={0}/>
            </linearGradient>
            <linearGradient id="colorLiquidity" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#3B82F6" stopOpacity={0.3}/>
              <stop offset="95%" stopColor="#3B82F6" stopOpacity={0}/>
            </linearGradient>
            <linearGradient id="colorVolatility" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#8B5CF6" stopOpacity={0.3}/>
              <stop offset="95%" stopColor="#8B5CF6" stopOpacity={0}/>
            </linearGradient>
            <linearGradient id="colorStress" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#F59E0B" stopOpacity={0.3}/>
              <stop offset="95%" stopColor="#F59E0B" stopOpacity={0}/>
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" stroke="#374151" opacity={0.3} />
          <XAxis 
            dataKey="date" 
            stroke="#6B7280" 
            fontSize={12}
            tickLine={false}
          />
          <YAxis 
            stroke="#6B7280" 
            fontSize={12}
            tickLine={false}
            tickFormatter={(value) => `${(value * 100).toFixed(0)}%`}
          />
          <Tooltip content={<CustomTooltip />} />
          <Area
            type="monotone"
            dataKey={selectedMetric}
            stroke={selectedMetric === 'var' ? '#EF4444' : 
                   selectedMetric === 'liquidity' ? '#3B82F6' :
                   selectedMetric === 'volatility' ? '#8B5CF6' : '#F59E0B'}
            fillOpacity={1}
            fill={selectedMetric === 'var' ? 'url(#colorVar)' : 
                  selectedMetric === 'liquidity' ? 'url(#colorLiquidity)' :
                  selectedMetric === 'volatility' ? 'url(#colorVolatility)' : 'url(#colorStress)'}
            strokeWidth={2}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );

  const renderCompositionChart = () => (
    <ResponsiveContainer width="100%" height={300}>
      <PieChart>
        <Pie
          data={portfolioComposition}
          cx="50%"
          cy="50%"
          outerRadius={100}
          innerRadius={40}
          paddingAngle={2}
          dataKey="value"
        >
          {portfolioComposition.map((entry, index) => (
            <Cell key={`cell-${index}`} fill={entry.color} />
          ))}
        </Pie>
        <Tooltip
          formatter={(value: number) => [`${value}%`, 'Allocation']}
          contentStyle={{
            backgroundColor: 'rgba(255, 255, 255, 0.95)',
            border: '1px solid #E5E7EB',
            borderRadius: '8px',
            backdropFilter: 'blur(4px)'
          }}
        />
        <Legend 
          wrapperStyle={{ paddingTop: '20px' }}
          formatter={(value) => <span className="text-sm">{value}</span>}
        />
      </PieChart>
    </ResponsiveContainer>
  );

  const renderMetricsChart = () => (
    <ResponsiveContainer width="100%" height={300}>
      <RadialBarChart cx="50%" cy="50%" innerRadius="20%" outerRadius="90%" data={riskMetrics}>
        <RadialBar
          label={{ position: 'insideStart', fill: '#fff', fontSize: 12 }}
          background
          clockWise={true}
          dataKey="value"
          cornerRadius={4}
          fill="#8884d8"
        />
        <Legend 
          iconSize={10}
          layout="vertical"
          verticalAlign="middle"
          align="right"
          wrapperStyle={{ paddingLeft: '20px' }}
        />
        <Tooltip
          formatter={(value: number) => [`${value}%`, 'Risk Level']}
          contentStyle={{
            backgroundColor: 'rgba(255, 255, 255, 0.95)',
            border: '1px solid #E5E7EB',
            borderRadius: '8px',
            backdropFilter: 'blur(4px)'
          }}
        />
      </RadialBarChart>
    </ResponsiveContainer>
  );

  const renderStressTestChart = () => (
    <ResponsiveContainer width="100%" height={300}>
      <BarChart data={stressTestResults} layout="horizontal">
        <CartesianGrid strokeDasharray="3 3" stroke="#374151" opacity={0.3} />
        <XAxis 
          type="number" 
          stroke="#6B7280" 
          fontSize={12}
          tickLine={false}
          tickFormatter={(value) => `${value}%`}
        />
        <YAxis 
          type="category" 
          dataKey="scenario" 
          stroke="#6B7280" 
          fontSize={12}
          tickLine={false}
          width={120}
        />
        <Tooltip
          formatter={(value: number) => [`${value}%`, 'Impact']}
          contentStyle={{
            backgroundColor: 'rgba(255, 255, 255, 0.95)',
            border: '1px solid #E5E7EB',
            borderRadius: '8px',
            backdropFilter: 'blur(4px)'
          }}
        />
        <Bar 
          dataKey="impact" 
          fill="#EF4444"
          radius={[0, 4, 4, 0]}
        />
      </BarChart>
    </ResponsiveContainer>
  );

  const renderChart = () => {
    switch (activeChart) {
      case 'timeseries':
        return renderTimeSeriesChart();
      case 'composition':
        return renderCompositionChart();
      case 'metrics':
        return renderMetricsChart();
      case 'stress':
        return renderStressTestChart();
      default:
        return renderTimeSeriesChart();
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className={cn("space-y-6", className)}
    >
      <Card className="bg-gradient-to-br from-white/90 to-white/60 dark:from-gray-900/90 dark:to-gray-900/60 backdrop-blur-md border-white/20 dark:border-gray-700/50">
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-3">
              <div className="p-2 bg-gradient-to-r from-purple-500 to-pink-600 rounded-lg">
                <Activity className="w-5 h-5 text-white" />
              </div>
              Risk Analytics Dashboard
            </CardTitle>
            
            {/* Chart Type Selector */}
            <div className="flex gap-2">
              {chartOptions.map(({ key, label, icon: Icon, color }) => (
                <motion.button
                  key={key}
                  onClick={() => setActiveChart(key)}
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  className={cn(
                    "flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-all",
                    activeChart === key
                      ? "bg-blue-500 text-white shadow-lg"
                      : "bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600"
                  )}
                >
                  <Icon className="w-4 h-4" />
                  <span className="hidden md:inline">{label}</span>
                </motion.button>
              ))}
            </div>
          </div>
        </CardHeader>

        <CardContent>
          <motion.div
            key={activeChart}
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.3 }}
          >
            {renderChart()}
          </motion.div>
        </CardContent>
      </Card>

      {/* Risk Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {[
          { 
            title: 'Portfolio VaR', 
            value: '2.4%', 
            change: -0.3, 
            icon: Shield,
            description: '95% confidence, 1-day horizon'
          },
          { 
            title: 'Liquidity Risk', 
            value: 'Medium', 
            change: 0.1, 
            icon: Activity,
            description: '72% liquid assets'
          },
          { 
            title: 'Concentration Risk', 
            value: 'Low', 
            change: -0.2, 
            icon: Target,
            description: 'Well diversified'
          },
          { 
            title: 'Stress Test', 
            value: '-8.7%', 
            change: 0.5, 
            icon: AlertTriangle,
            description: 'Worst case scenario'
          }
        ].map((metric, index) => (
          <motion.div
            key={metric.title}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.1 }}
            whileHover={{ scale: 1.02 }}
          >
            <Card className="bg-gradient-to-br from-white/90 to-white/60 dark:from-gray-900/90 dark:to-gray-900/60 backdrop-blur-md border-white/20 dark:border-gray-700/50">
              <CardContent className="p-4">
                <div className="flex items-center justify-between mb-2">
                  <metric.icon className="w-5 h-5 text-gray-600 dark:text-gray-400" />
                  <Badge 
                    variant={metric.change > 0 ? 'success' : metric.change < 0 ? 'danger' : 'secondary'}
                    className="text-xs"
                  >
                    {metric.change > 0 ? <TrendingUp className="w-3 h-3" /> : <TrendingDown className="w-3 h-3" />}
                    {Math.abs(metric.change)}%
                  </Badge>
                </div>
                <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-1">
                  {metric.title}
                </h3>
                <p className="text-xl font-bold text-gray-900 dark:text-white mb-1">
                  {metric.value}
                </p>
                <p className="text-xs text-gray-500 dark:text-gray-400">
                  {metric.description}
                </p>
              </CardContent>
            </Card>
          </motion.div>
        ))}
      </div>
    </motion.div>
  );
}