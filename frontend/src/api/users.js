import axios from 'axios';

// Получаем базовый URL API из переменных окружения или используем дефолтное значение
const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api';

// Создаем экземпляр axios с базовым URL и настройками
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Добавляем interceptor для добавления токена авторизации к каждому запросу
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token'); // Получаем токен из localStorage
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

/**
 * Получает список пользователей с их лимитами отпуска на указанный год.
 * Требует прав администратора.
 * @param {number} year - Год, за который нужно получить лимиты.
 * @returns {Promise<Array<object>>} - Промис, который разрешается массивом пользователей с лимитами.
 */
export const getUsersWithLimits = async (year) => {
  try {
    // Добавляем параметр year к запросу
    const response = await apiClient.get(`/admin/users`, { params: { year } });
    return response.data; // Возвращаем данные из ответа (массив UserWithLimitDTO)
  } catch (error) {
    console.error('Ошибка при получении пользователей с лимитами:', error.response || error.message);
    // Пробрасываем ошибку дальше, чтобы ее можно было обработать в компоненте
    throw new Error(error.response?.data?.error || 'Не удалось получить список пользователей');
  }
};

// Можно добавить другие функции для работы с пользователями (CRUD и т.д.)
// export const createUser = async (userData) => { ... };
// export const updateUser = async (userId, userData) => { ... };
// export const deleteUser = async (userId) => { ... };
