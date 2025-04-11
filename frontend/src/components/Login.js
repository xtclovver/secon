import React, { useState } from 'react';
import { loginUser } from '../services/api';
import {
  Box,
  TextField,
  Button,
  CircularProgress,
  Alert,
  Typography,
} from '@mui/material';

function Login({ onLoginSuccess }) {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError('');
    setLoading(true);

    try {
      const response = await loginUser({ email, password });
      console.log('Login successful:', response.data);
      if (onLoginSuccess) {
        onLoginSuccess({
          token: response.data.token,
          userId: response.data.user_id,
          isAdmin: response.data.is_admin,
        });
      }
    } catch (err) {
      console.error('Login failed:', err);
      const errorMessage = err.response?.data?.error || 'Login failed. Please check your credentials.';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box
      component="form"
      onSubmit={handleSubmit}
      noValidate
      sx={{ mt: 1, width: '100%' }}
    >
      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}
      <TextField
        margin="normal"
        required
        fullWidth
        id="email"
        label="Email Address"
        name="email"
        autoComplete="email"
        autoFocus
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        disabled={loading}
      />
      <TextField
        margin="normal"
        required
        fullWidth
        name="password"
        label="Password"
        type="password"
        id="password"
        autoComplete="current-password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        disabled={loading}
      />
      <Button
        type="submit"
        fullWidth
        variant="contained"
        sx={{ mt: 3, mb: 2 }}
        disabled={loading}
      >
        {loading ? <CircularProgress size={24} color="inherit" /> : 'Войти'}
      </Button>
    </Box>
  );
}

export default Login;
