import axios from 'axios';

// Используем относительный URL для API, чтобы Nginx мог его проксировать
// Если REACT_APP_API_URL установлена, она будет использована.
const API_BASE_URL = process.env.REACT_APP_API_URL || '/api';

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

/**
 * Обновляет профиль пользователя.
 * @param {number} userId - ID пользователя для обновления.
 * @param {object} updateData - Объект с полями для обновления (например, { full_name: "Новое Имя", password: "новый_пароль", position_id: 3 }).
 *                               Поля, которые не нужно обновлять, не включаются в объект.
 * @returns {Promise<object>} - Промис, который разрешается объектом с сообщением об успехе.
 */
export const updateUserProfile = async (userId, updateData) => {
  try {
    // Удаляем пустые или null значения из updateData, чтобы не отправлять их
    const filteredData = Object.entries(updateData).reduce((acc, [key, value]) => {
      if (value !== null && value !== undefined && value !== '') {
        acc[key] = value;
      }
      return acc;
    }, {});

    // Если после фильтрации не осталось данных, не отправляем запрос
    if (Object.keys(filteredData).length === 0) {
      console.log("Нет данных для обновления профиля.");
      return { message: "Нет данных для обновления." }; // Или можно выбросить ошибку
    }


    const response = await apiClient.put(`/users/${userId}`, filteredData);
    return response.data; // Возвращаем ответ сервера (обычно сообщение об успехе)
  } catch (error) {
    console.error(`Ошибка при обновлении профиля пользователя ${userId}:`, error.response || error.message);
    throw new Error(error.response?.data?.error || 'Не удалось обновить профиль пользователя');
  }
};

/**
 * Получает данные профиля текущего аутентифицированного пользователя.
 * Включает иерархию организационных юнитов (Department, SubDepartment, Sector).
 * @returns {Promise<object>} - Промис, который разрешается объектом UserProfileDTO.
 */
export const getMyProfile = async () => {
  try {
    const response = await apiClient.get('/profile'); // Вызываем новый эндпоинт
    return response.data; // Возвращаем UserProfileDTO
  } catch (error) {
    console.error('Ошибка при получении профиля пользователя:', error.response || error.message);
    throw new Error(error.response?.data?.error || 'Не удалось получить профиль пользователя');
  }
};

/**
 * Получает список всех пользователей (для админ-панели).
 * Требует прав администратора.
 * @returns {Promise<Array<object>>} - Промис, который разрешается массивом UserProfileDTO.
 */
export const getAllUsersAdmin = async () => {
  try {
    const response = await apiClient.get('/admin/users'); // Вызываем новый эндпоинт GET /api/admin/users
    return response.data;
  } catch (error) {
    console.error('Ошибка при получении списка всех пользователей (админ):', error.response || error.message);
    throw new Error(error.response?.data?.error || 'Не удалось получить список пользователей');
  }
};

/**
 * Обновляет данные пользователя от имени администратора.
 * Требует прав администратора.
 * @param {number} userId - ID пользователя для обновления.
 * @param {object} updateData - Объект с полями для обновления (position_id, organizational_unit_id, is_admin, is_manager).
 * @returns {Promise<object>} - Промис, который разрешается объектом с сообщением об успехе.
 */
export const updateUserAdmin = async (userId, updateData) => {
  try {
    // Удаляем null/undefined значения, но оставляем false для is_admin/is_manager
    const filteredData = Object.entries(updateData).reduce((acc, [key, value]) => {
      if (value !== null && value !== undefined) {
        acc[key] = value;
      }
      return acc;
    }, {});

    if (Object.keys(filteredData).length === 0) {
      console.log("Нет данных для обновления пользователя (админ).");
      return { message: "Нет данных для обновления." };
    }

    const response = await apiClient.put(`/admin/users/${userId}`, filteredData); // Вызываем новый эндпоинт PUT /api/admin/users/{id}
    return response.data;
  } catch (error) {
    console.error(`Ошибка при обновлении пользователя ${userId} (админ):`, error.response || error.message);
    throw new Error(error.response?.data?.error || 'Не удалось обновить данные пользователя');
  }
};


// Можно добавить другие функции для работы с пользователями (CRUD и т.д.)
// export const createUser = async (userData) => { ... };
// export const getUserById = async (userId) => { ... }; // Может понадобиться для страницы профиля
// export const deleteUser = async (userId) => { ... };
