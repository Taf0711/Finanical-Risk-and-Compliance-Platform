'use client';

import React, { useState, useEffect, createContext, useContext, ReactNode } from 'react';
import { useRouter } from 'next/navigation';
import Cookies from 'js-cookie';
import { authApi } from '@/lib/api';
import { User, LoginCredentials, RegisterData } from '@/types';

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (credentials: LoginCredentials) => Promise<void>;
  register: (userData: RegisterData) => Promise<void>;
  logout: () => void;
  isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [mounted, setMounted] = useState(false);
  const router = useRouter();

  useEffect(() => {
    setMounted(true);
    checkAuthStatus();
  }, []);

  const checkAuthStatus = async () => {
    try {
      const token = Cookies.get('auth_token');
      if (token) {
        // Verify token with backend
        const userData = await authApi.getCurrentUser();
        setUser(userData);
      }
    } catch (error) {
      console.error('Auth check failed:', error);
      Cookies.remove('auth_token');
    } finally {
      setLoading(false);
    }
  };

  const login = async (credentials: LoginCredentials) => {
    try {
      setLoading(true);
      const response = await authApi.login(credentials);
      
      if (response.token && response.user) {
        // Store token in secure cookie
        Cookies.set('auth_token', response.token, { 
          expires: 1, // 1 day
          secure: process.env.NODE_ENV === 'production',
          sameSite: 'strict'
        });
        
        setUser(response.user);
        router.push('/dashboard');
      } else {
        throw new Error('Login failed');
      }
    } catch (error: any) {
      console.error('Login error:', error);
      throw new Error(error.response?.data?.error || 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  const register = async (userData: RegisterData) => {
    try {
      setLoading(true);
      const response = await authApi.register(userData);
      
      if (response.token && response.user) {
        // Store token in secure cookie
        Cookies.set('auth_token', response.token, { 
          expires: 1, // 1 day
          secure: process.env.NODE_ENV === 'production',
          sameSite: 'strict'
        });
        
        setUser(response.user);
        router.push('/dashboard');
      } else {
        throw new Error('Registration failed');
      }
    } catch (error: any) {
      console.error('Registration error:', error);
      throw new Error(error.response?.data?.error || 'Registration failed');
    } finally {
      setLoading(false);
    }
  };

  const logout = () => {
    Cookies.remove('auth_token');
    setUser(null);
    router.push('/login');
  };

  const value: AuthContextType = {
    user,
    loading,
    login,
    register,
    logout,
    isAuthenticated: !!user
  };

  // Prevent hydration mismatch
  if (!mounted) {
    return React.createElement(
      AuthContext.Provider,
      { value: { ...value, loading: true } },
      children
    );
  }

  return React.createElement(
    AuthContext.Provider,
    { value },
    children
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

// Hook for protected routes
export function useRequireAuth() {
  const { user, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading && !user) {
      router.push('/login');
    }
  }, [user, loading, router]);

  return { user, loading };
}
