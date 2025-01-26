import React from 'react';
import { render, screen, fireEvent, waitFor, createEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BrowserRouter } from 'react-router-dom';
import { AuthProvider } from '../context/AuthContext';
import Login from './Login';

// Mock the useNavigate hook
const mockNavigate = jest.fn();
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockNavigate,
}));

// Mock the useAuth hook
const mockLogin = jest.fn();
const mockError = null;

jest.mock('../context/AuthContext', () => ({
  useAuth: () => ({
    login: mockLogin,
    error: mockError,
  }),
  AuthProvider: ({ children }) => <div>{children}</div>,
}));

const renderLogin = () => {
  return render(
    <BrowserRouter>
      <AuthProvider>
        <Login />
      </AuthProvider>
    </BrowserRouter>
  );
};

describe('Login', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('renders login form', () => {
    renderLogin();
    
    expect(screen.getByText('Sign in to Claroz')).toBeInTheDocument();
    expect(screen.getByLabelText(/email address/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument();
    expect(screen.getByText(/don't have an account\? sign up/i)).toBeInTheDocument();
  });

  test('handles input changes', async () => {
    renderLogin();
    
    const emailInput = screen.getByLabelText(/email address/i);
    const passwordInput = screen.getByLabelText(/password/i);
    
    await userEvent.type(emailInput, 'test@example.com');
    await userEvent.type(passwordInput, 'password123');
    
    expect(emailInput.value).toBe('test@example.com');
    expect(passwordInput.value).toBe('password123');
  });

  test('submits form with correct data', async () => {
    renderLogin();
    
    const emailInput = screen.getByLabelText(/email address/i);
    const passwordInput = screen.getByLabelText(/password/i);
    const submitButton = screen.getByRole('button', { name: /sign in/i });
    
    await userEvent.type(emailInput, 'test@example.com');
    await userEvent.type(passwordInput, 'password123');
    
    mockLogin.mockResolvedValueOnce();
    
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith({
        email: 'test@example.com',
        password: 'password123',
      });
      expect(mockNavigate).toHaveBeenCalledWith('/');
    });
  });

  test('displays error message when login fails', async () => {
    // Mock the error state
    const mockErrorState = {
      login: mockLogin.mockRejectedValueOnce(new Error('Invalid credentials')),
      error: 'Invalid credentials',
    };

    // Override the mock for this test
    jest.spyOn(require('../context/AuthContext'), 'useAuth')
      .mockImplementation(() => mockErrorState);
    
    renderLogin();
    
    const emailInput = screen.getByLabelText(/email address/i);
    const passwordInput = screen.getByLabelText(/password/i);
    const submitButton = screen.getByRole('button', { name: /sign in/i });
    
    await userEvent.type(emailInput, 'test@example.com');
    await userEvent.type(passwordInput, 'wrongpassword');
    
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalled();
      expect(mockNavigate).not.toHaveBeenCalled();
      // The error message is handled by AuthContext and displayed through Alert
      expect(screen.getByRole('alert')).toBeInTheDocument();
    });
  });

  test('navigates to register page when clicking sign up link', () => {
    renderLogin();
    
    const signUpLink = screen.getByText(/don't have an account\? sign up/i);
    expect(signUpLink.getAttribute('href')).toBe('/register');
  });

  test('validates required fields', async () => {
    renderLogin();
    
    const submitButton = screen.getByRole('button', { name: /sign in/i });
    
    fireEvent.click(submitButton);
    
    // HTML5 validation will prevent form submission and show validation messages
    expect(mockLogin).not.toHaveBeenCalled();
    
    // Check for required field validation
    const emailInput = screen.getByLabelText(/email address/i);
    const passwordInput = screen.getByLabelText(/password/i);
    
    expect(emailInput).toBeRequired();
    expect(passwordInput).toBeRequired();
  });

  test('prevents default form submission', async () => {
    renderLogin();
    
    const form = screen.getByRole('form');
    const submitEvent = createEvent.submit(form);
    
    fireEvent(form, submitEvent);
    
    expect(submitEvent.defaultPrevented).toBeTruthy();
  });
});
