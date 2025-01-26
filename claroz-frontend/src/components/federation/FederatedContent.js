import React, { useState } from 'react';
import {
  TextField,
  Button,
  Box,
  CircularProgress,
  Alert,
  Typography,
} from '@mui/material';
import { Search as SearchIcon } from '@mui/icons-material';
import { federationAPI } from '../../services/api';

function FederatedContent({ onProfileResolved }) {
  const [handle, setHandle] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleResolve = async (e) => {
    e.preventDefault();
    if (!handle.trim()) return;

    try {
      setLoading(true);
      setError('');
      const response = await federationAPI.resolveProfile(handle);
      if (onProfileResolved) {
        onProfileResolved(response.data);
      }
      setHandle('');
    } catch (err) {
      setError(err.response?.data?.message || 'Failed to resolve profile');
      console.error('Error resolving profile:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box sx={{ mb: 3 }}>
      <Typography variant="h6" gutterBottom>
        Follow Federated Profile
      </Typography>
      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}
      <form onSubmit={handleResolve}>
        <Box sx={{ display: 'flex', gap: 1 }}>
          <TextField
            fullWidth
            placeholder="Enter handle (e.g. user.bsky.social)"
            value={handle}
            onChange={(e) => setHandle(e.target.value)}
            disabled={loading}
            size="small"
          />
          <Button
            type="submit"
            variant="contained"
            disabled={!handle.trim() || loading}
            startIcon={loading ? <CircularProgress size={20} /> : <SearchIcon />}
          >
            Resolve
          </Button>
        </Box>
      </form>
      <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
        Enter a handle from any AT Protocol compatible service to follow and view their posts
      </Typography>
    </Box>
  );
}

export default FederatedContent;
