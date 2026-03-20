import React, { createContext, useContext, useState, useEffect } from 'react';
import { userClient, cardClient, createOptions } from '../api/client';
import { User } from '../api/ruthless';
import { ConfigPublic_Limits, ConfigPublic_AuthProvider } from '../api/config';

interface AuthContextType {
  user: User | null;
  token: string | null;
  loading: boolean;
  isDevelopment: boolean;
  authProvider: ConfigPublic_AuthProvider;
  limits: ConfigPublic_Limits | null;
  loginWithToken: (token: string) => void;
  completeRegistration: (name: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(localStorage.getItem('ruthless_token'));
  const [loading, setLoading] = useState(true);
  const [isDevelopment, setIsDevelopment] = useState(false);
  const [authProvider, setAuthProvider] = useState<ConfigPublic_AuthProvider>(ConfigPublic_AuthProvider.UNSPECIFIED);
  const [limits, setLimits] = useState<ConfigPublic_Limits | null>(null);

  useEffect(() => {
    const fetchConfig = async () => {
      try {
        const response = await cardClient.getConfig({}, {});
        // GetConfig now returns ConfigPublic, so fields are directly on the response
        setIsDevelopment(response.response.isDevelopment || false);
        setAuthProvider(response.response.authProvider);
        if (response.response.limits) {
          setLimits(response.response.limits);
        }

      } catch (error) {
        console.error('Failed to fetch server config:', error);
      }
    };
    fetchConfig();

    // Check search params first for backwards compatibility
    const params = new URLSearchParams(window.location.search);
    let urlToken = params.get('token');

    // Also check hash fragment
    if (!urlToken && window.location.hash) {
      // Remove the leading '#'
      const hashParams = new URLSearchParams(window.location.hash.substring(1));
      urlToken = hashParams.get('token');
    }

    if (urlToken) {
      setToken(urlToken);
      localStorage.setItem('ruthless_token', urlToken);
      
      // Remove token from URL
      const url = new URL(window.location.href);
      url.searchParams.delete('token');
      url.hash = ''; // Clear hash
      window.history.replaceState({}, '', url.toString());
    }
  }, []);

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


  const completeRegistration = async (name: string) => {
    if (!token) return;
    setLoading(true);
    try {
      const response = await userClient.completeRegistration({ name }, createOptions(token));
      setUser(response.response);
    } catch (error: any) {
      console.error('Complete registration failed:', error);
      throw error;
    } finally {
      setLoading(false);
    }
  };

  const loginWithToken = (newToken: string) => {
    setToken(newToken);
    localStorage.setItem('ruthless_token', newToken);
  };

  const logout = async () => {
    if (token) {
      try {
        const baseUrl = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";
        await fetch(`${baseUrl}/auth/logout`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`
          }
        });
      } catch (error) {
        console.error('Logout failed:', error);
      }
    }
    
    setToken(null);
    setUser(null);
    localStorage.removeItem('ruthless_token');
  };

  return (
    <AuthContext.Provider value={{ user, token, loading, isDevelopment, authProvider, limits, loginWithToken, completeRegistration, logout }}>
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
