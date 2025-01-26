import axios from 'axios';

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

// Create axios instance with base configuration
const authApi = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Function to get JWT token from localStorage
export const getToken = () => localStorage.getItem('token');

// Function to set JWT token in localStorage
export const setToken = (token) => {
  if (token) {
    localStorage.setItem('token', token);
  } else {
    localStorage.removeItem('token');
  }
};

// Authentication service functions
export const authService = {
  login: async (credentials) => {
    const response = await authApi.post('/auth/login', credentials);
    if (response.data.token) {
      setToken(response.data.token);
    }
    return response.data;
  },

  register: async (userData) => {
    const response = await authApi.post('/auth/register', userData);
    if (response.data.token) {
      setToken(response.data.token);
    }
    return response.data;
  },

  logout: () => {
    setToken(null);
  },

  refreshToken: async () => {
    const token = getToken();
    if (!token) return null;

    try {
      const response = await authApi.post('/auth/refresh', null, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (response.data.token) {
        setToken(response.data.token);
      }
      return response.data;
    } catch (error) {
      setToken(null);
      throw error;
    }
  },
};

export default authService;
