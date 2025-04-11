import React, { useState } from 'react';
import { loginUser } from '../services/api';
import {
  Box,
  TextField,
  Button,
  CircularProgress,
  Alert,
  Typography,
} from '@mui/material'; // Импортируем компоненты MUI

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
      // response.data содержит { token, user_id, is_admin, expiresAt }
      console.log('Login successful:', response.data);
      if (onLoginSuccess) {
        // Передаем информацию о пользователе и токене в родительский компонент
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
      // Токен уже удаляется внутри loginUser при ошибке
    } finally {
      setLoading(false);
    }
  };

  // Используем Box и компоненты MUI для формы
  return (
    <Box
      component="form"
      onSubmit={handleSubmit}
      noValidate
      sx={{ mt: 1, width: '100%' }} // Добавляем отступ сверху и занимаем всю ширину контейнера
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
      {/* TODO: Добавить Checkbox "Remember me" если нужно */}
      <Button
        type="submit"
        fullWidth
        variant="contained"
        sx={{ mt: 3, mb: 2 }} // Отступы для кнопки
        disabled={loading}
      >
        {loading ? <CircularProgress size={24} color="inherit" /> : 'Войти'}
      </Button>
      {/* TODO: Добавить ссылки на регистрацию или восстановление пароля */}
      {/* <Grid container>
            <Grid item xs>
              <Link href="#" variant="body2">
                Forgot password?
              </Link>
            </Grid>
            <Grid item>
              <Link href="#" variant="body2">
                {"Don't have an account? Sign Up"}
              </Link>
            </Grid>
          </Grid> */}
    </Box>
  );
}

export default Login;
