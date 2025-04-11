import React, { useState, useEffect, useCallback } from 'react';
import {
  CssBaseline,
  AppBar,
  Toolbar,
  Typography,
  Container,
  Box,
  CircularProgress,
  Alert,
  ThemeProvider,
  createTheme,
  Paper,
  Fab,
  Button,
  Switch,
  FormControlLabel,
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import LogoutIcon from '@mui/icons-material/Logout';
import VacationRequestForm from './components/VacationRequestForm';
import VacationList from './components/VacationList';
import Login from './components/Login';
import {
  getAuthToken,
  removeAuthToken,
  logoutUser,
  getMyVacations, // API для получения своих заявок
  getAllVacations, // API для получения всех заявок (админ)
} from './services/api';
import { jwtDecode } from 'jwt-decode'; // Библиотека для декодирования JWT

// Тема MUI (оставляем как есть или кастомизируем дальше)
const theme = createTheme({
  palette: {
    primary: { main: '#1976d2' },
    secondary: { main: '#dc004e' },
    background: { default: '#f4f6f8', paper: '#ffffff' },
    text: { primary: '#333', secondary: '#666' },
  },
  typography: {
    fontFamily: '"Roboto", "Helvetica", "Arial", sans-serif',
    h4: { fontWeight: 600, marginBottom: '1.5rem', color: '#1a237e' },
    h6: { fontWeight: 500, color: '#3f51b5' },
    button: { textTransform: 'none', fontWeight: 'bold' }, 
  },
  shape: {
    borderRadius: 8, // Скругляем углы
  },
  components: {
    MuiPaper: {
      styleOverrides: {
        root: {
          transition: 'box-shadow 0.3s ease-in-out',
          boxShadow: '0 2px 4px rgba(0,0,0,0.05)', // Мягкая тень по умолчанию
          '&:hover': {
             boxShadow: '0 5px 15px rgba(0,0,0,0.1)', // Более заметная тень при наведении
          }
        }
      },
      defaultProps: {
        elevation: 0, // Убираем стандартную тень, используем свою
      }
    },
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: 6, 
          transition: 'background-color 0.2s ease-in-out, transform 0.1s ease-in-out',
          '&:hover': {
            transform: 'translateY(-1px)',
            boxShadow: '0 2px 5px rgba(0,0,0,0.1)',
          }
        },
        containedPrimary: {
            '&:hover': {
                backgroundColor: '#1565c0',
            }
        }
      }
    },
    MuiFab: {
      styleOverrides: {
        root: {
          transition: 'background-color 0.2s ease-in-out, transform 0.2s ease-in-out, box-shadow 0.2s ease-in-out',
          boxShadow: '0 4px 10px rgba(0,0,0,0.2)',
          '&:hover': {
            transform: 'scale(1.08)', 
            boxShadow: '0 6px 15px rgba(0,0,0,0.3)',
          }
        }
      }
    },
    MuiAppBar: {
        styleOverrides: {
            root: {
                backgroundColor: '#ffffff', // Белый AppBar
                color: '#333', // Темный текст
                boxShadow: '0 2px 4px rgba(0,0,0,0.05)',
            }
        }
    },
    MuiSwitch: { 
        styleOverrides: {
            root: {
            },
            switchBase: {
                '&.Mui-checked': {
                    color: '#fff',
                    '& + .MuiSwitch-track': {
                        backgroundColor: '#dc004e',
                        opacity: 1,
                        border: 0,
                    },
                },
            },
            thumb: {
            },
            track: {
                borderRadius: 34 / 2,
                opacity: 1,
                backgroundColor: 'rgba(0,0,0,.25)',
                boxSizing: 'border-box',
            },
        }
    },
    MuiAlert: {
        styleOverrides: {
            root: {
                borderRadius: 6, 
            }
        }
    }
  },
});

// Функция для извлечения данных пользователя из токена
const getUserDataFromToken = (token) => {
  if (!token) return null;
  try {
    const decoded = jwtDecode(token); // Декодируем токен
    // Предполагаем, что токен содержит user_id и is_admin в payload
    return {
      token: token,
      userId: decoded.user_id, // Имя поля должно совпадать с тем, что в бэкенде (Claims)
      isAdmin: decoded.is_admin, // Имя поля должно совпадать с тем, что в бэкенде (Claims)
      // Добавляем время истечения для возможной проверки
      expiresAt: decoded.exp ? new Date(decoded.exp * 1000) : null,
    };
  } catch (error) {
    console.error("Error decoding token:", error);
    removeAuthToken(); // Удаляем невалидный токен
    return null;
  }
};


