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
