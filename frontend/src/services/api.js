import axios from 'axios';


const API_BASE_URL = 'http://localhost:8080/api'; 

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});


const TOKEN_KEY = 'authToken';

export const setAuthToken = (token) => {
  if (token) {
    localStorage.setItem(TOKEN_KEY, token);
    apiClient.defaults.headers.common['Authorization'] = `Bearer ${token}`;
  } else {
    removeAuthToken();
  }
};

export const removeAuthToken = () => {
  localStorage.removeItem(TOKEN_KEY);
  delete apiClient.defaults.headers.common['Authorization'];
};

export const getAuthToken = () => {
  return localStorage.getItem(TOKEN_KEY);
};

const initialToken = getAuthToken();
if (initialToken) {
  apiClient.defaults.headers.common['Authorization'] = `Bearer ${initialToken}`;
}

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
    throw error;
  }
};

export const logoutUser = () => {
  removeAuthToken();
};


export const createVacation = (vacationData) => {
  return apiClient.post('/vacations', vacationData);
};

export const getMyVacations = () => {
  return apiClient.get('/vacations/my');
};

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


export default apiClient;
