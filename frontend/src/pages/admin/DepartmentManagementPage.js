import React, { useState, useEffect, useCallback } from 'react';
import { getUnitChildren } from '../../api/units'; // Импортируем НОВОЕ API
import Loader from '../../components/ui/Loader/Loader';
import './DepartmentManagementPage.css'; // Подключаем стили

// Компонент для отображения элемента списка (юнит или пользователь)
const ListItem = ({ item, onNavigate }) => {
  const isUnit = item.type === 'unit';

  const handleNavigateClick = () => {
    if (isUnit && onNavigate) {
      onNavigate(item.id, item.name); // Передаем ID и имя юнита для навигации и хлебных крошек
    }
  };

  // TODO: Добавить обработчик для кнопки "Управлять лимитами" для пользователей
  const handleManageLimitsClick = () => {
    alert(`Управление лимитами для пользователя ${item.name} (ID: ${item.id}) - в разработке`);
  };

  return (
    <li className="list-item">
      <div className="item-info">
        <strong>{item.name}</strong>
        {isUnit && item.unit_type && ` (${item.unit_type})`}
        {!isUnit && item.position && ` (${item.position})`}
      </div>
      <div className="item-actions">
        {isUnit ? (
          <button onClick={handleNavigateClick} className="navigate-button">
            Войти <span aria-hidden="true">➔</span>
          </button>
        ) : (
          <button onClick={handleManageLimitsClick} className="manage-button">
            Управлять лимитами
          </button>
        )}
        {/* TODO: Добавить кнопки Edit/Delete для админа */}
      </div>
    </li>
  );
};

// Основной компонент страницы
const DepartmentManagementPage = () => {
  const [items, setItems] = useState([]);
  const [currentParentId, setCurrentParentId] = useState(null); // null для корневого уровня
  const [breadcrumbs, setBreadcrumbs] = useState([{ id: null, name: 'Корень' }]); // Начинаем с корня
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Функция для загрузки дочерних элементов
  const fetchChildren = useCallback(async (parentId) => {
    setLoading(true);
    setError(null);
    try {
      const data = await getUnitChildren(parentId);
      setItems(data || []);
    } catch (err) {
      setError(err.message || 'Не удалось загрузить дочерние элементы');
      setItems([]); // Очищаем список при ошибке
    } finally {
      setLoading(false);
    }
  }, []);

  // Загрузка данных при изменении currentParentId
  useEffect(() => {
    fetchChildren(currentParentId);
  }, [currentParentId, fetchChildren]);

  // Обработчик навигации вглубь
  const handleNavigateDown = (unitId, unitName) => {
    // Добавляем текущий уровень в хлебные крошки
    setBreadcrumbs(prev => [...prev, { id: unitId, name: unitName }]);
    // Устанавливаем новый родительский ID
    setCurrentParentId(unitId);
  };

  // Обработчик навигации по хлебным крошкам (вверх)
  const handleBreadcrumbClick = (index) => {
    // Обрезаем хлебные крошки до выбранного уровня
    const newBreadcrumbs = breadcrumbs.slice(0, index + 1);
    setBreadcrumbs(newBreadcrumbs);
    // Устанавливаем ID родителя из последнего элемента новых крошек
    setCurrentParentId(newBreadcrumbs[newBreadcrumbs.length - 1].id);
  };

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
                <span>{crumb.name}</span> // Текущий уровень не кликабельный
              )}
            </li>
          ))}
        </ol>
      </nav>

      {/* TODO: Добавить кнопку "Создать юнит/пользователя" здесь */}

      {loading && <Loader />}
      {error && <p className="error-message">Ошибка: {error}</p>}

      {!loading && !error && (
        <ul className="items-list">
          {items.length > 0 ? (
            items.map(item => (
              <ListItem
                key={`${item.type}-${item.id}`} // Уникальный ключ для юнитов и пользователей
                item={item}
                onNavigate={handleNavigateDown}
              />
            ))
          ) : (
            <p>В этом подразделении нет дочерних элементов.</p>
          )}
        </ul>
      )}
    </div>
  );
};

export default DepartmentManagementPage;
