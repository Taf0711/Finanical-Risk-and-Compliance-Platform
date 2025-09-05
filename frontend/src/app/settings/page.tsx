'use client';

import { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import {
  CogIcon,
  BellIcon,
  ShieldCheckIcon,
  ChartBarIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
  CheckCircleIcon,
  PlusIcon,
  TrashIcon,
} from '@heroicons/react/24/outline';
import { useRequireAuth } from '@/hooks/useAuth';
import DashboardLayout from '@/components/DashboardLayout';

const alertRuleSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  type: z.enum(['risk_breach', 'price_movement', 'volume_spike', 'position_limit']),
  condition_field: z.string().min(1, 'Field is required'),
  condition_operator: z.enum(['>', '>=', '<', '<=', '==', '!=']),
  condition_value: z.number().min(0, 'Value must be positive'),
  severity: z.enum(['low', 'medium', 'high', 'critical']),
  enabled: z.boolean(),
  cooldown_minutes: z.number().min(1, 'Cooldown must be at least 1 minute'),
});

const riskLimitsSchema = z.object({
  max_position_size: z.number().min(0.01).max(1, 'Must be between 1% and 100%'),
  max_sector_concentration: z.number().min(0.01).max(1, 'Must be between 1% and 100%'),
  var_limit: z.number().min(1000, 'Minimum VaR limit is $1,000'),
  max_drawdown: z.number().min(0.01).max(1, 'Must be between 1% and 100%'),
  liquidity_threshold: z.number().min(1).max(30, 'Must be between 1 and 30 days'),
  volatility_limit: z.number().min(0.05).max(2, 'Must be between 5% and 200%'),
});

const tradingSettingsSchema = z.object({
  auto_rebalancing: z.boolean(),
  rebalancing_threshold: z.number().min(0.01).max(0.5, 'Must be between 1% and 50%'),
  stop_loss_enabled: z.boolean(),
  default_stop_loss: z.number().min(0.01).max(0.5, 'Must be between 1% and 50%'),
  take_profit_enabled: z.boolean(),
  default_take_profit: z.number().min(0.01).max(2, 'Must be between 1% and 200%'),
  paper_trading_mode: z.boolean(),
  max_daily_trades: z.number().min(1).max(100, 'Must be between 1 and 100'),
});

type AlertRuleFormData = z.infer<typeof alertRuleSchema>;
type RiskLimitsFormData = z.infer<typeof riskLimitsSchema>;
type TradingSettingsFormData = z.infer<typeof tradingSettingsSchema>;

interface AlertRule {
  id: string;
  name: string;
  type: string;
  condition_field: string;
  condition_operator: string;
  condition_value: number;
  severity: string;
  enabled: boolean;
  cooldown_minutes: number;
  created_at: string;
}

