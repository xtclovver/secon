import authApi from './auth'; // Импортируем настроенный экземпляр axios из auth.js

/**
 * Получение лимита отпуска на указанный год
 * @param {number} year - Год
 * @returns {Promise<Object>} - Данные о лимите отпуска { totalDays, usedDays }
 * @throws {Error} - В случае ошибки запроса
 */
export const getVacationLimit = async (year) => {
  try {
    const response = await authApi.get(`/vacations/limits/${year}`);
    return response.data;
  } catch (error) {
    console.error("API Error in getVacationLimit:", error);
    if (error.response && error.response.data && error.response.data.error) {
      throw new Error(error.response.data.error);
    }
    throw new Error('Не удалось получить лимит отпуска.'); // Исправлено сообщение об ошибке
  } 
};

/**
 * Создание заявки на отпуск
 * @param {Object} request - Данные заявки { year, periods: [{ startDate, endDate, daysCount }], comment? }
 * @returns {Promise<Object>} - Созданная заявка с ID
 * @throws {Error} - В случае ошибки запроса
 */
export const createVacationRequest = async (request) => {
  try {
    const response = await authApi.post('/vacations/requests', request);
    return response.data;
  } catch (error) {
    console.error("API Error in createVacationRequest:", error);
    if (error.response && error.response.data && error.response.data.error) {
      throw new Error(error.response.data.error);
    }
    throw new Error('Не удалось создать заявку на отпуск.');
  }
};

/**
 * Обновление заявки на отпуск (например, сохранение черновика)
 * @param {number} id - ID заявки
 * @param {Object} request - Обновленные данные заявки
 * @returns {Promise<Object>} - Обновленная заявка
 * @throws {Error} - В случае ошибки запроса
 */
export const updateVacationRequest = async (id, request) => {
  try {
    const response = await authApi.put(`/vacations/requests/${id}`, request);
    return response.data;
  } catch (error) {
    console.error("API Error in updateVacationRequest:", error);
     if (error.response && error.response.data && error.response.data.error) {
      throw new Error(error.response.data.error);
    }
    throw new Error('Не удалось обновить заявку на отпуск.');
  }
};

/**
 * Отправка заявки на рассмотрение руководителю
 * @param {number} id - ID заявки
 * @returns {Promise<Object>} - Результат операции (например, { message: "..." })
 * @throws {Error} - В случае ошибки запроса
 */
export const submitVacationRequest = async (id) => {
  try {
    const response = await authApi.post(`/vacations/requests/${id}/submit`);
    return response.data;
  } catch (error) {
     console.error("API Error in submitVacationRequest:", error);
     if (error.response && error.response.data && error.response.data.error) {
      throw new Error(error.response.data.error);
    }
    throw new Error('Не удалось отправить заявку руководителю.');
  }
};

/**
 * Получение собственных заявок на отпуск пользователя
 * @param {number} year - Год
 * @returns {Promise<Array>} - Список заявок пользователя
 * @param {number | null} [status] - Опциональный ID статуса для фильтрации
 * @returns {Promise<Array>} - Список заявок пользователя
 * @throws {Error} - В случае ошибки запроса
 */
export const getMyVacations = async (year, status = null) => {
  try {
    const params = { year };
    if (status !== null) {
      params.status = status;
    }
    const response = await authApi.get('/vacations/my', { params });
    return response.data;
  } catch (error) {
    console.error("API Error in getMyVacations:", error);
    if (error.response && error.response.data && error.response.data.error) {
      throw new Error(error.response.data.error);
    }
    throw new Error('Не удалось получить список ваших заявок.');
  }
};

/**
 * Получение заявок сотрудников организационного юнита (для руководителя/админа)
 * @param {number} unitId - ID организационного юнита
 * @param {number} year - Год
 * @param {number | null} [status] - Опциональный ID статуса для фильтрации
 * @returns {Promise<Array>} - Список заявок сотрудников юнита
 * @throws {Error} - В случае ошибки запроса
 */
export const getUnitVacations = async (unitId, year, status = null) => { // Renamed function and parameter
  try {
    const params = { year };
    if (status !== null) {
      params.status = status;
    }
    // unitId передается как параметр пути, исправлен URL
    const response = await authApi.get(`/vacations/unit/${unitId}`, { params }); 
    return response.data;
  } catch (error) {
    console.error("API Error in getUnitVacations:", error); // Renamed function in log
    if (error.response && error.response.data && error.response.data.error) {
      throw new Error(error.response.data.error);
    }
    throw new Error('Не удалось получить заявки сотрудников подразделения.'); // Error message remains the same for user
  }
};

