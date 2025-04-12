import React, { useState, useEffect, useRef } from 'react'; // Убран useContext
// import { UserContext } from '../../context/UserContext'; // Больше не нужен
import { updateUserProfile, getMyProfile } from '../../api/users'; // Добавлен getMyProfile
// import { getPositions } from '../../api/positions'; // TODO: Если нужно будет редактировать должность
import './UserProfilePage.css';
import Loader from '../../components/ui/Loader/Loader';

function UserProfilePage() {
  // Убрано использование UserContext, данные профиля загружаются отдельно
  const [profileData, setProfileData] = useState(null); // Состояние для данных профиля
  const [isProfileLoading, setIsProfileLoading] = useState(true); // Состояние загрузки профиля
  const [formData, setFormData] = useState({
    full_name: '',
    password: '',
    confirmPassword: '',
    // position_id: '', // Пока не редактируем должность для себя
  });
  const [isFormLoading, setIsFormLoading] = useState(false); // Переименовано isLoading в isFormLoading
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');
  const successTimeoutRef = useRef(null); // Ref для хранения ID тайм-аута
  // const [positions, setPositions] = useState([]); // TODO: Если нужно будет редактировать должность

  // Загрузка данных профиля при монтировании компонента
  useEffect(() => {
    const fetchProfile = async () => {
      setIsProfileLoading(true);
      setError('');
      try {
        const data = await getMyProfile();
        setProfileData(data);
        // Заполняем форму данными из профиля
        setFormData((prevData) => ({
          ...prevData,
          full_name: data.full_name || '',
          // Очищаем поля пароля при загрузке
          password: '',
          confirmPassword: '',
          // position_id: data.position_id || '', // Пока не редактируем
        }));
      } catch (err) {
        setError(err.message || 'Не удалось загрузить профиль');
        console.error("Ошибка загрузки профиля:", err);
      } finally {
        setIsProfileLoading(false);
      }
    };

    fetchProfile();

    // TODO: Загрузить список должностей, если нужно (остается)
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
  }, []); // Пустой массив зависимостей, чтобы выполнилось один раз при монтировании


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

    setIsFormLoading(true); // Используем isFormLoading

    const updateData = {};
    // Используем fullName из profileData
    if (formData.full_name && formData.full_name !== profileData.full_name) {
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

    // Проверяем, есть ли изменения
    if (Object.keys(updateData).length === 0) {
      setError('Нет изменений для сохранения.');
      setIsFormLoading(false); // <-- Заменяем на setIsFormLoading
      return;
    }

    try {
      // Используем ID из profileData
      const response = await updateUserProfile(profileData.id, updateData);
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


      // Обновляем только поле full_name в formData, пароли очищаются
      setFormData((prevData) => ({
        ...prevData,
        password: '', // Очищаем пароль
        confirmPassword: '', // Очищаем подтверждение
      }));

      // Обновляем данные в локальном состоянии profileData
      if (updateData.full_name) {
          setProfileData(prevProfile => ({
              ...prevProfile,
              full_name: updateData.full_name,
              updated_at: new Date().toISOString(), // Обновляем время обновления (приблизительно)
          }));
          // Обновляем localStorage (если используется для хранения данных пользователя)
          // Важно: сохраняем ПОЛНЫЙ объект профиля, а не только user из старого контекста
          try {
            const storedUser = JSON.parse(localStorage.getItem('user') || '{}');
            const updatedStoredUser = { ...storedUser, fullName: updateData.full_name };
            localStorage.setItem('user', JSON.stringify(updatedStoredUser));
             // Можно также обновить UserContext глобально, если он используется где-то еще,
             // но для этой страницы достаточно локального обновления profileData
             // const { setUser } = useContext(UserContext); // Потребуется снова импортировать useContext и UserContext
             // setUser(updatedStoredUser); // Если UserContext все еще нужен глобально
          } catch (e) {
            console.error("Ошибка обновления localStorage:", e);
          }
      }


    } catch (err) {
      setError(err.message || 'Произошла ошибка при обновлении профиля.');
    } finally {
      setIsFormLoading(false); // Используем isFormLoading
    }
  };

  // Показываем загрузчик, пока профиль грузится
  if (isProfileLoading) {
    return <Loader />;
  }

  // Показываем ошибку, если профиль не загрузился
  if (error && !profileData) {
      return <div className="error-message">{error}</div>;
  }

  // Если профиль не загружен (маловероятно после isProfileLoading), показываем сообщение
  if (!profileData) {
     return <p>Не удалось загрузить данные профиля.</p>;
  }

  return (
    <div className="profile-page">
      <h2>Профиль пользователя</h2>
      <div className="profile-info">
        {/* Отображаем данные из profileData */}
        <p><strong>ФИО:</strong> {profileData.full_name || 'Не указано'}</p>
        <p><strong>Логин:</strong> {profileData.login}</p>
        {/* Отображаем иерархию юнитов */}
        {profileData.department && <p><strong>Департамент:</strong> {profileData.department}</p>}
        {profileData.subDepartment && <p><strong>Подотдел:</strong> {profileData.subDepartment}</p>}
        {profileData.sector && <p><strong>Сектор:</strong> {profileData.sector}</p>}
        <p>
            <strong>Должность:</strong>{' '}
            <span className={`user-position-badge ${getPositionLevelClass(profileData.positionName)}`}>
                {profileData.positionName || 'Не указана'}
            </span>
        </p>
        <p><strong>Дата создания:</strong> {new Date(profileData.created_at).toLocaleString()}</p>
        <p><strong>Дата обновления:</strong> {new Date(profileData.updated_at).toLocaleString()}</p>
        {/* Отображение ролей, если нужно */}
        {(profileData.is_admin || profileData.is_manager) && (
            <p><strong>Роли:</strong>
                {profileData.is_admin && <span className="role-badge admin-badge">Администратор</span>}
                {profileData.is_manager && <span className="role-badge manager-badge">Руководитель</span>}
            </p>
        )}
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
           // Используем profileData для сравнения
           const originalFullName = profileData?.full_name || '';
           const hasNameChanged = formData.full_name !== originalFullName;
           const hasPasswordChanged = !!formData.password;
           // TODO: Добавить проверку изменения должности, если она будет добавлена
           // const originalPositionId = profileData?.position_id || '';
           // const hasPositionChanged = formData.position_id !== originalPositionId;
           const hasChanges = hasNameChanged || hasPasswordChanged; // || hasPositionChanged;

           return (
              <button type="submit" className="save-button" disabled={isFormLoading || !hasChanges}>
                {isFormLoading ? <Loader size="small" /> : 'Сохранить изменения'}
              </button>
           );
        })()}
      </form>
    </div>
  );
}

// Вспомогательная функция для определения класса стиля должности
const getPositionLevelClass = (positionName) => {
  if (!positionName) return 'position-level-0'; // По умолчанию
  const lowerPos = positionName.toLowerCase();
  // Уровни важности (можно настроить)
  if (lowerPos.includes('директор') || lowerPos.includes('руководитель') || lowerPos.includes('начальник')) {
    return 'position-level-3'; // Красный
  } else if (lowerPos.includes('менеджер') || lowerPos.includes('ведущий') || lowerPos.includes('старший')) {
    return 'position-level-2'; // Оранжевый
  } else if (lowerPos.includes('специалист') || lowerPos.includes('инженер') || lowerPos.includes('бухгалтер')) {
     return 'position-level-1'; // Желтый
  } else {
    return 'position-level-0'; // Синий (обычный сотрудник)
  }
};

export default UserProfilePage;
