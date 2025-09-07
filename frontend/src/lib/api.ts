import axios, { AxiosResponse } from 'axios';
import Cookies from 'js-cookie';
import { 
  AuthResponse, 
  LoginCredentials, 
  RegisterData, 
  Portfolio, 
  Transaction, 
  CreateTransactionData,
  VaRResult,
  LiquidityResult,
  RiskMetric,
  Alert,
  ApiResponse,
  PaginatedResponse,
  TransactionFilters,
  DashboardData
} from '@/types';

// API Configuration
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

// Create axios instance
const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = Cookies.get('auth_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Token expired or invalid
      Cookies.remove('auth_token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// Authentication API
export const authApi = {
  login: async (credentials: LoginCredentials): Promise<AuthResponse> => {
    const response: AxiosResponse<AuthResponse> = await api.post('/auth/login', credentials);
    return response.data;
  },

  register: async (userData: RegisterData): Promise<AuthResponse> => {
    const response: AxiosResponse<AuthResponse> = await api.post('/auth/register', userData);
    return response.data;
  },

  logout: async (): Promise<void> => {
    await api.post('/auth/logout');
    Cookies.remove('auth_token');
  },

  getCurrentUser: async (): Promise<any> => {
    const response = await api.get('/protected/auth/me');
    return response.data;
  }
};

// Portfolio API
export const portfolioApi = {
  getAll: async (): Promise<Portfolio[]> => {
    try {
      const response: AxiosResponse<Portfolio[] | { portfolios: Portfolio[] }> = await api.get('/protected/portfolios');
      // Handle both array and object responses
      return Array.isArray(response.data) ? response.data : (response.data as any).portfolios || [];
    } catch (error) {
      console.error('Failed to fetch portfolios:', error);
      throw new Error('Unable to load portfolios. Please try again.');
    }
  },

  getById: async (id: string): Promise<Portfolio> => {
    try {
      const response: AxiosResponse<Portfolio> = await api.get(`/protected/portfolios/${id}`);
      return response.data;
    } catch (error) {
      console.error(`Failed to fetch portfolio ${id}:`, error);
      throw new Error('Unable to load portfolio details. Please try again.');
    }
  },

  create: async (portfolioData: Partial<Portfolio>): Promise<Portfolio> => {
    try {
      const response: AxiosResponse<Portfolio> = await api.post('/protected/portfolios', portfolioData);
      return response.data;
    } catch (error) {
      console.error('Failed to create portfolio:', error);
      throw new Error('Unable to create portfolio. Please check your data and try again.');
    }
  },

  update: async (id: string, portfolioData: Partial<Portfolio>): Promise<Portfolio> => {
    try {
      const response: AxiosResponse<Portfolio> = await api.put(`/protected/portfolios/${id}`, portfolioData);
      return response.data;
    } catch (error) {
      console.error(`Failed to update portfolio ${id}:`, error);
      throw new Error('Unable to update portfolio. Please try again.');
    }
  },

  delete: async (id: string): Promise<void> => {
    try {
      await api.delete(`/protected/portfolios/${id}`);
    } catch (error) {
      console.error(`Failed to delete portfolio ${id}:`, error);
      throw new Error('Unable to delete portfolio. Please try again.');
    }
  }
};

// Transaction API
export const transactionApi = {
  getAll: async (filters?: TransactionFilters): Promise<Transaction[]> => {
    try {
      const params = new URLSearchParams();
      if (filters) {
        Object.entries(filters).forEach(([key, value]) => {
          if (value) params.append(key, value);
        });
      }
      const response: AxiosResponse<Transaction[] | { data: Transaction[] }> = await api.get(
        `/protected/transactions?${params.toString()}`
      );
      // Handle both array and paginated responses
      return Array.isArray(response.data) ? response.data : (response.data as any).data || [];
    } catch (error) {
      console.error('Failed to fetch transactions:', error);
      return []; // Return empty array for non-critical data
    }
  },

  getById: async (id: string): Promise<Transaction> => {
    const response: AxiosResponse<Transaction> = await api.get(`/transactions/${id}`);
    return response.data;
  },

  create: async (transactionData: CreateTransactionData): Promise<Transaction> => {
    const response: AxiosResponse<Transaction> = await api.post('/transactions', transactionData);
    return response.data;
  },

  update: async (id: string, transactionData: Partial<Transaction>): Promise<Transaction> => {
    const response: AxiosResponse<Transaction> = await api.put(`/transactions/${id}`, transactionData);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/transactions/${id}`);
  }
};

// Risk API
export const riskApi = {
  getVaR: async (portfolioId: string): Promise<VaRResult> => {
    const response: AxiosResponse<VaRResult> = await api.get(`/risk/portfolio/${portfolioId}/var`);
    return response.data;
  },

  getLiquidityRisk: async (portfolioId: string): Promise<LiquidityResult> => {
    const response: AxiosResponse<LiquidityResult> = await api.get(`/risk/portfolio/${portfolioId}/liquidity`);
    return response.data;
  },

  getRiskMetrics: async (portfolioId: string): Promise<RiskMetric[]> => {
    const response: AxiosResponse<RiskMetric[]> = await api.get(`/risk/portfolio/${portfolioId}/metrics`);
    return response.data;
  },

  getRiskHistory: async (portfolioId: string, metricType?: string): Promise<any[]> => {
    const params = metricType ? `?metric_type=${metricType}` : '';
    const response = await api.get(`/risk/portfolio/${portfolioId}/history${params}`);
    return response.data;
  }
};

// Monte Carlo API
export const monteCarloApi = {
  runSimulation: async (portfolioId: string, config?: any): Promise<any> => {
    const response = await api.post(`/monte-carlo/portfolio/${portfolioId}/simulation`, config);
    return response.data;
  },

  runQuickValidation: async (portfolioId: string): Promise<any> => {
    const response = await api.get(`/monte-carlo/portfolio/${portfolioId}/validation`);
    return response.data;
  },

  getSimulationHistory: async (portfolioId: string): Promise<any[]> => {
    const response = await api.get(`/monte-carlo/portfolio/${portfolioId}/history`);
    return response.data;
  },

  compareSimulations: async (portfolioId: string, scenarios: any[]): Promise<any> => {
    const response = await api.post(`/monte-carlo/portfolio/${portfolioId}/compare`, { scenarios });
    return response.data;
  }
};

// Alert API
export const alertApi = {
  getAll: async (): Promise<Alert[]> => {
    try {
      const response: AxiosResponse<Alert[]> = await api.get('/alerts');
      return Array.isArray(response.data) ? response.data : [];
    } catch (error) {
      console.error('Failed to fetch alerts:', error);
      throw new Error('Unable to load alerts. Please try again.');
    }
  },

  getActive: async (): Promise<Alert[]> => {
    try {
      const response: AxiosResponse<Alert[]> = await api.get('/protected/alerts');
      // Filter for active alerts since backend might not have /active endpoint
      const alerts = Array.isArray(response.data) ? response.data : [];
      return alerts.filter(alert => alert.status === 'ACTIVE');
    } catch (error) {
      console.error('Failed to fetch active alerts:', error);
      return []; // Return empty array instead of throwing for non-critical data
    }
  },

  acknowledge: async (id: string): Promise<Alert> => {
    try {
      const response: AxiosResponse<Alert> = await api.put(`/alerts/${id}/acknowledge`);
      return response.data;
    } catch (error) {
      console.error(`Failed to acknowledge alert ${id}:`, error);
      throw new Error('Unable to acknowledge alert. Please try again.');
    }
  },

  resolve: async (id: string, resolution?: string): Promise<Alert> => {
    try {
      const response: AxiosResponse<Alert> = await api.put(`/alerts/${id}/resolve`, { resolution });
      return response.data;
    } catch (error) {
      console.error(`Failed to resolve alert ${id}:`, error);
      throw new Error('Unable to resolve alert. Please try again.');
    }
  },

  delete: async (id: string): Promise<void> => {
    try {
      await api.delete(`/alerts/${id}`);
    } catch (error) {
      console.error(`Failed to delete alert ${id}:`, error);
      throw new Error('Unable to delete alert. Please try again.');
    }
  }
};

// Dashboard API
export const dashboardApi = {
  getDashboardData: async (): Promise<DashboardData> => {
    try {
      // Fetch all required data in parallel with proper error handling
      const results = await Promise.allSettled([
        portfolioApi.getAll(),
        alertApi.getActive(),
        transactionApi.getAll()
      ]);

      // Extract successful results or use defaults
      const portfolios = results[0].status === 'fulfilled' ? results[0].value : [];
      const alerts = results[1].status === 'fulfilled' ? results[1].value : [];
      const transactions = results[2].status === 'fulfilled' ? results[2].value : [];

      // Log any failed requests
      results.forEach((result, index) => {
        if (result.status === 'rejected') {
          const endpoints = ['portfolios', 'alerts', 'transactions'];
          console.warn(`Failed to fetch ${endpoints[index]}:`, result.reason);
        }
      });

      // Calculate totals with error handling
      const totalValue = portfolios.reduce((sum, p) => {
        const value = parseFloat(p.total_value);
        return sum + (isNaN(value) ? 0 : value);
      }, 0);

      const totalPnL = portfolios.reduce((sum, p) => {
        const portfolioPnL = p.positions?.reduce((pSum, pos) => {
          const pnl = parseFloat(pos.pnl || '0');
          return pSum + (isNaN(pnl) ? 0 : pnl);
        }, 0) || 0;
        return sum + portfolioPnL;
      }, 0);

      return {
        portfolios,
        total_value: totalValue.toString(),
        total_pnl: totalPnL.toString(),
        total_pnl_percent: totalValue > 0 ? ((totalPnL / totalValue) * 100).toFixed(2) : '0',
        risk_metrics: [], // Will be populated by WebSocket
        recent_transactions: transactions.slice(0, 10), // Ensure we only show 10
        active_alerts: alerts.slice(0, 20), // Limit alerts for performance
        price_updates: {}
      };
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error);
      // Return minimal working data structure instead of throwing
      return {
        portfolios: [],
        total_value: '0',
        total_pnl: '0',
        total_pnl_percent: '0',
        risk_metrics: [],
        recent_transactions: [],
        active_alerts: [],
        price_updates: {}
      };
    }
  }
};

// WebSocket connection
export class WebSocketClient {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;

  constructor(userId?: string) {
    const wsUrl = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080';
    this.url = userId ? `${wsUrl}/ws?user_id=${userId}` : `${wsUrl}/ws`;
  }

  connect(onMessage: (data: any) => void, onError?: (error: Event) => void): void {
    try {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.reconnectAttempts = 0;
      };

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          onMessage(data);
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      this.ws.onclose = () => {
        console.log('WebSocket disconnected');
        this.reconnect(onMessage, onError);
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        if (onError) onError(error);
      };
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      if (onError) onError(error as Event);
    }
  }

  private reconnect(onMessage: (data: any) => void, onError?: (error: Event) => void): void {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      setTimeout(() => {
        console.log(`Attempting to reconnect... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
        this.connect(onMessage, onError);
      }, this.reconnectDelay * this.reconnectAttempts);
    }
  }

  disconnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  send(data: any): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }
}

