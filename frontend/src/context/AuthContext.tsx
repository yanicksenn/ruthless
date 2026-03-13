import React, { createContext, useContext, useState, useEffect } from 'react';
import { userClient, createOptions } from '../api/client';
import { User } from '../api/ruthless';

interface AuthContextType {
  user: User | null;
  token: string | null;
  loading: boolean;
  login: (token: string) => Promise<void>;
  register: (token: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(localStorage.getItem('ruthless_token'));
  const [loading, setLoading] = useState(true);

  const fetchUser = async (authToken: string) => {
    try {
      const response = await userClient.getMe({}, createOptions(authToken));
      setUser(response.response);
    } catch (error: any) {
      console.log('Fetch user failed:', { code: error.code, message: error.message });
      // Clear token if user is not found or permission denied
      setToken(null);
      localStorage.removeItem('ruthless_token');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (token) {
      fetchUser(token);
    } else {
      setLoading(false);
    }
  }, [token]);

  const login = async (newToken: string) => {
    setLoading(true);
    try {
      const response = await userClient.getMe({}, createOptions(newToken));
      setUser(response.response);
      setToken(newToken);
      localStorage.setItem('ruthless_token', newToken);
    } catch (error: any) {
      console.error('Login failed:', error);
      throw error;
    } finally {
      setLoading(false);
    }
  };

  const register = async (newToken: string) => {
    setLoading(true);
    try {
      const response = await userClient.register({}, createOptions(newToken));
      setUser(response.response);
      setToken(newToken);
      localStorage.setItem('ruthless_token', newToken);
    } catch (error: any) {
      console.error('Registration failed:', error);
      throw error;
    } finally {
      setLoading(false);
    }
  };

  const logout = () => {
    setToken(null);
    setUser(null);
    localStorage.removeItem('ruthless_token');
  };

  return (
    <AuthContext.Provider value={{ user, token, loading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