function App() {
  const [currentUser, setCurrentUser] = useState(null); // Хранит { token, userId, isAdmin, expiresAt }
  const [requests, setRequests] = useState([]);
  const [loading, setLoading] = useState(true); // Общий лоадер для инициализации и загрузки данных
  const [error, setError] = useState(null);
  const [showForm, setShowForm] = useState(false);
  const [viewAsAdmin, setViewAsAdmin] = useState(false); // Для админа: переключение вида

  // Проверка токена при загрузке приложения
  useEffect(() => {
    const token = getAuthToken();
    const userData = getUserDataFromToken(token);
    if (userData && userData.expiresAt && userData.expiresAt > new Date()) {
      setCurrentUser(userData);
      // Если пользователь админ, по умолчанию смотрим как админ
      if (userData.isAdmin) {
        setViewAsAdmin(true);
      }
    } else {
      // Если токена нет или он истек
      removeAuthToken();
      setCurrentUser(null);
    }
    setLoading(false); // Завершаем начальную загрузку состояния аутентификации
  }, []);

  // Функция для загрузки заявок в зависимости от роли и вида
  const fetchRequests = useCallback(async () => {
    if (!currentUser) return; // Не загружаем, если не авторизован

    setLoading(true);
    setError(null);
    try {
      let response;
      if (currentUser.isAdmin && viewAsAdmin) {
        // Админ смотрит все заявки
        response = await getAllVacations();
      } else {
        // Обычный пользователь или админ смотрит свои заявки
        response = await getMyVacations();
      }
      setRequests(response.data || []);
    } catch (err) {
      console.error("Fetch error:", err);
      setError(err.response?.data?.error || 'Не удалось загрузить заявки.');
      if (err.response?.status === 401) {
          handleLogout(); // Разлогиниваем при ошибке авторизации
      }
      setRequests([]);
    } finally {
      setLoading(false);
    }
  }, [currentUser, viewAsAdmin]); // Зависимости: пользователь и режим просмотра

  // Загружаем заявки при изменении состояния аутентификации или режима просмотра
  useEffect(() => {
    if (currentUser) {
      fetchRequests();
    } else {
      // Если пользователь разлогинился, очищаем заявки
      setRequests([]);
    }
  }, [currentUser, fetchRequests]);

  // Обработчик успешного входа
  const handleLoginSuccess = (userData) => {
    const fullUserData = getUserDataFromToken(userData.token); // Декодируем для получения всех данных
    if (fullUserData) {
        setCurrentUser(fullUserData);
        if (fullUserData.isAdmin) {
            setViewAsAdmin(true); // Админ по умолчанию видит все
        }
    } else {
        // Обработка ошибки декодирования, если необходимо
        setError("Ошибка декодирования");
        removeAuthToken();
    }
  };

  // Обработчик выхода
  const handleLogout = () => {
    logoutUser();
    setCurrentUser(null);
    setViewAsAdmin(false);
    setShowForm(false);
    console.log("Пользователь вышел");
  };

  const handleRequestCreated = () => {
    fetchRequests(); 
    setShowForm(false);
  };

  const handleViewToggle = (event) => {
    setViewAsAdmin(event.target.checked);
  };


  if (loading && !currentUser) {
    // Показываем лоадер только при самой первой загрузке, пока проверяется токен
    return (
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <Box display="flex" justifyContent="center" alignItems="center" minHeight="100vh">
          <CircularProgress />
        </Box>
      </ThemeProvider>
    );
  }

  if (!currentUser) {
    // Если пользователь не аутентифицирован, показываем форму входа
    return (
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <Container component="main" maxWidth="xs" sx={{ mt: 8 }}>
          <Paper elevation={3} sx={{ p: 4, display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
            <Typography component="h1" variant="h5">
              Vacation Scheduler Login
            </Typography>
            <Login onLoginSuccess={handleLoginSuccess} />
          </Paper>
        </Container>
      </ThemeProvider>
    );
  }

  // Пользователь аутентифицирован
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <AppBar position="static" elevation={1} sx={{ marginBottom: '2rem' }}>
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            График Отпусков
          </Typography>
          {currentUser.isAdmin && (
            <FormControlLabel
              control={
                <Switch
                  checked={viewAsAdmin}
                  onChange={handleViewToggle}
                  color="secondary"
                />
              }
              label="Admin View"
              sx={{ color: 'white', mr: 2 }}
            />
          )}
          <Button color="inherit" onClick={handleLogout} startIcon={<LogoutIcon />}>
            Выход
          </Button>
        </Toolbar>
      </AppBar>

      <Container maxWidth="lg">
        <Typography variant="h4" gutterBottom align="center">
          {currentUser.isAdmin && viewAsAdmin ? 'Все Заявки (Admin)' : 'Мои Заявки на Отпуск'}
        </Typography>

        {error && (
          <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
            {error}
          </Alert>
        )}

        {loading ? (
          <Box display="flex" justifyContent="center" my={4}>
            <CircularProgress />
          </Box>
        ) : (
          <Paper elevation={2} sx={{ p: 3, mb: 3 }}>
            <VacationList
                requests={requests}
                isAdminView={currentUser.isAdmin && viewAsAdmin}
                currentUserId={currentUser.userId}
                onUpdateRequest={fetchRequests}
                onDeleteRequest={fetchRequests}
            />
          </Paper>
        )}

        {!viewAsAdmin && showForm && (
          <Paper elevation={3} sx={{ p: 3, mt: 3, mb: 8 }}>
            <Typography variant="h6" gutterBottom>Новая заявка</Typography>
            <VacationRequestForm
              onSuccess={handleRequestCreated}
              onCancel={() => setShowForm(false)}
            />
          </Paper>
        )}

        {/* Кнопка FAB для добавления, только если не в режиме админа */}
        {!viewAsAdmin && (
          <Fab
            color="primary"
            aria-label="add"
            sx={{ position: 'fixed', bottom: 32, right: 32 }}
            onClick={() => setShowForm(!showForm)}
          >
            <AddIcon />
          </Fab>
        )}
      </Container>
    </ThemeProvider>
  );
}

export default App;