// Trading API
export const tradingApi = {
  // Account operations
  getAccount: async (): Promise<any> => {
    const response = await api.get('/trading/account');
    return response.data.account;
  },

  getTradingStatus: async (): Promise<any> => {
    const response = await api.get('/trading/status');
    return response.data.status;
  },

  // Order operations
  placeOrder: async (orderData: any): Promise<any> => {
    const response = await api.post('/trading/orders', orderData);
    return response.data.order;
  },

  getOrders: async (status?: string, limit?: number): Promise<any[]> => {
    const params = new URLSearchParams();
    if (status) params.append('status', status);
    if (limit) params.append('limit', limit.toString());
    
    const response = await api.get(`/trading/orders?${params.toString()}`);
    return response.data.orders || [];
  },

  getOrder: async (orderId: string): Promise<any> => {
    const response = await api.get(`/trading/orders/${orderId}`);
    return response.data.order;
  },

  cancelOrder: async (orderId: string): Promise<void> => {
    await api.delete(`/trading/orders/${orderId}`);
  },

  validateOrder: async (orderData: any): Promise<any> => {
    const response = await api.post('/trading/orders/validate', orderData);
    return response.data;
  },

  // Position operations
  getPositions: async (): Promise<any[]> => {
    const response = await api.get('/trading/positions');
    return response.data.positions || [];
  },

  getPosition: async (symbol: string): Promise<any> => {
    const response = await api.get(`/trading/positions/${symbol}`);
    return response.data.position;
  },

  closePosition: async (symbol: string): Promise<any> => {
    const response = await api.delete(`/trading/positions/${symbol}`);
    return response.data.order;
  },
};

export default api;
