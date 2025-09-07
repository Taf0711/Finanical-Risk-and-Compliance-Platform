// API Service for Financial Risk Monitor
// Handles all backend API interactions with proper error handling and typing

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Types
export interface Portfolio {
  id: string;
  user_id: string;
  name: string;
  description: string;
  currency: string;
  total_value: string;
  created_at: string;
  updated_at: string;
  positions?: Position[];
}

export interface Position {
  id: string;
  portfolio_id: string;
  symbol: string;
  asset_type: string;
  quantity: string;
  average_price: string;
  current_price: string;
  market_value: string;
  unrealized_pnl: string;
  currency: string;
  liquidity: string;
  created_at: string;
  updated_at: string;
}

export interface Transaction {
  id: string;
  portfolio_id: string;
  transaction_type: string;
  symbol: string;
  quantity: string;
  price: string;
  amount: string;
  currency: string;
  status: string;
  executed_at: string;
  notes?: string;
  kyc_verified: boolean;
  aml_checked: boolean;
  risk_score: number;
}

export interface FundTransactionRequest {
  amount: string;
  type: 'deposit' | 'withdrawal';
  description?: string;
  method: 'bank_transfer' | 'wire' | 'credit_card';
}

export interface FundTransactionResponse {
  transaction_id: string;
  portfolio_id: string;
  amount: string;
  type: string;
  status: string;
  method: string;
  description?: string;
  processed_at: string;
  new_balance: string;
}

export interface OrderRequest {
  symbol: string;
  quantity: number;
  side: 'buy' | 'sell';
  order_type: 'market' | 'limit' | 'stop' | 'stop_limit';
  time_in_force: 'day' | 'gtc' | 'ioc' | 'fok';
  limit_price?: number;
  stop_price?: number;
  portfolio_id: string;
}

export interface OrderResponse {
  id: string;
  symbol: string;
  quantity: string;
  side: string;
  order_type: string;
  time_in_force: string;
  limit_price?: string;
  stop_price?: string;
  status: string;
  filled_quantity: string;
  filled_price?: string;
  submitted_at: string;
  filled_at?: string;
  cancelled_at?: string;
  expired_at?: string;
  asset_class: string;
  extended_hours: boolean;
}

export interface Quote {
  symbol: string;
  price: number;
  change: number;
  change_percent: number;
  volume: number;
  timestamp: string;
}

export interface RiskMetrics {
  portfolio_id: string;
  var_95: string;
  var_99: string;
  liquidity_ratio: string;
  concentration_risk: string;
  calculated_at: string;
}

// API Client Class
class ApiClient {
  private baseURL: string;
  private token: string | null = null;

  constructor(baseURL: string) {
    this.baseURL = baseURL;
    
    // Initialize token from localStorage if available
    if (typeof window !== 'undefined') {
      this.token = localStorage.getItem('auth_token');
    }
  }

  setToken(token: string) {
    this.token = token;
    if (typeof window !== 'undefined') {
      localStorage.setItem('auth_token', token);
    }
  }

  clearToken() {
    this.token = null;
    if (typeof window !== 'undefined') {
      localStorage.removeItem('auth_token');
    }
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    
    const config: RequestInit = {
      headers: {
        'Content-Type': 'application/json',
        ...(this.token && { Authorization: `Bearer ${this.token}` }),
        ...options.headers,
      },
      ...options,
    };

    try {
      const response = await fetch(url, config);
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      return data;
    } catch (error) {
      console.error(`API Error [${endpoint}]:`, error);
      throw error;
    }
  }

  // Portfolio API methods
  async getPortfolios(): Promise<Portfolio[]> {
    return this.request<Portfolio[]>('/api/v1/protected/portfolios');
  }

  async getPortfolio(portfolioId: string): Promise<Portfolio> {
    return this.request<Portfolio>(`/api/v1/protected/portfolios/${portfolioId}`);
  }

