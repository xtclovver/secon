import React, { useState, useEffect, useCallback } from 'react';
import {
  getUnitChildren,
  getUnitUsersWithLimits, // <-- Импортируем функцию для получения пользователей с лимитами
  updateUserVacationLimit // <-- Импортируем функцию для обновления лимита
} from '../../api/units';
import Loader from '../../components/ui/Loader/Loader';
import './DepartmentManagementPage.css'; // Подключаем стили

// --- Компонент для отображения элемента списка юнитов (для навигации) ---
const UnitListItem = ({ item, onNavigate }) => {
  const handleNavigateClick = () => {
    if (onNavigate) {
      onNavigate(item.id, item.name); // Передаем ID и имя юнита для навигации и хлебных крошек
    }
  };

  return (
    <li className="list-item list-item--unit"> {/* Добавляем класс для стилизации */}
      <div className="item-info">
        <strong>{item.name}</strong>
        {item.unit_type && ` (${item.unit_type})`}
      </div>
      <div className="item-actions">
        <button onClick={handleNavigateClick} className="navigate-button">
          Войти <span aria-hidden="true">➔</span>
        </button>
        {/* TODO: Добавить кнопки Edit/Delete для админа */}
      </div>
    </li>
  );
};

// --- Компонент для отображения пользователя с редактируемым лимитом ---
const UserLimitItem = ({ user, year, onSave, initialLimit }) => {
  const [limit, setLimit] = useState(initialLimit); // Локальное состояние для инпута
  const [isSaving, setIsSaving] = useState(false);
  const [saveError, setSaveError] = useState(null);
  const [saveSuccess, setSaveSuccess] = useState(false);

  // Синхронизируем локальное состояние с initialLimit при его изменении
  useEffect(() => {
    setLimit(initialLimit);
  }, [initialLimit]);

  const handleInputChange = (e) => {
    setSaveError(null); // Сбрасываем ошибку при изменении
    setSaveSuccess(false); // Сбрасываем успех при изменении
    const value = e.target.value;
    // Разрешаем пустую строку или число >= 0
    if (value === '' || /^\d+$/.test(value)) {
        const numValue = value === '' ? '' : parseInt(value, 10);
         if (numValue === '' || numValue >= 0) {
             setLimit(numValue);
         }
    }
  };


  const handleSaveClick = async () => {
    if (limit === '' || limit < 0) {
        setSaveError("Лимит должен быть 0 или больше.");
        return;
      }
    setIsSaving(true);
    setSaveError(null);
    setSaveSuccess(false);
    try {
      await onSave(user.id, year, parseInt(limit, 10));
      setSaveSuccess(true);
      setTimeout(() => setSaveSuccess(false), 3000); // Скрываем сообщение об успехе через 3 сек
    } catch (err) {
      setSaveError(err.message || 'Ошибка сохранения');
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <li className="list-item list-item--user"> {/* Добавляем класс для стилизации */}
      <div className="item-info">
        <strong>{user.full_name}</strong>
        {user.position_name && ` (${user.position_name})`}
      </div>
      <div className="item-actions item-actions--limit">
        <input
          type="number"
          value={limit}
          onChange={handleInputChange}
          className="limit-input"
          min="0"
          disabled={isSaving}
        />
        <button
          onClick={handleSaveClick}
          className={`save-button ${saveSuccess ? 'success' : ''}`}
          disabled={isSaving || limit === initialLimit || limit === ''} // Блокируем, если не изменено или пусто
        >
          {isSaving ? 'Сохранение...' : saveSuccess ? 'Сохранено!' : 'Сохранить'}
        </button>
        {saveError && <span className="error-message save-error">{saveError}</span>}
      </div>
    </li>
  );
};


// --- Основной компонент страницы ---
const DepartmentManagementPage = () => {
  // Состояния для навигации по юнитам
  const [childUnits, setChildUnits] = useState([]); // Только дочерние ЮНИТЫ для навигации
  const [currentParentId, setCurrentParentId] = useState(null);
  const [breadcrumbs, setBreadcrumbs] = useState([{ id: null, name: 'Корень' }]);
  const [loadingUnits, setLoadingUnits] = useState(true);
  const [unitsError, setUnitsError] = useState(null);

  // Состояния для списка пользователей выбранного юнита
  const [unitUsers, setUnitUsers] = useState([]);
  const [loadingUsers, setLoadingUsers] = useState(false);
  const [usersError, setUsersError] = useState(null);
  const [selectedYear, setSelectedYear] = useState(new Date().getFullYear());
  // Состояние для хранения РЕДАКТИРУЕМЫХ значений лимитов
  // Ключ: userId, Значение: текущее значение в инпуте (строка или число)
  const [editingLimits, setEditingLimits] = useState({});

  // Функция для загрузки дочерних ЮНИТОВ (только юнитов)
  const fetchChildUnits = useCallback(async (parentId) => {
    setLoadingUnits(true);
    setUnitsError(null);
    setUnitUsers([]); // Очищаем список пользователей при переходе
    setEditingLimits({}); // Очищаем редактируемые лимиты
    try {
      const data = await getUnitChildren(parentId);
      // Фильтруем, оставляем только 'unit'
      setChildUnits(data.filter(item => item.type === 'unit') || []);
    } catch (err) {
      setUnitsError(err.message || 'Не удалось загрузить дочерние подразделения');
      setChildUnits([]);
    } finally {
      setLoadingUnits(false);
    }
  }, []);

  // Функция для загрузки пользователей выбранного юнита с лимитами
  const fetchUnitUsers = useCallback(async (unitId, year) => {
    if (unitId === null) {
        setUnitUsers([]);
        setEditingLimits({});
        return; // Не загружаем пользователей для корня
    }
    setLoadingUsers(true);
    setUsersError(null);
    try {
      const data = await getUnitUsersWithLimits(unitId, year);
      setUnitUsers(data || []);
      // Инициализируем editingLimits на основе полученных данных
      const initialLimits = {};
      (data || []).forEach(user => {
        // Используем 28 как дефолтное значение, если лимит null/undefined
        initialLimits[user.id] = user.total_days ?? 28;
      });
      setEditingLimits(initialLimits);
    } catch (err) {
      setUsersError(err.message || 'Не удалось загрузить сотрудников подразделения');
      setUnitUsers([]);
      setEditingLimits({});
    } finally {
      setLoadingUsers(false);
    }
  }, []);

  // Загрузка дочерних ЮНИТОВ при изменении currentParentId
  useEffect(() => {
    fetchChildUnits(currentParentId);
  }, [currentParentId, fetchChildUnits]);

  // Загрузка ПОЛЬЗОВАТЕЛЕЙ при изменении currentParentId (если не null) или selectedYear
  useEffect(() => {
    if (currentParentId !== null) {
        fetchUnitUsers(currentParentId, selectedYear);
    } else {
        setUnitUsers([]); // Очистка списка пользователей, если мы в корне
        setEditingLimits({});
    }
  }, [currentParentId, selectedYear, fetchUnitUsers]);


  // Обработчик навигации вглубь по юнитам
  const handleNavigateDown = (unitId, unitName) => {
    // Добавляем текущий уровень в хлебные крошки
    setBreadcrumbs(prev => [...prev, { id: unitId, name: unitName }]);
    // Устанавливаем новый родительский ID
    setCurrentParentId(unitId);
    // Загрузка пользователей для нового unitId сработает через useEffect
  };

  // Обработчик навигации по хлебным крошкам (вверх)
  const handleBreadcrumbClick = (index) => {
    const newBreadcrumbs = breadcrumbs.slice(0, index + 1);
    setBreadcrumbs(newBreadcrumbs);
    const newParentId = newBreadcrumbs[newBreadcrumbs.length - 1].id;
    setCurrentParentId(newParentId);
     // Если вернулись в корень, очищаем список пользователей
     if (newParentId === null) {
        setUnitUsers([]);
        setEditingLimits({});
     }
     // Загрузка юнитов и пользователей сработает через useEffect
  };

   // Обработчик сохранения лимита пользователя
   const handleSaveUserLimit = async (userId, year, limitValue) => {
    // Вызываем API для обновления
    await updateUserVacationLimit(userId, year, limitValue);
    // Обновляем initialLimit в состоянии editingLimits после успешного сохранения
    setEditingLimits(prev => ({ ...prev, [userId]: limitValue }));
    // Можно добавить логику для показа общего сообщения об успехе или обработать его в UserLimitItem
    // Примечание: перезагрузка списка пользователей не обязательна,
    // так как мы обновляем локальное состояние initialLimit
    // Если API вернет ошибку, она будет обработана в UserLimitItem
   };

  // Обработчик изменения года
  const handleYearChange = (e) => {
    setSelectedYear(parseInt(e.target.value, 10));
    // Загрузка пользователей для нового года сработает через useEffect
  };

  const currentUnitName = breadcrumbs[breadcrumbs.length - 1]?.name;

  return (
    <div className="department-management-page">
      <h1>Управление подразделениями</h1>

      {/* Хлебные крошки */}
      <nav aria-label="breadcrumb">
        <ol className="breadcrumbs">
          {breadcrumbs.map((crumb, index) => (
            <li key={crumb.id ?? 'root'} className={`breadcrumb-item ${index === breadcrumbs.length - 1 ? 'active' : ''}`}>
              {index < breadcrumbs.length - 1 ? (
                <button onClick={() => handleBreadcrumbClick(index)} className="breadcrumb-link">
                  {crumb.name}
                </button>
              ) : (
                <span>{crumb.name}</span>
              )}
            </li>
          ))}
        </ol>
      </nav>

      {/* --- Блок навигации по юнитам (виден всегда) --- */}
      <div className="unit-navigation">
        <h2>Дочерние подразделения:</h2>
        {loadingUnits && <Loader />}
        {unitsError && <p className="error-message">Ошибка: {unitsError}</p>}
        {!loadingUnits && !unitsError && (
          <ul className="items-list">
            {childUnits.length > 0 ? (
              childUnits.map(item => (
                <UnitListItem
                  key={`unit-${item.id}`}
                  item={item}
                  onNavigate={handleNavigateDown}
                />
              ))
            ) : (
              <p>Нет дочерних подразделений.</p>
            )}
          </ul>
        )}
      </div>

      {/* --- Блок управления пользователями (виден только если выбран юнит) --- */}
      {currentParentId !== null && (
        <div className="user-management">
          <h2>Сотрудники подразделения: {currentUnitName}</h2>

           {/* Выбор года */}
           <div className="year-selector">
             <label htmlFor="year-select">Год: </label>
             <select id="year-select" value={selectedYear} onChange={handleYearChange}>
               {/* Генерируем года, например, текущий +/- 2 года */}
               {[...Array(5)].map((_, i) => {
                 const year = new Date().getFullYear() - 2 + i;
                 return <option key={year} value={year}>{year}</option>;
               })}
             </select>
           </div>

          {loadingUsers && <Loader />}
          {usersError && <p className="error-message">Ошибка: {usersError}</p>}

          {!loadingUsers && !usersError && (
            <ul className="items-list">
              {unitUsers.length > 0 ? (
                unitUsers.map(user => (
                  <UserLimitItem
                    key={`user-${user.id}`}
                    user={user}
                    year={selectedYear}
                    onSave={handleSaveUserLimit}
                    // Передаем initialLimit из editingLimits, дефолт 28 если нет
                    initialLimit={editingLimits[user.id] ?? 28}
                  />
                ))
              ) : (
                <p>В этом подразделении нет сотрудников.</p>
              )}
            </ul>
          )}
        </div>
      )}

      {/* TODO: Добавить кнопку "Создать юнит/пользователя" */}

      {/* Старый блок отображения смешанного списка удален */}
    </div>
  );
};

export default DepartmentManagementPage; // <-- Добавляем экспорт в самый конец
