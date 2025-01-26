import React from 'react';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { ThemeProvider } from '@mui/material';
import { AuthProvider } from './context/AuthContext';
import theme from './theme/theme';
import App from './App';

// Mock the useAuth hook
jest.mock('./context/AuthContext', () => ({
  useAuth: () => ({
    isAuthenticated: false,
    user: null,
  }),
  AuthProvider: ({ children }) => <div>{children}</div>,
}));

// Mock the router hooks
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => jest.fn(),
}));

test('renders app without crashing', () => {
  render(
    <BrowserRouter>
      <ThemeProvider theme={theme}>
        <AuthProvider>
          <App />
        </AuthProvider>
      </ThemeProvider>
    </BrowserRouter>
  );
  
  // Basic smoke test
  expect(screen.getByRole('main')).toBeInTheDocument();
});