/**
 * Получение пересечений отпусков в подразделении (для руководителя)
 * @param {number} unitId - ID организационного юнита
 * @param {number} year - Год
 * @returns {Promise<Array>} - Список пересечений
 * @throws {Error} - В случае ошибки запроса
 */
export const getVacationIntersections = async (unitId, year) => { // Renamed parameter
  try {
    const response = await authApi.get('/vacations/intersections', {
      params: { unitId, year } // Renamed query parameter to unitId
    });
    return response.data;
  } catch (error) {
    console.error("API Error in getVacationIntersections:", error);
    if (error.response && error.response.data && error.response.data.error) {
      throw new Error(error.response.data.error);
    }
    throw new Error('Не удалось получить пересечения отпусков.');
  }
};

/**
 * Получение списка статусов заявок
 * @returns {Promise<Array>} - Список статусов { id, name, description }
 * @throws {Error} - В случае ошибки запроса
 */
export const getVacationStatuses = async () => {
  try {
    const response = await authApi.get('/vacations/statuses'); // Предполагаем такой эндпоинт
    return response.data;
  } catch (error) {
    console.error("API Error in getVacationStatuses:", error);
     if (error.response && error.response.data && error.response.data.error) {
      throw new Error(error.response.data.error);
    }
    throw new Error('Не удалось получить список статусов заявок.');
  }
};

/**
 * Утверждение заявки на отпуск (для руководителя)
 * @param {number} id - ID заявки
 * @param {boolean} [force=false] - Флаг для принудительного утверждения при конфликтах
 * @returns {Promise<Object>} - Результат операции ({ message: "...", warnings?: [...] })
 * @throws {Error|ConflictError} - Обычная ошибка или ошибка конфликта с данными { conflicts: [...] }
 */
export const approveVacationRequest = async (id, force = false) => {
  try {
    const url = `/vacations/requests/${id}/approve${force ? '?force=true' : ''}`;
    const response = await authApi.post(url);
    // Возвращаем весь объект data, так как он может содержать warnings при force=true
    return response.data;
  } catch (error) {
    console.error("API Error in approveVacationRequest:", error);
    // Специальная обработка для конфликта 409
    if (error.response && error.response.status === 409 && error.response.data) {
        // Создаем специфическую ошибку или объект, содержащий конфликты
        const conflictError = new Error(error.response.data.error || 'Обнаружены конфликты.');
        conflictError.isConflict = true; // Флаг для идентификации ошибки конфликта
        conflictError.conflicts = error.response.data.conflicts || []; // Данные о конфликтах
        throw conflictError;
    }
    // Общая обработка ошибок
    if (error.response && error.response.data && error.response.data.error) {
      throw new Error(error.response.data.error);
    }
    throw new Error('Не удалось утвердить заявку.');
  }
};

/**
 * Отклонение заявки на отпуск (для руководителя)
 * @param {number} id - ID заявки
 * @param {string} reason - Причина отклонения
 * @returns {Promise<Object>} - Результат операции
 * @throws {Error} - В случае ошибки запроса
 */
export const rejectVacationRequest = async (id, reason = '') => { // Делаем reason опциональным
  try {
    // Эндпоинт подтвержден в main.go
    const response = await authApi.post(`/vacations/requests/${id}/reject`, { reason });
    return response.data;
  } catch (error) {
    console.error("API Error in rejectVacationRequest:", error);
    if (error.response && error.response.data && error.response.data.error) {
      throw new Error(error.response.data.error);
    }
    throw new Error('Не удалось отклонить заявку.');
  } 
}; // <-- Эта скобка закрывает функцию rejectVacationRequest

/**
 * Получение всех УТВЕРЖДЕННЫХ заявок с ФИО (для календаря)
 * @param {Object} filters - Объект с фильтрами { year?, userId?, unitId? }
 * @returns {Promise<Array>} - Список утвержденных заявок в формате VacationRequestAdminView
 * @throws {Error} - В случае ошибки запроса
 */
export const getApprovedVacationsForCalendar = async (filters = {}) => {
  // Устанавливаем фильтр по статусу "Утверждена" (ID 3)
  const requiredFilters = { ...filters, status: 3 };
  return getAllVacations(requiredFilters); // Используем существующую функцию getAllVacations
};

/**
 * Установка лимита отпуска для пользователя (для администратора)
 * @param {number} userId - ID пользователя
 * @param {number} year - Год
 * @param {number} totalDays - Новое количество дней отпуска
 * @returns {Promise<Object>} - Результат операции (например, { message: "..." })
 * @throws {Error} - В случае ошибки запроса
 */
