import React, { useState, useContext } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { login } from '../../api/auth'; // Исправляем импорт на 'login'
import { UserContext } from '../../context/UserContext'; // Для обновления состояния пользователя
import './LoginPage.css'; // Импортируем новые стили

function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { setUser } = useContext(UserContext); // Получаем функцию для установки пользователя

  const handleSubmit = async (event) => {
    // console.log('Login form submitted. handleSubmit triggered.'); // <-- Убираем лог
    event.preventDefault();
    setError('');
    setLoading(true);

    try {
      const responseData = await login({ email, password }); // Переименуем для ясности
      // console.log('Login successful:', responseData); // Лог успешного входа
      
      // Передаем только объект пользователя в setUser
      // Используем optional chaining на случай, если user не пришел
      if (responseData?.user) {
        setUser(responseData.user); 
      } else {
         // Обработка случая, когда пользователь не вернулся в ответе, но ошибки не было
         console.error("Данные пользователя не получены после входа.");
         setError('Не удалось получить данные пользователя.');
         setLoading(false); // Останавливаем загрузку
         return; // Прерываем выполнение handleSubmit
      }

      // Проверяем роль внутри объекта user, используем optional chaining
      if (responseData?.user?.role === 'admin') {
        navigate('/admin-dashboard');
      } else if (responseData?.user?.role === 'manager') {
        navigate('/manager-dashboard');
      } else {
        navigate('/profile'); // Или '/dashboard' для обычных пользователей (включая случай, если role не определена)
      }
    } catch (err) {
      // console.error('Login error:', err); // Лог ошибки
      setError(err.message || 'Ошибка входа. Проверьте введенные данные.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-page login-page"> {/* Добавляем класс login-page для специфичных стилей */}
      <div className="auth-container login-container-size"> {/* Новый класс для управления размером */}
        <div className="form-container"> {/* Убираем login-container класс отсюда */}
          <h2>Вход</h2>
          <form> {/* Снова убираем onSubmit отсюда */}
            {error && <p className="error-message">{error}</p>}
            <div className="form-group">
              <label htmlFor="login-email">Email</label>
              <input
                type="email"
                id="login-email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                disabled={loading}
              />
            </div>
            <div className="form-group">
              <label htmlFor="login-password">Пароль</label>
              <input
                type="password"
                id="login-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                disabled={loading}
              />
            </div>
            {/* Снова добавляем onClick сюда */}
            <button type="submit" onClick={handleSubmit} disabled={loading}> 
              {loading ? 'Вход...' : 'Войти'}
            </button>
          </form>
          <Link to="/register" className="toggle-link"> {/* Используем Link вместо button */}
            Нет аккаунта? Зарегистрируйтесь
          </Link>
        </div>
      </div>
    </div>
  );
}

export default LoginPage;