  async createPortfolio(data: {
    name: string;
    description?: string;
    currency?: string;
  }): Promise<Portfolio> {
    return this.request<Portfolio>('/api/v1/protected/portfolios', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updatePortfolio(portfolioId: string, data: {
    name?: string;
    description?: string;
  }): Promise<{ message: string; data: Portfolio }> {
    return this.request(`/api/v1/protected/portfolios/${portfolioId}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deletePortfolio(portfolioId: string): Promise<{ message: string }> {
    return this.request(`/api/v1/protected/portfolios/${portfolioId}`, {
      method: 'DELETE',
    });
  }

  // Fund Management
  async addFunds(portfolioId: string, data: FundTransactionRequest): Promise<{ message: string; data: FundTransactionResponse }> {
    return this.request(`/api/v1/protected/portfolios/${portfolioId}/add-funds`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async withdrawFunds(portfolioId: string, data: FundTransactionRequest): Promise<{ message: string; data: FundTransactionResponse }> {
    return this.request(`/api/v1/protected/portfolios/${portfolioId}/withdraw-funds`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getCashBalance(portfolioId: string): Promise<{ portfolio_id: string; cash_balance: string; currency: string }> {
    return this.request(`/api/v1/protected/portfolios/${portfolioId}/cash-balance`);
  }

  async getTransactionHistory(portfolioId: string, limit?: number): Promise<{ portfolio_id: string; transactions: Transaction[]; total_records: number }> {
    const params = limit ? `?limit=${limit}` : '';
    return this.request(`/api/v1/protected/portfolios/${portfolioId}/transactions${params}`);
  }

  // Trading API methods
  async placeOrder(orderData: OrderRequest): Promise<OrderResponse> {
    return this.request<OrderResponse>('/api/v1/protected/trading/orders', {
      method: 'POST',
      body: JSON.stringify(orderData),
    });
  }

  async getOrders(status?: string, limit?: number): Promise<OrderResponse[]> {
    const params = new URLSearchParams();
    if (status) params.append('status', status);
    if (limit) params.append('limit', limit.toString());
    
    const queryString = params.toString();
    return this.request<OrderResponse[]>(`/api/v1/protected/trading/orders${queryString ? `?${queryString}` : ''}`);
  }

  async getOrder(orderId: string): Promise<OrderResponse> {
    return this.request<OrderResponse>(`/api/v1/protected/trading/orders/${orderId}`);
  }

  async cancelOrder(orderId: string): Promise<{ message: string }> {
    return this.request(`/api/v1/protected/trading/orders/${orderId}`, {
      method: 'DELETE',
    });
  }

  async getPositions(): Promise<any[]> {
    return this.request<any[]>('/api/v1/protected/trading/positions');
  }

  async getPosition(symbol: string): Promise<any> {
    return this.request(`/api/v1/protected/trading/positions/${symbol}`);
  }

  async closePosition(symbol: string): Promise<OrderResponse> {
    return this.request<OrderResponse>(`/api/v1/protected/trading/positions/${symbol}`, {
      method: 'DELETE',
    });
  }

  async getTradingAccount(): Promise<any> {
    return this.request('/api/v1/protected/trading/account');
  }

  // Market Data API methods
  async getQuote(symbol: string): Promise<Quote> {
    return this.request<Quote>(`/api/v1/marketdata/quote/${symbol}`);
  }

  async getMultipleQuotes(symbols: string[]): Promise<{ [symbol: string]: Quote }> {
    const symbolsParam = symbols.join(',');
    return this.request(`/api/v1/marketdata/quotes?symbols=${symbolsParam}`);
  }

  async getHistoricalData(symbol: string, period: string = '1mo'): Promise<any> {
    return this.request(`/api/v1/marketdata/historical/${symbol}?period=${period}`);
  }

  async getCompanyInfo(symbol: string): Promise<any> {
    return this.request(`/api/v1/marketdata/company/${symbol}`);
  }

  async searchSymbols(query: string): Promise<any[]> {
    return this.request(`/api/v1/marketdata/search?q=${encodeURIComponent(query)}`);
  }

  async validateSymbol(symbol: string): Promise<{ valid: boolean; symbol: string }> {
    return this.request(`/api/v1/marketdata/validate/${symbol}`);
  }

  // Risk Management API methods
  async calculateVaR(portfolioId: string, options: {
    confidence_level?: number;
    time_horizon?: number;
    method?: string;
  } = {}): Promise<any> {
    return this.request(`/api/v1/protected/risk/var/${portfolioId}`, {
      method: 'POST',
      body: JSON.stringify(options),
    });
  }

  async getLiquidityRisk(portfolioId: string): Promise<any> {
    return this.request(`/api/v1/protected/risk/liquidity/${portfolioId}`);
  }

  async checkPositionLimits(portfolioId: string, maxLimitPercent: number = 25): Promise<any> {
    return this.request(`/api/v1/protected/risk/position-limits/${portfolioId}?max_limit=${maxLimitPercent}`);
  }

  async runMonteCarloSimulation(portfolioId: string, options: {
    simulations?: number;
    time_horizon?: number;
    confidence_levels?: number[];
  } = {}): Promise<any> {
    return this.request(`/api/v1/protected/monte-carlo/simulate/${portfolioId}`, {
      method: 'POST',
      body: JSON.stringify(options),
    });
  }

  // Alerts API methods
  async getAlerts(portfolioId?: string): Promise<any[]> {
    const params = portfolioId ? `?portfolio_id=${portfolioId}` : '';
    return this.request(`/api/v1/protected/alerts${params}`);
  }

  async createAlert(alertData: any): Promise<any> {
    return this.request('/api/v1/protected/alerts', {
      method: 'POST',
      body: JSON.stringify(alertData),
    });
  }

  async updateAlert(alertId: string, data: any): Promise<any> {
    return this.request(`/api/v1/protected/alerts/${alertId}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteAlert(alertId: string): Promise<{ message: string }> {
    return this.request(`/api/v1/protected/alerts/${alertId}`, {
      method: 'DELETE',
    });
  }

  // Compliance API methods
  async getComplianceStatus(portfolioId: string): Promise<any> {
    return this.request(`/api/v1/protected/compliance/status/${portfolioId}`);
  }

  async runKYCAMLCheck(transactionId: string): Promise<any> {
    return this.request(`/api/v1/protected/compliance/kyc-aml/${transactionId}`, {
      method: 'POST',
    });
  }
}

// Create singleton instance
const apiClient = new ApiClient(API_BASE_URL);

export default apiClient;

// Export convenience methods
export const {
  setToken,
  clearToken,
  getPortfolios,
  getPortfolio,
  createPortfolio,
  updatePortfolio,
  deletePortfolio,
  addFunds,
  withdrawFunds,
  getCashBalance,
  getTransactionHistory,
  placeOrder,
  getOrders,
  getOrder,
  cancelOrder,
  getPositions,
  getPosition,
  closePosition,
  getTradingAccount,
  getQuote,
  getMultipleQuotes,
  getHistoricalData,
  getCompanyInfo,
  searchSymbols,
  validateSymbol,
  calculateVaR,
  getLiquidityRisk,
  checkPositionLimits,
  runMonteCarloSimulation,
  getAlerts,
  createAlert,
  updateAlert,
  deleteAlert,
  getComplianceStatus,
  runKYCAMLCheck,
} = apiClient;