export const setVacationLimit = async (userId, year, totalDays) => {
  try {
    const response = await authApi.post('/admin/vacation-limits', { 
      user_id: userId, // Используем snake_case, как ожидает бэкенд
      year: year,
      total_days: totalDays 
    });
    return response.data;
  } catch (error) {
    console.error("API Error in setVacationLimit:", error);
    if (error.response && error.response.data && error.response.data.error) {
      throw new Error(error.response.data.error);
    }
    throw new Error('Не удалось установить лимит отпуска.');
  }
};

/**
 * Отмена заявки (для пользователя, если она еще не утверждена)
 * @param {number} id - ID заявки
 * @returns {Promise<Object>} - Результат операции
 * @throws {Error} - В случае ошибки запроса
 */
export const cancelVacationRequest = async (id) => {
    try {
        // Эндпоинт подтвержден в main.go
        const response = await authApi.post(`/vacations/requests/${id}/cancel`);
        return response.data;
    } catch (error) {
        console.error("API Error in cancelVacationRequest:", error);
        if (error.response && error.response.data && error.response.data.error) {
            throw new Error(error.response.data.error);
        }
        throw new Error('Не удалось отменить заявку.');
    } // <-- Добавлена пропущенная скобка для catch
};

/**
 * Получение всех заявок с фильтрами (для админа/менеджера)
 * @param {Object} filters - Объект с фильтрами { year?, status?, userId?, departmentId? }
 * @returns {Promise<Array>} - Список заявок в формате VacationRequestAdminView
 * @throws {Error} - В случае ошибки запроса
 */
export const getAllVacations = async (filters = {}) => {
  try {
    // Убираем null/undefined значения из фильтров
    const validFilters = {};
    for (const key in filters) {
      if (filters[key] !== null && filters[key] !== undefined) {
        validFilters[key] = filters[key];
      }
    }
    const response = await authApi.get('/vacations/all', { params: validFilters });
    return response.data;
  } catch (error) {
    console.error("API Error in getAllVacations:", error);
    if (error.response && error.response.data && error.response.data.error) {
      throw new Error(error.response.data.error);
    }
    throw new Error('Не удалось получить список всех заявок.');
  }
};

/**
 * Получение утвержденных конфликтов отпусков для видимых юнитов
 * @param {string} startDate - Начальная дата в формате YYYY-MM-DD
 * @param {string} endDate - Конечная дата в формате YYYY-MM-DD
 * @returns {Promise<Array>} - Список конфликтов (ConflictingPeriod)
 * @throws {Error} - В случае ошибки запроса
 */
export const getVacationConflicts = async (startDate, endDate) => {
  try {
    const params = { startDate, endDate };
    const response = await authApi.get('/vacations/conflicts', { params });
    return response.data;
  } catch (error) {
    console.error("API Error in getVacationConflicts:", error);
    if (error.response && error.response.data && error.response.data.error) {
      throw new Error(error.response.data.error);
    }
    throw new Error('Не удалось получить конфликты отпусков.');
  }
};

/**
 * Получение данных для дашборда руководителя
 * @returns {Promise<Object>} - Данные дашборда (ManagerDashboardData)
 * @throws {Error} - В случае ошибки запроса
 */
export const getManagerDashboardData = async () => {
  try {
    const response = await authApi.get('/dashboard/manager');
    return response.data;
  } catch (error) {
    console.error("API Error in getManagerDashboardData:", error);
    if (error.response && error.response.data && error.response.data.error) {
      throw new Error(error.response.data.error);
    }
    throw new Error('Не удалось получить данные дашборда.');
  }
};

/**
 * Получение данных отпусков для экспорта по выбранным юнитам (для админа)
 * @param {number[]} unitIds - Массив ID организационных юнитов
 * @param {number} [year] - Опциональный год (по умолчанию текущий на бэкенде)
 * @returns {Promise<Array>} - Массив объектов VacationExportRow
 * @throws {Error} - В случае ошибки запроса
 */
export const exportVacationsByUnits = async (unitIds, year = null) => {
  try {
    const payload = { unit_ids: unitIds };
    if (year !== null) {
      payload.year = year;
    }
    // Используем POST-запрос, как определено в main.go
    const response = await authApi.post('/admin/vacations/export', payload);
    return response.data; // Ожидаем массив VacationExportRow
  } catch (error) {
    console.error("API Error in exportVacationsByUnits:", error);
    if (error.response && error.response.data && error.response.data.error) {
      throw new Error(error.response.data.error);
    }
    throw new Error('Не удалось получить данные для экспорта отпусков.');
  }
};
