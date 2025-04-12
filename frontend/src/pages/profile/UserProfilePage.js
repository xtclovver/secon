import React, { useState, useEffect, useContext, useRef } from 'react'; // Добавлен useRef
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
  const successTimeoutRef = useRef(null); // Ref для хранения ID тайм-аута
  // const [positions, setPositions] = useState([]); // TODO: Если нужно будет редактировать должность

  // Заполняем форму данными текущего пользователя при загрузке и очищаем тайм-аут при размонтировании
  useEffect(() => {
    if (currentUser) {
      setFormData((prevData) => ({
        ...prevData,
        // Используем fullName (после исправления transformUserKeys)
        full_name: currentUser.fullName || '',
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

    // Очистка тайм-аута при размонтировании компонента
    return () => {
        if (successTimeoutRef.current) {
            clearTimeout(successTimeoutRef.current);
        }
    };
  }, [currentUser]);


  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prevData) => ({
      ...prevData,
      [name]: value,
    }));
    // Очищаем сообщения об ошибках и успехе при любом изменении поля
    setError('');
    if (successMessage) setSuccessMessage(''); // Очищаем сообщение об успехе
    if (successTimeoutRef.current) { // Очищаем таймер, если он был
        clearTimeout(successTimeoutRef.current);
        successTimeoutRef.current = null;
    }
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
    // Используем fullName
    if (formData.full_name && formData.full_name !== currentUser.fullName) {
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
      const message = response.message || 'Профиль успешно обновлен!';
      setSuccessMessage(message);

      // Устанавливаем таймер для скрытия сообщения об успехе через 3 секунды
       if (successTimeoutRef.current) {
            clearTimeout(successTimeoutRef.current); // Очищаем предыдущий таймер, если есть
       }
      successTimeoutRef.current = setTimeout(() => {
        setSuccessMessage('');
        successTimeoutRef.current = null;
      }, 3000);


      // Очищаем поля пароля после успешного обновления
      setFormData((prevData) => ({
        ...prevData,
        password: '',
        confirmPassword: '',
      }));

      // Обновляем данные пользователя в контексте, если ФИО изменилось
      if (updateData.full_name) {
          // Создаем новый объект пользователя с обновленным именем
          // Используем fullName
          const updatedUser = { ...currentUser, fullName: updateData.full_name };
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
        <p><strong>Дата создания:</strong> {new Date(currentUser.created_at).toLocaleString()}</p>
        <p><strong>Дата обновления:</strong> {new Date(currentUser.updated_at).toLocaleString()}</p>
        {/* Возвращаем isAdmin и isManager */}
        <p><strong>Роли:</strong> {[currentUser.isAdmin && 'Администратор', currentUser.isManager && 'Руководитель', !currentUser.isAdmin && !currentUser.isManager && 'Сотрудник'].filter(Boolean).join(', ') || 'Сотрудник'}</p> 
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

        {/* Вычисляем, были ли изменения */}
        {(() => {
           // Убедимся, что currentUser существует перед доступом к его свойствам
           // Используем fullName
           const originalFullName = currentUser?.fullName || '';
           const hasNameChanged = formData.full_name !== originalFullName;
           const hasPasswordChanged = !!formData.password;
           // TODO: Добавить проверку изменения должности, если она будет добавлена
           // const originalPositionId = currentUser?.position_id || '';
           // const hasPositionChanged = formData.position_id !== originalPositionId;
           const hasChanges = hasNameChanged || hasPasswordChanged; // || hasPositionChanged;

           return (
              <button type="submit" className="save-button" disabled={isLoading || !hasChanges}>
                {isLoading ? <Loader size="small" /> : 'Сохранить изменения'}
              </button>
           );
        })()}
      </form>
    </div>
  );
}

export default UserProfilePage;
