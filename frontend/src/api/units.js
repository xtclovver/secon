import authApi from './auth'; // Используем настроенный экземпляр axios с перехватчиком

// API_URL теперь берется из настроек authApi (process.env.REACT_APP_API_URL || '/api')

// Получить дерево организационной структуры
export const getOrganizationalUnitTree = async () => {
  try {
    // Используем authApi, но обращаемся к ПУБЛИЧНОМУ эндпоинту /api/units/tree
    // authApi все еще полезен для базового URL и возможных других настроек, но он НЕ будет добавлять заголовок Authorization сюда, так как это GET запрос без явной аутентификации.
    const response = await authApi.get(`/units/tree`); // Убрали /admin
    return response.data;
  } catch (error) {
    console.error("Ошибка при получении дерева организационной структуры:", error.response?.data?.error || error.message);
    throw error.response?.data?.error || new Error("Не удалось загрузить структуру организации");
    }
  };

// Получить дочерние элементы (юниты и пользователи) для родительского юнита
export const getUnitChildren = async (parentId = null) => {
  try {
    // Используем authApi. GET-запрос, аутентификация не требуется явно, но базовый URL берется
    const params = {};
    if (parentId !== null) {
      params.parentId = parentId;
    }
    // Эндпоинт изменен на /units/children
    const response = await authApi.get(`/units/children`, { params });
    return response.data; // Ожидаем массив UnitListItemDTO
  } catch (error) {
    console.error(`Ошибка при получении дочерних элементов для parentId ${parentId}:`, error.response?.data?.error || error.message);
    throw error.response?.data?.error || new Error("Не удалось загрузить дочерние элементы");
  }
};

// Создать новый организационный юнит
export const createOrganizationalUnit = async (unitData) => {
  try {
    // Используем authApi и убираем ручное добавление заголовков
    const response = await authApi.post(`/admin/units`, unitData); 
    return response.data;
  } catch (error) {
    console.error("Ошибка при создании организационного юнита:", error.response?.data?.error || error.message);
    throw error.response?.data?.error || new Error("Не удалось создать юнит");
  }
};

// Обновить организационный юнит
export const updateOrganizationalUnit = async (unitId, unitData) => {
  try {
    // Используем authApi и убираем ручное добавление заголовков
    const response = await authApi.put(`/admin/units/${unitId}`, unitData); 
    return response.data;
  } catch (error) {
    console.error(`Ошибка при обновлении организационного юнита ${unitId}:`, error.response?.data?.error || error.message);
    throw error.response?.data?.error || new Error("Не удалось обновить юнит");
  }
};

// Удалить организационный юнит
export const deleteOrganizationalUnit = async (unitId) => {
  try {
    // Используем authApi и убираем ручное добавление заголовков
    const response = await authApi.delete(`/admin/units/${unitId}`); 
    return response.data; // Обычно возвращает сообщение об успехе
  } catch (error) {
    console.error(`Ошибка при удалении организационного юнита ${unitId}:`, error.response?.data?.error || error.message);
    throw error.response?.data?.error || new Error("Не удалось удалить юнит");
  }
};

// Получить организационный юнит по ID (может понадобиться для редактирования)
export const getOrganizationalUnitById = async (unitId) => {
    try {
      // Используем authApi и убираем ручное добавление заголовков
      const response = await authApi.get(`/admin/units/${unitId}`); 
      return response.data;
    } catch (error) {
      console.error(`Ошибка при получении организационного юнита ${unitId}:`, error.response?.data?.error || error.message);
      throw error.response?.data?.error || new Error("Не удалось загрузить данные юнита");
    }
  };

// Получить пользователей юнита с их лимитами отпуска на год
export const getUnitUsersWithLimits = async (unitId, year) => {
  try {
    // Используем authApi для GET запроса с параметрами
    const response = await authApi.get(`/admin/units/${unitId}/users-with-limits`, {
      params: { year },
    });
    // Ожидаем массив UserWithLimitAdminDTO
    return response.data;
  } catch (error) {
    console.error(`Ошибка при получении пользователей с лимитами для юнита ${unitId}, год ${year}:`, error.response?.data?.error || error.message);
    throw error.response?.data?.error || new Error("Не удалось загрузить пользователей юнита с лимитами");
  }
};

// Обновить лимит отпуска для пользователя на год
export const updateUserVacationLimit = async (userId, year, totalDays) => {
  try {
    // Используем authApi для PUT запроса
    const response = await authApi.put(`/admin/users/${userId}/vacation-limit`, {
      year: year,
      total_days: totalDays,
    });
    return response.data; // Обычно возвращает сообщение об успехе
  } catch (error) {
    console.error(`Ошибка при обновлении лимита отпуска для пользователя ${userId}, год ${year}:`, error.response?.data?.error || error.message);
    throw error.response?.data?.error || new Error("Не удалось обновить лимит отпуска пользователя");
  }
};
