// User and Authentication Types
export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  role: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  user: User;
  token: string;
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterData {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
}

// Portfolio Types
export interface Portfolio {
  id: string;
  user_id: string;
  name: string;
  description: string;
  total_value: string;
  currency: string;
  positions: Position[];
  created_at: string;
  updated_at: string;
}

export interface Position {
  id: string;
  portfolio_id: string;
  symbol: string;
  quantity: string;
  average_price: string;
  current_price: string;
  market_value: string;
  pnl: string;
  pnl_percent: string;
  weight: string;
  asset_type: string;
  liquidity: string;
  updated_at: string;
}

// Transaction Types
export interface Transaction {
  id: string;
  portfolio_id: string;
  transaction_type: 'BUY' | 'SELL';
  symbol: string;
  quantity: string;
  price: string;
  amount: string;
  currency: string;
  status: string;
  executed_at: string;
  kyc_verified: boolean;
  aml_checked: boolean;
  risk_score: number;
  created_at: string;
}

export interface CreateTransactionData {
  portfolio_id: string;
  transaction_type: 'BUY' | 'SELL';
  symbol: string;
  quantity: string;
  price: string;
  currency: string;
}

// Risk Types
export interface VaRResult {
  portfolio_id: string;
  var_value: string;
  var_percentage: number;
  confidence_level: number;
  time_horizon: number;
  method: string;
  status: 'SAFE' | 'WARNING' | 'CRITICAL';
  threshold: string;
  calculated_at: string;
}

export interface LiquidityResult {
  portfolio_id: string;
  liquidity_ratio: string;
  liquidity_score: string;
  days_to_liquidate: string;
  risk_assessment: 'LOW_RISK' | 'MEDIUM_RISK' | 'HIGH_RISK';
  calculated_at: string;
  breakdown: {
    HIGH: string;
    MEDIUM: string;
    LOW: string;
  };
}

export interface RiskMetric {
  id: string;
  portfolio_id: string;
  metric_type: string;
  value: string;
  threshold: string;
  status: 'SAFE' | 'WARNING' | 'CRITICAL';
  time_horizon?: number;
  confidence_level?: string;
  details: Record<string, any>;
  calculated_at: string;
}

// Alert Types
export interface Alert {
  id: string;
  portfolio_id: string;
  alert_type: string;
  severity: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
  title: string;
  description: string;
  source: string;
  status: 'ACTIVE' | 'ACKNOWLEDGED' | 'RESOLVED';
  triggered_by: Record<string, any>;
  resolution?: string;
  acknowledged_by?: string;
  acknowledged_at?: string;
  resolved_by?: string;
  resolved_at?: string;
  created_at: string;
  updated_at: string;
}

// WebSocket Types
export interface WebSocketMessage {
  type: 'price_update' | 'risk_update' | 'new_transaction' | 'new_alert' | 'aml_alert';
  data: any;
}

export interface PriceUpdate {
  [symbol: string]: {
    price: number;
    change: number;
    change_percent: number;
    volume: number;
    timestamp: number;
    provider: string;
    is_market_open: boolean;
  };
}

export interface RiskUpdate {
  portfolio_id: string;
  var: VaRResult;
  liquidity: LiquidityResult;
  timestamp: number;
}

// Chart Data Types
export interface ChartDataPoint {
  date: string;
  value: number;
  label?: string;
}

export interface PerformanceData {
  portfolio_value: ChartDataPoint[];
  var_history: ChartDataPoint[];
  liquidity_history: ChartDataPoint[];
}

// Dashboard Types
export interface DashboardData {
  portfolios: Portfolio[];
  total_value: string;
  total_pnl: string;
  total_pnl_percent: string;
  risk_metrics: RiskMetric[];
  recent_transactions: Transaction[];
  active_alerts: Alert[];
  price_updates: PriceUpdate;
}

// API Response Types
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    pages: number;
  };
}

// Filter and Sort Types
export interface TransactionFilters {
  portfolio_id?: string;
  transaction_type?: 'BUY' | 'SELL';
  symbol?: string;
  status?: string;
  date_from?: string;
  date_to?: string;
}

export interface SortOption {
  field: string;
  direction: 'asc' | 'desc';
}

// Theme and UI Types
export type Theme = 'light' | 'dark';

export interface UIState {
  theme: Theme;
  sidebarOpen: boolean;
  loading: boolean;
  error: string | null;
}
