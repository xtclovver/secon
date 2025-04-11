import axios from 'axios';

// Настройте базовый URL вашего API
// Если фронтенд и бэкенд работают на разных портах во время разработки (например, React на 3000, Go на 8080),
// вам нужно будет указать полный URL бэкенда. CORS уже настроен на бэкенде.
const API_BASE_URL = 'http://localhost:8080/api'; // Указываем полный URL бэкенда

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// --- Token Management ---

const TOKEN_KEY = 'authToken'; // Ключ для хранения токена в localStorage

// Функция для сохранения токена
export const setAuthToken = (token) => {
  if (token) {
    localStorage.setItem(TOKEN_KEY, token);
    // Устанавливаем токен для будущих запросов через этот инстанс apiClient
    apiClient.defaults.headers.common['Authorization'] = `Bearer ${token}`;
  } else {
    removeAuthToken(); // Удаляем токен, если передан null или undefined
  }
};

// Функция для удаления токена
export const removeAuthToken = () => {
  localStorage.removeItem(TOKEN_KEY);
  // Удаляем заголовок из настроек apiClient по умолчанию
  delete apiClient.defaults.headers.common['Authorization'];
};

// Функция для получения токена
export const getAuthToken = () => {
  return localStorage.getItem(TOKEN_KEY);
};

// --- Interceptor для автоматической установки токена при загрузке ---
// При инициализации модуля проверяем, есть ли токен в localStorage
const initialToken = getAuthToken();
if (initialToken) {
  apiClient.defaults.headers.common['Authorization'] = `Bearer ${initialToken}`;
}

// --- API Functions ---

// Функция для регистрации пользователя
export const registerUser = (userData) => {
  return apiClient.post('/users/register', userData);
};

// Функция для входа пользователя
export const loginUser = async (credentials) => {
  try {
    const response = await apiClient.post('/users/login', credentials);
    if (response.data && response.data.token) {
      setAuthToken(response.data.token); // Сохраняем токен при успешном входе
    }
    return response; // Возвращаем полный ответ для дальнейшей обработки
  } catch (error) {
    removeAuthToken(); // Удаляем токен в случае ошибки входа
    throw error; // Пробрасываем ошибку дальше
  }
};

// Функция для выхода (просто удаляет токен)
export const logoutUser = () => {
  removeAuthToken();
  // Здесь можно добавить вызов API для инвалидации токена на сервере, если это реализовано
};

// --- Существующие или будущие функции API для отпусков ---
// (Пример: адаптируем под использование apiClient с автоматическим токеном)

export const createVacation = (vacationData) => {
  // UserID больше не передается, бэкенд берет его из токена
  return apiClient.post('/vacations', vacationData);
};

export const getMyVacations = () => {
  // Используем новый маршрут /my
  return apiClient.get('/vacations/my');
};

// Функции для администратора (примеры)
export const getAllVacations = () => {
  return apiClient.get('/vacations'); // Админский маршрут
};

export const updateVacationStatus = (id, status) => {
  return apiClient.put(`/vacations/${id}`, { status }); // Админский маршрут
};

export const deleteVacation = (id) => {
  return apiClient.delete(`/vacations/${id}`); // Админский маршрут
};

export const getUserVacations = (userId) => {
    return apiClient.get(`/vacations/user/${userId}`); // Админский маршрут
};

export const checkOverlaps = () => {
    return apiClient.get('/vacations/overlaps'); // Админский маршрут
}


// Экспортируем сам клиент axios, если он нужен где-то еще напрямую,
// но предпочтительнее использовать экспортированные функции.
export default apiClient;
