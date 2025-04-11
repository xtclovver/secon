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
  Button, // Для кнопки выхода
  Switch, // Для переключения вида админ/юзер (пример)
  FormControlLabel, // Для Switch
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import LogoutIcon from '@mui/icons-material/Logout'; // Иконка для выхода
import VacationRequestForm from './components/VacationRequestForm';
import VacationList from './components/VacationList';
import Login from './components/Login'; // Импортируем компонент входа
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
    background: { default: '#f4f6f8', paper: '#ffffff' }, // Светлый фон
    text: { primary: '#333', secondary: '#666' },
  },
  typography: {
    fontFamily: '"Roboto", "Helvetica", "Arial", sans-serif', // Стандартный шрифт MUI
    h4: { fontWeight: 600, marginBottom: '1.5rem', color: '#1a237e' }, // Темно-синий заголовок
    h6: { fontWeight: 500, color: '#3f51b5' }, // Синий подзаголовок
    button: { textTransform: 'none', fontWeight: 'bold' }, // Кнопки без капса, жирные
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
          borderRadius: 6, // Скругление кнопок
          transition: 'background-color 0.2s ease-in-out, transform 0.1s ease-in-out',
          '&:hover': {
            transform: 'translateY(-1px)', // Легкий подъем при наведении
            boxShadow: '0 2px 5px rgba(0,0,0,0.1)',
          }
        },
        containedPrimary: { // Стиль для основной кнопки
            '&:hover': {
                backgroundColor: '#1565c0', // Чуть темнее при наведении
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
            transform: 'scale(1.08)', // Увеличение FAB при наведении
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
                boxShadow: '0 2px 4px rgba(0,0,0,0.05)', // Легкая тень снизу
            }
        }
    },
    MuiSwitch: { // Стили для переключателя Admin View
        styleOverrides: {
            root: {
                 // Увеличим немного размер
                 // width: 62,
                 // height: 34,
            },
            switchBase: {
                // Стили для кружка переключателя
                // padding: 1,
                '&.Mui-checked': {
                    // transform: 'translateX(28px)', // Сдвиг для увеличенного размера
                    color: '#fff',
                    '& + .MuiSwitch-track': {
                        backgroundColor: '#dc004e', // Используем прямое значение цвета
                        opacity: 1,
                        border: 0,
                    },
                },
            },
            thumb: { // Сам кружок
                // boxShadow: '0 2px 4px 0 rgb(0 35 11 / 20%)',
                // width: 32,
                // height: 32,
            },
            track: { // Полоса переключателя
                borderRadius: 34 / 2,
                opacity: 1,
                backgroundColor: 'rgba(0,0,0,.25)',
                boxSizing: 'border-box',
            },
        }
    },
    MuiAlert: { // Стили для сообщений об ошибках
        styleOverrides: {
            root: {
                borderRadius: 6, // Скругление
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
        console.log("Fetching all vacations (admin view)");
      } else {
        // Обычный пользователь или админ смотрит свои заявки
        response = await getMyVacations();
         console.log("Fetching my vacations");
      }
      setRequests(response.data || []);
    } catch (err) {
      console.error("Fetch error:", err);
      setError(err.response?.data?.error || 'Не удалось загрузить заявки.');
      // Проверяем на ошибку токена (например, 401 Unauthorized)
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
  }, [currentUser, fetchRequests]); // fetchRequests уже включает viewAsAdmin в зависимостях

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
        setError("Failed to process login data.");
        removeAuthToken();
    }
  };

  // Обработчик выхода
  const handleLogout = () => {
    logoutUser(); // Удаляет токен из localStorage и apiClient
    setCurrentUser(null);
    setViewAsAdmin(false); // Сбрасываем вид
    setShowForm(false); // Скрываем форму, если была открыта
    console.log("User logged out");
  };

  // Обработчик успешного создания заявки
  const handleRequestCreated = () => {
    fetchRequests(); // Обновляем список
    setShowForm(false); // Скрываем форму
  };

  // Обработчик переключения вида для админа
  const handleViewToggle = (event) => {
    setViewAsAdmin(event.target.checked);
  };

  // --- Рендеринг ---

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
            {/* TODO: Добавить возможность переключения на регистрацию */}
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
          {/* Показываем переключатель вида для админа */}
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
            {/* Передаем isAdmin в VacationList для условного рендеринга кнопок управления */}
            <VacationList
                requests={requests}
                isAdminView={currentUser.isAdmin && viewAsAdmin}
                currentUserId={currentUser.userId} // Передаем ID для возможной логики (например, выделение своих заявок)
                onUpdateRequest={fetchRequests} // Передаем функцию для обновления списка после действий
                onDeleteRequest={fetchRequests} // Передаем функцию для обновления списка после действий
            />
          </Paper>
        )}

        {/* Форма показывается по кнопке, только если не в режиме админа */}
        {!viewAsAdmin && showForm && (
          <Paper elevation={3} sx={{ p: 3, mt: 3, mb: 8 }}>
            <Typography variant="h6" gutterBottom>Новая заявка</Typography>
            <VacationRequestForm
              // userId больше не нужен, бэкенд берет из токена
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
