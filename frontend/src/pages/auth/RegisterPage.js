import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { register, getPositions } from '../../api/auth'; // Импортируем API функции
import './RegisterPage.css'; // Стили для страницы регистрации
import Loader from '../../components/ui/Loader/Loader'; // Компонент загрузчика

function RegisterPage() {
  const [formData, setFormData] = useState({
    fullName: '',
    username: '',
    email: '', // Добавим поле email, так как оно есть в модели и API
    password: '',
    confirmPassword: '',
    positionId: '', // ID выбранной должности
  });
  const [positions, setPositions] = useState([]); // Список должностей для dropdown
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');
  const navigate = useNavigate();

  // Загрузка должностей при монтировании компонента
  useEffect(() => {
    const fetchPositions = async () => {
      setLoading(true);
      try {
        const data = await getPositions();
        setPositions(data || []); // Устанавливаем пустой массив, если данные не пришли
      } catch (err) {
        setError('Ошибка загрузки списка должностей: ' + (err.message || 'Неизвестная ошибка'));
      } finally {
        setLoading(false);
      }
    };
    fetchPositions();
  }, []);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prevState => ({
      ...prevState,
      [name]: value
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setSuccessMessage('');

    if (formData.password !== formData.confirmPassword) {
      setError('Пароли не совпадают');
      return;
    }

    setLoading(true);
    try {
      const registrationData = {
        full_name: formData.fullName,
        username: formData.username,
        email: formData.email,
        password: formData.password,
        confirm_password: formData.confirmPassword,
        // Преобразуем positionId в число или null, если не выбрано
        position_id: formData.positionId ? parseInt(formData.positionId, 10) : null,
      };
      await register(registrationData);
      setSuccessMessage('Регистрация прошла успешно! Теперь вы можете войти.');
      // Очистка формы после успеха
      setFormData({
        fullName: '', username: '', email: '', password: '', confirmPassword: '', positionId: ''
      });
      // Можно добавить небольшую задержку перед редиректом
      setTimeout(() => {
        navigate('/login');
      }, 2000); // Редирект на страницу входа через 2 секунды
    } catch (err) {
      setError('Ошибка регистрации: ' + (err.message || 'Неизвестная ошибка'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="register-page">
      <h2>Регистрация</h2>
      <form onSubmit={handleSubmit} className="register-form">
        {error && <p className="error-message">{error}</p>}
        {successMessage && <p className="success-message">{successMessage}</p>}
        {loading && <Loader />}

        <div className="form-group">
          <label htmlFor="fullName">ФИО:</label>
          <input
            type="text"
            id="fullName"
            name="fullName"
            value={formData.fullName}
            onChange={handleChange}
            required
            disabled={loading}
          />
        </div>
        <div className="form-group">
          <label htmlFor="username">Логин (имя пользователя):</label>
          <input
            type="text"
            id="username"
            name="username"
            value={formData.username}
            onChange={handleChange}
            required
            disabled={loading}
          />
        </div>
         <div className="form-group">
          <label htmlFor="email">Email:</label>
          <input
            type="email"
            id="email"
            name="email"
            value={formData.email}
            onChange={handleChange}
            required
            disabled={loading}
          />
        </div>
        <div className="form-group">
          <label htmlFor="password">Пароль:</label>
          <input
            type="password"
            id="password"
            name="password"
            value={formData.password}
            onChange={handleChange}
            required
            disabled={loading}
          />
        </div>
        <div className="form-group">
          <label htmlFor="confirmPassword">Повторите пароль:</label>
          <input
            type="password"
            id="confirmPassword"
            name="confirmPassword"
            value={formData.confirmPassword}
            onChange={handleChange}
            required
            disabled={loading}
          />
        </div>
        <div className="form-group">
          <label htmlFor="positionId">Должность:</label>
          <select
            id="positionId"
            name="positionId"
            value={formData.positionId}
            onChange={handleChange}
            required // Сделать обязательным? Решите сами
            disabled={loading || positions.length === 0} // Блокируем, если должности не загружены
          >
            <option value="">-- Выберите должность --</option>
            {positions.map(group => (
              <optgroup label={group.name} key={group.id}>
                {group.positions && group.positions.map(position => (
                  <option value={position.id} key={position.id}>
                    {position.name}
                  </option>
                ))}
              </optgroup>
            ))}
          </select>
        </div>
        <button type="submit" disabled={loading}>
          {loading ? 'Регистрация...' : 'Зарегистрироваться'}
        </button>
      </form>
       <p className="login-link">
         Уже есть аккаунт? <a href="/login">Войти</a>
       </p>
    </div>
  );
}

export default RegisterPage;
