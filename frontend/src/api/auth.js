import axios from 'axios';

// Используем относительный URL для API, чтобы Nginx мог его проксировать
// Если REACT_APP_API_URL установлена (например, для локальной разработки с другим портом бэкенда), она будет использована.
const API_URL = process.env.REACT_APP_API_URL || '/api';

// Создание экземпляра axios с базовым URL
const authApi = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json'
  }
});

// Перехватчик для добавления токена к запросам
authApi.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`;
    }
    return config;
  },
  error => {
    return Promise.reject(error);
  }
);

// Перехватчик для обработки ошибок ответа
authApi.interceptors.response.use(
  response => response,
  error => {
    // Обработка 401 Unauthorized (например, истекший токен)
    if (error.response && error.response.status === 401) {
      // Удаляем старый токен и данные пользователя
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      // Перенаправляем на страницу логина, если мы не на ней
      if (window.location.pathname !== '/login') {
         window.location.href = '/login';
      }
    }
    // Возвращаем ошибку для дальнейшей обработки
    return Promise.reject(error);
  }
);

/**
 * Преобразует объект пользователя от бэкенда (snake_case) в формат фронтенда (camelCase).
 * @param {Object} backendUser - Объект пользователя с бэкенда.
 * @returns {Object|null} - Преобразованный объект пользователя или null.
 */
const transformUserKeys = (backendUser) => {
  if (!backendUser) return null;
  
  // Создаем новый объект, чтобы не мутировать исходный
  const frontendUser = { ...backendUser }; 

  // Преобразуем is_admin в isAdmin
  if (frontendUser.hasOwnProperty('is_admin')) {
    frontendUser.isAdmin = frontendUser.is_admin;
    delete frontendUser.is_admin; // Удаляем старый ключ
  }

  // Преобразуем is_manager в isManager
  if (frontendUser.hasOwnProperty('is_manager')) {
    frontendUser.isManager = frontendUser.is_manager;
    delete frontendUser.is_manager; // Удаляем старый ключ
  }
  
  // Добавьте здесь преобразование других ключей при необходимости
  // Например: frontendUser.fullName = frontendUser.full_name; delete frontendUser.full_name;

  return frontendUser;
};


/**
 * Авторизация пользователя
 * @param {Object} credentials - Объект с данными для входа (например, { email, password })
 * @returns {Promise<Object>} - Объект с данными пользователя и токеном
 */
export const login = async (credentials) => {
  try {
    // Отправляем объект с полями username и password, как ожидает бэкенд
    const payload = {
      username: credentials.email, // Используем email как username
      password: credentials.password
    };
    const response = await authApi.post('/auth/login', payload); // Отправляем измененный payload
    // Сохраняем токен и данные пользователя в localStorage при успешном входе
    const token = response.data?.token;
    const backendUser = response.data?.user;

    if (token) {
        localStorage.setItem('token', token);
    }
    
    // Преобразуем пользователя перед сохранением и возвратом
    const frontendUser = transformUserKeys(backendUser); 

    if (frontendUser) {
        localStorage.setItem('user', JSON.stringify(frontendUser));
    }
    
    // Возвращаем объект с токеном и преобразованным пользователем
    return { token, user: frontendUser }; 
  } catch (error) {
    // Обработка ошибок от сервера
    if (error.response && error.response.data && error.response.data.error) {
      throw new Error(error.response.data.error);
    } else if (error.request) {
      // Ошибка сети или сервер недоступен
      throw new Error('Не удалось подключиться к серверу. Проверьте ваше соединение.');
    } else {
      // Другие ошибки
      throw new Error('Произошла неизвестная ошибка при входе.');
    }
  }
};

/**
 * Выход из системы
 */
export const logout = () => {
  localStorage.removeItem('token');
  localStorage.removeItem('user');
  // Перенаправляем на страницу логина
  window.location.href = '/login';
};

/**
 * Получение информации о текущем пользователе (если есть токен)
 * @returns {Promise<Object|null>} - Объект с данными пользователя или null
 */
export const getCurrentUser = async () => {
  // Проверяем наличие токена перед запросом
  if (!isAuthenticated()) {
    return null;
  }
  
  try {
    const response = await authApi.get('/auth/me'); // Предполагаем, что есть такой эндпоинт
    const backendUser = response.data;

    // Преобразуем пользователя перед сохранением и возвратом
    const frontendUser = transformUserKeys(backendUser); 

    // Обновляем данные пользователя в localStorage преобразованными данными
    if (frontendUser) {
       localStorage.setItem('user', JSON.stringify(frontendUser));
    }
    
    // Возвращаем преобразованного пользователя
    return frontendUser; 
  } catch (error) {
    // Не бросаем ошибку, если это 401 (токен невалиден), просто возвращаем null
    if (error.response && error.response.status !== 401) {
      console.error('Ошибка получения данных пользователя:', error);
      // Можно выбросить ошибку для других статусов, если нужно
      // throw new Error(error.response?.data?.error || 'Ошибка получения данных пользователя');
    }
    // Если токен невалиден или другая ошибка, считаем пользователя неавторизованным
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    return null;
  }
};

/**
 * Проверка наличия валидного токена в localStorage
 * @returns {boolean} - true, если токен есть
 */
export const isAuthenticated = () => {
  return !!localStorage.getItem('token');
  // В более сложном случае можно проверять срок действия токена
};

/**
 * Регистрация нового пользователя
 * @param {Object} userData - Данные пользователя (username, password, confirm_password, full_name, email, position_id)
 * @returns {Promise<Object>} - Объект с данными созданного пользователя
 */
export const register = async (userData) => {
  try {
    // Преобразуем ключи перед отправкой на бэкенд (если нужно)
    // const backendUserData = { ...userData };
    // if (backendUserData.hasOwnProperty('positionId')) {
    //   backendUserData.position_id = backendUserData.positionId;
    //   delete backendUserData.positionId;
    // }
    // ... другие преобразования ...

    const response = await authApi.post('/auth/register', userData); // Отправляем userData как есть
    const backendUser = response.data;

    // Преобразуем пользователя перед возвратом
    const frontendUser = transformUserKeys(backendUser);

    // Не сохраняем пользователя в localStorage при регистрации,
    // пользователь должен будет войти после регистрации.
    return frontendUser;
  } catch (error) {
    if (error.response && error.response.data && error.response.data.error) {
      throw new Error(error.response.data.error);
    } else if (error.request) {
      throw new Error('Не удалось подключиться к серверу. Проверьте ваше соединение.');
    } else {
      throw new Error('Произошла неизвестная ошибка при регистрации.');
    }
  }
};

/**
 * Получение списка должностей
 * @returns {Promise<Array>} - Массив групп должностей с вложенными должностями
 */
export const getPositions = async () => {
  try {
    const response = await authApi.get('/positions');
    // Данные с бэкенда уже должны быть в нужном формате (массив PositionGroup)
    // Если ключи в Position или PositionGroup в snake_case, их нужно будет преобразовать здесь.
    // Пример:
    // return response.data.map(group => ({
    //   ...group,
    //   sortOrder: group.sort_order, // Преобразование sort_order
    //   positions: group.positions.map(pos => ({ ...pos })) // Преобразование ключей в Position, если нужно
    // }));
    return response.data;
  } catch (error) {
    if (error.response && error.response.data && error.response.data.error) {
      throw new Error(error.response.data.error);
    } else if (error.request) {
      throw new Error('Не удалось подключиться к серверу. Проверьте ваше соединение.');
    } else {
      throw new Error('Произошла неизвестная ошибка при получении списка должностей.');
    }
  }
};


// Экспорт экземпляра axios для возможного использования в других API-клиентах
export default authApi;
