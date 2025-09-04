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
    const response = await api.get('/auth/me');
    return response.data;
  }
};

// Portfolio API
export const portfolioApi = {
  getAll: async (): Promise<Portfolio[]> => {
    const response: AxiosResponse<{ portfolios: Portfolio[] }> = await api.get('/portfolios');
    return response.data.portfolios;
  },

  getById: async (id: string): Promise<Portfolio> => {
    const response: AxiosResponse<Portfolio> = await api.get(`/portfolios/${id}`);
    return response.data;
  },

  create: async (portfolioData: Partial<Portfolio>): Promise<Portfolio> => {
    const response: AxiosResponse<Portfolio> = await api.post('/portfolios', portfolioData);
    return response.data;
  },

  update: async (id: string, portfolioData: Partial<Portfolio>): Promise<Portfolio> => {
    const response: AxiosResponse<Portfolio> = await api.put(`/portfolios/${id}`, portfolioData);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/portfolios/${id}`);
  }
};

// Transaction API
export const transactionApi = {
  getAll: async (filters?: TransactionFilters): Promise<PaginatedResponse<Transaction>> => {
    const params = new URLSearchParams();
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value) params.append(key, value);
      });
    }
    const response: AxiosResponse<PaginatedResponse<Transaction>> = await api.get(
      `/transactions?${params.toString()}`
    );
    return response.data;
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
    const response: AxiosResponse<Alert[]> = await api.get('/alerts');
    return response.data;
  },

  getActive: async (): Promise<Alert[]> => {
    const response: AxiosResponse<Alert[]> = await api.get('/alerts/active');
    return response.data;
  },

  acknowledge: async (id: string): Promise<Alert> => {
    const response: AxiosResponse<Alert> = await api.put(`/alerts/${id}/acknowledge`);
    return response.data;
  },

  resolve: async (id: string, resolution?: string): Promise<Alert> => {
    const response: AxiosResponse<Alert> = await api.put(`/alerts/${id}/resolve`, { resolution });
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/alerts/${id}`);
  }
};

// Dashboard API
export const dashboardApi = {
  getDashboardData: async (): Promise<DashboardData> => {
    try {
      // Fetch all required data in parallel
      const [portfolios, alerts, transactions] = await Promise.all([
        portfolioApi.getAll(),
        alertApi.getActive(),
        transactionApi.getAll({ limit: 10 } as any)
      ]);

      // Calculate totals
      const totalValue = portfolios.reduce((sum, p) => sum + parseFloat(p.total_value), 0);
      const totalPnL = portfolios.reduce((sum, p) => {
        const portfolioPnL = p.positions?.reduce((pSum, pos) => pSum + parseFloat(pos.pnl || '0'), 0) || 0;
        return sum + portfolioPnL;
      }, 0);

      return {
        portfolios,
        total_value: totalValue.toString(),
        total_pnl: totalPnL.toString(),
        total_pnl_percent: totalValue > 0 ? ((totalPnL / totalValue) * 100).toFixed(2) : '0',
        risk_metrics: [], // Will be populated by WebSocket
        recent_transactions: transactions.data || [],
        active_alerts: alerts,
        price_updates: {}
      };
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error);
      throw error;
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

export default api;