export default function SettingsPage() {
  const { user, loading } = useRequireAuth();
  const [activeTab, setActiveTab] = useState('alerts');
  const [alertRules, setAlertRules] = useState<AlertRule[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [updateSuccess, setUpdateSuccess] = useState('');
  const [updateError, setUpdateError] = useState('');
  const [showAddRule, setShowAddRule] = useState(false);

  const alertRuleForm = useForm<AlertRuleFormData>({
    resolver: zodResolver(alertRuleSchema),
    defaultValues: {
      enabled: true,
      cooldown_minutes: 60,
      severity: 'medium',
    },
  });

  const riskLimitsForm = useForm<RiskLimitsFormData>({
    resolver: zodResolver(riskLimitsSchema),
    defaultValues: {
      max_position_size: 0.2,
      max_sector_concentration: 0.3,
      var_limit: 50000,
      max_drawdown: 0.15,
      liquidity_threshold: 5,
      volatility_limit: 0.25,
    },
  });

  const tradingSettingsForm = useForm<TradingSettingsFormData>({
    resolver: zodResolver(tradingSettingsSchema),
    defaultValues: {
      auto_rebalancing: false,
      rebalancing_threshold: 0.05,
      stop_loss_enabled: true,
      default_stop_loss: 0.05,
      take_profit_enabled: false,
      default_take_profit: 0.15,
      paper_trading_mode: true,
      max_daily_trades: 10,
    },
  });

  useEffect(() => {
    if (!loading && user) {
      loadSettings();
    }
  }, [user, loading]);

  const loadSettings = async () => {
    try {
      setIsLoading(true);
      // TODO: Load actual settings from API
      // For now, use mock data
      setAlertRules([
        {
          id: '1',
          name: 'High VaR Alert',
          type: 'risk_breach',
          condition_field: 'var_1day',
          condition_operator: '>',
          condition_value: 50000,
          severity: 'high',
          enabled: true,
          cooldown_minutes: 60,
          created_at: '2024-01-01T00:00:00Z',
        },
        {
          id: '2',
          name: 'Large Position Alert',
          type: 'position_limit',
          condition_field: 'position_size',
          condition_operator: '>',
          condition_value: 0.25,
          severity: 'medium',
          enabled: true,
          cooldown_minutes: 30,
          created_at: '2024-01-01T00:00:00Z',
        },
      ]);
    } catch (error) {
      console.error('Failed to load settings:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const onAlertRuleSubmit = async (data: AlertRuleFormData) => {
    try {
      setUpdateError('');
      setUpdateSuccess('');
      
      // TODO: Submit to API
      const newRule: AlertRule = {
        id: Date.now().toString(),
        ...data,
        created_at: new Date().toISOString(),
      };
      
      setAlertRules(prev => [...prev, newRule]);
      setUpdateSuccess('Alert rule created successfully!');
      alertRuleForm.reset();
      setShowAddRule(false);
    } catch (error: any) {
      setUpdateError(error.message || 'Failed to create alert rule');
    }
  };

  const onRiskLimitsSubmit = async (data: RiskLimitsFormData) => {
    try {
      setUpdateError('');
      setUpdateSuccess('');
      
      // TODO: Submit to API
      await new Promise(resolve => setTimeout(resolve, 1000));
      setUpdateSuccess('Risk limits updated successfully!');
    } catch (error: any) {
      setUpdateError(error.message || 'Failed to update risk limits');
    }
  };

  const onTradingSettingsSubmit = async (data: TradingSettingsFormData) => {
    try {
      setUpdateError('');
      setUpdateSuccess('');
      
      // TODO: Submit to API
      await new Promise(resolve => setTimeout(resolve, 1000));
      setUpdateSuccess('Trading settings updated successfully!');
    } catch (error: any) {
      setUpdateError(error.message || 'Failed to update trading settings');
    }
  };

  const deleteAlertRule = async (ruleId: string) => {
    try {
      // TODO: Delete from API
      setAlertRules(prev => prev.filter(rule => rule.id !== ruleId));
      setUpdateSuccess('Alert rule deleted successfully!');
    } catch (error: any) {
      setUpdateError(error.message || 'Failed to delete alert rule');
    }
  };

  const toggleAlertRule = async (ruleId: string, enabled: boolean) => {
    try {
      // TODO: Update API
      setAlertRules(prev => 
        prev.map(rule => 
          rule.id === ruleId ? { ...rule, enabled } : rule
        )
      );
      setUpdateSuccess(`Alert rule ${enabled ? 'enabled' : 'disabled'} successfully!`);
    } catch (error: any) {
      setUpdateError(error.message || 'Failed to update alert rule');
    }
  };

  if (loading || isLoading) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center min-h-96">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
        </div>
      </DashboardLayout>
    );
  }

  const tabs = [
    { id: 'alerts', label: 'Alert Rules', icon: BellIcon },
    { id: 'risk', label: 'Risk Limits', icon: ShieldCheckIcon },
    { id: 'trading', label: 'Trading Settings', icon: ChartBarIcon },
    { id: 'system', label: 'System Settings', icon: CogIcon },
  ];

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'bg-red-500/20 text-red-400 border-red-500/30';
      case 'high':
        return 'bg-orange-500/20 text-orange-400 border-orange-500/30';
      case 'medium':
        return 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30';
      case 'low':
        return 'bg-blue-500/20 text-blue-400 border-blue-500/30';
      default:
        return 'bg-gray-500/20 text-gray-400 border-gray-500/30';
    }
  };

  return (
    <DashboardLayout>
      <div className="p-6">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white">Application Settings</h1>
          <p className="text-slate-400 mt-2">Configure alerts, risk limits, and trading parameters</p>
        </div>

        {/* Success/Error Messages */}
        {updateSuccess && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className="mb-6 p-4 bg-green-500/20 border border-green-500/30 rounded-lg text-green-200 flex items-center"
          >
            <CheckCircleIcon className="h-5 w-5 mr-2" />
            {updateSuccess}
          </motion.div>
        )}

        {updateError && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className="mb-6 p-4 bg-red-500/20 border border-red-500/30 rounded-lg text-red-200 flex items-center"
          >
            <ExclamationTriangleIcon className="h-5 w-5 mr-2" />
            {updateError}
          </motion.div>
        )}

        {/* Tabs */}
        <div className="mb-6">
          <div className="flex space-x-1 bg-slate-800/30 p-1 rounded-lg">
            {tabs.map((tab) => {
              const Icon = tab.icon;
              return (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`flex-1 flex items-center justify-center px-4 py-2 text-sm rounded-md transition-colors ${
                    activeTab === tab.id
                      ? 'bg-slate-700 text-white'
                      : 'text-slate-400 hover:text-white hover:bg-slate-700/50'
                  }`}
                >
                  <Icon className="w-4 h-4 mr-2" />
                  {tab.label}
                </button>
              );
            })}
          </div>
        </div>

        {/* Tab Content */}
        <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6">
          {activeTab === 'alerts' && (
            <div>
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-xl font-semibold text-white">Alert Rules</h2>
                <button
                  onClick={() => setShowAddRule(true)}
                  className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                >
                  <PlusIcon className="w-4 h-4 mr-2" />
                  Add Rule
                </button>
              </div>

              {/* Add Rule Form */}
              {showAddRule && (
                <motion.div
                  initial={{ opacity: 0, height: 0 }}
                  animate={{ opacity: 1, height: 'auto' }}
                  className="mb-6 p-4 bg-slate-700/50 border border-slate-600 rounded-lg"
                >
                  <h3 className="text-lg font-semibold text-white mb-4">Create Alert Rule</h3>
                  <form onSubmit={alertRuleForm.handleSubmit(onAlertRuleSubmit)} className="space-y-4">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-slate-300 mb-2">
                          Rule Name
                        </label>
                        <input
                          {...alertRuleForm.register('name')}
                          type="text"
                          className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                        {alertRuleForm.formState.errors.name && (
                          <p className="mt-1 text-sm text-red-400">
                            {alertRuleForm.formState.errors.name.message}
                          </p>
                        )}
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-slate-300 mb-2">
                          Alert Type
                        </label>
                        <select
                          {...alertRuleForm.register('type')}
                          className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                        >
                          <option value="risk_breach">Risk Breach</option>
                          <option value="price_movement">Price Movement</option>
                          <option value="volume_spike">Volume Spike</option>
                          <option value="position_limit">Position Limit</option>
                        </select>
                      </div>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-slate-300 mb-2">
                          Field
                        </label>
                        <select
                          {...alertRuleForm.register('condition_field')}
                          className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                        >
                          <option value="var_1day">VaR (1 Day)</option>
                          <option value="position_size">Position Size</option>
                          <option value="portfolio_value">Portfolio Value</option>
                          <option value="volatility">Volatility</option>
                        </select>
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-slate-300 mb-2">
                          Operator
                        </label>
                        <select
                          {...alertRuleForm.register('condition_operator')}
                          className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                        >
                          <option value=">">Greater than</option>
                          <option value=">=">Greater than or equal</option>
                          <option value="<">Less than</option>
                          <option value="<=">Less than or equal</option>
                          <option value="==">Equal to</option>
                          <option value="!=">Not equal to</option>
                        </select>
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-slate-300 mb-2">
                          Value
                        </label>
                        <input
                          {...alertRuleForm.register('condition_value', { valueAsNumber: true })}
                          type="number"
                          step="0.01"
                          className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                      </div>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-slate-300 mb-2">
                          Severity
                        </label>
                        <select
                          {...alertRuleForm.register('severity')}
                          className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                        >
                          <option value="low">Low</option>
                          <option value="medium">Medium</option>
                          <option value="high">High</option>
                          <option value="critical">Critical</option>
                        </select>
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-slate-300 mb-2">
                          Cooldown (minutes)
                        </label>
                        <input
                          {...alertRuleForm.register('cooldown_minutes', { valueAsNumber: true })}
                          type="number"
                          min="1"
                          className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                      </div>

                      <div className="flex items-center">
                        <label className="flex items-center mt-8">
                          <input
                            {...alertRuleForm.register('enabled')}
                            type="checkbox"
                            className="w-4 h-4 text-blue-600 bg-slate-700 border-slate-600 rounded focus:ring-blue-500"
                          />
                          <span className="ml-2 text-slate-300">Enabled</span>
                        </label>
                      </div>
                    </div>

                    <div className="flex justify-end space-x-4">
                      <button
                        type="button"
                        onClick={() => setShowAddRule(false)}
                        className="px-4 py-2 text-slate-400 hover:text-white transition-colors"
                      >
                        Cancel
                      </button>
                      <button
                        type="submit"
                        className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                      >
                        Create Rule
                      </button>
                    </div>
                  </form>
                </motion.div>
              )}

              {/* Alert Rules List */}
              <div className="space-y-4">
                {alertRules.map((rule) => (
                  <motion.div
                    key={rule.id}
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    className="bg-slate-700/50 border border-slate-600 rounded-lg p-4"
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-3">
                          <h3 className="text-lg font-semibold text-white">{rule.name}</h3>
                          <span className={`px-2 py-1 text-xs rounded-full border ${getSeverityColor(rule.severity)}`}>
                            {rule.severity.toUpperCase()}
                          </span>
                          <span className="text-slate-400 text-sm capitalize">
                            {rule.type.replace('_', ' ')}
                          </span>
                        </div>
                        <div className="mt-2 text-sm text-slate-300">
                          Trigger when <strong>{rule.condition_field}</strong> {rule.condition_operator} <strong>{rule.condition_value}</strong>
                        </div>
                        <div className="mt-1 text-xs text-slate-400">
                          Cooldown: {rule.cooldown_minutes} minutes
                        </div>
                      </div>
                      <div className="flex items-center space-x-2">
                        <label className="flex items-center">
                          <input
                            type="checkbox"
                            checked={rule.enabled}
                            onChange={(e) => toggleAlertRule(rule.id, e.target.checked)}
                            className="w-4 h-4 text-blue-600 bg-slate-700 border-slate-600 rounded focus:ring-blue-500"
                          />
                          <span className="ml-2 text-slate-300 text-sm">Enabled</span>
                        </label>
                        <button
                          onClick={() => deleteAlertRule(rule.id)}
                          className="p-2 text-red-400 hover:text-red-300 hover:bg-red-500/20 rounded-lg transition-colors"
                        >
                          <TrashIcon className="w-4 h-4" />
                        </button>
                      </div>
                    </div>
                  </motion.div>
                ))}
              </div>
            </div>
          )}

          {activeTab === 'risk' && (
            <div>
              <h2 className="text-xl font-semibold text-white mb-6">Risk Management Limits</h2>
              <form onSubmit={riskLimitsForm.handleSubmit(onRiskLimitsSubmit)} className="space-y-6">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div>
                    <label className="block text-sm font-medium text-slate-300 mb-2">
                      Maximum Position Size (% of portfolio)
                    </label>
                    <input
                      {...riskLimitsForm.register('max_position_size', { valueAsNumber: true })}
                      type="number"
                      step="0.01"
                      min="0.01"
                      max="1"
                      className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-400">
                      Maximum percentage of portfolio in a single position
                    </p>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-slate-300 mb-2">
                      Maximum Sector Concentration (%)
                    </label>
                    <input
                      {...riskLimitsForm.register('max_sector_concentration', { valueAsNumber: true })}
                      type="number"
                      step="0.01"
                      min="0.01"
                      max="1"
                      className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-400">
                      Maximum exposure to any single sector
                    </p>
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div>
                    <label className="block text-sm font-medium text-slate-300 mb-2">
                      VaR Limit ($)
                    </label>
                    <input
                      {...riskLimitsForm.register('var_limit', { valueAsNumber: true })}
                      type="number"
                      min="1000"
                      className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-400">
                      Maximum Value at Risk (1-day, 95% confidence)
                    </p>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-slate-300 mb-2">
                      Maximum Drawdown (%)
                    </label>
                    <input
                      {...riskLimitsForm.register('max_drawdown', { valueAsNumber: true })}
                      type="number"
                      step="0.01"
                      min="0.01"
                      max="1"
                      className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-400">
                      Maximum acceptable portfolio drawdown
                    </p>
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div>
                    <label className="block text-sm font-medium text-slate-300 mb-2">
                      Liquidity Threshold (days)
                    </label>
                    <input
                      {...riskLimitsForm.register('liquidity_threshold', { valueAsNumber: true })}
                      type="number"
                      min="1"
                      max="30"
                      className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-400">
                      Maximum days to liquidate portfolio
                    </p>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-slate-300 mb-2">
                      Volatility Limit (%)
                    </label>
                    <input
                      {...riskLimitsForm.register('volatility_limit', { valueAsNumber: true })}
                      type="number"
                      step="0.01"
                      min="0.05"
                      max="2"
                      className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-400">
                      Maximum portfolio volatility (annualized)
                    </p>
                  </div>
                </div>

                <div className="flex justify-end">
                  <button
                    type="submit"
                    className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                  >
                    Update Risk Limits
                  </button>
                </div>
              </form>
            </div>
          )}

          {activeTab === 'trading' && (
            <div>
              <h2 className="text-xl font-semibold text-white mb-6">Trading Settings</h2>
              <form onSubmit={tradingSettingsForm.handleSubmit(onTradingSettingsSubmit)} className="space-y-6">
                <div className="space-y-4">
                  <div className="flex items-center justify-between p-4 bg-slate-700/50 rounded-lg">
                    <div>
                      <h3 className="text-white font-medium">Auto Rebalancing</h3>
                      <p className="text-slate-400 text-sm">Automatically rebalance portfolio when thresholds are exceeded</p>
                    </div>
                    <label className="flex items-center">
                      <input
                        {...tradingSettingsForm.register('auto_rebalancing')}
                        type="checkbox"
                        className="w-4 h-4 text-blue-600 bg-slate-700 border-slate-600 rounded focus:ring-blue-500"
                      />
                    </label>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-slate-300 mb-2">
                      Rebalancing Threshold (%)
                    </label>
                    <input
                      {...tradingSettingsForm.register('rebalancing_threshold', { valueAsNumber: true })}
                      type="number"
                      step="0.01"
                      min="0.01"
                      max="0.5"
                      className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-400">
                      Trigger rebalancing when allocation deviates by this percentage
                    </p>
                  </div>
                </div>

                <div className="space-y-4">
                  <div className="flex items-center justify-between p-4 bg-slate-700/50 rounded-lg">
                    <div>
                      <h3 className="text-white font-medium">Stop Loss Orders</h3>
                      <p className="text-slate-400 text-sm">Automatically place stop loss orders on new positions</p>
                    </div>
                    <label className="flex items-center">
                      <input
                        {...tradingSettingsForm.register('stop_loss_enabled')}
                        type="checkbox"
                        className="w-4 h-4 text-blue-600 bg-slate-700 border-slate-600 rounded focus:ring-blue-500"
                      />
                    </label>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-slate-300 mb-2">
                      Default Stop Loss (%)
                    </label>
                    <input
                      {...tradingSettingsForm.register('default_stop_loss', { valueAsNumber: true })}
                      type="number"
                      step="0.01"
                      min="0.01"
                      max="0.5"
                      className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </div>
                </div>

                <div className="space-y-4">
                  <div className="flex items-center justify-between p-4 bg-slate-700/50 rounded-lg">
                    <div>
                      <h3 className="text-white font-medium">Take Profit Orders</h3>
                      <p className="text-slate-400 text-sm">Automatically place take profit orders on new positions</p>
                    </div>
                    <label className="flex items-center">
                      <input
                        {...tradingSettingsForm.register('take_profit_enabled')}
                        type="checkbox"
                        className="w-4 h-4 text-blue-600 bg-slate-700 border-slate-600 rounded focus:ring-blue-500"
                      />
                    </label>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-slate-300 mb-2">
                      Default Take Profit (%)
                    </label>
                    <input
                      {...tradingSettingsForm.register('default_take_profit', { valueAsNumber: true })}
                      type="number"
                      step="0.01"
                      min="0.01"
                      max="2"
                      className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="flex items-center justify-between p-4 bg-slate-700/50 rounded-lg">
                    <div>
                      <h3 className="text-white font-medium">Paper Trading Mode</h3>
                      <p className="text-slate-400 text-sm">Use paper trading instead of real money</p>
                    </div>
                    <label className="flex items-center">
                      <input
                        {...tradingSettingsForm.register('paper_trading_mode')}
                        type="checkbox"
                        className="w-4 h-4 text-blue-600 bg-slate-700 border-slate-600 rounded focus:ring-blue-500"
                      />
                    </label>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-slate-300 mb-2">
                      Max Daily Trades
                    </label>
                    <input
                      {...tradingSettingsForm.register('max_daily_trades', { valueAsNumber: true })}
                      type="number"
                      min="1"
                      max="100"
                      className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-400">
                      Maximum number of trades per day
                    </p>
                  </div>
                </div>

                <div className="flex justify-end">
                  <button
                    type="submit"
                    className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                  >
                    Update Trading Settings
                  </button>
                </div>
              </form>
            </div>
          )}

          {activeTab === 'system' && (
            <div>
              <h2 className="text-xl font-semibold text-white mb-6">System Settings</h2>
              <div className="space-y-6">
                <div className="bg-blue-500/20 border border-blue-500/30 rounded-lg p-4">
                  <div className="flex items-center">
                    <InformationCircleIcon className="h-5 w-5 text-blue-400 mr-2" />
                    <h3 className="text-blue-400 font-medium">System Information</h3>
                  </div>
                  <div className="mt-3 grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                    <div>
                      <p className="text-slate-300">Version: <span className="text-white">1.0.0</span></p>
                      <p className="text-slate-300">Environment: <span className="text-white">Production</span></p>
                      <p className="text-slate-300">Last Updated: <span className="text-white">{new Date().toLocaleDateString()}</span></p>
                    </div>
                    <div>
                      <p className="text-slate-300">API Status: <span className="text-green-400">Connected</span></p>
                      <p className="text-slate-300">WebSocket: <span className="text-green-400">Active</span></p>
                      <p className="text-slate-300">Market Data: <span className="text-green-400">Live</span></p>
                    </div>
                  </div>
                </div>

                <div className="bg-yellow-500/20 border border-yellow-500/30 rounded-lg p-4">
                  <div className="flex items-center">
                    <ExclamationTriangleIcon className="h-5 w-5 text-yellow-400 mr-2" />
                    <h3 className="text-yellow-400 font-medium">Maintenance Mode</h3>
                  </div>
                  <p className="mt-2 text-slate-300 text-sm">
                    Enable maintenance mode to temporarily disable trading and alerts.
                  </p>
                  <div className="mt-3 flex items-center">
                    <label className="flex items-center">
                      <input
                        type="checkbox"
                        className="w-4 h-4 text-blue-600 bg-slate-700 border-slate-600 rounded focus:ring-blue-500"
                      />
                      <span className="ml-2 text-slate-300">Enable Maintenance Mode</span>
                    </label>
                  </div>
                </div>

                <div className="bg-red-500/20 border border-red-500/30 rounded-lg p-4">
                  <div className="flex items-center">
                    <ExclamationTriangleIcon className="h-5 w-5 text-red-400 mr-2" />
                    <h3 className="text-red-400 font-medium">Emergency Stop</h3>
                  </div>
                  <p className="mt-2 text-slate-300 text-sm">
                    Immediately stop all trading activities and cancel pending orders.
                  </p>
                  <button className="mt-3 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors">
                    Emergency Stop
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </DashboardLayout>
  );
}
