import React, { useState, useEffect, useCallback } from 'react';
import { toast } from 'react-toastify';
import { getAllUsersAdmin, updateUserAdmin } from '../../api/users'; // <-- Возвращаем именованный импорт
import { getOrganizationalUnitTree } from '../../api/units'; // Для получения списка юнитов
import { getPositions } from '../../api/auth'; // Исправлено: getAllPositions -> getPositions
import Loader from '../../components/ui/Loader/Loader';
import './UserManagementPage.css'; // Создадим файл стилей

const UserManagementPage = () => {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [editingUser, setEditingUser] = useState(null); // ID пользователя, которого редактируем
  const [editFormData, setEditFormData] = useState({}); // Данные формы редактирования

  // Состояния для хранения списков юнитов и должностей
  const [units, setUnits] = useState([]);
  const [positions, setPositions] = useState([]);
  const [loadingOptions, setLoadingOptions] = useState(true);

  // Функция для загрузки пользователей
  const fetchUsers = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await getAllUsersAdmin(); // <-- Используем прямой импорт
      setUsers(data || []); // Убедимся, что users всегда массив
    } catch (err) {
      setError(err.message || 'Не удалось загрузить пользователей');
      toast.error(err.message || 'Не удалось загрузить пользователей');
    } finally {
      setLoading(false);
    }
  }, []);

  // Функция для загрузки юнитов и должностей
  const fetchOptions = useCallback(async () => {
    setLoadingOptions(true);
    try {
      const [unitTree, positionsData] = await Promise.all([
        getOrganizationalUnitTree(), // Получаем дерево юнитов
        getPositions() // Исправлено: getAllPositions -> getPositions
      ]);

      // Преобразуем дерево юнитов в плоский список для select
      const flattenUnits = (nodes) => {
        let flatList = [];
        nodes.forEach(node => {
          flatList.push({ id: node.id, name: node.name, type: node.unit_type }); // Сохраняем ID и имя
          if (node.children && node.children.length > 0) {
            flatList = flatList.concat(flattenUnits(node.children));
          }
        });
        return flatList;
      };
      setUnits(flattenUnits(unitTree || []));
      setPositions(positionsData || []);

    } catch (err) {
      console.error("Ошибка загрузки опций:", err);
      toast.error('Не удалось загрузить списки подразделений или должностей');
    } finally {
      setLoadingOptions(false);
    }
  }, []);

  // Загрузка данных при монтировании компонента
  useEffect(() => {
    fetchUsers();
    fetchOptions();
  }, [fetchUsers, fetchOptions]);

  // Обработчик начала редактирования
  const handleEditClick = (user) => {
    setEditingUser(user.id);
    // Ищем текущий юнит и должность пользователя
    const currentUnit = units.find(u => u.name === user.department); // Сравниваем по имени, т.к. в DTO приходит имя
    const currentPosition = positions.find(p => p.name === user.positionName);

    setEditFormData({
      position_id: currentPosition ? currentPosition.id : '', // Устанавливаем ID
      organizational_unit_id: currentUnit ? currentUnit.id : '', // Устанавливаем ID
      is_admin: user.is_admin,
      is_manager: user.is_manager,
    });
  };

  // Обработчик отмены редактирования
  const handleCancelClick = () => {
    setEditingUser(null);
    setEditFormData({});
  };

  // Обработчик изменения данных в форме редактирования
  const handleInputChange = (event) => {
    const { name, value, type, checked } = event.target;
    setEditFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value,
    }));
  };

  // Обработчик сохранения изменений
  const handleSaveClick = async (userId) => {
    // Преобразуем ID в числа, если они строки
    const dataToSend = {
      position_id: editFormData.position_id ? parseInt(editFormData.position_id, 10) : null,
      organizational_unit_id: editFormData.organizational_unit_id ? parseInt(editFormData.organizational_unit_id, 10) : null,
      is_admin: editFormData.is_admin,
      is_manager: editFormData.is_manager,
    };

    // Удаляем null значения, чтобы не отправлять их, если не выбрано
    if (dataToSend.position_id === null) delete dataToSend.position_id;
    if (dataToSend.organizational_unit_id === null) delete dataToSend.organizational_unit_id;


    try {
      await updateUserAdmin(userId, dataToSend); // <-- Используем прямой импорт
      toast.success('Данные пользователя успешно обновлены');
      setEditingUser(null);
      fetchUsers(); // Обновляем список пользователей
    } catch (err) {
      toast.error(err.message || 'Не удалось обновить данные пользователя');
    }
  };

  if (loading || loadingOptions) {
    return <Loader />;
  }

  if (error) {
    return <div className="error-message">Ошибка: {error}</div>;
  }

  return (
    <div className="user-management-page">
      <h2>Управление пользователями</h2>
      <table className="users-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Логин</th>
            <th>ФИО</th>
            <th>Должность</th>
            <th>Подразделение</th>
            <th>Админ</th>
            <th>Менеджер</th>
            <th>Действия</th>
          </tr>
        </thead>
        <tbody>
          {users.map(user => (
            <tr key={user.id}>
              <td>{user.id}</td>
              <td>{user.login}</td>
              <td>{user.full_name}</td>
              <td>
                {editingUser === user.id ? (
                  <select
                    name="position_id"
                    value={editFormData.position_id || ''}
                    onChange={handleInputChange}
                  >
                    <option value="">-- Выберите должность --</option>
                    {positions.map(pos => (
                      <option key={pos.id} value={pos.id}>{pos.name}</option>
                    ))}
                  </select>
                ) : (
                  user.positionName || 'Не указана'
                )}
              </td>
              <td>
                {editingUser === user.id ? (
                  <select
                    name="organizational_unit_id"
                    value={editFormData.organizational_unit_id || ''}
                    onChange={handleInputChange}
                  >
                    <option value="">-- Выберите подразделение --</option>
                    {units.map(unit => (
                      <option key={unit.id} value={unit.id}>{unit.name} ({unit.type})</option>
                    ))}
                  </select>
                ) : (
                  // Отображаем Department, т.к. GetAllUsers возвращает только его
                  user.department || 'Не указано'
                )}
              </td>
              <td>
                {editingUser === user.id ? (
                  <input
                    type="checkbox"
                    name="is_admin"
                    checked={editFormData.is_admin || false}
                    onChange={handleInputChange}
                  />
                ) : (
                  user.is_admin ? 'Да' : 'Нет'
                )}
              </td>
              <td>
                {editingUser === user.id ? (
                  <input
                    type="checkbox"
                    name="is_manager"
                    checked={editFormData.is_manager || false}
                    onChange={handleInputChange}
                  />
                ) : (
                  user.is_manager ? 'Да' : 'Нет'
                )}
              </td>
              <td>
                {editingUser === user.id ? (
                  <>
                    <button onClick={() => handleSaveClick(user.id)} className="save-btn">Сохранить</button>
                    <button onClick={handleCancelClick} className="cancel-btn">Отмена</button>
                  </>
                ) : (
                  <button onClick={() => handleEditClick(user)} className="edit-btn">Редактировать</button>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default UserManagementPage;
