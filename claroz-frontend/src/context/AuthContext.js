import React, { createContext, useState, useContext, useEffect } from 'react';
import { userAPI } from '../services/api';
import { authService } from '../services/auth';

const AuthContext = createContext(null);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const initAuth = async () => {
      try {
        const response = await userAPI.getCurrentUser();
        setUser(response.data);
      } catch (err) {
        console.error('Failed to fetch user:', err);
      } finally {
        setLoading(false);
      }
    };

    if (localStorage.getItem('token')) {
      initAuth();
    } else {
      setLoading(false);
    }
  }, []);

  const login = async (credentials) => {
    try {
      setError('');
      const response = await authService.login(credentials);
      const userResponse = await userAPI.getCurrentUser();
      setUser(userResponse.data);
      return response;
    } catch (err) {
      setError(err.response?.data?.message || 'Failed to login');
      throw err;
    }
  };

  const register = async (userData) => {
    try {
      setError('');
      console.log('AuthContext - Registering with data:', userData);
      const response = await authService.register(userData);
      console.log('AuthContext - Registration response:', response);
      const userResponse = await userAPI.getCurrentUser();
      setUser(userResponse.data);
      return response;
    } catch (err) {
      console.error('AuthContext - Registration error:', err.response?.data);
      setError(err.response?.data?.error || err.response?.data?.message || 'Failed to register');
      throw err;
    }
  };

  const logout = () => {
    authService.logout();
    setUser(null);
  };

  const updateUser = (userData) => {
    setUser(userData);
  };

  const value = {
    user,
    loading,
    error,
    login,
    register,
    logout,
    updateUser,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export default AuthContext;
