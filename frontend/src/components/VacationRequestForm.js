import React, { useState } from 'react';
import {
  Box,
  TextField,
  Button,
  CircularProgress,
  Alert,
  Grid,
  Typography,
  Fade,
} from '@mui/material';

import { createVacation } from '../services/api';

const formatDateForInput = (date) => {
    if (!date) return '';
    const d = new Date(date);
    const month = `${d.getMonth() + 1}`.padStart(2, '0');
    const day = `${d.getDate()}`.padStart(2, '0');
    const year = d.getFullYear();
    return [year, month, day].join('-');
};

const VacationRequestForm = ({ onSuccess, onCancel }) => {
  const today = formatDateForInput(new Date());
  const [startDate, setStartDate] = useState(today);
  const [endDate, setEndDate] = useState(today);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError(null);

    // Простая валидация
    if (!startDate || !endDate) {
      setError('Пожалуйста, выберите даты начала и окончания.');
      return;
    }
    if (new Date(endDate) < new Date(startDate)) {
      setError('Дата окончания не может быть раньше даты начала.');
      return;
    }

    setLoading(true);
    try {
      // Убираем user_id из payload, бэкенд возьмет его из токена
      const payload = {
        start_date: startDate,
        end_date: endDate,
      };
      await createVacation(payload);
      onSuccess();
    } catch (err) {
      console.error("Submit error:", err);
      setError(err.response?.data?.error || 'Не удалось создать заявку.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Fade in={true} timeout={500}>
        <Box component="form" onSubmit={handleSubmit} noValidate>
            {error && (
                <Alert severity="error" sx={{ mb: 2 }}>
                {error}
                </Alert>
            )}
            <Grid container spacing={2}>
                <Grid item xs={12} sm={6}>
                <TextField
                    label="Дата начала"
                    type="date"
                    value={startDate}
                    onChange={(e) => setStartDate(e.target.value)}
                    InputLabelProps={{
                    shrink: true,
                    }}
                    fullWidth
                    required
                    inputProps={{
                        min: today
                    }}
                />
                </Grid>
                <Grid item xs={12} sm={6}>
                <TextField
                    label="Дата окончания"
                    type="date"
                    value={endDate}
                    onChange={(e) => setEndDate(e.target.value)}
                    InputLabelProps={{
                    shrink: true,
                    }}
                    fullWidth
                    required
                    inputProps={{
                        min: startDate || today
                    }}
                />
                </Grid>
            </Grid>
            <Box sx={{ mt: 3, display: 'flex', justifyContent: 'flex-end', gap: 1 }}>
                <Button onClick={onCancel} color="secondary" disabled={loading}>
                    Отмена
                </Button>
                <Button
                    type="submit"
                    variant="contained"
                    color="primary"
                    disabled={loading}
                    startIcon={loading ? <CircularProgress size={20} color="inherit" /> : null}
                >
                    {loading ? 'Отправка...' : 'Создать заявку'}
                </Button>
            </Box>
        </Box>
    </Fade>
  );
};

export default VacationRequestForm;
