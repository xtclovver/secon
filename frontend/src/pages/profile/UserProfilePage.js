import React, { useState, useEffect, useContext } from 'react';
import { UserContext } from '../../context/UserContext';
import { updateUserProfile } from '../../api/users'; // Предполагаем, что функция API создана
// import { getPositions } from '../../api/positions'; // TODO: Если нужно будет редактировать должность
import './UserProfilePage.css';
import Loader from '../../components/ui/Loader/Loader';

function UserProfilePage() {
  const { user: currentUser, setUser } = useContext(UserContext); // Получаем текущего пользователя и функцию для его обновления
  const [formData, setFormData] = useState({
    full_name: '',
    password: '',
    confirmPassword: '',
    // position_id: '', // Пока не редактируем должность для себя
  });
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');
  // const [positions, setPositions] = useState([]); // TODO: Если нужно будет редактировать должность

  // Заполняем форму данными текущего пользователя при загрузке
  useEffect(() => {
    if (currentUser) {
      setFormData((prevData) => ({
        ...prevData,
        full_name: currentUser.full_name || '',
        // position_id: currentUser.position_id || '', // Пока не редактируем
      }));
    }
    // TODO: Загрузить список должностей, если нужно
    // const fetchPositions = async () => {
    //   try {
    //     const data = await getPositions();
    //     setPositions(data);
    //   } catch (err) {
    //     console.error("Ошибка загрузки должностей:", err);
    //   }
    // };
    // fetchPositions();
  }, [currentUser]);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prevData) => ({
      ...prevData,
      [name]: value,
    }));
    // Очищаем сообщения при изменении полей
    setError('');
    setSuccessMessage('');
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setSuccessMessage('');

    if (formData.password && formData.password !== formData.confirmPassword) {
      setError('Пароли не совпадают');
      return;
    }

    setIsLoading(true);

    const updateData = {};
    if (formData.full_name && formData.full_name !== currentUser.full_name) {
      updateData.full_name = formData.full_name;
    }
    if (formData.password) {
      updateData.password = formData.password;
    }
    // TODO: Добавить position_id, если редактирование разрешено
    // if (currentUser?.is_admin || currentUser?.is_manager) { // Пример условия
    //   if (formData.position_id && formData.position_id !== currentUser.position_id) {
    //      updateData.position_id = parseInt(formData.position_id, 10); // Убедимся, что это число
    //   }
    // }

    if (Object.keys(updateData).length === 0) {
      setError('Нет изменений для сохранения.');
      setIsLoading(false);
      return;
    }

    try {
      const response = await updateUserProfile(currentUser.id, updateData);
      setSuccessMessage(response.message || 'Профиль успешно обновлен!');
      // Очищаем поля пароля после успешного обновления
      setFormData((prevData) => ({
        ...prevData,
        password: '',
        confirmPassword: '',
      }));

      // Обновляем данные пользователя в контексте, если ФИО изменилось
      if (updateData.full_name) {
          // Создаем новый объект пользователя с обновленным именем
          const updatedUser = { ...currentUser, full_name: updateData.full_name };
          // Обновляем контекст
          setUser(updatedUser);
          // Обновляем localStorage (если используется для хранения данных пользователя)
          localStorage.setItem('user', JSON.stringify(updatedUser));
      }


    } catch (err) {
      setError(err.message || 'Произошла ошибка при обновлении профиля.');
    } finally {
      setIsLoading(false);
    }
  };

  if (!currentUser) {
    return <Loader />; // Показываем загрузчик, пока данные пользователя не загружены
  }

  return (
    <div className="profile-page">
      <h2>Профиль пользователя</h2>
      <div className="profile-info">
        <p><strong>Имя пользователя:</strong> {currentUser.username}</p>
        <p><strong>Email:</strong> {currentUser.email}</p>
        {/* TODO: Отобразить название отдела и должности, если они есть */}
        {/* <p><strong>Отдел:</strong> {currentUser.department_name || 'Не указан'}</p> */}
        {/* <p><strong>Должность:</strong> {currentUser.position_name || 'Не указана'}</p> */}
        <p><strong>Дата создания:</strong> {new Date(currentUser.created_at).toLocaleDateString()}</p>
        <p><strong>Дата обновления:</strong> {new Date(currentUser.updated_at).toLocaleDateString()}</p>
        <p><strong>Роли:</strong> {currentUser.is_admin ? 'Администратор' : ''} {currentUser.is_manager ? 'Руководитель' : ''} {!currentUser.is_admin && !currentUser.is_manager ? 'Сотрудник' : ''}</p>
      </div>

      <form onSubmit={handleSubmit} className="profile-form">
        <h3>Редактировать профиль</h3>

        {error && <p className="error-message">{error}</p>}
        {successMessage && <p className="success-message">{successMessage}</p>}

        <div className="form-group">
          <label htmlFor="full_name">ФИО:</label>
          <input
            type="text"
            id="full_name"
            name="full_name"
            value={formData.full_name}
            onChange={handleChange}
            required
          />
        </div>

        <div className="form-group">
          <label htmlFor="password">Новый пароль (оставьте пустым, чтобы не менять):</label>
          <input
            type="password"
            id="password"
            name="password"
            value={formData.password}
            onChange={handleChange}
            autoComplete="new-password"
          />
        </div>

        {formData.password && ( // Показываем поле подтверждения только если введен новый пароль
          <div className="form-group">
            <label htmlFor="confirmPassword">Подтвердите новый пароль:</label>
            <input
              type="password"
              id="confirmPassword"
              name="confirmPassword"
              value={formData.confirmPassword}
              onChange={handleChange}
              required={!!formData.password} // Обязательно, если введен новый пароль
              autoComplete="new-password"
            />
          </div>
        )}

        {/* TODO: Поле для редактирования должности (только для админа/менеджера) */}
        {/* { (currentUser?.is_admin || currentUser?.is_manager) && (
          <div className="form-group">
            <label htmlFor="position_id">Должность:</label>
            <select
              id="position_id"
              name="position_id"
              value={formData.position_id}
              onChange={handleChange}
            >
              <option value="">Выберите должность</option>
              {positions.map(group => (
                <optgroup label={group.name} key={group.id}>
                  {group.positions.map(pos => (
                    <option key={pos.id} value={pos.id}>{pos.name}</option>
                  ))}
                </optgroup>
              ))}
            </select>
          </div>
        )} */}

        <button type="submit" disabled={isLoading}>
          {isLoading ? <Loader size="small" /> : 'Сохранить изменения'}
        </button>
      </form>
    </div>
  );
}

export default UserProfilePage;
